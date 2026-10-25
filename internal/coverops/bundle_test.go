package coverops

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func writeTestBundle(t *testing.T) (string, manifest) {
	t.Helper()

	directory := t.TempDir()
	main := writeTestPNG(t, directory, "main.png", 4, 6)
	thumbnail := writeTestPNG(t, directory, "thumb.png", 2, 3)
	document := manifest{Version: 1, BundleID: "synthetic-tests-only-v1"}
	for _, slug := range pilotSlugs {
		document.Covers = append(document.Covers, manifestCover{
			Slug:       slug,
			IsReviewed: true,
			UsageBasis: "Generated solid-color test fixture, not approved demo artwork.",
			Main:       main,
			Thumbnail:  &thumbnail,
		})
	}

	return directory, document
}

func writeTestPNG(t *testing.T, directory, name string, width, height int) manifestFile {
	t.Helper()

	var data bytes.Buffer
	if err := png.Encode(&data, image.NewRGBA(image.Rect(0, 0, width, height))); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(directory, name), data.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}

	return manifestFile{File: name, SHA256: checksum(data.Bytes())}
}

func writeTestManifest(t *testing.T, directory string, document manifest) string {
	t.Helper()

	data, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "manifest.json"), data, 0600); err != nil {
		t.Fatal(err)
	}

	return checksum(data)
}

func checksum(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func TestLoadBundleValidatesAllFiveBooks(t *testing.T) {
	directory, document := writeTestBundle(t)
	digest := writeTestManifest(t, directory, document)

	bundle, err := LoadBundle(directory, digest)
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.entries) != 5 || bundle.ID != document.BundleID {
		t.Fatalf("bundle = %+v", bundle)
	}
	for _, entry := range bundle.entries {
		if entry.local.Main.Width != 4 || entry.local.Thumbnail == nil || entry.local.Thumbnail.Width != 2 {
			t.Fatalf("prepared metadata = %+v", entry.local)
		}
	}
}

func TestLoadBundleRejectsInvalidDelivery(t *testing.T) {
	cases := []struct {
		name   string
		change func(*testing.T, string, *manifest)
	}{
		{name: "unsupported version", change: func(_ *testing.T, _ string, m *manifest) { m.Version = 2 }},
		{name: "missing bundle ID", change: func(_ *testing.T, _ string, m *manifest) { m.BundleID = "" }},
		{name: "missing pilot", change: func(_ *testing.T, _ string, m *manifest) { m.Covers = m.Covers[:4] }},
		{name: "duplicate slug", change: func(_ *testing.T, _ string, m *manifest) { m.Covers[4].Slug = m.Covers[0].Slug }},
		{name: "unknown slug", change: func(_ *testing.T, _ string, m *manifest) { m.Covers[4].Slug = "unrelated-book" }},
		{name: "unreviewed artwork", change: func(_ *testing.T, _ string, m *manifest) { m.Covers[4].IsReviewed = false }},
		{name: "missing usage evidence", change: func(_ *testing.T, _ string, m *manifest) { m.Covers[4].UsageBasis = " " }},
		{name: "missing last file", change: func(_ *testing.T, _ string, m *manifest) { m.Covers[4].Main.File = "missing.png" }},
		{name: "wrong file checksum", change: func(_ *testing.T, _ string, m *manifest) { m.Covers[4].Main.SHA256 = "incorrect" }},
		{name: "traversal", change: func(_ *testing.T, _ string, m *manifest) { m.Covers[4].Main.File = "../main.png" }},
		{name: "backslash path", change: func(_ *testing.T, _ string, m *manifest) { m.Covers[4].Main.File = `..\main.png` }},
		{name: "absolute path", change: func(_ *testing.T, _ string, m *manifest) { m.Covers[4].Main.File = "/main.png" }},
		{name: "wrong thumbnail proportions", change: func(t *testing.T, dir string, m *manifest) {
			file := writeTestPNG(t, dir, "wrong.png", 2, 4)
			m.Covers[4].Thumbnail = &file
		}},
		{name: "main reused as thumbnail", change: func(_ *testing.T, _ string, m *manifest) {
			file := m.Covers[4].Main
			m.Covers[4].Thumbnail = &file
		}},
		{name: "symlink", change: func(t *testing.T, dir string, m *manifest) {
			if err := os.Symlink("main.png", filepath.Join(dir, "link.png")); err != nil {
				t.Fatal(err)
			}
			m.Covers[4].Main.File = "link.png"
		}},
	}

	for _, check := range cases {
		t.Run(check.name, func(t *testing.T) {
			directory, document := writeTestBundle(t)
			check.change(t, directory, &document)
			digest := writeTestManifest(t, directory, document)

			if _, err := LoadBundle(directory, digest); err == nil {
				t.Fatal("invalid delivery accepted")
			}
		})
	}
}

func TestLoadBundleRejectsUnpinnedOrAmbiguousManifest(t *testing.T) {
	cases := []struct {
		name string
		data string
	}{
		{name: "unknown field", data: `{"unknown":true}`},
		{name: "trailing document", data: `{} {}`},
		{name: "array instead of manifest", data: `[]`},
		{name: "oversized manifest", data: string(make([]byte, 64*1024+1))},
	}

	for _, check := range cases {
		t.Run(check.name, func(t *testing.T) {
			data := []byte(check.data)
			directory := t.TempDir()
			if err := os.WriteFile(filepath.Join(directory, "manifest.json"), data, 0600); err != nil {
				t.Fatal(err)
			}

			if _, err := LoadBundle(directory, checksum(data)); err == nil {
				t.Fatal("malformed manifest accepted")
			}
		})
	}

	directory, document := writeTestBundle(t)
	writeTestManifest(t, directory, document)
	if _, err := LoadBundle(directory, "wrong-pinned-checksum"); err == nil {
		t.Fatal("unpinned manifest accepted")
	}
}
