package analyze

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDir(t *testing.T) {
	root := t.TempDir()
	write := func(p string, n int) {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, make([]byte, n), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(root, "big", "a", "b.bin"), 3000)
	write(filepath.Join(root, "small.txt"), 10)
	write(filepath.Join(root, "mid", "c.bin"), 500)
	// 符号链接不跟随，避免把外部目录重复计入
	if err := os.Symlink(filepath.Join(root, "big"), filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}

	res, err := Dir(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if res.Entries[0].Name != "big" || res.Entries[0].SizeBytes != 3000 || !res.Entries[0].IsDir {
		t.Errorf("largest entry = %+v", res.Entries[0])
	}
	if res.Entries[1].Name != "mid" {
		t.Errorf("entries not sorted by size: %+v", res.Entries)
	}
	if res.SizeBytes < 3510 || res.SizeBytes > 3510+512 {
		t.Errorf("total = %d", res.SizeBytes)
	}
}
