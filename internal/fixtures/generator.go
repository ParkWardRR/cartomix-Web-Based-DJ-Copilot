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
	// New fixture types
	IncludePhrase      bool     // Phrase track with intro/verse/build/drop/outro
	PhraseBPM          float64  // BPM for phrase track
	IncludeHarmonicSet bool     // Set of harmonically compatible tracks
	HarmonicSetKeys    []string // Keys for harmonic set (e.g., ["8A", "9A", "7A"])
	IncludeClubNoise   bool     // Club ambient noise fixture
}

// Manifest describes generated fixtures for tests/consumers.
type Manifest struct {
	SampleRate int               `json:"sample_rate"`
	Seed       int64             `json:"seed"`
	Fixtures   []ManifestFixture `json:"fixtures"`
}

type ManifestFixture struct {
	File        string             `json:"file"`
	Type        string             `json:"type"`
	BPM         float64            `json:"bpm,omitempty"`
	TargetBPM   float64            `json:"target_bpm,omitempty"`
	Beats       int                `json:"beats,omitempty"`
	DurationSec float64            `json:"duration_sec"`
	SwingRatio  float64            `json:"swing_ratio,omitempty"`
	Key         string             `json:"key,omitempty"`
	Sections    []ManifestSection  `json:"sections,omitempty"`    // For phrase tracks
	Energy      int                `json:"energy,omitempty"`      // 1-10 energy level
	SetID       string             `json:"set_id,omitempty"`      // Group ID for harmonic sets
	NoiseType   string             `json:"noise_type,omitempty"`  // For noise fixtures
}

// ManifestSection describes a section within a phrase track
type ManifestSection struct {
	Type       string  `json:"type"`       // intro, verse, build, drop, breakdown, outro
	StartBeat  int     `json:"start_beat"`
	EndBeat    int     `json:"end_beat"`
	StartTime  float64 `json:"start_time"`
	EndTime    float64 `json:"end_time"`
	Energy     int     `json:"energy"`     // 1-10
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

	// 5) Phrase track with sections
	if cfg.IncludePhrase {
		bpm := cfg.PhraseBPM
		if bpm == 0 {
			bpm = 128
		}
		filename := "phrase_track.wav"
		path := filepath.Join(cfg.OutputDir, filename)
		durationSec, sections := renderPhraseTrack(path, cfg.SampleRate, bpm)
		manifest.Fixtures = append(manifest.Fixtures, ManifestFixture{
			File:        filename,
			Type:        "phrase_track",
			BPM:         bpm,
			Key:         "8A",
			DurationSec: durationSec,
			Sections:    sections,
			Energy:      7,
		})
	}

	// 6) Harmonic set (multiple compatible tracks)
	if cfg.IncludeHarmonicSet {
		keys := cfg.HarmonicSetKeys
		if len(keys) == 0 {
			keys = []string{"8A", "9A", "7A", "8B"} // Camelot wheel neighbors
		}
		setID := fmt.Sprintf("harmonic_set_%d", cfg.Seed)
		bpms := []float64{126, 128, 130, 124} // Slightly varying BPMs
		for i, key := range keys {
			filename := fmt.Sprintf("harmonic_set_%d_%s.wav", i+1, key)
			path := filepath.Join(cfg.OutputDir, filename)
			bpm := bpms[i%len(bpms)]
			durationSec, sections := renderHarmonicSetTrack(path, cfg.SampleRate, key, bpm, int64(i))
			manifest.Fixtures = append(manifest.Fixtures, ManifestFixture{
				File:        filename,
				Type:        "harmonic_set_track",
				BPM:         bpm,
				Key:         key,
				DurationSec: durationSec,
				Sections:    sections,
				Energy:      5 + (i % 5),
				SetID:       setID,
			})
		}
	}

	// 7) Club noise fixture
	if cfg.IncludeClubNoise {
		noiseTypes := []string{"crowd", "reverb_tail", "pink_noise"}
		for _, noiseType := range noiseTypes {
			filename := fmt.Sprintf("club_noise_%s.wav", noiseType)
			path := filepath.Join(cfg.OutputDir, filename)
			durationSec := renderClubNoise(path, cfg.SampleRate, noiseType, cfg.Seed)
			manifest.Fixtures = append(manifest.Fixtures, ManifestFixture{
				File:        filename,
				Type:        "club_noise",
				NoiseType:   noiseType,
				DurationSec: durationSec,
			})
		}
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
	case "9B": // G major
		return []float64{196.0, 246.94, 293.66}
	case "7B": // F major
		return []float64{174.61, 220.0, 261.63}
	default:
		return []float64{220.0, 261.63, 329.63}
	}
}

