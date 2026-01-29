package exporter

import (
	"archive/tar"
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cartomix/cancun/gen/go/common"
)

// TrackExport bundles a track path with its analysis.
type TrackExport struct {
	Path     string
	Analysis *common.TrackAnalysis
}

// Result contains paths to generated export artifacts.
type Result struct {
	PlaylistPath     string
	AnalysisJSONPath string
	CuesCSVPath      string
	VendorExports    []string
	BundlePath       string
	ChecksumsPath    string
}

// WriteGeneric writes M3U8, analysis JSON, and cues CSV exports.
func WriteGeneric(outputDir, playlistName string, tracks []TrackExport) (*Result, error) {
	if len(tracks) == 0 {
		return nil, fmt.Errorf("no tracks to export")
	}

	if playlistName == "" {
		playlistName = "set"
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, err
	}

	result := &Result{
		PlaylistPath:     filepath.Join(outputDir, playlistName+".m3u8"),
		AnalysisJSONPath: filepath.Join(outputDir, playlistName+"-analysis.json"),
		CuesCSVPath:      filepath.Join(outputDir, playlistName+"-cues.csv"),
		VendorExports:    []string{},
		BundlePath:       filepath.Join(outputDir, playlistName+"-bundle.tar.gz"),
		ChecksumsPath:    filepath.Join(outputDir, playlistName+"-checksums.txt"),
	}

	if err := writeM3U(result.PlaylistPath, tracks); err != nil {
		return nil, err
	}
	if err := writeAnalysisJSON(result.AnalysisJSONPath, tracks); err != nil {
		return nil, err
	}
	if err := writeCuesCSV(result.CuesCSVPath, tracks); err != nil {
		return nil, err
	}

	if err := writeChecksums(result.ChecksumsPath, result.PlaylistPath, result.AnalysisJSONPath, result.CuesCSVPath); err != nil {
		return nil, err
	}

	if err := writeBundle(result.BundlePath, result.PlaylistPath, result.AnalysisJSONPath, result.CuesCSVPath, result.ChecksumsPath); err != nil {
		return nil, err
	}

	return result, nil
}

func writeM3U(path string, tracks []TrackExport) error {
	var b strings.Builder
	b.WriteString("#EXTM3U\n")
	for _, t := range tracks {
		title := filepath.Base(t.Path)
		if meta := t.Analysis.GetId(); meta != nil && meta.Path != "" {
			title = filepath.Base(meta.Path)
		}
		b.WriteString(fmt.Sprintf("#EXTINF:0,%s\n", title))
		b.WriteString(fmt.Sprintln(t.Path))
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func writeAnalysisJSON(path string, tracks []TrackExport) error {
	analyses := make([]*common.TrackAnalysis, 0, len(tracks))
	for _, t := range tracks {
		analyses = append(analyses, t.Analysis)
	}
	bytes, err := json.MarshalIndent(analyses, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0o644)
}

func writeCuesCSV(path string, tracks []TrackExport) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"track_path", "cue_type", "beat_index", "time_seconds", "label"}); err != nil {
		return err
	}

	for _, t := range tracks {
		for _, cue := range t.Analysis.GetCuePoints() {
			if err := writer.Write([]string{
				t.Path,
				cue.GetType().String(),
				fmt.Sprintf("%d", cue.GetBeatIndex()),
				fmt.Sprintf("%.3f", cue.GetTime().AsDuration().Seconds()),
				cue.GetType().String(),
			}); err != nil {
				return err
			}
		}
	}

	writer.Flush()
	return writer.Error()
}

// writeChecksums writes a SHA256 manifest for the exported artifacts.
func writeChecksums(path string, files ...string) error {
	var b strings.Builder
	for _, fp := range files {
		sum, err := fileSHA256(fp)
		if err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf("%s  %s\n", sum, filepath.Base(fp)))
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// writeBundle creates a tar.gz containing the primary artifacts for quick sharing.
func writeBundle(bundlePath string, files ...string) error {
	f, err := os.Create(bundlePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, fp := range files {
		info, err := os.Stat(fp)
		if err != nil {
			return err
		}

		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = filepath.Base(fp)

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		data, err := os.ReadFile(fp)
		if err != nil {
			return err
		}
		if _, err := tw.Write(data); err != nil {
			return err
		}
	}

	return nil
}

// fileSHA256 is kept unexported for internal writer use.
func fileSHA256(path string) (string, error) { return FileSHA256(path) }
