package probe

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"

	"devlemon/internal/config"
	"devlemon/internal/model"
)

// Probe 探针接口定义
type Probe interface {
	Name() string
	Category() model.Category
	Scan(ctx context.Context, cfg *config.Config) (*model.Group, error)
}

// dirSizeSem 全局目录遍历并发上限，由所有探针的 FastDirSize 调用共享，避免多探针并发时线程爆炸
var dirSizeSem = make(chan struct{}, runtime.NumCPU()*2)

// FastDirSize 并行计算目录总大小（不跟随符号链接）。子目录在有空闲并发额度时交给新 goroutine，
// 否则在当前 goroutine 内递归，因此不会死锁。macOS 上单目录读取使用 getattrlistbulk 批量获取大小。
func FastDirSize(path string) int64 {
	info, err := os.Lstat(path)
	if err != nil {
		return 0
	}
	if !info.IsDir() {
		return info.Size()
	}

	var total atomic.Int64
	var wg sync.WaitGroup
	var walk func(dir string)
	walk = func(dir string) {
		defer wg.Done()
		size, subdirs := readDirSizes(dir)
		total.Add(size)
		for _, sub := range subdirs {
			wg.Add(1)
			select {
			case dirSizeSem <- struct{}{}:
				go func() {
					defer func() { <-dirSizeSem }()
					walk(sub)
				}()
			default:
				walk(sub)
			}
		}
	}

	wg.Add(1)
	walk(path)
	wg.Wait()
	return total.Load()
}

// genericReadDirSizes 通用实现：ReadDir + 逐项 lstat
func genericReadDirSizes(dir string) (int64, []string) {
	f, err := os.Open(dir)
	if err != nil {
		return 0, nil
	}
	entries, _ := f.ReadDir(-1)
	f.Close()

	var total int64
	var subdirs []string
	for _, e := range entries {
		if e.IsDir() {
			subdirs = append(subdirs, filepath.Join(dir, e.Name()))
			continue
		}
		if fi, err := e.Info(); err == nil {
			total += fi.Size()
		}
	}
	return total, subdirs
}
