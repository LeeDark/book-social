package coverops

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// pinnedCatalogManifestSHA256 is the installer manifest checksum recorded in
// assets/demo-covers/catalog-v1/README.md; operators pass it to coverctl install.
const pinnedCatalogManifestSHA256 = "28329d42d6ad6db7aaffac97cd5f220a7db50f45eebd3133bef9aa1ad293973d"

var catalogInventoryDir = filepath.Join("..", "..", "assets", "demo-covers", "catalog-v1")

// catalogInventory reads only the fields that assemble.py copies into the
// installer manifest. The artwork files themselves are kept outside Git.
type catalogInventory struct {
	DatasetID string                  `json:"dataset_id"`
	Covers    []catalogInventoryCover `json:"covers"`
}

type catalogInventoryCover struct {
	Slug       string       `json:"slug"`
	ReviewedBy string       `json:"reviewed_by"`
	ReviewedOn string       `json:"reviewed_on"`
	Main       manifestFile `json:"main"`
	Thumbnail  manifestFile `json:"thumbnail"`
}

func readCatalogInventory(t *testing.T) catalogInventory {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(catalogInventoryDir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}

	var inventory catalogInventory
	if err := json.Unmarshal(data, &inventory); err != nil {
		t.Fatalf("decode catalog inventory: %v", err)
	}

	return inventory
}

// TestCatalogInventoryReproducesPinnedManifest mirrors assemble.py. A changed
// inventory must come with a newly reviewed pin in the catalog README.
func TestCatalogInventoryReproducesPinnedManifest(t *testing.T) {
	inventory := readCatalogInventory(t)

	type installerCover struct {
		Slug       string       `json:"slug"`
		IsReviewed bool         `json:"reviewed"`
		UsageBasis string       `json:"usage_basis"`
		Main       manifestFile `json:"main"`
		Thumbnail  manifestFile `json:"thumbnail"`
	}

	document := struct {
		Version  int              `json:"version"`
		BundleID string           `json:"bundle_id"`
		Covers   []installerCover `json:"covers"`
	}{Version: 2, BundleID: inventory.DatasetID}
	slugs := make(map[string]bool, len(inventory.Covers))
	files := make(map[string]bool, 2*len(inventory.Covers))
	for _, cover := range inventory.Covers {
		document.Covers = append(document.Covers, installerCover{
			Slug:       cover.Slug,
			IsReviewed: true,
			UsageBasis: "Project demo illustration, not a specific edition; reviewed by " +
				cover.ReviewedBy + " on " + cover.ReviewedOn + ". See retained provenance and review records.",
			Main:      cover.Main,
			Thumbnail: cover.Thumbnail,
		})
		slugs[cover.Slug] = true
		files[cover.Main.File] = true
		files[cover.Thumbnail.File] = true
	}

	if len(document.Covers) != 109 || len(slugs) != 109 || len(files) != 218 {
		t.Fatalf("inventory has %d entries, %d unique slugs and %d unique files; want 109, 109 and 218",
			len(document.Covers), len(slugs), len(files))
	}

	// Match Python's json.dumps(indent=2, ensure_ascii=False) plus a final newline.
	var data bytes.Buffer
	encoder := json.NewEncoder(&data)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(document); err != nil {
		t.Fatal(err)
	}

	if data.Len() > maxManifestBytes {
		t.Fatalf("installer manifest is %d bytes, above the %d-byte loader limit", data.Len(), maxManifestBytes)
	}
	if got := checksum(data.Bytes()); got != pinnedCatalogManifestSHA256 {
		t.Fatalf("installer manifest SHA-256 = %s, want pinned %s", got, pinnedCatalogManifestSHA256)
	}

	readme, err := os.ReadFile(filepath.Join(catalogInventoryDir, "README.md"))
	if err != nil || !bytes.Contains(readme, []byte(pinnedCatalogManifestSHA256)) {
		t.Fatalf("catalog README does not record the pinned manifest checksum: %v", err)
	}
}
