import { useEffect, useState, useCallback } from 'react';
import { motion } from 'framer-motion';
import * as api from '../api';
import type {
  TrainingLabelResponse,
  TrainingLabelStats,
  TrainingJobResponse,
  ModelVersionResponse,
  DJSectionLabel,
} from '../api';
import { useStore } from '../store';
import { WaveformLabeler } from './WaveformLabeler';

// Label configuration
const LABEL_CONFIG: Record<DJSectionLabel, { displayName: string; color: string }> = {
  intro: { displayName: 'Intro', color: '#22c55e' },
  build: { displayName: 'Build-up', color: '#eab308' },
  drop: { displayName: 'Drop', color: '#ef4444' },
  break: { displayName: 'Breakdown', color: '#a855f7' },
  outro: { displayName: 'Outro', color: '#3b82f6' },
  verse: { displayName: 'Verse', color: '#4b5563' },
  chorus: { displayName: 'Chorus', color: '#ec4899' },
};

interface TrainingProgressCardProps {
  job: TrainingJobResponse | null;
  onStartTraining: () => void;
  isReady: boolean;
  stats: TrainingLabelStats | null;
}

function TrainingProgressCard({ job, onStartTraining, isReady, stats }: TrainingProgressCardProps) {
  const isTraining = job && ['pending', 'preparing', 'training', 'evaluating'].includes(job.status);

  return (
    <div className="training-progress-card">
      <div className="card-header">
        <h4>Model Training</h4>
        {job && (
          <span className={`status-badge ${job.status}`}>
            {job.status}
          </span>
        )}
      </div>

      {/* Stats grid showing label counts */}
      {stats && (
        <div className="mini-stats-grid">
          {Object.entries(LABEL_CONFIG).map(([label, config]) => {
            const count = stats.label_counts[label] || 0;
            const minRequired = stats.min_samples_required;
            const isEnough = count >= minRequired;
            return (
              <div key={label} className={`mini-stat ${isEnough ? 'enough' : ''}`}>
                <div className="mini-stat-color" style={{ backgroundColor: config.color }} />
                <span className="mini-stat-count">{count}</span>
              </div>
            );
          })}
        </div>
      )}

      {isTraining && job && (
        <div className="progress-section">
          <div className="progress-bar">
            <motion.div
              className="progress-fill"
              animate={{ width: `${job.progress * 100}%` }}
              transition={{ duration: 0.3 }}
            />
          </div>
          <div className="progress-info">
            <span className="progress-label">{Math.round(job.progress * 100)}%</span>
            {job.current_epoch !== undefined && job.total_epochs !== undefined && (
              <span className="epoch-label">
                Epoch {job.current_epoch}/{job.total_epochs}
              </span>
            )}
            {job.current_loss !== undefined && (
              <span className="loss-label">Loss: {job.current_loss.toFixed(4)}</span>
            )}
          </div>
        </div>
      )}

      {job?.status === 'completed' && (
        <div className="results-section">
          <div className="result-row">
            <span className="result-label">Accuracy</span>
            <span className="result-value success">{((job.accuracy || 0) * 100).toFixed(1)}%</span>
          </div>
          <div className="result-row">
            <span className="result-label">F1 Score</span>
            <span className="result-value success">{((job.f1_score || 0) * 100).toFixed(1)}%</span>
          </div>
          {job.model_version && (
            <div className="result-row">
              <span className="result-label">Model</span>
              <span className="result-value">v{job.model_version}</span>
            </div>
          )}
        </div>
      )}

      {job?.status === 'failed' && (
        <div className="error-section">
          <span className="error-icon">⚠</span>
          <span className="error-message">{job.error_message || 'Training failed'}</span>
        </div>
      )}

      {!isTraining && (
        <button
          className="start-training-btn"
          onClick={onStartTraining}
          disabled={!isReady}
        >
          {isReady ? 'Start Training' : `Need ${stats?.min_samples_required || 10} samples per label`}
        </button>
      )}
    </div>
  );
}

