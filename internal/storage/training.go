package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// TrainingLabel represents a user-labeled audio segment
type TrainingLabel struct {
	ID               int64     `json:"id"`
	TrackID          int64     `json:"track_id"`
	ContentHash      string    `json:"content_hash"` // Derived from track
	TrackPath        string    `json:"track_path"`   // Derived from track
	LabelValue       string    `json:"label_value"`
	StartBeat        int       `json:"start_beat"`
	EndBeat          int       `json:"end_beat"`
	StartTimeSeconds float64   `json:"start_time_seconds"`
	EndTimeSeconds   float64   `json:"end_time_seconds"`
	Source           string    `json:"source"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TrainingJob represents a model training job
type TrainingJob struct {
	ID           int64              `json:"id"`
	JobID        string             `json:"job_id"`
	Status       string             `json:"status"`
	Progress     float64            `json:"progress"`
	CurrentEpoch *int               `json:"current_epoch,omitempty"`
	TotalEpochs  *int               `json:"total_epochs,omitempty"`
	CurrentLoss  *float64           `json:"current_loss,omitempty"`
	Accuracy     *float64           `json:"accuracy,omitempty"`
	F1Score      *float64           `json:"f1_score,omitempty"`
	ModelPath    *string            `json:"model_path,omitempty"`
	ModelVersion *int               `json:"model_version,omitempty"`
	ErrorMessage *string            `json:"error_message,omitempty"`
	LabelCounts  map[string]int     `json:"label_counts,omitempty"`
	StartedAt    *time.Time         `json:"started_at,omitempty"`
	CompletedAt  *time.Time         `json:"completed_at,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
}

