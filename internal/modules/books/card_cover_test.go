package books

import "testing"

func TestSelectCardFrontCover(t *testing.T) {
	width, height := 600, 900
	thumbnail := &StoredCoverFile{StorageKey: "thumbnail.webp", Width: 240, Height: 360}
	main := Cover{
		Variant:   "front",
		URL:       "/media/covers/main.webp",
		Width:     &width,
		Height:    &height,
		Thumbnail: thumbnail,
	}
	back := Cover{Variant: "back", URL: "https://example.test/back.jpg"}
	mainOnly := main
	mainOnly.Thumbnail = nil

	cases := []struct {
		name   string
		covers []Cover
		url    string
		width  int
	}{
		{name: "no cover"},
		{name: "back only", covers: []Cover{back}},
		{name: "empty front URL", covers: []Cover{{Variant: "front"}}},
		{name: "external main", covers: []Cover{{Variant: "front", URL: "https://example.test/front.jpg"}}, url: "https://example.test/front.jpg"},
		{name: "thumbnail preferred after back variant", covers: []Cover{back, main}, url: thumbnail.URL(), width: 240},
		{name: "local main without thumbnail", covers: []Cover{mainOnly}, url: mainOnly.URL, width: 600},
	}

	for _, check := range cases {
		t.Run(check.name, func(t *testing.T) {
			cover := SelectCardFrontCover(check.covers)
			if check.url == "" {
				if cover != nil {
					t.Fatalf("SelectCardFrontCover() = %+v, want nil", cover)
				}

				return
			}

			if cover == nil || cover.URL != check.url {
				t.Fatalf("SelectCardFrontCover() = %+v, want URL %q", cover, check.url)
			}
			if check.width != 0 && (cover.Width == nil || *cover.Width != check.width) {
				t.Fatalf("cover width = %v, want %d", cover.Width, check.width)
			}
		})
	}
}

func TestDetailCoverUsesMainDespiteThumbnail(t *testing.T) {
	book := Book{Covers: []Cover{{
		Variant:   "front",
		URL:       "/media/covers/main.webp",
		Thumbnail: &StoredCoverFile{StorageKey: "thumbnail.webp"},
	}}}

	details := mapBookToDetailsView(book)
	cards := mapBooksToCards([]Book{book})

	if details.FrontCover == nil || details.FrontCover.URL != book.Covers[0].URL {
		t.Fatalf("detail front cover = %+v, want main", details.FrontCover)
	}
	if cards[0].FrontCover == nil || cards[0].FrontCover.URL != book.Covers[0].Thumbnail.URL() {
		t.Fatalf("card front cover = %+v, want thumbnail", cards[0].FrontCover)
	}
}
