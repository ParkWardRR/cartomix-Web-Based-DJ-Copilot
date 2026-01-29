import type { Track } from '../types';

type Props = {
  track?: Track;
};

export function TrackDetail({ track }: Props) {
  if (!track) {
    return (
      <div className="card">
        <h3>Track detail</h3>
        <p className="muted">Select a track to see waveform, sections, and cues.</p>
      </div>
    );
  }

  return (
    <div className="card">
      <header className="detail-header">
        <div>
          <div className="detail-title">{track.title}</div>
          <div className="detail-artist">{track.artist}</div>
        </div>
        <div className="pill-row">
          <span className="pill pill-primary">{track.bpm} BPM</span>
          <span className="pill pill-secondary">{track.key}</span>
          <span className="pill pill-secondary">Energy {track.energy}</span>
        </div>
      </header>

      <div className="waveform">
        <div className="energy-arc">
          {track.waveformSummary.map((level, idx) => (
            <div key={idx} className="energy-bar" style={{ height: `${level * 6}%` }} />
          ))}
        </div>
        {track.sections.map((section) => (
          <div
            key={`${section.label}-${section.start}`}
            className={`section-marker ${section.label.toLowerCase()}`}
            style={{
              left: `${section.start / 4}%`,
              width: `${(section.end - section.start) / 4}%`,
            }}
            title={`${section.label} (${section.start}–${section.end})`}
          />
        ))}
        {track.cues.map((cue) => (
          <div
            key={`${cue.label}-${cue.beat}`}
            className="cue-marker"
            style={{ left: `${cue.beat / 4}%`, backgroundColor: 'var(--color-accent)' }}
            title={`${cue.type} @ beat ${cue.beat}`}
          />
        ))}
      </div>

      <div className="detail-grid">
        <div>
          <h4>Cue points</h4>
          <ul className="list">
            {track.cues.map((cue) => (
              <li key={`${cue.type}-${cue.beat}`}>
                <span className="pill pill-secondary">{cue.type}</span> Beat {cue.beat} — {cue.label}
              </li>
            ))}
          </ul>
        </div>
        <div>
          <h4>Transition windows</h4>
          <ul className="list">
            {track.transitionWindows.map((win) => (
              <li key={`${win.label}-${win.start}`}>
                <span className="pill pill-secondary">{win.label}</span> Beats {win.start}–{win.end}
              </li>
            ))}
          </ul>
        </div>
      </div>
    </div>
  );
}
