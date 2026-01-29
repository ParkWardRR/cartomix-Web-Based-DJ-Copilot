import { type SetPlan, type Track } from './types';

// Generate realistic waveform data with EDM-style dynamics
function generateWaveform(energy: number, bars: number = 64, style: 'techno' | 'house' | 'progressive' | 'ambient' = 'techno'): number[] {
  const waveform: number[] = [];
  // Seed random for consistent waveforms
  let seed = energy * 1000;
  const seededRandom = () => {
    seed = (seed * 9301 + 49297) % 233280;
    return seed / 233280;
  };

  for (let i = 0; i < bars; i++) {
    const position = i / bars;
    let base = energy * 0.6;

    // Style-specific dynamics
    if (style === 'techno') {
      // Hard drops, punchy
      if (position < 0.08) base *= position / 0.08;
      if (position > 0.9) base *= (1 - position) / 0.1;
      if (position > 0.4 && position < 0.5) base *= 0.6; // breakdown
    } else if (style === 'progressive') {
      // Long builds
      if (position < 0.15) base *= position / 0.15;
      if (position > 0.85) base *= (1 - position) / 0.15;
      if (position > 0.3 && position < 0.45) base *= 0.4 + position; // slow build
    } else if (style === 'house') {
      // Groovy, steady
      if (position < 0.1) base *= position / 0.1;
      if (position > 0.88) base *= (1 - position) / 0.12;
    } else {
      // Ambient - flowing
      base *= 0.5 + Math.sin(position * Math.PI) * 0.5;
    }

    // Add variation
    const variation = Math.sin(i * 0.7) * 1.2 + seededRandom() * 2.5;
    waveform.push(Math.max(1, Math.min(10, base + variation)));
  }
  return waveform;
}

