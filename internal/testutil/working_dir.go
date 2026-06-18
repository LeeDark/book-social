package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func ChdirProjectRoot(t *testing.T) {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "../.."))
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("os.Chdir(%q) error = %v", root, err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore working dir: %v", err)
		}
	})
}
