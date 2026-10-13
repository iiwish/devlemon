package probe

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"devlemon/internal/config"
	"devlemon/internal/i18n"
	"devlemon/internal/model"
)

// ManualReviewProbe 保守策略下不进入快速清理的项目：不属于用户数据，但删除后可能产生可感知的影响
// （网站离线数据丢失、需重新下载、排查材料丢失等）。仅在深度清理中展示，风险等级均为非 safe，
// 因此 UI 默认不勾选、`clean --safe` 也绝不会清理，必须由用户逐项手动确认。
type ManualReviewProbe struct{}

func NewManualReviewProbe() *ManualReviewProbe {
	return &ManualReviewProbe{}
}

func (p *ManualReviewProbe) Name() string {
	return "Manual Review Items Probe"
}

func (p *ManualReviewProbe) Category() model.Category {
	return model.CategorySystemCache
}

// ManualReviewGroupID 手动确认分组 ID，引擎据此将其排在所有分组最后
const ManualReviewGroupID = "manual_review"

// tempFileMinAge 临时文件（含其内部所有文件）超过该时长未被修改才纳入候选
const tempFileMinAge = 7 * 24 * time.Hour

// libraryCachesSkip ~/Library/Caches 下不纳入「其他应用缓存」的目录：
// 已由其他探针或规则覆盖（避免重复统计），或明确存放了用户数据（绝不触碰）
var libraryCachesSkip = map[string]bool{
	// 用户数据：CodexBar 在 Caches 中保存 Token 用量账单数据库
	"CodexBar": true,
	// 已由 CacheProbe 覆盖
	"go-build": true, "pnpm": true, "pip": true, "Homebrew": true, "ms-playwright": true,
	"CocoaPods": true, "node-gyp": true, "Yarn": true, "deno": true, "composer": true,
	// 已由 AppCacheProbe 覆盖
	"com.microsoft.VSCode": true, "Codex": true, "com.tencent.xinWeChat": true, "com.netease.163music": true,
	// 已由浏览器规则 / SystemMaintenanceProbe 覆盖
	"Google": true, "Microsoft Edge": true, "company.thebrowser.Arc": true, "BraveSoftware": true,
	"Firefox": true, "com.apple.Safari": true, "com.apple.dt.Xcode": true,
}

