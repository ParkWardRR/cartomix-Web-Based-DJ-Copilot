import { useState, useCallback } from 'react';
import type { DragEvent } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useStore } from '../store';

type WizardStep = 'welcome' | 'addMusic' | 'scanning' | 'complete';

export function IntroWizard() {
  const [step, setStep] = useState<WizardStep>('welcome');
  const [folderPath, setFolderPath] = useState('');
  const [scanProgress, setScanProgress] = useState({ processed: 0, total: 0, newTracks: 0 });
  const [error, setError] = useState<string | null>(null);
  const [isDragging, setIsDragging] = useState(false);

  const { scanLibrary, completeOnboarding, useDemoData } = useStore();

  const handleAddFolder = useCallback(async () => {
    if (!folderPath.trim()) {
      setError('Please enter a folder path');
      return;
    }

    setError(null);
    setStep('scanning');

    try {
      const result = await scanLibrary([folderPath.trim()]);
      setScanProgress({
        processed: result.processed,
        total: result.total,
        newTracks: result.newTracks.length,
      });
      setStep('complete');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Scan failed');
      setStep('addMusic');
    }
  }, [folderPath, scanLibrary]);

  const handleComplete = useCallback(() => {
    completeOnboarding();
  }, [completeOnboarding]);

  const handleSkipWithDemo = useCallback(() => {
    useDemoData();
    completeOnboarding();
  }, [useDemoData, completeOnboarding]);

  const handleDragOver = useCallback((e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  }, []);

  const handleDragLeave = useCallback((e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  }, []);

  const handleDrop = useCallback((e: DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);

    const items = e.dataTransfer.items;
    if (items && items.length > 0) {
      // Get the first item - check if it's a file/folder
      const item = items[0];
      if (item.kind === 'file') {
        const file = item.getAsFile();
        if (file) {
          // For web, we can't get the actual path, but the webkitRelativePath might help
          // In the native app context, we'd have access to the path
          // For now, show the file name as a hint
          const path = (file as File & { path?: string }).path || file.name;
          if (path) {
            setFolderPath(path);
            setError(null);
          }
        }
      }
    }

    // Also check files array
    const files = e.dataTransfer.files;
    if (files && files.length > 0) {
      const file = files[0] as File & { path?: string };
      if (file.path) {
        setFolderPath(file.path);
        setError(null);
      }
    }
  }, []);

  return (
    <div className="intro-wizard">
      <AnimatePresence mode="wait">
        {step === 'welcome' && (
          <motion.div
            key="welcome"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            className="wizard-step"
          >
            <div className="wizard-icon">◈</div>
            <h1>Welcome to Algiers</h1>
            <p className="wizard-subtitle">
              Your AI-powered DJ set prep copilot
            </p>
            <div className="wizard-features">
              <div className="feature">
                <span className="feature-icon">🎵</span>
                <div>
                  <strong>Audio Analysis</strong>
                  <p>BPM, key, energy, sections, and cue points</p>
                </div>
              </div>
              <div className="feature">
                <span className="feature-icon">🧠</span>
                <div>
                  <strong>Vibe Matching</strong>
                  <p>OpenL3 ML embeddings find tracks that "feel" similar</p>
                </div>
              </div>
              <div className="feature">
                <span className="feature-icon">📊</span>
                <div>
                  <strong>Set Planning</strong>
                  <p>Optimal track ordering with transition explanations</p>
                </div>
              </div>
            </div>
            <div className="wizard-actions">
              <button className="btn-primary" onClick={() => setStep('addMusic')}>
                Get Started
              </button>
              <button className="btn-secondary" onClick={handleSkipWithDemo}>
                Try with Demo Tracks
              </button>
            </div>
          </motion.div>
        )}

        {step === 'addMusic' && (
          <motion.div
            key="addMusic"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            className="wizard-step"
          >
            <div className="wizard-icon">📁</div>
            <h1>Add Your Music</h1>
            <p className="wizard-subtitle">
              Drag & drop a folder, or enter the path to your music
            </p>

            <div
              className={`drop-zone ${isDragging ? 'dragging' : ''} ${folderPath ? 'has-path' : ''}`}
              onDragOver={handleDragOver}
              onDragLeave={handleDragLeave}
              onDrop={handleDrop}
            >
              <div className="drop-zone-content">
                {isDragging ? (
                  <>
                    <span className="drop-icon">⬇️</span>
                    <span className="drop-text">Drop folder here</span>
                  </>
                ) : folderPath ? (
                  <>
                    <span className="drop-icon">✓</span>
                    <span className="drop-text path">{folderPath}</span>
                  </>
                ) : (
                  <>
                    <span className="drop-icon">🎵</span>
                    <span className="drop-text">Drag music folder here</span>
                    <span className="drop-hint">or enter path below</span>
                  </>
                )}
              </div>
            </div>

            <div className="folder-input-group">
              <input
                type="text"
                className="folder-input"
                placeholder="/Users/you/Music/DJ Library"
                value={folderPath}
                onChange={(e) => setFolderPath(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleAddFolder()}
              />
              <p className="input-hint">
                Supports: MP3, WAV, FLAC, AIFF, M4A, AAC
              </p>
            </div>
            {error && <p className="wizard-error">{error}</p>}
            <div className="wizard-actions">
              <button className="btn-primary" onClick={handleAddFolder} disabled={!folderPath.trim()}>
                Scan Folder
              </button>
              <button className="btn-text" onClick={() => setStep('welcome')}>
                Back
              </button>
            </div>
            <p className="wizard-note">
              🔒 100% local analysis using Apple Neural Engine.
              Your music never leaves your device.
            </p>
          </motion.div>
        )}

        {step === 'scanning' && (
          <motion.div
            key="scanning"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            className="wizard-step"
          >
            <div className="wizard-icon spinning">◎</div>
            <h1>Scanning Library</h1>
            <p className="wizard-subtitle">
              Finding and analyzing your tracks...
            </p>
            <div className="scan-progress">
              <div className="progress-bar">
                <div className="progress-fill" style={{ width: '100%' }} />
              </div>
              <p className="progress-text">Processing...</p>
            </div>
          </motion.div>
        )}

        {step === 'complete' && (
          <motion.div
            key="complete"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            className="wizard-step"
          >
            <div className="wizard-icon success">✓</div>
            <h1>Library Ready</h1>
            <p className="wizard-subtitle">
              Your music is ready to explore
            </p>
            <div className="scan-results">
              <div className="result-stat">
                <span className="stat-value">{scanProgress.total}</span>
                <span className="stat-label">Files Scanned</span>
              </div>
              <div className="result-stat">
                <span className="stat-value">{scanProgress.newTracks}</span>
                <span className="stat-label">New Tracks</span>
              </div>
            </div>
            <div className="wizard-actions">
              <button className="btn-primary" onClick={handleComplete}>
                Go to Library
              </button>
            </div>
            <p className="wizard-note">
              Tracks are analyzed in the background. Analysis includes BPM, key,
              energy, sections, and vibe embeddings.
            </p>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
