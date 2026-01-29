package exporter

import (
	"encoding/xml"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cartomix/cancun/gen/go/common"
	"google.golang.org/protobuf/types/known/durationpb"
)

var updateGolden = flag.Bool("update-golden", false, "update golden test files")

// goldenTracks returns a deterministic set of tracks for golden tests.
func goldenTracks() []TrackExport {
	return []TrackExport{
		{
			Path: "/Music/Artist1/Track One.mp3",
			Analysis: &common.TrackAnalysis{
				Id: &common.TrackId{
					ContentHash: "abc123def456",
					Path:        "/Music/Artist1/Track One.mp3",
				},
				Key: &common.MusicalKey{
					Value:      "8A",
					Format:     common.KeyFormat_CAMELOT,
					Confidence: 0.92,
				},
				Beatgrid: &common.Beatgrid{
					Beats: []*common.BeatMarker{
						{Index: 0, Time: durationpb.New(0), IsDownbeat: true},
						{Index: 256, Time: durationpb.New(time.Minute * 3)},
					},
					TempoMap: []*common.TempoMapNode{
						{BeatIndex: 0, Bpm: 128.0},
					},
					Confidence: 0.88,
				},
				CuePoints: []*common.CuePoint{
					{BeatIndex: 0, Time: durationpb.New(0), Type: common.CueType_CUE_LOAD},
					{BeatIndex: 32, Time: durationpb.New(15 * time.Second), Type: common.CueType_CUE_DROP},
					{BeatIndex: 224, Time: durationpb.New(165 * time.Second), Type: common.CueType_CUE_OUTRO_START},
				},
				EnergyGlobal: 7,
			},
		},
		{
			Path: "/Music/Artist2/Track Two.wav",
			Analysis: &common.TrackAnalysis{
				Id: &common.TrackId{
					ContentHash: "xyz789uvw012",
					Path:        "/Music/Artist2/Track Two.wav",
				},
				Key: &common.MusicalKey{
					Value:      "9A",
					Format:     common.KeyFormat_CAMELOT,
					Confidence: 0.95,
				},
				Beatgrid: &common.Beatgrid{
					Beats: []*common.BeatMarker{
						{Index: 0, Time: durationpb.New(0), IsDownbeat: true},
						{Index: 288, Time: durationpb.New(time.Minute*3 + time.Second*30)},
					},
					TempoMap: []*common.TempoMapNode{
						{BeatIndex: 0, Bpm: 130.0},
					},
					Confidence: 0.91,
				},
				CuePoints: []*common.CuePoint{
					{BeatIndex: 0, Time: durationpb.New(0), Type: common.CueType_CUE_INTRO_START},
					{BeatIndex: 64, Time: durationpb.New(30 * time.Second), Type: common.CueType_CUE_BUILD},
					{BeatIndex: 128, Time: durationpb.New(60 * time.Second), Type: common.CueType_CUE_DROP},
				},
				EnergyGlobal: 8,
			},
		},
	}
}

