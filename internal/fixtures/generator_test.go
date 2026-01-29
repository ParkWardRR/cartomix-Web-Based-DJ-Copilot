package fixtures

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateProducesAudioAndManifest(t *testing.T) {
	dir := t.TempDir()

	cfg := Config{
		OutputDir:    dir,
		SampleRate:   48000,
		BPMLadder:    []float64{120, 128},
		SwingRatio:   0.6,
		IncludeSwing: true,
		IncludeRamp:  true,
		RampStartBPM: 128,
		RampEndBPM:   100,
		IncludeChord: true,
		ChordKey:     "8A",
	}

	manifest, err := Generate(cfg)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if len(manifest.Fixtures) < 4 {
		t.Fatalf("expected at least 4 fixtures, got %d", len(manifest.Fixtures))
	}

	// Validate one WAV
	wavPath := filepath.Join(dir, "click_120bpm.wav")
	if _, err := os.Stat(wavPath); err != nil {
		t.Fatalf("wav missing: %v", err)
	}

	data, err := os.ReadFile(wavPath)
	if err != nil {
		t.Fatalf("read wav: %v", err)
	}

	if string(data[:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		t.Fatalf("not a wav header")
	}

	// sample rate at byte offset 24
	sampleRate := binary.LittleEndian.Uint32(data[24:28])
	if sampleRate != uint32(cfg.SampleRate) {
		t.Fatalf("unexpected sample rate %d", sampleRate)
	}
}

