package covers

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"mime"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/webp"
)

const (
	MaxFileBytes = 5 * 1024 * 1024
	MaxDimension = 4096
	MaxPixels    = 12_000_000
)

var ErrInvalidImage = errors.New("invalid cover image")

type File struct {
	Key            string
	MIMEType       string
	ByteSize       int64
	Width          int
	Height         int
	ChecksumSHA256 string
}

func (f File) URL() string {
	return "/media/covers/" + f.Key
}

func inspectImage(data []byte, filename, declaredType string) (File, error) {
	invalid := func(reason string) (File, error) {
		return File{}, fmt.Errorf("%w: %s", ErrInvalidImage, reason)
	}
	if len(data) == 0 || len(data) > MaxFileBytes {
		return invalid("file size exceeds limits")
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return invalid("unrecognized image header")
	}
	if config.Width <= 0 || config.Height <= config.Width ||
		config.Width > MaxDimension || config.Height > MaxDimension ||
		int64(config.Width)*int64(config.Height) > MaxPixels {
		return invalid("image must be portrait and within dimension limits")
	}

	var extension, contentType string
	switch format {
	case "jpeg":
		extension, contentType = ".jpg", "image/jpeg"
		if !bytes.HasSuffix(data, []byte{0xff, 0xd9}) {
			return invalid("incomplete JPEG")
		}
	case "png":
		extension, contentType = ".png", "image/png"
		if err := validatePNGChunks(data); err != nil {
			return invalid("invalid or animated PNG")
		}
	case "webp":
		extension, contentType = ".webp", "image/webp"
		if err := validateWebPChunks(data); err != nil {
			return invalid("invalid or animated WebP")
		}
	default:
		return invalid("unsupported image format")
	}
	nameExtension := strings.ToLower(filepath.Ext(filename))
	if format == "jpeg" && nameExtension == ".jpeg" {
		nameExtension = ".jpg"
	}
	if nameExtension != extension {
		return invalid("filename extension does not match image")
	}
	if declaredType != "" {
		mediaType, _, err := mime.ParseMediaType(declaredType)
		if err != nil || mediaType != contentType {
			return invalid("declared content type does not match image")
		}
	}

	decoded, decodedFormat, err := image.Decode(bytes.NewReader(data))
	if err != nil || decodedFormat != format {
		return invalid("corrupt image data")
	}
	if decoded.Bounds().Dx() != config.Width || decoded.Bounds().Dy() != config.Height {
		return invalid("inconsistent image dimensions")
	}
	checksum := sha256.Sum256(data)
	digest := hex.EncodeToString(checksum[:])
	return File{
		Key:            digest + extension,
		MIMEType:       contentType,
		ByteSize:       int64(len(data)),
		Width:          config.Width,
		Height:         config.Height,
		ChecksumSHA256: digest,
	}, nil
}

// Standard decoders can return the first frame while ignoring animation chunks.
// Walk the bounded containers as well as fully decoding the image.
func validatePNGChunks(data []byte) error {
	for offset := 8; offset+12 <= len(data); {
		length := uint64(binary.BigEndian.Uint32(data[offset : offset+4]))
		end := uint64(offset) + 12 + length
		if end > uint64(len(data)) {
			return ErrInvalidImage
		}
		switch string(data[offset+4 : offset+8]) {
		case "acTL", "fcTL", "fdAT":
			return ErrInvalidImage
		case "IEND":
			if length == 0 && end == uint64(len(data)) {
				return nil
			}
			return ErrInvalidImage
		}
		offset = int(end)
	}
	return ErrInvalidImage
}

func validateWebPChunks(data []byte) error {
	if len(data) < 12 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WEBP" ||
		uint64(binary.LittleEndian.Uint32(data[4:8]))+8 != uint64(len(data)) {
		return ErrInvalidImage
	}
	frames := 0
	for offset := 12; offset < len(data); {
		if offset+8 > len(data) {
			return ErrInvalidImage
		}
		length := uint64(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		end := uint64(offset) + 8 + length + length%2
		if end > uint64(len(data)) {
			return ErrInvalidImage
		}
		switch string(data[offset : offset+4]) {
		case "ANIM", "ANMF":
			return ErrInvalidImage
		case "VP8X":
			if offset != 12 || length != 10 || data[offset+8]&2 != 0 {
				return ErrInvalidImage
			}
		case "VP8 ", "VP8L":
			frames++
		}
		offset = int(end)
	}
	if frames != 1 {
		return ErrInvalidImage
	}
	return nil
}
