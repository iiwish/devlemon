package probe

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// installedAppIndex 本机已安装应用的 Bundle ID 与名称索引，用于判断 ~/Library 下的数据是否属于已卸载应用
type installedAppIndex struct {
	bundleIDs map[string]bool // 小写 Bundle ID
	companies map[string]bool // Bundle ID 前两段（如 com.google），同一厂商只要还装着任一应用就不判为残留
	nameKeys  map[string]bool // 归一化应用名
	ok        bool            // /Applications 可读；不可读时一律不判定残留
}

var (
	appIndexOnce sync.Once
	appIndex     *installedAppIndex
)

// ResetScanCaches 清空单次扫描内共享的缓存（已安装应用索引、应用残留、清理规则），由引擎在每次扫描开始时调用，
// 保证 TUI 等常驻进程多次扫描时能感知应用的安装与卸载。不可与正在进行的扫描并发调用。
func ResetScanCaches() {
	appIndexOnce, appIndex = sync.Once{}, nil
	leftoverOnce, leftoverCache = sync.Once{}, nil
	rulesOnce, rulesSet = sync.Once{}, nil
}

// installedApps 构建一次并缓存（同一次扫描中多个探针共用）
func installedApps() *installedAppIndex {
	appIndexOnce.Do(func() {
		home, _ := os.UserHomeDir()
		appIndex = buildAppIndex(appSearchRoots(home))
		// 不在标准目录（如从「下载」中直接运行）但正在运行的应用同样视为已安装
		if bundles, ok := runningAppBundles(); ok {
			for b := range bundles {
				appIndex.addApp(b)
			}
		}
	})
	return appIndex
}

func appSearchRoots(home string) []string {
	return []string{
		"/Applications",
		"/System/Applications",
		"/Library/Input Methods",
		"/Library/PreferencePanes",
		filepath.Join(home, "Applications"),
		filepath.Join(home, "Library", "Input Methods"),
		filepath.Join(home, "Library", "PreferencePanes"),
	}
}

func buildAppIndex(roots []string) *installedAppIndex {
	idx := &installedAppIndex{
		bundleIDs: make(map[string]bool),
		companies: make(map[string]bool),
		nameKeys:  make(map[string]bool),
	}
	if len(roots) > 0 {
		_, err := os.ReadDir(roots[0])
		idx.ok = err == nil
	}
	for _, root := range roots {
		idx.walk(root, 3)
	}
	return idx
}

// walk 在 dir 下查找应用包（.app / .prefPane / 输入法），普通文件夹最多下钻 depth 层（如 /Applications/Setapp/）
func (idx *installedAppIndex) walk(dir string, depth int) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		p := filepath.Join(dir, e.Name())
		if isAppBundle(e.Name()) {
			idx.addApp(p)
			continue
		}
		// 不跟随符号链接，避免环路；/Applications 中的链接型应用依靠运行中检测与厂商规则兜底
		if e.IsDir() && depth > 0 {
			idx.walk(p, depth-1)
		}
	}
}

func isAppBundle(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".app" || ext == ".prefpane"
}

// addApp 记录应用包及其内嵌的登录项、辅助应用的 Bundle ID
func (idx *installedAppIndex) addApp(appPath string) {
	idx.addBundle(appPath)
	// 在 Mac 上运行的 iPhone / iPad 应用：外层只是壳，真正的 Info.plist 位于 Wrapper/<名称>.app/Info.plist
	if wrapped, err := os.ReadDir(filepath.Join(appPath, "Wrapper")); err == nil {
		for _, e := range wrapped {
			if strings.EqualFold(filepath.Ext(e.Name()), ".app") {
				idx.addIOSBundle(filepath.Join(appPath, "Wrapper", e.Name()))
			}
		}
	}
	for _, sub := range []string{"Library/LoginItems", "Library/LaunchServices", "Helpers", "MacOS", "PlugIns"} {
		entries, err := os.ReadDir(filepath.Join(appPath, "Contents", sub))
		if err != nil {
			continue
		}
		for _, e := range entries {
			ext := strings.ToLower(filepath.Ext(e.Name()))
			if ext == ".app" || ext == ".appex" || ext == ".xpc" {
				idx.addBundle(filepath.Join(appPath, "Contents", sub, e.Name()))
			}
		}
	}
}

func (idx *installedAppIndex) addBundle(bundlePath string) {
	if k := appNameKey(strings.TrimSuffix(filepath.Base(bundlePath), filepath.Ext(bundlePath))); k != "" {
		idx.nameKeys[k] = true
	}
	idx.addID(appBundleInfo(bundlePath, "CFBundleIdentifier"))
}

func (idx *installedAppIndex) addID(id string) {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" {
		return
	}
	idx.bundleIDs[id] = true
	if c := bundleCompany(id); c != "" {
		idx.companies[c] = true
	}
}

// addIOSBundle iOS 应用包的 Info.plist 直接位于包根目录（没有 Contents/）
func (idx *installedAppIndex) addIOSBundle(bundlePath string) {
	v, err := readPlistFile(filepath.Join(bundlePath, "Info.plist"))
	if err != nil {
		return
	}
	idx.addID(plistDictString(v, "CFBundleIdentifier"))
}

// bundleCompany 取 Bundle ID 的厂商前缀（前两段）
func bundleCompany(id string) string {
	parts := strings.SplitN(id, ".", 3)
	if len(parts) < 3 {
		return ""
	}
	return parts[0] + "." + parts[1]
}

// genericCompanies 前两段过于宽泛、不能代表同一厂商的前缀（如 com.github.<用户名>）
var genericCompanies = map[string]bool{
	"com.github": true, "io.github": true, "com.electron": true, "org.chromium": true,
	"com.example": true,
}

// hasApp 判断 Bundle ID 对应的应用（或其主应用、同厂商应用）是否仍安装
func (idx *installedAppIndex) hasApp(id string) bool {
	id = strings.ToLower(id)
	if idx.bundleIDs[id] {
		return true
	}
	// 主应用与其扩展 / 辅助程序互为前缀：com.foo.App 与 com.foo.App.ShareExtension
	for installed := range idx.bundleIDs {
		if strings.HasPrefix(id, installed+".") || strings.HasPrefix(installed, id+".") {
			return true
		}
	}
	c := bundleCompany(id)
	if c != "" && !genericCompanies[c] && idx.companies[c] {
		return true
	}
	// 二进制 / 损坏的 Info.plist 读不到 ID 时，用 Bundle ID 末段与应用名兜底匹配
	return idx.nameKeys[appNameKey(lastComponent(id))]
}

// looksLikeBundleID 判断目录 / 文件名是否为反向域名形式的 Bundle ID（至少三段，只含合法字符）
func looksLikeBundleID(name string) bool {
	parts := strings.Split(name, ".")
	if len(parts) < 3 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		for _, r := range p {
			if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
				return false
			}
		}
	}
	// 顶级段须为字母，排除版本号等形如 1.2.3 的名字
	first := parts[0][0]
	return first >= 'a' && first <= 'z' || first >= 'A' && first <= 'Z'
}
