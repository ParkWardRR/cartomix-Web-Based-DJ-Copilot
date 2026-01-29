import { create } from 'zustand';
import type { Track, SetPlan, SetEdge } from './types';
import * as api from './api';
import { demoTracks, demoSetPlan } from './mockData';

type ViewMode = 'library' | 'setBuilder' | 'graph';
type SortMode = 'bpm-asc' | 'bpm-desc' | 'energy-desc';
type SetMode = 'Warm-up' | 'Peak-time' | 'Open-format';

export type ExportFormat = 'rekordbox' | 'serato' | 'traktor';

export interface ExportResult {
  playlistPath: string;
  analysisJson: string;
  cuesCSV: string;
  bundlePath: string;
  vendorExports: string[];
}

interface AppState {
  // Data
  tracks: Track[];
  trackMap: Record<string, Track>;
  currentSetPlan: SetPlan | null;

  // Selection & UI
  selectedId: string | null;
  viewMode: ViewMode;
  query: string;
  onlyReview: boolean;
  onlyAnalyzed: boolean;
  highEnergyOnly: boolean;
  sortMode: SortMode;
  playheadPosition: number;
  isPlaying: boolean;
  chartMode: 'bpm' | 'key';

  // Loading & Error
  isLoading: boolean;
  error: string | null;
  apiAvailable: boolean;

  // Export
  isExporting: boolean;
  exportResult: ExportResult | null;
  exportError: string | null;

  // Computed
  filteredTracks: () => Track[];
  selectedTrack: () => Track | undefined;
  stats: () => {
    analyzed: number;
    pending: number;
    failed: number;
    avgBpm: number;
    avgEnergy: number;
    keyCount: number;
    avgEdgeScore: number;
    totalTracks: number;
  };

  // Actions
  setViewMode: (mode: ViewMode) => void;
  setQuery: (query: string) => void;
  setOnlyReview: (value: boolean) => void;
  setOnlyAnalyzed: (value: boolean) => void;
  setHighEnergyOnly: (value: boolean) => void;
  setSortMode: (mode: SortMode) => void;
  setSelectedId: (id: string | null) => void;
  setPlayheadPosition: (pos: number) => void;
  setIsPlaying: (playing: boolean) => void;
  setChartMode: (mode: 'bpm' | 'key') => void;

  // API Actions
  fetchTracks: () => Promise<void>;
  fetchTrackDetail: (id: string) => Promise<void>;
  proposeSet: (trackIds: string[], mode: SetMode) => Promise<void>;
  exportSet: (trackIds: string[], playlistName: string, formats: ExportFormat[]) => Promise<ExportResult | null>;
  checkApiHealth: () => Promise<void>;
  useMockData: () => void;
  clearExportResult: () => void;
}

// Generate waveform from energy level (used when API doesn't return waveform)
function generateWaveform(energy: number, bars: number = 64): number[] {
  const waveform: number[] = [];
  for (let i = 0; i < bars; i++) {
    const position = i / bars;
    let base = energy * 0.6;
    if (position < 0.1) base *= position / 0.1;
    if (position > 0.85) base *= (1 - position) / 0.15;
    const variation = Math.sin(i * 0.5) * 1.5 + Math.random() * 2;
    waveform.push(Math.max(1, Math.min(10, base + variation)));
  }
  return waveform;
}

// Convert API track to frontend Track type
function apiToTrack(apiTrack: api.TrackSummaryResponse): Track {
  return {
    id: apiTrack.content_hash,
    title: apiTrack.title || apiTrack.path.split('/').pop() || 'Unknown',
    artist: apiTrack.artist || 'Unknown Artist',
    bpm: apiTrack.bpm || 120,
    key: apiTrack.key || '8A',
    energy: apiTrack.energy || 5,
    status: (apiTrack.status === 'analyzed' ? 'analyzed' : apiTrack.status === 'failed' ? 'failed' : 'pending') as Track['status'],
    needsReview: apiTrack.needs_review,
    path: apiTrack.path,
    cues: [],
    sections: [],
    transitionWindows: [],
    waveformSummary: generateWaveform(apiTrack.energy || 5),
  };
}

