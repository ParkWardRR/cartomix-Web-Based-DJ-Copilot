-- Migration 002: OpenL3 embeddings for ML-powered similarity
-- Adds 512-dimensional OpenL3 embedding storage and similarity caching

-- Add OpenL3 embedding column to analyses table
ALTER TABLE analyses ADD COLUMN openl3_embedding BLOB;
ALTER TABLE analyses ADD COLUMN openl3_window_count INTEGER DEFAULT 0;

-- Embedding similarity cache for fast lookup
CREATE TABLE IF NOT EXISTS embedding_similarity (
    track_a_id INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    track_b_id INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,

    -- OpenL3 cosine similarity (0-1)
    openl3_similarity REAL,

    -- Combined weighted similarity (0-1)
    -- Weights: 0.5*openl3 + 0.2*tempo + 0.2*key + 0.1*energy
    combined_score REAL,

    -- Component scores for explanation
    tempo_similarity REAL,
    key_similarity REAL,
    energy_similarity REAL,

    -- Explanation string
    explanation TEXT,

    computed_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (track_a_id, track_b_id)
);

CREATE INDEX IF NOT EXISTS idx_similarity_a ON embedding_similarity(track_a_id);
CREATE INDEX IF NOT EXISTS idx_similarity_b ON embedding_similarity(track_b_id);
CREATE INDEX IF NOT EXISTS idx_similarity_score ON embedding_similarity(combined_score DESC);

-- Windowed OpenL3 embeddings for fine-grained section similarity
CREATE TABLE IF NOT EXISTS openl3_windows (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    track_id INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    analysis_version INTEGER NOT NULL,
    window_index INTEGER NOT NULL,
    timestamp_seconds REAL NOT NULL,
    duration_seconds REAL NOT NULL,
    embedding BLOB NOT NULL,  -- 512 x float32 = 2KB per window
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(track_id, analysis_version, window_index)
);

CREATE INDEX IF NOT EXISTS idx_openl3_windows_track ON openl3_windows(track_id);

-- ML model settings (per-user feature flags)
CREATE TABLE IF NOT EXISTS ml_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    setting_key TEXT NOT NULL UNIQUE,
    setting_value TEXT NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Default settings
INSERT OR IGNORE INTO ml_settings (setting_key, setting_value) VALUES
    ('openl3_enabled', 'true'),
    ('sound_analysis_enabled', 'false'),
    ('custom_model_enabled', 'false'),
    ('min_similarity_threshold', '0.5'),
    ('show_explanations', 'true');

INSERT OR IGNORE INTO schema_migrations (version) VALUES (2);
