import { useState, useCallback } from 'react';
import type { DragEvent } from 'react';
import { motion } from 'framer-motion';
import { useStore } from '../store';

interface FolderDropZoneProps {
  onScanComplete?: () => void;
  compact?: boolean;
}

export function FolderDropZone({ onScanComplete, compact = false }: FolderDropZoneProps) {
  const [isDragging, setIsDragging] = useState(false);
  const [isScanning, setIsScanning] = useState(false);
  const [folderPath, setFolderPath] = useState('');
  const [showInput, setShowInput] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { scanLibrary, fetchTracks } = useStore();

  const handleScan = useCallback(async (path: string) => {
    if (!path.trim()) {
      setError('Please enter a folder path');
      return;
    }

    setIsScanning(true);
    setError(null);

    try {
      await scanLibrary([path.trim()]);
      await fetchTracks();
      setFolderPath('');
      setShowInput(false);
      onScanComplete?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Scan failed');
    } finally {
      setIsScanning(false);
    }
  }, [scanLibrary, fetchTracks, onScanComplete]);

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

    // Try to get file path from native app context
    const files = e.dataTransfer.files;
    if (files && files.length > 0) {
      const file = files[0] as File & { path?: string };
      if (file.path) {
        handleScan(file.path);
        return;
      }
    }

    // Show input for manual entry if we can't get the path
    setShowInput(true);
  }, [handleScan]);

  if (compact) {
    return (
      <div
        className={`folder-drop-compact ${isDragging ? 'dragging' : ''}`}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={() => setShowInput(true)}
      >
        {isScanning ? (
          <span className="drop-scanning">Scanning...</span>
        ) : isDragging ? (
          <span className="drop-active">Drop to add</span>
        ) : (
          <span className="drop-idle">+ Add folder</span>
        )}
      </div>
    );
  }

  return (
    <motion.div
      className={`folder-drop-zone ${isDragging ? 'dragging' : ''} ${isScanning ? 'scanning' : ''}`}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
    >
      {isScanning ? (
        <div className="drop-zone-content">
          <div className="scanning-spinner" />
          <span className="drop-text">Scanning folder...</span>
        </div>
      ) : showInput ? (
        <div className="drop-zone-input">
          <input
            type="text"
            className="folder-path-input"
            placeholder="/Users/you/Music/DJ Library"
            value={folderPath}
            onChange={(e) => setFolderPath(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleScan(folderPath)}
            autoFocus
          />
          <div className="input-actions">
            <button
              className="btn-scan"
              onClick={() => handleScan(folderPath)}
              disabled={!folderPath.trim()}
            >
              Scan
            </button>
            <button
              className="btn-cancel"
              onClick={() => {
                setShowInput(false);
                setFolderPath('');
                setError(null);
              }}
            >
              Cancel
            </button>
          </div>
          {error && <p className="drop-error">{error}</p>}
        </div>
      ) : (
        <div
          className="drop-zone-content clickable"
          onClick={() => setShowInput(true)}
        >
          {isDragging ? (
            <>
              <span className="drop-icon large">⬇️</span>
              <span className="drop-text">Drop folder here</span>
            </>
          ) : (
            <>
              <span className="drop-icon">📁</span>
              <span className="drop-text">Add more music</span>
              <span className="drop-hint">Drag folder or click to browse</span>
            </>
          )}
        </div>
      )}
    </motion.div>
  );
}
