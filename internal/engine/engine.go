package engine

import (
	"context"
	"sort"
	"sync"
	"syscall"
	"time"

	"devlemon/internal/config"
	"devlemon/internal/model"
	"devlemon/internal/probe"
)

type Engine struct {
	probes []probe.Probe
	cfg    *config.Config
}

func NewEngine(cfg *config.Config) *Engine {
	e := &Engine{
		cfg: cfg,
		probes: []probe.Probe{
			probe.NewSystemMaintenanceProbe(),
			probe.NewAppCacheProbe(),
			probe.NewDockerProbe(),
			probe.NewSimulatorProbe(),
			probe.NewCacheProbe(),
			probe.NewWorkspaceProbe(),
		},
	}
	return e
}

// AddProbe 允许外部注册自定义探针
func (e *Engine) AddProbe(p probe.Probe) {
	e.probes = append(e.probes, p)
}

// Scan 并发执行所有注册探针并汇总报告
func (e *Engine) Scan(ctx context.Context) (*model.ScanReport, error) {
	report := &model.ScanReport{
		Groups:    make([]*model.Group, 0),
		ScannedAt: time.Now(),
	}

	// 1. 读取系统磁盘状态（优先查询 /System/Volumes/Data）
	readDiskStats(report)

	// 2. 并发调度探针
	type probeResult struct {
		group *model.Group
		err   error
	}

	results := make(chan probeResult, len(e.probes))
	var wg sync.WaitGroup

	for _, p := range e.probes {
		// 关键性能优化：在仅安全项模式 (SafeOnly) 下，仅执行纯系统维护探针，毫秒级快速返回
		if e.cfg != nil && e.cfg.SafeOnly {
			if p.Category() != model.CategorySystemCache {
				continue
			}
		}

		// 沙盒合规模式 (Mac App Store)：严格隔离，跳过无权限的底层 Docker 容器与未显式授权的工作区
		if e.cfg != nil && e.cfg.SandboxMode {
			if p.Category() == model.CategoryDocker {
				continue
			}
			if p.Category() == model.CategoryWorkspaceBuild && len(e.cfg.WorkspacePaths) == 0 {
				continue
			}
		}

		wg.Add(1)
		go func(pr probe.Probe) {
			defer wg.Done()
			grp, err := pr.Scan(ctx, e.cfg)
			results <- probeResult{group: grp, err: err}
		}(p)
	}

	wg.Wait()
	close(results)

	for res := range results {
		if res.err == nil && res.group != nil && len(res.group.Items) > 0 {
			// 每个分区内部的所有清理条目按占用体积从大到小严格降序排列
			sort.Slice(res.group.Items, func(i, j int) bool {
				return res.group.Items[i].SizeBytes > res.group.Items[j].SizeBytes
			})
			report.Groups = append(report.Groups, res.group)
			report.TotalReclaimableBytes += res.group.TotalReclaimableBytes
		}
	}

	// 3. 稳定排序组：按优先级展示各类目
	categoryOrder := map[model.Category]int{
		model.CategorySystemCache:    1, // 🧹 系统基础维护与废纸篓
		model.CategoryAppCache:       2, // 📱 应用垃圾与多媒体缓存
		model.CategoryDocker:         3, // 🐳 Docker / OrbStack 容器资源
		model.CategorySimulator:      4, // 🛠️ iOS 模拟器环境
		model.CategoryPackageCache:   5, // 📦 开发语言与包管理器缓存
		model.CategoryWorkspaceBuild: 6, // 📂 项目构建产物与依赖
	}

	sort.Slice(report.Groups, func(i, j int) bool {
		return categoryOrder[report.Groups[i].Category] < categoryOrder[report.Groups[j].Category]
	})

	return report, nil
}

func readDiskStats(r *model.ScanReport) {
	var stat syscall.Statfs_t
	path := "/System/Volumes/Data"
	if err := syscall.Statfs(path, &stat); err != nil {
		path = "/"
		_ = syscall.Statfs(path, &stat)
	}

	bsize := int64(stat.Bsize)
	total := int64(stat.Blocks) * bsize
	free := int64(stat.Bavail) * bsize
	used := total - free

	r.DiskTotal = total
	r.DiskFree = free
	r.DiskUsed = used
	if total > 0 {
		r.DiskCapacityPercent = int(float64(used) / float64(total) * 100)
	}
}
