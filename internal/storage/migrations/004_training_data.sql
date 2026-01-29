-- Training labels for custom DJ section model
CREATE TABLE IF NOT EXISTS training_labels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    track_id INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    label_value TEXT NOT NULL CHECK (label_value IN ('intro', 'build', 'drop', 'break', 'outro', 'verse', 'chorus')),
    start_beat INTEGER NOT NULL,
    end_beat INTEGER NOT NULL,
    start_time_seconds REAL NOT NULL,
    end_time_seconds REAL NOT NULL,
    source TEXT NOT NULL DEFAULT 'user' CHECK (source IN ('user', 'auto_detected', 'imported')),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(track_id, start_beat, end_beat)
);

CREATE INDEX IF NOT EXISTS idx_training_labels_track ON training_labels(track_id);
CREATE INDEX IF NOT EXISTS idx_training_labels_label ON training_labels(label_value);

-- Training jobs history
CREATE TABLE IF NOT EXISTS training_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'preparing', 'training', 'evaluating', 'completed', 'failed', 'cancelled')),
    progress REAL NOT NULL DEFAULT 0.0,
    current_epoch INTEGER,
    total_epochs INTEGER,
    current_loss REAL,
    accuracy REAL,
    f1_score REAL,
    model_path TEXT,
    model_version INTEGER,
    error_message TEXT,
    label_counts TEXT,  -- JSON: {"intro": 10, "drop": 15, ...}
    started_at TEXT,
    completed_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_training_jobs_status ON training_jobs(status);
CREATE INDEX IF NOT EXISTS idx_training_jobs_job_id ON training_jobs(job_id);

-- Model versions
CREATE TABLE IF NOT EXISTS model_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    model_type TEXT NOT NULL DEFAULT 'dj_section',
    version INTEGER NOT NULL,
    model_path TEXT NOT NULL,
    accuracy REAL NOT NULL DEFAULT 0.0,
    f1_score REAL NOT NULL DEFAULT 0.0,
    is_active INTEGER NOT NULL DEFAULT 0,
    label_counts TEXT,  -- JSON
    training_job_id TEXT REFERENCES training_jobs(job_id),
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(model_type, version)
);

CREATE INDEX IF NOT EXISTS idx_model_versions_type ON model_versions(model_type);
CREATE INDEX IF NOT EXISTS idx_model_versions_active ON model_versions(is_active);

-- Trigger to ensure only one active model per type
CREATE TRIGGER IF NOT EXISTS ensure_single_active_model
BEFORE UPDATE ON model_versions
WHEN NEW.is_active = 1
BEGIN
    UPDATE model_versions SET is_active = 0 WHERE model_type = NEW.model_type AND id != NEW.id;
END;
