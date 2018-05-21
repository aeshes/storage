package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// LocalFile Describes file saved on disk
type LocalFile struct {
	Path   string
	Hash   string
	Handle *os.File
	Prev   *LocalFile
}

// Sha256 Calculates sha-256 checksum of this file
func (f *LocalFile) Sha256() string {
	file, err := os.Open(f.Path)
	if err != nil {
		log.Println(err)
		return ""
	}

	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		log.Println(err)
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%x", hasher.Sum(nil))
	return b.String()
}

// Remove removes file from disk
func (f *LocalFile) Remove() {
	err := os.Remove(f.Path)
	if err != nil {
		log.Println(err)
	}
}
