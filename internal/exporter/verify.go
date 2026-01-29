package exporter

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// VerifyChecksums reads a sha256 manifest (format: "<hex>  <filename>") and
// verifies the referenced files exist and match. Returns nil if everything is OK.
func VerifyChecksums(manifestPath, baseDir string) error {
	f, err := os.Open(manifestPath)
	if err != nil {
		return fmt.Errorf("open manifest: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			return fmt.Errorf("invalid manifest line %d: %q", lineNo, line)
		}
		want := parts[0]
		name := parts[len(parts)-1]
		path := filepath.Join(baseDir, name)

		got, err := FileSHA256(path)
		if err != nil {
			return fmt.Errorf("hash %s: %w", path, err)
		}
		if !strings.EqualFold(got, want) {
			return fmt.Errorf("checksum mismatch for %s: want %s got %s", name, want, got)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read manifest: %w", err)
	}

	return nil
}

// fileSHA256 returns the hex SHA256 of a file.
func FileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
