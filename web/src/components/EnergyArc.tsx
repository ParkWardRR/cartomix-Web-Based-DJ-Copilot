import { motion } from 'framer-motion';
import { useMemo } from 'react';

interface EnergyArcProps {
  values: number[]; // Array of energy values 1-10
  labels?: string[];
  height?: number;
  showLabels?: boolean;
  highlightIndex?: number;
  animate?: boolean;
}

export function EnergyArc({
  values,
  labels = [],
  height = 100,
  showLabels = true,
  highlightIndex,
  animate = true,
}: EnergyArcProps) {
  // Generate smooth curve points using cubic bezier interpolation
  const pathData = useMemo(() => {
    if (values.length < 2) return '';

    const points = values.map((v, i) => ({
      x: (i / (values.length - 1)) * 100,
      y: 100 - (v / 10) * 85 - 5, // Scale to leave padding
    }));

    // Create smooth curve using bezier
    let path = `M ${points[0].x} ${points[0].y}`;

    for (let i = 0; i < points.length - 1; i++) {
      const current = points[i];
      const next = points[i + 1];
      const controlX = (current.x + next.x) / 2;

      path += ` C ${controlX} ${current.y}, ${controlX} ${next.y}, ${next.x} ${next.y}`;
    }

    return path;
  }, [values]);

  // Fill area under curve
  const fillPath = useMemo(() => {
    if (!pathData) return '';
    return `${pathData} L 100 100 L 0 100 Z`;
  }, [pathData]);

  // Energy level colors
  const getEnergyColor = (value: number): string => {
    if (value >= 8) return 'var(--color-error)';
    if (value >= 6) return 'var(--color-warning)';
    if (value >= 4) return 'var(--color-primary)';
    return 'var(--color-success)';
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4 }}
      style={{
        width: '100%',
        height,
        position: 'relative',
        background: 'var(--color-bg-tertiary)',
        borderRadius: 8,
        padding: '8px 12px',
        boxSizing: 'border-box',
      }}
    >
      {/* Y-axis labels */}
      <div
        style={{
          position: 'absolute',
          left: 4,
          top: 8,
          bottom: showLabels ? 24 : 8,
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'space-between',
          fontSize: '0.55rem',
          color: 'var(--color-text-muted)',
          width: 12,
        }}
      >
        <span>10</span>
        <span>5</span>
        <span>1</span>
      </div>

      {/* SVG Chart */}
      <svg
        viewBox="0 0 100 100"
        preserveAspectRatio="none"
        style={{
          width: 'calc(100% - 20px)',
          height: showLabels ? 'calc(100% - 20px)' : '100%',
          marginLeft: 16,
        }}
      >
        {/* Grid lines */}
        <defs>
          <linearGradient id="energyGradient" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stopColor="var(--color-error)" stopOpacity="0.6" />
            <stop offset="50%" stopColor="var(--color-warning)" stopOpacity="0.4" />
            <stop offset="100%" stopColor="var(--color-success)" stopOpacity="0.2" />
          </linearGradient>
          <linearGradient id="lineGradient" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="var(--color-primary)" />
            <stop offset="50%" stopColor="var(--color-accent)" />
            <stop offset="100%" stopColor="var(--color-error)" />
          </linearGradient>
        </defs>

        {/* Horizontal grid lines */}
        {[15, 32.5, 50, 67.5, 85].map((y, i) => (
          <line
            key={i}
            x1="0"
            y1={y}
            x2="100"
            y2={y}
            stroke="var(--color-border)"
            strokeWidth="0.3"
            strokeDasharray="2,2"
          />
        ))}

        {/* Area fill */}
        <motion.path
          d={fillPath}
          fill="url(#energyGradient)"
          initial={animate ? { opacity: 0 } : undefined}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.8, delay: 0.2 }}
        />

        {/* Main curve line */}
        <motion.path
          d={pathData}
          fill="none"
          stroke="url(#lineGradient)"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          initial={animate ? { pathLength: 0 } : undefined}
          animate={{ pathLength: 1 }}
          transition={{ duration: 1, ease: 'easeOut' }}
        />

        {/* Data points */}
        {values.map((v, i) => {
          const x = (i / (values.length - 1)) * 100;
          const y = 100 - (v / 10) * 85 - 5;
          const isHighlighted = highlightIndex === i;

          return (
            <motion.g key={i}>
              {/* Glow effect for highlighted */}
              {isHighlighted && (
                <motion.circle
                  cx={x}
                  cy={y}
                  r="6"
                  fill={getEnergyColor(v)}
                  opacity="0.3"
                  initial={{ scale: 0 }}
                  animate={{ scale: [1, 1.5, 1] }}
                  transition={{ duration: 1, repeat: Infinity }}
                />
              )}

              {/* Point */}
              <motion.circle
                cx={x}
                cy={y}
                r={isHighlighted ? 4 : 3}
                fill={getEnergyColor(v)}
                stroke="var(--color-bg)"
                strokeWidth="1"
                initial={animate ? { scale: 0 } : undefined}
                animate={{ scale: 1 }}
                transition={{ delay: i * 0.05, duration: 0.3 }}
                style={{ cursor: 'pointer' }}
              />

              {/* Value label on hover area */}
              <title>
                {labels[i] || `Track ${i + 1}`}: Energy {v}
              </title>
            </motion.g>
          );
        })}
      </svg>

      {/* X-axis labels */}
      {showLabels && labels.length > 0 && (
        <div
          style={{
            position: 'absolute',
            bottom: 4,
            left: 20,
            right: 8,
            display: 'flex',
            justifyContent: 'space-between',
            fontSize: '0.55rem',
            color: 'var(--color-text-muted)',
            overflow: 'hidden',
          }}
        >
          {labels.slice(0, 8).map((label, i) => (
            <span
              key={i}
              style={{
                textOverflow: 'ellipsis',
                overflow: 'hidden',
                whiteSpace: 'nowrap',
                maxWidth: `${100 / Math.min(labels.length, 8)}%`,
                textAlign: 'center',
                fontWeight: highlightIndex === i ? 600 : 400,
                color: highlightIndex === i ? 'var(--color-text)' : undefined,
              }}
            >
              {label.slice(0, 8)}
            </span>
          ))}
        </div>
      )}

      {/* Legend */}
      <div
        style={{
          position: 'absolute',
          top: 4,
          right: 8,
          display: 'flex',
          gap: 8,
          fontSize: '0.55rem',
        }}
      >
        <span style={{ color: 'var(--color-success)' }}>Low</span>
        <span style={{ color: 'var(--color-warning)' }}>Mid</span>
        <span style={{ color: 'var(--color-error)' }}>Peak</span>
      </div>
    </motion.div>
  );
}
