package view

type Page struct {
	Title       string
	Description string
	ActiveNav   string
	Nav         []NavItem
	//CurrentUser *CurrentUser
	//Flash []FlashMessage
}

type CurrentUser struct {
	ID          int64
	DisplayName string
}

type FlashMessage struct {
	Type string
	Text string
}
