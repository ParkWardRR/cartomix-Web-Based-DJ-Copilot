package storage

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cartomix/cancun/gen/go/common"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestAnalysisRoundTrip(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	dir := t.TempDir()

	db, err := Open(dir, logger)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	track := &Track{
		ContentHash:    "abc123",
		Path:           filepath.Join(dir, "demo.wav"),
		FileSize:       1024,
		FileModifiedAt: time.Now(),
	}
	id, err := db.UpsertTrack(track)
	if err != nil {
		t.Fatalf("upsert track: %v", err)
	}
	track.ID = id

	analysis := &common.TrackAnalysis{
		Id:              &common.TrackId{ContentHash: track.ContentHash, Path: track.Path},
		DurationSeconds: 180,
		Beatgrid: &common.Beatgrid{
			Beats: []*common.BeatMarker{
				{Index: 0, Time: durationpb.New(0)},
				{Index: 1, Time: durationpb.New(500 * time.Millisecond)},
			},
			TempoMap:   []*common.TempoMapNode{{BeatIndex: 0, Bpm: 120}},
			Confidence: 0.8,
		},
		Key:          &common.MusicalKey{Value: "8A", Format: common.KeyFormat_CAMELOT, Confidence: 0.9},
		EnergyGlobal: 7,
		CuePoints: []*common.CuePoint{
			{BeatIndex: 0, Time: durationpb.New(0), Type: common.CueType_CUE_LOAD},
		},
		Loudness: &common.Loudness{IntegratedLufs: -10, TruePeakDb: -1},
	}

	record, err := AnalysisRecordFromProto(track.ID, 1, analysis)
	if err != nil {
		t.Fatalf("record from proto: %v", err)
	}
	if err := db.UpsertAnalysis(record); err != nil {
		t.Fatalf("upsert analysis: %v", err)
	}

	loaded, err := db.LatestCompleteAnalysis(track.ID)
	if err != nil {
		t.Fatalf("latest analysis: %v", err)
	}
	if loaded.GetKey().GetValue() != "8A" {
		t.Fatalf("unexpected key: %s", loaded.GetKey().GetValue())
	}
}

func TestTrackSummariesIncludeCues(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	dir := t.TempDir()
	db, err := Open(dir, logger)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	track := &Track{
		ContentHash:    "hash-summary",
		Path:           filepath.Join(dir, "summary.wav"),
		FileModifiedAt: time.Now(),
	}
	id, _ := db.UpsertTrack(track)
	record := &AnalysisRecord{
		TrackID:       id,
		Version:       1,
		Status:        AnalysisStatusComplete,
		CuePointsJSON: `[{"beatIndex":0,"time":"0s","type":"CUE_LOAD"}]`,
		BPMConfidence: 0.1,
	}
	if err := db.UpsertAnalysis(record); err != nil {
		t.Fatalf("upsert analysis: %v", err)
	}

	summaries, err := db.TrackSummaries("", true, 10)
	if err != nil {
		t.Fatalf("track summaries: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].GetCueCount() != 1 {
		t.Fatalf("expected cue count 1, got %d", summaries[0].GetCueCount())
	}
}

// Ensure migrations table is populated to avoid regression.
func TestMigrationsApplied(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	dir := t.TempDir()
	db, err := Open(dir, logger)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("schema migrations missing: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected at least one migration row")
	}
}
