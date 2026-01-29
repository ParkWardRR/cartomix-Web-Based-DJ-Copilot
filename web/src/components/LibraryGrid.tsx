import type { Track } from '../types';

type Props = {
  tracks: Track[];
  selectedId?: string;
  onSelect: (id: string) => void;
};

export function LibraryGrid({ tracks, selectedId, onSelect }: Props) {
  if (!tracks.length) {
    return <div className="muted" style={{ padding: '1rem', textAlign: 'center' }}>No tracks match that filter.</div>;
  }

  return (
    <div className="track-grid">
      {tracks.map((track) => (
        <button
          key={track.id}
          className={`track-card ${selectedId === track.id ? 'active' : ''}`}
          onClick={() => onSelect(track.id)}
        >
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
