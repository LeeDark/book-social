package library

import "testing"

func TestReadingStatusLabel(t *testing.T) {
	tests := []struct {
		name   string
		status ReadingStatus
		want   string
	}{
		{name: "want to read", status: ReadingStatusWantToRead, want: "Want to read"},
		{name: "reading", status: ReadingStatusReading, want: "Reading"},
		{name: "read", status: ReadingStatusRead, want: "Read"},
		{name: "unknown status uses default label", status: "unknown", want: "Want to read"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.Label(); got != tt.want {
				t.Fatalf("Label() = %q, want %q", got, tt.want)
			}
		})
	}
}
