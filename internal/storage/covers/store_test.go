package covers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"hash/crc32"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// A generated solid-color 2×3 WebP, not artwork or a development-seed dependency.
const portraitWebP = "524946463600000057454250565038202a000000b001009d012a0200030002003425a00274ba0004668000feee767ff70667d5d8e1fdcdc5b9fa14000000"

func TestPutFormatsAndReopen(t *testing.T) {
	for _, format := range []string{"png", "jpeg", "webp"} {
		t.Run(format, func(t *testing.T) {
			directory := t.TempDir()
			store, err := New(directory)
			if err != nil {
				t.Fatal(err)
			}
			data := imageData(t, format)
			file, err := store.Put(context.Background(), bytes.NewReader(data), "../../untrusted."+format, "image/"+format)
			if err != nil {
				t.Fatal(err)
			}
			_ = store.Close()
			reopened := newStore(t, directory)
			got, metadata, err := reopened.Read(file.Key)
			if err != nil || !bytes.Equal(got, data) || metadata != file {
				t.Fatalf("Read() = %v, %+v, %v", got, metadata, err)
			}
			checksum := sha256.Sum256(data)
			if file.ChecksumSHA256 != hex.EncodeToString(checksum[:]) || file.Width != 2 || file.Height != 3 ||
				file.ByteSize != int64(len(data)) || file.MIMEType != "image/"+format ||
				file.URL() != "/media/covers/"+file.Key {
				t.Fatalf("unexpected metadata: %+v", file)
			}
			assertOnlyFile(t, directory, file.Key)
		})
	}
}

func TestPutRejectsInvalidImagesWithoutWriting(t *testing.T) {
	validPNG := imageData(t, "png")
	validJPEG := imageData(t, "jpeg")
	validWebP := imageData(t, "webp")
	badCRC := bytes.Clone(validPNG)
	badCRC[len(badCRC)-1] ^= 1
	oversizedDimensions := bytes.Clone(validPNG)
	binary.BigEndian.PutUint32(oversizedDimensions[16:20], MaxDimension+1)
	binary.BigEndian.PutUint32(oversizedDimensions[20:24], MaxDimension+2)
	binary.BigEndian.PutUint32(oversizedDimensions[29:33], crc32.ChecksumIEEE(oversizedDimensions[12:29]))
	tooManyPixels := bytes.Clone(oversizedDimensions)
	binary.BigEndian.PutUint32(tooManyPixels[16:20], 3500)
	binary.BigEndian.PutUint32(tooManyPixels[20:24], 4000)
	binary.BigEndian.PutUint32(tooManyPixels[29:33], crc32.ChecksumIEEE(tooManyPixels[12:29]))
	animatedPNG := append(bytes.Clone(validPNG[:33]), pngChunk("acTL", make([]byte, 8))...)
	animatedPNG = append(animatedPNG, validPNG[33:]...)
	animatedWebP := append(bytes.Clone(validWebP), []byte("ANIM\x06\x00\x00\x00\x00\x00\x00\x00\x00\x00")...)
	binary.LittleEndian.PutUint32(animatedWebP[4:8], uint32(len(animatedWebP)-8))
	landscape := new(bytes.Buffer)
	if err := png.Encode(landscape, image.NewRGBA(image.Rect(0, 0, 3, 2))); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, filename, mime string
		data                 []byte
	}{
		{name: "empty", filename: "a.png"},
		{name: "SVG", filename: "a.svg", data: []byte(`<svg xmlns="http://www.w3.org/2000/svg"/>`)},
		{name: "too many bytes", filename: "a.png", data: make([]byte, MaxFileBytes+1)},
		{name: "extension mismatch", filename: "a.jpg", data: validPNG},
		{name: "MIME mismatch", filename: "a.png", mime: "image/jpeg", data: validPNG},
		{name: "corrupt PNG", filename: "a.png", data: badCRC},
		{name: "truncated PNG", filename: "a.png", data: validPNG[:len(validPNG)-4]},
		{name: "truncated JPEG", filename: "a.jpg", data: validJPEG[:len(validJPEG)-2]},
		{name: "truncated WebP", filename: "a.webp", data: validWebP[:len(validWebP)-2]},
		{name: "dimensions", filename: "a.png", data: oversizedDimensions},
		{name: "pixel count", filename: "a.png", data: tooManyPixels},
		{name: "landscape", filename: "a.png", data: landscape.Bytes()},
		{name: "animated PNG", filename: "a.png", data: animatedPNG},
		{name: "animated WebP", filename: "a.webp", data: animatedWebP},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			directory := t.TempDir()
			store := newStore(t, directory)
			_, err := store.Put(context.Background(), bytes.NewReader(tt.data), tt.filename, tt.mime)
			if !errors.Is(err, ErrInvalidImage) {
				t.Fatalf("Put() error = %v, want ErrInvalidImage", err)
			}
			assertOnlyFile(t, directory, "")
		})
	}
}

