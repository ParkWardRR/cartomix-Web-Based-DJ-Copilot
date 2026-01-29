package exporter

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf16"
)

// SeratoCrate represents the structure of a Serato crate file.
// Serato crate files use a simple binary format with versioning and track paths.

const (
	seratoCrateVersion = "81.0"
	seratoCrateMagic   = "vrsn"
	seratoTrackMagic   = "otrk"
	seratoPathMagic    = "ptrk"
)

// WriteSerato exports tracks to Serato crate format (.crate).
// This creates a Serato DJ-compatible crate file that can be imported.
func WriteSerato(outputDir, playlistName string, tracks []TrackExport) (string, error) {
	if len(tracks) == 0 {
		return "", fmt.Errorf("no tracks to export")
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", err
	}

	var buf bytes.Buffer

	// Write version header
	// Format: "vrsn" + 4-byte length (big endian) + UTF-16BE version string
	versionBytes := encodeUTF16BE(seratoCrateVersion)
	buf.WriteString(seratoCrateMagic)
	binary.Write(&buf, binary.BigEndian, uint32(len(versionBytes)))
	buf.Write(versionBytes)

	// Write each track
	for _, t := range tracks {
		absPath, err := filepath.Abs(t.Path)
		if err != nil {
			absPath = t.Path
		}

		// Track entry format: "otrk" + length + path
		// Path format: "ptrk" + length + UTF-16BE path string
		pathBytes := encodeUTF16BE(absPath)

		// Write path chunk
		var pathChunk bytes.Buffer
		pathChunk.WriteString(seratoPathMagic)
		binary.Write(&pathChunk, binary.BigEndian, uint32(len(pathBytes)))
		pathChunk.Write(pathBytes)

		// Write track chunk containing path chunk
		buf.WriteString(seratoTrackMagic)
		binary.Write(&buf, binary.BigEndian, uint32(pathChunk.Len()))
		buf.Write(pathChunk.Bytes())
	}

	outputPath := filepath.Join(outputDir, playlistName+".crate")
	if err := os.WriteFile(outputPath, buf.Bytes(), 0o644); err != nil {
		return "", fmt.Errorf("failed to write serato crate: %w", err)
	}

	// Also write a CSV with cue point data for Serato marker reference
	csvPath := filepath.Join(outputDir, playlistName+"-serato-cues.csv")
	if err := writeSeratoCuesCSV(csvPath, tracks); err != nil {
		// Non-fatal - main crate file was written
		return outputPath, nil
	}

	return outputPath, nil
}

// encodeUTF16BE encodes a string as UTF-16 Big Endian.
func encodeUTF16BE(s string) []byte {
	runes := []rune(s)
	u16 := utf16.Encode(runes)

	buf := make([]byte, len(u16)*2)
	for i, r := range u16 {
		buf[i*2] = byte(r >> 8)
		buf[i*2+1] = byte(r)
	}
	return buf
}

// writeSeratoCuesCSV writes a supplementary CSV file with cue point data.
// Serato stores cues in ID3 GEOB tags, but a CSV provides reference data.
func writeSeratoCuesCSV(path string, tracks []TrackExport) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Header
	file.WriteString("track_path,cue_index,cue_type,position_ms,color_hex,name\n")

	for _, t := range tracks {
		for i, cue := range t.Analysis.GetCuePoints() {
			positionMs := int64(cue.GetTime().AsDuration().Milliseconds())
			color := seratoCueColor(cue.GetType().String())
			name := cue.GetType().String()

			line := fmt.Sprintf("%q,%d,%s,%d,%s,%q\n",
				t.Path, i, cue.GetType().String(), positionMs, color, name)
			file.WriteString(line)
		}
	}

	return nil
}

// seratoCueColor returns the hex color code for a Serato cue point.
// Serato uses specific colors for hot cues.
func seratoCueColor(cueType string) string {
	colors := map[string]string{
		"CUE_TYPE_INTRO":    "CC0000", // Red
		"CUE_TYPE_DROP":     "CC4400", // Orange
		"CUE_TYPE_BREAK":    "CC8800", // Yellow
		"CUE_TYPE_BUILD":    "00CC00", // Green
		"CUE_TYPE_OUTRO":    "0088CC", // Blue
		"CUE_TYPE_LOOP":     "0000CC", // Dark Blue
		"CUE_TYPE_FADE_IN":  "8800CC", // Purple
		"CUE_TYPE_FADE_OUT": "CC0088", // Pink
	}
	if color, ok := colors[cueType]; ok {
		return color
	}
	return "888888" // Default gray
}

// SeratoMarkersV2 represents Serato's cue point marker format.
// This structure is used for writing Serato markers to ID3 GEOB tags.
// Note: Full ID3 tag writing requires additional library support.
type SeratoMarkersV2 struct {
	Version   uint8
	NumCues   uint8
	CuePoints []SeratoCuePoint
}

// SeratoCuePoint represents a single cue point in Serato format.
type SeratoCuePoint struct {
	Index      uint8
	Position   uint32 // Milliseconds
	Color      [3]byte
	LoopActive uint8
	LoopEnd    uint32
	Name       string
}

// EncodeSeratoMarkers encodes cue points to Serato markers binary format.
// This can be written to ID3 GEOB tags with frame ID "Serato Markers2".
func EncodeSeratoMarkers(tracks []TrackExport) map[string][]byte {
	markers := make(map[string][]byte)

	for _, t := range tracks {
		var buf bytes.Buffer
		cues := t.Analysis.GetCuePoints()

		// Version
		buf.WriteByte(0x02)

		// Cue count
		buf.WriteByte(byte(len(cues)))

		for i, cue := range cues {
			if i >= 8 {
				break // Serato supports max 8 hot cues
			}

			// Cue index
			buf.WriteByte(byte(i))

			// Position in milliseconds (big endian uint32)
			posMs := uint32(cue.GetTime().AsDuration().Milliseconds())
			binary.Write(&buf, binary.BigEndian, posMs)

			// Color RGB
			color := cueTypeToRGB(cue.GetType().String())
			buf.Write(color[:])

			// Loop active (0 = not a loop)
			buf.WriteByte(0)

			// Loop end (0 if not a loop)
			binary.Write(&buf, binary.BigEndian, uint32(0))
		}

		markers[t.Path] = buf.Bytes()
	}

	return markers
}

// cueTypeToRGB returns RGB bytes for a cue type.
func cueTypeToRGB(cueType string) [3]byte {
	colors := map[string][3]byte{
		"CUE_TYPE_INTRO":    {0xCC, 0x00, 0x00}, // Red
		"CUE_TYPE_DROP":     {0xCC, 0x44, 0x00}, // Orange
		"CUE_TYPE_BREAK":    {0xCC, 0x88, 0x00}, // Yellow
		"CUE_TYPE_BUILD":    {0x00, 0xCC, 0x00}, // Green
		"CUE_TYPE_OUTRO":    {0x00, 0x88, 0xCC}, // Blue
		"CUE_TYPE_LOOP":     {0x00, 0x00, 0xCC}, // Dark Blue
		"CUE_TYPE_FADE_IN":  {0x88, 0x00, 0xCC}, // Purple
		"CUE_TYPE_FADE_OUT": {0xCC, 0x00, 0x88}, // Pink
	}
	if color, ok := colors[cueType]; ok {
		return color
	}
	return [3]byte{0x88, 0x88, 0x88} // Default gray
}