// ModelVersion represents a trained model version
type ModelVersion struct {
	ID            int64          `json:"id"`
	ModelType     string         `json:"model_type"`
	Version       int            `json:"version"`
	ModelPath     string         `json:"model_path"`
	Accuracy      float64        `json:"accuracy"`
	F1Score       float64        `json:"f1_score"`
	IsActive      bool           `json:"is_active"`
	LabelCounts   map[string]int `json:"label_counts,omitempty"`
	TrainingJobID *string        `json:"training_job_id,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

// TrainingLabelStats represents label distribution statistics
type TrainingLabelStats struct {
	TotalLabels   int            `json:"total_labels"`
	LabelCounts   map[string]int `json:"label_counts"`
	TracksCovered int            `json:"tracks_covered"`
	AvgPerTrack   float64        `json:"avg_per_track"`
}

// AddTrainingLabel adds a new training label
func (db *DB) AddTrainingLabel(ctx context.Context, label *TrainingLabel) error {
	query := `
		INSERT INTO training_labels (track_id, label_value, start_beat, end_beat, start_time_seconds, end_time_seconds, source)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (track_id, start_beat, end_beat) DO UPDATE SET
			label_value = excluded.label_value,
			start_time_seconds = excluded.start_time_seconds,
			end_time_seconds = excluded.end_time_seconds,
			source = excluded.source,
			updated_at = datetime('now')
	`
	result, err := db.db.ExecContext(ctx, query,
		label.TrackID, label.LabelValue, label.StartBeat, label.EndBeat,
		label.StartTimeSeconds, label.EndTimeSeconds, label.Source,
	)
	if err != nil {
		return fmt.Errorf("failed to add training label: %w", err)
	}

	id, _ := result.LastInsertId()
	label.ID = id
	return nil
}

// GetTrainingLabels retrieves all training labels with optional filtering
func (db *DB) GetTrainingLabels(ctx context.Context, trackID *int64, labelValue *string) ([]TrainingLabel, error) {
	query := `
		SELECT tl.id, tl.track_id, t.content_hash, t.path, tl.label_value, tl.start_beat, tl.end_beat,
		       tl.start_time_seconds, tl.end_time_seconds, tl.source, tl.created_at, tl.updated_at
		FROM training_labels tl
		JOIN tracks t ON tl.track_id = t.id
		WHERE 1=1
	`
	var args []interface{}

	if trackID != nil {
		query += " AND tl.track_id = ?"
		args = append(args, *trackID)
	}
	if labelValue != nil {
		query += " AND tl.label_value = ?"
		args = append(args, *labelValue)
	}

	query += " ORDER BY tl.track_id, tl.start_beat"

	rows, err := db.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query training labels: %w", err)
	}
	defer rows.Close()

	var labels []TrainingLabel
	for rows.Next() {
		var l TrainingLabel
		var createdAt, updatedAt string
		err := rows.Scan(
			&l.ID, &l.TrackID, &l.ContentHash, &l.TrackPath, &l.LabelValue,
			&l.StartBeat, &l.EndBeat, &l.StartTimeSeconds, &l.EndTimeSeconds,
			&l.Source, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan training label: %w", err)
		}
		l.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		l.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		labels = append(labels, l)
	}

	return labels, nil
}

// DeleteTrainingLabel removes a training label
func (db *DB) DeleteTrainingLabel(ctx context.Context, id int64) error {
	_, err := db.db.ExecContext(ctx, "DELETE FROM training_labels WHERE id = ?", id)
	return err
}

// GetTrainingLabelStats returns statistics about training labels
func (db *DB) GetTrainingLabelStats(ctx context.Context) (*TrainingLabelStats, error) {
	var stats TrainingLabelStats
	stats.LabelCounts = make(map[string]int)

	// Get label counts
	rows, err := db.db.QueryContext(ctx, `
		SELECT label_value, COUNT(*) FROM training_labels GROUP BY label_value
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var label string
		var count int
		if err := rows.Scan(&label, &count); err != nil {
			return nil, err
		}
		stats.LabelCounts[label] = count
		stats.TotalLabels += count
	}

	// Get tracks covered
	err = db.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT track_id) FROM training_labels
	`).Scan(&stats.TracksCovered)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if stats.TracksCovered > 0 {
		stats.AvgPerTrack = float64(stats.TotalLabels) / float64(stats.TracksCovered)
	}

	return &stats, nil
}

// CreateTrainingJob creates a new training job
func (db *DB) CreateTrainingJob(ctx context.Context, jobID string, labelCounts map[string]int) error {
	countsJSON, _ := json.Marshal(labelCounts)
	_, err := db.db.ExecContext(ctx, `
		INSERT INTO training_jobs (job_id, status, label_counts, started_at)
		VALUES (?, 'pending', ?, datetime('now'))
	`, jobID, string(countsJSON))
	return err
}

// UpdateTrainingJobProgress updates the progress of a training job
func (db *DB) UpdateTrainingJobProgress(ctx context.Context, jobID string, status string, progress float64, epoch, totalEpochs *int, loss *float64) error {
	query := `
		UPDATE training_jobs
		SET status = ?, progress = ?, current_epoch = ?, total_epochs = ?, current_loss = ?
		WHERE job_id = ?
	`
	_, err := db.db.ExecContext(ctx, query, status, progress, epoch, totalEpochs, loss, jobID)
	return err
}

// CompleteTrainingJob marks a training job as completed
func (db *DB) CompleteTrainingJob(ctx context.Context, jobID string, accuracy, f1Score float64, modelPath string, modelVersion int) error {
	query := `
		UPDATE training_jobs
		SET status = 'completed', progress = 1.0, accuracy = ?, f1_score = ?,
		    model_path = ?, model_version = ?, completed_at = datetime('now')
		WHERE job_id = ?
	`
	_, err := db.db.ExecContext(ctx, query, accuracy, f1Score, modelPath, modelVersion, jobID)
	return err
}

// FailTrainingJob marks a training job as failed
func (db *DB) FailTrainingJob(ctx context.Context, jobID string, errorMsg string) error {
	query := `
		UPDATE training_jobs
		SET status = 'failed', error_message = ?, completed_at = datetime('now')
		WHERE job_id = ?
	`
	_, err := db.db.ExecContext(ctx, query, errorMsg, jobID)
	return err
}

// GetTrainingJob retrieves a training job by ID
func (db *DB) GetTrainingJob(ctx context.Context, jobID string) (*TrainingJob, error) {
	var job TrainingJob
	var labelCountsJSON sql.NullString
	var createdAt string
	var startedAt, completedAt sql.NullString

	err := db.db.QueryRowContext(ctx, `
		SELECT id, job_id, status, progress, current_epoch, total_epochs, current_loss,
		       accuracy, f1_score, model_path, model_version, error_message, label_counts,
		       started_at, completed_at, created_at
		FROM training_jobs WHERE job_id = ?
	`, jobID).Scan(
		&job.ID, &job.JobID, &job.Status, &job.Progress,
		&job.CurrentEpoch, &job.TotalEpochs, &job.CurrentLoss,
		&job.Accuracy, &job.F1Score, &job.ModelPath, &job.ModelVersion,
		&job.ErrorMessage, &labelCountsJSON, &startedAt, &completedAt, &createdAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	job.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	if startedAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", startedAt.String)
		job.StartedAt = &t
	}
	if completedAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", completedAt.String)
		job.CompletedAt = &t
	}
	if labelCountsJSON.Valid {
		json.Unmarshal([]byte(labelCountsJSON.String), &job.LabelCounts)
	}

	return &job, nil
}

// ListTrainingJobs retrieves recent training jobs
func (db *DB) ListTrainingJobs(ctx context.Context, limit int) ([]TrainingJob, error) {
	rows, err := db.db.QueryContext(ctx, `
		SELECT id, job_id, status, progress, current_epoch, total_epochs, current_loss,
		       accuracy, f1_score, model_path, model_version, error_message, label_counts,
		       started_at, completed_at, created_at
		FROM training_jobs ORDER BY created_at DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []TrainingJob
	for rows.Next() {
		var job TrainingJob
		var labelCountsJSON sql.NullString
		var createdAt string
		var startedAt, completedAt sql.NullString

		err := rows.Scan(
			&job.ID, &job.JobID, &job.Status, &job.Progress,
			&job.CurrentEpoch, &job.TotalEpochs, &job.CurrentLoss,
			&job.Accuracy, &job.F1Score, &job.ModelPath, &job.ModelVersion,
			&job.ErrorMessage, &labelCountsJSON, &startedAt, &completedAt, &createdAt,
		)
		if err != nil {
			return nil, err
		}

		job.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		if startedAt.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", startedAt.String)
			job.StartedAt = &t
		}
		if completedAt.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", completedAt.String)
			job.CompletedAt = &t
		}
		if labelCountsJSON.Valid {
			json.Unmarshal([]byte(labelCountsJSON.String), &job.LabelCounts)
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// AddModelVersion adds a new model version
func (db *DB) AddModelVersion(ctx context.Context, mv *ModelVersion) error {
	labelCountsJSON, _ := json.Marshal(mv.LabelCounts)
	result, err := db.db.ExecContext(ctx, `
		INSERT INTO model_versions (model_type, version, model_path, accuracy, f1_score, is_active, label_counts, training_job_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, mv.ModelType, mv.Version, mv.ModelPath, mv.Accuracy, mv.F1Score, mv.IsActive, string(labelCountsJSON), mv.TrainingJobID)
	if err != nil {
		return err
	}
	mv.ID, _ = result.LastInsertId()
	return nil
}

// GetModelVersions retrieves all model versions of a type
func (db *DB) GetModelVersions(ctx context.Context, modelType string) ([]ModelVersion, error) {
	rows, err := db.db.QueryContext(ctx, `
		SELECT id, model_type, version, model_path, accuracy, f1_score, is_active, label_counts, training_job_id, created_at
		FROM model_versions WHERE model_type = ? ORDER BY version DESC
	`, modelType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []ModelVersion
	for rows.Next() {
		var mv ModelVersion
		var labelCountsJSON sql.NullString
		var createdAt string

		err := rows.Scan(
			&mv.ID, &mv.ModelType, &mv.Version, &mv.ModelPath,
			&mv.Accuracy, &mv.F1Score, &mv.IsActive, &labelCountsJSON,
			&mv.TrainingJobID, &createdAt,
		)
		if err != nil {
			return nil, err
		}

		mv.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		if labelCountsJSON.Valid {
			json.Unmarshal([]byte(labelCountsJSON.String), &mv.LabelCounts)
		}

		versions = append(versions, mv)
	}

	return versions, nil
}

// ActivateModelVersion activates a specific model version
func (db *DB) ActivateModelVersion(ctx context.Context, modelType string, version int) error {
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Deactivate all versions of this type
	_, err = tx.ExecContext(ctx, "UPDATE model_versions SET is_active = 0 WHERE model_type = ?", modelType)
	if err != nil {
		return err
	}

	// Activate the specified version
	result, err := tx.ExecContext(ctx, "UPDATE model_versions SET is_active = 1 WHERE model_type = ? AND version = ?", modelType, version)
	if err != nil {
		return err
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("model version %d not found", version)
	}

	return tx.Commit()
}

// GetActiveModelVersion gets the currently active model version
func (db *DB) GetActiveModelVersion(ctx context.Context, modelType string) (*ModelVersion, error) {
	var mv ModelVersion
	var labelCountsJSON sql.NullString
	var createdAt string

	err := db.db.QueryRowContext(ctx, `
		SELECT id, model_type, version, model_path, accuracy, f1_score, is_active, label_counts, training_job_id, created_at
		FROM model_versions WHERE model_type = ? AND is_active = 1
	`, modelType).Scan(
		&mv.ID, &mv.ModelType, &mv.Version, &mv.ModelPath,
		&mv.Accuracy, &mv.F1Score, &mv.IsActive, &labelCountsJSON,
		&mv.TrainingJobID, &createdAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	mv.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	if labelCountsJSON.Valid {
		json.Unmarshal([]byte(labelCountsJSON.String), &mv.LabelCounts)
	}

	return &mv, nil
}

// DeleteModelVersion removes a model version
func (db *DB) DeleteModelVersion(ctx context.Context, modelType string, version int) error {
	_, err := db.db.ExecContext(ctx, "DELETE FROM model_versions WHERE model_type = ? AND version = ?", modelType, version)
	return err
}
