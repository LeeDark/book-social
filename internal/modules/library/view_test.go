package library

import (
	"testing"

	"github.com/LeeDark/book-social/internal/modules/books"
)

func TestLibraryViewUsesCardCover(t *testing.T) {
	for _, hasThumbnail := range []bool{false, true} {
		name := "main only"
		if hasThumbnail {
			name = "thumbnail preferred"
		}

		t.Run(name, func(t *testing.T) {
			cover := books.Cover{Variant: "front", URL: "/media/covers/main.png"}
			wantURL := cover.URL
			if hasThumbnail {
				cover.Thumbnail = &books.StoredCoverFile{StorageKey: "thumb.png"}
				wantURL = cover.Thumbnail.URL()
			}

			items := mapItemsToViews([]Item{{Book: books.Book{ID: 7, Covers: []books.Cover{cover}}}})
			if items[0].FrontCover == nil || items[0].FrontCover.URL != wantURL || items[0].CoverClass != "cover-2" {
				t.Fatalf("library cover = %+v, class %q", items[0].FrontCover, items[0].CoverClass)
			}
		})
	}
}
