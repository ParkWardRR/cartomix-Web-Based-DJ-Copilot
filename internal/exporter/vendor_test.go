package exporter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cartomix/cancun/gen/go/common"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestWriteRekordboxCreatesXML(t *testing.T) {
	dir := t.TempDir()
	tracks := makeTestTracks()

	path, err := WriteRekordbox(dir, "test-set", tracks)
	if err != nil {
		t.Fatalf("WriteRekordbox failed: %v", err)
	}

	if !strings.HasSuffix(path, "-rekordbox.xml") {
		t.Errorf("expected path ending in -rekordbox.xml, got %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "DJ_PLAYLISTS") {
		t.Error("expected DJ_PLAYLISTS element in Rekordbox XML")
	}
	if !strings.Contains(content, "COLLECTION") {
		t.Error("expected COLLECTION element in Rekordbox XML")
	}
	if !strings.Contains(content, "POSITION_MARK") {
		t.Error("expected POSITION_MARK element for cue points")
	}
}

func TestWriteSeratoCreatesCrate(t *testing.T) {
	dir := t.TempDir()
	tracks := makeTestTracks()

	path, err := WriteSerato(dir, "test-set", tracks)
	if err != nil {
		t.Fatalf("WriteSerato failed: %v", err)
	}

	if !strings.HasSuffix(path, ".crate") {
		t.Errorf("expected path ending in .crate, got %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	// Check for Serato crate magic bytes
	if len(data) < 4 {
		t.Error("crate file too small")
	}
	if string(data[:4]) != "vrsn" {
		t.Errorf("expected vrsn magic, got %q", string(data[:4]))
	}
}

func TestWriteTraktorCreatesNML(t *testing.T) {
	dir := t.TempDir()
	tracks := makeTestTracks()

	path, err := WriteTraktor(dir, "test-set", tracks)
	if err != nil {
		t.Fatalf("WriteTraktor failed: %v", err)
	}

	if !strings.HasSuffix(path, "-traktor.nml") {
		t.Errorf("expected path ending in -traktor.nml, got %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "<NML") {
		t.Error("expected NML root element in Traktor export")
	}
	if !strings.Contains(content, "COLLECTION") {
		t.Error("expected COLLECTION element in Traktor NML")
	}
	if !strings.Contains(content, "CUE_V2") {
		t.Error("expected CUE_V2 element for cue points")
	}
}

func TestCamelotToRekordbox(t *testing.T) {
	tests := []struct {
		camelot string
		want    string
	}{
		{"8A", "Am"},
		{"8B", "C"},
		{"1A", "Abm"},
		{"5B", "Eb"},
	}

	for _, tc := range tests {
		got := camelotToRekordbox(tc.camelot)
		if got != tc.want {
			t.Errorf("camelotToRekordbox(%s) = %s, want %s", tc.camelot, got, tc.want)
		}
	}
}

func TestCamelotToTraktorKey(t *testing.T) {
	tests := []struct {
		camelot string
		want    int
	}{
		{"8A", 21}, // Am
		{"8B", 0},  // C
		{"1A", 20}, // Abm
		{"5B", 3},  // Eb
	}

	for _, tc := range tests {
		got := camelotToTraktorKey(tc.camelot)
		if got != tc.want {
			t.Errorf("camelotToTraktorKey(%s) = %d, want %d", tc.camelot, got, tc.want)
		}
	}
}

func makeTestTracks() []TrackExport {
	return []TrackExport{
		{
			Path: filepath.Join(os.TempDir(), "track1.mp3"),
			Analysis: &common.TrackAnalysis{
				Id: &common.TrackId{
					ContentHash: "hash1",
					Path:        filepath.Join(os.TempDir(), "track1.mp3"),
				},
				Key: &common.MusicalKey{
					Value:      "8A",
					Format:     common.KeyFormat_CAMELOT,
					Confidence: 0.9,
				},
				Beatgrid: &common.Beatgrid{
					Beats: []*common.BeatMarker{
						{Index: 0, Time: durationpb.New(0), IsDownbeat: true},
						{Index: 100, Time: durationpb.New(180000000000), IsDownbeat: true}, // 3 minutes
					},
					TempoMap: []*common.TempoMapNode{
						{BeatIndex: 0, Bpm: 128.0},
					},
					Confidence: 0.85,
				},
				CuePoints: []*common.CuePoint{
					{BeatIndex: 0, Time: durationpb.New(0), Type: common.CueType_CUE_LOAD},
					{BeatIndex: 32, Time: durationpb.New(15000000000), Type: common.CueType_CUE_DROP},
				},
				EnergyGlobal: 7,
			},
		},
		{
			Path: filepath.Join(os.TempDir(), "track2.wav"),
			Analysis: &common.TrackAnalysis{
				Id: &common.TrackId{
					ContentHash: "hash2",
					Path:        filepath.Join(os.TempDir(), "track2.wav"),
				},
				Key: &common.MusicalKey{
					Value:      "9A",
					Format:     common.KeyFormat_CAMELOT,
					Confidence: 0.88,
				},
				Beatgrid: &common.Beatgrid{
					Beats: []*common.BeatMarker{
						{Index: 0, Time: durationpb.New(0), IsDownbeat: true},
						{Index: 120, Time: durationpb.New(210000000000), IsDownbeat: true}, // 3.5 minutes
					},
					TempoMap: []*common.TempoMapNode{
						{BeatIndex: 0, Bpm: 130.0},
					},
					Confidence: 0.9,
				},
				CuePoints: []*common.CuePoint{
					{BeatIndex: 0, Time: durationpb.New(0), Type: common.CueType_CUE_INTRO_START},
					{BeatIndex: 64, Time: durationpb.New(30000000000), Type: common.CueType_CUE_BUILD},
				},
				EnergyGlobal: 8,
			},
		},
	}
}