func (p *ManualReviewProbe) Scan(ctx context.Context, cfg *config.Config) (*model.Group, error) {
	group := &model.Group{
		ID:       ManualReviewGroupID,
		Title:    i18n.T("🔍 Needs Manual Review (unchecked by default)", "🔍 需手动确认的清理项（默认不勾选）"),
		Category: model.CategorySystemCache,
		Items:    make([]*model.Item, 0),
	}
	// 快速清理只关心安全项，跳过本探针以保证菜单栏秒级返回
	if cfg != nil && cfg.SafeOnly {
		return group, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return group, nil
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	add := func(item *model.Item) {
		if item == nil {
			return
		}
		mu.Lock()
		group.Items = append(group.Items, item)
		group.TotalSizeBytes += item.SizeBytes
		group.TotalReclaimableBytes += item.SizeBytes
		mu.Unlock()
	}
	run := func(f func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
			}
			f()
		}()
	}

	// 1. 崩溃诊断报告：可能是用户排查问题、反馈 Bug 的材料
	run(func() {
		dir := filepath.Join(home, "Library", "Logs", "DiagnosticReports")
		add(newCacheItem("crash_reports",
			i18n.T("Crash Diagnostic Reports", "应用崩溃诊断报告"),
			i18n.T("Crash reports in ~/Library/Logs/DiagnosticReports. Keep them if you need to troubleshoot or report a bug.", "~/Library/Logs/DiagnosticReports 下的崩溃报告。如需排查问题或向开发者反馈 Bug，请保留"),
			childPaths(dir, nil), 1*1024*1024, model.RiskCaution))
	})

	// 2. Chromium 系浏览器：组件更新下载缓存 + 网站 Service Worker 离线数据
	for _, b := range chromiumBrowsers {
		b := b
		root := filepath.Join(home, b.dataRel)
		if !dirExists(root) {
			continue
		}
		run(func() {
			add(newCacheItem(b.id+"_component_cache",
				fmt.Sprintf("%s %s", b.name, i18n.T("Component Download Cache", "组件更新下载缓存")),
				i18n.T("Installers downloaded by the browser component updater. Re-downloaded when components update.", "浏览器组件更新器下载的安装包，不含用户数据；清理后组件更新时需重新下载"),
				[]string{filepath.Join(root, "component_crx_cache")}, 10*1024*1024, model.RiskRebuildable))
		})
		run(func() {
			var paths []string
			for _, profile := range chromiumProfiles(root) {
				paths = append(paths, serviceWorkerCaches(profile)...)
			}
			add(newCacheItem(b.id+"_service_worker",
				fmt.Sprintf("%s %s", b.name, i18n.T("Website Offline Data (Service Worker)", "网站离线数据 (Service Worker)")),
				serviceWorkerDesc(), paths, 10*1024*1024, model.RiskCaution))
		})
	}

	// 3. Electron 应用的 Service Worker 离线数据
	for _, t := range electronCacheTargets(home) {
		t := t
		if isLeftoverName(filepath.Base(t.userDataRoot)) {
			continue // 已卸载应用的离线数据随残留一并统计
		}
		run(func() {
			add(newCacheItem("electron_sw_"+strings.TrimPrefix(t.id, "electron_cache_"),
				fmt.Sprintf("%s (%s)", i18n.T("App Offline Data (Service Worker)", "应用离线数据 (Service Worker)"), filepath.Base(t.userDataRoot)),
				serviceWorkerDesc(), serviceWorkerCaches(t.userDataRoot), 10*1024*1024, model.RiskCaution))
		})
	}

	// 4. Xcode 归档：含发布版本的 dSYM 符号文件，删除后无法再符号化线上崩溃
	run(func() {
		dir := filepath.Join(home, "Library", "Developer", "Xcode", "Archives")
		add(newCacheItem("xcode_archives",
			i18n.T("Xcode Archives", "Xcode 打包归档 (Archives)"),
			i18n.T("App builds archived for distribution, including dSYM symbols. Keep the ones you may need to symbolicate crash reports of released versions.", "为发布打包的应用归档，内含 dSYM 符号文件；已发布版本的归档用于解析线上崩溃，请确认不再需要后再勾选"),
			childPaths(dir, nil), 50*1024*1024, model.RiskCaution))
	})

	// 5. iOS 设备本地备份：真实用户数据，可能是设备的唯一备份
	for _, backup := range childPaths(filepath.Join(home, "Library", "Application Support", "MobileSync", "Backup"), nil) {
		backup := backup
		run(func() {
			name := backupDeviceName(backup)
			add(newCacheItem("ios_backup_"+filepath.Base(backup),
				fmt.Sprintf("%s (%s)", i18n.T("iPhone / iPad Local Backup", "iPhone / iPad 本地备份"), name),
				i18n.T("Device backup made by Finder. This may be the only copy of the device's data; delete only backups of devices you no longer need.", "Finder 生成的设备备份，可能是设备数据的唯一副本；仅删除确定不再需要的旧设备备份"),
				[]string{backup}, 100*1024*1024, model.RiskCaution))
		})
	}

	// 6. 本地大模型：用户主动拉取的模型文件，重新下载耗时很长，只列出不预选
	run(func() {
		add(newCacheItem("ollama_models",
			i18n.T("Ollama Local Models", "Ollama 本地大模型"),
			i18n.T("Models pulled with ollama pull (~/.ollama/models). Prefer `ollama rm <model>` to remove single models; cleaning here removes all of them and they must be pulled again.", "通过 ollama pull 下载的模型 (~/.ollama/models)。建议用 `ollama rm <模型名>` 单独删除；在此清理会删除全部模型，需要重新拉取"),
			[]string{filepath.Join(home, ".ollama", "models")}, 100*1024*1024, model.RiskCaution))
	})
	run(func() {
		add(newCacheItem("lmstudio_models",
			i18n.T("LM Studio Local Models", "LM Studio 本地大模型"),
			i18n.T("Models downloaded in LM Studio. Removing them means downloading again in the app.", "LM Studio 中下载的模型，删除后需在应用内重新下载"),
			[]string{filepath.Join(home, ".lmstudio", "models"), filepath.Join(home, ".cache", "lm-studio", "models")}, 100*1024*1024, model.RiskCaution))
	})

	// 7. 过期临时文件：用户临时目录中 7 天以上未修改的条目
	run(func() { add(scanStaleTempFiles()) })

	// 8. ~/Library/Caches 下其他应用缓存：按 Apple 规范可清除，但个别应用会违规存放数据，必须逐项确认
	// 更新安装包由 SystemMaintenanceProbe 统计：整个目录被覆盖则跳过，仅子目录被覆盖则统计其余子项
	cachesDir := filepath.Join(home, "Library", "Caches")
	claimed := updaterClaimed(cachesDir)
	if entries, err := os.ReadDir(cachesDir); err == nil {
		for _, e := range entries {
			name := e.Name()
			dir := filepath.Join(cachesDir, name)
			if !e.IsDir() || libraryCachesSkip[name] || claimed[dir] || isLeftoverName(name) || strings.HasPrefix(name, "com.apple.") ||
				strings.Contains(strings.ToLower(name), "devlemon") {
				continue
			}
			skip := make(map[string]bool)
			for p := range claimed {
				if filepath.Dir(p) == dir {
					skip[filepath.Base(p)] = true
				}
			}
			paths := []string{dir}
			if len(skip) > 0 {
				paths = childPaths(dir, skip)
			}
			run(func() {
				add(newCacheItem("other_cache_"+name,
					fmt.Sprintf("%s (%s)", i18n.T("Other App Cache", "其他应用缓存"), name),
					i18n.T("Cache folder in ~/Library/Caches. Usually rebuildable, but a few apps store data here; confirm you recognize this app before cleaning.", "~/Library/Caches 下的应用缓存，通常可自动重建；但少数应用会在此存放数据，请确认认识该应用后再勾选"),
					paths, 20*1024*1024, model.RiskCaution))
			})
		}
	}

	wg.Wait()
	return group, nil
}

