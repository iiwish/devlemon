package probe

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// buildTree 构造覆盖各类边界的目录树：Unicode/长文件名、符号链接、空目录、需多次批量读取的大目录
func buildTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write := func(rel string, n int) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, make([]byte, n), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("a.txt", 123)
	write("中文目录/文件 😀.bin", 4567)
	write("deep/1/2/3/4/5/x", 89)
	write(strings.Repeat("长", 80)+"/y", 10)
	if err := os.MkdirAll(filepath.Join(root, "empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3000; i++ { // 超过单次 64KB 缓冲可容纳的条目数
		write(fmt.Sprintf("many/file-%04d-%s", i, strings.Repeat("n", i%60)), i)
	}
	// 指向目录外大文件的符号链接：不得跟随
	outside := filepath.Join(t.TempDir(), "big")
	if err := os.WriteFile(outside, make([]byte, 1<<20), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Dir(outside), filepath.Join(root, "dirlink")); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestReadDirSizesMatchesGeneric(t *testing.T) {
	root := buildTree(t)
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		gs, gd := genericReadDirSizes(p)
		bs, bd := readDirSizes(p)
		sort.Strings(gd)
		sort.Strings(bd)
		if gs != bs || strings.Join(gd, "\n") != strings.Join(bd, "\n") {
			t.Errorf("%s: generic=(%d,%v) fast=(%d,%v)", p, gs, gd, bs, bd)
		}
		return nil
	})
}

func TestFastDirSize(t *testing.T) {
	root := buildTree(t)
	var want int64
	_ = filepath.WalkDir(root, func(_ string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			if fi, err := d.Info(); err == nil {
				want += fi.Size()
			}
		}
		return nil
	})
	if got := FastDirSize(root); got != want {
		t.Errorf("FastDirSize = %d, want %d", got, want)
	}
	if got := FastDirSize(filepath.Join(root, "a.txt")); got != 123 {
		t.Errorf("FastDirSize(file) = %d, want 123", got)
	}
	if got := FastDirSize(filepath.Join(root, "missing")); got != 0 {
		t.Errorf("FastDirSize(missing) = %d, want 0", got)
	}
}
