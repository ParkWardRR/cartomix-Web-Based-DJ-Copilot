import { useCallback, useEffect, useRef, useState } from 'react';
import { motion } from 'framer-motion';
import type { Cue, Section } from '../types';

interface WaveformCanvasProps {
  peaks: number[];
  sections?: Section[];
  cues?: Cue[];
  playheadPosition?: number; // 0-1
  duration?: number;
  isPlaying?: boolean;
  onSeek?: (position: number) => void;
  height?: number;
  showBeatGrid?: boolean;
  bpm?: number;
}

const SECTION_COLORS: Record<string, string> = {
  Intro: 'rgba(34, 197, 94, 0.25)',
  Drop: 'rgba(239, 68, 68, 0.25)',
  Break: 'rgba(168, 85, 247, 0.25)',
  Breakdown: 'rgba(168, 85, 247, 0.25)',
  Build: 'rgba(251, 191, 36, 0.25)',
  'Build/Drop': 'rgba(251, 191, 36, 0.25)',
  Body: 'rgba(59, 130, 246, 0.15)',
  Outro: 'rgba(234, 179, 8, 0.25)',
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

export function WaveformCanvas({
  peaks,
  sections = [],
  cues = [],
  playheadPosition = 0,
  duration = 180,
  isPlaying = false,
  onSeek,
  height = 120,
  showBeatGrid = true,
  bpm = 128,
}: WaveformCanvasProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const animationRef = useRef<number>(0);
  const [hoveredCue] = useState<Cue | null>(null);
  const [dimensions, setDimensions] = useState({ width: 800, height });

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
    const totalBeats = peaks.length > 0 ? peaks.length : 384;
    sections.forEach((section) => {
      const x = (section.start / totalBeats) * width;
      const w = ((section.end - section.start) / totalBeats) * width;
      ctx.fillStyle = SECTION_COLORS[section.label] || 'rgba(100, 100, 100, 0.1)';
      ctx.fillRect(x, 0, w, h);
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
        // Stronger line every 4 beats (bar)
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

      // Gradient based on position relative to playhead
      const position = i / peaks.length;
      if (position < playheadPosition) {
        // Played portion - accent color
        const gradient = ctx.createLinearGradient(x, centerY - barHeight / 2, x, centerY + barHeight / 2);
        gradient.addColorStop(0, 'rgba(167, 139, 250, 0.9)');
        gradient.addColorStop(0.5, 'rgba(139, 92, 246, 1)');
        gradient.addColorStop(1, 'rgba(167, 139, 250, 0.9)');
        ctx.fillStyle = gradient;
      } else {
        // Unplayed portion - primary color
        const gradient = ctx.createLinearGradient(x, centerY - barHeight / 2, x, centerY + barHeight / 2);
        gradient.addColorStop(0, 'rgba(59, 130, 246, 0.7)');
        gradient.addColorStop(0.5, 'rgba(59, 130, 246, 1)');
        gradient.addColorStop(1, 'rgba(59, 130, 246, 0.7)');
        ctx.fillStyle = gradient;
      }

      // Draw mirrored bars for stereo effect
      ctx.beginPath();
      ctx.roundRect(x, centerY - barHeight / 2, barWidth, barHeight, 1);
      ctx.fill();
    });

    // Draw cue markers
    cues.forEach((cue) => {
      const x = (cue.beat / totalBeats) * width;
      const color = CUE_COLORS[cue.type] || '#ffffff';

      // Cue line
      ctx.strokeStyle = color;
      ctx.lineWidth = 2;
      ctx.beginPath();
      ctx.moveTo(x, 0);
      ctx.lineTo(x, h);
      ctx.stroke();

      // Cue triangle marker at top
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

      // Glow effect
      ctx.shadowColor = '#a78bfa';
      ctx.shadowBlur = 10;
      ctx.strokeStyle = '#a78bfa';
      ctx.lineWidth = 2;
      ctx.beginPath();
      ctx.moveTo(playheadX, 0);
      ctx.lineTo(playheadX, h);
      ctx.stroke();
      ctx.shadowBlur = 0;

      // Playhead handle
      ctx.fillStyle = '#a78bfa';
      ctx.beginPath();
      ctx.arc(playheadX, h - 8, 6, 0, Math.PI * 2);
      ctx.fill();
    }
  }, [peaks, sections, cues, playheadPosition, duration, dimensions, showBeatGrid, bpm]);

  // Animation loop for smooth playhead
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

  // Re-render on state changes
  useEffect(() => {
    render();
  }, [render]);

  const handleClick = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!onSeek || !canvasRef.current) return;
    const rect = canvasRef.current.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const position = x / rect.width;
    onSeek(Math.max(0, Math.min(1, position)));
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
        onClick={handleClick}
        style={{
          cursor: onSeek ? 'pointer' : 'default',
          display: 'block',
        }}
      />

      {/* Section labels */}
      <div
        style={{
          position: 'absolute',
          top: 4,
          left: 0,
          right: 0,
          display: 'flex',
          pointerEvents: 'none',
        }}
      >
        {sections.map((section, i) => {
          const totalBeats = peaks.length > 0 ? peaks.length : 384;
          const left = `${(section.start / totalBeats) * 100}%`;
          return (
            <span
              key={i}
              style={{
                position: 'absolute',
                left,
                fontSize: '0.65rem',
                fontWeight: 500,
                color: 'var(--color-text-secondary)',
                textTransform: 'uppercase',
                letterSpacing: '0.05em',
                padding: '2px 4px',
                background: 'rgba(0,0,0,0.4)',
                borderRadius: 2,
              }}
            >
              {section.label}
            </span>
          );
        })}
      </div>

      {/* Cue tooltips */}
      {hoveredCue && (
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          style={{
            position: 'absolute',
            top: 20,
            left: `calc(${(hoveredCue.beat / (peaks.length || 384)) * 100}% - 40px)`,
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
    </motion.div>
  );
}