// Convert API analysis to full Track type
function analysisToTrack(analysis: api.TrackAnalysisResponse): Track {
  const cues = (analysis.cues || []).map((c) => ({
    beat: c.beat,
    label: c.label,
    type: c.type,
    color: c.color ? `rgb(${c.color.r},${c.color.g},${c.color.b})` : undefined,
  }));

  const sections = (analysis.sections || []).map((s) => ({
    start: s.start_beat,
    end: s.end_beat,
    label: s.tag,
  }));

  const transitionWindows = (analysis.transition_windows || []).map((tw) => ({
    start: tw.start_beat,
    end: tw.end_beat,
    label: tw.tag,
  }));

  // Extract BPM from tempo map
  let bpm = 128;
  if (analysis.beatgrid?.tempo_map?.length > 0) {
    bpm = analysis.beatgrid.tempo_map[0].bpm;
  }

  return {
    id: analysis.id.content_hash,
    title: analysis.title || analysis.id.path.split('/').pop() || 'Unknown',
    artist: analysis.artist || 'Unknown Artist',
    bpm,
    key: analysis.key?.value || '8A',
    energy: analysis.energy_global || 5,
    status: 'analyzed',
    needsReview: false,
    path: analysis.id.path,
    cues,
    sections,
    transitionWindows,
    waveformSummary: analysis.waveform_summary || generateWaveform(analysis.energy_global || 5),
  };
}

// Convert API explanation to SetEdge
function explanationToEdge(expl: api.TransitionExplanation): SetEdge {
  return {
    from: expl.from.content_hash,
    to: expl.to.content_hash,
    score: expl.score,
    tempoDelta: expl.tempo_delta,
    energyDelta: expl.energy_delta,
    keyRelation: expl.key_relation,
    window: expl.window_tag,
    reason: expl.reasons?.join('; ') || '',
  };
}

