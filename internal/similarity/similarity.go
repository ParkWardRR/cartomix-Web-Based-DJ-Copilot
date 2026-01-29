// Package similarity provides ML-powered track similarity search using OpenL3 embeddings.
package similarity

import (
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"strings"
)

// EmbeddingDim is the dimensionality of OpenL3 embeddings.
const EmbeddingDim = 512

// Weights for combined similarity score.
const (
	WeightOpenL3  = 0.50 // OpenL3 embedding similarity
	WeightTempo   = 0.20 // BPM compatibility
	WeightKey     = 0.20 // Key compatibility
	WeightEnergy  = 0.10 // Energy level similarity
)

// TrackFeatures contains the features needed for similarity computation.
type TrackFeatures struct {
	TrackID         int64
	ContentHash     string
	Title           string
	Artist          string
	BPM             float64
	KeyValue        string  // Camelot notation (e.g., "8A", "12B")
	Energy          int32   // 1-10 scale
	OpenL3Embedding []byte  // 512 x float32 = 2048 bytes
}

// SimilarityResult represents a similarity match with explanation.
type SimilarityResult struct {
	TrackID      int64   `json:"track_id"`
	ContentHash  string  `json:"content_hash"`
	Title        string  `json:"title"`
	Artist       string  `json:"artist"`
	Score        float64 `json:"score"`         // Combined score 0-1
	Explanation  string  `json:"explanation"`   // Human-readable rationale
	VibeMatch    float64 `json:"vibe_match"`    // OpenL3 similarity %
	TempoMatch   float64 `json:"tempo_match"`   // BPM similarity %
	KeyMatch     float64 `json:"key_match"`     // Key compatibility %
	EnergyMatch  float64 `json:"energy_match"`  // Energy similarity %
	BPMDelta     float64 `json:"bpm_delta"`     // Absolute BPM difference
	KeyRelation  string  `json:"key_relation"`  // "same", "compatible", "harmonic", "clash"
	EnergyDelta  int32   `json:"energy_delta"`  // Signed energy difference
}

// TransitionMatch represents a potential mix transition point.
type TransitionMatch struct {
	TrackID        int64   `json:"track_id"`
	ContentHash    string  `json:"content_hash"`
	Title          string  `json:"title"`
	Artist         string  `json:"artist"`
	Score          float64 `json:"score"`
	Explanation    string  `json:"explanation"`
	SuggestedOutBeat int32 `json:"suggested_out_beat"` // From current track
	SuggestedInBeat  int32 `json:"suggested_in_beat"`  // Into target track
}

// FindSimilar finds tracks similar to the query track.
func FindSimilar(query *TrackFeatures, candidates []*TrackFeatures, limit int) []SimilarityResult {
	if query == nil || len(candidates) == 0 {
		return nil
	}

	queryEmb := bytesToFloats(query.OpenL3Embedding)

	results := make([]SimilarityResult, 0, len(candidates))

	for _, candidate := range candidates {
		if candidate.TrackID == query.TrackID {
			continue // Skip self
		}

		// Compute component similarities
		vibeMatch := computeCosineSimilarity(queryEmb, bytesToFloats(candidate.OpenL3Embedding))
		tempoMatch := computeTempoSimilarity(query.BPM, candidate.BPM)
		keyMatch, keyRelation := computeKeySimilarity(query.KeyValue, candidate.KeyValue)
		energyMatch := computeEnergySimilarity(query.Energy, candidate.Energy)

		// Combined weighted score
		score := WeightOpenL3*vibeMatch + WeightTempo*tempoMatch + WeightKey*keyMatch + WeightEnergy*energyMatch

		// Build explanation
		explanation := buildExplanation(vibeMatch, tempoMatch, keyMatch, keyRelation, energyMatch, query.BPM, candidate.BPM, query.Energy, candidate.Energy)

		results = append(results, SimilarityResult{
			TrackID:     candidate.TrackID,
			ContentHash: candidate.ContentHash,
			Title:       candidate.Title,
			Artist:      candidate.Artist,
			Score:       score,
			Explanation: explanation,
			VibeMatch:   vibeMatch * 100,
			TempoMatch:  tempoMatch * 100,
			KeyMatch:    keyMatch * 100,
			EnergyMatch: energyMatch * 100,
			BPMDelta:    math.Abs(query.BPM - candidate.BPM),
			KeyRelation: keyRelation,
			EnergyDelta: candidate.Energy - query.Energy,
		})
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit results
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results
}

// computeCosineSimilarity calculates cosine similarity between two embedding vectors.
func computeCosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	denominator := math.Sqrt(normA) * math.Sqrt(normB)
	if denominator == 0 {
		return 0
	}

	// Cosine similarity is -1 to 1, normalize to 0-1
	sim := dotProduct / denominator
	return (sim + 1) / 2
}

