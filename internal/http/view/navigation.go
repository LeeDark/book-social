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
		{Key: "about", Label: "About", Href: "/about", Public: true},
		{Key: "catalog", Label: "Books", Href: "/books", Public: true},
	}
}
