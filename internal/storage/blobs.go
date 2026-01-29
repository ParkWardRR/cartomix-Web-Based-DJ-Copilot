package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// BlobType defines the type of blob stored.
type BlobType string

const (
	BlobTypeWaveformTile BlobType = "waveform_tile"
	BlobTypeEmbedding    BlobType = "embedding"
)

// Blob represents a content-addressed blob.
type Blob struct {
	Hash      string
	Type      BlobType
	Level     int
	TrackID   int64
	Data      []byte
	Size      int
	CreatedAt time.Time
}

// PutBlob stores a blob with content-addressed hashing.
// Returns the hash of the stored blob.
func (d *DB) PutBlob(blobType BlobType, level int, trackID int64, data []byte) (string, error) {
	hash := hashData(data)

	_, err := d.db.Exec(`
		INSERT OR IGNORE INTO blobs (hash, type, level, track_id, data, size)
		VALUES (?, ?, ?, ?, ?, ?)
	`, hash, string(blobType), level, trackID, data, len(data))
	if err != nil {
		return "", err
	}

	return hash, nil
}

// GetBlob retrieves a blob by hash.
func (d *DB) GetBlob(hash string) (*Blob, error) {
	b := &Blob{}
	var blobType string
	var createdAt string

	row := d.db.QueryRow(`
		SELECT hash, type, level, track_id, data, size, created_at
		FROM blobs WHERE hash = ?
	`, hash)

	if err := row.Scan(&b.Hash, &blobType, &b.Level, &b.TrackID, &b.Data, &b.Size, &createdAt); err != nil {
		return nil, err
	}

	b.Type = BlobType(blobType)
	b.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)

	return b, nil
}

// GetBlobsForTrack retrieves all blobs for a track, optionally filtered by type.
func (d *DB) GetBlobsForTrack(trackID int64, blobType BlobType) ([]*Blob, error) {
	query := "SELECT hash, type, level, track_id, data, size, created_at FROM blobs WHERE track_id = ?"
	args := []any{trackID}

	if blobType != "" {
		query += " AND type = ?"
		args = append(args, string(blobType))
	}

	query += " ORDER BY level ASC"

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blobs []*Blob
	for rows.Next() {
		b := &Blob{}
		var bt string
		var createdAt string

		if err := rows.Scan(&b.Hash, &bt, &b.Level, &b.TrackID, &b.Data, &b.Size, &createdAt); err != nil {
			return nil, err
		}

		b.Type = BlobType(bt)
		b.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		blobs = append(blobs, b)
	}

	return blobs, rows.Err()
}

// DeleteBlobsForTrack deletes all blobs for a track.
func (d *DB) DeleteBlobsForTrack(trackID int64) error {
	_, err := d.db.Exec("DELETE FROM blobs WHERE track_id = ?", trackID)
	return err
}

func hashData(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