interface ModelVersionsListProps {
  versions: ModelVersionResponse[];
  onActivate: (version: number) => void;
  onDelete: (version: number) => void;
}

function ModelVersionsList({ versions, onActivate, onDelete }: ModelVersionsListProps) {
  if (versions.length === 0) {
    return (
      <div className="models-empty">
        <span className="empty-icon">🤖</span>
        <p>No trained models yet</p>
        <span className="empty-hint">Add labels and train your first model</span>
      </div>
    );
  }

  return (
    <div className="model-versions-list">
      {versions.map((version) => (
        <motion.div
          key={version.version}
          className={`model-version-card ${version.is_active ? 'active' : ''}`}
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
        >
          <div className="version-header">
            <span className="version-number">v{version.version}</span>
            {version.is_active && <span className="active-badge">Active</span>}
          </div>

          <div className="version-metrics">
            <div className="metric">
              <span className="metric-label">Accuracy</span>
              <span className="metric-value">{(version.accuracy * 100).toFixed(1)}%</span>
            </div>
            <div className="metric">
              <span className="metric-label">F1</span>
              <span className="metric-value">{(version.f1_score * 100).toFixed(1)}%</span>
            </div>
          </div>

          <div className="version-actions">
            {!version.is_active && (
              <>
                <button className="activate-btn" onClick={() => onActivate(version.version)}>
                  Activate
                </button>
                <button className="delete-version-btn" onClick={() => onDelete(version.version)}>
                  Delete
                </button>
              </>
            )}
          </div>
        </motion.div>
      ))}
    </div>
  );
}

interface RecentLabelsProps {
  labels: TrainingLabelResponse[];
  onDelete: (id: number) => void;
}

function RecentLabels({ labels, onDelete }: RecentLabelsProps) {
  // Show only the most recent 10 labels
  const recentLabels = labels.slice(0, 10);

  if (recentLabels.length === 0) {
    return (
      <div className="recent-labels-empty">
        <p>No labels yet</p>
        <span>Select a track above to start labeling</span>
      </div>
    );
  }

  return (
    <div className="recent-labels">
      {recentLabels.map((label) => {
        const config = LABEL_CONFIG[label.label_value as DJSectionLabel];
        return (
          <motion.div
            key={label.id}
            className="recent-label-item"
            initial={{ opacity: 0, x: -10 }}
            animate={{ opacity: 1, x: 0 }}
          >
            <span
              className="label-badge"
              style={{ backgroundColor: config?.color || '#6b7280' }}
            >
              {config?.displayName || label.label_value}
            </span>
            <span className="label-track" title={label.track_path}>
              {label.track_path.split('/').pop()?.slice(0, 25)}...
            </span>
            <span className="label-beats">{label.start_beat}-{label.end_beat}</span>
            <button className="label-delete" onClick={() => onDelete(label.id)}>×</button>
          </motion.div>
        );
      })}
    </div>
  );
}

