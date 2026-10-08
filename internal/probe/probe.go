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

// FastDirSize 计算目录总大小（字节）
func FastDirSize(path string) int64 {
	var total int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total
}
