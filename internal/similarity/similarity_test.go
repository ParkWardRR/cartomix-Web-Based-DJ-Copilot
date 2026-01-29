package similarity

import (
	"math"
	"testing"
)

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        []float32
		b        []float32
		wantMin  float64
		wantMax  float64
	}{
		{
			name:    "identical vectors",
			a:       []float32{1, 2, 3, 4, 5},
			b:       []float32{1, 2, 3, 4, 5},
			wantMin: 0.99,
			wantMax: 1.01,
		},
		{
			name:    "orthogonal vectors",
			a:       []float32{1, 0, 0},
			b:       []float32{0, 1, 0},
			wantMin: 0.49,
			wantMax: 0.51,
		},
		{
			name:    "opposite vectors",
			a:       []float32{1, 2, 3},
			b:       []float32{-1, -2, -3},
			wantMin: -0.01,
			wantMax: 0.01,
		},
		{
			name:    "empty vectors",
			a:       []float32{},
			b:       []float32{},
			wantMin: -0.01,
			wantMax: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeCosineSimilarity(tt.a, tt.b)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("computeCosineSimilarity() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestComputeTempoSimilarity(t *testing.T) {
	tests := []struct {
		name    string
		bpmA    float64
		bpmB    float64
		wantMin float64
	}{
		{"same tempo", 128.0, 128.0, 1.0},
		{"1 BPM diff", 128.0, 129.0, 0.89},
		{"5 BPM diff", 128.0, 133.0, 0.49},
		{"10 BPM diff", 128.0, 138.0, 0.0},
		{"half tempo match", 128.0, 64.0, 0.99},
		{"double tempo match", 64.0, 128.0, 0.99},
		{"zero tempo", 0, 128.0, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeTempoSimilarity(tt.bpmA, tt.bpmB)
			if got < tt.wantMin {
				t.Errorf("computeTempoSimilarity(%v, %v) = %v, want >= %v", tt.bpmA, tt.bpmB, got, tt.wantMin)
			}
		})
	}
}

func TestComputeKeySimilarity(t *testing.T) {
	tests := []struct {
		name         string
		keyA         string
		keyB         string
		wantScore    float64
		wantRelation string
	}{
		{"same key", "8A", "8A", 1.0, "same"},
		{"relative major/minor", "8A", "8B", 0.9, "relative"},
		{"adjacent key", "8A", "9A", 0.85, "compatible"},
		{"adjacent key wrap", "12A", "1A", 0.85, "compatible"},
		{"distant key", "8A", "2A", 0.2, "clash"},
		{"unknown key", "", "8A", 0.5, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, relation := computeKeySimilarity(tt.keyA, tt.keyB)
			if math.Abs(score-tt.wantScore) > 0.01 {
				t.Errorf("computeKeySimilarity(%v, %v) score = %v, want %v", tt.keyA, tt.keyB, score, tt.wantScore)
			}
			if relation != tt.wantRelation {
				t.Errorf("computeKeySimilarity(%v, %v) relation = %v, want %v", tt.keyA, tt.keyB, relation, tt.wantRelation)
			}
		})
	}
}

func TestComputeEnergySimilarity(t *testing.T) {
	tests := []struct {
		name    string
		energyA int32
		energyB int32
		want    float64
	}{
		{"same energy", 5, 5, 1.0},
		{"1 diff", 5, 6, 0.8},
		{"2 diff", 5, 7, 0.6},
		{"5+ diff", 1, 10, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeEnergySimilarity(tt.energyA, tt.energyB)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("computeEnergySimilarity(%v, %v) = %v, want %v", tt.energyA, tt.energyB, got, tt.want)
			}
		})
	}
}

func TestFindSimilar(t *testing.T) {
	// Create a fake embedding (512 floats)
	createEmbedding := func(seed float32) []byte {
		floats := make([]float32, 512)
		for i := range floats {
			floats[i] = seed + float32(i)*0.001
		}
		return FloatsToBytes(floats)
	}

	query := &TrackFeatures{
		TrackID:         1,
		ContentHash:     "query-hash",
		Title:           "Query Track",
		Artist:          "Query Artist",
		BPM:             128.0,
		KeyValue:        "8A",
		Energy:          7,
		OpenL3Embedding: createEmbedding(1.0),
	}

	candidates := []*TrackFeatures{
		{
			TrackID:         2,
			ContentHash:     "similar-hash",
			Title:           "Similar Track",
			Artist:          "Similar Artist",
			BPM:             129.0,
			KeyValue:        "8A",
			Energy:          7,
			OpenL3Embedding: createEmbedding(1.0), // Same seed = similar
		},
		{
			TrackID:         3,
			ContentHash:     "different-hash",
			Title:           "Different Track",
			Artist:          "Different Artist",
			BPM:             140.0,
			KeyValue:        "2B",
			Energy:          3,
			OpenL3Embedding: createEmbedding(100.0), // Different seed = different
		},
	}

	results := FindSimilar(query, candidates, 10)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// First result should be the similar track (higher score)
	if results[0].TrackID != 2 {
		t.Errorf("expected similar track first, got track ID %d", results[0].TrackID)
	}

	if results[0].Score <= results[1].Score {
		t.Errorf("expected first result to have higher score: %v <= %v", results[0].Score, results[1].Score)
	}

	// Check explanation is populated
	if results[0].Explanation == "" {
		t.Error("expected explanation to be populated")
	}
}

func TestBytesFloatsRoundTrip(t *testing.T) {
	original := []float32{1.5, 2.5, 3.5, -4.5, 0.0}
	bytes := FloatsToBytes(original)
	recovered := bytesToFloats(bytes)

	if len(recovered) != len(original) {
		t.Fatalf("length mismatch: %d vs %d", len(recovered), len(original))
	}

	for i := range original {
		if recovered[i] != original[i] {
			t.Errorf("value mismatch at %d: %v vs %v", i, recovered[i], original[i])
		}
	}
}
