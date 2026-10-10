package probe

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// Chromium / Electron 内核产生的纯缓存目录（HTTP 磁盘缓存、V8 字节码、GPU/着色器缓存）。
// 这些目录本身就会被 Chromium 按 LRU 随时自动淘汰，删除等价于一次缓存淘汰，下次使用时自动重建。
// 严格排除 Cookies、Local Storage、IndexedDB、Service Worker/CacheStorage（PWA 离线数据）等站点与登录数据。
var chromiumCacheSubdirs = []string{
	"Cache",
	"Code Cache",
	"GPUCache",
	"DawnCache",
	"DawnGraphiteCache",
	"DawnWebGPUCache",
	"GrShaderCache",
	"GraphiteDawnCache",
	"ShaderCache",
}

func fileExists(p string) bool {
	info, err := os.Lstat(p)
	return err == nil && info.Mode().IsRegular()
}

// isSimpleCacheLayout Chromium Simple Cache 后端：index 文件 + index-dir 目录
func isSimpleCacheLayout(dir string) bool {
	return fileExists(filepath.Join(dir, "index")) && dirExists(filepath.Join(dir, "index-dir"))
}

// isBlockfileCacheLayout Chromium Blockfile 后端：index 文件 + data_0 块文件
func isBlockfileCacheLayout(dir string) bool {
	return fileExists(filepath.Join(dir, "index")) && fileExists(filepath.Join(dir, "data_0"))
}

// isChromiumCacheDir 按目录名精确校验 Chromium 磁盘缓存的文件结构，结构不符一律不认，
// 避免误删恰好同名的业务数据目录（宁可漏扫，不可误删）
func isChromiumCacheDir(dir string) bool {
	switch filepath.Base(dir) {
	case "Cache":
		return isSimpleCacheLayout(filepath.Join(dir, "Cache_Data")) || isBlockfileCacheLayout(dir)
	case "Code Cache":
		return isSimpleCacheLayout(filepath.Join(dir, "js")) || isSimpleCacheLayout(filepath.Join(dir, "wasm"))
	default:
		return isBlockfileCacheLayout(dir)
	}
}

// collectChromiumCaches 返回 root 目录（Electron 应用数据根或浏览器 Profile）下已存在的 Chromium 缓存子目录
func collectChromiumCaches(root string) []string {
	var found []string
	for _, sub := range chromiumCacheSubdirs {
		p := filepath.Join(root, sub)
		if dirExists(p) && isChromiumCacheDir(p) {
			found = append(found, p)
		}
	}
	return found
}

// collectChromiumBrowserCaches 扫描 Chromium 系浏览器用户数据根目录（如 Application Support/Google/Chrome）
// 下所有 Profile 的 GPU/着色器缓存，以及根级别的着色器缓存
func collectChromiumBrowserCaches(userDataRoot string) []string {
	found := collectChromiumCaches(userDataRoot)
	for _, profile := range chromiumProfiles(userDataRoot) {
		found = append(found, collectChromiumCaches(profile)...)
	}
	return found
}

// dirExists 使用 Lstat，符号链接一律视为不存在，绝不跟随链接到缓存目录之外
func dirExists(p string) bool {
	info, err := os.Lstat(p)
	return err == nil && info.IsDir()
}

// chromiumLockHeld 检查 Chromium ProcessSingleton 锁：用户数据根目录下的 SingletonLock 符号链接，
// 内容为 "主机名-PID"。进程存活即视为应用正在运行；锁存在但无法解析时保守视为运行中。
func chromiumLockHeld(userDataRoot string) bool {
	target, err := os.Readlink(filepath.Join(userDataRoot, "SingletonLock"))
	if err != nil {
		return false
	}
	idx := strings.LastIndex(target, "-")
	if idx < 0 {
		return true
	}
	pid, err := strconv.Atoi(target[idx+1:])
	if err != nil || pid <= 0 {
		return true
	}
	err = syscall.Kill(pid, 0)
	return err == nil || err == syscall.EPERM
}

// leveldbLockSubpaths Chromium/Electron 运行期间一定会打开的 LevelDB 数据库。LevelDB 打开后会用 fcntl
// 对其 LOCK 文件加写锁，进程退出即自动释放，因此锁是否被持有可准确反映应用是否运行。
var leveldbLockSubpaths = []string{
	filepath.Join("Local Storage", "leveldb", "LOCK"),
	filepath.Join("Session Storage", "LOCK"),
	filepath.Join("shared_proto_db", "LOCK"),
	filepath.Join("Extension State", "LOCK"),
	filepath.Join("GCM Store", "LOCK"),
	filepath.Join("Site Characteristics Database", "LOCK"),
}

// fcntlLockHeld 用 F_GETLK 查询文件上是否存在其他进程持有的 fcntl 锁。
// F_GETLK 只读查询、绝不实际加锁，不会干扰目标应用打开数据库。
func fcntlLockHeld(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	lk := syscall.Flock_t{Type: syscall.F_WRLCK, Whence: 0}
	if err := syscall.FcntlFlock(f.Fd(), syscall.F_GETLK, &lk); err != nil {
		return false
	}
	return lk.Type != syscall.F_UNLCK && int(lk.Pid) != os.Getpid()
}

// runningChecker 判断缓存所属应用是否正在运行。运行中的应用缓存不进入快速清理（安全项），
// 仅在深度清理中展示并由用户手动勾选，避免应用正在使用时缓存被抽走导致页面重载。
// 全部检测手段（sysctl 进程表、fcntl 锁查询、SingletonLock）均不依赖外部命令，在 App Sandbox 内同样有效。
type runningChecker struct {
	procNames map[string]bool // 进程名（内核 p_comm，最长 16 字符）
	procsOK   bool
}

func newRunningChecker() *runningChecker {
	names, ok := processNames()
	return &runningChecker{procNames: names, procsOK: ok}
}

// chromiumRunning 判断 Chromium/Electron 应用是否运行中：SingletonLock 指向存活进程，
// 或用户数据根目录（及其下一级 Profile 目录）中任一 LevelDB LOCK 文件被持有
func (rc *runningChecker) chromiumRunning(userDataRoot string) bool {
	if chromiumLockHeld(userDataRoot) {
		return true
	}
	bases := []string{userDataRoot}
	if entries, err := os.ReadDir(userDataRoot); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				bases = append(bases, filepath.Join(userDataRoot, e.Name()))
			}
		}
	}
	for _, base := range bases {
		for _, sub := range leveldbLockSubpaths {
			if fcntlLockHeld(filepath.Join(base, sub)) {
				return true
			}
		}
	}
	return false
}

// processRunning 判断是否存在指定进程名的进程；进程表不可读时保守视为运行中
func (rc *runningChecker) processRunning(name string) bool {
	if !rc.procsOK {
		return true
	}
	return rc.procNames[name]
}
