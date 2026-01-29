import type { SetEdge, Track } from '../types';

type Props = {
  from?: Track;
  to?: Track;
  edge?: SetEdge;
};

export function TransitionRehearsal({ from, to, edge }: Props) {
  return (
    <div className="card">
      <header className="set-header">
        <div>
          <h3>Transition rehearsal</h3>
          <p className="muted">Dual deck preview (static demo)</p>
        </div>
        {edge && <span className="pill pill-primary">Score {edge.score.toFixed(1)}</span>}
      </header>

      <div className="dual-deck">
        <Deck label="Deck A" track={from} side="A" />
        <Deck label="Deck B" track={to} side="B" />
      </div>
      {edge && (
        <div className="transition-reason">
          {edge.reason} — {edge.window}
        </div>
      )}
    </div>
  );
}

function Deck({ label, track, side }: { label: string; track?: Track; side: 'A' | 'B' }) {
  return (
    <div className="deck">
      <div className="deck-label">
        <span className="pill pill-secondary">{label}</span>
        <span className="pill pill-secondary">{side}</span>
      </div>
      {track ? (
        <>
          <div className="track-title">{track.title}</div>
          <div className="track-artist">{track.artist}</div>
          <div className="track-meta">
            <span>{track.bpm} BPM</span>
            <span>{track.key}</span>
            <span>Energy {track.energy}</span>
          </div>
          <div className="waveform mini">
            <div className="energy-arc">
              {track.waveformSummary.slice(0, 16).map((level, idx) => (
                <div key={idx} className="energy-bar" style={{ height: `${level * 6}%` }} />
              ))}
            </div>
          </div>
        </>
      ) : (
        <p className="muted">Select a track to rehearse.</p>
      )}
    </div>
  );
}
