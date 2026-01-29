package planner

import (
	"math"
	"testing"

	"github.com/cartomix/cancun/gen/go/common"
	eng "github.com/cartomix/cancun/gen/go/engine"
	"google.golang.org/protobuf/types/known/durationpb"
)

// TestPlanOutputContainsAllInputs verifies that the planner doesn't drop tracks.
func TestPlanOutputContainsAllInputs(t *testing.T) {
	testCases := []int{1, 2, 5, 10, 20}

	for _, n := range testCases {
		analyses := generateAnalyses(n)
		order, _, err := Plan(analyses, Options{Mode: eng.SetMode_PEAK_TIME})
		if err != nil {
			t.Errorf("Plan(%d tracks) failed: %v", n, err)
			continue
		}

		if len(order) != n {
			t.Errorf("Plan(%d tracks): expected %d in output, got %d", n, n, len(order))
		}

		// Verify no duplicates
		seen := make(map[string]bool)
		for _, id := range order {
			if seen[id.GetContentHash()] {
				t.Errorf("Plan(%d tracks): duplicate track %s in output", n, id.GetContentHash())
			}
			seen[id.GetContentHash()] = true
		}
	}
}

// TestPlanDeterministic verifies that the same input produces the same output.
func TestPlanDeterministic(t *testing.T) {
	analyses := generateAnalyses(10)
	opts := Options{Mode: eng.SetMode_WARM_UP}

	order1, _, err1 := Plan(analyses, opts)
	if err1 != nil {
		t.Fatalf("first Plan() failed: %v", err1)
	}

	order2, _, err2 := Plan(analyses, opts)
	if err2 != nil {
		t.Fatalf("second Plan() failed: %v", err2)
	}

	if len(order1) != len(order2) {
		t.Fatal("determinism failed: different lengths")
	}

	for i := range order1 {
		if order1[i].GetContentHash() != order2[i].GetContentHash() {
			t.Errorf("determinism failed at index %d: %s != %s",
				i, order1[i].GetContentHash(), order2[i].GetContentHash())
		}
	}
}

// TestBanExcludesTracks verifies that banned tracks are excluded.
func TestBanExcludesTracks(t *testing.T) {
	analyses := generateAnalyses(5)
	bannedHash := analyses[2].GetId().GetContentHash()

	opts := Options{
		Mode:      eng.SetMode_PEAK_TIME,
		BanHashes: map[string]bool{bannedHash: true},
	}

	order, _, err := Plan(analyses, opts)
	if err != nil {
		t.Fatalf("Plan() failed: %v", err)
	}

	if len(order) != 4 {
		t.Errorf("expected 4 tracks after ban, got %d", len(order))
	}

	for _, id := range order {
		if id.GetContentHash() == bannedHash {
			t.Error("banned track appeared in output")
		}
	}
}

// TestMustPlayIncluded verifies that must-play tracks are included.
func TestMustPlayIncluded(t *testing.T) {
	analyses := generateAnalyses(10)
	mustPlayHash := analyses[5].GetId().GetContentHash()

	opts := Options{
		Mode:           eng.SetMode_OPEN_FORMAT,
		MustPlayHashes: map[string]bool{mustPlayHash: true},
	}

	order, _, err := Plan(analyses, opts)
	if err != nil {
		t.Fatalf("Plan() failed: %v", err)
	}

	found := false
	for _, id := range order {
		if id.GetContentHash() == mustPlayHash {
			found = true
			break
		}
	}

	if !found {
		t.Error("must-play track not found in output")
	}
}

// TestExplanationsMatchOrder verifies that explanations correspond to transitions.
func TestExplanationsMatchOrder(t *testing.T) {
	analyses := generateAnalyses(5)
	order, explanations, err := Plan(analyses, Options{Mode: eng.SetMode_PEAK_TIME})
	if err != nil {
		t.Fatalf("Plan() failed: %v", err)
	}

	if len(order) < 2 {
		t.Skip("not enough tracks for transition explanations")
	}

	// Explanations should be n-1 for n tracks
	expectedExpls := len(order) - 1
	if len(explanations) != expectedExpls {
		t.Errorf("expected %d explanations for %d tracks, got %d",
			expectedExpls, len(order), len(explanations))
	}

	// Each explanation should connect consecutive tracks
	for i, expl := range explanations {
		if i >= len(order)-1 {
			break
		}
		fromHash := order[i].GetContentHash()
		toHash := order[i+1].GetContentHash()

		if expl.GetFrom().GetContentHash() != fromHash {
			t.Errorf("explanation %d: from hash mismatch", i)
		}
		if expl.GetTo().GetContentHash() != toHash {
			t.Errorf("explanation %d: to hash mismatch", i)
		}
	}
}

