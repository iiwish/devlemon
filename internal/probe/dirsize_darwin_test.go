//go:build darwin

package probe

import (
	"os"
	"path/filepath"
	"testing"
)

// 确保 macOS 上真正走的是 getattrlistbulk 快速路径，而不是被回退逻辑掩盖的通用实现
func TestBulkReadDirSizesNoFallback(t *testing.T) {
	root := buildTree(t)
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if _, _, ok := bulkReadDirSizes(p); !ok {
			t.Errorf("bulkReadDirSizes fell back for %s", p)
		}
		return nil
	})
}
