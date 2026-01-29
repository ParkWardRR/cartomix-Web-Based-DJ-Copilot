import { useEffect } from 'react';
import { motion } from 'framer-motion';
import { useStore, type SimilarTrack } from '../store';

interface SimilarTracksProps {
  trackId: string;
  onSelectTrack?: (id: string) => void;
  limit?: number;
  compact?: boolean;
}

function ScoreBar({ value, label, color }: { value: number; label: string; color: string }) {
  return (
    <div className="score-bar">
      <span className="score-label">{label}</span>
      <div className="score-track">
        <motion.div
          className="score-fill"
          style={{ backgroundColor: color }}
          initial={{ width: 0 }}
          animate={{ width: `${value}%` }}
          transition={{ duration: 0.5, ease: 'easeOut' }}
        />
      </div>
      <span className="score-value">{Math.round(value)}%</span>
    </div>
  );
}

function SimilarTrackCard({
  track,
  showExplanations,
  onSelect,
}: {
  track: SimilarTrack;
  showExplanations: boolean;
  onSelect?: (id: string) => void;
}) {
  const keyRelationColor = {
    same: '#10b981',
    relative: '#8b5cf6',
    compatible: '#3b82f6',
    harmonic: '#f59e0b',
    clash: '#ef4444',
    unknown: '#6b7280',
  }[track.keyRelation] || '#6b7280';

  return (
    <motion.div
      className="similar-track-card"
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      whileHover={{ scale: 1.01 }}
      onClick={() => onSelect?.(track.contentHash)}
    >
      <div className="similar-track-header">
        <div className="similar-track-info">
          <span className="similar-track-title">{track.title}</span>
          <span className="similar-track-artist">{track.artist}</span>
        </div>
        <div className="similar-track-score">
          <span className="score-number">{Math.round(track.score * 100)}</span>
          <span className="score-label">match</span>
        </div>
      </div>

      {showExplanations && (
        <>
          <div className="similar-track-bars">
            <ScoreBar value={track.vibeMatch} label="Vibe" color="#8b5cf6" />
            <ScoreBar value={track.tempoMatch} label="Tempo" color="#3b82f6" />
            <ScoreBar value={track.keyMatch} label="Key" color="#10b981" />
            <ScoreBar value={track.energyMatch} label="Energy" color="#f59e0b" />
          </div>

          <div className="similar-track-tags">
            <span
              className="tag"
              style={{ backgroundColor: keyRelationColor, color: 'white' }}
            >
              {track.keyRelation}
            </span>
            {track.bpmDelta !== 0 && (
              <span className="tag tag-muted">
                {track.bpmDelta > 0 ? '+' : ''}{track.bpmDelta.toFixed(1)} BPM
              </span>
            )}
            {track.energyDelta !== 0 && (
              <span className="tag tag-muted">
                energy {track.energyDelta > 0 ? '+' : ''}{track.energyDelta}
              </span>
            )}
          </div>

          <div className="similar-track-explanation">
            {track.explanation}
          </div>
        </>
      )}
    </motion.div>
  );
}

export function SimilarTracks({ trackId, onSelectTrack, limit = 10, compact }: SimilarTracksProps) {
  const {
    similarTracks,
    similarTracksLoading,
    similarTracksError,
    mlSettings,
    fetchSimilarTracks,
    apiAvailable,
  } = useStore();

  useEffect(() => {
    if (trackId && apiAvailable) {
      fetchSimilarTracks(trackId, limit);
    }
  }, [trackId, limit, apiAvailable, fetchSimilarTracks]);

  if (!apiAvailable) {
    return (
      <div className="similar-tracks-empty">
        <span className="empty-icon">⚡</span>
        <p>Similar tracks require API connection</p>
      </div>
    );
  }

  if (similarTracksLoading) {
    return (
      <div className="similar-tracks-loading">
        <div className="loading-spinner small" />
        <span>Finding similar tracks...</span>
      </div>
    );
  }

  if (similarTracksError) {
    return (
      <div className="similar-tracks-error">
        <span className="error-icon">⚠</span>
        <p>{similarTracksError}</p>
      </div>
    );
  }

  if (similarTracks.length === 0) {
    return (
      <div className="similar-tracks-empty">
        <span className="empty-icon">🎵</span>
        <p>No similar tracks found</p>
        <span className="empty-hint">Analyze more tracks to find matches</span>
      </div>
    );
  }

  return (
    <div className={`similar-tracks ${compact ? 'compact' : ''}`}>
      <div className="similar-tracks-header">
        <h4>Similar Tracks</h4>
        <span className="count-badge">{similarTracks.length} matches</span>
      </div>
      <div className="similar-tracks-list">
        {similarTracks.map((track) => (
          <SimilarTrackCard
            key={track.contentHash}
            track={track}
            showExplanations={mlSettings.showExplanations && !compact}
            onSelect={onSelectTrack}
          />
        ))}
      </div>
    </div>
  );
}