// computeTempoSimilarity returns similarity based on BPM difference.
// Perfect match = 1.0, >10 BPM difference = 0.0
func computeTempoSimilarity(bpmA, bpmB float64) float64 {
	if bpmA == 0 || bpmB == 0 {
		return 0.5 // Unknown tempo, neutral score
	}

	diff := math.Abs(bpmA - bpmB)

	// Also check half/double tempo compatibility
	halfDiff := math.Min(
		math.Abs(bpmA - bpmB*2),
		math.Abs(bpmA*2 - bpmB),
	)
	diff = math.Min(diff, halfDiff)

	if diff <= 1 {
		return 1.0
	} else if diff >= 10 {
		return 0.0
	}
	return 1.0 - (diff / 10.0)
}

// computeKeySimilarity returns key compatibility score and relation type.
func computeKeySimilarity(keyA, keyB string) (float64, string) {
	if keyA == "" || keyB == "" {
		return 0.5, "unknown"
	}

	keyA = strings.ToUpper(strings.TrimSpace(keyA))
	keyB = strings.ToUpper(strings.TrimSpace(keyB))

	if keyA == keyB {
		return 1.0, "same"
	}

	// Parse Camelot notation (e.g., "8A", "12B")
	numA, modeA := parseCamelot(keyA)
	numB, modeB := parseCamelot(keyB)

	if numA == 0 || numB == 0 {
		return 0.3, "unknown"
	}

	// Same number, different mode (relative major/minor)
	if numA == numB && modeA != modeB {
		return 0.9, "relative"
	}

	// Adjacent numbers, same mode (+1/-1)
	if modeA == modeB {
		diff := (numA - numB + 12) % 12
		if diff == 1 || diff == 11 {
			return 0.85, "compatible"
		}
	}

	// Energy boost (+1 semitone)
	if modeA == modeB {
		diff := (numA - numB + 12) % 12
		if diff == 2 || diff == 10 {
			return 0.7, "harmonic"
		}
	}

	// Cross-mode adjacent
	if modeA != modeB {
		diff := (numA - numB + 12) % 12
		if diff == 0 || diff == 1 || diff == 11 {
			return 0.75, "harmonic"
		}
	}

	return 0.2, "clash"
}

// parseCamelot extracts number and mode from Camelot notation.
func parseCamelot(key string) (int, string) {
	if len(key) < 2 {
		return 0, ""
	}

	mode := string(key[len(key)-1])
	numStr := key[:len(key)-1]

	var num int
	fmt.Sscanf(numStr, "%d", &num)

	if num < 1 || num > 12 {
		return 0, ""
	}

	if mode != "A" && mode != "B" {
		return 0, ""
	}

	return num, mode
}

// computeEnergySimilarity returns similarity based on energy level difference.
func computeEnergySimilarity(energyA, energyB int32) float64 {
	diff := math.Abs(float64(energyA - energyB))
	if diff == 0 {
		return 1.0
	} else if diff >= 5 {
		return 0.0
	}
	return 1.0 - (diff / 5.0)
}

// buildExplanation creates a human-readable explanation string.
func buildExplanation(vibeMatch, tempoMatch, keyMatch float64, keyRelation string, energyMatch float64, bpmA, bpmB float64, energyA, energyB int32) string {
	var parts []string

	// Vibe match
	if vibeMatch >= 0.7 {
		parts = append(parts, fmt.Sprintf("similar vibe (%.0f%%)", vibeMatch*100))
	}

	// Tempo
	bpmDiff := bpmB - bpmA
	if math.Abs(bpmDiff) <= 2 {
		parts = append(parts, "tempo match")
	} else if bpmDiff > 0 {
		parts = append(parts, fmt.Sprintf("Δ+%.1f BPM", bpmDiff))
	} else {
		parts = append(parts, fmt.Sprintf("Δ%.1f BPM", bpmDiff))
	}

	// Key
	switch keyRelation {
	case "same":
		parts = append(parts, "same key")
	case "relative":
		parts = append(parts, "relative key")
	case "compatible":
		parts = append(parts, "key compatible")
	case "harmonic":
		parts = append(parts, "harmonic key")
	case "clash":
		parts = append(parts, "key clash ⚠")
	}

	// Energy
	energyDiff := energyB - energyA
	if energyDiff == 0 {
		parts = append(parts, "same energy")
	} else if energyDiff > 0 {
		parts = append(parts, fmt.Sprintf("energy +%d", energyDiff))
	} else {
		parts = append(parts, fmt.Sprintf("energy %d", energyDiff))
	}

	return strings.Join(parts, "; ")
}

// bytesToFloats converts a byte slice to float32 slice (little-endian).
func bytesToFloats(data []byte) []float32 {
	if len(data) == 0 {
		return nil
	}

	floatCount := len(data) / 4
	result := make([]float32, floatCount)

	for i := 0; i < floatCount; i++ {
		bits := binary.LittleEndian.Uint32(data[i*4 : (i+1)*4])
		result[i] = math.Float32frombits(bits)
	}

	return result
}

// FloatsToBytes converts a float32 slice to byte slice (little-endian).
func FloatsToBytes(floats []float32) []byte {
	data := make([]byte, len(floats)*4)
	for i, f := range floats {
		binary.LittleEndian.PutUint32(data[i*4:], math.Float32bits(f))
	}
	return data
}
