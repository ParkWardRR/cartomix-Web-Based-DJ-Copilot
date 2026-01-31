import { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';

export interface AnalysisStatus {
  isAnalyzing: boolean;
  currentTrack: string | null;
  completed: number;
  total: number;
  errors: string[];
}

interface AnalysisProgressProps {
  status: AnalysisStatus;
  onClose?: () => void;
}

export function AnalysisProgress({ status, onClose }: AnalysisProgressProps) {
  const [isVisible, setIsVisible] = useState(false);
  const [lastStatus, setLastStatus] = useState<AnalysisStatus | null>(null);

  useEffect(() => {
    if (status.isAnalyzing || status.completed > 0) {
      setIsVisible(true);
      setLastStatus(status);
    }
  }, [status]);

  const handleClose = useCallback(() => {
    setIsVisible(false);
    onClose?.();
  }, [onClose]);

  // Auto-hide after completion
  useEffect(() => {
    if (!status.isAnalyzing && status.completed > 0 && status.completed === status.total) {
      const timer = setTimeout(() => {
        setIsVisible(false);
      }, 5000);
      return () => clearTimeout(timer);
    }
  }, [status.isAnalyzing, status.completed, status.total]);

  const displayStatus = status.isAnalyzing ? status : lastStatus;
  if (!displayStatus || !isVisible) return null;

  const progress = displayStatus.total > 0
    ? (displayStatus.completed / displayStatus.total) * 100
    : 0;

  const isComplete = !displayStatus.isAnalyzing && displayStatus.completed === displayStatus.total;

  return (
    <AnimatePresence>
      <motion.div
        className="analysis-progress-bar"
        initial={{ y: 100, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        exit={{ y: 100, opacity: 0 }}
        transition={{ type: 'spring', stiffness: 300, damping: 30 }}
      >
        <div className="analysis-icon">
          {isComplete ? '✓' : '🎵'}
        </div>

        <div className="analysis-info">
          <div className="analysis-title">
            {isComplete ? 'Analysis Complete' : 'Analyzing Tracks'}
          </div>
          <div className="analysis-subtitle">
            {displayStatus.currentTrack && !isComplete && (
              <span className="analysis-track-name">
                {displayStatus.currentTrack}
              </span>
            )}
            {isComplete && (
              <span>{displayStatus.completed} tracks analyzed</span>
            )}
          </div>
        </div>

        <div className="analysis-progress-track">
          <motion.div
            className="analysis-progress-fill"
            initial={{ width: 0 }}
            animate={{ width: `${progress}%` }}
            transition={{ duration: 0.3 }}
          />
        </div>

        <div className="analysis-stats">
          <span className="analysis-count">
            <strong>{displayStatus.completed}</strong> / {displayStatus.total}
          </span>
          {displayStatus.errors.length > 0 && (
            <span className="analysis-errors" style={{ color: 'var(--color-error)' }}>
              {displayStatus.errors.length} errors
            </span>
          )}
        </div>

        <button className="analysis-close" onClick={handleClose} aria-label="Close">
          ×
        </button>
      </motion.div>
    </AnimatePresence>
  );
}

// Hook to manage analysis state
export function useAnalysisProgress() {
  const [status, setStatus] = useState<AnalysisStatus>({
    isAnalyzing: false,
    currentTrack: null,
    completed: 0,
    total: 0,
    errors: [],
  });

  const startAnalysis = useCallback((total: number) => {
    setStatus({
      isAnalyzing: true,
      currentTrack: null,
      completed: 0,
      total,
      errors: [],
    });
  }, []);

  const updateProgress = useCallback((currentTrack: string, completed: number) => {
    setStatus(prev => ({
      ...prev,
      currentTrack,
      completed,
    }));
  }, []);

  const completeAnalysis = useCallback(() => {
    setStatus(prev => ({
      ...prev,
      isAnalyzing: false,
    }));
  }, []);

  const addError = useCallback((error: string) => {
    setStatus(prev => ({
      ...prev,
      errors: [...prev.errors, error],
    }));
  }, []);

  const reset = useCallback(() => {
    setStatus({
      isAnalyzing: false,
      currentTrack: null,
      completed: 0,
      total: 0,
      errors: [],
    });
  }, []);

  return {
    status,
    startAnalysis,
    updateProgress,
    completeAnalysis,
    addError,
    reset,
  };
}
