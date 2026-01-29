import { type SetPlan, type Track } from '../types';

type Props = {
  plan: SetPlan;
  tracks: Record<string, Track>;
};

export function SetBuilder({ plan, tracks }: Props) {
  return (
    <div className="card">
      <header className="set-header">
        <div>
          <h3>Set builder</h3>
          <p className="muted">Mode: {plan.mode}</p>
        </div>
        <span className="pill pill-primary">{plan.order.length} tracks</span>
      </header>

      <div className="set-builder">
        {plan.order.map((id, idx) => {
          const track = tracks[id];
          const edge = plan.edges.find((e) => e.to === id && e.from === plan.order[idx - 1]);
          return (
            <div key={id}>
              <div className="set-track">
                <div className="set-track-number">{idx + 1}</div>
                <div>
                  <div className="track-title">{track?.title ?? 'Unknown'}</div>
                  <div className="track-artist">{track?.artist}</div>
                  <div className="track-meta">
                    <span>{track?.bpm} BPM</span>
                    <span>{track?.key}</span>
                    <span>Energy {track?.energy}</span>
                  </div>
                </div>
                {edge && <span className="pill pill-secondary">Score {edge.score.toFixed(1)}</span>}
              </div>
              {edge && (
                <div className="transition-reason">
                  {edge.reason} • {edge.keyRelation} • Δ{edge.tempoDelta} BPM • {edge.window}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
