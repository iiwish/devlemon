package probe

import (
	"io/fs"
	"os"
	"path/filepath"
)

// logicalSize 按文件逻辑长度统计（默认口径）：目录用 FastDirSize，普通文件用 Size
func logicalSize(p string) int64 {
	info, err := os.Lstat(p)
	if err != nil {
		return 0
	}
	if info.IsDir() {
		return FastDirSize(p)
	}
	if info.Mode().IsRegular() {
		return info.Size()
	}
	return 0
}

// allocatedSize 按实际占用的磁盘块统计（与 du 口径一致，不跟随符号链接）。
// 用于稀疏文件：模拟器磁盘镜像的逻辑长度可达 10GB，实际只占用写入过的几百 MB，按逻辑长度会严重夸大可释放空间
func allocatedSize(p string) int64 {
	info, err := os.Lstat(p)
	if err != nil {
		return 0
	}
	if !info.IsDir() {
		if info.Mode().IsRegular() {
			return fileBlocks(info)
		}
		return 0
	}
	var total int64
	_ = filepath.WalkDir(p, func(_ string, d fs.DirEntry, err error) error {
		if err != nil || !d.Type().IsRegular() {
			return nil
		}
		if fi, err := d.Info(); err == nil {
			total += fileBlocks(fi)
		}
		return nil
	})
	return total
}
