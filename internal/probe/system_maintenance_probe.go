package probe

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"devlemon/internal/config"
	"devlemon/internal/i18n"
	"devlemon/internal/model"
)

type SystemMaintenanceProbe struct{}

func NewSystemMaintenanceProbe() *SystemMaintenanceProbe {
	return &SystemMaintenanceProbe{}
}

func (p *SystemMaintenanceProbe) Name() string {
	return "System Maintenance & Trash Probe"
}

func (p *SystemMaintenanceProbe) Category() model.Category {
	return model.CategorySystemCache
}

func (p *SystemMaintenanceProbe) Scan(ctx context.Context, cfg *config.Config) (*model.Group, error) {
	group := &model.Group{
		ID:       "system_maintenance",
		Title:    i18n.T("🧹 System Maintenance & Trash", "🧹 系统基础维护与废纸篓"),
		Category: model.CategorySystemCache,
		Items:    make([]*model.Item, 0),
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return group, nil
	}

	// 1. 用户废纸篓扫描
	trashSize := getTrashSize(ctx, home)
	if trashSize > 0 {
		item := &model.Item{
			ID:            "user_trash",
			Title:         i18n.T("macOS User Trash", "macOS 用户废纸篓"),
			Description:   i18n.T("Deleted files waiting in Trash (~/.Trash), safe to purge", "已丢入废纸篓等待清空的文件，清理后彻底释放物理磁盘空间"),
			Path:          filepath.Join(home, ".Trash"),
			SizeBytes:     trashSize,
			SizeFormatted: model.FormatBytes(trashSize),
			Risk:          model.RiskSafe,
			Category:      model.CategorySystemCache,
			IsProtected:   false,
		}

		if runtime.GOOS == "darwin" {
			item.CleanType = model.CleanTypeCommand
			item.CleanCommand = []string{"osascript", "-e", "tell application \"Finder\" to empty trash"}
		} else {
			item.CleanType = model.CleanTypeRemovePath
			item.CleanPath = filepath.Join(home, ".local/share/Trash")
		}

		group.Items = append(group.Items, item)
		group.TotalSizeBytes += trashSize
		group.TotalReclaimableBytes += trashSize
	}

	// 2. 桌面软件自动更新安装包残留 (Electron / App updaters)
	cachesDir := filepath.Join(home, "Library", "Caches")
	if entries, err := os.ReadDir(cachesDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			if strings.HasSuffix(name, "-updater") || strings.Contains(name, "desktop-updater") {
				uDir := filepath.Join(cachesDir, name)
				sz := FastDirSize(uDir)
				if sz > 10*1024*1024 { // 大于 10MB
					uName := filepath.Base(uDir)
					item := &model.Item{
						ID:            "updater_" + uName,
						Title:         fmt.Sprintf("%s (%s)", i18n.T("App Update Cache", "应用更新残留包"), uName),
						Description:   i18n.T("Residual update archives downloaded by auto-updaters", "应用自动升级后遗留在 Caches 目录下的更新安装包残留"),
						Path:          uDir,
						SizeBytes:     sz,
						SizeFormatted: model.FormatBytes(sz),
						Risk:          model.RiskSafe,
						Category:      model.CategorySystemCache,
						IsProtected:   false,
						CleanType:     model.CleanTypeRemovePath,
						CleanPath:     uDir,
					}
					group.Items = append(group.Items, item)
					group.TotalSizeBytes += sz
					group.TotalReclaimableBytes += sz
				}
			}
		}
	}

	// 3. 过期系统与应用日志 (~/Library/Logs，超过 20MB 时提示)
	logsDir := filepath.Join(home, "Library", "Logs")
	if info, err := os.Stat(logsDir); err == nil && info.IsDir() {
		logsSize := FastDirSize(logsDir)
		if logsSize > 20*1024*1024 {
			item := &model.Item{
				ID:            "system_logs",
				Title:         i18n.T("Expired Application Logs", "过期应用与系统诊断日志"),
				Description:   i18n.T("Old application logs and crash dumps in ~/Library/Logs", "~/Library/Logs 下的历史运行日志与崩溃报告"),
				Path:          logsDir,
				SizeBytes:     logsSize,
				SizeFormatted: model.FormatBytes(logsSize),
				Risk:          model.RiskSafe,
				Category:      model.CategorySystemCache,
				IsProtected:   false,
				CleanType:     model.CleanTypeCommand,
				CleanCommand:  []string{"find", logsDir, "-mindepth", "1", "-delete"},
			}
			group.Items = append(group.Items, item)
			group.TotalSizeBytes += logsSize
			group.TotalReclaimableBytes += logsSize
		}
	}

	// 4. Xcode 辅助索引与临时缓存 (如果存在且大于 50MB)
	xcodeCacheDir := filepath.Join(home, "Library", "Caches", "com.apple.dt.Xcode")
	if info, err := os.Stat(xcodeCacheDir); err == nil && info.IsDir() {
		xcSize := FastDirSize(xcodeCacheDir)
		if xcSize > 50*1024*1024 {
			item := &model.Item{
				ID:            "xcode_aux_cache",
				Title:         i18n.T("Xcode Auxiliary Caches", "Xcode 辅助索引与临时缓存"),
				Description:   i18n.T("Documentation and runtime caches for Xcode", "Xcode 文档索引与临时构建辅助缓存，删除后自动按需生成"),
				Path:          xcodeCacheDir,
				SizeBytes:     xcSize,
				SizeFormatted: model.FormatBytes(xcSize),
				Risk:          model.RiskSafe,
				Category:      model.CategorySystemCache,
				IsProtected:   false,
				CleanType:     model.CleanTypeRemovePath,
				CleanPath:     xcodeCacheDir,
			}
			group.Items = append(group.Items, item)
			group.TotalSizeBytes += xcSize
			group.TotalReclaimableBytes += xcSize
		}
	}

	// 5. 常见浏览器与主流应用缓存 (主流浏览器网络/渲染缓存，删除后自动重新生成，不影响书签与登录)
	type browserTarget struct {
		id    string
		name  string
		rel   string
		minSz int64
	}

	browserTargets := []browserTarget{
		{id: "chrome_cache", name: "Google Chrome 浏览器网络缓存", rel: filepath.Join("Library", "Caches", "Google", "Chrome"), minSz: 10 * 1024 * 1024},
		{id: "safari_cache", name: "Safari 浏览器网络缓存", rel: filepath.Join("Library", "Caches", "com.apple.Safari"), minSz: 10 * 1024 * 1024},
		{id: "edge_cache", name: "Microsoft Edge 浏览器缓存", rel: filepath.Join("Library", "Caches", "Microsoft Edge"), minSz: 10 * 1024 * 1024},
		{id: "arc_cache", name: "Arc 浏览器网络缓存", rel: filepath.Join("Library", "Caches", "company.thebrowser.Arc"), minSz: 10 * 1024 * 1024},
		{id: "brave_cache", name: "Brave 浏览器网络缓存", rel: filepath.Join("Library", "Caches", "BraveSoftware", "Brave-Browser"), minSz: 10 * 1024 * 1024},
		{id: "firefox_cache", name: "Firefox 浏览器网络缓存", rel: filepath.Join("Library", "Caches", "Firefox"), minSz: 10 * 1024 * 1024},
		{id: "wechat_cache", name: "微信应用临时图片与媒体缓存", rel: filepath.Join("Library", "Caches", "com.tencent.xinWeChat"), minSz: 20 * 1024 * 1024},
	}

	for _, bt := range browserTargets {
		p := filepath.Join(home, bt.rel)
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			sz := FastDirSize(p)
			if sz >= bt.minSz {
				item := &model.Item{
					ID:            bt.id,
					Title:         bt.name,
					Description:   i18n.T("Browser/App web cache. Safe to clean without affecting accounts or bookmarks.", "应用运行产生的临时网络请求与渲染缓存，清理后不影响账号登录与书签"),
					Path:          p,
					SizeBytes:     sz,
					SizeFormatted: model.FormatBytes(sz),
					Risk:          model.RiskSafe,
					Category:      model.CategorySystemCache,
					IsProtected:   false,
					CleanType:     model.CleanTypeRemovePath,
					CleanPath:     p,
				}
				group.Items = append(group.Items, item)
				group.TotalSizeBytes += sz
				group.TotalReclaimableBytes += sz
			}
		}
	}

	return group, nil
}

func getTrashSize(ctx context.Context, home string) int64 {
	trashPath := filepath.Join(home, ".Trash")
	// 1. 优先尝试直接 POSIX 统计 (如果授予了完全磁盘访问权限或非 Mac)
	if size := FastDirSize(trashPath); size > 0 {
		return size
	}

	// 2. macOS 下若因 TCC 权限被锁，通过 Finder AppleScript 统计 (不弹权限窗口，秒级返回)
	if runtime.GOOS == "darwin" {
		script := `tell application "Finder"
	set totalSize to 0
	repeat with anItem in (get every item of trash)
		try
			set totalSize to totalSize + (size of anItem as integer)
		end try
	end repeat
	return totalSize
end tell`
		cmd := exec.CommandContext(ctx, "osascript", "-e", script)
		out, err := cmd.Output()
		if err == nil {
			str := strings.TrimSpace(string(out))
			if val, parseErr := strconv.ParseInt(str, 10, 64); parseErr == nil {
				return val
			}
		}
	}
	return 0
}