export const demoTracks: Track[] = [
  {
    id: 'hash-berlin-sunrise',
    title: 'Berghain Sunrise',
    artist: 'Amelie Lens',
    bpm: 136,
    key: '8A',
    energy: 8,
    status: 'analyzed',
    needsReview: false,
    path: '/music/Amelie Lens/Berghain Sunrise.wav',
    cues: [
      { beat: 0, label: 'Load', type: 'Load' },
      { beat: 64, label: 'Drop', type: 'Drop' },
      { beat: 256, label: 'Break', type: 'Breakdown' },
      { beat: 320, label: 'Peak', type: 'Drop' },
      { beat: 448, label: 'Outro', type: 'OutroStart' },
    ],
    sections: [
      { start: 0, end: 64, label: 'Intro' },
      { start: 64, end: 256, label: 'Drop' },
      { start: 256, end: 320, label: 'Breakdown' },
      { start: 320, end: 448, label: 'Peak' },
      { start: 448, end: 512, label: 'Outro' },
    ],
    transitionWindows: [
      { start: 0, end: 32, label: 'intro_mix' },
      { start: 480, end: 512, label: 'outro_mix' },
    ],
    waveformSummary: generateWaveform(8, 64, 'techno'),
  },
  {
    id: 'hash-midnight-arc',
    title: 'Oxia - Domino (Matador Remix)',
    artist: 'Matador',
    bpm: 128,
    key: '9A',
    energy: 7,
    status: 'analyzed',
    needsReview: false,
    path: '/music/Matador/Domino Remix.flac',
    cues: [
      { beat: 0, label: 'Load', type: 'Load' },
      { beat: 64, label: 'Keys In', type: 'FirstDownbeat' },
      { beat: 128, label: 'Break', type: 'Breakdown' },
      { beat: 192, label: 'Build', type: 'Build' },
      { beat: 256, label: 'Main Drop', type: 'Drop' },
    ],
    sections: [
      { start: 0, end: 64, label: 'Intro' },
      { start: 64, end: 128, label: 'Build' },
      { start: 128, end: 192, label: 'Breakdown' },
      { start: 192, end: 256, label: 'Tension' },
      { start: 256, end: 384, label: 'Peak' },
      { start: 384, end: 448, label: 'Outro' },
    ],
    transitionWindows: [
      { start: 16, end: 48, label: 'intro_mix' },
      { start: 400, end: 448, label: 'outro_mix' },
    ],
    waveformSummary: generateWaveform(7, 64, 'progressive'),
  },
  {
    id: 'hash-neon-bridge',
    title: 'Neon Bridge',
    artist: 'MonoKite',
    bpm: 124,
    key: '7B',
    energy: 5,
    status: 'analyzed',
    needsReview: true,
    path: '/music/MonoKite/Neon Bridge.mp3',
    cues: [
      { beat: 0, label: 'Load', type: 'Load' },
      { beat: 48, label: 'First Downbeat', type: 'FirstDownbeat' },
      { beat: 320, label: 'Outro', type: 'OutroStart' },
    ],
    sections: [
      { start: 0, end: 48, label: 'Intro' },
      { start: 48, end: 288, label: 'Body' },
      { start: 288, end: 352, label: 'Outro' },
    ],
    transitionWindows: [
      { start: 0, end: 24, label: 'intro_mix' },
      { start: 304, end: 336, label: 'outro_mix' },
    ],
    waveformSummary: generateWaveform(5),
  },
  {
    id: 'hash-delta-peak',
    title: 'Enrico Sangiuliano - Biomorph',
    artist: 'Enrico Sangiuliano',
    bpm: 133,
    key: '10A',
    energy: 9,
    status: 'analyzed',
    needsReview: false,
    path: '/music/Enrico Sangiuliano/Biomorph.aiff',
    cues: [
      { beat: 0, label: 'Load', type: 'Load' },
      { beat: 32, label: 'Kick', type: 'FirstDownbeat' },
      { beat: 96, label: 'Synth', type: 'Build' },
      { beat: 160, label: 'DROP', type: 'Drop' },
      { beat: 288, label: 'Break', type: 'Breakdown' },
      { beat: 352, label: 'Final', type: 'Drop' },
    ],
    sections: [
      { start: 0, end: 96, label: 'Intro' },
      { start: 96, end: 160, label: 'Build' },
      { start: 160, end: 288, label: 'Drop' },
      { start: 288, end: 352, label: 'Breakdown' },
      { start: 352, end: 448, label: 'Peak' },
      { start: 448, end: 512, label: 'Outro' },
    ],
    transitionWindows: [
      { start: 0, end: 32, label: 'intro_mix' },
      { start: 480, end: 512, label: 'outro_mix' },
    ],
    waveformSummary: generateWaveform(9, 64, 'techno'),
  },
  {
    id: 'hash-cascade-flow',
    title: 'Cascade Flow',
    artist: 'Synthetic Dreams',
    bpm: 122,
    key: '6A',
    energy: 4,
    status: 'analyzed',
    needsReview: false,
    path: '/music/Synthetic Dreams/Cascade Flow.wav',
    cues: [
      { beat: 0, label: 'Load', type: 'Load' },
      { beat: 32, label: 'First Beat', type: 'FirstDownbeat' },
    ],
    sections: [
      { start: 0, end: 64, label: 'Intro' },
      { start: 64, end: 256, label: 'Body' },
      { start: 256, end: 320, label: 'Outro' },
    ],
    transitionWindows: [
      { start: 0, end: 48, label: 'intro_mix' },
      { start: 280, end: 320, label: 'outro_mix' },
    ],
    waveformSummary: generateWaveform(4),
  },
  {
    id: 'hash-pulse-drive',
    title: 'Pulse Drive',
    artist: 'Technocraft',
    bpm: 134,
    key: '11A',
    energy: 9,
    status: 'analyzed',
    needsReview: false,
    path: '/music/Technocraft/Pulse Drive.flac',
    cues: [
      { beat: 0, label: 'Load', type: 'Load' },
      { beat: 32, label: 'Kick In', type: 'FirstDownbeat' },
      { beat: 128, label: 'Drop', type: 'Drop' },
      { beat: 256, label: 'Build', type: 'Build' },
      { beat: 288, label: 'Peak', type: 'Drop' },
    ],
    sections: [
      { start: 0, end: 32, label: 'Intro' },
      { start: 32, end: 128, label: 'Build' },
      { start: 128, end: 256, label: 'Drop' },
      { start: 256, end: 320, label: 'Breakdown' },
      { start: 320, end: 384, label: 'Outro' },
    ],
    transitionWindows: [
      { start: 0, end: 24, label: 'intro_mix' },
      { start: 350, end: 384, label: 'outro_mix' },
    ],
    waveformSummary: generateWaveform(9),
  },
  {
    id: 'hash-ambient-drift',
    title: 'Ambient Drift',
    artist: 'Horizon',
    bpm: 118,
    key: '5B',
    energy: 3,
    status: 'analyzed',
    needsReview: false,
    path: '/music/Horizon/Ambient Drift.mp3',
    cues: [
      { beat: 0, label: 'Load', type: 'Load' },
      { beat: 64, label: 'Melody', type: 'FirstDownbeat' },
    ],
    sections: [
      { start: 0, end: 96, label: 'Intro' },
      { start: 96, end: 288, label: 'Body' },
      { start: 288, end: 384, label: 'Outro' },
    ],
    transitionWindows: [
      { start: 0, end: 64, label: 'intro_mix' },
      { start: 320, end: 384, label: 'outro_mix' },
    ],
    waveformSummary: generateWaveform(3),
  },
  {
    id: 'hash-chrome-echo',
    title: 'Chrome Echo',
    artist: 'Digital Mind',
    bpm: 128,
    key: '8A',
    energy: 7,
    status: 'analyzed',
    needsReview: true,
    path: '/music/Digital Mind/Chrome Echo.wav',
    cues: [
      { beat: 0, label: 'Load', type: 'Load' },
      { beat: 64, label: 'Drop', type: 'Drop' },
      { beat: 192, label: 'Break', type: 'Breakdown' },
      { beat: 256, label: 'Peak', type: 'Drop' },
    ],
    sections: [
      { start: 0, end: 64, label: 'Intro' },
      { start: 64, end: 192, label: 'Drop' },
      { start: 192, end: 256, label: 'Break' },
      { start: 256, end: 352, label: 'Outro' },
    ],
    transitionWindows: [
      { start: 16, end: 48, label: 'intro_mix' },
      { start: 300, end: 352, label: 'outro_mix' },
    ],
    waveformSummary: generateWaveform(7),
  },
  {
    id: 'hash-solar-wind',
    title: 'Solar Wind',
    artist: 'Cosmos',
    bpm: 126,
    key: '9B',
    energy: 6,
    status: 'analyzed',
    needsReview: false,
    path: '/music/Cosmos/Solar Wind.flac',
    cues: [
      { beat: 0, label: 'Load', type: 'Load' },
      { beat: 48, label: 'Synth', type: 'FirstDownbeat' },
      { beat: 128, label: 'Build', type: 'Build' },
      { beat: 160, label: 'Drop', type: 'Drop' },
    ],
    sections: [
      { start: 0, end: 48, label: 'Intro' },
      { start: 48, end: 160, label: 'Build' },
      { start: 160, end: 288, label: 'Drop' },
      { start: 288, end: 352, label: 'Outro' },
    ],
    transitionWindows: [
      { start: 0, end: 32, label: 'intro_mix' },
      { start: 310, end: 352, label: 'outro_mix' },
    ],
    waveformSummary: generateWaveform(6),
  },
  {
    id: 'hash-quantum-state',
    title: 'Quantum State',
    artist: 'Particle',
    bpm: 132,
    key: '10B',
    energy: 8,
    status: 'pending',
    needsReview: false,
    path: '/music/Particle/Quantum State.aiff',
    cues: [],
    sections: [{ start: 0, end: 64, label: 'Intro' }],
    transitionWindows: [],
    waveformSummary: generateWaveform(8),
  },
  {
    id: 'hash-deep-current',
    title: 'Deep Current',
    artist: 'Submerge',
    bpm: 120,
    key: '4A',
    energy: 5,
    status: 'analyzed',
    needsReview: false,
    path: '/music/Submerge/Deep Current.mp3',
    cues: [
      { beat: 0, label: 'Load', type: 'Load' },
      { beat: 64, label: 'Bass', type: 'FirstDownbeat' },
      { beat: 256, label: 'Outro', type: 'OutroStart' },
    ],
    sections: [
      { start: 0, end: 64, label: 'Intro' },
      { start: 64, end: 256, label: 'Body' },
      { start: 256, end: 320, label: 'Outro' },
    ],
    transitionWindows: [
      { start: 0, end: 48, label: 'intro_mix' },
      { start: 280, end: 320, label: 'outro_mix' },
    ],
    waveformSummary: generateWaveform(5),
  },
  {
    id: 'hash-neon-rush',
    title: 'Neon Rush',
    artist: 'Velocity',
    bpm: 140,
    key: '12A',
    energy: 9,
    status: 'analyzed',
    needsReview: false,
    path: '/music/Velocity/Neon Rush.wav',
    cues: [
      { beat: 0, label: 'Load', type: 'Load' },
      { beat: 16, label: 'Kick', type: 'FirstDownbeat' },
      { beat: 64, label: 'Drop', type: 'Drop' },
      { beat: 192, label: 'Build', type: 'Build' },
      { beat: 224, label: 'Peak', type: 'Drop' },
    ],
    sections: [
      { start: 0, end: 64, label: 'Intro' },
      { start: 64, end: 192, label: 'Drop' },
      { start: 192, end: 256, label: 'Build' },
      { start: 256, end: 320, label: 'Outro' },
    ],
    transitionWindows: [
      { start: 0, end: 16, label: 'intro_mix' },
      { start: 280, end: 320, label: 'outro_mix' },
    ],
    waveformSummary: generateWaveform(9),
  },
  {
    id: 'hash-twilight-zone',
    title: 'Twilight Zone',
    artist: 'Dusk',
    bpm: 124,
    key: '7A',
    energy: 6,
    status: 'analyzed',
    needsReview: false,
    path: '/music/Dusk/Twilight Zone.flac',
    cues: [
      { beat: 0, label: 'Load', type: 'Load' },
      { beat: 32, label: 'Pad', type: 'FirstDownbeat' },
      { beat: 96, label: 'Drop', type: 'Drop' },
    ],
    sections: [
      { start: 0, end: 96, label: 'Intro' },
      { start: 96, end: 256, label: 'Drop' },
      { start: 256, end: 320, label: 'Outro' },
    ],
    transitionWindows: [
      { start: 0, end: 64, label: 'intro_mix' },
      { start: 280, end: 320, label: 'outro_mix' },
    ],
    waveformSummary: generateWaveform(6),
  },
  {
    id: 'hash-acid-rain',
    title: 'Acid Rain',
    artist: 'TB-303',
    bpm: 135,
    key: '3A',
    energy: 8,
    status: 'analyzed',
    needsReview: true,
    path: '/music/TB-303/Acid Rain.wav',
    cues: [
      { beat: 0, label: 'Load', type: 'Load' },
      { beat: 32, label: '303', type: 'FirstDownbeat' },
      { beat: 128, label: 'Filter', type: 'Build' },
      { beat: 192, label: 'Peak', type: 'Drop' },
    ],
    sections: [
      { start: 0, end: 64, label: 'Intro' },
      { start: 64, end: 192, label: 'Build' },
      { start: 192, end: 288, label: 'Drop' },
      { start: 288, end: 352, label: 'Outro' },
    ],
    transitionWindows: [
      { start: 0, end: 32, label: 'intro_mix' },
      { start: 310, end: 352, label: 'outro_mix' },
    ],
    waveformSummary: generateWaveform(8),
  },
  {
    id: 'hash-stellar-drift',
    title: 'Stellar Drift',
    artist: 'Nebula',
    bpm: 116,
    key: '2B',
    energy: 3,
    status: 'pending',
    needsReview: false,
    path: '/music/Nebula/Stellar Drift.mp3',
    cues: [],
    sections: [{ start: 0, end: 64, label: 'Intro' }],
    transitionWindows: [],
    waveformSummary: generateWaveform(3),
  },
  {
    id: 'hash-machine-heart',
    title: 'Machine Heart',
    artist: 'Circuit',
    bpm: 128,
    key: '8B',
    energy: 7,
    status: 'analyzed',
    needsReview: false,
    path: '/music/Circuit/Machine Heart.flac',
    cues: [
      { beat: 0, label: 'Load', type: 'Load' },
      { beat: 64, label: 'Main', type: 'Drop' },
      { beat: 192, label: 'Break', type: 'Breakdown' },
      { beat: 256, label: 'Return', type: 'Drop' },
    ],
    sections: [
      { start: 0, end: 64, label: 'Intro' },
      { start: 64, end: 192, label: 'Drop' },
      { start: 192, end: 256, label: 'Break' },
      { start: 256, end: 320, label: 'Outro' },
    ],
    transitionWindows: [
      { start: 0, end: 32, label: 'intro_mix' },
      { start: 288, end: 320, label: 'outro_mix' },
    ],
    waveformSummary: generateWaveform(7),
  },
];

