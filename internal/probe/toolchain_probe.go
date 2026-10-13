package probe

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"devlemon/internal/config"
	"devlemon/internal/i18n"
	"devlemon/internal/model"
)

// ToolchainProbe 开发工具链与 SDK 中不再使用的版本：旧版 pnpm 存储、rustup 固定版本工具链、
// 未被模拟器引用的 Android 系统镜像、旧版 NDK / build-tools、旧版 Gradle Wrapper。
// 这些都可以按需重新下载，但需要先判断是否仍在使用，因此单独成组，风险等级为 rebuildable。
type ToolchainProbe struct{}

func NewToolchainProbe() *ToolchainProbe {
	return &ToolchainProbe{}
}

func (p *ToolchainProbe) Name() string {
	return "Developer Toolchains & SDKs Probe"
}

func (p *ToolchainProbe) Category() model.Category {
	return model.CategoryPackageCache
}

// toolchainMinIdle 最近仍有改动的版本视为在用，绝不列入
const toolchainMinIdle = 30 * 24 * time.Hour

func (p *ToolchainProbe) Scan(ctx context.Context, cfg *config.Config) (*model.Group, error) {
	group := &model.Group{
		ID:       "dev_toolchains",
		Title:    i18n.T("🧰 Developer Toolchains & SDKs", "🧰 开发工具链与 SDK 旧版本"),
		Category: model.CategoryPackageCache,
		Items:    make([]*model.Item, 0),
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return group, nil
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	add := func(items ...*model.Item) {
		mu.Lock()
		defer mu.Unlock()
		for _, item := range items {
			if item == nil {
				continue
			}
			item.Category = model.CategoryPackageCache
			group.Items = append(group.Items, item)
			group.TotalSizeBytes += item.SizeBytes
			group.TotalReclaimableBytes += item.SizeBytes
		}
	}
	for _, f := range []func(string) []*model.Item{
		pnpmOldStores, rustupToolchains, androidSDKItems, gradleWrapperDists,
	} {
		f := f
		wg.Add(1)
		go func() {
			defer wg.Done()
			if ctx.Err() != nil {
				return
			}
			add(f(home)...)
		}()
	}
	wg.Wait()
	return group, nil
}

// ---------- pnpm ----------

var pnpmStoreVersionRe = regexp.MustCompile(`^v(\d+)$`)

func pnpmStoreRoots(home string) []string {
	return []string{
		filepath.Join(home, "Library", "pnpm", "store"),
		filepath.Join(home, ".local", "share", "pnpm", "store"),
		filepath.Join(home, ".pnpm-store"),
	}
}

// pnpmOldStores 每个 pnpm 大版本使用独立的存储目录（v3 / v10 / v11…），升级后旧目录不会自动删除。
// 已安装项目的 node_modules 是从存储克隆或硬链接出去的独立文件，删除旧存储不影响它们。
func pnpmOldStores(home string) []*model.Item {
	var items []*model.Item
	for _, root := range pnpmStoreRoots(home) {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		type store struct {
			ver  int
			path string
		}
		var stores []store
		for _, e := range entries {
			if m := pnpmStoreVersionRe.FindStringSubmatch(e.Name()); m != nil && e.IsDir() {
				v, _ := strconv.Atoi(m[1])
				stores = append(stores, store{v, filepath.Join(root, e.Name())})
			}
		}
		if len(stores) < 2 {
			continue
		}
		sort.Slice(stores, func(i, j int) bool { return stores[i].ver > stores[j].ver })
		for _, s := range stores[1:] {
			if modifiedWithin(s.path, toolchainMinIdle) {
				continue
			}
			items = append(items, newCacheItem("pnpm_store_v"+strconv.Itoa(s.ver),
				fmt.Sprintf(i18n.T("pnpm Store from an Older pnpm (v%d)", "旧版 pnpm 遗留的存储 (v%d)"), s.ver),
				fmt.Sprintf(i18n.T("Package store of an older pnpm major version; the current store is v%d. Installed node_modules are independent copies and keep working; only reinstalling with the old pnpm downloads packages again.",
					"旧版 pnpm 使用的包存储，当前在用的是 v%d。已安装项目的 node_modules 是独立副本，不受影响；只有用旧版 pnpm 重新安装时才需要重新下载"), stores[0].ver),
				[]string{s.path}, 50*1024*1024, model.RiskRebuildable))
		}
	}
	return items
}

// ---------- rustup ----------

var (
	rustupDefaultRe  = regexp.MustCompile(`(?m)^\s*default_toolchain\s*=\s*"([^"]+)"`)
	rustupOverrideRe = regexp.MustCompile(`^\s*"[^"]+"\s*=\s*"([^"]+)"`)
)

// rustupToolchains 列出带版本号的工具链（通常由项目的 rust-toolchain.toml 固定后自动安装），
// 保留默认工具链、被 rustup override 绑定的工具链，以及 stable / beta / nightly 等频道工具链
func rustupToolchains(home string) []*model.Item {
	root := os.Getenv("RUSTUP_HOME")
	if root == "" {
		root = filepath.Join(home, ".rustup")
	}
	entries, err := os.ReadDir(filepath.Join(root, "toolchains"))
	if err != nil {
		return nil
	}
	keep := rustupKeptToolchains(filepath.Join(root, "settings.toml"))

	var items []*model.Item
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() || name == "" || name[0] < '0' || name[0] > '9' || rustupKept(name, keep) {
			continue
		}
		dir := filepath.Join(root, "toolchains", name)
		if modifiedWithin(dir, toolchainMinIdle) {
			continue
		}
		paths := []string{dir}
		if h := filepath.Join(root, "update-hashes", name); fileExists(h) {
			paths = append(paths, h)
		}
		items = append(items, newCacheItem("rustup_"+name,
			fmt.Sprintf("%s (%s)", i18n.T("Pinned Rust Toolchain", "固定版本 Rust 工具链"), name),
			i18n.T("Version-pinned toolchain, usually installed for a project's rust-toolchain.toml. Not your default toolchain; rustup reinstalls it automatically the next time a project needs it.",
				"按版本固定的工具链，通常是某个项目的 rust-toolchain.toml 指定后自动安装的；不是默认工具链，项目下次需要时 rustup 会自动重新安装"),
			paths, 50*1024*1024, model.RiskRebuildable))
	}
	return items
}

