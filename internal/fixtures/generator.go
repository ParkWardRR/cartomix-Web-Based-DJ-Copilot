package fixtures

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

// Config controls which fixtures are emitted.
type Config struct {
	OutputDir    string
	SampleRate   int
	Seed         int64
	BPMLadder    []float64
	SwingRatio   float64 // e.g., 0.6 means offbeat delayed to 60% of beat duration
	IncludeSwing bool
	IncludeRamp  bool
	RampStartBPM float64
	RampEndBPM   float64
	IncludeChord bool
	ChordKey     string
}

// Manifest describes generated fixtures for tests/consumers.
type Manifest struct {
	SampleRate int               `json:"sample_rate"`
	Seed       int64             `json:"seed"`
	Fixtures   []ManifestFixture `json:"fixtures"`
}

type ManifestFixture struct {
	File        string  `json:"file"`
	Type        string  `json:"type"`
	BPM         float64 `json:"bpm,omitempty"`
	TargetBPM   float64 `json:"target_bpm,omitempty"`
	Beats       int     `json:"beats,omitempty"`
	DurationSec float64 `json:"duration_sec"`
	SwingRatio  float64 `json:"swing_ratio,omitempty"`
	Key         string  `json:"key,omitempty"`
}

// Generate writes WAV fixtures and a manifest.json into OutputDir.
func Generate(cfg Config) (*Manifest, error) {
	if cfg.SampleRate == 0 {
		cfg.SampleRate = 48000
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "./testdata/audio"
	}

	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir output: %w", err)
	}

	manifest := &Manifest{SampleRate: cfg.SampleRate, Seed: cfg.Seed}

	// 1) BPM ladder click tracks
	for _, bpm := range cfg.BPMLadder {
		filename := fmt.Sprintf("click_%dbpm.wav", int(bpm))
		path := filepath.Join(cfg.OutputDir, filename)
		durationSec := renderClickTrack(path, cfg.SampleRate, bpm, 32 /*beats*/, 0, 1.0)
		manifest.Fixtures = append(manifest.Fixtures, ManifestFixture{
			File:        filename,
			Type:        "click",
			BPM:         bpm,
			Beats:       32,
			DurationSec: durationSec,
		})
	}

	// 2) Swing click
	if cfg.IncludeSwing {
		bpm := cfg.BPMLadder[len(cfg.BPMLadder)/2]
		filename := "swing_click.wav"
		path := filepath.Join(cfg.OutputDir, filename)
		durationSec := renderClickTrack(path, cfg.SampleRate, bpm, 32, cfg.SwingRatio, 1.0)
		manifest.Fixtures = append(manifest.Fixtures, ManifestFixture{
			File:        filename,
			Type:        "swing_click",
			BPM:         bpm,
			SwingRatio:  cfg.SwingRatio,
			Beats:       32,
			DurationSec: durationSec,
		})
	}

	// 3) Tempo ramp
	if cfg.IncludeRamp {
		filename := "tempo_ramp.wav"
		path := filepath.Join(cfg.OutputDir, filename)
		durationSec := renderTempoRamp(path, cfg.SampleRate, cfg.RampStartBPM, cfg.RampEndBPM, 64)
		manifest.Fixtures = append(manifest.Fixtures, ManifestFixture{
			File:        filename,
			Type:        "tempo_ramp",
			BPM:         cfg.RampStartBPM,
			TargetBPM:   cfg.RampEndBPM,
			Beats:       64,
			DurationSec: durationSec,
		})
	}

	// 4) Harmonic chord pad (simple triad)
	if cfg.IncludeChord {
		filename := fmt.Sprintf("chord_%s.wav", cfg.ChordKey)
		path := filepath.Join(cfg.OutputDir, filename)
		durationSec := renderChord(path, cfg.SampleRate, cfg.ChordKey, 8.0)
		manifest.Fixtures = append(manifest.Fixtures, ManifestFixture{
			File:        filename,
			Type:        "harmonic_chord",
			Key:         cfg.ChordKey,
			DurationSec: durationSec,
		})
	}

	manifestPath := filepath.Join(cfg.OutputDir, "manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal manifest: %w", err)
	}
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		return nil, fmt.Errorf("write manifest: %w", err)
	}

	return manifest, nil
}

