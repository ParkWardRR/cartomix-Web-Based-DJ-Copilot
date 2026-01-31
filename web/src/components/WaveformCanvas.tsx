import { useCallback, useEffect, useRef, useState } from 'react';
import type { MouseEvent as ReactMouseEvent } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import type { Cue, Section } from '../types';

interface WaveformCanvasProps {
  peaks: number[];
  sections?: Section[];
  cues?: Cue[];
  playheadPosition?: number; // 0-1
  duration?: number;
  isPlaying?: boolean;
  onSeek?: (position: number) => void;
  onSectionsChange?: (sections: Section[]) => void;
  height?: number;
  showBeatGrid?: boolean;
  bpm?: number;
  editable?: boolean;
}

const SECTION_LABELS = ['Intro', 'Build', 'Drop', 'Breakdown', 'Outro', 'Verse', 'Chorus', 'Body'] as const;

const SECTION_COLORS: Record<string, string> = {
  Intro: 'rgba(34, 197, 94, 0.25)',
  Drop: 'rgba(239, 68, 68, 0.25)',
  Break: 'rgba(168, 85, 247, 0.25)',
  Breakdown: 'rgba(168, 85, 247, 0.25)',
  Build: 'rgba(251, 191, 36, 0.25)',
  'Build/Drop': 'rgba(251, 191, 36, 0.25)',
  Body: 'rgba(59, 130, 246, 0.15)',
  Outro: 'rgba(234, 179, 8, 0.25)',
  Verse: 'rgba(96, 165, 250, 0.25)',
  Chorus: 'rgba(244, 114, 182, 0.25)',
};

const SECTION_BORDER_COLORS: Record<string, string> = {
  Intro: '#22c55e',
  Drop: '#ef4444',
  Break: '#a855f7',
  Breakdown: '#a855f7',
  Build: '#fbbf24',
  Body: '#3b82f6',
  Outro: '#eab308',
  Verse: '#60a5fa',
  Chorus: '#f472b6',
};

const CUE_COLORS: Record<string, string> = {
  Load: '#22c55e',
  FirstDownbeat: '#3b82f6',
  Drop: '#ef4444',
  Breakdown: '#a855f7',
  Build: '#f59e0b',
  OutroStart: '#eab308',
  SafetyLoop: '#06b6d4',
};

interface EditingState {
  mode: 'none' | 'creating' | 'resizing-start' | 'resizing-end' | 'moving';
  sectionIndex?: number;
  startBeat?: number;
  startMouseX?: number;
  originalStart?: number;
  originalEnd?: number;
}

