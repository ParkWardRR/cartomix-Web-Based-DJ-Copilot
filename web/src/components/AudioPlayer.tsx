import { useEffect, useCallback } from 'react';
import { useAudioPlayer } from '../hooks/useAudioPlayer';
import { useStore } from '../store';

type Props = {
  trackPath?: string;
  duration?: number;
  onTimeUpdate?: (time: number, position: number) => void;
};

function formatTime(seconds: number): string {
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins}:${secs.toString().padStart(2, '0')}`;
}

export function AudioPlayer({ trackPath, duration: trackDuration, onTimeUpdate }: Props) {
  const [state, controls] = useAudioPlayer();
  const { setIsPlaying, setPlayheadPosition } = useStore();

  // Load track when path changes
  useEffect(() => {
    if (trackPath) {
      // Construct URL for audio file - use API endpoint for serving audio
      const audioUrl = `/api/audio?path=${encodeURIComponent(trackPath)}`;
      controls.loadTrack(audioUrl).catch(() => {
        // If API fails, try direct file path (for local development)
        controls.loadTrack(trackPath);
      });
    }
  }, [trackPath, controls]);

  // Sync state with store
  useEffect(() => {
    setIsPlaying(state.isPlaying);
    if (state.duration > 0) {
      const position = state.currentTime / state.duration;
      setPlayheadPosition(position);
      onTimeUpdate?.(state.currentTime, position);
    }
  }, [state.isPlaying, state.currentTime, state.duration, setIsPlaying, setPlayheadPosition, onTimeUpdate]);

  const handlePlayPause = useCallback(async () => {
    if (state.isPlaying) {
      controls.pause();
    } else {
      await controls.play();
    }
  }, [state.isPlaying, controls]);

  const handleStop = useCallback(() => {
    controls.stop();
  }, [controls]);

  const handleSeek = useCallback((e: React.MouseEvent<HTMLDivElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const position = (e.clientX - rect.left) / rect.width;
    controls.seekToPosition(Math.max(0, Math.min(1, position)));
  }, [controls]);

  const handleSpeedChange = useCallback((speed: number) => {
    controls.setPlaybackRate(speed);
  }, [controls]);

  const displayDuration = state.duration || trackDuration || 0;
  const progress = displayDuration > 0 ? (state.currentTime / displayDuration) * 100 : 0;

  return (
    <div className="audio-player">
      <div className="player-controls">
        <button
          className={`player-btn ${state.isPlaying ? 'playing' : ''}`}
          onClick={handlePlayPause}
          disabled={state.isLoading || (!trackPath && !state.duration)}
          title={state.isPlaying ? 'Pause' : 'Play'}
        >
          {state.isLoading ? (
            <span className="player-spinner" />
          ) : state.isPlaying ? (
            <span className="player-icon">||</span>
          ) : (
            <span className="player-icon play-icon">▶</span>
          )}
        </button>
        <button
          className="player-btn stop"
          onClick={handleStop}
          disabled={!state.isPlaying && state.currentTime === 0}
          title="Stop"
        >
          <span className="player-icon">■</span>
        </button>
      </div>

      <div className="player-timeline" onClick={handleSeek}>
        <div className="timeline-track">
          <div
            className="timeline-progress"
            style={{ width: `${progress}%` }}
          />
          <div
            className="timeline-playhead"
            style={{ left: `${progress}%` }}
          />
        </div>
      </div>

      <div className="player-time">
        <span className="time-current">{formatTime(state.currentTime)}</span>
        <span className="time-separator">/</span>
        <span className="time-total">{formatTime(displayDuration)}</span>
      </div>

      <div className="player-speed">
        {[0.5, 0.75, 1, 1.25, 1.5].map((speed) => (
          <button
            key={speed}
            className={`speed-btn ${state.playbackRate === speed ? 'active' : ''}`}
            onClick={() => handleSpeedChange(speed)}
            title={`${speed}x speed`}
          >
            {speed}x
          </button>
        ))}
      </div>

      {state.error && (
        <div className="player-error" title={state.error}>
          <span className="error-icon">!</span>
        </div>
      )}
    </div>
  );
}