// renderClickTrack writes a mono WAV with short clicks per beat.
func renderClickTrack(path string, sampleRate int, bpm float64, beats int, swingRatio float64, amplitude float64) float64 {
	secondsPerBeat := 60.0 / bpm
	totalDuration := secondsPerBeat * float64(beats)
	samples := int(totalDuration * float64(sampleRate))
	data := make([]float64, samples)

	clickLen := int(0.01 * float64(sampleRate)) // 10ms click
	for i := 0; i < beats; i++ {
		offsetSec := secondsPerBeat * float64(i)
		// Swing applies to off-beats (odd beats)
		if swingRatio > 0 && i%2 == 1 {
			offsetSec = secondsPerBeat*float64(i-1) + secondsPerBeat*swingRatio
		}
		offset := int(offsetSec * float64(sampleRate))
		for j := 0; j < clickLen && offset+j < len(data); j++ {
			// simple exponential decay
			data[offset+j] += amplitude * math.Exp(-4*float64(j)/float64(clickLen))
		}
	}

	writeWAV(path, data, sampleRate)
	return totalDuration
}

// renderTempoRamp writes clicks whose interval ramps linearly from start to end BPM.
func renderTempoRamp(path string, sampleRate int, startBPM, endBPM float64, beats int) float64 {
	data := []float64{}
	currentTime := 0.0
	clickLen := int(0.01 * float64(sampleRate))

	for i := 0; i < beats; i++ {
		progress := float64(i) / float64(beats-1)
		bpm := startBPM + (endBPM-startBPM)*progress
		secondsPerBeat := 60.0 / bpm
		offset := int(currentTime * float64(sampleRate))

		ensure := offset + clickLen
		if ensure > len(data) {
			data = append(data, make([]float64, ensure-len(data))...)
		}

		for j := 0; j < clickLen; j++ {
			data[offset+j] += math.Exp(-4 * float64(j) / float64(clickLen))
		}

		currentTime += secondsPerBeat
	}

	writeWAV(path, data, sampleRate)
	return currentTime
}

// renderChord writes a simple triad pad for the given Camelot-like key (e.g., 8A).
func renderChord(path string, sampleRate int, key string, durationSec float64) float64 {
	freqs := camelotFrequencies(key)
	totalSamples := int(durationSec * float64(sampleRate))
	data := make([]float64, totalSamples)

	for _, f := range freqs {
		for i := 0; i < totalSamples; i++ {
			t := float64(i) / float64(sampleRate)
			data[i] += 0.2 * math.Sin(2*math.Pi*f*t) // low amplitude to avoid clipping when summed
		}
	}

	// Apply simple fade in/out (50ms)
	fadeSamples := int(0.05 * float64(sampleRate))
	for i := 0; i < fadeSamples; i++ {
		gain := float64(i) / float64(fadeSamples)
		data[i] *= gain
		data[totalSamples-1-i] *= gain
	}

	writeWAV(path, data, sampleRate)
	return durationSec
}

// writeWAV writes mono 16-bit PCM WAV.
func writeWAV(path string, samples []float64, sampleRate int) {
	// Clamp and convert to int16
	buf := make([]int16, len(samples))
	for i, s := range samples {
		if s > 1 {
			s = 1
		} else if s < -1 {
			s = -1
		}
		buf[i] = int16(s * 32767)
	}

	byteRate := sampleRate * 2
	blockAlign := int16(2)
	bitsPerSample := int16(16)
	dataSize := len(buf) * 2
	riffSize := 36 + dataSize

	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// RIFF header
	f.Write([]byte("RIFF"))
	binary.Write(f, binary.LittleEndian, uint32(riffSize))
	f.Write([]byte("WAVE"))
	// fmt chunk
	f.Write([]byte("fmt "))
	binary.Write(f, binary.LittleEndian, uint32(16))         // PCM header size
	binary.Write(f, binary.LittleEndian, uint16(1))          // PCM
	binary.Write(f, binary.LittleEndian, uint16(1))          // mono
	binary.Write(f, binary.LittleEndian, uint32(sampleRate)) // sample rate
	binary.Write(f, binary.LittleEndian, uint32(byteRate))   // byte rate
	binary.Write(f, binary.LittleEndian, blockAlign)         // block align
	binary.Write(f, binary.LittleEndian, bitsPerSample)      // bits per sample
	// data chunk
	f.Write([]byte("data"))
	binary.Write(f, binary.LittleEndian, uint32(dataSize))
	for _, v := range buf {
		binary.Write(f, binary.LittleEndian, v)
	}
}

// camelotFrequencies returns a simple triad mapped from Camelot notation (approximate).
func camelotFrequencies(key string) []float64 {
	// Map a handful of keys; defaults to A minor if unknown.
	switch key {
	case "8A": // A minor
		return []float64{220.0, 261.63, 329.63}
	case "9A": // E minor
		return []float64{164.81, 246.94, 329.63}
	case "7A": // D minor
		return []float64{146.83, 220.0, 293.66}
	case "8B": // C major
		return []float64{261.63, 329.63, 392.0}
	default:
		return []float64{220.0, 261.63, 329.63}
	}
}
