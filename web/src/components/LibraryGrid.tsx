import type { Track } from '../types';

type Props = {
  tracks: Track[];
  selectedId?: string;
  onSelect: (id: string) => void;
  batchMode?: boolean;
  batchSelectedIds?: Set<string>;
  onBatchToggle?: (id: string) => void;
};

export function LibraryGrid({
  tracks,
  selectedId,
  onSelect,
  batchMode = false,
  batchSelectedIds = new Set(),
  onBatchToggle,
}: Props) {
  if (!tracks.length) {
    return <div className="muted" style={{ padding: '1rem', textAlign: 'center' }}>No tracks match that filter.</div>;
  }

  const handleClick = (id: string, e: React.MouseEvent) => {
    if (batchMode && onBatchToggle) {
      e.preventDefault();
      onBatchToggle(id);
    } else {
      onSelect(id);
    }
  };

  return (
    <div className="track-grid">
      {tracks.map((track) => (
        <button
          key={track.id}
          className={`track-card ${selectedId === track.id ? 'active' : ''} ${batchMode && batchSelectedIds.has(track.id) ? 'batch-selected' : ''}`}
          onClick={(e) => handleClick(track.id, e)}
        >
          {batchMode && (
            <div className="batch-checkbox">
              <input
                type="checkbox"
                checked={batchSelectedIds.has(track.id)}
                onChange={() => onBatchToggle?.(track.id)}
                onClick={(e) => e.stopPropagation()}
              />
            </div>
          )}
          <div className="track-info">
            <div className="track-title">{track.title}</div>
            <div className="track-artist">{track.artist}</div>
          </div>
          <div className="track-meta">
            <span>{track.bpm}</span>
            <span>{track.key}</span>
            <span>E{track.energy}</span>
          </div>
          <div className="track-flags">
            <span className={`pill pill-${track.status}`}>{track.status === 'analyzed' ? '✓' : '◷'}</span>
            {track.needsReview && <span className="pill pill-warn">!</span>}
          </div>
        </button>
      ))}
    </div>
  );
}
