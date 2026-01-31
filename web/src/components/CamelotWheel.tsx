import { useMemo } from 'react';
import { motion } from 'framer-motion';

interface CamelotWheelProps {
  selectedKey?: string;
  tracks?: { key: string; count: number }[];
  onKeySelect?: (key: string) => void;
  size?: number;
}

// Camelot wheel positions (clockwise from 12 o'clock)
const CAMELOT_KEYS = [
  { camelot: '12B', openKey: '1d', major: 'E', position: 0 },
  { camelot: '1B', openKey: '2d', major: 'B', position: 1 },
  { camelot: '2B', openKey: '3d', major: 'F#', position: 2 },
  { camelot: '3B', openKey: '4d', major: 'Db', position: 3 },
  { camelot: '4B', openKey: '5d', major: 'Ab', position: 4 },
  { camelot: '5B', openKey: '6d', major: 'Eb', position: 5 },
  { camelot: '6B', openKey: '7d', major: 'Bb', position: 6 },
  { camelot: '7B', openKey: '8d', major: 'F', position: 7 },
  { camelot: '8B', openKey: '9d', major: 'C', position: 8 },
  { camelot: '9B', openKey: '10d', major: 'G', position: 9 },
  { camelot: '10B', openKey: '11d', major: 'D', position: 10 },
  { camelot: '11B', openKey: '12d', major: 'A', position: 11 },
  { camelot: '12A', openKey: '1m', minor: 'Dbm', position: 0 },
  { camelot: '1A', openKey: '2m', minor: 'Abm', position: 1 },
  { camelot: '2A', openKey: '3m', minor: 'Ebm', position: 2 },
  { camelot: '3A', openKey: '4m', minor: 'Bbm', position: 3 },
  { camelot: '4A', openKey: '5m', minor: 'Fm', position: 4 },
  { camelot: '5A', openKey: '6m', minor: 'Cm', position: 5 },
  { camelot: '6A', openKey: '7m', minor: 'Gm', position: 6 },
  { camelot: '7A', openKey: '8m', minor: 'Dm', position: 7 },
  { camelot: '8A', openKey: '9m', minor: 'Am', position: 8 },
  { camelot: '9A', openKey: '10m', minor: 'Em', position: 9 },
  { camelot: '10A', openKey: '11m', minor: 'Bm', position: 10 },
  { camelot: '11A', openKey: '12m', minor: 'F#m', position: 11 },
];

// Get compatible keys for a given Camelot key
function getCompatibleKeys(key: string): { perfect: string[]; good: string[]; energy: string[] } {
  const num = parseInt(key.slice(0, -1));
  const mode = key.slice(-1);

  // Same key is perfect
  const perfect = [key];

  // Adjacent keys on wheel (+1, -1) are good
  const prevNum = num === 1 ? 12 : num - 1;
  const nextNum = num === 12 ? 1 : num + 1;
  const good = [
    `${prevNum}${mode}`,
    `${nextNum}${mode}`,
  ];

  // Relative major/minor (same number, different letter) for energy change
  const otherMode = mode === 'A' ? 'B' : 'A';
  const energy = [`${num}${otherMode}`];

  return { perfect, good, energy };
}

// Normalize key to Camelot notation
function normalizeKey(key: string): string {
  // Already in Camelot format
  if (/^\d{1,2}[AB]$/.test(key)) {
    return key;
  }

  // Find matching key
  const found = CAMELOT_KEYS.find(k =>
    k.camelot === key ||
    k.major === key ||
    k.minor === key ||
    k.openKey === key
  );

  return found?.camelot || key;
}