export const demoSetPlan: SetPlan = {
  mode: 'Peak-time',
  order: [
    'hash-ambient-drift',
    'hash-cascade-flow',
    'hash-neon-bridge',
    'hash-twilight-zone',
    'hash-berlin-sunrise',
    'hash-chrome-echo',
    'hash-midnight-arc',
    'hash-delta-peak',
    'hash-pulse-drive',
    'hash-neon-rush',
  ],
  edges: [
    {
      from: 'hash-ambient-drift',
      to: 'hash-cascade-flow',
      score: 8.2,
      tempoDelta: 4,
      energyDelta: 1,
      keyRelation: '5B → 6A (+1 Camelot)',
      window: 'outro_mix → intro_mix',
      reason: '🎵 Smooth opener: +4 BPM lift, energy building, harmonic movement up the wheel',
    },
    {
      from: 'hash-cascade-flow',
      to: 'hash-neon-bridge',
      score: 7.8,
      tempoDelta: 2,
      energyDelta: 1,
      keyRelation: '+1 Camelot',
      window: 'outro_mix → intro_mix',
      reason: 'Smooth progression: +2 BPM, +1 energy, compatible keys',
    },
    {
      from: 'hash-neon-bridge',
      to: 'hash-twilight-zone',
      score: 8.2,
      tempoDelta: 0,
      energyDelta: 1,
      keyRelation: 'Same key family',
      window: 'outro_mix → intro_mix',
      reason: 'Perfect tempo match, +1 energy, harmonic blend',
    },
    {
      from: 'hash-twilight-zone',
      to: 'hash-berlin-sunrise',
      score: 7.9,
      tempoDelta: 2,
      energyDelta: 0,
      keyRelation: '+1 Camelot',
      window: 'outro_mix → intro_mix',
      reason: 'Maintain energy, +2 BPM groove shift',
    },
    {
      from: 'hash-berlin-sunrise',
      to: 'hash-chrome-echo',
      score: 8.5,
      tempoDelta: 2,
      energyDelta: 1,
      keyRelation: 'Same key',
      window: 'outro_mix → intro_mix',
      reason: 'Same key 8A, +2 BPM, building energy',
    },
    {
      from: 'hash-chrome-echo',
      to: 'hash-midnight-arc',
      score: 9.1,
      tempoDelta: 0,
      energyDelta: 0,
      keyRelation: '8A → 9A (+1 Camelot)',
      window: 'outro_mix → intro_mix',
      reason: '🔥 Perfect match: same tempo 128 BPM, smooth harmonic lift to 9A, 87% vibe similarity',
    },
    {
      from: 'hash-midnight-arc',
      to: 'hash-delta-peak',
      score: 8.8,
      tempoDelta: 5,
      energyDelta: 2,
      keyRelation: '9A → 10A (+1 Camelot)',
      window: 'outro_mix → intro_mix',
      reason: '📈 Peak time ramp: +5 BPM to 133, energy 7→9, continuing Camelot climb',
    },
    {
      from: 'hash-delta-peak',
      to: 'hash-pulse-drive',
      score: 8.7,
      tempoDelta: 1,
      energyDelta: 0,
      keyRelation: '10A → 11A (+1 Camelot)',
      window: 'outro_mix → intro_mix',
      reason: '🚀 Hold the energy: near-perfect tempo at 134, sustain peak level 9',
    },
    {
      from: 'hash-pulse-drive',
      to: 'hash-neon-rush',
      score: 9.4,
      tempoDelta: 6,
      energyDelta: 0,
      keyRelation: '11A → 12A (→ 1A cycle)',
      window: 'outro_mix → intro_mix',
      reason: '💥 FINALE: Push to 140 BPM, max energy, complete the Camelot wheel!',
    },
  ],
};
