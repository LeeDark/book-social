package books

type Genre struct {
	ID          int
	Name        string
	Slug        string
	Description string
}

type Author struct {
	ID          int
	FirstName   string
	SecondName  string
	SurName     string
	Slug        string
	Description string
}

type Cover struct {
	ID             int
	Variant        string
	URL            string
	MIMEType       *string
	ByteSize       *int64
	Width          *int
	Height         *int
	ChecksumSHA256 *string
}

type Book struct {
	ID          int
	Title       string
	Slug        string
	Description string
	Authors     []Author
	Genres      []Genre
	Covers      []Cover
}
