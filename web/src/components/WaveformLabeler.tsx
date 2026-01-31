import { useState, useRef, useCallback, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import type { Track } from '../types';
import type { DJSectionLabel, TrainingLabelResponse } from '../api';
import * as api from '../api';

// Label configuration with colors and keyboard shortcuts
const LABEL_CONFIG: Record<DJSectionLabel, { displayName: string; color: string; key: string }> = {
  intro: { displayName: 'Intro', color: '#22c55e', key: '1' },
  build: { displayName: 'Build', color: '#eab308', key: '2' },
  drop: { displayName: 'Drop', color: '#ef4444', key: '3' },
  break: { displayName: 'Break', color: '#a855f7', key: '4' },
  outro: { displayName: 'Outro', color: '#3b82f6', key: '5' },
  verse: { displayName: 'Verse', color: '#4b5563', key: '6' },
  chorus: { displayName: 'Chorus', color: '#ec4899', key: '7' },
};

interface WaveformLabelerProps {
  track: Track;
  existingLabels: TrainingLabelResponse[];
  onLabelAdded: () => void;
  onLabelDeleted: (id: number) => void;
}

interface SelectionRange {
  startBeat: number;
  endBeat: number;
}

export function WaveformLabeler({ track, existingLabels, onLabelAdded, onLabelDeleted }: WaveformLabelerProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [selectedLabel, setSelectedLabel] = useState<DJSectionLabel>('drop');
  const [selection, setSelection] = useState<SelectionRange | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState<number | null>(null);
  const [isAdding, setIsAdding] = useState(false);
  const [hoveredLabel, setHoveredLabel] = useState<number | null>(null);

  // Calculate total beats from BPM and duration
  const bpm = track.bpm || 120;
  const duration = track.sections?.reduce((max, s) => Math.max(max, s.end), 0) || 512;
  const totalBeats = Math.ceil((duration / 60) * bpm);

  // Convert position to beat
  const positionToBeat = useCallback((clientX: number): number => {
    if (!containerRef.current) return 0;
    const rect = containerRef.current.getBoundingClientRect();
    const x = clientX - rect.left;
    const ratio = Math.max(0, Math.min(1, x / rect.width));
    return Math.round(ratio * totalBeats);
  }, [totalBeats]);

  // Handle mouse events for drag selection
  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    if (e.button !== 0) return; // Left click only
    const beat = positionToBeat(e.clientX);
    setDragStart(beat);
    setIsDragging(true);
    setSelection({ startBeat: beat, endBeat: beat });
  }, [positionToBeat]);

  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    if (!isDragging || dragStart === null) return;
    const beat = positionToBeat(e.clientX);
    setSelection({
      startBeat: Math.min(dragStart, beat),
      endBeat: Math.max(dragStart, beat),
    });
  }, [isDragging, dragStart, positionToBeat]);

  const handleMouseUp = useCallback(() => {
    setIsDragging(false);
    setDragStart(null);
    // Keep selection for labeling
  }, []);

  // Keyboard shortcuts for label types
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return;
      }

      // Number keys 1-7 for label types
      const labelKeys = Object.entries(LABEL_CONFIG);
      const found = labelKeys.find(([, config]) => config.key === e.key);
      if (found) {
        e.preventDefault();
        setSelectedLabel(found[0] as DJSectionLabel);
      }

      // Enter to add label
      if (e.key === 'Enter' && selection && selection.endBeat > selection.startBeat) {
        e.preventDefault();
        handleAddLabel();
      }

      // Escape to clear selection
      if (e.key === 'Escape') {
        setSelection(null);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [selection, selectedLabel]);

  // Add label
  const handleAddLabel = async () => {
    if (!selection || selection.endBeat <= selection.startBeat) return;

    setIsAdding(true);
    try {
      // Calculate time from beats
      const startTime = (selection.startBeat / bpm) * 60;
      const endTime = (selection.endBeat / bpm) * 60;

      // Get track ID from API
      const tracks = await api.listTracks({ query: track.title, limit: 10 });
      const apiTrack = tracks.find(t => t.content_hash === track.id);
      if (!apiTrack?.id) {
        throw new Error('Track not found in database');
      }

      await api.addTrainingLabel({
        track_id: apiTrack.id,
        label_value: selectedLabel,
        start_beat: selection.startBeat,
        end_beat: selection.endBeat,
        start_time_seconds: startTime,
        end_time_seconds: endTime,
        source: 'user',
      });

      setSelection(null);
      onLabelAdded();
    } catch (error) {
      console.error('Failed to add label:', error);
      alert('Failed to add label: ' + (error as Error).message);
    } finally {
      setIsAdding(false);
    }
  };

  // Get labels for this track
  const trackLabels = existingLabels.filter(l => l.content_hash === track.id);

  return (
    <div className="waveform-labeler">
      {/* Header with track info */}
      <div className="labeler-header">
        <div className="track-info">
          <span className="track-title">{track.title}</span>
          <span className="track-meta">{track.artist} • {bpm} BPM • {track.key}</span>
        </div>
        <div className="shortcut-hint">
          Press <kbd>1</kbd>-<kbd>7</kbd> for labels, <kbd>Enter</kbd> to add
        </div>
      </div>

      {/* Label selector */}
      <div className="label-selector-bar">
        {(Object.entries(LABEL_CONFIG) as [DJSectionLabel, typeof LABEL_CONFIG[DJSectionLabel]][]).map(
          ([value, config]) => (
            <button
              key={value}
              className={`label-chip ${selectedLabel === value ? 'active' : ''}`}
              style={{
                '--label-color': config.color,
              } as React.CSSProperties}
              onClick={() => setSelectedLabel(value)}
            >
              <kbd>{config.key}</kbd>
              <span>{config.displayName}</span>
            </button>
          )
        )}
      </div>

      {/* Waveform visualization with labels */}
      <div
        ref={containerRef}
        className="waveform-container"
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
      >
        {/* Waveform bars */}
        <div className="waveform-bars">
          {track.waveformSummary?.map((val, i) => (
            <div
              key={i}
              className="waveform-bar"
              style={{ height: `${val * 100}%` }}
            />
          )) || (
            // Placeholder waveform if no data
            Array.from({ length: 100 }).map((_, i) => (
              <div
                key={i}
                className="waveform-bar placeholder"
                style={{ height: `${30 + Math.random() * 40}%` }}
              />
            ))
          )}
        </div>

        {/* Existing labels overlay */}
        <div className="labels-overlay">
          {trackLabels.map((label) => {
            const config = LABEL_CONFIG[label.label_value as DJSectionLabel];
            const left = (label.start_beat / totalBeats) * 100;
            const width = ((label.end_beat - label.start_beat) / totalBeats) * 100;
            return (
              <motion.div
                key={label.id}
                className={`label-region ${hoveredLabel === label.id ? 'hovered' : ''}`}
                style={{
                  left: `${left}%`,
                  width: `${width}%`,
                  backgroundColor: config?.color || '#6b7280',
                }}
                initial={{ opacity: 0, scaleY: 0 }}
                animate={{ opacity: 0.6, scaleY: 1 }}
                onMouseEnter={() => setHoveredLabel(label.id)}
                onMouseLeave={() => setHoveredLabel(null)}
              >
                <span className="label-name">{config?.displayName}</span>
                <button
                  className="label-delete"
                  onClick={(e) => {
                    e.stopPropagation();
                    onLabelDeleted(label.id);
                  }}
                >
                  ×
                </button>
              </motion.div>
            );
          })}
        </div>

        {/* Selection overlay */}
        <AnimatePresence>
          {selection && selection.endBeat > selection.startBeat && (
            <motion.div
              className="selection-overlay"
              style={{
                left: `${(selection.startBeat / totalBeats) * 100}%`,
                width: `${((selection.endBeat - selection.startBeat) / totalBeats) * 100}%`,
                backgroundColor: LABEL_CONFIG[selectedLabel].color,
              }}
              initial={{ opacity: 0 }}
              animate={{ opacity: 0.4 }}
              exit={{ opacity: 0 }}
            >
              <span className="selection-label">
                {LABEL_CONFIG[selectedLabel].displayName}
                <span className="selection-beats">
                  ({selection.startBeat} - {selection.endBeat})
                </span>
              </span>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Beat markers */}
        <div className="beat-markers">
          {Array.from({ length: Math.min(16, Math.ceil(totalBeats / 32)) }).map((_, i) => {
            const beat = i * 32;
            const left = (beat / totalBeats) * 100;
            return (
              <div key={i} className="beat-marker" style={{ left: `${left}%` }}>
                <span>{beat}</span>
              </div>
            );
          })}
        </div>
      </div>

      {/* Add button */}
      {selection && selection.endBeat > selection.startBeat && (
        <motion.div
          className="add-label-bar"
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
        >
          <span className="selection-info">
            <span
              className="color-dot"
              style={{ backgroundColor: LABEL_CONFIG[selectedLabel].color }}
            />
            {LABEL_CONFIG[selectedLabel].displayName}: beats {selection.startBeat} - {selection.endBeat}
          </span>
          <div className="action-buttons">
            <button className="clear-btn" onClick={() => setSelection(null)}>
              Clear
            </button>
            <button
              className="add-btn"
              onClick={handleAddLabel}
              disabled={isAdding}
            >
              {isAdding ? 'Adding...' : 'Add Label (Enter)'}
            </button>
          </div>
        </motion.div>
      )}

      {/* Instructions when no selection */}
      {!selection && (
        <div className="labeler-hint">
          Click and drag on the waveform to select a region, then press <kbd>Enter</kbd> to add a label
        </div>
      )}
    </div>
  );
}
