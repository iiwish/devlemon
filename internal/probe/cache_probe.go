package probe

import (
	"context"
	"os"
	"path/filepath"

	"devlemon/internal/config"
	"devlemon/internal/model"
)

type CacheProbe struct{}

func NewCacheProbe() *CacheProbe {
	return &CacheProbe{}
}

func (p *CacheProbe) Name() string {
	return "Developer & System Cache Probe"
}

func (p *CacheProbe) Category() model.Category {
	return model.CategoryPackageCache
}

type cacheTarget struct {
	id          string
	title       string
	description string
	subpath     string
	risk        model.RiskLevel
	cleanCmd    []string
}

func (p *CacheProbe) Scan(ctx context.Context, cfg *config.Config) (*model.Group, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	targets := []cacheTarget{
		{
			id:          "go_build_cache",
			title:       "Go 编译中间缓存",
			description: "Go 编译器生成的临时目标文件，删除后随时按需自动重新生成",
			subpath:     filepath.Join("Library", "Caches", "go-build"),
			risk:        model.RiskSafe,
			cleanCmd:    []string{"go", "clean", "-cache"},
		},
		{
			id:          "go_mod_cache",
			title:       "Go 模块下载缓存",
			description: "本地下载缓存的第三方 Go 依赖包 (pkg/mod)",
			subpath:     filepath.Join("go", "pkg", "mod"),
			risk:        model.RiskSafe,
			cleanCmd:    []string{"go", "clean", "-modcache"},
		},
		{
			id:          "npm_cache",
			title:       "npm 离线安装包缓存",
			description: "npm 全局安装包缓存 (~/.npm)，删除无副作用",
			subpath:     ".npm",
			risk:        model.RiskSafe,
			cleanCmd:    []string{"npm", "cache", "clean", "--force"},
		},
		{
			id:          "pnpm_cache",
			title:       "pnpm 临时存储缓存",
			description: "pnpm 客户端本地下载缓存",
			subpath:     filepath.Join("Library", "Caches", "pnpm"),
			risk:        model.RiskSafe,
		},
		{
			id:          "uv_cache",
			title:       "Python uv 高速包缓存",
			description: "Python uv 工具链下载与预构建 wheel 缓存",
			subpath:     filepath.Join(".cache", "uv"),
			risk:        model.RiskSafe,
			cleanCmd:    []string{"uv", "cache", "clean"},
		},
		{
			id:          "pip_cache",
			title:       "Python pip 下载缓存",
			description: "pip 安装包离线缓存",
			subpath:     filepath.Join("Library", "Caches", "pip"),
			risk:        model.RiskSafe,
			cleanCmd:    []string{"pip", "cache", "purge"},
		},
		{
			id:          "homebrew_cache",
			title:       "Homebrew 安装包与 Bottle 缓存",
			description: "Brew 下载的历史版本二进制包与安装归档",
			subpath:     filepath.Join("Library", "Caches", "Homebrew"),
			risk:        model.RiskSafe,
			cleanCmd:    []string{"brew", "cleanup", "-s"},
		},
		{
			id:          "vscode_shipit",
			title:       "VS Code 软件更新临时安装包",
			description: "VS Code 自动更新后遗留在 Caches 里的更新安装包",
			subpath:     filepath.Join("Library", "Caches", "com.microsoft.VSCode.ShipIt"),
			risk:        model.RiskSafe,
		},
		{
			id:          "playwright_browsers",
			title:       "Playwright 自动化无头浏览器包",
			description: "用于 E2E 自动化测试的 Chromium/WebKit 独立浏览器包",
			subpath:     filepath.Join("Library", "Caches", "ms-playwright"),
			risk:        model.RiskSafe,
		},
	}

	group := &model.Group{
		ID:       "package_caches",
		Title:    "📦 开发语言与包管理器缓存",
		Category: model.CategoryPackageCache,
		Items:    make([]*model.Item, 0),
	}

	for _, tgt := range targets {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		fullPath := filepath.Join(home, tgt.subpath)
		info, err := os.Stat(fullPath)
		if err != nil || !info.IsDir() {
			continue
		}

		size := FastDirSize(fullPath)
		if size < 10*1024*1024 { // 小于 10MB 的微小缓存暂时忽略，保持输出干净
			continue
		}

		item := &model.Item{
			ID:            tgt.id,
			Title:         tgt.title,
			Description:   tgt.description,
			Path:          fullPath,
			SizeBytes:     size,
			SizeFormatted: model.FormatBytes(size),
			Risk:          tgt.risk,
			Category:      model.CategoryPackageCache,
			IsProtected:   false,
			CleanType:     model.CleanTypeRemovePath,
			CleanPath:     fullPath,
		}

		if len(tgt.cleanCmd) > 0 {
			item.CleanType = model.CleanTypeCommand
			item.CleanCommand = tgt.cleanCmd
		}

		group.Items = append(group.Items, item)
		group.TotalSizeBytes += size
		group.TotalReclaimableBytes += size
	}

	return group, nil
}
