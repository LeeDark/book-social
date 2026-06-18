package books

import "testing"

func TestMapBooksToCards(t *testing.T) {
	books := []Book{
		{
			ID:          7,
			Title:       "Signal in the Stacks",
			Slug:        "signal-in-the-stacks",
			Description: "A library mystery.",
			Author: Author{
				ID:         2,
				FirstName:  "Jon",
				SecondName: "A.",
				SurName:    "Vale",
			},
			Genre: Genre{
				Name: "Mystery",
				Slug: "mystery",
			},
		},
	}

	cards := mapBooksToCards(books)

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
	if card.AuthorName != "Jon A. Vale" {
		t.Errorf("AuthorName = %q", card.AuthorName)
	}
	if card.AuthorURL != "/authors/2" {
		t.Errorf("AuthorURL = %q", card.AuthorURL)
	}
	if card.GenreName != "Mystery" {
		t.Errorf("GenreName = %q", card.GenreName)
	}
	if card.GenreURL != "/books?genre=mystery" {
		t.Errorf("GenreURL = %q", card.GenreURL)
	}
	if card.CoverClass != "cover-2" {
		t.Errorf("CoverClass = %q", card.CoverClass)
	}
}

func TestMapBookToDetailsView(t *testing.T) {
	book := Book{
		ID:          8,
		Title:       "The Quiet Atlas",
		Slug:        "the-quiet-atlas",
		Description: "A reflective journey.",
		Author: Author{
			ID:         1,
			FirstName:  "Mira",
			SecondName: "L.",
			SurName:    "Stone",
		},
		Genre: Genre{
			Name: "Literary Fiction",
			Slug: "literary-fiction",
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
	if details.Authors[0].Name != "Mira L. Stone" {
		t.Errorf("Author name = %q", details.Authors[0].Name)
	}
	if details.Authors[0].URL != "/authors/1" {
		t.Errorf("Author URL = %q", details.Authors[0].URL)
	}

	if got, want := len(details.Genres), 1; got != want {
		t.Fatalf("len(Genres) = %d, want %d", got, want)
	}
	if details.Genres[0].Name != "Literary Fiction" {
		t.Errorf("Genre name = %q", details.Genres[0].Name)
	}
	if details.Genres[0].URL != "/books?genre=literary-fiction" {
		t.Errorf("Genre URL = %q", details.Genres[0].URL)
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
