package app

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/LeeDark/book-social/internal/modules/books"
	"github.com/LeeDark/book-social/internal/storage/covers"
)

func uploadPNG(t *testing.T, width, height int) []byte {
	t.Helper()

	var data bytes.Buffer
	if err := png.Encode(&data, image.NewRGBA(image.Rect(0, 0, width, height))); err != nil {
		t.Fatal(err)
	}

	return data.Bytes()
}

func TestCoverPreparerValidatesPairBeforePublishing(t *testing.T) {
	for _, invalid := range []string{"main", "thumbnail", "proportions", "same file"} {
		t.Run(invalid, func(t *testing.T) {
			directory := t.TempDir()
			store, err := covers.New(directory)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = store.Close() })

			main, thumbnail := uploadPNG(t, 4, 6), uploadPNG(t, 2, 3)
			switch invalid {
			case "main":
				main = []byte("invalid")
			case "thumbnail":
				thumbnail = []byte("invalid")
			case "proportions":
				thumbnail = uploadPNG(t, 2, 4)
			case "same file":
				thumbnail = main
			}

			_, err = NewCoverPreparer(store).ValidateAndStore(context.Background(),
				books.CoverUploadFile{Reader: bytes.NewReader(main), Filename: "main.png", MIMEType: "image/png"},
				&books.CoverUploadFile{Reader: bytes.NewReader(thumbnail), Filename: "thumb.png", MIMEType: "image/png"})

			var validation books.CatalogValidationError
			if !errors.As(err, &validation) {
				t.Fatalf("ValidateAndStore() error = %v", err)
			}

			entries, err := os.ReadDir(directory)
			if err != nil || len(entries) != 0 {
				t.Fatalf("invalid pair published files: %v, %v", entries, err)
			}
		})
	}
}

func TestCoverPreparerPublishesVerifiedMetadata(t *testing.T) {
	store, err := covers.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })

	main, thumbnail := uploadPNG(t, 4, 6), uploadPNG(t, 2, 3)
	prepared, err := NewCoverPreparer(store).ValidateAndStore(context.Background(),
		books.CoverUploadFile{Reader: bytes.NewReader(main), Filename: "main.png", MIMEType: "image/png"},
		&books.CoverUploadFile{Reader: bytes.NewReader(thumbnail), Filename: "thumb.png", MIMEType: "image/png"})
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range []books.StoredCoverFile{prepared.Main, *prepared.Thumbnail} {
		data, metadata, err := store.Read(file.StorageKey)
		if err != nil || int64(len(data)) != file.ByteSize || storedCoverFile(metadata) != file {
			t.Fatalf("stored file = %+v, %v", metadata, err)
		}
	}

	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := NewCoverPreparer(store).ValidateAndStore(context.Background(),
		books.CoverUploadFile{Reader: bytes.NewReader(main), Filename: "main.png"}, nil); err == nil {
		t.Fatal("closed store accepted upload")
	}
}