// rustupKeptToolchains 从 settings.toml 读取默认工具链与 [overrides] 中绑定的工具链；读不到时返回 nil
func rustupKeptToolchains(settings string) []string {
	f, err := os.Open(settings)
	if err != nil {
		return nil
	}
	defer f.Close()
	var keep []string
	inOverrides := false
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if t := strings.TrimSpace(line); strings.HasPrefix(t, "[") {
			inOverrides = t == "[overrides]"
			continue
		}
		if m := rustupDefaultRe.FindStringSubmatch(line); m != nil && !inOverrides {
			keep = append(keep, m[1])
		}
		if m := rustupOverrideRe.FindStringSubmatch(line); m != nil && inOverrides {
			keep = append(keep, m[1])
		}
	}
	return keep
}

// rustupKept 工具链目录名带有目标三元组（1.88.0-aarch64-apple-darwin），设置里可能只写了 1.88.0
func rustupKept(dir string, keep []string) bool {
	for _, k := range keep {
		if dir == k || strings.HasPrefix(dir, k+"-") {
			return true
		}
	}
	return false
}

// ---------- Android ----------

func androidSDKRoots(home string) []string {
	var roots []string
	for _, env := range []string{"ANDROID_HOME", "ANDROID_SDK_ROOT"} {
		if v := os.Getenv(env); v != "" {
			roots = append(roots, v)
		}
	}
	roots = append(roots, filepath.Join(home, "Library", "Android", "sdk"), filepath.Join(home, "Android", "Sdk"))
	seen := map[string]bool{}
	var out []string
	for _, r := range roots {
		r = filepath.Clean(r)
		if !seen[r] && dirExists(r) {
			seen[r] = true
			out = append(out, r)
		}
	}
	return out
}

