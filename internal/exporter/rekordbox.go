package exporter

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// RekordboxXML is the root element of a Rekordbox XML export.
type RekordboxXML struct {
	XMLName xml.Name          `xml:"DJ_PLAYLISTS"`
	Version string            `xml:"Version,attr"`
	Product RekordboxProduct  `xml:"PRODUCT"`
	Collection RekordboxCollection `xml:"COLLECTION"`
	Playlists  RekordboxPlaylists  `xml:"PLAYLISTS"`
}

// RekordboxProduct identifies the exporting application.
type RekordboxProduct struct {
	Name    string `xml:"Name,attr"`
	Version string `xml:"Version,attr"`
	Company string `xml:"Company,attr"`
}

// RekordboxCollection holds all tracks.
type RekordboxCollection struct {
	Entries int             `xml:"Entries,attr"`
	Tracks  []RekordboxTrack `xml:"TRACK"`
}

// RekordboxTrack represents a single track in Rekordbox format.
type RekordboxTrack struct {
	TrackID     int    `xml:"TrackID,attr"`
	Name        string `xml:"Name,attr"`
	Artist      string `xml:"Artist,attr"`
	Album       string `xml:"Album,attr,omitempty"`
	Genre       string `xml:"Genre,attr,omitempty"`
	Kind        string `xml:"Kind,attr,omitempty"`
	Size        int64  `xml:"Size,attr,omitempty"`
	TotalTime   int    `xml:"TotalTime,attr"`
	DateAdded   string `xml:"DateAdded,attr,omitempty"`
	BitRate     int    `xml:"BitRate,attr,omitempty"`
	SampleRate  int    `xml:"SampleRate,attr,omitempty"`
	AverageBpm  string `xml:"AverageBpm,attr"`
	Tonality    string `xml:"Tonality,attr,omitempty"`
	Rating      int    `xml:"Rating,attr,omitempty"`
	Location    string `xml:"Location,attr"`
	PositionMarks []RekordboxPositionMark `xml:"POSITION_MARK,omitempty"`
	Tempo       []RekordboxTempo `xml:"TEMPO,omitempty"`
}

// RekordboxPositionMark represents a cue point or memory cue.
type RekordboxPositionMark struct {
	Name    string  `xml:"Name,attr,omitempty"`
	Type    int     `xml:"Type,attr"`
	Start   string  `xml:"Start,attr"`
	End     string  `xml:"End,attr,omitempty"`
	Num     int     `xml:"Num,attr"`
	Red     int     `xml:"Red,attr,omitempty"`
	Green   int     `xml:"Green,attr,omitempty"`
	Blue    int     `xml:"Blue,attr,omitempty"`
}

// RekordboxTempo represents a tempo marker.
type RekordboxTempo struct {
	Inizio string `xml:"Inizio,attr"`
	Bpm    string `xml:"Bpm,attr"`
	Metro  string `xml:"Metro,attr"`
	Battito int   `xml:"Battito,attr"`
}

// RekordboxPlaylists is the playlists container.
type RekordboxPlaylists struct {
	Node RekordboxPlaylistNode `xml:"NODE"`
}

// RekordboxPlaylistNode represents a playlist folder or playlist.
type RekordboxPlaylistNode struct {
	Type     int                     `xml:"Type,attr"`
	Name     string                  `xml:"Name,attr"`
	Count    int                     `xml:"Count,attr,omitempty"`
	Entries  int                     `xml:"Entries,attr,omitempty"`
	KeyType  int                     `xml:"KeyType,attr,omitempty"`
	Children []RekordboxPlaylistNode `xml:"NODE,omitempty"`
	Tracks   []RekordboxPlaylistTrack `xml:"TRACK,omitempty"`
}

// RekordboxPlaylistTrack is a track reference in a playlist.
type RekordboxPlaylistTrack struct {
	Key int `xml:"Key,attr"`
}

