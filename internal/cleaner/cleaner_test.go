package cleaner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateRemovePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}

	bad := []string{
		"",
		"relative/path",
		"/",
		"/Users",
		"/var/folders",
		home,
		filepath.Dir(home),
		home + "/",
	}
	for _, p := range bad {
		if err := validateRemovePath(p); err == nil {
			t.Errorf("validateRemovePath(%q) should be rejected", p)
		}
	}

	good := []string{
		filepath.Join(home, "Library", "Caches", "SomeApp"),
		filepath.Join(home, "Library", "Application Support", "Claude", "GPUCache"),
	}
	for _, p := range good {
		if err := validateRemovePath(p); err != nil {
			t.Errorf("validateRemovePath(%q) unexpected error: %v", p, err)
		}
	}
}