func androidSDKItems(home string) []*model.Item {
	var items []*model.Item
	used := androidUsedSystemImages(home)
	for _, sdk := range androidSDKRoots(home) {
		items = append(items, androidUnusedSystemImages(sdk, used)...)
		items = append(items, olderVersionItems(filepath.Join(sdk, "ndk"), "android_ndk_",
			i18n.T("Older Android NDK", "旧版 Android NDK"),
			i18n.T("An older NDK version; the newest installed one is kept. Android Gradle Plugin downloads the version a project pins when it is missing.",
				"较旧的 NDK 版本（已保留最新版本）；项目指定的版本缺失时，Android Gradle 插件会自动下载"))...)
		items = append(items, olderVersionItems(filepath.Join(sdk, "build-tools"), "android_build_tools_",
			i18n.T("Older Android Build Tools", "旧版 Android Build Tools"),
			i18n.T("An older build-tools version; the newest installed one is kept. Downloaded again automatically if a project requires it.",
				"较旧的 build-tools 版本（已保留最新版本）；项目需要时会自动重新下载"))...)
	}
	return items
}

// androidUsedSystemImages 读取全部 AVD 的 config.ini，收集其引用的系统镜像（如 system-images/android-36/google_apis/arm64-v8a）
func androidUsedSystemImages(home string) map[string]bool {
	avdHome := os.Getenv("ANDROID_AVD_HOME")
	if avdHome == "" {
		avdHome = filepath.Join(home, ".android", "avd")
	}
	used := map[string]bool{}
	entries, _ := os.ReadDir(avdHome)
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".avd") {
			continue
		}
		f, err := os.Open(filepath.Join(avdHome, e.Name(), "config.ini"))
		if err != nil {
			continue
		}
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			k, v, ok := strings.Cut(sc.Text(), "=")
			if ok && strings.HasPrefix(strings.TrimSpace(k), "image.sysdir.") {
				used[strings.Trim(filepath.ToSlash(strings.TrimSpace(v)), "/")] = true
			}
		}
		f.Close()
	}
	return used
}

// androidUnusedSystemImages 系统镜像位于 system-images/<api>/<tag>/<abi>，未被任何 AVD 引用即可删除
func androidUnusedSystemImages(sdk string, used map[string]bool) []*model.Item {
	var items []*model.Item
	base := filepath.Join(sdk, "system-images")
	apis, _ := os.ReadDir(base)
	for _, api := range apis {
		tags, _ := os.ReadDir(filepath.Join(base, api.Name()))
		for _, tag := range tags {
			abis, _ := os.ReadDir(filepath.Join(base, api.Name(), tag.Name()))
			for _, abi := range abis {
				if !abi.IsDir() {
					continue
				}
				rel := strings.Join([]string{"system-images", api.Name(), tag.Name(), abi.Name()}, "/")
				if used[rel] {
					continue
				}
				dir := filepath.Join(base, api.Name(), tag.Name(), abi.Name())
				if modifiedWithin(dir, toolchainMinIdle) {
					continue
				}
				items = append(items, newCacheItem("android_sysimg_"+strings.ReplaceAll(rel, "/", "_"),
					fmt.Sprintf("%s (%s / %s / %s)", i18n.T("Unused Android System Image", "未使用的 Android 系统镜像"), api.Name(), tag.Name(), abi.Name()),
					i18n.T("Emulator system image not used by any Android Virtual Device. Re-download it from Android Studio's SDK Manager when you create an emulator with it.",
						"没有任何 Android 模拟器 (AVD) 使用的系统镜像；以后创建模拟器时可在 Android Studio 的 SDK Manager 中重新下载"),
					[]string{dir}, 100*1024*1024, model.RiskRebuildable))
			}
		}
	}
	return items
}

