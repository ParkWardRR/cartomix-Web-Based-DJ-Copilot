package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cartomix/cancun/gen/go/common"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

// AnalysisStatus represents the lifecycle of an analysis row.
type AnalysisStatus string

const (
	AnalysisStatusPending  AnalysisStatus = "pending"
	AnalysisStatusRunning  AnalysisStatus = "analyzing"
	AnalysisStatusComplete AnalysisStatus = "complete"
	AnalysisStatusFailed   AnalysisStatus = "failed"
)

// AnalysisRecord mirrors the analyses table.
type AnalysisRecord struct {
	ID                    int64
	TrackID               int64
	Version               int32
	Status                AnalysisStatus
	Error                 string
	DurationSeconds       float64
	BPM                   float64
	BPMConfidence         float64
	IsDynamicTempo        bool
	KeyValue              string
	KeyFormat             string
	KeyConfidence         float64
	EnergyGlobal          int32
	IntegratedLufs        float64
	TruePeakDb            float64
	BeatgridJSON          string
	SectionsJSON          string
	CuePointsJSON         string
	EnergySegmentsJSON    string
	TransitionWindowsJSON string
	TempoMapJSON          string
	Embedding             []byte
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// AnalysisRecordFromProto builds a record ready for persistence.
func AnalysisRecordFromProto(trackID int64, version int32, analysis *common.TrackAnalysis) (*AnalysisRecord, error) {
	if analysis == nil {
		return nil, errors.New("analysis is required")
	}

	beatgridJSON, err := marshalProto(analysis.Beatgrid)
	if err != nil {
		return nil, fmt.Errorf("marshal beatgrid: %w", err)
	}
	sectionsJSON, err := marshalProtoSlice(analysis.Sections)
	if err != nil {
		return nil, fmt.Errorf("marshal sections: %w", err)
	}
	cuesJSON, err := marshalProtoSlice(analysis.CuePoints)
	if err != nil {
		return nil, fmt.Errorf("marshal cues: %w", err)
	}
	energySegmentsJSON, err := marshalProtoSlice(analysis.EnergySegments)
	if err != nil {
		return nil, fmt.Errorf("marshal energy segments: %w", err)
	}
	transitionJSON, err := marshalProtoSlice(analysis.TransitionWindows)
	if err != nil {
		return nil, fmt.Errorf("marshal transition windows: %w", err)
	}
	tempoMapJSON, err := marshalProtoSlice(analysis.Beatgrid.GetTempoMap())
	if err != nil {
		return nil, fmt.Errorf("marshal tempo map: %w", err)
	}

	keyFormat := analysis.GetKey().GetFormat().String()
	if keyFormat == "" || keyFormat == "KEY_FORMAT_UNSPECIFIED" {
		keyFormat = "camelot" // default for UI expectations
	}

	record := &AnalysisRecord{
		TrackID:               trackID,
		Version:               version,
		Status:                AnalysisStatusComplete,
		DurationSeconds:       analysis.GetDurationSeconds(),
		BPM:                   inferBPM(analysis),
		BPMConfidence:         float64(analysis.GetBeatgrid().GetConfidence()),
		IsDynamicTempo:        analysis.GetBeatgrid().GetIsDynamic(),
		KeyValue:              analysis.GetKey().GetValue(),
		KeyFormat:             keyFormat,
		KeyConfidence:         float64(analysis.GetKey().GetConfidence()),
		EnergyGlobal:          analysis.GetEnergyGlobal(),
		IntegratedLufs:        float64(analysis.GetLoudness().GetIntegratedLufs()),
		TruePeakDb:            float64(analysis.GetLoudness().GetTruePeakDb()),
		BeatgridJSON:          beatgridJSON,
		SectionsJSON:          sectionsJSON,
		CuePointsJSON:         cuesJSON,
		EnergySegmentsJSON:    energySegmentsJSON,
		TransitionWindowsJSON: transitionJSON,
		TempoMapJSON:          tempoMapJSON,
		Embedding:             analysis.GetEmbedding(),
	}

	return record, nil
}

// UpsertAnalysis writes or updates an analysis row (identified by track_id + version).
func (d *DB) UpsertAnalysis(rec *AnalysisRecord) error {
	_, err := d.db.Exec(`
		INSERT INTO analyses (
			track_id, version, status, error,
			duration_seconds, bpm, bpm_confidence, is_dynamic_tempo,
			key_value, key_format, key_confidence,
			energy_global, integrated_lufs, true_peak_db,
			beatgrid_json, sections_json, cue_points_json, energy_segments_json, transition_windows_json, tempo_map_json,
			embedding, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(track_id, version) DO UPDATE SET
			status = excluded.status,
			error = excluded.error,
			duration_seconds = excluded.duration_seconds,
			bpm = excluded.bpm,
			bpm_confidence = excluded.bpm_confidence,
			is_dynamic_tempo = excluded.is_dynamic_tempo,
			key_value = excluded.key_value,
			key_format = excluded.key_format,
			key_confidence = excluded.key_confidence,
			energy_global = excluded.energy_global,
			integrated_lufs = excluded.integrated_lufs,
			true_peak_db = excluded.true_peak_db,
			beatgrid_json = excluded.beatgrid_json,
			sections_json = excluded.sections_json,
			cue_points_json = excluded.cue_points_json,
			energy_segments_json = excluded.energy_segments_json,
			transition_windows_json = excluded.transition_windows_json,
			tempo_map_json = excluded.tempo_map_json,
			embedding = excluded.embedding,
			updated_at = CURRENT_TIMESTAMP
	`, rec.TrackID, rec.Version, rec.Status, rec.Error,
		rec.DurationSeconds, rec.BPM, rec.BPMConfidence, rec.IsDynamicTempo,
		rec.KeyValue, rec.KeyFormat, rec.KeyConfidence,
		rec.EnergyGlobal, rec.IntegratedLufs, rec.TruePeakDb,
		rec.BeatgridJSON, rec.SectionsJSON, rec.CuePointsJSON, rec.EnergySegmentsJSON, rec.TransitionWindowsJSON, rec.TempoMapJSON,
		rec.Embedding)

	return err
}

// MarkAnalysisFailure records a failed analysis attempt with the given version.
func (d *DB) MarkAnalysisFailure(trackID int64, version int32, errMsg string) error {
	_, err := d.db.Exec(`
		INSERT INTO analyses (track_id, version, status, error, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(track_id, version) DO UPDATE SET
			status = excluded.status,
			error = excluded.error,
			updated_at = CURRENT_TIMESTAMP
	`, trackID, version, string(AnalysisStatusFailed), errMsg)
	return err
}

// LatestAnalysisRecord fetches the most recent analysis (any status) for a track.
func (d *DB) LatestAnalysisRecord(trackID int64) (*AnalysisRecord, error) {
	row := d.db.QueryRow(`
		SELECT id, track_id, version, status, error, duration_seconds, bpm, bpm_confidence, is_dynamic_tempo,
		       key_value, key_format, key_confidence, energy_global, integrated_lufs, true_peak_db,
		       beatgrid_json, sections_json, cue_points_json, energy_segments_json, transition_windows_json, tempo_map_json,
		       embedding, created_at, updated_at
		FROM analyses
		WHERE track_id = ?
		ORDER BY version DESC
		LIMIT 1
	`, trackID)

	rec := &AnalysisRecord{}
	var status string
	var createdAt, updatedAt sql.NullTime
	if err := row.Scan(
		&rec.ID, &rec.TrackID, &rec.Version, &status, &rec.Error, &rec.DurationSeconds, &rec.BPM, &rec.BPMConfidence, &rec.IsDynamicTempo,
		&rec.KeyValue, &rec.KeyFormat, &rec.KeyConfidence, &rec.EnergyGlobal, &rec.IntegratedLufs, &rec.TruePeakDb,
		&rec.BeatgridJSON, &rec.SectionsJSON, &rec.CuePointsJSON, &rec.EnergySegmentsJSON, &rec.TransitionWindowsJSON, &rec.TempoMapJSON,
		&rec.Embedding, &createdAt, &updatedAt,
	); err != nil {
		return nil, err
	}

	rec.Status = AnalysisStatus(status)
	if createdAt.Valid {
		rec.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		rec.UpdatedAt = updatedAt.Time
	}
	return rec, nil
}

// LatestCompleteAnalysis returns the latest completed analysis proto for a track.
func (d *DB) LatestCompleteAnalysis(trackID int64) (*common.TrackAnalysis, error) {
	rec, err := d.latestByStatus(trackID, AnalysisStatusComplete)
	if err != nil {
		return nil, err
	}
	track, err := d.GetTrackByID(trackID)
	if err != nil {
		return nil, err
	}
	return rec.ToProto(track)
}

// ToProto converts an AnalysisRecord back into the protobuf representation.
func (rec *AnalysisRecord) ToProto(track *Track) (*common.TrackAnalysis, error) {
	if track == nil {
		return nil, errors.New("track is required for analysis reconstruction")
	}

	analysis := &common.TrackAnalysis{
		Id: &common.TrackId{
			ContentHash: track.ContentHash,
			Path:        track.Path,
		},
		DurationSeconds: rec.DurationSeconds,
		EnergyGlobal:    rec.EnergyGlobal,
		AnalysisVersion: rec.Version,
		Loudness: &common.Loudness{
			IntegratedLufs: float32(rec.IntegratedLufs),
			TruePeakDb:     float32(rec.TruePeakDb),
		},
		Embedding: rec.Embedding,
	}

	if rec.BeatgridJSON != "" {
		analysis.Beatgrid = &common.Beatgrid{}
		if err := unmarshalProto(rec.BeatgridJSON, analysis.Beatgrid); err != nil {
			return nil, fmt.Errorf("unmarshal beatgrid: %w", err)
		}
	}

	if rec.SectionsJSON != "" {
		if err := unmarshalRepeated(rec.SectionsJSON, func() proto.Message { return &common.Section{} }, func(msg proto.Message) {
			analysis.Sections = append(analysis.Sections, msg.(*common.Section))
		}); err != nil {
			return nil, fmt.Errorf("unmarshal sections: %w", err)
		}
	}

	if rec.CuePointsJSON != "" {
		if err := unmarshalRepeated(rec.CuePointsJSON, func() proto.Message { return &common.CuePoint{} }, func(msg proto.Message) {
			analysis.CuePoints = append(analysis.CuePoints, msg.(*common.CuePoint))
		}); err != nil {
			return nil, fmt.Errorf("unmarshal cues: %w", err)
		}
	}

	if rec.EnergySegmentsJSON != "" {
		if err := unmarshalRepeated(rec.EnergySegmentsJSON, func() proto.Message { return &common.EnergySegment{} }, func(msg proto.Message) {
			analysis.EnergySegments = append(analysis.EnergySegments, msg.(*common.EnergySegment))
		}); err != nil {
			return nil, fmt.Errorf("unmarshal energy segments: %w", err)
		}
	}

	if rec.TransitionWindowsJSON != "" {
		if err := unmarshalRepeated(rec.TransitionWindowsJSON, func() proto.Message { return &common.TransitionWindow{} }, func(msg proto.Message) {
			analysis.TransitionWindows = append(analysis.TransitionWindows, msg.(*common.TransitionWindow))
		}); err != nil {
			return nil, fmt.Errorf("unmarshal transition windows: %w", err)
		}
	}

	analysis.Key = &common.MusicalKey{
		Value:      rec.KeyValue,
		Format:     keyFormatFromString(rec.KeyFormat),
		Confidence: float32(rec.KeyConfidence),
	}

	return analysis, nil
}

// TrackSummaries returns library summaries joined with the latest analysis.
func (d *DB) TrackSummaries(query string, needsGridReview bool, limit int) ([]*common.TrackSummary, error) {
	conditions := []string{}
	args := []any{}

	if query != "" {
		conditions = append(conditions, "(t.title LIKE ? OR t.artist LIKE ? OR t.path LIKE ?)")
		pattern := "%" + query + "%"
		args = append(args, pattern, pattern, pattern)
	}

	if needsGridReview {
		conditions = append(conditions, "(a.bpm_confidence < 0.5)")
	}

	sqlStr := `
		SELECT t.id, t.content_hash, t.path, t.title, t.artist,
		       COALESCE(a.bpm, 0),
		       COALESCE(a.key_value, ''),
		       COALESCE(a.key_format, ''),
		       COALESCE(a.energy_global, 0),
		       COALESCE(a.cue_points_json, ''),
		       COALESCE(a.status, 'pending'),
		       COALESCE(a.bpm_confidence, 0)
		FROM tracks t
		LEFT JOIN analyses a ON a.id = (
			SELECT id FROM analyses a2 WHERE a2.track_id = t.id ORDER BY a2.version DESC LIMIT 1
		)
	`

	if len(conditions) > 0 {
		sqlStr += " WHERE " + strings.Join(conditions, " AND ")
	}

	sqlStr += " ORDER BY COALESCE(a.updated_at, t.updated_at) DESC"

	if limit > 0 {
		sqlStr += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := d.db.Query(sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []*common.TrackSummary
	for rows.Next() {
		var (
			id            sql.NullInt64
			contentHash   sql.NullString
			path          sql.NullString
			title         sql.NullString
			artist        sql.NullString
			bpm           sql.NullFloat64
			keyValue      sql.NullString
			keyFormat     sql.NullString
			energyGlobal  sql.NullInt64
			cuesJSON      sql.NullString
			status        sql.NullString
			bpmConfidence sql.NullFloat64
		)

		if err := rows.Scan(&id, &contentHash, &path, &title, &artist, &bpm, &keyValue, &keyFormat, &energyGlobal, &cuesJSON, &status, &bpmConfidence); err != nil {
			return nil, err
		}
		_ = id

		summary := &common.TrackSummary{
			Id: &common.TrackId{
				ContentHash: contentHash.String,
				Path:        path.String,
			},
			Title: title.String,
			Artist: func() string {
				if artist.Valid {
					return artist.String
				}
				return ""
			}(),
			Bpm:    bpm.Float64,
			Key:    &common.MusicalKey{Value: keyValue.String, Format: keyFormatFromString(keyFormat.String), Confidence: 1},
			Energy: int32(energyGlobal.Int64),
			Status: func() string {
				if status.Valid {
					return status.String
				}
				return "pending"
			}(),
		}

		if cuesJSON.Valid && cuesJSON.String != "" {
			var raw []json.RawMessage
			if err := json.Unmarshal([]byte(cuesJSON.String), &raw); err == nil {
				summary.CueCount = int32(len(raw))
			}
		}

		// If grid review is requested but no analysis exists, skip.
		if needsGridReview && (!bpmConfidence.Valid || bpmConfidence.Float64 >= 0.5) {
			continue
		}

		summaries = append(summaries, summary)
	}

	return summaries, rows.Err()
}

// latestByStatus fetches the latest analysis matching the given status.
func (d *DB) latestByStatus(trackID int64, status AnalysisStatus) (*AnalysisRecord, error) {
	row := d.db.QueryRow(`
		SELECT id, track_id, version, status, error, duration_seconds, bpm, bpm_confidence, is_dynamic_tempo,
		       key_value, key_format, key_confidence, energy_global, integrated_lufs, true_peak_db,
		       beatgrid_json, sections_json, cue_points_json, energy_segments_json, transition_windows_json, tempo_map_json,
		       embedding, created_at, updated_at
		FROM analyses
		WHERE track_id = ? AND status = ?
		ORDER BY version DESC
		LIMIT 1
	`, trackID, string(status))

	rec := &AnalysisRecord{}
	var statusStr string
	var createdAt, updatedAt sql.NullTime
	if err := row.Scan(
		&rec.ID, &rec.TrackID, &rec.Version, &statusStr, &rec.Error, &rec.DurationSeconds, &rec.BPM, &rec.BPMConfidence, &rec.IsDynamicTempo,
		&rec.KeyValue, &rec.KeyFormat, &rec.KeyConfidence, &rec.EnergyGlobal, &rec.IntegratedLufs, &rec.TruePeakDb,
		&rec.BeatgridJSON, &rec.SectionsJSON, &rec.CuePointsJSON, &rec.EnergySegmentsJSON, &rec.TransitionWindowsJSON, &rec.TempoMapJSON,
		&rec.Embedding, &createdAt, &updatedAt,
	); err != nil {
		return nil, err
	}

	rec.Status = AnalysisStatus(statusStr)
	if createdAt.Valid {
		rec.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		rec.UpdatedAt = updatedAt.Time
	}
	return rec, nil
}

func marshalProto(msg proto.Message) (string, error) {
	if msg == nil {
		return "", nil
	}
	bytes, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(msg)
	return string(bytes), err
}

func marshalProtoSlice[T proto.Message](items []T) (string, error) {
	if len(items) == 0 {
		return "", nil
	}

	var raw []json.RawMessage
	for _, item := range items {
		bytes, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(item)
		if err != nil {
			return "", err
		}
		raw = append(raw, json.RawMessage(bytes))
	}

	bytes, err := json.Marshal(raw)
	return string(bytes), err
}

func unmarshalProto(data string, msg proto.Message) error {
	return protojson.Unmarshal([]byte(data), msg)
}

func unmarshalRepeated(data string, factory func() proto.Message, sink func(proto.Message)) error {
	var raw []json.RawMessage
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return err
	}
	for _, r := range raw {
		msg := factory()
		if err := protojson.Unmarshal(r, msg); err != nil {
			return err
		}
		sink(msg)
	}
	return nil
}

func inferBPM(analysis *common.TrackAnalysis) float64 {
	if analysis.GetBeatgrid() == nil {
		return 0
	}
	if tm := analysis.GetBeatgrid().GetTempoMap(); len(tm) > 0 {
		return tm[0].GetBpm()
	}
	beats := analysis.GetBeatgrid().GetBeats()
	if len(beats) >= 2 {
		first := beats[0].GetTime().AsDuration()
		second := beats[1].GetTime().AsDuration()
		if delta := second - first; delta > 0 {
			return 60.0 / delta.Seconds()
		}
	}
	return 0
}

func keyFormatFromString(value string) common.KeyFormat {
	switch strings.ToLower(value) {
	case "camelot", "key_format_camelot":
		return common.KeyFormat_CAMELOT
	case "open_key", "openkey", "key_format_open_key":
		return common.KeyFormat_OPEN_KEY
	default:
		return common.KeyFormat_KEY_FORMAT_UNSPECIFIED
	}
}

// SafeDuration converts a protobuf duration pointer to a time.Duration.
func SafeDuration(d *durationpb.Duration) time.Duration {
	if d == nil {
		return 0
	}
	return d.AsDuration()
}