// renderPhraseTrack creates a track with DJ-style phrase structure
// Structure: 16-bar intro, 32-bar verse, 16-bar build, 32-bar drop, 16-bar breakdown, 16-bar outro
func renderPhraseTrack(path string, sampleRate int, bpm float64) (float64, []ManifestSection) {
	secondsPerBeat := 60.0 / bpm
	beatsPerBar := 4

	// Define sections (beats)
	sectionDefs := []struct {
		typ    string
		bars   int
		energy int
	}{
		{"intro", 16, 3},
		{"verse", 32, 5},
		{"build", 16, 7},
		{"drop", 32, 10},
		{"breakdown", 16, 4},
		{"outro", 16, 2},
	}

	totalBeats := 0
	sections := []ManifestSection{}
	for _, def := range sectionDefs {
		beats := def.bars * beatsPerBar
		startBeat := totalBeats
		endBeat := totalBeats + beats
		sections = append(sections, ManifestSection{
			Type:      def.typ,
			StartBeat: startBeat,
			EndBeat:   endBeat,
			StartTime: float64(startBeat) * secondsPerBeat,
			EndTime:   float64(endBeat) * secondsPerBeat,
			Energy:    def.energy,
		})
		totalBeats = endBeat
	}

	totalDuration := float64(totalBeats) * secondsPerBeat
	totalSamples := int(totalDuration * float64(sampleRate))
	data := make([]float64, totalSamples)

	// Base frequencies for A minor
	bassFreq := 110.0  // A2
	leadFreq := 440.0  // A4
	padFreqs := []float64{220.0, 261.63, 329.63} // A minor chord

	// Render each section with appropriate energy
	for _, section := range sections {
		startSample := int(section.StartTime * float64(sampleRate))
		endSample := int(section.EndTime * float64(sampleRate))
		energy := float64(section.Energy) / 10.0

		// Add kick on downbeats
		for beat := section.StartBeat; beat < section.EndBeat; beat++ {
			beatTime := float64(beat) * secondsPerBeat
			beatSample := int(beatTime * float64(sampleRate))

			// Kick drum (simple sine decay)
			if beat%beatsPerBar == 0 || (section.Type == "drop" && beat%2 == 0) {
				kickLen := int(0.15 * float64(sampleRate))
				for i := 0; i < kickLen && beatSample+i < totalSamples; i++ {
					t := float64(i) / float64(sampleRate)
					kickFreq := 60.0 * math.Exp(-15*t) // Pitch envelope
					amplitude := energy * 0.7 * math.Exp(-10*t)
					data[beatSample+i] += amplitude * math.Sin(2*math.Pi*kickFreq*t)
				}
			}

			// Hi-hat on off-beats
			if (beat%2 == 1 || section.Type == "drop") && energy > 0.3 {
				hatLen := int(0.02 * float64(sampleRate))
				for i := 0; i < hatLen && beatSample+i < totalSamples; i++ {
					t := float64(i) / float64(sampleRate)
					noise := (float64(uint32(beat*1337+i)%65536)/32768.0 - 1.0) // Deterministic noise
					amplitude := energy * 0.15 * math.Exp(-30*t)
					data[beatSample+i] += amplitude * noise
				}
			}
		}

		// Add bass line in verse, build, drop
		if section.Type == "verse" || section.Type == "build" || section.Type == "drop" {
			for i := startSample; i < endSample; i++ {
				t := float64(i) / float64(sampleRate)
				beatPos := t / secondsPerBeat
				barPos := beatPos / float64(beatsPerBar)
				// Simple bass pattern
				bassAmp := energy * 0.3 * (0.5 + 0.5*math.Sin(2*math.Pi*barPos))
				data[i] += bassAmp * math.Sin(2*math.Pi*bassFreq*t)
			}
		}

		// Add lead in build and drop
		if section.Type == "build" || section.Type == "drop" {
			for i := startSample; i < endSample; i++ {
				t := float64(i) / float64(sampleRate)
				leadAmp := energy * 0.2
				data[i] += leadAmp * math.Sin(2*math.Pi*leadFreq*t)
			}
		}

		// Add pad throughout
		for _, freq := range padFreqs {
			for i := startSample; i < endSample; i++ {
				t := float64(i) / float64(sampleRate)
				padAmp := energy * 0.1
				data[i] += padAmp * math.Sin(2*math.Pi*freq*t)
			}
		}
	}

	// Apply global fade in/out
	fadeSamples := int(0.5 * float64(sampleRate))
	for i := 0; i < fadeSamples; i++ {
		gain := float64(i) / float64(fadeSamples)
		data[i] *= gain
		data[totalSamples-1-i] *= gain
	}

	writeWAV(path, data, sampleRate)
	return totalDuration, sections
}

