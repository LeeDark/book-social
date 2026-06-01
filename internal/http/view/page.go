package view

type Page struct {
	Title       string
	Description string
	ActiveNav   string
	Nav         []NavItem
	Breadcrumbs []Breadcrumb
	//CurrentUser *CurrentUser
	//Flash []FlashMessage
}

type Breadcrumb struct {
	Label string
	Href  string
}

type CurrentUser struct {
	ID          int64
	DisplayName string
}

type FlashMessage struct {
	Type string
	Text string
}
