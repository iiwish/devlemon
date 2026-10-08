package probe

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devlemon/internal/config"
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

func (p *WorkspaceProbe) Scan(ctx context.Context, cfg *config.Config) (*model.Group, error) {
	group := &model.Group{
		ID:       "workspace_builds",
		Title:    "📂 项目构建产物与依赖（target/、.build、node_modules）",
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

	now := time.Now()
	dormancyDuration := time.Duration(cfg.DormancyDays) * 24 * time.Hour

	for _, wsRoot := range cfg.WorkspacePaths {
		info, err := os.Stat(wsRoot)
		if err != nil || !info.IsDir() {
			continue
		}

		// 遍历工作区下的每个子目录
		_ = filepath.Walk(wsRoot, func(currentPath string, info os.FileInfo, err error) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if err != nil || !info.IsDir() {
				return nil
			}

			// 计算相对深度，避免深度死循环
			rel, err := filepath.Rel(wsRoot, currentPath)
			if err != nil {
				return nil
			}
			depth := len(strings.Split(rel, string(filepath.Separator)))
			if depth > cfg.MaxDepth && rel != "." {
				return filepath.SkipDir
			}

			baseName := filepath.Base(currentPath)

			// 如果命中目标构建目录（如 target, .build, node_modules）
			if targetMap[baseName] && rel != "." {
				size := FastDirSize(currentPath)
				// 仅收集大于 100MB 的构建产物，避免过细碎片
				if size > 100*1024*1024 {
					projectName := filepath.Base(filepath.Dir(currentPath))
					daysAgo := int(now.Sub(info.ModTime()) / (24 * time.Hour))

					isDormant := now.Sub(info.ModTime()) >= dormancyDuration

					desc := fmt.Sprintf("项目 [%s] 的中间构建产物，距最后修改约 %d 天", projectName, daysAgo)
					if isDormant {
						desc += "（已休眠，可安全清理）"
					} else {
						desc += "（近期活跃，重新编译将耗时）"
					}

					item := &model.Item{
						ID:            fmt.Sprintf("build_%s_%s", projectName, baseName),
						Title:         fmt.Sprintf("%s / %s", projectName, baseName),
						Description:   desc,
						Path:          currentPath,
						SizeBytes:     size,
						SizeFormatted: model.FormatBytes(size),
						Risk:          model.RiskRebuildable,
						Category:      model.CategoryWorkspaceBuild,
						IsProtected:   false,
						CleanType:     model.CleanTypeRemovePath,
						CleanPath:     currentPath,
					}

					group.Items = append(group.Items, item)
					group.TotalSizeBytes += size
					group.TotalReclaimableBytes += size
				}

				// 已经收集了该构建目录，不需要继续往下递归其子目录
				return filepath.SkipDir
			}

			return nil
		})
	}

	return group, nil
}
