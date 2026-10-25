// Package coverops implements explicit, offline cover maintenance, not HTTP uploads.
package coverops

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/storage/covers"
)

var pilotSlugs = [...]string{
	"a-christmas-carol",
	"pride-and-prejudice",
	"frankenstein",
	"alices-adventures-in-wonderland",
	"twenty-thousand-leagues-under-the-seas",
}

type manifest struct {
	Version  int             `json:"version"`
	BundleID string          `json:"bundle_id"`
	Covers   []manifestCover `json:"covers"`
}

type manifestCover struct {
	Slug       string        `json:"slug"`
	UsageBasis string        `json:"usage_basis"`
	IsReviewed bool          `json:"reviewed"`
	Main       manifestFile  `json:"main"`
	Thumbnail  *manifestFile `json:"thumbnail,omitempty"`
}

type manifestFile struct {
	File   string `json:"file"`
	SHA256 string `json:"sha256"`
}

// Bundle can only acquire prepared bytes through LoadBundle validation.
// Source artwork and rights evidence stay in the separately retained bundle.
type Bundle struct {
	ID      string
	entries []preparedCover
}

type preparedCover struct {
	slug      string
	local     books.LocalFrontCover
	main      []byte
	thumbnail []byte
}

// LoadBundle verifies a caller-pinned manifest checksum and the entire five-book
// delivery before any publication or database write. Files must be flat regular files.
func LoadBundle(directory, manifestSHA256 string) (*Bundle, error) {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return nil, fmt.Errorf("open bundle: %w", err)
	}
	defer func() { _ = root.Close() }()

	data, err := readBundleFile(root, "manifest.json", 64*1024)
	if err != nil {
		return nil, err
	}
	if !matchesChecksum(data, manifestSHA256) {
		return nil, errors.New("manifest checksum does not match the pinned release")
	}

	var document manifest
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return nil, errors.New("manifest has trailing JSON data")
	}
	if document.Version != 1 || strings.TrimSpace(document.BundleID) == "" || len(document.Covers) != len(pilotSlugs) {
		return nil, errors.New("manifest requires version 1, a bundle ID and exactly five pilot books")
	}

	remaining := make(map[string]bool, len(pilotSlugs))
	for _, slug := range pilotSlugs {
		remaining[slug] = true
	}

	bundle := &Bundle{ID: document.BundleID}
	for _, entry := range document.Covers {
		if !remaining[entry.Slug] || !entry.IsReviewed || strings.TrimSpace(entry.UsageBasis) == "" {
			return nil, errors.New("manifest requires unique pilot slugs, reviewed artwork and usage evidence")
		}
		delete(remaining, entry.Slug)

		prepared, err := prepareEntry(root, entry)
		if err != nil {
			return nil, fmt.Errorf("validate %s: %w", entry.Slug, err)
		}
		bundle.entries = append(bundle.entries, prepared)
	}

	return bundle, nil
}

func prepareEntry(root *os.Root, entry manifestCover) (preparedCover, error) {
	main, metadata, err := inspectBundleImage(root, entry.Main)
	if err != nil {
		return preparedCover{}, err
	}

	prepared := preparedCover{
		slug:  entry.Slug,
		main:  main,
		local: books.LocalFrontCover{Main: metadata},
	}
	if entry.Thumbnail != nil {
		thumbnail, metadata, err := inspectBundleImage(root, *entry.Thumbnail)
		if err != nil {
			return preparedCover{}, err
		}

		prepared.thumbnail = thumbnail
		prepared.local.Thumbnail = &metadata
	}

	if err := prepared.local.Validate(); err != nil {
		return preparedCover{}, err
	}

	return prepared, nil
}

func inspectBundleImage(root *os.Root, file manifestFile) ([]byte, books.StoredCoverFile, error) {
	data, err := readBundleFile(root, file.File, covers.MaxFileBytes)
	if err != nil {
		return nil, books.StoredCoverFile{}, err
	}
	if !matchesChecksum(data, file.SHA256) {
		return nil, books.StoredCoverFile{}, errors.New("bundle image checksum mismatch")
	}

	metadata, err := covers.Inspect(data, file.File, "")
	if err != nil {
		return nil, books.StoredCoverFile{}, err
	}

	return data, storedFile(metadata), nil
}

func readBundleFile(root *os.Root, name string, limit int64) ([]byte, error) {
	if name == "" || name == "." || filepath.Base(name) != name || strings.ContainsAny(name, `/\`) {
		return nil, errors.New("bundle filenames must be flat relative names")
	}

	info, err := root.Lstat(name)
	if err != nil {
		return nil, fmt.Errorf("stat bundle file: %w", err)
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, errors.New("bundle file must be regular and within size limits")
	}

	file, err := root.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open bundle file: %w", err)
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("bundle file exceeds size limit")
	}

	return data, nil
}

func matchesChecksum(data []byte, checksum string) bool {
	digest := sha256.Sum256(data)
	return checksum == hex.EncodeToString(digest[:])
}

func storedFile(file covers.File) books.StoredCoverFile {
	return books.StoredCoverFile{
		StorageKey:     file.Key,
		MIMEType:       file.MIMEType,
		ByteSize:       file.ByteSize,
		Width:          file.Width,
		Height:         file.Height,
		ChecksumSHA256: file.ChecksumSHA256,
	}
}
