import { motion, useSpring, useTransform } from 'framer-motion';
import { useEffect, useState } from 'react';

interface StatCardProps {
  label: string;
  value: number;
  suffix?: string;
  color?: string;
  icon?: string;
  trend?: 'up' | 'down' | 'neutral';
  animate?: boolean;
}

function AnimatedNumber({ value, suffix = '' }: { value: number; suffix?: string }) {
  const spring = useSpring(0, { stiffness: 100, damping: 20 });
  const display = useTransform(spring, (v) =>
    suffix === '%' || suffix.includes('.') ? v.toFixed(1) : Math.round(v).toString()
  );
  const [displayValue, setDisplayValue] = useState('0');

  useEffect(() => {
    spring.set(value);
  }, [spring, value]);

  useEffect(() => {
    return display.on('change', (v) => setDisplayValue(v));
  }, [display]);

  return (
    <span>
      {displayValue}
      {suffix}
    </span>
  );
}

function StatCard({ label, value, suffix = '', color, icon, trend, animate = true }: StatCardProps) {
  const trendIcon = trend === 'up' ? '↑' : trend === 'down' ? '↓' : '';
  const trendColor = trend === 'up' ? 'var(--color-success)' : trend === 'down' ? 'var(--color-error)' : '';

  return (
    <motion.div
      initial={animate ? { opacity: 0, y: 20, scale: 0.95 } : undefined}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      transition={{ duration: 0.4 }}
      whileHover={{ scale: 1.02, y: -2 }}
      style={{
        background: 'var(--color-bg-tertiary)',
        border: '1px solid var(--color-border)',
        borderRadius: 12,
        padding: '16px 20px',
        display: 'flex',
        flexDirection: 'column',
        gap: 8,
        position: 'relative',
        overflow: 'hidden',
      }}
    >
      {/* Background glow */}
      {color && (
        <div
          style={{
            position: 'absolute',
            top: -20,
            right: -20,
            width: 80,
            height: 80,
            background: color,
            opacity: 0.1,
            borderRadius: '50%',
            filter: 'blur(20px)',
          }}
        />
      )}

      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <span
          style={{
            fontSize: '0.75rem',
            color: 'var(--color-text-secondary)',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            fontWeight: 500,
          }}
        >
          {icon && <span style={{ marginRight: 6 }}>{icon}</span>}
          {label}
        </span>
        {trend && (
          <span style={{ color: trendColor, fontSize: '0.7rem', fontWeight: 600 }}>{trendIcon}</span>
        )}
      </div>

      <div
        style={{
          fontSize: '1.75rem',
          fontWeight: 700,
          color: color || 'var(--color-text)',
          fontVariantNumeric: 'tabular-nums',
        }}
      >
        {animate ? <AnimatedNumber value={value} suffix={suffix} /> : `${value}${suffix}`}
      </div>
    </motion.div>
  );
}

interface ProgressRingProps {
  value: number; // 0-100
  size?: number;
  strokeWidth?: number;
  color?: string;
  label?: string;
}

function ProgressRing({ value, size = 80, strokeWidth = 6, color = 'var(--color-primary)', label }: ProgressRingProps) {
  const radius = (size - strokeWidth) / 2;
  const circumference = radius * 2 * Math.PI;
  const offset = circumference - (value / 100) * circumference;

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.8 }}
      animate={{ opacity: 1, scale: 1 }}
      style={{
        position: 'relative',
        width: size,
        height: size,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <svg width={size} height={size} style={{ transform: 'rotate(-90deg)' }}>
        {/* Background circle */}
        <circle cx={size / 2} cy={size / 2} r={radius} fill="none" stroke="var(--color-bg-tertiary)" strokeWidth={strokeWidth} />
        {/* Progress circle */}
        <motion.circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke={color}
          strokeWidth={strokeWidth}
          strokeLinecap="round"
          strokeDasharray={circumference}
          initial={{ strokeDashoffset: circumference }}
          animate={{ strokeDashoffset: offset }}
          transition={{ duration: 1, ease: 'easeOut' }}
        />
      </svg>
      <div
        style={{
          position: 'absolute',
          textAlign: 'center',
        }}
      >
        <div style={{ fontSize: '1rem', fontWeight: 700, color }}>{Math.round(value)}%</div>
        {label && (
          <div style={{ fontSize: '0.55rem', color: 'var(--color-text-muted)', marginTop: 2 }}>{label}</div>
        )}
      </div>
    </motion.div>
  );
}

interface LiveStatsProps {
  stats: {
    analyzed: number;
    pending: number;
    failed?: number;
    avgBpm: number;
    avgEnergy: number;
    keyCount: number;
    avgEdgeScore: number;
    totalTracks: number;
  };
  animate?: boolean;
}

export function LiveStats({ stats, animate = true }: LiveStatsProps) {
  const analysisProgress = stats.totalTracks > 0 ? (stats.analyzed / stats.totalTracks) * 100 : 0;

  return (
    <motion.div
      initial={animate ? { opacity: 0 } : undefined}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.5, staggerChildren: 0.1 }}
      style={{
        display: 'flex',
        flexDirection: 'column',
        gap: 16,
      }}
    >
      {/* Top row - Key metrics */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(140px, 1fr))',
          gap: 12,
        }}
      >
        <StatCard label="Analyzed" value={stats.analyzed} icon="✓" color="var(--color-success)" animate={animate} />
        <StatCard label="Pending" value={stats.pending} icon="◷" color="var(--color-warning)" animate={animate} />
        <StatCard label="Avg BPM" value={stats.avgBpm} color="var(--color-primary)" animate={animate} />
        <StatCard label="Avg Energy" value={stats.avgEnergy} suffix="/10" color="var(--color-accent)" animate={animate} />
      </div>

      {/* Bottom row - Secondary metrics + progress */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(120px, 1fr))',
          gap: 12,
          alignItems: 'center',
        }}
      >
        <StatCard label="Keys" value={stats.keyCount} icon="♪" animate={animate} />
        <StatCard
          label="Edge Score"
          value={stats.avgEdgeScore}
          suffix="/10"
          color="var(--color-primary)"
          trend={stats.avgEdgeScore >= 7 ? 'up' : stats.avgEdgeScore < 5 ? 'down' : 'neutral'}
          animate={animate}
        />

        {/* Analysis progress ring */}
        <motion.div
          initial={animate ? { opacity: 0, scale: 0.8 } : undefined}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: 0.3 }}
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            background: 'var(--color-bg-tertiary)',
            border: '1px solid var(--color-border)',
            borderRadius: 12,
            padding: 16,
          }}
        >
          <ProgressRing
            value={analysisProgress}
            color={analysisProgress === 100 ? 'var(--color-success)' : 'var(--color-primary)'}
            label="Analyzed"
          />
        </motion.div>
      </div>
    </motion.div>
  );
}

export { StatCard, ProgressRing };
