//go:build !darwin && !linux

package probe

import "os"

// fileBlocks 不支持读取分配块数的平台回退为逻辑长度
func fileBlocks(fi os.FileInfo) int64 {
	return fi.Size()
}
