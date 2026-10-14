package probe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"devlemon/internal/i18n"
)

// 应用自动更新器在 ~/Library/Caches 中下载 / 暂存的安装包。三种主流更新框架：
//   - Squirrel.Mac (VS Code、Slack 等 Electron 应用)：<bundleID>.ShipIt/，退出应用时由 ShipIt 进程安装
//   - Sparkle (大量原生应用)：<bundleID>/org.sparkle-project.Sparkle/，退出应用时由 Autoupdate / Updater 安装
//   - electron-updater：<app>-updater/ 或 <bundleID>/pending/（含 update-info.json），退出应用时安装
//
// 应用未运行时这些安装包要么是安装失败的残留，要么会在下次检查更新时重新下载，删除无副作用；
// 应用运行中则可能是「退出时待安装」的更新，交由深度清理手动确认。

// 各更新框架的安装进程名：运行中说明可能正在替换应用，绝不能抽走对应框架的安装包
var (
	squirrelInstallerProcs = []string{"ShipIt"}
	sparkleInstallerProcs  = []string{"Autoupdate", "Updater"}
)

// updaterRecentWindow 安装包在该时间内有改动视为正在下载
const updaterRecentWindow = 10 * time.Minute

type updaterTarget struct {
	dirName                         string // ~/Library/Caches 下的目录名，用于去重
	path                            string
	bundleID                        string // 所属应用 Bundle ID（可能为空）
	appPath                         string // 所属应用 .app 路径（ShipIt 可从状态文件获得）
	nameHint                        string // 用于与运行中应用名模糊匹配
	stagedVersion, installedVersion string
	installerProcs                  []string // 对应更新框架的安装进程名
}

func (t updaterTarget) displayName() string {
	if t.appPath != "" {
		return strings.TrimSuffix(filepath.Base(t.appPath), ".app")
	}
	return t.dirName
}

// updaterCacheTargets 枚举 ~/Library/Caches 下所有更新器安装包目录
func updaterCacheTargets(cachesDir string) []updaterTarget {
	entries, err := os.ReadDir(cachesDir)
	if err != nil {
		return nil
	}
	var targets []updaterTarget
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		dir := filepath.Join(cachesDir, name)
		switch {
		case strings.HasSuffix(name, ".ShipIt"):
			t := updaterTarget{dirName: name, path: dir, bundleID: strings.TrimSuffix(name, ".ShipIt"), installerProcs: squirrelInstallerProcs}
			t.nameHint = lastComponent(t.bundleID)
			readShipItState(&t)
			targets = append(targets, t)
		case strings.HasSuffix(name, "-updater") || strings.Contains(name, "desktop-updater"):
			targets = append(targets, updaterTarget{dirName: name, path: dir, nameHint: strings.TrimSuffix(name, "-updater"), installerProcs: squirrelInstallerProcs})
		default:
			if p := filepath.Join(dir, "org.sparkle-project.Sparkle"); dirExists(p) {
				targets = append(targets, updaterTarget{dirName: name, path: p, bundleID: name, nameHint: lastComponent(name), installerProcs: sparkleInstallerProcs})
			}
			// electron-updater 在 macOS 上最终交给 Squirrel 的 ShipIt 安装
			if p := filepath.Join(dir, "pending"); fileExists(filepath.Join(p, "update-info.json")) {
				targets = append(targets, updaterTarget{dirName: name, path: p, bundleID: name, nameHint: lastComponent(name), installerProcs: squirrelInstallerProcs})
			}
		}
	}
	return targets
}

// updaterClaimed 返回已被更新器规则覆盖的路径集合，供「其他应用缓存」去重
func updaterClaimed(cachesDir string) map[string]bool {
	claimed := make(map[string]bool)
	for _, t := range updaterCacheTargets(cachesDir) {
		claimed[t.path] = true
	}
	return claimed
}