// WriteRekordbox exports tracks to Rekordbox XML format.
func WriteRekordbox(outputDir, playlistName string, tracks []TrackExport) (string, error) {
	if len(tracks) == 0 {
		return "", fmt.Errorf("no tracks to export")
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", err
	}

	rbTracks := make([]RekordboxTrack, 0, len(tracks))
	playlistTracks := make([]RekordboxPlaylistTrack, 0, len(tracks))

	for i, t := range tracks {
		trackID := i + 1
		analysis := t.Analysis

		// Parse duration from sections or estimate from beatgrid
		totalTime := 0
		if grid := analysis.GetBeatgrid(); grid != nil && len(grid.GetBeats()) > 0 {
			beats := grid.GetBeats()
			lastBeat := beats[len(beats)-1]
			totalTime = int(lastBeat.GetTime().AsDuration().Seconds())
		}

		// Convert key to Rekordbox tonality format (e.g., "8A" -> "Am")
		tonality := ""
		if key := analysis.GetKey(); key != nil {
			tonality = camelotToRekordbox(key.GetValue())
		}

		// Build position marks from cue points
		positionMarks := make([]RekordboxPositionMark, 0)
		for j, cue := range analysis.GetCuePoints() {
			color := cueTypeToColor(cue.GetType().String())
			positionMarks = append(positionMarks, RekordboxPositionMark{
				Name:  cue.GetType().String(),
				Type:  0, // 0 = cue, 1 = fade-in, 2 = fade-out, 3 = load, 4 = loop
				Start: fmt.Sprintf("%.3f", cue.GetTime().AsDuration().Seconds()),
				Num:   j,
				Red:   color[0],
				Green: color[1],
				Blue:  color[2],
			})
		}

		// Build tempo markers
		tempoMarks := make([]RekordboxTempo, 0)
		if grid := analysis.GetBeatgrid(); grid != nil && len(grid.GetBeats()) > 0 {
			bpm := 120.0
			if len(grid.GetTempoMap()) > 0 {
				bpm = grid.GetTempoMap()[0].GetBpm()
			}
			tempoMarks = append(tempoMarks, RekordboxTempo{
				Inizio:  "0.000",
				Bpm:     fmt.Sprintf("%.2f", bpm),
				Metro:   "4/4",
				Battito: 1,
			})
		}

		// Convert file path to file:// URL format for Rekordbox
		location := pathToFileURL(t.Path)

		rbTracks = append(rbTracks, RekordboxTrack{
			TrackID:       trackID,
			Name:          filepath.Base(strings.TrimSuffix(t.Path, filepath.Ext(t.Path))),
			Artist:        analysis.GetId().GetPath(), // Placeholder - should come from tags
			TotalTime:     totalTime,
			DateAdded:     time.Now().Format("2006-01-02"),
			AverageBpm:    fmt.Sprintf("%.2f", analysis.GetBeatgrid().GetTempoMap()[0].GetBpm()),
			Tonality:      tonality,
			Location:      location,
			PositionMarks: positionMarks,
			Tempo:         tempoMarks,
		})

		playlistTracks = append(playlistTracks, RekordboxPlaylistTrack{Key: trackID})
	}

	// Build the XML structure
	rbXML := RekordboxXML{
		Version: "1.0.0",
		Product: RekordboxProduct{
			Name:    "Algiers",
			Version: "0.1.0",
			Company: "Cartomix",
		},
		Collection: RekordboxCollection{
			Entries: len(rbTracks),
			Tracks:  rbTracks,
		},
		Playlists: RekordboxPlaylists{
			Node: RekordboxPlaylistNode{
				Type: 0, // Root folder
				Name: "ROOT",
				Children: []RekordboxPlaylistNode{
					{
						Type:    1, // Playlist
						Name:    playlistName,
						Entries: len(playlistTracks),
						KeyType: 0,
						Tracks:  playlistTracks,
					},
				},
			},
		},
	}

	// Marshal to XML
	outputPath := filepath.Join(outputDir, playlistName+"-rekordbox.xml")
	output, err := xml.MarshalIndent(rbXML, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal rekordbox XML: %w", err)
	}

	// Add XML declaration
	xmlContent := []byte(xml.Header + string(output))
	if err := os.WriteFile(outputPath, xmlContent, 0o644); err != nil {
		return "", fmt.Errorf("failed to write rekordbox XML: %w", err)
	}

	return outputPath, nil
}

// camelotToRekordbox converts Camelot key notation to Rekordbox tonality.
func camelotToRekordbox(camelot string) string {
	// Camelot to musical key mapping
	camelotMap := map[string]string{
		"1A": "Abm", "1B": "B",
		"2A": "Ebm", "2B": "Gb",
		"3A": "Bbm", "3B": "Db",
		"4A": "Fm",  "4B": "Ab",
		"5A": "Cm",  "5B": "Eb",
		"6A": "Gm",  "6B": "Bb",
		"7A": "Dm",  "7B": "F",
		"8A": "Am",  "8B": "C",
		"9A": "Em",  "9B": "G",
		"10A": "Bm", "10B": "D",
		"11A": "Gbm", "11B": "A",
		"12A": "Dbm", "12B": "E",
	}
	if key, ok := camelotMap[camelot]; ok {
		return key
	}
	return camelot
}

// cueTypeToColor returns RGB values for cue point colors.
func cueTypeToColor(cueType string) [3]int {
	colors := map[string][3]int{
		"CUE_TYPE_INTRO":    {40, 226, 20},   // Green
		"CUE_TYPE_DROP":     {230, 20, 20},   // Red
		"CUE_TYPE_BREAK":    {20, 130, 230},  // Blue
		"CUE_TYPE_BUILD":    {230, 150, 20},  // Orange
		"CUE_TYPE_OUTRO":    {200, 20, 200},  // Purple
		"CUE_TYPE_LOOP":     {230, 230, 20},  // Yellow
		"CUE_TYPE_FADE_IN":  {20, 200, 200},  // Cyan
		"CUE_TYPE_FADE_OUT": {150, 150, 150}, // Gray
	}
	if color, ok := colors[cueType]; ok {
		return color
	}
	return [3]int{200, 200, 200} // Default gray
}

// pathToFileURL converts a file path to file:// URL format.
func pathToFileURL(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	// Rekordbox expects file://localhost/ prefix on macOS
	return "file://localhost" + absPath
}

// intAttr is a helper to convert int to attribute string.
func intAttr(val int) string {
	return strconv.Itoa(val)
}
