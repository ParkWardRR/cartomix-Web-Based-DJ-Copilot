package exporter

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TraktorNML represents the root element of a Traktor NML export.
type TraktorNML struct {
	XMLName  xml.Name           `xml:"NML"`
	Version  int                `xml:"VERSION,attr"`
	Head     TraktorHead        `xml:"HEAD"`
	MusicFolders TraktorMusicFolders `xml:"MUSICFOLDERS,omitempty"`
	Collection TraktorCollection `xml:"COLLECTION"`
	Sets     TraktorSets        `xml:"SETS,omitempty"`
	Playlists TraktorPlaylists  `xml:"PLAYLISTS"`
}

// TraktorHead contains metadata about the NML file.
type TraktorHead struct {
	Company string `xml:"COMPANY,attr"`
	Program string `xml:"PROGRAM,attr"`
}

// TraktorMusicFolders contains paths to music directories.
type TraktorMusicFolders struct {
	Folders []TraktorMusicFolder `xml:"FOLDER,omitempty"`
}

// TraktorMusicFolder is a single music folder entry.
type TraktorMusicFolder struct {
	Path string `xml:"PATH,attr"`
}

// TraktorCollection holds all tracks in the collection.
type TraktorCollection struct {
	Entries int            `xml:"ENTRIES,attr"`
	Tracks  []TraktorEntry `xml:"ENTRY"`
}

// TraktorEntry represents a single track entry.
type TraktorEntry struct {
	ModifiedDate  string              `xml:"MODIFIED_DATE,attr,omitempty"`
	ModifiedTime  int                 `xml:"MODIFIED_TIME,attr,omitempty"`
	AudioID       string              `xml:"AUDIO_ID,attr,omitempty"`
	Title         string              `xml:"TITLE,attr"`
	Artist        string              `xml:"ARTIST,attr,omitempty"`
	Location      TraktorLocation     `xml:"LOCATION"`
	Album         TraktorAlbum        `xml:"ALBUM,omitempty"`
	Info          TraktorInfo         `xml:"INFO,omitempty"`
	Tempo         TraktorTempo        `xml:"TEMPO,omitempty"`
	Loudness      TraktorLoudness     `xml:"LOUDNESS,omitempty"`
	MusicalKey    TraktorMusicalKey   `xml:"MUSICAL_KEY,omitempty"`
	CuePoints     []TraktorCueV2      `xml:"CUE_V2,omitempty"`
}

// TraktorLocation contains the file path information.
type TraktorLocation struct {
	Dir     string `xml:"DIR,attr"`
	File    string `xml:"FILE,attr"`
	Volume  string `xml:"VOLUME,attr,omitempty"`
	VolumeID string `xml:"VOLUMEID,attr,omitempty"`
}

// TraktorAlbum contains album information.
type TraktorAlbum struct {
	Title    string `xml:"TITLE,attr,omitempty"`
	Track    int    `xml:"TRACK,attr,omitempty"`
	OfTracks int    `xml:"OF_TRACKS,attr,omitempty"`
}

// TraktorInfo contains track metadata.
type TraktorInfo struct {
	Bitrate     int     `xml:"BITRATE,attr,omitempty"`
	Genre       string  `xml:"GENRE,attr,omitempty"`
	Label       string  `xml:"LABEL,attr,omitempty"`
	Comment     string  `xml:"COMMENT,attr,omitempty"`
	CoverArtID  string  `xml:"COVERARTID,attr,omitempty"`
	Key         string  `xml:"KEY,attr,omitempty"`
	PlayCount   int     `xml:"PLAYCOUNT,attr,omitempty"`
	Playtime    int     `xml:"PLAYTIME,attr,omitempty"`
	PlaytimeF   float64 `xml:"PLAYTIME_FLOAT,attr,omitempty"`
	ImportDate  string  `xml:"IMPORT_DATE,attr,omitempty"`
	LastPlayed  string  `xml:"LAST_PLAYED,attr,omitempty"`
	Ranking     int     `xml:"RANKING,attr,omitempty"`
	ReleaseDate string  `xml:"RELEASE_DATE,attr,omitempty"`
}

// TraktorTempo contains BPM and timing information.
type TraktorTempo struct {
	BPM        float64 `xml:"BPM,attr"`
	BPMQuality float64 `xml:"BPM_QUALITY,attr,omitempty"`
}

