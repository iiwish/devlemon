package probe

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAllocatedSizeOfSparseFile(t *testing.T) {
	dir := t.TempDir()
	sparse(t, filepath.Join(dir, "sparse.img"), 1<<30)
	if err := os.WriteFile(filepath.Join(dir, "data.bin"), make([]byte, 3<<20), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := logicalSize(dir); got < (1<<30)+(3<<20) {
		t.Errorf("logicalSize = %d, want >= 1GB+3MB", got)
	}
	// 稀疏文件几乎不占磁盘；写入的 3MB（全零也已真实分配）应被计入
	if got := allocatedSize(dir); got < 3<<20 || got > 16<<20 {
		t.Errorf("allocatedSize = %d, want about 3MB", got)
	}
	if got := allocatedSize(filepath.Join(dir, "missing")); got != 0 {
		t.Errorf("missing path = %d", got)
	}
}
