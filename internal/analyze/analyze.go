// Package analyze 磁盘空间分析：统计目录下每个子项的占用并按体积排序，帮助定位缓存规则之外的大目录
package analyze

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"

	"devlemon/internal/probe"
)

// Entry 目录下的一个子项
type Entry struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	SizeBytes int64  `json:"size_bytes"`
	IsDir     bool   `json:"is_dir"`
}

// Result 一次分析的结果，Entries 按体积降序
type Result struct {
	Path      string  `json:"path"`
	SizeBytes int64   `json:"size_bytes"`
	Entries   []Entry `json:"entries"`
}

// Dir 统计 dir 下每个直接子项的大小（不跟随符号链接）
func Dir(ctx context.Context, dir string) (*Result, error) {
	dir = filepath.Clean(dir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	res := &Result{Path: dir, Entries: make([]Entry, len(entries))}
	var wg sync.WaitGroup
	sem := make(chan struct{}, runtime.NumCPU()) // 子项可能成千上万，限制顶层并发
	for i, e := range entries {
		i, e := i, e
		res.Entries[i] = Entry{Name: e.Name(), Path: filepath.Join(dir, e.Name()), IsDir: e.IsDir()}
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer func() { <-sem; wg.Done() }()
			if ctx.Err() != nil {
				return
			}
			res.Entries[i].SizeBytes = probe.FastDirSize(res.Entries[i].Path)
		}()
	}
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	sort.Slice(res.Entries, func(a, b int) bool { return res.Entries[a].SizeBytes > res.Entries[b].SizeBytes })
	for _, e := range res.Entries {
		res.SizeBytes += e.SizeBytes
	}
	return res, nil
}
