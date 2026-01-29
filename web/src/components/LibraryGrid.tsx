import type { Track } from '../types';

type Props = {
  tracks: Track[];
  selectedId?: string;
  onSelect: (id: string) => void;
};

export function LibraryGrid({ tracks, selectedId, onSelect }: Props) {
  if (!tracks.length) {
    return <div className="muted">No tracks match that filter.</div>;
  }

  return (
    <div className="track-grid">
      {tracks.map((track) => (
        <button
          key={track.id}
          className={`track-card ${selectedId === track.id ? 'active' : ''}`}
          onClick={() => onSelect(track.id)}
        >
          <div className="track-title">{track.title}</div>
          <div className="track-artist">{track.artist}</div>
          <div className="track-meta">
            <span>{track.bpm} BPM</span>
            <span>{track.key}</span>
            <span>Energy {track.energy}</span>
          </div>
          <div className="track-flags">
            <span className={`pill pill-${track.status}`}>{track.status}</span>
            {track.needsReview && <span className="pill pill-warn">Grid review</span>}
          </div>
        </button>
      ))}
    </div>
  );
}
