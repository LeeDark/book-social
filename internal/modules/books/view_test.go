package books

import "testing"

func TestMapBooksToCardsMapsAllRelationships(t *testing.T) {
	bookList := []Book{
		{
			ID:          7,
			Title:       "Signal in the Stacks",
			Slug:        "signal-in-the-stacks",
			Description: "A library mystery.",
			Authors: []Author{
				{ID: 2, FirstName: "Jon", SecondName: "A.", SurName: "Vale", Slug: "jon-a-vale"},
				{ID: 3, FirstName: "Ada", SurName: "Kern", Slug: "ada-kern"},
			},
			Genres: []Genre{
				{ID: 1, Name: "Mystery", Slug: "mystery"},
				{ID: 2, Name: "Thriller", Slug: "thriller"},
			},
		},
	}

	cards := mapBooksToCards(bookList)
	if got, want := len(cards), 1; got != want {
		t.Fatalf("len(cards) = %d, want %d", got, want)
	}

	card := cards[0]
	if card.Title != "Signal in the Stacks" {
		t.Errorf("Title = %q", card.Title)
	}
	if card.Slug != "signal-in-the-stacks" {
		t.Errorf("Slug = %q", card.Slug)
	}
	if card.BookURL != "/books/signal-in-the-stacks" {
		t.Errorf("BookURL = %q", card.BookURL)
	}
	if card.CoverClass != "cover-2" {
		t.Errorf("CoverClass = %q", card.CoverClass)
	}

	if got, want := len(card.Authors), 2; got != want {
		t.Fatalf("len(Authors) = %d, want %d", got, want)
	}
	if got, want := card.Authors[0].Name, "Jon A. Vale"; got != want {
		t.Errorf("first author name = %q, want %q", got, want)
	}
	if got, want := card.Authors[0].URL, "/authors/jon-a-vale"; got != want {
		t.Errorf("first author URL = %q, want %q", got, want)
	}
	if got, want := card.Authors[0].FilterURL, "/books?author=jon-a-vale"; got != want {
		t.Errorf("first author filter URL = %q, want %q", got, want)
	}
	if got, want := card.Authors[1].Name, "Ada Kern"; got != want {
		t.Errorf("second author name = %q, want %q", got, want)
	}

	if got, want := len(card.Genres), 2; got != want {
		t.Fatalf("len(Genres) = %d, want %d", got, want)
	}
	if got, want := card.Genres[0].Name, "Mystery"; got != want {
		t.Errorf("first genre name = %q, want %q", got, want)
	}
	if got, want := card.Genres[1].URL, "/books?genre=thriller"; got != want {
		t.Errorf("second genre URL = %q, want %q", got, want)
	}
}

func TestMapBookToDetailsViewMapsRelationshipsAndCovers(t *testing.T) {
	mimeType := "image/jpeg"
	byteSize := int64(245760)
	width := 600
	height := 900
	checksum := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	book := Book{
		ID:          8,
		Title:       "The Quiet Atlas",
		Slug:        "the-quiet-atlas",
		Description: "A reflective journey.",
		Authors: []Author{
			{ID: 1, FirstName: "Mira", SecondName: "L.", SurName: "Stone", Slug: "mira-l-stone"},
		},
		Genres: []Genre{
			{ID: 4, Name: "Literary Fiction", Slug: "literary-fiction"},
		},
		Covers: []Cover{
			{
				ID:             5,
				Variant:        "front",
				URL:            "https://example.test/covers/the-quiet-atlas.jpg",
				MIMEType:       &mimeType,
				ByteSize:       &byteSize,
				Width:          &width,
				Height:         &height,
				ChecksumSHA256: &checksum,
			},
		},
	}

	details := mapBookToDetailsView(book)
	if details.ID != 8 {
		t.Errorf("ID = %d", details.ID)
	}
	if details.Title != "The Quiet Atlas" {
		t.Errorf("Title = %q", details.Title)
	}
	if details.Slug != "the-quiet-atlas" {
		t.Errorf("Slug = %q", details.Slug)
	}
	if details.Description != "A reflective journey." {
		t.Errorf("Description = %q", details.Description)
	}
	if details.CoverClass != "cover-3" {
		t.Errorf("CoverClass = %q", details.CoverClass)
	}

	if got, want := len(details.Authors), 1; got != want {
		t.Fatalf("len(Authors) = %d, want %d", got, want)
	}
	if got, want := details.Authors[0].Name, "Mira L. Stone"; got != want {
		t.Errorf("author name = %q, want %q", got, want)
	}
	if got, want := len(details.Genres), 1; got != want {
		t.Fatalf("len(Genres) = %d, want %d", got, want)
	}
	if got, want := details.Genres[0].URL, "/books?genre=literary-fiction"; got != want {
		t.Errorf("genre URL = %q, want %q", got, want)
	}
	if got, want := len(details.Covers), 1; got != want {
		t.Fatalf("len(Covers) = %d, want %d", got, want)
	}
	if got, want := details.Covers[0].Variant, "front"; got != want {
		t.Errorf("cover variant = %q, want %q", got, want)
	}
	if got, want := details.Covers[0].URL, "https://example.test/covers/the-quiet-atlas.jpg"; got != want {
		t.Errorf("cover URL = %q, want %q", got, want)
	}
	if details.FrontCover == nil || details.FrontCover.URL != "https://example.test/covers/the-quiet-atlas.jpg" {
		t.Errorf("FrontCover = %#v", details.FrontCover)
	}
	if details.Covers[0].MIMEType != &mimeType {
		t.Errorf("cover MIMEType pointer was not preserved")
	}
	if details.Covers[0].ByteSize != &byteSize || details.Covers[0].Width != &width || details.Covers[0].Height != &height {
		t.Errorf("cover numeric metadata pointers were not preserved")
	}
	if details.Covers[0].ChecksumSHA256 != &checksum {
		t.Errorf("cover checksum pointer was not preserved")
	}
}

func TestMapBookToDetailsViewHandlesMissingRelationships(t *testing.T) {
	details := mapBookToDetailsView(Book{ID: 1, Title: "Unlinked Book", Slug: "unlinked-book"})

	if len(details.Authors) != 0 || len(details.Genres) != 0 || len(details.Covers) != 0 || details.FrontCover != nil {
		t.Fatalf(
			"relationship lengths = %d/%d/%d, want 0/0/0",
			len(details.Authors),
			len(details.Genres),
			len(details.Covers),
		)
	}
}

func TestMapBookToDetailsViewSelectsFrontCover(t *testing.T) {
	details := mapBookToDetailsView(Book{
		Covers: []Cover{
			{Variant: "back", URL: "https://example.test/covers/back.jpg"},
			{Variant: "front", URL: "https://example.test/covers/front.jpg"},
		},
	})

	if details.FrontCover == nil {
		t.Fatal("FrontCover = nil")
	}
	if details.FrontCover.URL != "https://example.test/covers/front.jpg" {
		t.Errorf("FrontCover.URL = %q", details.FrontCover.URL)
	}
}

func TestCoverClassForBook(t *testing.T) {
	tests := []struct {
		name string
		id   int
		want string
	}{
		{name: "zero", id: 0, want: "cover-0"},
		{name: "within range", id: 4, want: "cover-4"},
		{name: "wraps after five", id: 5, want: "cover-0"},
		{name: "wraps larger id", id: 12, want: "cover-2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := coverClassForBook(tt.id)
			if got != tt.want {
				t.Fatalf("coverClassForBook(%d) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}
