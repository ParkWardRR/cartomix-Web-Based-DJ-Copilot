import { motion } from 'framer-motion';
import type { Track } from '../types';
import { useStore } from '../store';

interface AnalysisPanelProps {
  track: Track;
  compact?: boolean;
}

interface MetricRowProps {
  label: string;
  value: string | number;
  confidence?: number;
  unit?: string;
  badge?: string;
  badgeColor?: string;
}

function MetricRow({ label, value, confidence, unit, badge, badgeColor }: MetricRowProps) {
  return (
    <div className="metric-row">
      <div className="metric-label">
        {label}
        {badge && (
          <span className="metric-badge" style={{ backgroundColor: badgeColor }}>
            {badge}
          </span>
        )}
      </div>
      <div className="metric-value">
        <span className="value">{value}</span>
        {unit && <span className="unit">{unit}</span>}
        {confidence !== undefined && (
          <div className="confidence-bar">
            <motion.div
              className="confidence-fill"
              initial={{ width: 0 }}
              animate={{ width: `${confidence * 100}%` }}
              transition={{ duration: 0.4 }}
            />
            <span className="confidence-label">{Math.round(confidence * 100)}%</span>
          </div>
        )}
      </div>
    </div>
  );
}

interface QAFlagProps {
  type: string;
  reason: string;
  onDismiss?: () => void;
}

function QAFlag({ type, reason, onDismiss }: QAFlagProps) {
  const flagConfig = {
    needs_review: { icon: '⚠', color: '#f59e0b', label: 'Needs Review' },
    mixed_content: { icon: '🔀', color: '#8b5cf6', label: 'Mixed Content' },
    low_confidence: { icon: '❓', color: '#6b7280', label: 'Low Confidence' },
    speech_detected: { icon: '🎤', color: '#3b82f6', label: 'Speech Detected' },
  }[type] || { icon: '📋', color: '#6b7280', label: type };

  return (
    <motion.div
      className="qa-flag"
      initial={{ opacity: 0, x: -10 }}
      animate={{ opacity: 1, x: 0 }}
      style={{ borderLeftColor: flagConfig.color }}
    >
      <span className="qa-flag-icon">{flagConfig.icon}</span>
      <div className="qa-flag-content">
        <span className="qa-flag-label">{flagConfig.label}</span>
        <span className="qa-flag-reason">{reason}</span>
      </div>
      {onDismiss && (
        <button className="qa-flag-dismiss" onClick={onDismiss}>×</button>
      )}
    </motion.div>
  );
}

interface SoundEventProps {
  label: string;
  category: string;
  confidence: number;
  startTime: number;
  endTime: number;
  duration: number;
}

function SoundEventBar({ label, category, confidence, startTime, endTime, duration }: SoundEventProps) {
  const left = (startTime / duration) * 100;
  const width = ((endTime - startTime) / duration) * 100;

  const categoryColor = {
    music: '#10b981',
    speech: '#3b82f6',
    noise: '#6b7280',
    silence: '#374151',
  }[category.toLowerCase()] || '#8b5cf6';

  return (
    <div
      className="sound-event-bar"
      style={{
        left: `${left}%`,
        width: `${Math.max(width, 1)}%`,
        backgroundColor: categoryColor,
        opacity: 0.3 + confidence * 0.7,
      }}
      title={`${label} (${category}) - ${Math.round(confidence * 100)}% confidence`}
    />
  );
}

