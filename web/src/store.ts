import { create } from 'zustand';
import type { Track, SetPlan, SetEdge } from './types';
import * as api from './api';
import { demoTracks, demoSetPlan } from './mockData';

// Check localStorage for onboarding state
const ONBOARDING_KEY = 'algiers-onboarding-complete';
function getOnboardingComplete(): boolean {
  try {
    return localStorage.getItem(ONBOARDING_KEY) === 'true';
  } catch {
    return false;
  }
}
function setOnboardingComplete(value: boolean): void {
  try {
    localStorage.setItem(ONBOARDING_KEY, value ? 'true' : 'false');
  } catch {
    // localStorage not available
  }
}

type ViewMode = 'library' | 'setBuilder' | 'graph' | 'settings' | 'training';
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

export interface MLSettings {
  openl3Enabled: boolean;
  soundAnalysisEnabled: boolean;
  customModelEnabled: boolean;
  minSimilarityThreshold: number;
  showExplanations: boolean;
}

export interface SimilarTrack {
  trackId: number;
  contentHash: string;
  title: string;
  artist: string;
  score: number;
  explanation: string;
  vibeMatch: number;
  tempoMatch: number;
  keyMatch: number;
  energyMatch: number;
  bpmDelta: number;
  keyRelation: string;
  energyDelta: number;
}

interface AppState {
  // Data
  tracks: Track[];
  trackMap: Record<string, Track>;
  currentSetPlan: SetPlan | null;

  // Selection & UI
  selectedId: string | null;
  batchSelectedIds: Set<string>;
  batchMode: boolean;
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

  // Onboarding
  hasCompletedOnboarding: boolean;

  // Export
  isExporting: boolean;
  exportResult: ExportResult | null;
  exportError: string | null;

  // ML Settings
  mlSettings: MLSettings;
  mlSettingsLoading: boolean;

  // Similar Tracks
  similarTracks: SimilarTrack[];
  similarTracksLoading: boolean;
  similarTracksError: string | null;

  // Batch Operations
  isAnalyzing: boolean;
  analyzeProgress: { current: number; total: number; analyzing: string[] };

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
  useDemoData: () => void;
  clearExportResult: () => void;

  // Onboarding Actions
  scanLibrary: (roots: string[]) => Promise<{ processed: number; total: number; newTracks: string[] }>;
  completeOnboarding: () => void;

  // ML Actions
  fetchMLSettings: () => Promise<void>;
  updateMLSettings: (settings: Partial<MLSettings>) => Promise<void>;
  fetchSimilarTracks: (trackId: string, limit?: number) => Promise<void>;
  clearSimilarTracks: () => void;

