package storage

import (
	"database/sql"
	"errors"
	"time"

	"github.com/cartomix/cancun/gen/go/common"
)

// Track represents a track in the database.
type Track struct {
	ID             int64
	ContentHash    string
	Path           string
	Title          string
	Artist         string
	Album          string
	FileSize       int64
	FileModifiedAt time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// UpsertTrack inserts or updates a track by content hash.
func (d *DB) UpsertTrack(t *Track) (int64, error) {
	result, err := d.db.Exec(`
		INSERT INTO tracks (content_hash, path, title, artist, album, file_size, file_modified_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(content_hash) DO UPDATE SET
			path = excluded.path,
			title = excluded.title,
			artist = excluded.artist,
			album = excluded.album,
			file_size = excluded.file_size,
			file_modified_at = excluded.file_modified_at,
			updated_at = CURRENT_TIMESTAMP
	`, t.ContentHash, t.Path, t.Title, t.Artist, t.Album, t.FileSize, t.FileModifiedAt)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		// On conflict, fetch the existing ID
		row := d.db.QueryRow("SELECT id FROM tracks WHERE content_hash = ?", t.ContentHash)
		if err := row.Scan(&id); err != nil {
			return 0, err
		}
	}

	return id, nil
}

// GetTrackByHash retrieves a track by content hash.
func (d *DB) GetTrackByHash(hash string) (*Track, error) {
	t := &Track{}
	row := d.db.QueryRow(`
		SELECT id, content_hash, path, title, artist, album, file_size, file_modified_at, created_at, updated_at
		FROM tracks WHERE content_hash = ?
	`, hash)

	var fileModifiedAt, createdAt, updatedAt sql.NullTime
	var title, artist, album sql.NullString
	var fileSize sql.NullInt64

	err := row.Scan(&t.ID, &t.ContentHash, &t.Path, &title, &artist, &album, &fileSize, &fileModifiedAt, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	if title.Valid {
		t.Title = title.String
	}
	if artist.Valid {
		t.Artist = artist.String
	}
	if album.Valid {
		t.Album = album.String
	}
	if fileSize.Valid {
		t.FileSize = fileSize.Int64
	}
	if fileModifiedAt.Valid {
		t.FileModifiedAt = fileModifiedAt.Time
	}
	if createdAt.Valid {
		t.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		t.UpdatedAt = updatedAt.Time
	}

	return t, nil
}

// GetTrackByID retrieves a track by ID.
func (d *DB) GetTrackByID(id int64) (*Track, error) {
	t := &Track{}
	row := d.db.QueryRow(`
		SELECT id, content_hash, path, title, artist, album, file_size, file_modified_at, created_at, updated_at
		FROM tracks WHERE id = ?
	`, id)

	var fileModifiedAt, createdAt, updatedAt sql.NullTime
	var title, artist, album sql.NullString
	var fileSize sql.NullInt64

	err := row.Scan(&t.ID, &t.ContentHash, &t.Path, &title, &artist, &album, &fileSize, &fileModifiedAt, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	if title.Valid {
		t.Title = title.String
	}
	if artist.Valid {
		t.Artist = artist.String
	}
	if album.Valid {
		t.Album = album.String
	}
	if fileSize.Valid {
		t.FileSize = fileSize.Int64
	}
	if fileModifiedAt.Valid {
		t.FileModifiedAt = fileModifiedAt.Time
	}
	if createdAt.Valid {
		t.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		t.UpdatedAt = updatedAt.Time
	}

	return t, nil
}

// GetTrackByPath retrieves a track by its file path.
func (d *DB) GetTrackByPath(path string) (*Track, error) {
	t := &Track{}
	row := d.db.QueryRow(`
		SELECT id, content_hash, path, title, artist, album, file_size, file_modified_at, created_at, updated_at
		FROM tracks WHERE path = ?
	`, path)

	var fileModifiedAt, createdAt, updatedAt sql.NullTime
	var title, artist, album sql.NullString
	var fileSize sql.NullInt64

	err := row.Scan(&t.ID, &t.ContentHash, &t.Path, &title, &artist, &album, &fileSize, &fileModifiedAt, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	if title.Valid {
		t.Title = title.String
	}
	if artist.Valid {
		t.Artist = artist.String
	}
	if album.Valid {
		t.Album = album.String
	}
	if fileSize.Valid {
		t.FileSize = fileSize.Int64
	}
	if fileModifiedAt.Valid {
		t.FileModifiedAt = fileModifiedAt.Time
	}
	if createdAt.Valid {
		t.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		t.UpdatedAt = updatedAt.Time
	}

	return t, nil
}

// ResolveTrack attempts to find a track using the provided proto TrackId.
// It prefers content_hash when available, falling back to path.
func (d *DB) ResolveTrack(id *common.TrackId) (*Track, error) {
	if id == nil {
		return nil, errors.New("track id is required")
	}

	if id.ContentHash != "" {
		if t, err := d.GetTrackByHash(id.ContentHash); err == nil {
			return t, nil
		} else if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}

	if id.Path != "" {
		return d.GetTrackByPath(id.Path)
	}

	return nil, sql.ErrNoRows
}

// ListTracks returns tracks matching the query.
func (d *DB) ListTracks(query string, limit int) ([]*Track, error) {
	sqlQuery := `
		SELECT id, content_hash, path, title, artist, album, file_size, file_modified_at, created_at, updated_at
		FROM tracks
	`
	args := []any{}

	if query != "" {
		sqlQuery += " WHERE title LIKE ? OR artist LIKE ? OR path LIKE ?"
		pattern := "%" + query + "%"
		args = append(args, pattern, pattern, pattern)
	}

	sqlQuery += " ORDER BY updated_at DESC"

	if limit > 0 {
		sqlQuery += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := d.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []*Track
	for rows.Next() {
		t := &Track{}
		var fileModifiedAt, createdAt, updatedAt sql.NullTime
		var title, artist, album sql.NullString
		var fileSize sql.NullInt64

		if err := rows.Scan(&t.ID, &t.ContentHash, &t.Path, &title, &artist, &album, &fileSize, &fileModifiedAt, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		if title.Valid {
			t.Title = title.String
		}
		if artist.Valid {
			t.Artist = artist.String
		}
		if album.Valid {
			t.Album = album.String
		}
		if fileSize.Valid {
			t.FileSize = fileSize.Int64
		}
		if fileModifiedAt.Valid {
			t.FileModifiedAt = fileModifiedAt.Time
		}
		if createdAt.Valid {
			t.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			t.UpdatedAt = updatedAt.Time
		}

		tracks = append(tracks, t)
	}

	return tracks, rows.Err()
}
