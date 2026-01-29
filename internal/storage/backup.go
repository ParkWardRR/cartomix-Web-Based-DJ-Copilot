package storage

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// BackupMetadata describes a database backup.
type BackupMetadata struct {
	Version         string    `json:"version"`
	CreatedAt       time.Time `json:"created_at"`
	TrackCount      int       `json:"track_count"`
	AnalysisCount   int       `json:"analysis_count"`
	SchemaVersion   int       `json:"schema_version"`
	DatabaseSize    int64     `json:"database_size_bytes"`
	Checksum        string    `json:"checksum_sha256"`
	AnalyzerVersion string    `json:"analyzer_version,omitempty"`
}

// CacheVersionInfo describes the analysis cache versioning.
type CacheVersionInfo struct {
	AnalyzerVersion   int32  `json:"analyzer_version"`
	ModelVersion      int32  `json:"model_version,omitempty"`
	ParamsHash        string `json:"params_hash,omitempty"`
	ContentHash       string `json:"content_hash,omitempty"`
	CachedAt          time.Time `json:"cached_at"`
	IsValid           bool   `json:"is_valid"`
}

// DatabaseInfo returns information about the database state.
func (d *DB) DatabaseInfo() (*BackupMetadata, error) {
	meta := &BackupMetadata{
		Version:   "1.0",
		CreatedAt: time.Now(),
	}

	// Get track count
	var trackCount int
	if err := d.db.QueryRow("SELECT COUNT(*) FROM tracks").Scan(&trackCount); err != nil {
		return nil, fmt.Errorf("count tracks: %w", err)
	}
	meta.TrackCount = trackCount

	// Get analysis count
	var analysisCount int
	if err := d.db.QueryRow("SELECT COUNT(*) FROM analyses").Scan(&analysisCount); err != nil {
		return nil, fmt.Errorf("count analyses: %w", err)
	}
	meta.AnalysisCount = analysisCount

	// Get schema version
	var schemaVersion int
	row := d.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations")
	if err := row.Scan(&schemaVersion); err != nil {
		return nil, fmt.Errorf("get schema version: %w", err)
	}
	meta.SchemaVersion = schemaVersion

	return meta, nil
}

// CreateBackup creates a backup archive of the database.
// Returns the path to the backup file and metadata.
func (d *DB) CreateBackup(backupDir string) (string, *BackupMetadata, error) {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", nil, fmt.Errorf("create backup dir: %w", err)
	}

	// Get database path
	var dbPath string
	row := d.db.QueryRow("PRAGMA database_list")
	var seq int
	var name string
	if err := row.Scan(&seq, &name, &dbPath); err != nil {
		return "", nil, fmt.Errorf("get db path: %w", err)
	}

	// Checkpoint WAL to ensure all data is in the main file
	if _, err := d.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		d.logger.Warn("WAL checkpoint failed", "error", err)
	}

	// Get metadata
	meta, err := d.DatabaseInfo()
	if err != nil {
		return "", nil, err
	}

	// Get file size
	info, err := os.Stat(dbPath)
	if err != nil {
		return "", nil, fmt.Errorf("stat db: %w", err)
	}
	meta.DatabaseSize = info.Size()

	// Create backup filename
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("algiers-backup-%s.tar.gz", timestamp)
	backupPath := filepath.Join(backupDir, backupName)

	// Create tar.gz archive
	backupFile, err := os.Create(backupPath)
	if err != nil {
		return "", nil, fmt.Errorf("create backup file: %w", err)
	}
	defer backupFile.Close()

	gzWriter := gzip.NewWriter(backupFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Calculate checksum and add database file
	checksum, err := addFileToTar(tarWriter, dbPath, "algiers.db")
	if err != nil {
		return "", nil, fmt.Errorf("add db to archive: %w", err)
	}
	meta.Checksum = checksum

	// Add metadata file
	metaJSON, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return "", nil, fmt.Errorf("marshal metadata: %w", err)
	}

	metaHeader := &tar.Header{
		Name:    "backup-metadata.json",
		Size:    int64(len(metaJSON)),
		Mode:    0644,
		ModTime: time.Now(),
	}
	if err := tarWriter.WriteHeader(metaHeader); err != nil {
		return "", nil, fmt.Errorf("write meta header: %w", err)
	}
	if _, err := tarWriter.Write(metaJSON); err != nil {
		return "", nil, fmt.Errorf("write meta content: %w", err)
	}

	d.logger.Info("backup created",
		"path", backupPath,
		"tracks", meta.TrackCount,
		"analyses", meta.AnalysisCount,
		"size_mb", float64(meta.DatabaseSize)/(1024*1024),
	)

	return backupPath, meta, nil
}