  // Batch Actions
  setBatchMode: (enabled: boolean) => void;
  toggleBatchSelection: (id: string) => void;
  selectAllTracks: () => void;
  selectNoneTracks: () => void;
  analyzeBatchSelected: () => Promise<void>;
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

// Build trackMap from tracks array
function buildTrackMap(tracks: Track[]): Record<string, Track> {
  const map: Record<string, Track> = {};
  for (const t of tracks) {
    map[t.id] = t;
  }
  return map;
}

// Initialize based on onboarding state
const hasOnboarded = getOnboardingComplete();
const initialTracks = hasOnboarded ? demoTracks.slice() : [];
const initialTrackMap = buildTrackMap(initialTracks);
const initialSetPlan = hasOnboarded ? JSON.parse(JSON.stringify(demoSetPlan)) as typeof demoSetPlan : null;

export const useStore = create<AppState>((set, get) => ({
  // Initial state - empty until onboarding or API load
  tracks: initialTracks,
  trackMap: initialTrackMap,
  currentSetPlan: initialSetPlan,
  selectedId: initialTracks[0]?.id || null,
  batchSelectedIds: new Set<string>(),
  batchMode: false,
  viewMode: 'library',
  query: '',
  onlyReview: false,
  onlyAnalyzed: false,
  highEnergyOnly: false,
  sortMode: 'energy-desc',
  playheadPosition: 0.3,
  isPlaying: false,
  chartMode: 'bpm',
  isLoading: false,
  error: null,
  apiAvailable: false,
  hasCompletedOnboarding: hasOnboarded,
  isExporting: false,
  exportResult: null,
  exportError: null,

  // ML Settings
  mlSettings: {
    openl3Enabled: true,
    soundAnalysisEnabled: false,
    customModelEnabled: false,
    minSimilarityThreshold: 0.5,
    showExplanations: true,
  },
  mlSettingsLoading: false,

  // Similar Tracks
  similarTracks: [],
  similarTracksLoading: false,
  similarTracksError: null,

  // Batch Operations
  isAnalyzing: false,
  analyzeProgress: { current: 0, total: 0, analyzing: [] },

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

  // Use demo data
  useDemoData: () => {
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

  // Scan music library folders
  scanLibrary: async (roots: string[]) => {
    set({ isLoading: true, error: null });
    try {
      const result = await api.scanLibrary(roots, false);
      // After scanning, fetch the tracks
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
      return {
        processed: result.processed,
        total: result.total,
        newTracks: result.new_tracks,
      };
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Scan failed',
        isLoading: false,
      });
      throw err;
    }
  },

  // Complete onboarding
  completeOnboarding: () => {
    setOnboardingComplete(true);
    set({ hasCompletedOnboarding: true });
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
    const currentTracks = get().tracks;

    // Skip API fetch if we already have demo data - just do health check
    if (currentTracks.length > 0) {
      try {
        await api.listTracks({ limit: 1 });
        set({ apiAvailable: true });
      } catch {
        set({ apiAvailable: false });
      }
      return;
    }

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
      // Fall back to demo data on API failure
      get().useDemoData();
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

  // Fetch ML settings
  fetchMLSettings: async () => {
    set({ mlSettingsLoading: true });
    try {
      const response = await api.getMLSettings();
      set({
        mlSettings: {
          openl3Enabled: response.openl3_enabled,
          soundAnalysisEnabled: response.sound_analysis_enabled,
          customModelEnabled: response.custom_model_enabled,
          minSimilarityThreshold: response.min_similarity_threshold,
          showExplanations: response.show_explanations,
        },
        mlSettingsLoading: false,
      });
    } catch (err) {
      console.warn('Failed to fetch ML settings:', err);
      set({ mlSettingsLoading: false });
    }
  },

  // Update ML settings
  updateMLSettings: async (settings: Partial<MLSettings>) => {
    set({ mlSettingsLoading: true });
    try {
      const response = await api.updateMLSettings({
        openl3_enabled: settings.openl3Enabled,
        sound_analysis_enabled: settings.soundAnalysisEnabled,
        custom_model_enabled: settings.customModelEnabled,
        min_similarity_threshold: settings.minSimilarityThreshold,
        show_explanations: settings.showExplanations,
      });
      set({
        mlSettings: {
          openl3Enabled: response.openl3_enabled,
          soundAnalysisEnabled: response.sound_analysis_enabled,
          customModelEnabled: response.custom_model_enabled,
          minSimilarityThreshold: response.min_similarity_threshold,
          showExplanations: response.show_explanations,
        },
        mlSettingsLoading: false,
      });
    } catch (err) {
      console.warn('Failed to update ML settings:', err);
      set({ mlSettingsLoading: false });
    }
  },

  // Fetch similar tracks
  fetchSimilarTracks: async (trackId: string, limit = 10) => {
    set({ similarTracksLoading: true, similarTracksError: null });
    try {
      const response = await api.getSimilarTracks(trackId, limit);
      const similar = response.similar.map((s) => ({
        trackId: s.track_id,
        contentHash: s.content_hash,
        title: s.title,
        artist: s.artist,
        score: s.score,
        explanation: s.explanation,
        vibeMatch: s.vibe_match,
        tempoMatch: s.tempo_match,
        keyMatch: s.key_match,
        energyMatch: s.energy_match,
        bpmDelta: s.bpm_delta,
        keyRelation: s.key_relation,
        energyDelta: s.energy_delta,
      }));
      set({ similarTracks: similar, similarTracksLoading: false });
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : 'Failed to fetch similar tracks';
      set({ similarTracksError: errorMsg, similarTracksLoading: false, similarTracks: [] });
    }
  },

  // Clear similar tracks
  clearSimilarTracks: () => {
    set({ similarTracks: [], similarTracksError: null });
  },

  // Batch operations
  setBatchMode: (enabled: boolean) => {
    set({
      batchMode: enabled,
      batchSelectedIds: enabled ? get().batchSelectedIds : new Set<string>(),
    });
  },

  toggleBatchSelection: (id: string) => {
    const current = get().batchSelectedIds;
    const next = new Set(current);
    if (next.has(id)) {
      next.delete(id);
    } else {
      next.add(id);
    }
    set({ batchSelectedIds: next });
  },

  selectAllTracks: () => {
    const filtered = get().filteredTracks();
    set({ batchSelectedIds: new Set(filtered.map((t) => t.id)) });
  },

  selectNoneTracks: () => {
    set({ batchSelectedIds: new Set<string>() });
  },

  analyzeBatchSelected: async () => {
    const { batchSelectedIds, apiAvailable, tracks } = get();
    if (!apiAvailable || batchSelectedIds.size === 0) return;

    const selectedIds = Array.from(batchSelectedIds);
    set({
      isAnalyzing: true,
      analyzeProgress: { current: 0, total: selectedIds.length, analyzing: selectedIds },
    });

    try {
      const result = await api.analyzeTracks({ track_ids: selectedIds });

      // Refresh track list after analysis
      const response = await api.listTracks({ limit: 500 });
      const updatedTracks = response.map((t) => {
        // Find existing track to preserve extended data
        const existing = tracks.find((e) => e.id === t.content_hash);
        return existing ? { ...existing, status: t.status as Track['status'] } : {
          id: t.content_hash,
          title: t.title || t.path.split('/').pop() || 'Unknown',
          artist: t.artist || 'Unknown Artist',
          bpm: t.bpm || 120,
          key: t.key || '8A',
          energy: t.energy || 5,
          status: (t.status === 'analyzed' ? 'analyzed' : t.status === 'failed' ? 'failed' : 'pending') as Track['status'],
          needsReview: t.needs_review,
          path: t.path,
          cues: [],
          sections: [],
          transitionWindows: [],
          waveformSummary: [],
        };
      });

      const trackMap: Record<string, Track> = {};
      for (const t of updatedTracks) {
        trackMap[t.id] = t;
      }

      set({
        tracks: updatedTracks,
        trackMap,
        isAnalyzing: false,
        analyzeProgress: { current: result.analyzed.length, total: selectedIds.length, analyzing: [] },
        batchSelectedIds: new Set<string>(),
        batchMode: false,
      });
    } catch (err) {
      set({
        isAnalyzing: false,
        error: err instanceof Error ? err.message : 'Analysis failed',
      });
    }
  },
}));
