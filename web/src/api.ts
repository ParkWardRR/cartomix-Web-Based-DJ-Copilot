/**
 * API client for Algiers HTTP REST endpoints.
 * All endpoints are proxied through Vite dev server.
 */

const API_BASE = '/api';

export type ApiError = {
  error: string;
};

// Track list response from GET /api/tracks
export type TrackSummaryResponse = {
  id?: number;
  content_hash: string;
  path: string;
  title: string;
  artist: string;
  bpm: number;
  key: string;
  energy: number;
  cue_count: number;
  status: string;
  needs_review: boolean;
  analyzed_at?: string;
};

// Full track analysis response from GET /api/tracks/{id}
export type TrackAnalysisResponse = {
  id: {
    content_hash: string;
    path: string;
  };
  title: string;
  artist: string;
  key: {
    value: string;
    format: string;
    confidence: number;
  };
  beatgrid: {
    beats: Array<{
      index: number;
      time: string;
      is_downbeat: boolean;
    }>;
    tempo_map: Array<{
      beat_index: number;
      bpm: number;
    }>;
    confidence: number;
  };
  cues: Array<{
    beat: number;
    label: string;
    type: string;
    color?: {
      r: number;
      g: number;
      b: number;
    };
  }>;
  sections: Array<{
    start_beat: number;
    end_beat: number;
    tag: string;
    confidence: number;
  }>;
  transition_windows: Array<{
    start_beat: number;
    end_beat: number;
    tag: string;
    confidence: number;
  }>;
  energy_global: number;
  energy_curve?: number[];
  waveform_summary?: number[];
};

// Scan request/response
export type ScanRequest = {
  roots: string[];
  force_rescan: boolean;
};

export type ScanResponse = {
  processed: number;
  total: number;
  new_tracks: string[];
};

// Analyze request/response
export type AnalyzeRequest = {
  paths?: string[];
  track_ids?: string[];
  force?: boolean;
};

export type AnalyzeResponse = {
  analyzed: string[];
  skipped: string[];
  errors: string[];
};

// Set proposal request/response
export type ProposeSetRequest = {
  track_ids: string[];
  mode: 'WARM_UP' | 'PEAK_TIME' | 'OPEN_FORMAT';
  allow_key_jumps?: boolean;
  max_bpm_step?: number;
  must_play?: string[];
  ban?: string[];
};

export type TransitionExplanation = {
  from: { content_hash: string; path: string };
  to: { content_hash: string; path: string };
  score: number;
  tempo_delta: number;
  energy_delta: number;
  key_relation: string;
  window_tag: string;
  reasons: string[];
};

export type ProposeSetResponse = {
  order: Array<{ content_hash: string; path: string }>;
  explanations: TransitionExplanation[];
};

// Export request/response
export type ExportRequest = {
  track_ids: string[];
  playlist_name?: string;
  output_dir?: string;
  formats?: string[];
};

export type ExportResponse = {
  playlist_path: string;
  analysis_json: string;
  cues_csv: string;
  bundle_path: string;
  vendor_exports: string[];
};

async function fetchJson<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error((data as ApiError).error || `HTTP ${response.status}`);
  }

  return data as T;
}

/**
 * Check API health.
 */
export async function checkHealth(): Promise<{ status: string }> {
  return fetchJson(`${API_BASE}/health`);
}

/**
 * List tracks with optional filtering.
 */
export async function listTracks(options?: {
  query?: string;
  needsReview?: boolean;
  limit?: number;
}): Promise<TrackSummaryResponse[]> {
  const params = new URLSearchParams();
  if (options?.query) params.set('q', options.query);
  if (options?.needsReview) params.set('needs_review', 'true');
  if (options?.limit) params.set('limit', options.limit.toString());

  const queryString = params.toString();
  const url = queryString ? `${API_BASE}/tracks?${queryString}` : `${API_BASE}/tracks`;
  return fetchJson(url);
}

/**
 * Get full track analysis by content hash.
 */
export async function getTrack(id: string): Promise<TrackAnalysisResponse> {
  return fetchJson(`${API_BASE}/tracks/${encodeURIComponent(id)}`);
}

/**
 * Scan music library directories.
 */
export async function scanLibrary(roots: string[], forceRescan = false): Promise<ScanResponse> {
  return fetchJson(`${API_BASE}/scan`, {
    method: 'POST',
    body: JSON.stringify({ roots, force_rescan: forceRescan }),
  });
}

/**
 * Analyze tracks by path or ID.
 */
export async function analyzeTracks(
  request: AnalyzeRequest
): Promise<AnalyzeResponse> {
  return fetchJson(`${API_BASE}/analyze`, {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

/**
 * Propose a DJ set ordering.
 */
export async function proposeSet(request: ProposeSetRequest): Promise<ProposeSetResponse> {
  return fetchJson(`${API_BASE}/set/propose`, {
    method: 'POST',
    body: JSON.stringify(request),
  });
}

/**
 * Export a set in various formats.
 */
export async function exportSet(request: ExportRequest): Promise<ExportResponse> {
  return fetchJson(`${API_BASE}/export`, {
    method: 'POST',
    body: JSON.stringify(request),
  });
}