// RestoreBackup restores a database from a backup archive.
// The current database must be closed before calling this.
func RestoreBackup(backupPath, dataDir string) (*BackupMetadata, error) {
	// Open the backup archive
	backupFile, err := os.Open(backupPath)
	if err != nil {
		return nil, fmt.Errorf("open backup: %w", err)
	}
	defer backupFile.Close()

	gzReader, err := gzip.NewReader(backupFile)
	if err != nil {
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	var meta *BackupMetadata
	var dbData []byte

	// Read archive contents
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tar: %w", err)
		}

		switch header.Name {
		case "algiers.db":
			dbData, err = io.ReadAll(tarReader)
			if err != nil {
				return nil, fmt.Errorf("read db data: %w", err)
			}
		case "backup-metadata.json":
			metaData, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, fmt.Errorf("read metadata: %w", err)
			}
			meta = &BackupMetadata{}
			if err := json.Unmarshal(metaData, meta); err != nil {
				return nil, fmt.Errorf("unmarshal metadata: %w", err)
			}
		}
	}

	if dbData == nil {
		return nil, fmt.Errorf("backup does not contain database file")
	}

	// Verify checksum
	if meta != nil && meta.Checksum != "" {
		hash := sha256.Sum256(dbData)
		actualChecksum := hex.EncodeToString(hash[:])
		if actualChecksum != meta.Checksum {
			return nil, fmt.Errorf("checksum mismatch: expected %s, got %s", meta.Checksum, actualChecksum)
		}
	}

	// Backup existing database
	existingDB := filepath.Join(dataDir, "algiers.db")
	if _, err := os.Stat(existingDB); err == nil {
		backupName := fmt.Sprintf("algiers.db.backup-%s", time.Now().Format("20060102-150405"))
		if err := os.Rename(existingDB, filepath.Join(dataDir, backupName)); err != nil {
			return nil, fmt.Errorf("backup existing db: %w", err)
		}
	}

	// Write restored database
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	if err := os.WriteFile(existingDB, dbData, 0644); err != nil {
		return nil, fmt.Errorf("write restored db: %w", err)
	}

	return meta, nil
}

// ExportAnalysisCache exports all analyses to JSON for portability.
func (d *DB) ExportAnalysisCache(outputPath string) error {
	rows, err := d.db.Query(`
		SELECT a.id, a.track_id, t.content_hash, t.path, a.version, a.status,
		       a.duration_seconds, a.bpm, a.bpm_confidence, a.key_value, a.key_format,
		       a.key_confidence, a.energy_global, a.integrated_lufs, a.true_peak_db,
		       a.beatgrid_json, a.sections_json, a.cue_points_json,
		       a.created_at, a.updated_at
		FROM analyses a
		JOIN tracks t ON t.id = a.track_id
		WHERE a.status = 'complete'
		ORDER BY a.updated_at DESC
	`)
	if err != nil {
		return fmt.Errorf("query analyses: %w", err)
	}
	defer rows.Close()

	type ExportedAnalysis struct {
		ID              int64   `json:"id"`
		TrackID         int64   `json:"track_id"`
		ContentHash     string  `json:"content_hash"`
		Path            string  `json:"path"`
		Version         int32   `json:"version"`
		Status          string  `json:"status"`
		DurationSeconds float64 `json:"duration_seconds"`
		BPM             float64 `json:"bpm"`
		BPMConfidence   float64 `json:"bpm_confidence"`
		KeyValue        string  `json:"key_value"`
		KeyFormat       string  `json:"key_format"`
		KeyConfidence   float64 `json:"key_confidence"`
		EnergyGlobal    int32   `json:"energy_global"`
		IntegratedLufs  float64 `json:"integrated_lufs"`
		TruePeakDb      float64 `json:"true_peak_db"`
		BeatgridJSON    string  `json:"beatgrid,omitempty"`
		SectionsJSON    string  `json:"sections,omitempty"`
		CuePointsJSON   string  `json:"cue_points,omitempty"`
		CreatedAt       string  `json:"created_at"`
		UpdatedAt       string  `json:"updated_at"`
	}

	var analyses []ExportedAnalysis
	for rows.Next() {
		var a ExportedAnalysis
		var createdAt, updatedAt sql.NullTime
		if err := rows.Scan(
			&a.ID, &a.TrackID, &a.ContentHash, &a.Path, &a.Version, &a.Status,
			&a.DurationSeconds, &a.BPM, &a.BPMConfidence, &a.KeyValue, &a.KeyFormat,
			&a.KeyConfidence, &a.EnergyGlobal, &a.IntegratedLufs, &a.TruePeakDb,
			&a.BeatgridJSON, &a.SectionsJSON, &a.CuePointsJSON,
			&createdAt, &updatedAt,
		); err != nil {
			return fmt.Errorf("scan row: %w", err)
		}
		if createdAt.Valid {
			a.CreatedAt = createdAt.Time.Format(time.RFC3339)
		}
		if updatedAt.Valid {
			a.UpdatedAt = updatedAt.Time.Format(time.RFC3339)
		}
		analyses = append(analyses, a)
	}

	export := struct {
		Version   string             `json:"version"`
		ExportedAt string            `json:"exported_at"`
		Count     int                `json:"count"`
		Analyses  []ExportedAnalysis `json:"analyses"`
	}{
		Version:    "1.0",
		ExportedAt: time.Now().Format(time.RFC3339),
		Count:      len(analyses),
		Analyses:   analyses,
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal export: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("write export: %w", err)
	}

	d.logger.Info("exported analysis cache", "path", outputPath, "count", len(analyses))
	return nil
}