export function WaveformCanvas({
  peaks,
  sections = [],
  cues = [],
  playheadPosition = 0,
  duration = 180,
  isPlaying = false,
  onSeek,
  onSectionsChange,
  height = 120,
  showBeatGrid = true,
  bpm = 128,
  editable = false,
}: WaveformCanvasProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const animationRef = useRef<number>(0);
  const [hoveredCue] = useState<Cue | null>(null);
  const [dimensions, setDimensions] = useState({ width: 800, height });
  const [editMode, setEditMode] = useState(false);
  const [selectedSection, setSelectedSection] = useState<number | null>(null);
  const [editingState, setEditingState] = useState<EditingState>({ mode: 'none' });
  const [showLabelMenu, setShowLabelMenu] = useState<{ x: number; y: number; index: number } | null>(null);
  const [hoveredSectionEdge, setHoveredSectionEdge] = useState<{ index: number; edge: 'start' | 'end' } | null>(null);

  const totalBeats = peaks.length > 0 ? peaks.length : 384;

  // Resize observer for responsive canvas
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { width } = entry.contentRect;
        setDimensions({ width, height });
      }
    });

    resizeObserver.observe(container);
    return () => resizeObserver.disconnect();
  }, [height]);

  // Convert pixel position to beat
  const pixelToBeat = useCallback((x: number) => {
    const beat = Math.round((x / dimensions.width) * totalBeats);
    return Math.max(0, Math.min(totalBeats, beat));
  }, [dimensions.width, totalBeats]);

  // Convert beat to pixel position
  const beatToPixel = useCallback((beat: number) => {
    return (beat / totalBeats) * dimensions.width;
  }, [dimensions.width, totalBeats]);

  // Check if mouse is near section edge
  const getSectionEdgeAtPosition = useCallback((x: number): { index: number; edge: 'start' | 'end' } | null => {
    const threshold = 8; // pixels
    for (let i = 0; i < sections.length; i++) {
      const startX = beatToPixel(sections[i].start);
      const endX = beatToPixel(sections[i].end);
      if (Math.abs(x - startX) < threshold) {
        return { index: i, edge: 'start' };
      }
      if (Math.abs(x - endX) < threshold) {
        return { index: i, edge: 'end' };
      }
    }
    return null;
  }, [sections, beatToPixel]);

  // Check if mouse is inside a section
  const getSectionAtPosition = useCallback((x: number): number | null => {
    const beat = pixelToBeat(x);
    for (let i = 0; i < sections.length; i++) {
      if (beat >= sections[i].start && beat <= sections[i].end) {
        return i;
      }
    }
    return null;
  }, [sections, pixelToBeat]);

  // High-performance canvas rendering
  const render = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d', { alpha: true });
    if (!ctx) return;

    const { width, height: h } = dimensions;
    const dpr = window.devicePixelRatio || 1;

    // Set canvas size with device pixel ratio for crisp rendering
    canvas.width = width * dpr;
    canvas.height = h * dpr;
    canvas.style.width = `${width}px`;
    canvas.style.height = `${h}px`;
    ctx.scale(dpr, dpr);

    // Clear
    ctx.clearRect(0, 0, width, h);

    // Draw sections background
    sections.forEach((section, i) => {
      const x = (section.start / totalBeats) * width;
      const w = ((section.end - section.start) / totalBeats) * width;
      ctx.fillStyle = SECTION_COLORS[section.label] || 'rgba(100, 100, 100, 0.1)';
      ctx.fillRect(x, 0, w, h);

      // Draw selection highlight
      if (editMode && selectedSection === i) {
        ctx.strokeStyle = SECTION_BORDER_COLORS[section.label] || '#a855f7';
        ctx.lineWidth = 2;
        ctx.strokeRect(x + 1, 1, w - 2, h - 2);
      }

      // Draw resize handles in edit mode
      if (editMode) {
        const handleWidth = 6;
        const handleColor = SECTION_BORDER_COLORS[section.label] || '#a855f7';

        // Start handle
        ctx.fillStyle = hoveredSectionEdge?.index === i && hoveredSectionEdge?.edge === 'start'
          ? handleColor : `${handleColor}88`;
        ctx.fillRect(x, 0, handleWidth, h);

        // End handle
        ctx.fillStyle = hoveredSectionEdge?.index === i && hoveredSectionEdge?.edge === 'end'
          ? handleColor : `${handleColor}88`;
        ctx.fillRect(x + w - handleWidth, 0, handleWidth, h);
      }
    });

    // Draw beat grid
    if (showBeatGrid && bpm > 0) {
      const beatsPerSecond = bpm / 60;
      const totalBeatsInDuration = beatsPerSecond * duration;
      const beatWidth = width / totalBeatsInDuration;

      ctx.strokeStyle = 'rgba(255, 255, 255, 0.08)';
      ctx.lineWidth = 1;

      for (let i = 0; i < totalBeatsInDuration; i++) {
        const x = i * beatWidth;
        if (i % 4 === 0) {
          ctx.strokeStyle = 'rgba(255, 255, 255, 0.15)';
        } else {
          ctx.strokeStyle = 'rgba(255, 255, 255, 0.05)';
        }
        ctx.beginPath();
        ctx.moveTo(x, 0);
        ctx.lineTo(x, h);
        ctx.stroke();
      }
    }

    // Draw waveform
    const barWidth = Math.max(2, width / peaks.length - 1);
    const centerY = h / 2;

    peaks.forEach((peak, i) => {
      const x = (i / peaks.length) * width;
      const normalizedPeak = peak / 10;
      const barHeight = normalizedPeak * (h * 0.8);

      const position = i / peaks.length;
      if (position < playheadPosition) {
        const gradient = ctx.createLinearGradient(x, centerY - barHeight / 2, x, centerY + barHeight / 2);
        gradient.addColorStop(0, 'rgba(167, 139, 250, 0.9)');
        gradient.addColorStop(0.5, 'rgba(139, 92, 246, 1)');
        gradient.addColorStop(1, 'rgba(167, 139, 250, 0.9)');
        ctx.fillStyle = gradient;
      } else {
        const gradient = ctx.createLinearGradient(x, centerY - barHeight / 2, x, centerY + barHeight / 2);
        gradient.addColorStop(0, 'rgba(59, 130, 246, 0.7)');
        gradient.addColorStop(0.5, 'rgba(59, 130, 246, 1)');
        gradient.addColorStop(1, 'rgba(59, 130, 246, 0.7)');
        ctx.fillStyle = gradient;
      }

      ctx.beginPath();
      ctx.roundRect(x, centerY - barHeight / 2, barWidth, barHeight, 1);
      ctx.fill();
    });

    // Draw cue markers
    cues.forEach((cue) => {
      const x = (cue.beat / totalBeats) * width;
      const color = CUE_COLORS[cue.type] || '#ffffff';

      ctx.strokeStyle = color;
      ctx.lineWidth = 2;
      ctx.beginPath();
      ctx.moveTo(x, 0);
      ctx.lineTo(x, h);
      ctx.stroke();

      ctx.fillStyle = color;
      ctx.beginPath();
      ctx.moveTo(x, 0);
      ctx.lineTo(x - 6, 12);
      ctx.lineTo(x + 6, 12);
      ctx.closePath();
      ctx.fill();
    });

    // Draw playhead
    if (playheadPosition > 0) {
      const playheadX = playheadPosition * width;

      ctx.shadowColor = '#a78bfa';
      ctx.shadowBlur = 10;
      ctx.strokeStyle = '#a78bfa';
      ctx.lineWidth = 2;
      ctx.beginPath();
      ctx.moveTo(playheadX, 0);
      ctx.lineTo(playheadX, h);
      ctx.stroke();
      ctx.shadowBlur = 0;

      ctx.fillStyle = '#a78bfa';
      ctx.beginPath();
      ctx.arc(playheadX, h - 8, 6, 0, Math.PI * 2);
      ctx.fill();
    }
  }, [peaks, sections, cues, playheadPosition, duration, dimensions, showBeatGrid, bpm, editMode, selectedSection, totalBeats, hoveredSectionEdge, editingState]);

  // Animation loop
  useEffect(() => {
    const animate = () => {
      render();
      if (isPlaying) {
        animationRef.current = requestAnimationFrame(animate);
      }
    };
    animate();
    return () => {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
      }
    };
  }, [render, isPlaying]);

  useEffect(() => {
    render();
  }, [render]);

  // Mouse handlers for section editing
  const handleMouseDown = (e: ReactMouseEvent<HTMLCanvasElement>) => {
    if (!canvasRef.current) return;
    const rect = canvasRef.current.getBoundingClientRect();
    const x = e.clientX - rect.left;

    if (editMode && onSectionsChange) {
      // Check if clicking on section edge for resize
      const edge = getSectionEdgeAtPosition(x);
      if (edge) {
        setEditingState({
          mode: edge.edge === 'start' ? 'resizing-start' : 'resizing-end',
          sectionIndex: edge.index,
          startMouseX: x,
          originalStart: sections[edge.index].start,
          originalEnd: sections[edge.index].end,
        });
        return;
      }

      // Check if clicking inside a section for selection
      const sectionIndex = getSectionAtPosition(x);
      if (sectionIndex !== null) {
        setSelectedSection(sectionIndex);
        return;
      }

      // Start creating new section
      const beat = pixelToBeat(x);
      setEditingState({
        mode: 'creating',
        startBeat: beat,
        startMouseX: x,
      });
      setSelectedSection(null);
    } else if (onSeek) {
      const position = x / rect.width;
      onSeek(Math.max(0, Math.min(1, position)));
    }
  };

  const handleMouseMove = (e: ReactMouseEvent<HTMLCanvasElement>) => {
    if (!canvasRef.current) return;
    const rect = canvasRef.current.getBoundingClientRect();
    const x = e.clientX - rect.left;

    if (editMode) {
      // Update cursor based on hover
      const edge = getSectionEdgeAtPosition(x);
      setHoveredSectionEdge(edge);

      if (editingState.mode === 'resizing-start' && editingState.sectionIndex !== undefined) {
        const newStart = pixelToBeat(x);
        const newSections = [...sections];
        newSections[editingState.sectionIndex] = {
          ...newSections[editingState.sectionIndex],
          start: Math.min(newStart, newSections[editingState.sectionIndex].end - 4),
        };
        onSectionsChange?.(newSections);
      } else if (editingState.mode === 'resizing-end' && editingState.sectionIndex !== undefined) {
        const newEnd = pixelToBeat(x);
        const newSections = [...sections];
        newSections[editingState.sectionIndex] = {
          ...newSections[editingState.sectionIndex],
          end: Math.max(newEnd, newSections[editingState.sectionIndex].start + 4),
        };
        onSectionsChange?.(newSections);
      }
    }
  };

  const handleMouseUp = (e: ReactMouseEvent<HTMLCanvasElement>) => {
    if (!canvasRef.current) return;
    const rect = canvasRef.current.getBoundingClientRect();
    const x = e.clientX - rect.left;

    if (editingState.mode === 'creating' && editingState.startBeat !== undefined) {
      const endBeat = pixelToBeat(x);
      const start = Math.min(editingState.startBeat, endBeat);
      const end = Math.max(editingState.startBeat, endBeat);

      // Only create if selection is big enough (at least 4 beats)
      if (end - start >= 4) {
        const newSection: Section = {
          start,
          end,
          label: 'Body', // Default label
        };
        const newSections = [...sections, newSection].sort((a, b) => a.start - b.start);
        onSectionsChange?.(newSections);
        setSelectedSection(newSections.findIndex(s => s.start === start && s.end === end));
      }
    }

    setEditingState({ mode: 'none' });
  };

  const handleContextMenu = (e: ReactMouseEvent<HTMLCanvasElement>) => {
    e.preventDefault();
    if (!editMode || !canvasRef.current) return;

    const rect = canvasRef.current.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const sectionIndex = getSectionAtPosition(x);

    if (sectionIndex !== null) {
      setShowLabelMenu({ x: e.clientX, y: e.clientY, index: sectionIndex });
      setSelectedSection(sectionIndex);
    }
  };

  const handleLabelChange = (label: string) => {
    if (showLabelMenu && onSectionsChange) {
      const newSections = [...sections];
      newSections[showLabelMenu.index] = {
        ...newSections[showLabelMenu.index],
        label,
      };
      onSectionsChange(newSections);
    }
    setShowLabelMenu(null);
  };

  const handleDeleteSection = () => {
    if (selectedSection !== null && onSectionsChange) {
      const newSections = sections.filter((_, i) => i !== selectedSection);
      onSectionsChange(newSections);
      setSelectedSection(null);
    }
    setShowLabelMenu(null);
  };

  // Close menu on click outside
  useEffect(() => {
    const handleClickOutside = () => setShowLabelMenu(null);
    if (showLabelMenu) {
      document.addEventListener('click', handleClickOutside);
      return () => document.removeEventListener('click', handleClickOutside);
    }
  }, [showLabelMenu]);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Delete' || e.key === 'Backspace') {
        if (selectedSection !== null && editMode && onSectionsChange) {
          handleDeleteSection();
        }
      }
      if (e.key === 'Escape') {
        setSelectedSection(null);
        setShowLabelMenu(null);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [selectedSection, editMode, onSectionsChange]);

  const getCursor = () => {
    if (!editMode) return onSeek ? 'pointer' : 'default';
    if (hoveredSectionEdge) return 'ew-resize';
    if (editingState.mode === 'creating') return 'crosshair';
    return 'crosshair';
  };

  return (
    <motion.div
      ref={containerRef}
      className="waveform-container"
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      style={{
        position: 'relative',
        width: '100%',
        height,
        borderRadius: 8,
        overflow: 'hidden',
        background: 'var(--color-bg-tertiary)',
      }}
    >
      <canvas
        ref={canvasRef}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={() => {
          setHoveredSectionEdge(null);
          if (editingState.mode !== 'none') {
            setEditingState({ mode: 'none' });
          }
        }}
        onContextMenu={handleContextMenu}
        style={{
          cursor: getCursor(),
          display: 'block',
        }}
      />

      {/* Edit mode toggle */}
      {editable && (
        <button
          className={`waveform-edit-btn ${editMode ? 'active' : ''}`}
          onClick={() => {
            setEditMode(!editMode);
            setSelectedSection(null);
          }}
          title={editMode ? 'Exit edit mode' : 'Edit sections'}
        >
          {editMode ? '✓' : '✎'}
        </button>
      )}

      {/* Section labels */}
      <div
        style={{
          position: 'absolute',
          top: 4,
          left: 0,
          right: 0,
          display: 'flex',
          pointerEvents: editMode ? 'auto' : 'none',
        }}
      >
        {sections.map((section, i) => {
          const left = `${(section.start / totalBeats) * 100}%`;
          return (
            <span
              key={i}
              onClick={() => editMode && setSelectedSection(i)}
              style={{
                position: 'absolute',
                left,
                fontSize: '0.65rem',
                fontWeight: 500,
                color: selectedSection === i ? '#fff' : 'var(--color-text-secondary)',
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
                padding: '2px 6px',
                background: selectedSection === i
                  ? SECTION_BORDER_COLORS[section.label] || '#a855f7'
                  : 'rgba(0,0,0,0.5)',
                borderRadius: 3,
                cursor: editMode ? 'pointer' : 'default',
                transition: 'all 0.15s',
              }}
            >
              {section.label}
            </span>
          );
        })}
      </div>

      {/* Label selection menu */}
      <AnimatePresence>
        {showLabelMenu && (
          <motion.div
            initial={{ opacity: 0, scale: 0.9 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.9 }}
            className="section-label-menu"
            style={{
              position: 'fixed',
              left: showLabelMenu.x,
              top: showLabelMenu.y,
              zIndex: 1000,
            }}
            onClick={(e) => e.stopPropagation()}
          >
            <div className="menu-title">Section Type</div>
            {SECTION_LABELS.map((label) => (
              <button
                key={label}
                className={`menu-item ${sections[showLabelMenu.index]?.label === label ? 'active' : ''}`}
                onClick={() => handleLabelChange(label)}
                style={{
                  borderLeft: `3px solid ${SECTION_BORDER_COLORS[label] || '#666'}`,
                }}
              >
                {label}
              </button>
            ))}
            <div className="menu-divider" />
            <button className="menu-item delete" onClick={handleDeleteSection}>
              Delete Section
            </button>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Cue tooltips */}
      {hoveredCue && (
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          style={{
            position: 'absolute',
            top: 20,
            left: `calc(${(hoveredCue.beat / totalBeats) * 100}% - 40px)`,
            background: 'var(--color-bg-secondary)',
            border: '1px solid var(--color-border)',
            borderRadius: 4,
            padding: '4px 8px',
            fontSize: '0.75rem',
            pointerEvents: 'none',
            zIndex: 10,
          }}
        >
          <strong>{hoveredCue.label}</strong>
          <br />
          Beat {hoveredCue.beat}
        </motion.div>
      )}

      {/* Time markers */}
      <div
        style={{
          position: 'absolute',
          bottom: 2,
          left: 4,
          right: 4,
          display: 'flex',
          justifyContent: 'space-between',
          fontSize: '0.6rem',
          color: 'var(--color-text-muted)',
          pointerEvents: 'none',
        }}
      >
        <span>0:00</span>
        <span>{Math.floor(duration / 60)}:{String(duration % 60).padStart(2, '0')}</span>
      </div>

      {/* Edit mode indicator */}
      {editMode && (
        <div className="edit-mode-indicator">
          EDIT MODE · Click to select · Drag to create · Right-click for options
        </div>
      )}
    </motion.div>
  );
}
