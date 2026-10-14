//go:build darwin || linux

package probe

import (
	"os"
	"syscall"
)

// fileBlocks 文件实际分配的磁盘空间（st_blocks 以 512 字节为单位）
func fileBlocks(fi os.FileInfo) int64 {
	if st, ok := fi.Sys().(*syscall.Stat_t); ok {
		return st.Blocks * 512
	}
	return fi.Size()
}