// olderVersionItems 目录下按版本号命名的子目录中，保留最新版本，其余 30 天未改动的逐个列出
func olderVersionItems(dir, idPrefix, title, desc string) []*model.Item {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var versions []string
	for _, e := range entries {
		if e.IsDir() && parseVersion(e.Name()) != nil {
			versions = append(versions, e.Name())
		}
	}
	if len(versions) < 2 {
		return nil
	}
	sort.Slice(versions, func(i, j int) bool { return compareVersions(versions[i], versions[j]) > 0 })
	var items []*model.Item
	for _, v := range versions[1:] {
		p := filepath.Join(dir, v)
		if modifiedWithin(p, toolchainMinIdle) {
			continue
		}
		items = append(items, newCacheItem(idPrefix+v, fmt.Sprintf("%s (%s)", title, v), desc,
			[]string{p}, 20*1024*1024, model.RiskRebuildable))
	}
	return items
}

// ---------- Gradle ----------

var gradleDistRe = regexp.MustCompile(`^gradle-(\d+(?:\.\d+)*)(?:-[a-z0-9]+)*-(?:bin|all)$`)

// gradleWrapperDists 保留最新的 Gradle Wrapper 发行版，其余 30 天未使用的列出（项目构建时 Wrapper 会自动重新下载）
func gradleWrapperDists(home string) []*model.Item {
	gradleHome := os.Getenv("GRADLE_USER_HOME")
	if gradleHome == "" {
		gradleHome = filepath.Join(home, ".gradle")
	}
	dir := filepath.Join(gradleHome, "wrapper", "dists")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	type dist struct{ name, version string }
	var dists []dist
	for _, e := range entries {
		if m := gradleDistRe.FindStringSubmatch(e.Name()); m != nil && e.IsDir() {
			dists = append(dists, dist{e.Name(), m[1]})
		}
	}
	if len(dists) < 2 {
		return nil
	}
	sort.Slice(dists, func(i, j int) bool { return compareVersions(dists[i].version, dists[j].version) > 0 })
	var items []*model.Item
	for _, d := range dists[1:] {
		if d.version == dists[0].version {
			continue // 同一版本的 bin 与 all 发行版都保留
		}
		p := filepath.Join(dir, d.name)
		if modifiedWithin(p, toolchainMinIdle) {
			continue
		}
		items = append(items, newCacheItem("gradle_dist_"+d.name,
			fmt.Sprintf("%s (%s)", i18n.T("Older Gradle Wrapper Distribution", "旧版 Gradle Wrapper 发行版"), d.name),
			i18n.T("Gradle distribution downloaded by a project's wrapper; the newest one is kept. The wrapper downloads it again if a project still needs it.",
				"项目 Wrapper 下载的 Gradle 发行版（已保留最新版本）；仍需要它的项目构建时会自动重新下载"),
			[]string{p}, 20*1024*1024, model.RiskRebuildable))
	}
	return items
}

// ---------- 版本号工具 ----------

// parseVersion 解析 1.2.3 / 27.0.12077973 / 26.1.10909125 等纯数字点分版本，非版本号返回 nil
func parseVersion(s string) []int {
	parts := strings.Split(s, ".")
	nums := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		nums = append(nums, n)
	}
	return nums
}

func compareVersions(a, b string) int {
	va, vb := parseVersion(a), parseVersion(b)
	for i := 0; i < len(va) || i < len(vb); i++ {
		var x, y int
		if i < len(va) {
			x = va[i]
		}
		if i < len(vb) {
			y = vb[i]
		}
		if x != y {
			if x > y {
				return 1
			}
			return -1
		}
	}
	return 0
}
