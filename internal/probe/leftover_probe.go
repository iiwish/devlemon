package probe

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"devlemon/internal/config"
	"devlemon/internal/i18n"
	"devlemon/internal/model"
)

// LeftoverProbe 已卸载应用的残留数据、失效的登录启动项、下载目录中的安装包。
// 这些都不是可自动重建的缓存，可能含用户数据，因此一律为 caution 等级，仅在深度清理中手动勾选。
type LeftoverProbe struct{}

func NewLeftoverProbe() *LeftoverProbe {
	return &LeftoverProbe{}
}

func (p *LeftoverProbe) Name() string {
	return "Uninstalled App Leftovers Probe"
}

func (p *LeftoverProbe) Category() model.Category {
	return model.CategoryAppCache
}

// LeftoverGroupID 残留分组 ID，引擎据此将其排在手动确认分组之前
const LeftoverGroupID = "app_leftovers"

// leftoverMinIdle 残留数据须在该时长内无任何改动：仍有进程在写入说明应用（或同名命令行工具）还在使用
const leftoverMinIdle = 30 * 24 * time.Hour

// leftoverHandledIDs 已由其他探针按缓存子目录处理的应用，避免与整个容器目录重复统计
var leftoverHandledIDs = map[string]bool{
	"com.tencent.xinwechat": true, "com.tencent.weworkmac": true, "com.alibaba.dingtalkmac": true,
	"com.electron.lark": true, "com.bytedance.ee.lark": true, "com.tencent.qqmusicmac": true,
	"com.netease.163music": true, "com.tencent.qq": true, "com.microsoft.vscode": true,
}

// leftoverLocation ~/Library 下以 Bundle ID 命名数据的位置：dir 为目录，suffix 为条目名中 Bundle ID 之后的后缀；
// guiOnly 表示只有图形界面应用才会产生（偏好设置、WebKit、沙盒容器等），命令行工具不会
type leftoverLocation struct {
	dir     string
	suffix  string
	guiOnly bool
}

func leftoverLocations(home string) []leftoverLocation {
	lib := filepath.Join(home, "Library")
	return []leftoverLocation{
		{filepath.Join(lib, "Application Support"), "", false},
		{filepath.Join(lib, "Caches"), "", false},
		{filepath.Join(lib, "Logs"), "", false},
		{filepath.Join(lib, "Containers"), "", true},
		{filepath.Join(lib, "Application Scripts"), "", true},
		{filepath.Join(lib, "HTTPStorages"), "", true},
		{filepath.Join(lib, "HTTPStorages"), ".binarycookies", true},
		{filepath.Join(lib, "WebKit"), "", true},
		{filepath.Join(lib, "Cookies"), ".binarycookies", true},
		{filepath.Join(lib, "Preferences"), ".plist", true},
		{filepath.Join(lib, "Saved Application State"), ".savedState", true},
	}
}

var (
	leftoverOnce  sync.Once
	leftoverCache map[string][]string
)

// appLeftovers 返回「已卸载应用 Bundle ID（小写）-> 残留路径」，同一次扫描内缓存，供其他探针去重
func appLeftovers() map[string][]string {
	leftoverOnce.Do(func() {
		home, err := os.UserHomeDir()
		if err != nil {
			return
		}
		leftoverCache = findLeftovers(home, installedApps(), updaterClaimed(filepath.Join(home, "Library", "Caches")))
	})
	return leftoverCache
}

// isLeftoverName 判断 ~/Library 下的条目名是否属于已卸载应用
func isLeftoverName(name string) bool {
	_, ok := appLeftovers()[strings.ToLower(name)]
	return ok
}

