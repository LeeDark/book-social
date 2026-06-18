package view

import "testing"

func TestMainNavigationReturnsPublicNavigationItems(t *testing.T) {
	nav := MainNavigation()

	if got, want := len(nav), 5; got != want {
		t.Fatalf("len(MainNavigation()) = %d, want %d", got, want)
	}

	tests := []struct {
		index int
		key   string
		label string
		href  string
	}{
		{index: 0, key: "home", label: "Home", href: "/"},
		{index: 1, key: "catalog", label: "Catalog", href: "/books"},
		{index: 4, key: "about", label: "About", href: "/about"},
	}

	for _, tt := range tests {
		item := nav[tt.index]

		if item.Key != tt.key {
			t.Errorf("nav[%d].Key = %q, want %q", tt.index, item.Key, tt.key)
		}
		if item.Label != tt.label {
			t.Errorf("nav[%d].Label = %q, want %q", tt.index, item.Label, tt.label)
		}
		if item.Href != tt.href {
			t.Errorf("nav[%d].Href = %q, want %q", tt.index, item.Href, tt.href)
		}
	}

	for _, item := range nav {
		if !item.Public {
			t.Errorf("nav item %q is not public", item.Key)
		}
	}
}
