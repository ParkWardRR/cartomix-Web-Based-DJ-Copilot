-- Initial schema for Algiers DJ set prep copilot

-- Tracks table - core metadata and file info
CREATE TABLE IF NOT EXISTS tracks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content_hash TEXT NOT NULL UNIQUE,
    path TEXT NOT NULL,
    title TEXT,
    artist TEXT,
    album TEXT,
    file_size INTEGER,
    file_modified_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tracks_content_hash ON tracks(content_hash);
CREATE INDEX IF NOT EXISTS idx_tracks_path ON tracks(path);

-- Analyses table - full analysis results
CREATE TABLE IF NOT EXISTS analyses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    track_id INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    version INTEGER NOT NULL DEFAULT 1,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, analyzing, complete, failed
    error TEXT,

    -- Core analysis data
    duration_seconds REAL,
    bpm REAL,
    bpm_confidence REAL,
    is_dynamic_tempo BOOLEAN DEFAULT FALSE,

    -- Key detection
    key_value TEXT,
    key_format TEXT, -- camelot, openkey
    key_confidence REAL,

    -- Energy
    energy_global INTEGER,

    -- Loudness
    integrated_lufs REAL,
    true_peak_db REAL,

    -- Serialized data (JSON/protobuf)
    beatgrid_json TEXT,
    sections_json TEXT,
    cue_points_json TEXT,
    energy_segments_json TEXT,
    transition_windows_json TEXT,
    tempo_map_json TEXT,

    -- Embedding for similarity (binary blob)
    embedding BLOB,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(track_id, version)
);

CREATE INDEX IF NOT EXISTS idx_analyses_track_id ON analyses(track_id);
CREATE INDEX IF NOT EXISTS idx_analyses_status ON analyses(status);

-- Cue edits table - user-authored cue modifications
CREATE TABLE IF NOT EXISTS cue_edits (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    track_id INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    cue_index INTEGER NOT NULL,
    beat_index INTEGER NOT NULL,
    cue_type TEXT NOT NULL,
    label TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(track_id, cue_index)
);

CREATE INDEX IF NOT EXISTS idx_cue_edits_track_id ON cue_edits(track_id);

-- Graph edges table - transition compatibility scores
CREATE TABLE IF NOT EXISTS graph_edges (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    from_track_id INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    to_track_id INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    score REAL NOT NULL,
    tempo_delta REAL,
    energy_delta INTEGER,
    key_relation TEXT,
    window_overlap TEXT,
    reason TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(from_track_id, to_track_id)
);

CREATE INDEX IF NOT EXISTS idx_graph_edges_from ON graph_edges(from_track_id);
CREATE INDEX IF NOT EXISTS idx_graph_edges_to ON graph_edges(to_track_id);
CREATE INDEX IF NOT EXISTS idx_graph_edges_score ON graph_edges(score DESC);

-- Blobs table - content-addressed storage for waveform tiles and embeddings
CREATE TABLE IF NOT EXISTS blobs (
    hash TEXT PRIMARY KEY,
    type TEXT NOT NULL, -- waveform_tile, embedding, etc.
    level INTEGER, -- zoom level for waveform tiles
    track_id INTEGER REFERENCES tracks(id) ON DELETE CASCADE,
    data BLOB NOT NULL,
    size INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_blobs_track_id ON blobs(track_id);
CREATE INDEX IF NOT EXISTS idx_blobs_type ON blobs(type);

-- Job queue table - for resumable scan/analysis jobs
CREATE TABLE IF NOT EXISTS jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL, -- scan, analyze
    status TEXT NOT NULL DEFAULT 'pending', -- pending, running, complete, failed
    priority INTEGER DEFAULT 0,
    payload_json TEXT,
    result_json TEXT,
    error TEXT,
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    started_at DATETIME,
    completed_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_type ON jobs(type);

-- Schema version tracking
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO schema_migrations (version) VALUES (1);
