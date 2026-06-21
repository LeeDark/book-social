package books

type Genre struct {
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

type Book struct {
	ID          int
	Title       string
	Slug        string
	Description string
	Author      Author
	Genre       Genre
}