// TraktorLoudness contains loudness analysis data.
type TraktorLoudness struct {
	PeakDB     float64 `xml:"PEAK_DB,attr,omitempty"`
	PerceivedDB float64 `xml:"PERCEIVED_DB,attr,omitempty"`
	AnalyzedDB  float64 `xml:"ANALYZED_DB,attr,omitempty"`
}

// TraktorMusicalKey contains key detection data.
type TraktorMusicalKey struct {
	Value int `xml:"VALUE,attr"`
}

// TraktorCueV2 represents a cue point in Traktor format.
type TraktorCueV2 struct {
	Name     string  `xml:"NAME,attr,omitempty"`
	Displ    int     `xml:"DISPL_ORDER,attr"`
	Type     int     `xml:"TYPE,attr"`
	Start    float64 `xml:"START,attr"`
	Len      float64 `xml:"LEN,attr,omitempty"`
	Repeats  int     `xml:"REPEATS,attr,omitempty"`
	Hotcue   int     `xml:"HOTCUE,attr"`
}

// TraktorSets contains saved set/playlist history.
type TraktorSets struct {
	Entries int `xml:"ENTRIES,attr,omitempty"`
}

// TraktorPlaylists is the container for playlists.
type TraktorPlaylists struct {
	Node TraktorPlaylistNode `xml:"NODE"`
}

// TraktorPlaylistNode represents a playlist or folder.
type TraktorPlaylistNode struct {
	Type      string                `xml:"TYPE,attr"`
	Name      string                `xml:"NAME,attr"`
	Subnodes  []TraktorPlaylistNode `xml:"SUBNODES>NODE,omitempty"`
	Playlist  *TraktorPlaylist      `xml:"PLAYLIST,omitempty"`
}

// TraktorPlaylist contains the actual playlist entries.
type TraktorPlaylist struct {
	Entries int                    `xml:"ENTRIES,attr"`
	Type    string                 `xml:"TYPE,attr"`
	UUID    string                 `xml:"UUID,attr,omitempty"`
	Tracks  []TraktorPlaylistEntry `xml:"ENTRY"`
}

// TraktorPlaylistEntry is a track reference in a playlist.
type TraktorPlaylistEntry struct {
	PrimaryKey TraktorPrimaryKey `xml:"PRIMARYKEY"`
}

// TraktorPrimaryKey identifies a track by its file location.
type TraktorPrimaryKey struct {
	Type string `xml:"TYPE,attr"`
	Key  string `xml:"KEY,attr"`
}

