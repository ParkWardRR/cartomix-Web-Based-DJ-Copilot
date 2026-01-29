package main

import (
	"flag"
	"log"
	"path/filepath"

	"github.com/cartomix/cancun/internal/exporter"
)

// exportverify validates a checksum manifest emitted by ExportSet bundles.
func main() {
	manifest := flag.String("manifest", "", "path to checksums txt (e.g., set-checksums.txt)")
	dir := flag.String("dir", "", "directory containing files (defaults to manifest dir)")
	flag.Parse()

	if *manifest == "" {
		log.Fatal("manifest path required")
	}

	base := *dir
	if base == "" {
		base = filepath.Dir(*manifest)
	}

	if err := exporter.VerifyChecksums(*manifest, base); err != nil {
		log.Fatalf("verify failed: %v", err)
	}

	log.Printf("checksums OK for manifest %s", *manifest)
}