func findLeftovers(home string, idx *installedAppIndex, claimed map[string]bool) map[string][]string {
	result := make(map[string][]string)
	if idx == nil || !idx.ok {
		return result
	}
	gui := make(map[string]bool)
	for _, loc := range leftoverLocations(home) {
		entries, err := os.ReadDir(loc.dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			name := e.Name()
			if loc.suffix != "" {
				if !strings.HasSuffix(name, loc.suffix) {
					continue
				}
				name = strings.TrimSuffix(name, loc.suffix)
			} else if !e.IsDir() {
				continue
			}
			id := strings.ToLower(name)
			if !looksLikeBundleID(name) || strings.HasPrefix(id, "com.apple.") ||
				strings.Contains(id, "devlemon") || leftoverHandledIDs[id] || idx.hasApp(id) {
				continue
			}
			p := filepath.Join(loc.dir, e.Name())
			if claimed[p] {
				continue
			}
			result[id] = append(result[id], leftoverPaths(p, claimed)...)
			gui[id] = gui[id] || loc.guiOnly
		}
	}
	// 只在 Application Support / Caches / Logs 中出现的多为命令行工具（如 com.vercel.cli 存放登录凭证），不予判定
	for id := range result {
		if !gui[id] {
			delete(result, id)
		}
	}
	// 任一路径近期仍有改动则整组放弃：应用可能装在非标准位置，或是同名的命令行工具仍在使用
	for id, paths := range result {
		for _, p := range paths {
			if modifiedWithin(p, leftoverMinIdle) {
				delete(result, id)
				break
			}
		}
	}
	return result
}

// leftoverPaths 排除已由其他规则统计的子目录（更新安装包、Chromium 安全缓存），避免重复计算体积
func leftoverPaths(p string, claimed map[string]bool) []string {
	skip := make(map[string]bool)
	for c := range claimed {
		if filepath.Dir(c) == p {
			skip[filepath.Base(c)] = true
		}
	}
	for _, c := range collectChromiumCaches(p) {
		skip[filepath.Base(c)] = true
	}
	if len(skip) == 0 {
		return []string{p}
	}
	return childPaths(p, skip)
}

func (p *LeftoverProbe) Scan(ctx context.Context, cfg *config.Config) (*model.Group, error) {
	group := &model.Group{
		ID:       LeftoverGroupID,
		Title:    i18n.T("🧳 Uninstalled App Leftovers & Installers (unchecked by default)", "🧳 已卸载应用残留与安装包（默认不勾选）"),
		Category: model.CategoryAppCache,
		Items:    make([]*model.Item, 0),
	}
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
		item.Category = model.CategoryAppCache
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

	// 1. 已卸载应用残留：每个 Bundle ID 一项
	for id, paths := range appLeftovers() {
		id, paths := id, paths
		run(func() {
			add(newCacheItem("leftover_"+id,
				fmt.Sprintf("%s (%s)", i18n.T("Uninstalled App Leftovers", "已卸载应用残留"), id),
				fmt.Sprintf(i18n.T("Settings, caches and data left behind by an app that is no longer installed (%d locations, untouched for 30+ days). Containers may hold documents created in the app; confirm you no longer need them.",
					"已不在本机的应用遗留的设置、缓存与数据（%d 处，30 天以上无改动）。沙盒容器中可能有在该应用内创建的文档，请确认不再需要后再勾选"), len(paths)),
				paths, 1*1024*1024, model.RiskCaution))
		})
	}

	// 2. 失效的登录启动项：指向的程序已不存在
	run(func() {
		for _, item := range brokenLaunchAgents(filepath.Join(home, "Library", "LaunchAgents")) {
			add(item)
		}
	})

	// 3. 下载 / 桌面中的安装包
	run(func() {
		for _, item := range installerPackages(home, installedApps()) {
			add(item)
		}
	})

	wg.Wait()
	return group, nil
}

