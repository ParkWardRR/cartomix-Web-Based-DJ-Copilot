import { type SetPlan, type Track } from '../types';

type Props = {
  plan: SetPlan;
  tracks: Record<string, Track>;
  compact?: boolean;
};

export function SetBuilder({ plan, tracks, compact }: Props) {
  const displayOrder = compact ? plan.order.slice(0, 6) : plan.order;

  return (
    <div className={`set-builder ${compact ? 'compact' : ''}`}>
      {displayOrder.map((id, idx) => {
        const track = tracks[id];
        const edge = plan.edges.find((e) => e.to === id && e.from === plan.order[idx - 1]);
        return (
          <div key={id}>
            <div className="set-track">
              <div className="set-track-number">{idx + 1}</div>
              <div className="set-track-info">
                <div className="set-track-title">{track?.title ?? 'Unknown'}</div>
              </div>
              <div className="track-meta">
                <span>{track?.bpm}</span>
                <span>{track?.key}</span>
              </div>
              {edge && !compact && <span className="pill pill-secondary">{edge.score.toFixed(1)}</span>}
            </div>
            {edge && !compact && (
              <div className="transition-reason">
                {edge.keyRelation} • Δ{edge.tempoDelta} BPM
              </div>
            )}
          </div>
        );
      })}
      {compact && plan.order.length > 6 && (
        <div className="set-track" style={{ opacity: 0.6, justifyContent: 'center' }}>
          <span className="muted">+{plan.order.length - 6} more tracks</span>
        </div>
      )}
    </div>
  );
}