// readShipItState 解析 ShipItState.plist（Squirrel 以 JSON 写入）获得目标应用路径及新旧版本号
func readShipItState(t *updaterTarget) {
	data, err := os.ReadFile(filepath.Join(t.path, "ShipItState.plist"))
	if err != nil {
		return
	}
	var state struct {
		UpdateBundleURL string `json:"updateBundleURL"`
		TargetBundleURL string `json:"targetBundleURL"`
	}
	if json.Unmarshal(data, &state) != nil {
		return
	}
	if p := fileURLPath(state.TargetBundleURL); p != "" {
		t.appPath = p
		t.installedVersion = bundleVersion(p)
	}
	if p := fileURLPath(state.UpdateBundleURL); p != "" {
		t.stagedVersion = bundleVersion(p)
	}
}

func fileURLPath(s string) string {
	u, err := url.Parse(s)
	if err != nil || u.Scheme != "file" {
		return ""
	}
	return filepath.Clean(u.Path)
}

func (t updaterTarget) description() string {
	desc := i18n.T("Update installer downloaded by the app's auto-updater. Re-downloaded on the next update check if still needed.",
		"应用自动更新器下载的安装包，不含用户数据；如仍需要，应用下次检查更新时会重新下载")
	if t.stagedVersion != "" && t.installedVersion != "" {
		desc += fmt.Sprintf(i18n.T(" (installed %s, package %s)", "（已安装 %s，安装包 %s）"), t.installedVersion, t.stagedVersion)
	}
	return desc
}

// modifiedWithin 检查目录前三层内是否有条目在 d 时间内被修改（下载中的文件或新建的子目录都会体现）
func modifiedWithin(dir string, d time.Duration) bool {
	cutoff := time.Now().Add(-d)
	var walk func(p string, depth int) bool
	walk = func(p string, depth int) bool {
		info, err := os.Lstat(p)
		if err != nil {
			return false
		}
		if info.ModTime().After(cutoff) {
			return true
		}
		if !info.IsDir() || depth == 0 {
			return false
		}
		entries, _ := os.ReadDir(p)
		for _, e := range entries {
			if walk(filepath.Join(p, e.Name()), depth-1) {
				return true
			}
		}
		return false
	}
	return walk(dir, 3)
}

// lastComponent 取 Bundle ID 的最后一段，如 com.google.antigravity -> antigravity
func lastComponent(bundleID string) string {
	if i := strings.LastIndex(bundleID, "."); i >= 0 {
		return bundleID[i+1:]
	}
	return bundleID
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]`)

// appNameKey 归一化应用名用于模糊匹配：小写且只保留字母数字，过短的名字不参与匹配以免误判
func appNameKey(s string) string {
	k := nonAlnum.ReplaceAllString(strings.ToLower(s), "")
	if len(k) < 3 {
		return ""
	}
	return k
}

// appBundleInfo 读取 .app 包 Info.plist 中的字符串字段（支持 XML 与二进制格式）
func appBundleInfo(appPath, key string) string {
	v, err := readPlistFile(filepath.Join(appPath, "Contents", "Info.plist"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(plistDictString(v, key))
}

func bundleVersion(appPath string) string {
	return appBundleInfo(appPath, "CFBundleShortVersionString")
}

// outerAppBundle 从可执行文件路径中截取最外层 .app 包路径，如 Helper 进程归属到主应用
func outerAppBundle(exe string) string {
	if i := strings.Index(exe, ".app/"); i >= 0 {
		return exe[:i+len(".app")]
	}
	return ""
}

// parseProcArgs2 解析 macOS kern.procargs2 的返回：int32 argc | 可执行文件路径\0 | 若干 \0 填充 | argv[0..argc-1]（各以 \0 结尾）| 环境变量…
func parseProcArgs2(buf []byte) (exe string, args []string) {
	if len(buf) < 4 {
		return "", nil
	}
	argc := int(int32(uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24))
	rest := buf[4:]
	n := bytes.IndexByte(rest, 0)
	if n < 0 {
		return string(rest), nil
	}
	exe = string(rest[:n])
	rest = rest[n:]
	for len(rest) > 0 && rest[0] == 0 {
		rest = rest[1:]
	}
	for i := 0; i < argc && len(rest) > 0; i++ {
		n := bytes.IndexByte(rest, 0)
		if n < 0 {
			args = append(args, string(rest))
			break
		}
		args = append(args, string(rest[:n]))
		rest = rest[n+1:]
	}
	return exe, args
}
