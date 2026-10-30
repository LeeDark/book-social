package covers

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strings"
)

var (
	ErrInvalidKey  = errors.New("invalid cover storage key")
	ErrCorruptFile = errors.New("stored cover does not match its key")
	keyPattern     = regexp.MustCompile(`^[a-f0-9]{64}\.(jpg|png|webp)$`)
)

type Store struct {
	root *os.Root
}

func New(directory string) (*Store, error) {
	if strings.TrimSpace(directory) == "" {
		return nil, errors.New("cover storage directory is empty")
	}

	if err := os.MkdirAll(directory, 0750); err != nil {
		return nil, fmt.Errorf("create cover directory: %w", err)
	}

	root, err := os.OpenRoot(directory)
	if err != nil {
		return nil, fmt.Errorf("open cover directory: %w", err)
	}

	return &Store{root: root}, nil
}

func (s *Store) Close() error {
	return s.root.Close()
}

// Put publishes validated bytes without replacing an existing key. The caller owns
// the later SQL transaction; a published file is retained if that transaction fails.
func (s *Store) Put(ctx context.Context, source io.Reader, filename, declaredType string) (File, error) {
	data, err := readCoverInput(ctx, source)
	if err != nil {
		return File{}, fmt.Errorf("read cover input: %w", err)
	}

	file, err := inspectImage(data, filename, declaredType)
	if err != nil {
		return File{}, err
	}
	if err := ctx.Err(); err != nil {
		return File{}, err
	}

	temporary := ".upload-" + rand.Text()
	f, err := s.root.OpenFile(temporary, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0640)
	if err != nil {
		return File{}, fmt.Errorf("create temporary cover: %w", err)
	}
	defer func() { _ = s.root.Remove(temporary) }()

	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return File{}, fmt.Errorf("write cover: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return File{}, fmt.Errorf("sync cover: %w", err)
	}
	if err := f.Close(); err != nil {
		return File{}, fmt.Errorf("close cover: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return File{}, err
	}

	// A hard link publishes a complete file atomically and refuses to overwrite.
	// Both names are inside the same root/filesystem. Rename would overwrite on Unix.
	if err := s.root.Link(temporary, file.Key); err != nil {
		if !errors.Is(err, fs.ErrExist) {
			return File{}, fmt.Errorf("publish cover: %w", err)
		}

		existing, err := s.read(file.Key)
		if err != nil {
			return File{}, fmt.Errorf("verify existing cover: %w", err)
		}
		if !bytes.Equal(existing, data) {
			return File{}, ErrCorruptFile
		}
	}

	if err := s.syncDirectory(); err != nil {
		return File{}, err
	}

	return file, nil
}

// Read verifies bytes before serving, so missing, damaged or mismatched files do
// not acquire a public immutable cache entry. Reads are bounded by MaxFileBytes.
func (s *Store) Read(key string) ([]byte, File, error) {
	data, err := s.read(key)
	if err != nil {
		return nil, File{}, err
	}

	file, err := inspectImage(data, key, "")
	if err != nil || file.Key != key {
		return nil, File{}, ErrCorruptFile
	}

	return data, file, nil
}

func (s *Store) read(key string) ([]byte, error) {
	if !keyPattern.MatchString(key) {
		return nil, ErrInvalidKey
	}

	info, err := s.root.Lstat(key)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > MaxFileBytes {
		return nil, ErrCorruptFile
	}

	f, err := s.root.Open(key)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	return io.ReadAll(io.LimitReader(f, MaxFileBytes+1))
}

func (s *Store) syncDirectory() error {
	directory, err := s.root.Open(".")
	if err != nil {
		return fmt.Errorf("open cover directory for sync: %w", err)
	}
	defer func() { _ = directory.Close() }()

	if err := directory.Sync(); err != nil {
		return fmt.Errorf("sync cover directory: %w", err)
	}

	return nil
}

func readCoverInput(ctx context.Context, source io.Reader) ([]byte, error) {
	const bufferSize = 32 * 1024
	buffer := make([]byte, bufferSize)
	data := make([]byte, 0, bufferSize)

	// Read one byte beyond the accepted size so inspection can reject oversized
	// files. Check cancellation before each source read without retaining context.
	for len(data) < MaxFileBytes+1 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		remaining := MaxFileBytes + 1 - len(data)
		n, err := source.Read(buffer[:min(len(buffer), remaining)])
		data = append(data, buffer[:n]...)
		if err == io.EOF {
			return data, nil
		}
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}