export function CamelotWheel({
  selectedKey,
  tracks = [],
  onKeySelect,
  size = 200
}: CamelotWheelProps) {
  const normalizedSelected = selectedKey ? normalizeKey(selectedKey) : undefined;

  const compatibility = useMemo(() => {
    if (!normalizedSelected) return null;
    return getCompatibleKeys(normalizedSelected);
  }, [normalizedSelected]);

  // Count tracks per key
  const trackCounts = useMemo(() => {
    const counts: Record<string, number> = {};
    tracks.forEach(t => {
      const normalized = normalizeKey(t.key);
      counts[normalized] = (counts[normalized] || 0) + t.count;
    });
    return counts;
  }, [tracks]);

  const centerX = size / 2;
  const centerY = size / 2;
  const outerRadius = size / 2 - 10;
  const innerRadius = outerRadius * 0.6;
  const midRadius = (outerRadius + innerRadius) / 2;

  // Create path for a wheel segment
  const createSegmentPath = (position: number, isOuter: boolean) => {
    const startAngle = (position * 30 - 105) * (Math.PI / 180);
    const endAngle = ((position + 1) * 30 - 105) * (Math.PI / 180);

    const r1 = isOuter ? midRadius : innerRadius;
    const r2 = isOuter ? outerRadius : midRadius;

    const x1 = centerX + r1 * Math.cos(startAngle);
    const y1 = centerY + r1 * Math.sin(startAngle);
    const x2 = centerX + r2 * Math.cos(startAngle);
    const y2 = centerY + r2 * Math.sin(startAngle);
    const x3 = centerX + r2 * Math.cos(endAngle);
    const y3 = centerY + r2 * Math.sin(endAngle);
    const x4 = centerX + r1 * Math.cos(endAngle);
    const y4 = centerY + r1 * Math.sin(endAngle);

    return `M ${x1} ${y1} L ${x2} ${y2} A ${r2} ${r2} 0 0 1 ${x3} ${y3} L ${x4} ${y4} A ${r1} ${r1} 0 0 0 ${x1} ${y1}`;
  };

  // Get label position
  const getLabelPosition = (position: number, isOuter: boolean) => {
    const angle = (position * 30 - 90 + 15) * (Math.PI / 180);
    const r = isOuter ? (outerRadius + midRadius) / 2 : (midRadius + innerRadius) / 2;
    return {
      x: centerX + r * Math.cos(angle),
      y: centerY + r * Math.sin(angle),
    };
  };

  // Get segment color based on compatibility
  const getSegmentColor = (camelotKey: string, hasTrack: boolean) => {
    if (camelotKey === normalizedSelected) {
      return 'var(--color-accent)';
    }
    if (compatibility) {
      if (compatibility.perfect.includes(camelotKey)) {
        return 'rgba(34, 197, 94, 0.8)'; // Green - perfect match
      }
      if (compatibility.good.includes(camelotKey)) {
        return 'rgba(59, 130, 246, 0.7)'; // Blue - good match
      }
      if (compatibility.energy.includes(camelotKey)) {
        return 'rgba(251, 191, 36, 0.7)'; // Yellow - energy change
      }
    }
    if (hasTrack) {
      return 'rgba(255, 255, 255, 0.15)';
    }
    return 'rgba(255, 255, 255, 0.05)';
  };

  return (
    <div className="camelot-wheel-container">
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
        {/* Outer ring (Major keys - B) */}
        {CAMELOT_KEYS.filter(k => k.camelot.endsWith('B')).map((key) => {
          const hasTrack = (trackCounts[key.camelot] || 0) > 0;
          const isSelected = key.camelot === normalizedSelected;

          return (
            <g key={key.camelot}>
              <motion.path
                d={createSegmentPath(key.position, true)}
                fill={getSegmentColor(key.camelot, hasTrack)}
                stroke="var(--color-bg)"
                strokeWidth={1}
                initial={false}
                animate={{
                  scale: isSelected ? 1.02 : 1,
                  opacity: compatibility && !isSelected &&
                    !compatibility.perfect.includes(key.camelot) &&
                    !compatibility.good.includes(key.camelot) &&
                    !compatibility.energy.includes(key.camelot) ? 0.3 : 1,
                }}
                whileHover={{ scale: 1.05, opacity: 1 }}
                style={{ cursor: onKeySelect ? 'pointer' : 'default', transformOrigin: 'center' }}
                onClick={() => onKeySelect?.(key.camelot)}
              />
              <text
                x={getLabelPosition(key.position, true).x}
                y={getLabelPosition(key.position, true).y}
                textAnchor="middle"
                dominantBaseline="central"
                fill={isSelected ? 'white' : 'var(--color-text-secondary)'}
                fontSize={size / 20}
                fontWeight={isSelected ? 600 : 400}
                style={{ pointerEvents: 'none' }}
              >
                {key.camelot}
              </text>
              {hasTrack && (
                <circle
                  cx={getLabelPosition(key.position, true).x + 8}
                  cy={getLabelPosition(key.position, true).y - 6}
                  r={3}
                  fill="var(--color-accent)"
                />
              )}
            </g>
          );
        })}

        {/* Inner ring (Minor keys - A) */}
        {CAMELOT_KEYS.filter(k => k.camelot.endsWith('A')).map((key) => {
          const hasTrack = (trackCounts[key.camelot] || 0) > 0;
          const isSelected = key.camelot === normalizedSelected;

          return (
            <g key={key.camelot}>
              <motion.path
                d={createSegmentPath(key.position, false)}
                fill={getSegmentColor(key.camelot, hasTrack)}
                stroke="var(--color-bg)"
                strokeWidth={1}
                initial={false}
                animate={{
                  scale: isSelected ? 1.02 : 1,
                  opacity: compatibility && !isSelected &&
                    !compatibility.perfect.includes(key.camelot) &&
                    !compatibility.good.includes(key.camelot) &&
                    !compatibility.energy.includes(key.camelot) ? 0.3 : 1,
                }}
                whileHover={{ scale: 1.05, opacity: 1 }}
                style={{ cursor: onKeySelect ? 'pointer' : 'default', transformOrigin: 'center' }}
                onClick={() => onKeySelect?.(key.camelot)}
              />
              <text
                x={getLabelPosition(key.position, false).x}
                y={getLabelPosition(key.position, false).y}
                textAnchor="middle"
                dominantBaseline="central"
                fill={isSelected ? 'white' : 'var(--color-text-muted)'}
                fontSize={size / 22}
                fontWeight={isSelected ? 600 : 400}
                style={{ pointerEvents: 'none' }}
              >
                {key.camelot}
              </text>
              {hasTrack && (
                <circle
                  cx={getLabelPosition(key.position, false).x + 6}
                  cy={getLabelPosition(key.position, false).y - 5}
                  r={2.5}
                  fill="var(--color-accent)"
                />
              )}
            </g>
          );
        })}

        {/* Center circle */}
        <circle
          cx={centerX}
          cy={centerY}
          r={innerRadius - 5}
          fill="var(--color-bg-secondary)"
          stroke="var(--color-border)"
          strokeWidth={1}
        />

        {/* Center label */}
        <text
          x={centerX}
          y={centerY - 8}
          textAnchor="middle"
          dominantBaseline="central"
          fill="var(--color-text-muted)"
          fontSize={size / 18}
          fontWeight={500}
        >
          {normalizedSelected || 'Select'}
        </text>
        <text
          x={centerX}
          y={centerY + 10}
          textAnchor="middle"
          dominantBaseline="central"
          fill="var(--color-text-muted)"
          fontSize={size / 25}
        >
          {normalizedSelected ? 'key' : 'a track'}
        </text>
      </svg>

      {/* Legend */}
      {normalizedSelected && (
        <div className="camelot-legend">
          <div className="legend-item">
            <span className="legend-dot perfect" />
            <span>Perfect</span>
          </div>
          <div className="legend-item">
            <span className="legend-dot good" />
            <span>Compatible</span>
          </div>
          <div className="legend-item">
            <span className="legend-dot energy" />
            <span>Energy shift</span>
          </div>
        </div>
      )}
    </div>
  );
}
