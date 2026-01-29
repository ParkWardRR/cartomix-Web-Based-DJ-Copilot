package exporter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyChecksums(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	os.WriteFile(a, []byte("hello"), 0o644)
	os.WriteFile(b, []byte("world"), 0o644)

	sumA, _ := FileSHA256(a)
	sumB, _ := FileSHA256(b)
	manifest := filepath.Join(dir, "checksums.txt")
	os.WriteFile(manifest, []byte(sumA+"  a.txt\n"+sumB+"  b.txt\n"), 0o644)

	if err := VerifyChecksums(manifest, dir); err != nil {
		t.Fatalf("expected verify ok, got %v", err)
	}

	// Corrupt a file
	os.WriteFile(a, []byte("oops"), 0o644)
	if err := VerifyChecksums(manifest, dir); err == nil {
		t.Fatalf("expected checksum failure")
	}
}
