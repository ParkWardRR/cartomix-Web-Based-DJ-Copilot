package storage

import (
	"database/sql"

	"github.com/cartomix/cancun/internal/similarity"
)

// GetTrackFeaturesForSimilarity fetches track features needed for similarity search.
func (d *DB) GetTrackFeaturesForSimilarity(trackID int64) (*similarity.TrackFeatures, error) {
	row := d.db.QueryRow(`
		SELECT t.id, t.content_hash, t.title, t.artist,
		       COALESCE(a.bpm, 0), COALESCE(a.key_value, ''), COALESCE(a.energy_global, 5),
		       COALESCE(a.openl3_embedding, X'')
		FROM tracks t
		LEFT JOIN analyses a ON a.id = (
			SELECT id FROM analyses a2
			WHERE a2.track_id = t.id AND a2.status = 'complete'
			ORDER BY a2.version DESC LIMIT 1
		)
		WHERE t.id = ?
	`, trackID)

	var features similarity.TrackFeatures
	var title, artist sql.NullString

	if err := row.Scan(
		&features.TrackID, &features.ContentHash, &title, &artist,
		&features.BPM, &features.KeyValue, &features.Energy,
		&features.OpenL3Embedding,
	); err != nil {
		return nil, err
	}

	if title.Valid {
		features.Title = title.String
	}
	if artist.Valid {
		features.Artist = artist.String
	}

	return &features, nil
}

// GetAllTrackFeaturesForSimilarity fetches features for all analyzed tracks.
func (d *DB) GetAllTrackFeaturesForSimilarity() ([]*similarity.TrackFeatures, error) {
	rows, err := d.db.Query(`
		SELECT t.id, t.content_hash, t.title, t.artist,
		       COALESCE(a.bpm, 0), COALESCE(a.key_value, ''), COALESCE(a.energy_global, 5),
		       COALESCE(a.openl3_embedding, X'')
		FROM tracks t
		INNER JOIN analyses a ON a.id = (
			SELECT id FROM analyses a2
			WHERE a2.track_id = t.id AND a2.status = 'complete'
			ORDER BY a2.version DESC LIMIT 1
		)
		WHERE a.openl3_embedding IS NOT NULL AND LENGTH(a.openl3_embedding) > 0
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*similarity.TrackFeatures
	for rows.Next() {
		var features similarity.TrackFeatures
		var title, artist sql.NullString

		if err := rows.Scan(
			&features.TrackID, &features.ContentHash, &title, &artist,
			&features.BPM, &features.KeyValue, &features.Energy,
			&features.OpenL3Embedding,
		); err != nil {
			return nil, err
		}

		if title.Valid {
			features.Title = title.String
		}
		if artist.Valid {
			features.Artist = artist.String
		}

		results = append(results, &features)
	}

	return results, rows.Err()
}

// GetTrackFeaturesExcluding fetches features for all tracks except the specified ones.
func (d *DB) GetTrackFeaturesExcluding(excludeIDs []int64) ([]*similarity.TrackFeatures, error) {
	if len(excludeIDs) == 0 {
		return d.GetAllTrackFeaturesForSimilarity()
	}

	// Build placeholder string for IN clause
	placeholders := make([]byte, 0, len(excludeIDs)*2)
	args := make([]any, len(excludeIDs))
	for i, id := range excludeIDs {
		if i > 0 {
			placeholders = append(placeholders, ',')
		}
		placeholders = append(placeholders, '?')
		args[i] = id
	}

	query := `
		SELECT t.id, t.content_hash, t.title, t.artist,
		       COALESCE(a.bpm, 0), COALESCE(a.key_value, ''), COALESCE(a.energy_global, 5),
		       COALESCE(a.openl3_embedding, X'')
		FROM tracks t
		INNER JOIN analyses a ON a.id = (
			SELECT id FROM analyses a2
			WHERE a2.track_id = t.id AND a2.status = 'complete'
			ORDER BY a2.version DESC LIMIT 1
		)
		WHERE a.openl3_embedding IS NOT NULL
		  AND LENGTH(a.openl3_embedding) > 0
		  AND t.id NOT IN (` + string(placeholders) + `)`

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*similarity.TrackFeatures
	for rows.Next() {
		var features similarity.TrackFeatures
		var title, artist sql.NullString

		if err := rows.Scan(
			&features.TrackID, &features.ContentHash, &title, &artist,
			&features.BPM, &features.KeyValue, &features.Energy,
			&features.OpenL3Embedding,
		); err != nil {
			return nil, err
		}

		if title.Valid {
			features.Title = title.String
		}
		if artist.Valid {
			features.Artist = artist.String
		}

		results = append(results, &features)
	}

	return results, rows.Err()
}

// CacheSimilarity stores a computed similarity result.
func (d *DB) CacheSimilarity(trackAID, trackBID int64, openL3Sim, combinedScore, tempoSim, keySim, energySim float64, explanation string) error {
	_, err := d.db.Exec(`
		INSERT INTO embedding_similarity (
			track_a_id, track_b_id, openl3_similarity, combined_score,
			tempo_similarity, key_similarity, energy_similarity, explanation,
			computed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(track_a_id, track_b_id) DO UPDATE SET
			openl3_similarity = excluded.openl3_similarity,
			combined_score = excluded.combined_score,
			tempo_similarity = excluded.tempo_similarity,
			key_similarity = excluded.key_similarity,
			energy_similarity = excluded.energy_similarity,
			explanation = excluded.explanation,
			computed_at = CURRENT_TIMESTAMP
	`, trackAID, trackBID, openL3Sim, combinedScore, tempoSim, keySim, energySim, explanation)
	return err
}

// GetCachedSimilarTracks returns cached similar tracks for a given track.
func (d *DB) GetCachedSimilarTracks(trackID int64, limit int) ([]*similarity.SimilarityResult, error) {
	rows, err := d.db.Query(`
		SELECT s.track_b_id, t.content_hash, t.title, t.artist,
		       s.combined_score, s.openl3_similarity, s.tempo_similarity,
		       s.key_similarity, s.energy_similarity, s.explanation
		FROM embedding_similarity s
		JOIN tracks t ON t.id = s.track_b_id
		WHERE s.track_a_id = ?
		ORDER BY s.combined_score DESC
		LIMIT ?
	`, trackID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*similarity.SimilarityResult
	for rows.Next() {
		var r similarity.SimilarityResult
		var title, artist, explanation sql.NullString

		if err := rows.Scan(
			&r.TrackID, &r.ContentHash, &title, &artist,
			&r.Score, &r.VibeMatch, &r.TempoMatch,
			&r.KeyMatch, &r.EnergyMatch, &explanation,
		); err != nil {
			return nil, err
		}

		if title.Valid {
			r.Title = title.String
		}
		if artist.Valid {
			r.Artist = artist.String
		}
		if explanation.Valid {
			r.Explanation = explanation.String
		}

		results = append(results, &r)
	}

	return results, rows.Err()
}

// GetMLSettings retrieves ML feature settings.
func (d *DB) GetMLSettings() (map[string]string, error) {
	rows, err := d.db.Query(`SELECT setting_key, setting_value FROM ml_settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		settings[key] = value
	}

	return settings, rows.Err()
}

// SetMLSetting updates a single ML setting.
func (d *DB) SetMLSetting(key, value string) error {
	_, err := d.db.Exec(`
		INSERT INTO ml_settings (setting_key, setting_value, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(setting_key) DO UPDATE SET
			setting_value = excluded.setting_value,
			updated_at = CURRENT_TIMESTAMP
	`, key, value)
	return err
}