export function AnalysisPanel({ track, compact }: AnalysisPanelProps) {
  const { mlSettings } = useStore();

  // Mock data for demonstration - will be replaced with real API data
  const dspResults = {
    bpm: { value: track.bpm, confidence: 0.95 },
    key: { value: track.key, confidence: 0.88 },
    energy: { value: track.energy, confidence: 0.92 },
    loudness: { value: -8.2, unit: 'LUFS', confidence: 0.99 },
    peakdB: { value: -0.3, unit: 'dB' },
    dynamicRange: { value: 8.5, unit: 'dB' },
  };

  // Mock sound analysis results - will come from Layer 1 API
  const soundContext = {
    primary: 'music',
    confidence: 0.96,
  };

  const soundEvents: Array<{
    label: string;
    category: string;
    confidence: number;
    startTime: number;
    endTime: number;
  }> = mlSettings.soundAnalysisEnabled ? [
    { label: 'Electronic Music', category: 'music', confidence: 0.94, startTime: 0, endTime: 180 },
    { label: 'Build-up', category: 'music', confidence: 0.87, startTime: 45, endTime: 60 },
    { label: 'Drop', category: 'music', confidence: 0.91, startTime: 60, endTime: 90 },
  ] : [];

  // Mock QA flags - will come from Layer 1 API
  const qaFlags = track.needsReview ? [
    { type: 'needs_review', reason: 'Track requires manual verification' },
  ] : [];

  const estimatedDuration = 180; // seconds - will come from actual track data

  return (
    <div className={`analysis-panel ${compact ? 'compact' : ''}`}>
      <div className="analysis-section">
        <div className="section-header">
          <h4>DSP Analysis</h4>
          <span className="section-badge dsp">Accelerate vDSP</span>
        </div>
        <div className="metrics-grid">
          <MetricRow
            label="Tempo"
            value={dspResults.bpm.value.toFixed(1)}
            unit="BPM"
            confidence={dspResults.bpm.confidence}
          />
          <MetricRow
            label="Key"
            value={dspResults.key.value}
            confidence={dspResults.key.confidence}
            badge="Camelot"
            badgeColor="#8b5cf6"
          />
          <MetricRow
            label="Energy"
            value={dspResults.energy.value}
            unit="/ 10"
            confidence={dspResults.energy.confidence}
          />
          <MetricRow
            label="Loudness"
            value={dspResults.loudness.value.toFixed(1)}
            unit={dspResults.loudness.unit}
            confidence={dspResults.loudness.confidence}
            badge="EBU R128"
            badgeColor="#3b82f6"
          />
          {!compact && (
            <>
              <MetricRow
                label="Peak"
                value={dspResults.peakdB.value.toFixed(1)}
                unit={dspResults.peakdB.unit}
              />
              <MetricRow
                label="Dynamic Range"
                value={dspResults.dynamicRange.value.toFixed(1)}
                unit={dspResults.dynamicRange.unit}
              />
            </>
          )}
        </div>
      </div>

      {mlSettings.soundAnalysisEnabled && (
        <div className="analysis-section">
          <div className="section-header">
            <h4>Sound Analysis</h4>
            <span className="section-badge ml">Apple ML</span>
          </div>

          <div className="sound-context">
            <div className="context-primary">
              <span className="context-label">Primary Context:</span>
              <span className="context-value">{soundContext.primary}</span>
              <span className="context-confidence">{Math.round(soundContext.confidence * 100)}%</span>
            </div>
          </div>

          {soundEvents.length > 0 && !compact && (
            <div className="sound-events">
              <div className="events-label">Detected Events</div>
              <div className="events-timeline">
                <div className="timeline-track">
                  {soundEvents.map((event, idx) => (
                    <SoundEventBar
                      key={idx}
                      {...event}
                      duration={estimatedDuration}
                    />
                  ))}
                </div>
                <div className="timeline-markers">
                  <span>0:00</span>
                  <span>{Math.floor(estimatedDuration / 60)}:{String(estimatedDuration % 60).padStart(2, '0')}</span>
                </div>
              </div>
              <div className="events-list">
                {soundEvents.map((event, idx) => (
                  <div key={idx} className="event-item">
                    <span className="event-label">{event.label}</span>
                    <span className="event-time">
                      {Math.floor(event.startTime / 60)}:{String(Math.floor(event.startTime) % 60).padStart(2, '0')} -
                      {Math.floor(event.endTime / 60)}:{String(Math.floor(event.endTime) % 60).padStart(2, '0')}
                    </span>
                    <span className="event-confidence">{Math.round(event.confidence * 100)}%</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {qaFlags.length > 0 && (
        <div className="analysis-section qa-section">
          <div className="section-header">
            <h4>QA Flags</h4>
            <span className="qa-count">{qaFlags.length}</span>
          </div>
          <div className="qa-flags-list">
            {qaFlags.map((flag, idx) => (
              <QAFlag key={idx} type={flag.type} reason={flag.reason} />
            ))}
          </div>
        </div>
      )}

      {mlSettings.openl3Enabled && (
        <div className="analysis-section">
          <div className="section-header">
            <h4>Vibe Embedding</h4>
            <span className="section-badge ane">ANE</span>
          </div>
          <div className="embedding-status">
            <span className="embedding-icon">✓</span>
            <span className="embedding-text">512-dim OpenL3 embedding computed</span>
          </div>
        </div>
      )}
    </div>
  );
}
