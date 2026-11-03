package books

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseOperatorFormReportsSpecificCoverError(t *testing.T) {
	const (
		sizeError   = "Each image must contain data and be at most 5 MiB."
		actionError = "Choose one cover action and supply only its URL or files."
	)

	cases := []struct {
		name      string
		mainImage []byte
		want      string
	}{
		{name: "oversized main image reports size limit", mainImage: make([]byte, (5<<20)+1), want: sizeError},
		{name: "empty main image reports size limit", mainImage: []byte{}, want: sizeError},
		{name: "upload without main image reports action rule", want: actionError},
	}

	for _, check := range cases {
		t.Run(check.name, func(t *testing.T) {
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			if err := writer.WriteField("title", "Book"); err != nil {
				t.Fatal(err)
			}
			if err := writer.WriteField("front_cover_action", string(CoverUpload)); err != nil {
				t.Fatal(err)
			}
			if check.mainImage != nil {
				part, err := writer.CreateFormFile("front_cover_file", "cover.png")
				if err != nil {
					t.Fatal(err)
				}
				if _, err := part.Write(check.mainImage); err != nil {
					t.Fatal(err)
				}
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}

			request := httptest.NewRequest(http.MethodPost, "/admin/catalog/books/1/edit", &body)
			request.Header.Set("Content-Type", writer.FormDataContentType())
			form, input := parseOperatorForm(httptest.NewRecorder(), request, false)
			t.Cleanup(func() { cleanupOperatorForm(request, input) })

			if got := form.Errors["front_cover"]; got != check.want {
				t.Fatalf("front_cover error = %q, want %q", got, check.want)
			}
		})
	}
}