export const useStore = create<AppState>((set, get) => ({
  // Initial state
  tracks: [],
  trackMap: {},
  currentSetPlan: null,
  selectedId: null,
  viewMode: 'library',
  query: '',
  onlyReview: false,
  onlyAnalyzed: false,
  highEnergyOnly: false,
  sortMode: 'energy-desc',
  playheadPosition: 0.3,
  isPlaying: false,
  chartMode: 'bpm',
  isLoading: true,
  error: null,
  apiAvailable: false,
  isExporting: false,
  exportResult: null,
  exportError: null,

  // Computed getters
  filteredTracks: () => {
    const state = get();
    const base = state.tracks.filter((track) => {
      const matchesQuery =
        track.title.toLowerCase().includes(state.query.toLowerCase()) ||
        track.artist.toLowerCase().includes(state.query.toLowerCase()) ||
        track.key.toLowerCase().includes(state.query.toLowerCase());
      const matchesReview = !state.onlyReview || track.needsReview;
      const matchesAnalyzed = !state.onlyAnalyzed || track.status === 'analyzed';
      const matchesEnergy = !state.highEnergyOnly || track.energy >= 7;
      return matchesQuery && matchesReview && matchesAnalyzed && matchesEnergy;
    });
    return base.sort((a, b) => {
      switch (state.sortMode) {
        case 'bpm-asc':
          return a.bpm - b.bpm;
        case 'bpm-desc':
          return b.bpm - a.bpm;
        case 'energy-desc':
        default:
          return b.energy - a.energy || b.bpm - a.bpm;
      }
    });
  },

  selectedTrack: () => {
    const state = get();
    if (state.selectedId && state.trackMap[state.selectedId]) {
      return state.trackMap[state.selectedId];
    }
    const filtered = state.filteredTracks();
    return filtered[0];
  },

  stats: () => {
    const state = get();
    const tracks = state.tracks;
    const analyzed = tracks.filter((t) => t.status === 'analyzed').length;
    const pending = tracks.filter((t) => t.status === 'pending').length;
    const failed = tracks.filter((t) => t.status === 'failed').length;
    const avgBpm = tracks.reduce((acc, t) => acc + t.bpm, 0) / Math.max(tracks.length, 1);
    const avgEnergy = tracks.reduce((acc, t) => acc + t.energy, 0) / Math.max(tracks.length, 1);
    const keys = new Set(tracks.map((t) => t.key));
    const edges = state.currentSetPlan?.edges || [];
    const avgEdge = edges.reduce((acc, e) => acc + e.score, 0) / Math.max(edges.length, 1);
    return {
      analyzed,
      pending,
      failed,
      avgBpm: Math.round(avgBpm * 10) / 10,
      avgEnergy: Math.round(avgEnergy * 10) / 10,
      keyCount: keys.size,
      avgEdgeScore: Math.round(avgEdge * 10) / 10 || 0,
      totalTracks: tracks.length,
    };
  },

  // Simple setters
  setViewMode: (mode) => set({ viewMode: mode }),
  setQuery: (query) => set({ query }),
  setOnlyReview: (value) => set({ onlyReview: value }),
  setOnlyAnalyzed: (value) => set({ onlyAnalyzed: value }),
  setHighEnergyOnly: (value) => set({ highEnergyOnly: value }),
  setSortMode: (mode) => set({ sortMode: mode }),
  setSelectedId: (id) => set({ selectedId: id }),
  setPlayheadPosition: (pos) => set({ playheadPosition: pos }),
  setIsPlaying: (playing) => set({ isPlaying: playing }),
  setChartMode: (mode) => set({ chartMode: mode }),

  // Use mock data fallback
  useMockData: () => {
    const trackMap: Record<string, Track> = {};
    for (const t of demoTracks) {
      trackMap[t.id] = t;
    }
    set({
      tracks: demoTracks,
      trackMap,
      currentSetPlan: demoSetPlan,
      selectedId: demoTracks[0]?.id || null,
      isLoading: false,
      error: null,
      apiAvailable: false,
    });
  },

  // API health check
  checkApiHealth: async () => {
    try {
      const response = await api.checkHealth();
      set({ apiAvailable: response.status === 'ok' });
    } catch {
      set({ apiAvailable: false });
    }
  },

  // Fetch all tracks from API
  fetchTracks: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await api.listTracks({ limit: 500 });
      const tracks = response.map(apiToTrack);
      const trackMap: Record<string, Track> = {};
      for (const t of tracks) {
        trackMap[t.id] = t;
      }
      set({
        tracks,
        trackMap,
        selectedId: tracks[0]?.id || null,
        isLoading: false,
        apiAvailable: true,
      });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to fetch tracks',
        isLoading: false,
      });
      // Fall back to mock data on API failure
      get().useMockData();
    }
  },

  // Fetch full track detail with analysis
  fetchTrackDetail: async (id: string) => {
    try {
      const analysis = await api.getTrack(id);
      const track = analysisToTrack(analysis);
      set((state) => ({
        trackMap: { ...state.trackMap, [id]: track },
        tracks: state.tracks.map((t) => (t.id === id ? track : t)),
      }));
    } catch (err) {
      console.warn('Failed to fetch track detail:', err);
    }
  },

  // Propose a DJ set
  proposeSet: async (trackIds: string[], mode: SetMode) => {
    if (trackIds.length < 2) return;

    const apiMode = mode === 'Warm-up' ? 'WARM_UP' : mode === 'Open-format' ? 'OPEN_FORMAT' : 'PEAK_TIME';

    try {
      const response = await api.proposeSet({
        track_ids: trackIds,
        mode: apiMode,
      });

      const order = response.order.map((o) => o.content_hash);
      const edges = response.explanations.map(explanationToEdge);

      set({
        currentSetPlan: { mode, order, edges },
      });
    } catch (err) {
      console.warn('Failed to propose set:', err);
    }
  },

  // Export a DJ set
  exportSet: async (trackIds: string[], playlistName: string, formats: ExportFormat[]) => {
    if (trackIds.length === 0) return null;

    set({ isExporting: true, exportError: null, exportResult: null });

    try {
      const response = await api.exportSet({
        track_ids: trackIds,
        playlist_name: playlistName,
        formats,
      });

      const result: ExportResult = {
        playlistPath: response.playlist_path,
        analysisJson: response.analysis_json,
        cuesCSV: response.cues_csv,
        bundlePath: response.bundle_path,
        vendorExports: response.vendor_exports,
      };

      set({ exportResult: result, isExporting: false });
      return result;
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Export failed';
      set({ exportError: errorMsg, isExporting: false });
      return null;
    }
  },

  // Clear export result
  clearExportResult: () => {
    set({ exportResult: null, exportError: null });
  },
}));
