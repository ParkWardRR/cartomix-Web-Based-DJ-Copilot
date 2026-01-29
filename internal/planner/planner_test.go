package planner

import (
	"testing"
	"time"

	"github.com/cartomix/cancun/gen/go/common"
	eng "github.com/cartomix/cancun/gen/go/engine"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestPlanWarmupPrefersEnergyClimb(t *testing.T) {
	tracks := []*common.TrackAnalysis{
		buildAnalysis("a", 124, 5, "7A"),
		buildAnalysis("b", 126, 6, "8A"),
		buildAnalysis("c", 128, 7, "9A"),
	}

	order, edges, err := Plan(tracks, Options{Mode: eng.SetMode_WARM_UP})
	if err != nil {
		t.Fatalf("plan failed: %v", err)
	}

	if len(order) != 3 {
		t.Fatalf("expected 3 tracks, got %d", len(order))
	}
	if order[0].ContentHash != "a" {
		t.Errorf("warm-up should start low energy, got %s", order[0].ContentHash)
	}
	if len(edges) != 2 {
		t.Fatalf("expected 2 edges, got %d", len(edges))
	}
	if edges[0].Score <= 0 {
		t.Errorf("expected positive edge score, got %v", edges[0].Score)
	}
}

func TestKeyCompatibilityRespectsJumps(t *testing.T) {
	_, relation := keyCompatibility("8A", "9A", false)
	if relation != "+1 Camelot" {
		t.Fatalf("unexpected relation: %s", relation)
	}

	score, relation := keyCompatibility("8A", "11B", false)
	if score >= 0 {
		t.Fatalf("expected penalty for distant key, got %f (%s)", score, relation)
	}

	score, _ = keyCompatibility("8A", "11B", true)
	if score <= -3 {
		t.Fatalf("allowing jumps should soften penalty, got %f", score)
	}
}

func TestMaxBpmStepPenalty(t *testing.T) {
	from := buildAnalysis("x", 124, 6, "8A")
	to := buildAnalysis("y", 140, 7, "9A")

	score, _ := scoreEdge(from, to, Options{MaxBpmStep: 4})
	if score >= 0 {
		t.Fatalf("expected penalty for bpm jump, got %f", score)
	}
}

func buildAnalysis(hash string, bpm float64, energy int32, key string) *common.TrackAnalysis {
	return &common.TrackAnalysis{
		Id: &common.TrackId{ContentHash: hash, Path: "/tmp/" + hash},
		Beatgrid: &common.Beatgrid{
			Beats: []*common.BeatMarker{
				{Index: 0, Time: durationpb.New(0)},
				{Index: 1, Time: durationpb.New(time.Duration(float64(time.Second) * (60 / bpm)))},
			},
			TempoMap:   []*common.TempoMapNode{{BeatIndex: 0, Bpm: bpm}},
			Confidence: 0.7,
		},
		Key:          &common.MusicalKey{Value: key, Format: common.KeyFormat_CAMELOT, Confidence: 0.9},
		EnergyGlobal: energy,
		TransitionWindows: []*common.TransitionWindow{
			{StartBeat: 0, EndBeat: 32, Tag: "intro"},
			{StartBeat: 300, EndBeat: 332, Tag: "outro"},
		},
		CuePoints: []*common.CuePoint{
			{BeatIndex: 0, Time: durationpb.New(0), Type: common.CueType_CUE_LOAD},
		},
	}
}
