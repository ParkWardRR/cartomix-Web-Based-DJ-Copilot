import { useState, useRef, useCallback, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import type { Track } from '../types';

interface CrossfadePreviewProps {
  trackA?: Track;
  trackB?: Track;
  transitionPoint?: number; // 0-1 position in track A where transition starts
  crossfadeDuration?: number; // seconds
  onClose?: () => void;
}

interface AudioState {
  isPlaying: boolean;
  currentTime: number;
  duration: number;
  volume: number;
}

export function CrossfadePreview({
  trackA,
  trackB,
  transitionPoint = 0.8,
  crossfadeDuration = 16,
  onClose,
}: CrossfadePreviewProps) {
  const audioRefA = useRef<HTMLAudioElement | null>(null);
  const audioRefB = useRef<HTMLAudioElement | null>(null);
  const animationRef = useRef<number | null>(null);

  const [isPlaying, setIsPlaying] = useState(false);
  const [crossfadePosition, setCrossfadePosition] = useState(0); // 0 = all A, 1 = all B
  const [stateA, setStateA] = useState<AudioState>({
    isPlaying: false,
    currentTime: 0,
    duration: 0,
    volume: 1,
  });
  const [stateB, setStateB] = useState<AudioState>({
    isPlaying: false,
    currentTime: 0,
    duration: 0,
    volume: 0,
  });
  const [localTransitionPoint, setLocalTransitionPoint] = useState(transitionPoint);
  const [localCrossfadeDuration, setLocalCrossfadeDuration] = useState(crossfadeDuration);

  // Get audio URL from path
  const getAudioUrl = (path?: string) => {
    if (!path) return '';
    return `/api/audio?path=${encodeURIComponent(path)}`;
  };

  // Update volumes based on crossfade position
  const updateVolumes = useCallback((position: number) => {
    if (audioRefA.current) {
      audioRefA.current.volume = Math.cos(position * Math.PI / 2);
    }
    if (audioRefB.current) {
      audioRefB.current.volume = Math.sin(position * Math.PI / 2);
    }
  }, []);

  // Animation loop for crossfade
  const animate = useCallback(() => {
    if (!audioRefA.current || !audioRefB.current) return;

    const timeA = audioRefA.current.currentTime;
    const durationA = audioRefA.current.duration || 1;
    const transitionStartTime = durationA * localTransitionPoint;
    const transitionEndTime = transitionStartTime + localCrossfadeDuration;

    // Update state A
    setStateA({
      isPlaying: !audioRefA.current.paused,
      currentTime: timeA,
      duration: durationA,
      volume: audioRefA.current.volume,
    });

    // Update state B
    setStateB({
      isPlaying: !audioRefB.current.paused,
      currentTime: audioRefB.current.currentTime,
      duration: audioRefB.current.duration || 0,
      volume: audioRefB.current.volume,
    });

    // Calculate crossfade position
    if (timeA < transitionStartTime) {
      setCrossfadePosition(0);
      updateVolumes(0);
    } else if (timeA >= transitionEndTime) {
      setCrossfadePosition(1);
      updateVolumes(1);
    } else {
      const progress = (timeA - transitionStartTime) / localCrossfadeDuration;
      setCrossfadePosition(progress);
      updateVolumes(progress);

      // Start track B at the right time
      if (audioRefB.current.paused && progress > 0) {
        audioRefB.current.currentTime = 0;
        audioRefB.current.play().catch(() => {});
      }
    }

    animationRef.current = requestAnimationFrame(animate);
  }, [localTransitionPoint, localCrossfadeDuration, updateVolumes]);

  // Start/stop animation
  useEffect(() => {
    if (isPlaying) {
      animationRef.current = requestAnimationFrame(animate);
    } else if (animationRef.current) {
      cancelAnimationFrame(animationRef.current);
    }
    return () => {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
      }
    };
  }, [isPlaying, animate]);

  // Play/pause handler
  const togglePlay = useCallback(() => {
    if (!audioRefA.current) return;

    if (isPlaying) {
      audioRefA.current.pause();
      audioRefB.current?.pause();
      setIsPlaying(false);
    } else {
      // Start from transition point - 4 seconds
      const durationA = audioRefA.current.duration || 60;
      const startTime = Math.max(0, durationA * localTransitionPoint - 4);
      audioRefA.current.currentTime = startTime;
      audioRefA.current.play().catch(() => {});
      setIsPlaying(true);
    }
  }, [isPlaying, localTransitionPoint]);

  // Reset handler
  const handleReset = useCallback(() => {
    if (audioRefA.current) {
      audioRefA.current.pause();
      audioRefA.current.currentTime = 0;
    }
    if (audioRefB.current) {
      audioRefB.current.pause();
      audioRefB.current.currentTime = 0;
    }
    setIsPlaying(false);
    setCrossfadePosition(0);
    updateVolumes(0);
  }, [updateVolumes]);

  // Format time as MM:SS
  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  if (!trackA || !trackB) {
    return (
      <div className="crossfade-preview-empty">
        <p>Select two tracks to preview the transition</p>
      </div>
    );
  }

  return (
    <div className="crossfade-preview">
      <div className="crossfade-header">
        <h3>Transition Preview</h3>
        {onClose && (
          <button className="close-btn" onClick={onClose}>
            &times;
          </button>
        )}
      </div>

      {/* Hidden audio elements */}
      <audio ref={audioRefA} src={getAudioUrl(trackA.path)} preload="metadata" />
      <audio ref={audioRefB} src={getAudioUrl(trackB.path)} preload="metadata" />

      {/* Track info */}
      <div className="crossfade-tracks">
        <div className={`track-info track-a ${crossfadePosition < 0.5 ? 'active' : ''}`}>
          <span className="track-label">A</span>
          <div className="track-details">
            <span className="track-title">{trackA.title}</span>
            <span className="track-artist">{trackA.artist}</span>
          </div>
          <div className="track-meta">
            <span className="bpm">{trackA.bpm} BPM</span>
            <span className="key">{trackA.key}</span>
          </div>
        </div>

        <div className="crossfade-arrow">
          <motion.div
            className="arrow-fill"
            animate={{ width: `${crossfadePosition * 100}%` }}
            transition={{ duration: 0.1 }}
          />
        </div>

        <div className={`track-info track-b ${crossfadePosition >= 0.5 ? 'active' : ''}`}>
          <span className="track-label">B</span>
          <div className="track-details">
            <span className="track-title">{trackB.title}</span>
            <span className="track-artist">{trackB.artist}</span>
          </div>
          <div className="track-meta">
            <span className="bpm">{trackB.bpm} BPM</span>
            <span className="key">{trackB.key}</span>
          </div>
        </div>
      </div>

      {/* Crossfade visualization */}
      <div className="crossfade-viz">
        <div className="volume-meter meter-a">
          <div className="meter-label">A</div>
          <motion.div
            className="meter-fill"
            animate={{ height: `${stateA.volume * 100}%` }}
            transition={{ duration: 0.05 }}
          />
        </div>

        <div className="crossfade-slider-container">
          <div className="timeline">
            <div className="time-marker start">
              {formatTime(stateA.duration * localTransitionPoint - 4)}
            </div>
            <div className="time-marker end">
              {formatTime(stateA.duration * localTransitionPoint + localCrossfadeDuration)}
            </div>
          </div>

          <div className="crossfade-slider">
            <motion.div
              className="crossfade-progress"
              animate={{ width: `${crossfadePosition * 100}%` }}
              transition={{ duration: 0.1 }}
            />
            <div
              className="transition-marker"
              style={{ left: `${(4 / (4 + localCrossfadeDuration)) * 100}%` }}
            >
              <span>Mix</span>
            </div>
          </div>
        </div>

        <div className="volume-meter meter-b">
          <div className="meter-label">B</div>
          <motion.div
            className="meter-fill"
            animate={{ height: `${stateB.volume * 100}%` }}
            transition={{ duration: 0.05 }}
          />
        </div>
      </div>

      {/* Controls */}
      <div className="crossfade-controls">
        <motion.button
          className={`play-btn ${isPlaying ? 'playing' : ''}`}
          onClick={togglePlay}
          whileTap={{ scale: 0.95 }}
        >
          <AnimatePresence mode="wait">
            {isPlaying ? (
              <motion.span
                key="pause"
                initial={{ opacity: 0, scale: 0.8 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.8 }}
              >
                ⏸
              </motion.span>
            ) : (
              <motion.span
                key="play"
                initial={{ opacity: 0, scale: 0.8 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.8 }}
              >
                ▶
              </motion.span>
            )}
          </AnimatePresence>
          <span>{isPlaying ? 'Pause' : 'Preview Transition'}</span>
        </motion.button>

        <button className="reset-btn" onClick={handleReset}>
          Reset
        </button>
      </div>

      {/* Settings */}
      <div className="crossfade-settings">
        <div className="setting">
          <label>Transition Point</label>
          <input
            type="range"
            min="0.5"
            max="0.95"
            step="0.05"
            value={localTransitionPoint}
            onChange={(e) => setLocalTransitionPoint(parseFloat(e.target.value))}
          />
          <span>{Math.round(localTransitionPoint * 100)}%</span>
        </div>
        <div className="setting">
          <label>Crossfade Duration</label>
          <input
            type="range"
            min="4"
            max="32"
            step="4"
            value={localCrossfadeDuration}
            onChange={(e) => setLocalCrossfadeDuration(parseFloat(e.target.value))}
          />
          <span>{localCrossfadeDuration}s</span>
        </div>
      </div>
    </div>
  );
}
