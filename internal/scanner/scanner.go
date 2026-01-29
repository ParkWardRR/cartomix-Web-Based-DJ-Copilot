package scanner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cartomix/cancun/internal/storage"
)

// SupportedFormats lists the audio formats we can process.
var SupportedFormats = map[string]bool{
	".mp3":  true,
	".flac": true,
	".wav":  true,
	".aiff": true,
	".aif":  true,
	".m4a":  true,
	".ogg":  true,
	".opus": true,
}

// Scanner recursively scans directories for audio files.
type Scanner struct {
	db     *storage.DB
	logger *slog.Logger
}

// ScanResult holds the result of scanning a file.
type ScanResult struct {
	Path        string
	ContentHash string
	TrackID     int64
	IsNew       bool
	Error       error
}

// ScanProgress reports scanning progress.
type ScanProgress struct {
	Path        string
	Status      string // queued, processing, done, skipped, error
	Error       string
	Processed   int64
	Total       int64
	TrackID     int64
	IsNew       bool
	ContentHash string
}

// NewScanner creates a new file scanner.
func NewScanner(db *storage.DB, logger *slog.Logger) *Scanner {
	return &Scanner{db: db, logger: logger}
}

// Scan recursively scans the given roots for audio files.
// Progress is reported via the progress channel.
func (s *Scanner) Scan(ctx context.Context, roots []string, forceRescan bool, progress chan<- ScanProgress) error {
	defer close(progress)

	// First pass: count files
	var total int64
	for _, root := range roots {
		count, err := s.countFiles(root)
		if err != nil {
			s.logger.Warn("failed to count files in root", "root", root, "error", err)
			continue
		}
		total += count
	}

	// Second pass: process files
	var processed int64
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // Skip errors, continue scanning
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if d.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(path))
			if !SupportedFormats[ext] {
				return nil // Skip unsupported formats
			}

			result := s.processFile(path, forceRescan)
			processed++

			status := "done"
			errMsg := ""
			if result.Error != nil {
				status = "error"
				errMsg = result.Error.Error()
			} else if !result.IsNew {
				status = "skipped"
			}

			select {
			case progress <- ScanProgress{
				Path:        path,
				Status:      status,
				Error:       errMsg,
				Processed:   processed,
				Total:       total,
				TrackID:     result.TrackID,
				IsNew:       result.IsNew,
				ContentHash: result.ContentHash,
			}:
			case <-ctx.Done():
				return ctx.Err()
			}

			return nil
		})

		if err != nil && err != context.Canceled {
			s.logger.Error("scan error", "root", root, "error", err)
		}
	}

	return nil
}

func (s *Scanner) countFiles(root string) (int64, error) {
	var count int64
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if SupportedFormats[ext] {
			count++
		}
		return nil
	})
	return count, err
}

func (s *Scanner) processFile(path string, forceRescan bool) ScanResult {
	result := ScanResult{Path: path}

	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		result.Error = err
		return result
	}

	// Compute content hash (first 64KB for speed)
	hash, err := ComputeHash(path)
	if err != nil {
		result.Error = err
		return result
	}
	result.ContentHash = hash

	// Check if already in database
	if !forceRescan {
		existing, err := s.db.GetTrackByHash(hash)
		if err == nil && existing != nil {
			result.TrackID = existing.ID
			result.IsNew = false
			return result
		}
	}

	// Insert/update track
	track := &storage.Track{
		ContentHash:    hash,
		Path:           path,
		FileSize:       info.Size(),
		FileModifiedAt: info.ModTime(),
	}

	// TODO: Extract metadata (title, artist, album) from tags

	trackID, err := s.db.UpsertTrack(track)
	if err != nil {
		result.Error = err
		return result
	}

	result.TrackID = trackID
	result.IsNew = true
	return result
}

// EnqueueAnalysis creates analysis jobs for the given track IDs.
func (s *Scanner) EnqueueAnalysis(trackIDs []int64, priority int) error {
	for _, trackID := range trackIDs {
		_, err := s.db.CreateJob(storage.JobTypeAnalyze, priority, map[string]any{
			"track_id": trackID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// ComputeHash returns a deterministic, fast hash of the first 64KB.
func ComputeHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha256.New()
	// Hash first 64KB for speed - content hash is just for identity
	_, err = io.CopyN(h, file, 64*1024)
	if err != nil && err != io.EOF {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// HashCache provides a simple in-memory cache for file hashes.
type HashCache struct {
	cache map[string]cacheEntry
}

type cacheEntry struct {
	hash    string
	modTime time.Time
}

// NewHashCache creates a new hash cache.
func NewHashCache() *HashCache {
	return &HashCache{cache: make(map[string]cacheEntry)}
}

// Get returns a cached hash if the file hasn't been modified.
func (c *HashCache) Get(path string, modTime time.Time) (string, bool) {
	entry, ok := c.cache[path]
	if !ok {
		return "", false
	}
	if !entry.modTime.Equal(modTime) {
		return "", false
	}
	return entry.hash, true
}

// Set caches a hash for a file.
func (c *HashCache) Set(path string, hash string, modTime time.Time) {
	c.cache[path] = cacheEntry{hash: hash, modTime: modTime}
}
