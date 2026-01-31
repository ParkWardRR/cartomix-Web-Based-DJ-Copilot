import { useMemo } from 'react';
import { motion } from 'framer-motion';
import type { Track } from '../types';

interface LibraryStatsProps {
  tracks: Track[];
}

interface KeyDistribution {
  key: string;
  count: number;
  percentage: number;
}

interface BPMRange {
  min: number;
  max: number;
  avg: number;
  median: number;
}

interface EnergyDistribution {
  level: number;
  count: number;
  percentage: number;
}

export function LibraryStats({ tracks }: LibraryStatsProps) {
  // Calculate key distribution
  const keyDistribution = useMemo((): KeyDistribution[] => {
    const counts: Record<string, number> = {};
    tracks.forEach(t => {
      counts[t.key] = (counts[t.key] || 0) + 1;
    });

    return Object.entries(counts)
      .map(([key, count]) => ({
        key,
        count,
        percentage: (count / tracks.length) * 100,
      }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 8); // Top 8 keys
  }, [tracks]);

  // Calculate BPM range
  const bpmRange = useMemo((): BPMRange => {
    if (tracks.length === 0) {
      return { min: 0, max: 0, avg: 0, median: 0 };
    }

    const bpms = tracks.map(t => t.bpm).sort((a, b) => a - b);
    const sum = bpms.reduce((a, b) => a + b, 0);

    return {
      min: bpms[0],
      max: bpms[bpms.length - 1],
      avg: Math.round(sum / bpms.length),
      median: bpms[Math.floor(bpms.length / 2)],
    };
  }, [tracks]);

  // Calculate energy distribution
  const energyDistribution = useMemo((): EnergyDistribution[] => {
    const counts: Record<number, number> = {};
    for (let i = 1; i <= 10; i++) counts[i] = 0;

    tracks.forEach(t => {
      const level = Math.round(t.energy);
      if (level >= 1 && level <= 10) {
        counts[level]++;
      }
    });

    return Object.entries(counts).map(([level, count]) => ({
      level: parseInt(level),
      count,
      percentage: tracks.length > 0 ? (count / tracks.length) * 100 : 0,
    }));
  }, [tracks]);

  // Calculate status breakdown
  const statusBreakdown = useMemo(() => {
    const analyzed = tracks.filter(t => t.status === 'analyzed').length;
    const pending = tracks.filter(t => t.status === 'pending').length;
    const failed = tracks.filter(t => t.status === 'failed').length;

    return { analyzed, pending, failed };
  }, [tracks]);

  // Average energy
  const avgEnergy = useMemo(() => {
    if (tracks.length === 0) return 0;
    const sum = tracks.reduce((a, t) => a + t.energy, 0);
    return (sum / tracks.length).toFixed(1);
  }, [tracks]);

  if (tracks.length === 0) {
    return (
      <div className="library-stats-empty">
        <p>No tracks in library</p>
      </div>
    );
  }

  return (
    <div className="library-stats">
      {/* Summary Cards */}
      <div className="stats-summary">
        <div className="stat-card">
          <span className="stat-value">{tracks.length}</span>
          <span className="stat-label">Total Tracks</span>
        </div>
        <div className="stat-card">
          <span className="stat-value">{statusBreakdown.analyzed}</span>
          <span className="stat-label">Analyzed</span>
        </div>
        <div className="stat-card">
          <span className="stat-value">{bpmRange.avg}</span>
          <span className="stat-label">Avg BPM</span>
        </div>
        <div className="stat-card">
          <span className="stat-value">{avgEnergy}</span>
          <span className="stat-label">Avg Energy</span>
        </div>
      </div>

      {/* BPM Range */}
      <div className="stats-section">
        <h4>BPM Range</h4>
        <div className="bpm-range-bar">
          <div className="range-labels">
            <span>{bpmRange.min}</span>
            <span className="range-median">{bpmRange.median}</span>
            <span>{bpmRange.max}</span>
          </div>
          <div className="range-track">
            <motion.div
              className="range-fill"
              initial={{ width: 0 }}
              animate={{ width: '100%' }}
              transition={{ duration: 0.5 }}
            />
            <div
              className="range-marker median"
              style={{ left: `${((bpmRange.median - bpmRange.min) / (bpmRange.max - bpmRange.min)) * 100}%` }}
            />
          </div>
        </div>
      </div>

      {/* Energy Distribution */}
      <div className="stats-section">
        <h4>Energy Distribution</h4>
        <div className="energy-bars">
          {energyDistribution.map((e) => (
            <div key={e.level} className="energy-bar-item">
              <span className="energy-level">{e.level}</span>
              <div className="energy-bar-track">
                <motion.div
                  className="energy-bar-fill"
                  initial={{ width: 0 }}
                  animate={{ width: `${e.percentage}%` }}
                  transition={{ duration: 0.3, delay: e.level * 0.03 }}
                  style={{
                    background: `hsl(${(e.level - 1) * 12}, 70%, 50%)`,
                  }}
                />
              </div>
              <span className="energy-count">{e.count}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Top Keys */}
      <div className="stats-section">
        <h4>Top Keys</h4>
        <div className="key-distribution">
          {keyDistribution.map((k, i) => (
            <motion.div
              key={k.key}
              className="key-item"
              initial={{ opacity: 0, x: -10 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: i * 0.05 }}
            >
              <span className="key-name">{k.key}</span>
              <div className="key-bar-track">
                <motion.div
                  className="key-bar-fill"
                  initial={{ width: 0 }}
                  animate={{ width: `${k.percentage}%` }}
                  transition={{ duration: 0.4, delay: i * 0.05 }}
                />
              </div>
              <span className="key-count">{k.count}</span>
            </motion.div>
          ))}
        </div>
      </div>

      {/* Status Breakdown */}
      {(statusBreakdown.pending > 0 || statusBreakdown.failed > 0) && (
        <div className="stats-section">
          <h4>Analysis Status</h4>
          <div className="status-breakdown">
            <div className="status-item analyzed">
              <span className="status-dot" />
              <span className="status-label">Analyzed</span>
              <span className="status-count">{statusBreakdown.analyzed}</span>
            </div>
            {statusBreakdown.pending > 0 && (
              <div className="status-item pending">
                <span className="status-dot" />
                <span className="status-label">Pending</span>
                <span className="status-count">{statusBreakdown.pending}</span>
              </div>
            )}
            {statusBreakdown.failed > 0 && (
              <div className="status-item failed">
                <span className="status-dot" />
                <span className="status-label">Failed</span>
                <span className="status-count">{statusBreakdown.failed}</span>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
