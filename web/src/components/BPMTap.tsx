import { useState, useCallback, useEffect, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';

interface BPMTapProps {
  onBPMDetected?: (bpm: number) => void;
  compact?: boolean;
}

export function BPMTap({ onBPMDetected, compact = false }: BPMTapProps) {
  const [, setTaps] = useState<number[]>([]);
  const [bpm, setBpm] = useState<number | null>(null);
  const [isActive, setIsActive] = useState(false);
  const [tapCount, setTapCount] = useState(0);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Calculate BPM from tap intervals
  const calculateBPM = useCallback((tapTimes: number[]): number | null => {
    if (tapTimes.length < 2) return null;

    // Calculate intervals between taps
    const intervals: number[] = [];
    for (let i = 1; i < tapTimes.length; i++) {
      intervals.push(tapTimes[i] - tapTimes[i - 1]);
    }

    // Average the intervals
    const avgInterval = intervals.reduce((a, b) => a + b, 0) / intervals.length;

    // Convert to BPM (60000ms = 1 minute)
    const calculatedBpm = Math.round(60000 / avgInterval);

    // Clamp to reasonable DJ range
    if (calculatedBpm < 60) return calculatedBpm * 2; // Double if too slow
    if (calculatedBpm > 200) return Math.round(calculatedBpm / 2); // Halve if too fast

    return calculatedBpm;
  }, []);

  // Handle tap
  const handleTap = useCallback(() => {
    const now = Date.now();

    setTaps(prev => {
      // Reset if last tap was more than 2 seconds ago
      const newTaps = prev.length > 0 && now - prev[prev.length - 1] > 2000
        ? [now]
        : [...prev.slice(-7), now]; // Keep last 8 taps for averaging

      const newBpm = calculateBPM(newTaps);
      setBpm(newBpm);

      if (newBpm && onBPMDetected) {
        onBPMDetected(newBpm);
      }

      return newTaps;
    });

    setTapCount(prev => prev + 1);
    setIsActive(true);

    // Reset active state after animation
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }
    timeoutRef.current = setTimeout(() => {
      setIsActive(false);
    }, 100);
  }, [calculateBPM, onBPMDetected]);

  // Reset
  const handleReset = useCallback(() => {
    setTaps([]);
    setBpm(null);
    setTapCount(0);
  }, []);

  // Keyboard handler for spacebar
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Only handle if not in an input
      if (
        e.target instanceof HTMLInputElement ||
        e.target instanceof HTMLTextAreaElement
      ) {
        return;
      }

      if (e.key === 't' || e.key === 'T') {
        e.preventDefault();
        handleTap();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleTap]);

  // Cleanup timeout
  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  if (compact) {
    return (
      <div className="bpm-tap-compact">
        <motion.button
          className={`tap-btn-compact ${isActive ? 'active' : ''}`}
          onClick={handleTap}
          whileTap={{ scale: 0.95 }}
          title="Tap for BPM (T)"
        >
          <span className="tap-icon">♪</span>
          {bpm ? (
            <span className="tap-bpm">{bpm}</span>
          ) : (
            <span className="tap-label">TAP</span>
          )}
        </motion.button>
        {bpm && (
          <button className="tap-reset-compact" onClick={handleReset} title="Reset">
            ✕
          </button>
        )}
      </div>
    );
  }

  return (
    <div className="bpm-tap-container">
      <div className="bpm-tap-header">
        <h4>Tap Tempo</h4>
        <span className="tap-hint">Press <kbd>T</kbd> or click</span>
      </div>

      <motion.button
        className={`tap-btn ${isActive ? 'active' : ''}`}
        onClick={handleTap}
        whileTap={{ scale: 0.95 }}
      >
        <AnimatePresence mode="wait">
          {bpm ? (
            <motion.div
              key="bpm"
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.8 }}
              className="tap-bpm-display"
            >
              <span className="bpm-value">{bpm}</span>
              <span className="bpm-unit">BPM</span>
            </motion.div>
          ) : (
            <motion.div
              key="tap"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="tap-prompt"
            >
              TAP
            </motion.div>
          )}
        </AnimatePresence>
      </motion.button>

      <div className="tap-info">
        <span className="tap-count">{tapCount} taps</span>
        {bpm && (
          <button className="tap-reset" onClick={handleReset}>
            Reset
          </button>
        )}
      </div>

      {/* Visual beat indicator */}
      <div className="beat-indicator">
        {[0, 1, 2, 3].map((i) => (
          <motion.div
            key={i}
            className="beat-dot"
            animate={{
              scale: isActive && i === (tapCount % 4) ? 1.5 : 1,
              opacity: isActive && i === (tapCount % 4) ? 1 : 0.3,
            }}
            transition={{ duration: 0.1 }}
          />
        ))}
      </div>
    </div>
  );
}
