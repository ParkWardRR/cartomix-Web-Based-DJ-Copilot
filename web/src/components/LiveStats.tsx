import { motion, useSpring, useTransform } from 'framer-motion';
import { useEffect, useState } from 'react';

interface StatCardProps {
  label: string;
  value: number;
  suffix?: string;
  color?: string;
  animate?: boolean;
  compact?: boolean;
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

function StatCard({ label, value, suffix = '', color, animate = true, compact }: StatCardProps) {
  return (
    <motion.div
      initial={animate ? { opacity: 0, scale: 0.95 } : undefined}
      animate={{ opacity: 1, scale: 1 }}
      className="stat-item"
      style={{
        background: color ? `linear-gradient(135deg, var(--color-bg-tertiary), ${color}10)` : undefined,
      }}
    >
      <div className="stat-value" style={{ color: color || 'var(--color-primary)', fontSize: compact ? '0.9rem' : '1rem' }}>
        {animate ? <AnimatedNumber value={value} suffix={suffix} /> : `${value}${suffix}`}
      </div>
      <div className="stat-label">{label}</div>
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
  compact?: boolean;
}

export function LiveStats({ stats, animate = true, compact }: LiveStatsProps) {
  return (
    <motion.div
      initial={animate ? { opacity: 0 } : undefined}
      animate={{ opacity: 1 }}
      className={`live-stats ${compact ? 'compact' : ''}`}
    >
      <StatCard label="Tracks" value={stats.totalTracks} color="var(--color-primary)" animate={animate} compact={compact} />
      <StatCard label="Analyzed" value={stats.analyzed} color="var(--color-success)" animate={animate} compact={compact} />
      <StatCard label="BPM" value={stats.avgBpm} color="var(--color-accent)" animate={animate} compact={compact} />
      <StatCard label="Energy" value={stats.avgEnergy} suffix="/10" animate={animate} compact={compact} />
      <StatCard label="Keys" value={stats.keyCount} animate={animate} compact={compact} />
      <StatCard label="Score" value={stats.avgEdgeScore} suffix="/10" color="var(--color-primary)" animate={animate} compact={compact} />
    </motion.div>
  );
}

export { StatCard };