// GetCacheVersion returns the cache version info for a track.
func (d *DB) GetCacheVersion(trackID int64) (*CacheVersionInfo, error) {
	row := d.db.QueryRow(`
		SELECT a.version, a.updated_at, a.status
		FROM analyses a
		WHERE a.track_id = ?
		ORDER BY a.version DESC
		LIMIT 1
	`, trackID)

	var version int32
	var updatedAt sql.NullTime
	var status string

	if err := row.Scan(&version, &updatedAt, &status); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	info := &CacheVersionInfo{
		AnalyzerVersion: version,
		IsValid:         status == string(AnalysisStatusComplete),
	}
	if updatedAt.Valid {
		info.CachedAt = updatedAt.Time
	}

	return info, nil
}

// InvalidateCacheOlderThan marks analyses older than the given version as requiring re-analysis.
func (d *DB) InvalidateCacheOlderThan(minVersion int32) (int64, error) {
	result, err := d.db.Exec(`
		UPDATE analyses
		SET status = 'pending'
		WHERE version < ? AND status = 'complete'
	`, minVersion)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// VacuumDatabase optimizes the database and reclaims space.
func (d *DB) VacuumDatabase() error {
	// Checkpoint WAL first
	if _, err := d.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		d.logger.Warn("WAL checkpoint failed", "error", err)
	}

	// Run VACUUM
	if _, err := d.db.Exec("VACUUM"); err != nil {
		return fmt.Errorf("vacuum: %w", err)
	}

	// Analyze for query optimizer
	if _, err := d.db.Exec("ANALYZE"); err != nil {
		return fmt.Errorf("analyze: %w", err)
	}

	d.logger.Info("database vacuum complete")
	return nil
}

// IntegrityCheck performs a database integrity check.
func (d *DB) IntegrityCheck() error {
	row := d.db.QueryRow("PRAGMA integrity_check")
	var result string
	if err := row.Scan(&result); err != nil {
		return fmt.Errorf("integrity check: %w", err)
	}
	if result != "ok" {
		return fmt.Errorf("integrity check failed: %s", result)
	}
	d.logger.Info("database integrity check passed")
	return nil
}

// addFileToTar adds a file to the tar archive and returns its SHA256 checksum.
func addFileToTar(tw *tar.Writer, srcPath, destName string) (string, error) {
	file, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", err
	}

	header := &tar.Header{
		Name:    destName,
		Size:    info.Size(),
		Mode:    int64(info.Mode()),
		ModTime: info.ModTime(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return "", err
	}

	// Calculate checksum while writing
	hasher := sha256.New()
	teeReader := io.TeeReader(file, hasher)

	if _, err := io.Copy(tw, teeReader); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
