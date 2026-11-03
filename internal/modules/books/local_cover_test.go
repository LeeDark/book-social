package books

import (
	"errors"
	"strings"
	"testing"
)

func TestLocalFrontCoverValidation(t *testing.T) {
	fixture := func() LocalFrontCover {
		digest := strings.Repeat("a", 64)
		main := StoredCoverFile{
			StorageKey: digest + ".png", ChecksumSHA256: digest, MIMEType: "image/png",
			ByteSize: 1000, Width: 600, Height: 900,
		}
		thumbnail := main
		thumbnail.ChecksumSHA256 = strings.Repeat("b", 64)
		thumbnail.StorageKey = thumbnail.ChecksumSHA256 + ".png"
		thumbnail.Width, thumbnail.Height = 240, 360
		return LocalFrontCover{Main: main, Thumbnail: &thumbnail}
	}
	for _, check := range []struct {
		name   string
		change func(*LocalFrontCover)
		valid  bool
	}{
		{name: "main and thumbnail", valid: true},
		{name: "main only", valid: true, change: func(c *LocalFrontCover) { c.Thumbnail = nil }},
		{name: "JPEG", valid: true, change: func(c *LocalFrontCover) {
			c.Main.MIMEType, c.Main.StorageKey = "image/jpeg", c.Main.ChecksumSHA256+".jpg"
			c.Thumbnail = nil
		}},
		{name: "WebP", valid: true, change: func(c *LocalFrontCover) {
			c.Main.MIMEType, c.Main.StorageKey = "image/webp", c.Main.ChecksumSHA256+".webp"
			c.Thumbnail = nil
		}},
		{name: "unsafe key", change: func(c *LocalFrontCover) { c.Main.StorageKey = "../cover.png" }},
		{name: "wrong checksum", change: func(c *LocalFrontCover) { c.Main.ChecksumSHA256 = strings.Repeat("b", 64) }},
		{name: "unsupported MIME", change: func(c *LocalFrontCover) { c.Main.MIMEType = "image/svg+xml" }},
		{name: "empty bytes", change: func(c *LocalFrontCover) { c.Main.ByteSize = 0 }},
		{name: "too many bytes", change: func(c *LocalFrontCover) { c.Main.ByteSize = 5242881 }},
		{name: "zero width", change: func(c *LocalFrontCover) { c.Main.Width = 0 }},
		{name: "landscape", change: func(c *LocalFrontCover) { c.Main.Width = 1000 }},
		{name: "dimension limit", change: func(c *LocalFrontCover) { c.Main.Height = 4097 }},
		{name: "pixel limit", change: func(c *LocalFrontCover) { c.Main.Width, c.Main.Height = 3500, 4000 }},
		{name: "thumbnail format mismatch", change: func(c *LocalFrontCover) {
			c.Thumbnail.MIMEType = "image/jpeg"
			c.Thumbnail.StorageKey = c.Thumbnail.ChecksumSHA256 + ".jpg"
		}},
		{name: "thumbnail proportions", change: func(c *LocalFrontCover) { c.Thumbnail.Height = 300 }},
		{name: "thumbnail is not smaller", change: func(c *LocalFrontCover) { c.Thumbnail.Width, c.Thumbnail.Height = 600, 900 }},
		{name: "thumbnail incomplete", change: func(c *LocalFrontCover) { c.Thumbnail.StorageKey = "" }},
		{name: "thumbnail reuses main key", change: func(c *LocalFrontCover) {
			c.Thumbnail.StorageKey, c.Thumbnail.ChecksumSHA256 = c.Main.StorageKey, c.Main.ChecksumSHA256
		}},
	} {
		t.Run(check.name, func(t *testing.T) {
			cover := fixture()
			if check.change != nil {
				check.change(&cover)
			}
			err := cover.Validate()
			if check.valid {
				if err != nil {
					t.Fatal(err)
				}
				if cover.Main.URL() != "/media/covers/"+cover.Main.StorageKey {
					t.Fatal("unexpected public URL")
				}
				return
			}
			var validation CatalogValidationError
			if !errors.As(err, &validation) || validation.Field != "front_cover" {
				t.Fatalf("Validate() = %v", err)
			}
		})
	}
}
