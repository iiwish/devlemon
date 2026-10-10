package probe

import (
	"context"
	"os"
	"path/filepath"

	"devlemon/internal/config"
	"devlemon/internal/model"
)

// Probe 探针接口定义
type Probe interface {
	Name() string
	Category() model.Category
	Scan(ctx context.Context, cfg *config.Config) (*model.Group, error)
}

// FastDirSize 高性能计算目录总大小（使用 WalkDir 规避昂贵的单文件 lstat 系统调用，大幅提速）
func FastDirSize(path string) int64 {
	var total int64
	_ = filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			if info, err := d.Info(); err == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total
}
