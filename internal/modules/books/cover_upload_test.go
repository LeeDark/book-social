package books

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeCoverPreparer struct {
	calls int
	err   error
	cover LocalFrontCover
	after func()
}

func (p *fakeCoverPreparer) ValidateAndStore(context.Context, CoverUploadFile, *CoverUploadFile) (LocalFrontCover, error) {
	p.calls++
	if p.after != nil {
		p.after()
	}

	return p.cover, p.err
}

func TestOperatorUploadAuthorizationAndFailures(t *testing.T) {
	for _, failure := range []string{"forbidden", "read", "storage", "reauthorize", "audit", "commit", "none"} {
		t.Run(failure, func(t *testing.T) {
			repo := &fakeOperatorStore{admin: true, before: OperatorBook{ID: 42, BookCorrection: BookCorrection{Title: "Book"}}}
			preparer := &fakeCoverPreparer{cover: LocalFrontCover{Main: StoredCoverFile{
				StorageKey: strings.Repeat("a", 64) + ".png", ChecksumSHA256: strings.Repeat("a", 64),
				MIMEType: "image/png", ByteSize: 100, Width: 4, Height: 6,
			}}}
			switch failure {
			case "forbidden":
				repo.admin = false
			case "read", "audit", "commit":
				repo.fail = failure
			case "storage":
				preparer.err = errors.New("storage unavailable")
			case "reauthorize":
				preparer.after = func() { repo.admin = false }
			}

			input := BookCorrection{Title: "Book", FrontCoverAction: CoverUpload,
				FrontCoverFile: &CoverUploadFile{Reader: strings.NewReader("prepared by adapter"), Filename: "a.png"}}
			err := NewOperatorServiceWithCovers(repo, preparer).Correct(context.Background(), 7, 42, input)
			if failure == "none" {
				if err != nil || !repo.committed || repo.audit.Summary != "front_cover" {
					t.Fatalf("upload = %v, audit %+v", err, repo.audit)
				}
			} else if err == nil || repo.committed {
				t.Fatalf("failed upload = %v, committed %t", err, repo.committed)
			}

			wantCalls := 1
			if failure == "forbidden" || failure == "read" {
				wantCalls = 0
			}
			if preparer.calls != wantCalls {
				t.Fatalf("prepare calls = %d, want %d", preparer.calls, wantCalls)
			}
		})
	}
}

func TestOperatorKeepDoesNotWriteCover(t *testing.T) {
	repo := &fakeOperatorStore{admin: true, before: OperatorBook{BookCorrection: BookCorrection{
		Title: "Book", FrontCoverURL: "/media/covers/old.png",
	}}}

	if err := NewOperatorService(repo).Correct(context.Background(), 7, 42, BookCorrection{Title: "Book"}); err != nil {
		t.Fatal(err)
	}

	for _, step := range repo.steps {
		if step == "cover" {
			t.Fatal("default keep wrote cover")
		}
	}

	if repo.audit.Summary != "" {
		t.Fatalf("keep audit = %q", repo.audit.Summary)
	}
}
