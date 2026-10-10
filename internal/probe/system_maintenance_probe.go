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
	"sync"

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

	// 1. 用户废纸篓：无权限时需调用 Finder AppleScript（约 0.2~0.4 秒），与其余扫描并行执行
	trashCh := make(chan int64, 1)
	go func() { trashCh <- getTrashSize(ctx, home) }()

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

	// 3. 应用运行日志 (~/Library/Logs，超过 1MB 时提示)
	// 崩溃诊断报告 DiagnosticReports 可能是用户排查问题或反馈 Bug 的材料，不纳入安全项，交由深度清理手动确认
	logsDir := filepath.Join(home, "Library", "Logs")
	if entries, err := os.ReadDir(logsDir); err == nil {
		var logPaths []string
		for _, e := range entries {
			if e.Name() == "DiagnosticReports" {
				continue
			}
			logPaths = append(logPaths, filepath.Join(logsDir, e.Name()))
		}
		if item := newCacheItem("system_logs",
			i18n.T("Application Logs", "应用与系统运行日志"),
			i18n.T("Historical runtime logs in ~/Library/Logs (crash reports excluded)", "~/Library/Logs 下的历史运行日志（不含崩溃诊断报告）"),
			logPaths, 1*1024*1024, model.RiskSafe); item != nil {
			group.Items = append(group.Items, item)
			group.TotalSizeBytes += item.SizeBytes
			group.TotalReclaimableBytes += item.SizeBytes
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

	// 5. 浏览器与 Electron 应用缓存。运行中的应用缓存降级为非安全项：快速清理跳过，深度清理中由用户手动勾选
	targets := browserCacheTargets(home)
	targets = append(targets, electronCacheTargets(home)...)

	rc := newRunningChecker()

	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, t := range targets {
		t := t
		wg.Add(1)
		go func() {
			defer wg.Done()
			item := newCacheItem(t.id, t.title, t.description, t.paths, t.minSize, model.RiskSafe)
			if item == nil {
				return
			}
			running := false
			if t.processName != "" {
				running = rc.processRunning(t.processName)
			}
			if t.userDataRoot != "" {
				running = running || rc.chromiumRunning(t.userDataRoot)
			}
			if running {
				item.Risk = model.RiskRebuildable
				item.Title += i18n.T(" (Running)", "（运行中）")
				item.Description = i18n.T("App is running, skipped by quick clean. Quit the app before cleaning to avoid page reloads. ", "应用正在运行，快速清理已自动跳过；建议退出应用后再勾选清理，避免页面重新加载。") + item.Description
			}
			mu.Lock()
			group.Items = append(group.Items, item)
			group.TotalSizeBytes += item.SizeBytes
			group.TotalReclaimableBytes += item.SizeBytes
			mu.Unlock()
		}()
	}
	wg.Wait()

	// 汇总废纸篓结果
	trashSize := <-trashCh
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

	return group, nil
}

func getTrashSize(ctx context.Context, home string) int64 {
	trashPath := filepath.Join(home, ".Trash")
	// 1. 优先尝试直接 POSIX 统计 (如果授予了完全磁盘访问权限或非 Mac)；目录可读时直接信任结果（含空废纸篓）
	if _, err := os.ReadDir(trashPath); err == nil {
		return FastDirSize(trashPath)
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

// appCacheTarget 一个应用（浏览器 / Electron 应用）的缓存目录集合
type appCacheTarget struct {
	id           string
	title        string
	description  string
	paths        []string
	minSize      int64
	userDataRoot string // Chromium 用户数据根目录，用于 SingletonLock 运行检测
	processName  string // 进程名，用于非 Chromium 浏览器运行检测
}

type browserSpec struct {
	id, name, cacheRel, dataRel, process string
}

// chromiumBrowsers Chromium 系浏览器：~/Library/Caches 下的网络缓存 + 用户数据根目录下各 Profile 的 GPU/着色器缓存
var chromiumBrowsers = []browserSpec{
	{id: "chrome", name: "Google Chrome", cacheRel: "Library/Caches/Google/Chrome", dataRel: "Library/Application Support/Google/Chrome"},
	{id: "edge", name: "Microsoft Edge", cacheRel: "Library/Caches/Microsoft Edge", dataRel: "Library/Application Support/Microsoft Edge"},
	{id: "arc", name: "Arc", cacheRel: "Library/Caches/company.thebrowser.Arc", dataRel: "Library/Application Support/Arc/User Data"},
	{id: "brave", name: "Brave", cacheRel: "Library/Caches/BraveSoftware/Brave-Browser", dataRel: "Library/Application Support/BraveSoftware/Brave-Browser"},
}

var otherBrowsers = []browserSpec{
	{id: "safari", name: "Safari", cacheRel: "Library/Caches/com.apple.Safari", process: "Safari"},
	{id: "firefox", name: "Firefox", cacheRel: "Library/Caches/Firefox", process: "firefox"},
}

func browserCacheTargets(home string) []appCacheTarget {
	desc := i18n.T("Browser network, code and GPU shader caches. Safe to clean without affecting accounts, cookies or bookmarks.", "浏览器网络请求、脚本编译与 GPU 着色器缓存，清理后不影响账号登录、Cookie 与书签")
	var targets []appCacheTarget
	for _, b := range append(append([]browserSpec{}, chromiumBrowsers...), otherBrowsers...) {
		t := appCacheTarget{
			id:          b.id + "_cache",
			title:       fmt.Sprintf("%s %s", b.name, i18n.T("Browser Cache", "浏览器缓存")),
			description: desc,
			minSize:     10 * 1024 * 1024,
			processName: b.process,
		}
		if p := filepath.Join(home, b.cacheRel); dirExists(p) {
			t.paths = append(t.paths, p)
		}
		if b.dataRel != "" {
			t.userDataRoot = filepath.Join(home, b.dataRel)
			t.paths = append(t.paths, collectChromiumBrowserCaches(t.userDataRoot)...)
		}
		if len(t.paths) > 0 {
			targets = append(targets, t)
		}
	}
	return targets
}

// electronSkipDirs 已由其他探针或浏览器规则覆盖的 Application Support 子目录，避免重复统计
var electronSkipDirs = map[string]bool{
	"Code":           true, // VS Code，由 AppCacheProbe 处理
	"DingTalkMac":    true, // 钉钉，由 AppCacheProbe 处理
	"Google":         true,
	"Microsoft Edge": true,
	"BraveSoftware":  true,
	"Arc":            true,
}

// electronCacheTargets 枚举所有 Electron / Chromium 内核桌面应用的纯缓存目录，每个应用聚合为一项
func electronCacheTargets(home string) []appCacheTarget {
	supportDir := filepath.Join(home, "Library", "Application Support")
	entries, err := os.ReadDir(supportDir)
	if err != nil {
		return nil
	}

	var targets []appCacheTarget
	for _, e := range entries {
		if !e.IsDir() || electronSkipDirs[e.Name()] {
			continue
		}
		root := filepath.Join(supportDir, e.Name())
		paths := collectChromiumCaches(root)
		if len(paths) == 0 {
			continue
		}
		targets = append(targets, appCacheTarget{
			id:           "electron_cache_" + e.Name(),
			title:        fmt.Sprintf("%s (%s)", i18n.T("App Web Cache", "应用网页缓存"), e.Name()),
			description:  i18n.T("Chromium/Electron network, script and GPU caches. Rebuilt automatically; does not touch login state or user data.", "Chromium/Electron 内核的网络、脚本编译与 GPU 缓存，自动重建，不涉及登录状态与用户数据"),
			paths:        paths,
			minSize:      5 * 1024 * 1024,
			userDataRoot: root,
		})
	}
	return targets
}

// newCacheItem 将一个或多个目录聚合为一个清理项，体积不足 minSize 时返回 nil
func newCacheItem(id, title, desc string, paths []string, minSize int64, risk model.RiskLevel) *model.Item {
	var total int64
	var valid []string
	for _, p := range paths {
		info, err := os.Lstat(p)
		if err != nil {
			continue
		}
		var sz int64
		if info.IsDir() {
			sz = FastDirSize(p)
		} else if info.Mode().IsRegular() {
			sz = info.Size()
		}
		if sz > 0 {
			total += sz
			valid = append(valid, p)
		}
	}
	if total < minSize || len(valid) == 0 {
		return nil
	}

	return &model.Item{
		ID:            id,
		Title:         title,
		Description:   desc,
		Path:          valid[0],
		SizeBytes:     total,
		SizeFormatted: model.FormatBytes(total),
		Risk:          risk,
		Category:      model.CategorySystemCache,
		CleanType:     model.CleanTypeRemovePaths,
		CleanPaths:    valid,
	}
}
