package storage

import (
	"database/sql"
	"encoding/json"
	"time"
)

// JobType defines the type of job.
type JobType string

const (
	JobTypeScan    JobType = "scan"
	JobTypeAnalyze JobType = "analyze"
)

// JobStatus defines the status of a job.
type JobStatus string

const (
	JobStatusPending  JobStatus = "pending"
	JobStatusRunning  JobStatus = "running"
	JobStatusComplete JobStatus = "complete"
	JobStatusFailed   JobStatus = "failed"
)

// Job represents a job in the queue.
type Job struct {
	ID          int64
	Type        JobType
	Status      JobStatus
	Priority    int
	Payload     map[string]any
	Result      map[string]any
	Error       string
	Attempts    int
	MaxAttempts int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
}

// CreateJob creates a new job in the queue.
func (d *DB) CreateJob(jobType JobType, priority int, payload map[string]any) (int64, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	result, err := d.db.Exec(`
		INSERT INTO jobs (type, status, priority, payload_json)
		VALUES (?, ?, ?, ?)
	`, string(jobType), string(JobStatusPending), priority, string(payloadJSON))
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// ClaimJob atomically claims the next pending job of the given type.
func (d *DB) ClaimJob(jobType JobType) (*Job, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	row := tx.QueryRow(`
		SELECT id, type, status, priority, payload_json, attempts, max_attempts, created_at
		FROM jobs
		WHERE type = ? AND status = ? AND attempts < max_attempts
		ORDER BY priority DESC, created_at ASC
		LIMIT 1
	`, string(jobType), string(JobStatusPending))

	job := &Job{}
	var payloadJSON sql.NullString
	var createdAt string

	if err := row.Scan(&job.ID, &job.Type, &job.Status, &job.Priority, &payloadJSON, &job.Attempts, &job.MaxAttempts, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No jobs available
		}
		return nil, err
	}

	if payloadJSON.Valid {
		json.Unmarshal([]byte(payloadJSON.String), &job.Payload)
	}
	job.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

	// Mark as running
	now := time.Now()
	_, err = tx.Exec(`
		UPDATE jobs SET status = ?, started_at = ?, attempts = attempts + 1, updated_at = ?
		WHERE id = ?
	`, string(JobStatusRunning), now, now, job.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	job.Status = JobStatusRunning
	job.Attempts++
	job.StartedAt = &now

	return job, nil
}

// CompleteJob marks a job as complete with optional result.
func (d *DB) CompleteJob(jobID int64, result map[string]any) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = d.db.Exec(`
		UPDATE jobs SET status = ?, result_json = ?, completed_at = ?, updated_at = ?
		WHERE id = ?
	`, string(JobStatusComplete), string(resultJSON), now, now, jobID)

	return err
}

// FailJob marks a job as failed with an error message.
func (d *DB) FailJob(jobID int64, errMsg string) error {
	now := time.Now()
	_, err := d.db.Exec(`
		UPDATE jobs SET status = ?, error = ?, updated_at = ?
		WHERE id = ?
	`, string(JobStatusFailed), errMsg, now, jobID)

	return err
}

// RetryJob resets a job to pending for retry.
func (d *DB) RetryJob(jobID int64) error {
	now := time.Now()
	_, err := d.db.Exec(`
		UPDATE jobs SET status = ?, updated_at = ?
		WHERE id = ? AND attempts < max_attempts
	`, string(JobStatusPending), now, jobID)

	return err
}

// GetPendingJobCount returns the count of pending jobs by type.
func (d *DB) GetPendingJobCount(jobType JobType) (int, error) {
	var count int
	row := d.db.QueryRow(`
		SELECT COUNT(*) FROM jobs WHERE type = ? AND status = ?
	`, string(jobType), string(JobStatusPending))

	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// ResetStalledJobs resets jobs that have been running for too long.
func (d *DB) ResetStalledJobs(timeout time.Duration) (int64, error) {
	cutoff := time.Now().Add(-timeout)
	result, err := d.db.Exec(`
		UPDATE jobs SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE status = ? AND started_at < ? AND attempts < max_attempts
	`, string(JobStatusPending), string(JobStatusRunning), cutoff)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
