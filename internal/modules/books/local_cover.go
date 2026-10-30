package books

import "regexp"

// StoredCoverFile describes already validated and published bytes. It contains
// no host path; the upload service must verify the file before saving a reference.
type StoredCoverFile struct {
	StorageKey     string
	MIMEType       string
	ByteSize       int64
	Width          int
	Height         int
	ChecksumSHA256 string
}

func (f StoredCoverFile) URL() string {
	return "/media/covers/" + f.StorageKey
}

type LocalFrontCover struct {
	Main      StoredCoverFile
	Thumbnail *StoredCoverFile
}

var coverChecksumPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

// Validate checks reference/metadata consistency, not image bytes or file existence.
func (c LocalFrontCover) Validate() error {
	if !validStoredCoverFile(c.Main) {
		return CatalogValidationError{Field: "front_cover", Message: "Choose a valid prepared cover file."}
	}
	if thumbnail := c.Thumbnail; thumbnail != nil {
		if !validStoredCoverFile(*thumbnail) || thumbnail.MIMEType != c.Main.MIMEType ||
			thumbnail.StorageKey == c.Main.StorageKey ||
			thumbnail.Width >= c.Main.Width || thumbnail.Height >= c.Main.Height ||
			int64(thumbnail.Width)*int64(c.Main.Height) != int64(thumbnail.Height)*int64(c.Main.Width) {
			return CatalogValidationError{Field: "front_cover", Message: "Choose a smaller thumbnail with the same format and proportions."}
		}
	}
	return nil
}

func validStoredCoverFile(f StoredCoverFile) bool {
	var extension string
	switch f.MIMEType {
	case "image/jpeg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	case "image/webp":
		extension = ".webp"
	default:
		return false
	}
	return coverChecksumPattern.MatchString(f.ChecksumSHA256) &&
		f.StorageKey == f.ChecksumSHA256+extension &&
		f.ByteSize > 0 && f.ByteSize <= 5*1024*1024 &&
		f.Width > 0 && f.Height > f.Width && f.Width <= 4096 && f.Height <= 4096 &&
		int64(f.Width)*int64(f.Height) <= 12_000_000
}