export function TrainingScreen() {
  const { apiAvailable, selectedId, trackMap, tracks } = useStore();
  const selectedTrack = selectedId ? trackMap[selectedId] : undefined;

  const [labels, setLabels] = useState<TrainingLabelResponse[]>([]);
  const [stats, setStats] = useState<TrainingLabelStats | null>(null);
  const [currentJob, setCurrentJob] = useState<TrainingJobResponse | null>(null);
  const [modelVersions, setModelVersions] = useState<ModelVersionResponse[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const loadData = useCallback(async () => {
    if (!apiAvailable) return;

    try {
      const [labelsRes, statsRes, jobsRes, versionsRes] = await Promise.all([
        api.getTrainingLabels(),
        api.getTrainingLabelStats(),
        api.getTrainingJobs(1),
        api.getModelVersions(),
      ]);

      setLabels(labelsRes);
      setStats(statsRes);
      setCurrentJob(jobsRes[0] || null);
      setModelVersions(versionsRes);
    } catch (error) {
      console.error('Failed to load training data:', error);
    } finally {
      setIsLoading(false);
    }
  }, [apiAvailable]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // Poll for job progress when training
  useEffect(() => {
    if (!currentJob || !['pending', 'preparing', 'training', 'evaluating'].includes(currentJob.status)) {
      return;
    }

    const interval = setInterval(async () => {
      try {
        const job = await api.getTrainingJob(currentJob.job_id);
        setCurrentJob(job);

        if (job.status === 'completed' || job.status === 'failed') {
          loadData(); // Reload all data when done
        }
      } catch (error) {
        console.error('Failed to poll job:', error);
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [currentJob, loadData]);

  const handleDeleteLabel = async (id: number) => {
    try {
      await api.deleteTrainingLabel(id);
      loadData();
    } catch (error) {
      console.error('Failed to delete label:', error);
    }
  };

  const handleStartTraining = async () => {
    try {
      const result = await api.startTraining();
      const job = await api.getTrainingJob(result.job_id);
      setCurrentJob(job);
    } catch (error) {
      alert('Failed to start training: ' + (error as Error).message);
    }
  };

  const handleActivateModel = async (version: number) => {
    try {
      await api.activateModelVersion(version);
      loadData();
    } catch (error) {
      alert('Failed to activate model: ' + (error as Error).message);
    }
  };

  const handleDeleteModel = async (version: number) => {
    if (!confirm(`Delete model version ${version}?`)) return;

    try {
      await api.deleteModelVersion(version);
      loadData();
    } catch (error) {
      alert('Failed to delete model: ' + (error as Error).message);
    }
  };

  if (!apiAvailable) {
    return (
      <div className="training-screen-empty">
        <span className="empty-icon">🤖</span>
        <h3>Custom Model Training</h3>
        <p>Train a personalized AI model to understand your music style</p>
        <span className="empty-hint">Connect to the API to get started</span>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="training-screen-loading">
        <div className="loading-spinner" />
        <span>Loading training data...</span>
      </div>
    );
  }

  return (
    <div className="training-screen-v2">
      {/* Main area - Waveform Labeler */}
      <div className="training-main-area">
        <div className="labeler-panel">
          <div className="panel-header">
            <h3>Section Labeling</h3>
            <span className="panel-subtitle">
              Teach the AI to recognize sections in your style
            </span>
          </div>

          {selectedTrack ? (
            <WaveformLabeler
              track={selectedTrack}
              existingLabels={labels}
              onLabelAdded={loadData}
              onLabelDeleted={handleDeleteLabel}
            />
          ) : (
            <div className="no-track-selected">
              <span className="icon">🎵</span>
              <p>Select a track from your library to start labeling</p>
              <span className="hint">
                Use keyboard shortcuts <kbd>↓</kbd> <kbd>↑</kbd> to navigate tracks
              </span>
            </div>
          )}

          {/* Track selector for quick access */}
          <div className="quick-track-selector">
            <span className="selector-label">Quick select:</span>
            <div className="track-chips">
              {tracks.slice(0, 8).map((track) => (
                <button
                  key={track.id}
                  className={`track-chip ${selectedId === track.id ? 'active' : ''}`}
                  onClick={() => useStore.getState().setSelectedId(track.id)}
                >
                  {track.title.slice(0, 20)}
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Sidebar - Training Progress & Models */}
      <div className="training-sidebar-v2">
        <TrainingProgressCard
          job={currentJob}
          onStartTraining={handleStartTraining}
          isReady={stats?.ready_for_training || false}
          stats={stats}
        />

        <div className="panel">
          <div className="panel-header">
            <h4>Trained Models</h4>
          </div>
          <ModelVersionsList
            versions={modelVersions}
            onActivate={handleActivateModel}
            onDelete={handleDeleteModel}
          />
        </div>

        <div className="panel">
          <div className="panel-header">
            <h4>Recent Labels</h4>
            <span className="count-badge">{stats?.total_labels || 0}</span>
          </div>
          <RecentLabels labels={labels} onDelete={handleDeleteLabel} />
        </div>
      </div>
    </div>
  );
}
