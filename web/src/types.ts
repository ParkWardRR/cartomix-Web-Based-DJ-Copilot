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