func TestPutConcurrentIdenticalFiles(t *testing.T) {
	directory := t.TempDir()
	store := newStore(t, directory)
	data := imageData(t, "png")
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			if _, err := store.Put(context.Background(), bytes.NewReader(data), "a.png", ""); err != nil {
				t.Errorf("concurrent Put(): %v", err)
			}
		})
	}
	wg.Wait()
	file, err := store.Put(context.Background(), bytes.NewReader(data), "a.png", "")
	if err != nil {
		t.Fatal(err)
	}
	assertOnlyFile(t, directory, file.Key)
}

func TestPutDoesNotOverwriteCorruptExistingFile(t *testing.T) {
	directory := t.TempDir()
	store := newStore(t, directory)
	data := imageData(t, "png")
	file, err := store.Put(context.Background(), bytes.NewReader(data), "a.png", "")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, file.Key)
	// This is still a decodable PNG, but its bytes no longer match the key.
	changed := append(bytes.Clone(data[:33]), pngChunk("tEXt", []byte("Comment\x00modified"))...)
	changed = append(changed, data[33:]...)
	if err := os.WriteFile(path, changed, 0640); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Put(context.Background(), bytes.NewReader(data), "a.png", ""); !errors.Is(err, ErrCorruptFile) {
		t.Fatalf("Put() error = %v, want ErrCorruptFile", err)
	}
	if _, _, err := store.Read(file.Key); !errors.Is(err, ErrCorruptFile) {
		t.Fatalf("Read() error = %v, want ErrCorruptFile", err)
	}
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, changed) {
		t.Fatalf("existing file changed: %q, %v", got, err)
	}
	assertOnlyFile(t, directory, file.Key)
}

func TestReadRejectsUnsafeKeysAndSymlinks(t *testing.T) {
	directory := t.TempDir()
	store := newStore(t, directory)
	for _, key := range []string{"", ".", "../secret.png", "/etc/passwd", `..\secret`, strings.Repeat("a", 64) + ".svg", ".upload-test"} {
		if _, _, err := store.Read(key); !errors.Is(err, ErrInvalidKey) {
			t.Errorf("Read(%q) error = %v, want ErrInvalidKey", key, err)
		}
	}
	data := imageData(t, "png")
	file, err := inspectImage(data, "a.png", "")
	if err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside.png")
	if err := os.WriteFile(outside, data, 0640); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(directory, file.Key)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.Read(file.Key); err == nil {
		t.Fatal("Read() accepted symlink")
	}
	if _, err := store.Put(context.Background(), bytes.NewReader(data), "a.png", ""); err == nil {
		t.Fatal("Put() reused symlink")
	}
}

func TestPutInputAndStorageFailures(t *testing.T) {
	directory := t.TempDir()
	store := newStore(t, directory)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := store.Put(ctx, bytes.NewReader(imageData(t, "png")), "a.png", ""); !errors.Is(err, context.Canceled) {
		t.Fatalf("Put() error = %v, want cancellation", err)
	}
	inputErr := errors.New("broken input stream")
	if _, err := store.Put(context.Background(), failedReader{inputErr}, "a.png", ""); !errors.Is(err, inputErr) {
		t.Fatalf("Put() error = %v, want input error", err)
	}
	assertOnlyFile(t, directory, "")
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Put(context.Background(), bytes.NewReader(imageData(t, "png")), "a.png", ""); err == nil {
		t.Fatal("Put() succeeded with closed storage")
	}
	if _, err := New(""); err == nil {
		t.Fatal("New() accepted empty directory")
	}
	blocked := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocked, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := New(filepath.Join(blocked, "covers")); err == nil {
		t.Fatal("New() accepted directory below regular file")
	}
}

func newStore(t *testing.T, directory string) *Store {
	t.Helper()
	store, err := New(directory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func imageData(t *testing.T, format string) []byte {
	t.Helper()
	if format == "webp" {
		data, err := hex.DecodeString(portraitWebP)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	var data bytes.Buffer
	var err error
	img := image.NewRGBA(image.Rect(0, 0, 2, 3))
	if format == "jpeg" {
		err = jpeg.Encode(&data, img, nil)
	} else {
		err = png.Encode(&data, img)
	}
	if err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}

func pngChunk(kind string, payload []byte) []byte {
	chunk := make([]byte, len(payload)+12)
	binary.BigEndian.PutUint32(chunk[:4], uint32(len(payload)))
	copy(chunk[4:8], kind)
	copy(chunk[8:], payload)
	binary.BigEndian.PutUint32(chunk[len(chunk)-4:], crc32.ChecksumIEEE(chunk[4:len(chunk)-4]))
	return chunk
}

func assertOnlyFile(t *testing.T, directory, key string) {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if key == "" && len(entries) == 0 {
		return
	}
	if len(entries) != 1 || entries[0].Name() != key {
		t.Fatalf("storage entries = %v, want only %q", entries, key)
	}
}

type failedReader struct{ err error }

func (r failedReader) Read([]byte) (int, error) { return 0, r.err }

var _ io.Reader = failedReader{}