// TestKeyCompatibilitySymmetric verifies key compatibility properties.
func TestKeyCompatibilityProperties(t *testing.T) {
	testCases := []struct {
		key1, key2 string
	}{
		{"8A", "8A"},   // Same key
		{"8A", "9A"},   // Adjacent
		{"8A", "7A"},   // Adjacent
		{"8A", "8B"},   // Same wheel, different mode
		{"1A", "12A"},  // Wrap-around
	}

	for _, tc := range testCases {
		score1, _ := keyCompatibility(tc.key1, tc.key2, false)
		score2, _ := keyCompatibility(tc.key2, tc.key1, false)

		// Key compatibility should be symmetric
		if math.Abs(score1-score2) > 0.001 {
			t.Errorf("key compatibility not symmetric: (%s, %s) = %f, (%s, %s) = %f",
				tc.key1, tc.key2, score1, tc.key2, tc.key1, score2)
		}
	}
}

// TestScoreEdgeBounds verifies that edge scores are bounded.
func TestScoreEdgeBounds(t *testing.T) {
	analyses := generateAnalyses(20)
	opts := Options{Mode: eng.SetMode_PEAK_TIME}

	for i := 0; i < len(analyses)-1; i++ {
		score, expl := scoreEdge(analyses[i], analyses[i+1], opts)

		// Score should be finite
		if math.IsNaN(score) || math.IsInf(score, 0) {
			t.Errorf("invalid score for edge %d->%d: %f", i, i+1, score)
		}

		// Proto score should match
		if math.Abs(float64(expl.GetScore())-score) > 0.001 {
			t.Errorf("score mismatch: func returned %f, proto has %f", score, expl.GetScore())
		}
	}
}

// TestWarmUpModePreferencesLowEnergyStart verifies warm-up mode behavior.
func TestWarmUpModePreferencesLowEnergyStart(t *testing.T) {
	// Create tracks with varying energy
	analyses := []*common.TrackAnalysis{
		makeAnalysis("high1", 128, "8A", 9),
		makeAnalysis("high2", 130, "8A", 8),
		makeAnalysis("low1", 120, "8A", 3),
		makeAnalysis("low2", 122, "8A", 4),
		makeAnalysis("mid", 125, "8A", 6),
	}

	order, _, err := Plan(analyses, Options{Mode: eng.SetMode_WARM_UP})
	if err != nil {
		t.Fatalf("Plan() failed: %v", err)
	}

	// First track should be one of the low energy ones
	firstHash := order[0].GetContentHash()
	firstEnergy := findEnergy(analyses, firstHash)

	if firstEnergy > 5 {
		t.Errorf("warm-up mode started with high energy track (energy=%d)", firstEnergy)
	}
}

// TestPeakTimeModePreferencesHighEnergyStart verifies peak-time mode behavior.
func TestPeakTimeModePreferencesHighEnergyStart(t *testing.T) {
	// Create tracks with varying energy
	analyses := []*common.TrackAnalysis{
		makeAnalysis("high1", 128, "8A", 9),
		makeAnalysis("high2", 130, "8A", 8),
		makeAnalysis("low1", 120, "8A", 3),
		makeAnalysis("low2", 122, "8A", 4),
		makeAnalysis("mid", 125, "8A", 6),
	}

	order, _, err := Plan(analyses, Options{Mode: eng.SetMode_PEAK_TIME})
	if err != nil {
		t.Fatalf("Plan() failed: %v", err)
	}

	// First track should be one of the high energy ones
	firstHash := order[0].GetContentHash()
	firstEnergy := findEnergy(analyses, firstHash)

	if firstEnergy < 7 {
		t.Errorf("peak-time mode started with low energy track (energy=%d)", firstEnergy)
	}
}

// Helper functions

func generateAnalyses(n int) []*common.TrackAnalysis {
	analyses := make([]*common.TrackAnalysis, n)
	keys := []string{"1A", "2A", "3A", "4A", "5A", "6A", "7A", "8A", "9A", "10A", "11A", "12A"}

	for i := 0; i < n; i++ {
		bpm := 120.0 + float64(i%20)*2.0 // 120-158 BPM
		key := keys[i%len(keys)]
		energy := int32((i % 10) + 1) // 1-10 energy

		analyses[i] = makeAnalysis(
			string(rune('a'+i)),
			bpm,
			key,
			energy,
		)
	}
	return analyses
}

func makeAnalysis(hash string, bpm float64, key string, energy int32) *common.TrackAnalysis {
	return &common.TrackAnalysis{
		Id: &common.TrackId{
			ContentHash: hash,
			Path:        "/test/" + hash + ".mp3",
		},
		Key: &common.MusicalKey{
			Value:      key,
			Format:     common.KeyFormat_CAMELOT,
			Confidence: 0.9,
		},
		Beatgrid: &common.Beatgrid{
			Beats: []*common.BeatMarker{
				{Index: 0, Time: durationpb.New(0), IsDownbeat: true},
				{Index: 100, Time: durationpb.New(180000000000), IsDownbeat: true},
			},
			TempoMap: []*common.TempoMapNode{
				{BeatIndex: 0, Bpm: bpm},
			},
			Confidence: 0.85,
		},
		EnergyGlobal: energy,
		TransitionWindows: []*common.TransitionWindow{
			{StartBeat: 0, EndBeat: 16, Tag: "intro", Confidence: 0.8},
		},
	}
}

func findEnergy(analyses []*common.TrackAnalysis, hash string) int32 {
	for _, a := range analyses {
		if a.GetId().GetContentHash() == hash {
			return a.GetEnergyGlobal()
		}
	}
	return 0
}
