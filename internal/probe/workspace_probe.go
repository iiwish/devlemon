package probe

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"devlemon/internal/config"
	"devlemon/internal/i18n"
	"devlemon/internal/model"
)

type WorkspaceProbe struct{}

func NewWorkspaceProbe() *WorkspaceProbe {
	return &WorkspaceProbe{}
}

func (p *WorkspaceProbe) Name() string {
	return "Project Workspace & Build Artifact Probe"
}

func (p *WorkspaceProbe) Category() model.Category {
	return model.CategoryWorkspaceBuild
}

type candidateDir struct {
	path        string
	modTime     time.Time
	projectName string
	baseName    string
}

func (p *WorkspaceProbe) Scan(ctx context.Context, cfg *config.Config) (*model.Group, error) {
	group := &model.Group{
		ID:       "workspace_builds",
		Title:    i18n.T("📂 Workspace Builds & Dependencies (target, .build, node_modules)", "📂 项目构建产物与依赖（target/、.build、node_modules）"),
		Category: model.CategoryWorkspaceBuild,
		Items:    make([]*model.Item, 0),
	}

	if len(cfg.WorkspacePaths) == 0 {
		return group, nil
	}

	targetMap := make(map[string]bool)
	for _, name := range cfg.TargetDirNames {
		targetMap[name] = true
	}

	// 1. 第一步：极速收集目标构建目录（不在此步计算大小，耗时 < 0.2 秒）
	var candidates []candidateDir

	for _, wsRoot := range cfg.WorkspacePaths {
		info, err := os.Stat(wsRoot)
		if err != nil || !info.IsDir() {
			continue
		}

		_ = filepath.Walk(wsRoot, func(currentPath string, info os.FileInfo, err error) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if err != nil || !info.IsDir() {
				return nil
			}

			rel, err := filepath.Rel(wsRoot, currentPath)
			if err != nil {
				return nil
			}
			depth := len(strings.Split(rel, string(filepath.Separator)))
			if depth > cfg.MaxDepth && rel != "." {
				return filepath.SkipDir
			}

			baseName := filepath.Base(currentPath)

			// 如果命中目标构建目录名称
			if targetMap[baseName] && rel != "." {
				projectName := filepath.Base(filepath.Dir(currentPath))
				candidates = append(candidates, candidateDir{
					path:        currentPath,
					modTime:     info.ModTime(),
					projectName: projectName,
					baseName:    baseName,
				})
				// 命中后不再深入该目录
				return filepath.SkipDir
			}

			return nil
		})
	}

	if len(candidates) == 0 {
		return group, nil
	}

	// 2. 第二步：使用 16 并发 Worker 线程池极速并行计算各目录体积（速度提升 10 倍！）
	now := time.Now()
	dormancyDuration := time.Duration(cfg.DormancyDays) * 24 * time.Hour

	workerCount := 16
	jobs := make(chan candidateDir, len(candidates))
	type resultItem struct {
		item *model.Item
		size int64
	}
	results := make(chan resultItem, len(candidates))

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for c := range jobs {
				size := FastDirSize(c.path)
				// 仅保留大于 100MB 的构建目录，忽略碎片
				if size > 100*1024*1024 {
					daysAgo := int(now.Sub(c.modTime) / (24 * time.Hour))
					isDormant := now.Sub(c.modTime) >= dormancyDuration

					var desc string
					if isDormant {
						desc = fmt.Sprintf(i18n.T("Build artifacts for [%s], modified %d days ago (Dormant, safe to clean)", "项目 [%s] 的中间构建产物，距最后修改约 %d 天（已休眠，可安全清理）"), c.projectName, daysAgo)
					} else {
						desc = fmt.Sprintf(i18n.T("Build artifacts for [%s], modified %d days ago (Recently active, recompilation takes time)", "项目 [%s] 的中间构建产物，距最后修改约 %d 天（近期活跃，重新编译将耗时）"), c.projectName, daysAgo)
					}

					item := &model.Item{
						ID:            fmt.Sprintf("build_%s_%s", c.projectName, c.baseName),
						Title:         fmt.Sprintf("%s / %s", c.projectName, c.baseName),
						Description:   desc,
						Path:          c.path,
						SizeBytes:     size,
						SizeFormatted: model.FormatBytes(size),
						Risk:          model.RiskRebuildable,
						Category:      model.CategoryWorkspaceBuild,
						IsProtected:   false,
						CleanType:     model.CleanTypeRemovePath,
						CleanPath:     c.path,
					}
					results <- resultItem{item: item, size: size}
				}
			}
		}()
	}

	for _, c := range candidates {
		jobs <- c
	}
	close(jobs)

	wg.Wait()
	close(results)

	for res := range results {
		group.Items = append(group.Items, res.item)
		group.TotalSizeBytes += res.size
		group.TotalReclaimableBytes += res.size
	}

	return group, nil
}