// WriteTraktor exports tracks to Traktor NML format.
func WriteTraktor(outputDir, playlistName string, tracks []TrackExport) (string, error) {
	if len(tracks) == 0 {
		return "", fmt.Errorf("no tracks to export")
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", err
	}

	entries := make([]TraktorEntry, 0, len(tracks))
	playlistEntries := make([]TraktorPlaylistEntry, 0, len(tracks))
	musicFolders := make(map[string]bool)

	for _, t := range tracks {
		analysis := t.Analysis
		absPath, err := filepath.Abs(t.Path)
		if err != nil {
			absPath = t.Path
		}

		dir := filepath.Dir(absPath)
		file := filepath.Base(absPath)
		musicFolders[dir] = true

		// Get BPM
		bpm := 120.0
		if grid := analysis.GetBeatgrid(); grid != nil && len(grid.GetTempoMap()) > 0 {
			bpm = grid.GetTempoMap()[0].GetBpm()
		}

		// Get duration
		playtime := 0
		playtimeF := 0.0
		if grid := analysis.GetBeatgrid(); grid != nil && len(grid.GetBeats()) > 0 {
			beats := grid.GetBeats()
			lastBeat := beats[len(beats)-1]
			playtimeF = lastBeat.GetTime().AsDuration().Seconds()
			playtime = int(playtimeF)
		}

		// Convert key to Traktor format
		keyValue := camelotToTraktorKey(analysis.GetKey().GetValue())

		// Build cue points
		cues := make([]TraktorCueV2, 0)
		for i, cue := range analysis.GetCuePoints() {
			cues = append(cues, TraktorCueV2{
				Name:   cue.GetType().String(),
				Displ:  i,
				Type:   cueTypeToTraktorType(cue.GetType().String()),
				Start:  cue.GetTime().AsDuration().Seconds() * 1000, // Traktor uses milliseconds
				Hotcue: i,
			})
		}

		// Build location key for playlist reference
		locationKey := "/:file://localhost" + strings.ReplaceAll(absPath, "/", "/:")

		entry := TraktorEntry{
			Title:  strings.TrimSuffix(file, filepath.Ext(file)),
			Artist: "", // Would come from tags
			Location: TraktorLocation{
				Dir:    strings.ReplaceAll(dir, "/", "/:") + "/:",
				File:   file,
				Volume: "",
			},
			Info: TraktorInfo{
				Playtime:   playtime,
				PlaytimeF:  playtimeF,
				Key:        analysis.GetKey().GetValue(),
				ImportDate: "2026/01/29",
			},
			Tempo: TraktorTempo{
				BPM:        bpm,
				BPMQuality: float64(analysis.GetBeatgrid().GetConfidence()),
			},
			MusicalKey: TraktorMusicalKey{
				Value: keyValue,
			},
			CuePoints: cues,
		}

		entries = append(entries, entry)
		playlistEntries = append(playlistEntries, TraktorPlaylistEntry{
			PrimaryKey: TraktorPrimaryKey{
				Type: "TRACK",
				Key:  locationKey,
			},
		})
	}

	// Build music folders list
	folders := make([]TraktorMusicFolder, 0, len(musicFolders))
	for dir := range musicFolders {
		folders = append(folders, TraktorMusicFolder{
			Path: strings.ReplaceAll(dir, "/", "/:") + "/:",
		})
	}

	// Build the NML structure
	nml := TraktorNML{
		Version: 19,
		Head: TraktorHead{
			Company: "Algiers",
			Program: "Algiers DJ Prep",
		},
		MusicFolders: TraktorMusicFolders{
			Folders: folders,
		},
		Collection: TraktorCollection{
			Entries: len(entries),
			Tracks:  entries,
		},
		Sets: TraktorSets{Entries: 0},
		Playlists: TraktorPlaylists{
			Node: TraktorPlaylistNode{
				Type: "FOLDER",
				Name: "$ROOT",
				Subnodes: []TraktorPlaylistNode{
					{
						Type: "PLAYLIST",
						Name: playlistName,
						Playlist: &TraktorPlaylist{
							Entries: len(playlistEntries),
							Type:    "LIST",
							Tracks:  playlistEntries,
						},
					},
				},
			},
		},
	}

	// Marshal to XML
	outputPath := filepath.Join(outputDir, playlistName+"-traktor.nml")
	output, err := xml.MarshalIndent(nml, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal traktor NML: %w", err)
	}

	// Add XML declaration
	xmlContent := []byte(xml.Header + string(output))
	if err := os.WriteFile(outputPath, xmlContent, 0o644); err != nil {
		return "", fmt.Errorf("failed to write traktor NML: %w", err)
	}

	return outputPath, nil
}

// camelotToTraktorKey converts Camelot notation to Traktor's key value.
// Traktor uses integers 0-23 for keys (0=C, 1=C#, etc. for major, 12+ for minor)
func camelotToTraktorKey(camelot string) int {
	// Camelot to Traktor key value mapping
	camelotMap := map[string]int{
		"1A": 20,  // Abm
		"1B": 11,  // B
		"2A": 15,  // Ebm
		"2B": 6,   // Gb
		"3A": 22,  // Bbm
		"3B": 1,   // Db
		"4A": 17,  // Fm
		"4B": 8,   // Ab
		"5A": 12,  // Cm
		"5B": 3,   // Eb
		"6A": 19,  // Gm
		"6B": 10,  // Bb
		"7A": 14,  // Dm
		"7B": 5,   // F
		"8A": 21,  // Am
		"8B": 0,   // C
		"9A": 16,  // Em
		"9B": 7,   // G
		"10A": 23, // Bm
		"10B": 2,  // D
		"11A": 18, // F#m
		"11B": 9,  // A
		"12A": 13, // C#m
		"12B": 4,  // E
	}
	if val, ok := camelotMap[camelot]; ok {
		return val
	}
	return 0
}

// cueTypeToTraktorType converts cue type to Traktor's cue type integer.
func cueTypeToTraktorType(cueType string) int {
	// Traktor cue types: 0=cue, 1=fade-in, 2=fade-out, 3=load, 4=loop
	types := map[string]int{
		"CUE_TYPE_INTRO":    0,
		"CUE_TYPE_DROP":     0,
		"CUE_TYPE_BREAK":    0,
		"CUE_TYPE_BUILD":    0,
		"CUE_TYPE_OUTRO":    0,
		"CUE_TYPE_LOOP":     4,
		"CUE_TYPE_FADE_IN":  1,
		"CUE_TYPE_FADE_OUT": 2,
	}
	if t, ok := types[cueType]; ok {
		return t
	}
	return 0
}