// renderHarmonicSetTrack creates a shorter track for harmonic set testing
func renderHarmonicSetTrack(path string, sampleRate int, key string, bpm float64, trackIndex int64) (float64, []ManifestSection) {
	secondsPerBeat := 60.0 / bpm
	beatsPerBar := 4

	// Shorter structure: 8-bar intro, 16-bar main, 8-bar outro
	totalBars := 32
	totalBeats := totalBars * beatsPerBar
	totalDuration := float64(totalBeats) * secondsPerBeat
	totalSamples := int(totalDuration * float64(sampleRate))
	data := make([]float64, totalSamples)

	sections := []ManifestSection{
		{Type: "intro", StartBeat: 0, EndBeat: 32, StartTime: 0, EndTime: float64(32) * secondsPerBeat, Energy: 4},
		{Type: "main", StartBeat: 32, EndBeat: 96, StartTime: float64(32) * secondsPerBeat, EndTime: float64(96) * secondsPerBeat, Energy: 7},
		{Type: "outro", StartBeat: 96, EndBeat: 128, StartTime: float64(96) * secondsPerBeat, EndTime: float64(128) * secondsPerBeat, Energy: 3},
	}

	freqs := camelotFrequencies(key)
	bassFreq := freqs[0] / 2 // One octave down

	// Render with slight variations per track
	variation := float64(trackIndex%5) * 0.1

	for _, section := range sections {
		startSample := int(section.StartTime * float64(sampleRate))
		endSample := int(section.EndTime * float64(sampleRate))
		energy := float64(section.Energy) / 10.0

		for beat := section.StartBeat; beat < section.EndBeat; beat++ {
			beatTime := float64(beat) * secondsPerBeat
			beatSample := int(beatTime * float64(sampleRate))

			// Kick on downbeats
			if beat%beatsPerBar == 0 {
				kickLen := int(0.12 * float64(sampleRate))
				for i := 0; i < kickLen && beatSample+i < totalSamples; i++ {
					t := float64(i) / float64(sampleRate)
					kickFreq := 55.0 * math.Exp(-12*t)
					amplitude := (energy + variation) * 0.6 * math.Exp(-8*t)
					data[beatSample+i] += amplitude * math.Sin(2*math.Pi*kickFreq*t)
				}
			}
		}

		// Bass and chords
		for i := startSample; i < endSample; i++ {
			t := float64(i) / float64(sampleRate)
			bassAmp := (energy + variation) * 0.25
			data[i] += bassAmp * math.Sin(2*math.Pi*bassFreq*t)

			for j, freq := range freqs {
				padAmp := (energy + variation) * 0.08 * (1 - float64(j)*0.2)
				data[i] += padAmp * math.Sin(2*math.Pi*freq*t)
			}
		}
	}

	// Fade
	fadeSamples := int(0.3 * float64(sampleRate))
	for i := 0; i < fadeSamples && i < totalSamples; i++ {
		gain := float64(i) / float64(fadeSamples)
		data[i] *= gain
		if totalSamples-1-i >= 0 {
			data[totalSamples-1-i] *= gain
		}
	}

	writeWAV(path, data, sampleRate)
	return totalDuration, sections
}