var backupDeviceNameRe = regexp.MustCompile(`<key>Device Name</key>\s*<string>([^<]+)</string>`)

// backupDeviceName 从备份的 Info.plist 读取设备名，读不到时回退为备份目录名
func backupDeviceName(backup string) string {
	if data, err := os.ReadFile(filepath.Join(backup, "Info.plist")); err == nil {
		if m := backupDeviceNameRe.FindSubmatch(data); m != nil {
			return string(m[1])
		}
	}
	return filepath.Base(backup)
}

func serviceWorkerDesc() string {
	return i18n.T("Offline data cached by websites/web apps via Service Worker. Not login data, but offline-capable pages (PWAs) need to go online again to rebuild it.",
		"网站或网页应用通过 Service Worker 缓存的离线资源，不含登录信息；但支持离线使用的网页应用 (PWA) 清理后需联网重新加载")
}

// chromiumProfiles 返回 Chromium 用户数据根目录下的所有 Profile 目录（以存在 Preferences 文件为准）
func chromiumProfiles(userDataRoot string) []string {
	entries, err := os.ReadDir(userDataRoot)
	if err != nil {
		return nil
	}
	var profiles []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		profile := filepath.Join(userDataRoot, e.Name())
		if fileExists(filepath.Join(profile, "Preferences")) {
			profiles = append(profiles, profile)
		}
	}
	return profiles
}

// serviceWorkerCaches 返回 Service Worker 的 CacheStorage 与 ScriptCache 目录（不含注册数据库 Database）
func serviceWorkerCaches(root string) []string {
	var paths []string
	for _, sub := range []string{"CacheStorage", "ScriptCache"} {
		if p := filepath.Join(root, "Service Worker", sub); dirExists(p) {
			paths = append(paths, p)
		}
	}
	return paths
}

// childPaths 返回目录下的直接子项（跳过 skip 中的名称）
func childPaths(dir string, skip map[string]bool) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var paths []string
	for _, e := range entries {
		if !skip[e.Name()] {
			paths = append(paths, filepath.Join(dir, e.Name()))
		}
	}
	return paths
}

// scanStaleTempFiles 扫描用户临时目录 ($TMPDIR) 中整体超过 7 天未修改的条目。
// 只要条目内部任一文件近期被修改、或含有 socket/管道等运行期特殊文件，整个条目即跳过。
func scanStaleTempFiles() *model.Item {
	tmp := filepath.Clean(os.TempDir())
	// 仅处理 macOS 每用户临时目录，绝不触碰 /tmp 等共享目录
	if !strings.HasPrefix(tmp, "/var/folders/") && !strings.HasPrefix(tmp, "/private/var/folders/") {
		return nil
	}
	entries, err := os.ReadDir(tmp)
	if err != nil {
		return nil
	}

	cutoff := time.Now().Add(-tempFileMinAge)
	var stale []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "com.apple.") {
			continue
		}
		p := filepath.Join(tmp, e.Name())
		if isStale(p, cutoff) {
			stale = append(stale, p)
		}
	}
	return newCacheItem("stale_temp_files",
		i18n.T("Stale Temporary Files", "过期临时文件"),
		i18n.T("Entries in the user temp folder ($TMPDIR) untouched for over 7 days. Apps normally clean these themselves; review before cleaning.", "用户临时目录 ($TMPDIR) 中超过 7 天未被修改的文件，通常是应用退出时遗留的临时文件"),
		stale, 10*1024*1024, model.RiskCaution)
}

// isStale 路径本身及其内部所有文件均早于 cutoff，且不含 socket / 管道 / 设备等特殊文件
func isStale(path string, cutoff time.Time) bool {
	stale := true
	_ = filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			stale = false
			return filepath.SkipAll
		}
		info, err := d.Info()
		if err != nil {
			stale = false
			return filepath.SkipAll
		}
		mode := info.Mode()
		if !mode.IsRegular() && !mode.IsDir() && mode&fs.ModeSymlink == 0 {
			stale = false
			return filepath.SkipAll
		}
		if info.ModTime().After(cutoff) {
			stale = false
			return filepath.SkipAll
		}
		return nil
	})
	return stale
}
