package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// ModuleRoot returns the absolute path to the directory containing go.mod,
// found by searching upward from the *caller's* own source file (not this
// one). Prefer this over hardcoding a relative path like
// filepath.Dir(filepath.Dir(thisFile)) — that kind of depth-counting breaks
// silently the next time a test package moves (as tests/e2e already has
// once).
func ModuleRoot(t testing.TB) string {
	t.Helper()
	_, callerFile, _, ok := runtime.Caller(1)
	if !ok {
		t.Fatal("utils.ModuleRoot: runtime.Caller failed")
	}

	dir := filepath.Dir(callerFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("utils.ModuleRoot: no go.mod found above %s", callerFile)
		}
		dir = parent
	}
}
