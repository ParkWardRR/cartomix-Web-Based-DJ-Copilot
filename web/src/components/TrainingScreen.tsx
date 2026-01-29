import { useEffect, useState, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import * as api from '../api';
import type {
  TrainingLabelResponse,
  TrainingLabelStats,
  TrainingJobResponse,
  ModelVersionResponse,
  DJSectionLabel,
} from '../api';
import { useStore } from '../store';

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

interface LabelEditorProps {
  trackId: string;
  onLabelAdded: () => void;
}

function LabelEditor({ trackId, onLabelAdded }: LabelEditorProps) {
  const { trackMap } = useStore();
  const track = trackMap[trackId];

  const [selectedLabel, setSelectedLabel] = useState<DJSectionLabel>('drop');
  const [startBeat, setStartBeat] = useState(0);
  const [endBeat, setEndBeat] = useState(32);
  const [isAdding, setIsAdding] = useState(false);

  if (!track) {
    return (
      <div className="label-editor-empty">
        <p>Select a track to add training labels</p>
      </div>
    );
  }

  const handleAddLabel = async () => {
    if (startBeat >= endBeat) {
      alert('End beat must be greater than start beat');
      return;
    }

    setIsAdding(true);
    try {
      // Calculate time from beats (assuming 120 BPM as default)
      const bpm = track.bpm || 120;
      const startTime = (startBeat / bpm) * 60;
      const endTime = (endBeat / bpm) * 60;

      // Get track ID from API (using content hash)
      const tracks = await api.listTracks({ query: track.title, limit: 1 });
      const apiTrack = tracks.find(t => t.content_hash === trackId);
      if (!apiTrack?.id) {
        throw new Error('Track not found in database');
      }

      await api.addTrainingLabel({
        track_id: apiTrack.id,
        label_value: selectedLabel,
        start_beat: startBeat,
        end_beat: endBeat,
        start_time_seconds: startTime,
        end_time_seconds: endTime,
        source: 'user',
      });

      onLabelAdded();
      setStartBeat(endBeat);
      setEndBeat(endBeat + 32);
    } catch (error) {
      console.error('Failed to add label:', error);
      alert('Failed to add label: ' + (error as Error).message);
    } finally {
      setIsAdding(false);
    }
  };

  return (
    <div className="label-editor">
      <div className="label-editor-header">
        <h4>Add Label</h4>
        <span className="track-name">{track.title}</span>
      </div>

      <div className="label-selector">
        {(Object.entries(LABEL_CONFIG) as [DJSectionLabel, typeof LABEL_CONFIG[DJSectionLabel]][]).map(
          ([value, config]) => (
            <button
              key={value}
              className={`label-btn ${selectedLabel === value ? 'active' : ''}`}
              style={{
                borderColor: config.color,
                backgroundColor: selectedLabel === value ? config.color : 'transparent',
                color: selectedLabel === value ? 'white' : config.color,
              }}
              onClick={() => setSelectedLabel(value)}
            >
              {config.displayName}
            </button>
          )
        )}
      </div>

      <div className="beat-inputs">
        <div className="input-group">
          <label>Start Beat</label>
          <input
            type="number"
            value={startBeat}
            onChange={(e) => setStartBeat(parseInt(e.target.value) || 0)}
            min={0}
          />
        </div>
        <div className="input-group">
          <label>End Beat</label>
          <input
            type="number"
            value={endBeat}
            onChange={(e) => setEndBeat(parseInt(e.target.value) || 0)}
            min={1}
          />
        </div>
      </div>

      <button className="add-label-btn" onClick={handleAddLabel} disabled={isAdding}>
        {isAdding ? 'Adding...' : 'Add Label'}
      </button>
    </div>
  );
}

interface DatasetTableProps {
  labels: TrainingLabelResponse[];
  onDelete: (id: number) => void;
}

function DatasetTable({ labels, onDelete }: DatasetTableProps) {
  if (labels.length === 0) {
    return (
      <div className="dataset-empty">
        <span className="empty-icon">📋</span>
        <p>No training labels yet</p>
        <span className="empty-hint">Add labels to train your custom model</span>
      </div>
    );
  }

  return (
    <div className="dataset-table">
      <table>
        <thead>
          <tr>
            <th>Track</th>
            <th>Label</th>
            <th>Beats</th>
            <th>Source</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {labels.map((label) => {
            const config = LABEL_CONFIG[label.label_value as DJSectionLabel];
            return (
              <motion.tr
                key={label.id}
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
              >
                <td className="track-cell">
                  <span className="track-path" title={label.track_path}>
                    {label.track_path.split('/').pop()}
                  </span>
                </td>
                <td>
                  <span
                    className="label-badge"
                    style={{ backgroundColor: config?.color || '#6b7280' }}
                  >
                    {config?.displayName || label.label_value}
                  </span>
                </td>
                <td className="beats-cell">
                  {label.start_beat} - {label.end_beat}
                </td>
                <td className="source-cell">{label.source}</td>
                <td>
                  <button
                    className="delete-btn"
                    onClick={() => onDelete(label.id)}
                    title="Delete label"
                  >
                    ×
                  </button>
                </td>
              </motion.tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

interface TrainingProgressCardProps {
  job: TrainingJobResponse | null;
  onStartTraining: () => void;
  isReady: boolean;
}

function TrainingProgressCard({ job, onStartTraining, isReady }: TrainingProgressCardProps) {
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

      {isTraining && job && (
        <div className="progress-section">
          <div className="progress-bar">
            <motion.div
              className="progress-fill"
              animate={{ width: `${job.progress * 100}%` }}
              transition={{ duration: 0.3 }}
            />
          </div>
          <span className="progress-label">{Math.round(job.progress * 100)}%</span>

          {job.current_epoch !== undefined && job.total_epochs !== undefined && (
            <span className="epoch-label">
              Epoch {job.current_epoch} / {job.total_epochs}
            </span>
          )}

          {job.current_loss !== undefined && (
            <span className="loss-label">Loss: {job.current_loss.toFixed(4)}</span>
          )}
        </div>
      )}

      {job?.status === 'completed' && (
        <div className="results-section">
          <div className="result-row">
            <span className="result-label">Accuracy</span>
            <span className="result-value">{((job.accuracy || 0) * 100).toFixed(1)}%</span>
          </div>
          <div className="result-row">
            <span className="result-label">F1 Score</span>
            <span className="result-value">{((job.f1_score || 0) * 100).toFixed(1)}%</span>
          </div>
          {job.model_version && (
            <div className="result-row">
              <span className="result-label">Model Version</span>
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
          {isReady ? 'Start Training' : 'Need More Labels'}
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
                <button className="delete-btn" onClick={() => onDelete(version.version)}>
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

export function TrainingScreen() {
  const { apiAvailable, selectedId } = useStore();

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
        <p>Training requires API connection</p>
        <span className="empty-hint">Start the Go engine to train custom models</span>
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
    <div className="training-screen">
      <div className="training-main">
        <div className="panel">
          <div className="panel-header">
            <h3>Training Dataset</h3>
            <span className="count-badge">{stats?.total_labels || 0} labels</span>
          </div>

          {stats && (
            <div className="stats-grid">
              {Object.entries(stats.label_counts).map(([label, count]) => {
                const config = LABEL_CONFIG[label as DJSectionLabel];
                const minRequired = stats.min_samples_required;
                const isEnough = count >= minRequired;
                return (
                  <div key={label} className="stat-item">
                    <div
                      className="stat-color"
                      style={{ backgroundColor: config?.color || '#6b7280' }}
                    />
                    <span className="stat-label">{config?.displayName || label}</span>
                    <span className={`stat-count ${isEnough ? 'enough' : 'needs-more'}`}>
                      {count} / {minRequired}
                    </span>
                  </div>
                );
              })}
            </div>
          )}

          <AnimatePresence>
            <DatasetTable labels={labels} onDelete={handleDeleteLabel} />
          </AnimatePresence>
        </div>
      </div>

      <div className="training-sidebar">
        {selectedId && (
          <div className="panel">
            <LabelEditor trackId={selectedId} onLabelAdded={loadData} />
          </div>
        )}

        <div className="panel">
          <TrainingProgressCard
            job={currentJob}
            onStartTraining={handleStartTraining}
            isReady={stats?.ready_for_training || false}
          />
        </div>

        <div className="panel">
          <div className="panel-header">
            <h4>Model Versions</h4>
          </div>
          <ModelVersionsList
            versions={modelVersions}
            onActivate={handleActivateModel}
            onDelete={handleDeleteModel}
          />
        </div>
      </div>
    </div>
  );
}
