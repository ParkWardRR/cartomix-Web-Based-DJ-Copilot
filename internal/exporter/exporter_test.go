package exporter

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cartomix/cancun/gen/go/common"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestWriteGenericCreatesArtifacts(t *testing.T) {
	dir := t.TempDir()

	tracks := []TrackExport{
		{
			Path: "/music/test.wav",
			Analysis: &common.TrackAnalysis{
				Id: &common.TrackId{ContentHash: "hash1", Path: "/music/test.wav"},
				CuePoints: []*common.CuePoint{
					{BeatIndex: 0, Time: durationpb.New(time.Second), Type: common.CueType_CUE_LOAD},
				},
			},
		},
	}

	res, err := WriteGeneric(dir, "demo", tracks)
	if err != nil {
		t.Fatalf("write generic failed: %v", err)
	}

	for _, path := range []string{res.PlaylistPath, res.AnalysisJSONPath, res.CuesCSVPath} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected file %s: %v", path, err)
		}
	}

	playlistBytes, _ := os.ReadFile(res.PlaylistPath)
	if len(playlistBytes) == 0 {
		t.Fatalf("playlist should not be empty")
	}

	f, err := os.Open(res.CuesCSVPath)
	if err != nil {
		t.Fatalf("open csv: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("csv read: %v", err)
	}
	if len(rows) < 2 {
		t.Fatalf("expected header + at least one cue row, got %d", len(rows))
	}

	if filepath.Ext(res.PlaylistPath) != ".m3u8" {
		t.Fatalf("expected m3u8 playlist, got %s", res.PlaylistPath)
	}
}