// renderClubNoise creates ambient noise fixtures for testing noise detection
func renderClubNoise(path string, sampleRate int, noiseType string, seed int64) float64 {
	durationSec := 10.0
	totalSamples := int(durationSec * float64(sampleRate))
	data := make([]float64, totalSamples)

	// Simple LCG for deterministic pseudo-random
	rng := uint64(seed)
	nextRand := func() float64 {
		rng = rng*6364136223846793005 + 1442695040888963407
		return float64(rng>>33) / float64(1<<31) // 0 to 1
	}

	switch noiseType {
	case "crowd":
		// Crowd noise: low-passed noise with slow modulation
		var lowpass float64
		for i := 0; i < totalSamples; i++ {
			t := float64(i) / float64(sampleRate)
			// Multiple "voices" with slow envelope
			noise := nextRand()*2 - 1
			// Low pass filter
			lowpass = lowpass*0.98 + noise*0.02
			// Slow amplitude modulation (crowd swell)
			mod := 0.5 + 0.3*math.Sin(2*math.Pi*0.2*t) + 0.2*math.Sin(2*math.Pi*0.07*t)
			data[i] = lowpass * mod * 0.4
		}

	case "reverb_tail":
		// Reverb tail: decaying filtered noise
		var lowpass float64
		for i := 0; i < totalSamples; i++ {
			t := float64(i) / float64(sampleRate)
			noise := nextRand()*2 - 1
			// Longer decay at start
			decay := math.Exp(-0.5 * t)
			// Increase filtering over time
			filterCoef := 0.9 + 0.09*t/durationSec
			lowpass = lowpass*filterCoef + noise*(1-filterCoef)
			data[i] = lowpass * decay * 0.5
		}

	case "pink_noise":
		// Pink noise (1/f): use multiple filtered noise sources
		var b [7]float64
		for i := 0; i < totalSamples; i++ {
			white := nextRand()*2 - 1
			// Paul Kellet's approximation
			b[0] = 0.99886*b[0] + white*0.0555179
			b[1] = 0.99332*b[1] + white*0.0750759
			b[2] = 0.96900*b[2] + white*0.1538520
			b[3] = 0.86650*b[3] + white*0.3104856
			b[4] = 0.55000*b[4] + white*0.5329522
			b[5] = -0.7616*b[5] - white*0.0168980
			pink := b[0] + b[1] + b[2] + b[3] + b[4] + b[5] + b[6] + white*0.5362
			b[6] = white * 0.115926
			data[i] = pink * 0.11 // Scale to reasonable level
		}
	}

	// Fade in/out
	fadeSamples := int(0.2 * float64(sampleRate))
	for i := 0; i < fadeSamples; i++ {
		gain := float64(i) / float64(fadeSamples)
		data[i] *= gain
		data[totalSamples-1-i] *= gain
	}

	writeWAV(path, data, sampleRate)
	return durationSec
}
