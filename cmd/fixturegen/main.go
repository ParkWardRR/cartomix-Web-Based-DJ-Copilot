package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/cartomix/cancun/internal/fixtures"
)

// fixturegen produces deterministic WAV fixtures used by tests and demos.
func main() {
	outDir := flag.String("out", "./testdata/audio", "output directory for generated audio")
	seed := flag.Int("seed", 1337, "random seed for deterministic fixtures")
	bpmLadderStr := flag.String("bpm-ladder", "80,100,120,128,140,160", "comma-separated BPM ladder")
	includeSwing := flag.Bool("include-swing", true, "include swing/shuffle fixtures")
	includeTempoRamp := flag.Bool("include-tempo-ramp", true, "include dynamic tempo fixtures")
	rampStart := flag.Float64("ramp-start-bpm", 128, "tempo ramp start BPM")
	rampEnd := flag.Float64("ramp-end-bpm", 100, "tempo ramp end BPM")
	includeHarmonic := flag.String("include-harmonic-keys", "8A", "comma-separated keys for harmonic fixtures")

	// New fixture types
	includePhrase := flag.Bool("include-phrase", true, "include phrase track with sections")
	phraseBPM := flag.Float64("phrase-bpm", 128, "BPM for phrase track")
	includeHarmonicSet := flag.Bool("include-harmonic-set", true, "include harmonic set of tracks")
	harmonicSetKeys := flag.String("harmonic-set-keys", "8A,9A,7A,8B", "comma-separated keys for harmonic set")
	includeClubNoise := flag.Bool("include-club-noise", true, "include club ambient noise fixtures")

	flag.Parse()

	var ladder []float64
	for _, s := range strings.Split(*bpmLadderStr, ",") {
		var v float64
		_, err := fmt.Sscanf(strings.TrimSpace(s), "%f", &v)
		if err == nil {
			ladder = append(ladder, v)
		}
	}
	if len(ladder) == 0 {
		ladder = []float64{120}
	}

	keys := strings.Split(*includeHarmonic, ",")
	includeChord := len(keys) > 0 && keys[0] != ""

	// Parse harmonic set keys
	var harmonicKeys []string
	for _, k := range strings.Split(*harmonicSetKeys, ",") {
		k = strings.TrimSpace(k)
		if k != "" {
			harmonicKeys = append(harmonicKeys, k)
		}
	}

	cfg := fixtures.Config{
		OutputDir:          *outDir,
		SampleRate:         48000,
		Seed:               int64(*seed),
		BPMLadder:          ladder,
		SwingRatio:         0.6,
		IncludeSwing:       *includeSwing,
		IncludeRamp:        *includeTempoRamp,
		RampStartBPM:       *rampStart,
		RampEndBPM:         *rampEnd,
		IncludeChord:       includeChord,
		IncludePhrase:      *includePhrase,
		PhraseBPM:          *phraseBPM,
		IncludeHarmonicSet: *includeHarmonicSet,
		HarmonicSetKeys:    harmonicKeys,
		IncludeClubNoise:   *includeClubNoise,
	}
	if includeChord {
		cfg.ChordKey = strings.TrimSpace(keys[0])
		if cfg.ChordKey == "" {
			cfg.ChordKey = "8A"
		}
	}

	manifest, err := fixtures.Generate(cfg)
	if err != nil {
		log.Fatalf("generate fixtures: %v", err)
	}

	log.Printf("fixturegen wrote %d fixtures to %s (sample_rate=%d)", len(manifest.Fixtures), cfg.OutputDir, cfg.SampleRate)
}
