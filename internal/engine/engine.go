package engine

import (
	"context"
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
			probe.NewCacheProbe(),
			probe.NewDockerProbe(),
			probe.NewSimulatorProbe(),
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
			report.Groups = append(report.Groups, res.group)
			report.TotalReclaimableBytes += res.group.TotalReclaimableBytes
		}
	}

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