// brokenLaunchAgents 找出 Program / ProgramArguments[0] 指向不存在文件的用户级 LaunchAgent
func brokenLaunchAgents(dir string) []*model.Item {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var items []*model.Item
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".plist") || strings.HasPrefix(e.Name(), "com.apple.") {
			continue
		}
		p := filepath.Join(dir, e.Name())
		prog := launchAgentProgram(p)
		// 只判定绝对路径；外接卷上的程序可能只是暂未挂载
		if prog == "" || !filepath.IsAbs(prog) || strings.HasPrefix(prog, "/Volumes/") {
			continue
		}
		if _, err := os.Lstat(prog); !os.IsNotExist(err) {
			continue
		}
		label := strings.TrimSuffix(e.Name(), ".plist")
		item := newCacheItem("broken_agent_"+label,
			fmt.Sprintf("%s (%s)", i18n.T("Broken Login Item", "失效的登录启动项"), label),
			fmt.Sprintf(i18n.T("Launch agent whose program no longer exists: %s. Usually left behind after uninstalling an app; it fails on every login. Takes effect after you log in again.",
				"指向的程序已不存在：%s。通常是卸载应用后遗留，每次登录都会启动失败；删除后重新登录生效"), prog),
			[]string{p}, 0, model.RiskCaution)
		if item != nil {
			items = append(items, item)
		}
	}
	return items
}

func launchAgentProgram(path string) string {
	v, err := readPlistFile(path)
	if err != nil {
		return ""
	}
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	if s, ok := m["Program"].(string); ok && s != "" {
		return s
	}
	if args, ok := m["ProgramArguments"].([]any); ok && len(args) > 0 {
		if s, ok := args[0].(string); ok {
			return s
		}
	}
	return ""
}

// installerExts 应用安装包格式（不含 .zip / .iso 等也常用于存放其他内容的格式）
var installerExts = map[string]bool{".dmg": true, ".pkg": true, ".mpkg": true, ".xip": true}

// installerPackages 找出 ~/Downloads、~/Desktop（含一级子目录）中的应用安装包，每个文件一项
func installerPackages(home string, idx *installedAppIndex) []*model.Item {
	var files []string
	for _, root := range []string{filepath.Join(home, "Downloads"), filepath.Join(home, "Desktop")} {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, e := range entries {
			p := filepath.Join(root, e.Name())
			if e.IsDir() && !isAppBundle(e.Name()) && filepath.Ext(e.Name()) != ".pkg" {
				sub, err := os.ReadDir(p)
				if err != nil {
					continue
				}
				for _, s := range sub {
					if installerExts[strings.ToLower(filepath.Ext(s.Name()))] {
						files = append(files, filepath.Join(p, s.Name()))
					}
				}
				continue
			}
			if installerExts[strings.ToLower(filepath.Ext(e.Name()))] {
				files = append(files, p)
			}
		}
	}
	sort.Strings(files)

	var items []*model.Item
	for _, f := range files {
		name := filepath.Base(f)
		desc := i18n.T("Installer downloaded to install an app. Usually no longer needed once the app is installed.", "用于安装应用的安装包，应用装好后通常不再需要")
		if idx != nil && installerAppInstalled(name, idx) {
			desc += i18n.T(" A matching app appears to be installed.", "（检测到对应应用已安装）")
		}
		if item := newCacheItem(fmt.Sprintf("installer_%x", fnv64(f)), fmt.Sprintf("%s (%s)", i18n.T("Installer Package", "安装包"), name),
			desc, []string{f}, 1*1024*1024, model.RiskCaution); item != nil {
			items = append(items, item)
		}
	}
	return items
}

func fnv64(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

// installerAppInstalled 用安装包文件名的首个单词与已安装应用名匹配，如 Docker-4.30.dmg -> Docker
func installerAppInstalled(fileName string, idx *installedAppIndex) bool {
	base := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	word := strings.FieldsFunc(base, func(r rune) bool {
		return r == '-' || r == '_' || r == ' ' || r == '.' || (r >= '0' && r <= '9')
	})
	if len(word) == 0 {
		return false
	}
	if idx.nameKeys[appNameKey(base)] {
		return true
	}
	return idx.nameKeys[appNameKey(word[0])]
}