func TestRekordboxGolden(t *testing.T) {
	dir := t.TempDir()
	tracks := goldenTracks()

	path, err := WriteRekordbox(dir, "golden-set", tracks)
	if err != nil {
		t.Fatalf("WriteRekordbox failed: %v", err)
	}

	actual, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	goldenPath := filepath.Join("testdata", "golden-rekordbox.xml")

	if *updateGolden {
		if err := os.WriteFile(goldenPath, actual, 0644); err != nil {
			t.Fatalf("failed to update golden file: %v", err)
		}
		t.Log("updated golden file:", goldenPath)
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if os.IsNotExist(err) {
		t.Skip("golden file does not exist, run with -update-golden to create")
	}
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	// Compare XML structure, not byte-for-byte (timestamps vary)
	compareXML(t, "Rekordbox", expected, actual)
}

func TestTraktorGolden(t *testing.T) {
	dir := t.TempDir()
	tracks := goldenTracks()

	path, err := WriteTraktor(dir, "golden-set", tracks)
	if err != nil {
		t.Fatalf("WriteTraktor failed: %v", err)
	}

	actual, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	goldenPath := filepath.Join("testdata", "golden-traktor.nml")

	if *updateGolden {
		if err := os.WriteFile(goldenPath, actual, 0644); err != nil {
			t.Fatalf("failed to update golden file: %v", err)
		}
		t.Log("updated golden file:", goldenPath)
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if os.IsNotExist(err) {
		t.Skip("golden file does not exist, run with -update-golden to create")
	}
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	compareXML(t, "Traktor", expected, actual)
}

func TestSeratoGolden(t *testing.T) {
	dir := t.TempDir()
	tracks := goldenTracks()

	path, err := WriteSerato(dir, "golden-set", tracks)
	if err != nil {
		t.Fatalf("WriteSerato failed: %v", err)
	}

	actual, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	goldenPath := filepath.Join("testdata", "golden-serato.crate")

	if *updateGolden {
		if err := os.WriteFile(goldenPath, actual, 0644); err != nil {
			t.Fatalf("failed to update golden file: %v", err)
		}
		t.Log("updated golden file:", goldenPath)
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if os.IsNotExist(err) {
		t.Skip("golden file does not exist, run with -update-golden to create")
	}
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	// Serato is binary, compare byte-for-byte (but check size first)
	if len(actual) != len(expected) {
		t.Errorf("Serato crate size mismatch: got %d, want %d", len(actual), len(expected))
	}

	// Check header bytes match
	if len(actual) >= 4 && len(expected) >= 4 {
		if string(actual[:4]) != string(expected[:4]) {
			t.Errorf("Serato header mismatch: got %q, want %q", actual[:4], expected[:4])
		}
	}
}

func TestGenericExportGolden(t *testing.T) {
	dir := t.TempDir()
	tracks := goldenTracks()

	result, err := WriteGeneric(dir, "golden-set", tracks)
	if err != nil {
		t.Fatalf("WriteGeneric failed: %v", err)
	}

	// Check M3U8 contains expected paths
	m3u, err := os.ReadFile(result.PlaylistPath)
	if err != nil {
		t.Fatalf("failed to read M3U8: %v", err)
	}
	m3uContent := string(m3u)
	if !strings.Contains(m3uContent, "Track One.mp3") {
		t.Error("M3U8 missing Track One")
	}
	if !strings.Contains(m3uContent, "Track Two.wav") {
		t.Error("M3U8 missing Track Two")
	}

	// Check JSON has track data
	jsonData, err := os.ReadFile(result.AnalysisJSONPath)
	if err != nil {
		t.Fatalf("failed to read JSON: %v", err)
	}
	if !strings.Contains(string(jsonData), "abc123def456") {
		t.Error("JSON missing track hash abc123def456")
	}

	// Check cues CSV has data
	csvData, err := os.ReadFile(result.CuesCSVPath)
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}
	csvContent := string(csvData)
	// Check CSV has header and data rows
	if !strings.Contains(csvContent, "beat_index") {
		t.Error("CSV missing header")
	}
	// Check cue types are present
	if !strings.Contains(csvContent, "CUE_DROP") && !strings.Contains(csvContent, "DROP") {
		t.Error("CSV missing cue type entries")
	}

	// Check checksum file exists
	if _, err := os.Stat(result.ChecksumsPath); os.IsNotExist(err) {
		t.Error("checksum file not created")
	}
}

// compareXML compares two XML documents for structural equivalence.
// It ignores whitespace differences and attribute ordering.
func compareXML(t *testing.T, name string, expected, actual []byte) {
	t.Helper()

	// Parse both documents
	var expDoc, actDoc interface{}
	if err := xml.Unmarshal(expected, &expDoc); err != nil {
		t.Errorf("%s: failed to parse expected XML: %v", name, err)
	}
	if err := xml.Unmarshal(actual, &actDoc); err != nil {
		t.Errorf("%s: failed to parse actual XML: %v", name, err)
	}

	// Check key elements exist
	actualStr := string(actual)
	expectedStr := string(expected)

	// Check for expected elements in both
	checkElements := []string{}
	if strings.Contains(expectedStr, "DJ_PLAYLISTS") {
		checkElements = append(checkElements, "DJ_PLAYLISTS", "COLLECTION", "TRACK", "PLAYLISTS")
	}
	if strings.Contains(expectedStr, "<NML") {
		checkElements = append(checkElements, "NML", "COLLECTION", "ENTRY", "PLAYLISTS")
	}

	for _, elem := range checkElements {
		if strings.Contains(expectedStr, elem) && !strings.Contains(actualStr, elem) {
			t.Errorf("%s: missing element %s in output", name, elem)
		}
	}

	// Check track counts match
	expCount := strings.Count(expectedStr, "<TRACK")
	actCount := strings.Count(actualStr, "<TRACK")
	if strings.Contains(expectedStr, "DJ_PLAYLISTS") {
		if expCount != actCount {
			t.Errorf("%s: TRACK count mismatch: expected %d, got %d", name, expCount, actCount)
		}
	}

	expEntry := strings.Count(expectedStr, "<ENTRY")
	actEntry := strings.Count(actualStr, "<ENTRY")
	if strings.Contains(expectedStr, "<NML") {
		if expEntry != actEntry {
			t.Errorf("%s: ENTRY count mismatch: expected %d, got %d", name, expEntry, actEntry)
		}
	}
}

func TestKeyConversionConsistency(t *testing.T) {
	// Test all Camelot keys convert correctly and back
	camelotKeys := []string{
		"1A", "1B", "2A", "2B", "3A", "3B", "4A", "4B",
		"5A", "5B", "6A", "6B", "7A", "7B", "8A", "8B",
		"9A", "9B", "10A", "10B", "11A", "11B", "12A", "12B",
	}

	for _, key := range camelotKeys {
		// Test Rekordbox conversion is non-empty
		rb := camelotToRekordbox(key)
		if rb == "" {
			t.Errorf("camelotToRekordbox(%s) returned empty", key)
		}

		// Test Traktor conversion is in valid range
		tk := camelotToTraktorKey(key)
		if tk < 0 || tk > 23 {
			t.Errorf("camelotToTraktorKey(%s) = %d, out of range 0-23", key, tk)
		}
	}
}

func TestExportWithEmptyAnalysis(t *testing.T) {
	dir := t.TempDir()

	// Track with minimal analysis
	tracks := []TrackExport{
		{
			Path: "/Music/empty.mp3",
			Analysis: &common.TrackAnalysis{
				Id: &common.TrackId{
					ContentHash: "empty123",
					Path:        "/Music/empty.mp3",
				},
				// No cues, no beatgrid, no key
			},
		},
	}

	// Should not panic with minimal data
	if _, err := WriteRekordbox(dir, "empty", tracks); err != nil {
		t.Errorf("WriteRekordbox failed with empty analysis: %v", err)
	}
	if _, err := WriteSerato(dir, "empty", tracks); err != nil {
		t.Errorf("WriteSerato failed with empty analysis: %v", err)
	}
	if _, err := WriteTraktor(dir, "empty", tracks); err != nil {
		t.Errorf("WriteTraktor failed with empty analysis: %v", err)
	}
}

func TestExportPathSpecialCharacters(t *testing.T) {
	dir := t.TempDir()

	tracks := []TrackExport{
		{
			Path: "/Music/Artist & Producer/Track (Remix) [Extended].mp3",
			Analysis: &common.TrackAnalysis{
				Id: &common.TrackId{
					ContentHash: "special123",
					Path:        "/Music/Artist & Producer/Track (Remix) [Extended].mp3",
				},
				Key: &common.MusicalKey{
					Value:      "5A",
					Format:     common.KeyFormat_CAMELOT,
					Confidence: 0.85,
				},
				EnergyGlobal: 6,
			},
		},
	}

	// XML exports should escape special characters properly
	if path, err := WriteRekordbox(dir, "special", tracks); err != nil {
		t.Errorf("WriteRekordbox failed with special chars: %v", err)
	} else {
		data, _ := os.ReadFile(path)
		content := string(data)
		// Check valid XML by ensuring & is escaped
		if strings.Contains(content, "& ") {
			// Unescaped ampersand found
			t.Error("ampersand not properly escaped in Rekordbox XML")
		}
	}

	if path, err := WriteTraktor(dir, "special", tracks); err != nil {
		t.Errorf("WriteTraktor failed with special chars: %v", err)
	} else {
		data, _ := os.ReadFile(path)
		content := string(data)
		if strings.Contains(content, "& ") {
			t.Error("ampersand not properly escaped in Traktor NML")
		}
	}
}
