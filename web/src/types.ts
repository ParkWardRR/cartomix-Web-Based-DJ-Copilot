export type Cue = {
  beat: number;
  label: string;
  type: string;
  color?: string;
};

export type Section = {
  start: number;
  end: number;
  label: string;
};

export type TransitionWindow = {
  label: string;
  start: number;
  end: number;
};

export type Track = {
  id: string;
  title: string;
  artist: string;
  bpm: number;
  key: string;
  energy: number;
  status: 'pending' | 'analyzed' | 'failed';
  needsReview?: boolean;
  cues: Cue[];
  sections: Section[];
  transitionWindows: TransitionWindow[];
  waveformSummary: number[];
  path: string;
};

export type SetEdge = {
  from: string;
  to: string;
  score: number;
  tempoDelta: number;
  energyDelta: number;
  keyRelation: string;
  window: string;
  reason: string;
};

export type SetPlan = {
  mode: 'Warm-up' | 'Peak-time' | 'Open-format';
  order: string[];
  edges: SetEdge[];
};

export type SetSession = {
  id: string;
  name: string;
  createdAt: string;
  updatedAt: string;
  plan: SetPlan;
  trackCount: number;
  totalDuration: number; // minutes
  avgBpm: number;
  notes?: string;
};

export type SetHistoryStats = {
  totalSessions: number;
  totalTracks: number;
  avgTracksPerSet: number;
  favoriteMode: 'Warm-up' | 'Peak-time' | 'Open-format';
  lastSessionDate: string | null;
};
