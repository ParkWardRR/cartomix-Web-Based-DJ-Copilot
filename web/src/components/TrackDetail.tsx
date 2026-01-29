import { useState, useCallback } from 'react';
import type { Track } from '../types';
import { AudioPlayer } from './AudioPlayer';

type Props = {
  track?: Track;
};

export function TrackDetail({ track }: Props) {
  const [playheadPosition, setPlayheadPosition] = useState(0);

  const handleTimeUpdate = useCallback((_time: number, position: number) => {
    setPlayheadPosition(position);
  }, []);

  const handleWaveformClick = useCallback((e: React.MouseEvent<HTMLDivElement>) => {
    // Calculate click position relative to waveform
    const rect = e.currentTarget.getBoundingClientRect();
    const position = (e.clientX - rect.left) / rect.width;
    // The AudioPlayer will handle seeking via its timeline
    // This is just for visual feedback
    setPlayheadPosition(position);
  }, []);

  if (!track) {
    return (
      <div className="card">
        <h3>Track detail</h3>
        <p className="muted">Select a track to see waveform, sections, and cues.</p>
      </div>
    );
  }

  // Find current section based on playhead
  const currentBeat = Math.floor(playheadPosition * (track.cues[track.cues.length - 1]?.beat || 100));
  const currentSection = track.sections.find(
    (s) => currentBeat >= s.start && currentBeat < s.end
  );

  return (
    <div className="card track-detail">
      <header className="detail-header">
        <div>
          <div className="detail-title">{track.title}</div>
          <div className="detail-artist">{track.artist}</div>
        </div>
        <div className="pill-row">
          <span className="pill pill-primary">{track.bpm} BPM</span>
          <span className="pill pill-secondary">{track.key}</span>
          <span className="pill pill-secondary">Energy {track.energy}</span>
          {currentSection && (
            <span className={`pill pill-section ${currentSection.label.toLowerCase()}`}>
              {currentSection.label}
            </span>
          )}
        </div>
      </header>

      {/* Audio Player Controls */}
      <AudioPlayer
        trackPath={track.path}
        onTimeUpdate={handleTimeUpdate}
      />

      {/* Waveform with playhead */}
      <div className="waveform" onClick={handleWaveformClick}>
        <div className="energy-arc">
          {track.waveformSummary.map((level, idx) => {
            const barPosition = idx / track.waveformSummary.length;
            const isPlayed = barPosition <= playheadPosition;
            return (
              <div
                key={idx}
                className={`energy-bar ${isPlayed ? 'played' : ''}`}
                style={{ height: `${level * 6}%` }}
              />
            );
          })}
        </div>

        {/* Section overlays */}
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

        {/* Cue markers */}
        {track.cues.map((cue) => (
          <div
            key={`${cue.label}-${cue.beat}`}
            className="cue-marker"
            style={{ left: `${cue.beat / 4}%`, backgroundColor: cue.color || 'var(--color-accent)' }}
            title={`${cue.type} @ beat ${cue.beat}`}
          />
        ))}

        {/* Playhead */}
        <div
          className="waveform-playhead"
          style={{ left: `${playheadPosition * 100}%` }}
        />
      </div>

      <div className="detail-grid">
        <div>
          <h4>Cue points</h4>
          <ul className="list cue-list">
            {track.cues.map((cue) => {
              const isActive = Math.abs(cue.beat - currentBeat) < 4;
              return (
                <li key={`${cue.type}-${cue.beat}`} className={isActive ? 'active' : ''}>
                  <span className="cue-color" style={{ backgroundColor: cue.color || 'var(--color-accent)' }} />
                  <span className="pill pill-secondary">{cue.type}</span>
                  <span className="cue-info">Beat {cue.beat} — {cue.label}</span>
                </li>
              );
            })}
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
