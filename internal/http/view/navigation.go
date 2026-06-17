package view

type NavItem struct {
	Key    string
	Label  string
	Href   string
	Public bool
}

func MainNavigation() []NavItem {
	return []NavItem{
		{Key: "home", Label: "Home", Href: "/", Public: true},
		{Key: "catalog", Label: "Catalog", Href: "/books", Public: true},
		{Key: "authors", Label: "Authors", Href: "/authors", Public: true},
		{Key: "genres", Label: "Genres", Href: "/genres", Public: true},
		{Key: "about", Label: "About", Href: "/about", Public: true},
	}
}
