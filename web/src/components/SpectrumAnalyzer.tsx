import { useCallback, useEffect, useRef, useState } from 'react';
import { motion } from 'framer-motion';

interface SpectrumAnalyzerProps {
  bands?: number;
  height?: number;
  isActive?: boolean;
  simulatedData?: number[];
  colorScheme?: 'primary' | 'accent' | 'gradient';
  style?: 'bars' | 'mirror' | 'circular';
}

// Generate realistic-looking frequency distribution
function generateSpectrumData(bands: number, seed: number): number[] {
  const data: number[] = [];
  const time = Date.now() / 1000 + seed;

  for (let i = 0; i < bands; i++) {
    // Lower frequencies typically have more energy
    const freqFalloff = 1 - (i / bands) * 0.6;
    // Add some noise and variation
    const noise = Math.sin(time * 3 + i * 0.5) * 0.3 + Math.sin(time * 7 + i * 0.2) * 0.2;
    // Occasional peaks (like kick drums)
    const kick = i < 4 ? Math.max(0, Math.sin(time * 2) * 0.5) : 0;
    // High hat region
    const hihat = i > bands * 0.7 ? Math.max(0, Math.sin(time * 8 + i) * 0.3) : 0;

    const value = Math.max(0.05, Math.min(1, freqFalloff * 0.7 + noise + kick + hihat));
    data.push(value);
  }

  return data;
}

export function SpectrumAnalyzer({
  bands = 32,
  height = 80,
  isActive = true,
  simulatedData,
  colorScheme = 'gradient',
  style = 'mirror',
}: SpectrumAnalyzerProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const animationRef = useRef<number>(0);
  const seedRef = useRef(Math.random() * 1000);
  const [dimensions, setDimensions] = useState({ width: 400, height });

  // Resize observer
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

  const getColor = useCallback(
    (index: number, total: number, value: number): string => {
      switch (colorScheme) {
        case 'primary':
          return `rgba(59, 130, 246, ${0.5 + value * 0.5})`;
        case 'accent':
          return `rgba(167, 139, 250, ${0.5 + value * 0.5})`;
        case 'gradient':
        default: {
          // Gradient from blue through purple to pink based on frequency
          const hue = 220 + (index / total) * 80;
          const saturation = 70 + value * 30;
          const lightness = 50 + value * 20;
          return `hsla(${hue}, ${saturation}%, ${lightness}%, ${0.7 + value * 0.3})`;
        }
      }
    },
    [colorScheme]
  );

  const render = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d', { alpha: true });
    if (!ctx) return;

    const { width, height: h } = dimensions;
    const dpr = window.devicePixelRatio || 1;

    canvas.width = width * dpr;
    canvas.height = h * dpr;
    canvas.style.width = `${width}px`;
    canvas.style.height = `${h}px`;
    ctx.scale(dpr, dpr);

    ctx.clearRect(0, 0, width, h);

    const data = simulatedData || generateSpectrumData(bands, seedRef.current);
    const barWidth = (width / bands) * 0.8;
    const gap = (width / bands) * 0.2;
    const centerY = h / 2;

    data.forEach((value, i) => {
      const x = i * (barWidth + gap) + gap / 2;
      const color = getColor(i, bands, value);

      if (style === 'mirror') {
        // Mirror style - bars from center
        const barHeight = value * (h * 0.45);

        // Top half
        const gradientTop = ctx.createLinearGradient(x, centerY, x, centerY - barHeight);
        gradientTop.addColorStop(0, color);
        gradientTop.addColorStop(1, color.replace(/[\d.]+\)$/, '0.3)'));
        ctx.fillStyle = gradientTop;
        ctx.beginPath();
        ctx.roundRect(x, centerY - barHeight, barWidth, barHeight, [2, 2, 0, 0]);
        ctx.fill();

        // Bottom half (mirrored)
        const gradientBottom = ctx.createLinearGradient(x, centerY, x, centerY + barHeight);
        gradientBottom.addColorStop(0, color);
        gradientBottom.addColorStop(1, color.replace(/[\d.]+\)$/, '0.3)'));
        ctx.fillStyle = gradientBottom;
        ctx.beginPath();
        ctx.roundRect(x, centerY, barWidth, barHeight, [0, 0, 2, 2]);
        ctx.fill();
      } else if (style === 'circular') {
        // Circular style - radial bars
        const cx = width / 2;
        const cy = h / 2;
        const innerRadius = Math.min(width, h) * 0.2;
        const outerRadius = innerRadius + value * Math.min(width, h) * 0.3;
        const angle = (i / bands) * Math.PI * 2 - Math.PI / 2;
        const angleWidth = (Math.PI * 2) / bands * 0.8;

        ctx.fillStyle = color;
        ctx.beginPath();
        ctx.arc(cx, cy, outerRadius, angle - angleWidth / 2, angle + angleWidth / 2);
        ctx.arc(cx, cy, innerRadius, angle + angleWidth / 2, angle - angleWidth / 2, true);
        ctx.closePath();
        ctx.fill();
      } else {
        // Standard bars
        const barHeight = value * (h * 0.9);

        const gradient = ctx.createLinearGradient(x, h, x, h - barHeight);
        gradient.addColorStop(0, color);
        gradient.addColorStop(1, color.replace(/[\d.]+\)$/, '0.5)'));
        ctx.fillStyle = gradient;
        ctx.beginPath();
        ctx.roundRect(x, h - barHeight, barWidth, barHeight, [2, 2, 0, 0]);
        ctx.fill();
      }
    });

    // Add glow effect overlay
    if (style === 'mirror') {
      const glowGradient = ctx.createRadialGradient(width / 2, centerY, 0, width / 2, centerY, width / 2);
      glowGradient.addColorStop(0, 'rgba(167, 139, 250, 0.1)');
      glowGradient.addColorStop(1, 'rgba(167, 139, 250, 0)');
      ctx.fillStyle = glowGradient;
      ctx.fillRect(0, 0, width, h);
    }
  }, [dimensions, bands, simulatedData, getColor, style]);

  // Animation loop
  useEffect(() => {
    if (!isActive) {
      render();
      return;
    }

    const animate = () => {
      seedRef.current += 0.016; // ~60fps time step
      render();
      animationRef.current = requestAnimationFrame(animate);
    };

    animate();

    return () => {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
      }
    };
  }, [render, isActive]);

  return (
    <motion.div
      ref={containerRef}
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.4 }}
      style={{
        width: '100%',
        height,
        borderRadius: 8,
        overflow: 'hidden',
        background: 'var(--color-bg-tertiary)',
        position: 'relative',
      }}
    >
      <canvas
        ref={canvasRef}
        style={{
          display: 'block',
        }}
      />

      {/* Frequency labels */}
      <div
        style={{
          position: 'absolute',
          bottom: 2,
          left: 8,
          right: 8,
          display: 'flex',
          justifyContent: 'space-between',
          fontSize: '0.55rem',
          color: 'var(--color-text-muted)',
          pointerEvents: 'none',
          opacity: 0.6,
        }}
      >
        <span>20Hz</span>
        <span>200Hz</span>
        <span>2kHz</span>
        <span>20kHz</span>
      </div>
    </motion.div>
  );
}
