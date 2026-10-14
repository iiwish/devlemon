package probe

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"devlemon/internal/config"
	"devlemon/internal/i18n"
	"devlemon/internal/model"
)

// AndroidEmulatorProbe Android 模拟器 (AVD) 的数据：运行中的模拟器加锁保护；已关机的模拟器拆为两项——
// 快速启动快照（可重建，只影响下次启动速度）与「重置数据」（等同 Android Studio 的 Wipe Data，会丢失模拟器内的应用与数据）
type AndroidEmulatorProbe struct{}

func NewAndroidEmulatorProbe() *AndroidEmulatorProbe {
	return &AndroidEmulatorProbe{}
}

func (p *AndroidEmulatorProbe) Name() string {
	return "Android Emulator Active Shield Probe"
}

func (p *AndroidEmulatorProbe) Category() model.Category {
	return model.CategorySimulator
}

// avdInfo 一个 Android 虚拟设备
type avdInfo struct {
	name    string // AVD 名称（即 emulator -avd 使用的名称，来自 <name>.ini）
	dir     string // 数据目录 <name>.avd
	display string // 显示名称（config.ini 中的 avd.ini.displayname）
}

// androidAVDHome AVD 注册目录：$ANDROID_AVD_HOME，否则 $ANDROID_EMULATOR_HOME/avd，否则 ~/.android/avd
func androidAVDHome(home string) string {
	if d := os.Getenv("ANDROID_AVD_HOME"); d != "" {
		return d
	}
	if d := os.Getenv("ANDROID_EMULATOR_HOME"); d != "" {
		return filepath.Join(d, "avd")
	}
	return filepath.Join(home, ".android", "avd")
}

// listAVDs 读取 AVD 注册目录下的 <name>.ini，按其中的 path= 定位数据目录（可能不在注册目录内），
// 找不到时回退到注册目录下的 <name>.avd；另外包含没有 .ini 的孤立 .avd 目录
func listAVDs(home string) []avdInfo {
	avdHome := androidAVDHome(home)
	inis, _ := filepath.Glob(filepath.Join(avdHome, "*.ini"))
	sort.Strings(inis)
	var avds []avdInfo
	for _, ini := range inis {
		name := strings.TrimSuffix(filepath.Base(ini), ".ini")
		kv := readINI(ini)
		dir := kv["path"]
		if dir == "" || !dirExists(dir) {
			dir = filepath.Join(avdHome, name+".avd")
		}
		if !dirExists(dir) {
			continue
		}
		display := readINI(filepath.Join(dir, "config.ini"))["avd.ini.displayname"]
		if display == "" {
			display = strings.ReplaceAll(name, "_", " ")
		}
		avds = append(avds, avdInfo{name: name, dir: dir, display: display})
	}
	// 缺少 .ini 的孤立 .avd 目录同样计入：保守起见，其引用的系统镜像不视为未使用，数据也可正常清理
	seen := map[string]bool{}
	for _, a := range avds {
		seen[filepath.Clean(a.dir)] = true
	}
	dirs, _ := filepath.Glob(filepath.Join(avdHome, "*.avd"))
	sort.Strings(dirs)
	for _, dir := range dirs {
		if seen[filepath.Clean(dir)] || !dirExists(dir) {
			continue
		}
		name := strings.TrimSuffix(filepath.Base(dir), ".avd")
		display := readINI(filepath.Join(dir, "config.ini"))["avd.ini.displayname"]
		if display == "" {
			display = strings.ReplaceAll(name, "_", " ")
		}
		avds = append(avds, avdInfo{name: name, dir: dir, display: display})
	}
	return avds
}

// readINI 读取 key=value 格式文件（AVD 的 .ini / config.ini），读不到时返回空 map
func readINI(path string) map[string]string {
	kv := map[string]string{}
	f, err := os.Open(path)
	if err != nil {
		return kv
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if k, v, ok := strings.Cut(sc.Text(), "="); ok {
			kv[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return kv
}

// isEmulatorExe 模拟器进程：emulator 启动器与 qemu-system-* 内核（含 -headless 变体）
func isEmulatorExe(exe string) bool {
	base := filepath.Base(exe)
	return base == "emulator" || strings.HasPrefix(base, "qemu-system-")
}

// emulatorProcesses 读取模拟器进程的参数（测试中可替换，避免依赖宿主机上真实运行的模拟器）
var emulatorProcesses = func() ([][]string, bool) { return processArgs(isEmulatorExe) }

// runningAVDNames 从模拟器进程参数中解析正在运行的 AVD（-avd <name> 或 @<name>）。
// anyEmulator 表示是否有模拟器进程在运行；ok=false 表示进程信息不可读
func runningAVDNames() (names map[string]bool, anyEmulator, ok bool) {
	procs, ok := emulatorProcesses()
	if !ok {
		return nil, false, false
	}
	names = map[string]bool{}
	for _, args := range procs {
		anyEmulator = true
		for i, a := range args {
			if a == "-avd" && i+1 < len(args) {
				names[args[i+1]] = true
			}
			if strings.HasPrefix(a, "@") && len(a) > 1 {
				names[a[1:]] = true
			}
		}
	}
	return names, anyEmulator, true
}

// hasEmulatorLock 运行中的模拟器会在数据目录下创建 *.lock（如 hardware-qemu.ini.lock、multiinstance.lock）
func hasEmulatorLock(dir string) bool {
	locks, _ := filepath.Glob(filepath.Join(dir, "*.lock"))
	return len(locks) > 0
}

// avdWipePrefixes「Wipe Data」会删除的文件：用户数据、缓存分区与对应的加密密钥（含 .qcow2 增量层）
var avdWipePrefixes = []string{"userdata-qemu.img", "cache.img", "encryptionkey.img"}

const avdQuickBootSnapshot = "default_boot"

// avdSize 模拟器磁盘镜像是稀疏文件，必须按实际占用的磁盘块统计（测试中可替换）
var avdSize = allocatedSize

func (p *AndroidEmulatorProbe) Scan(ctx context.Context, cfg *config.Config) (*model.Group, error) {
	group := &model.Group{
		ID:       "android_emulators",
		Title:    i18n.T("🤖 Android Emulators (Active Shield Protected)", "🤖 Android 模拟器（带运行态避让保护）"),
		Category: model.CategorySimulator,
		Items:    make([]*model.Item, 0),
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return group, nil
	}
	avds := listAVDs(home)
	if len(avds) == 0 {
		return group, nil
	}
	running, anyEmulator, procsOK := runningAVDNames()

	add := func(item *model.Item) {
		if item == nil {
			return
		}
		item.Category = model.CategorySimulator
		group.Items = append(group.Items, item)
		group.TotalSizeBytes += item.SizeBytes
		if !item.IsProtected {
			group.TotalReclaimableBytes += item.SizeBytes
		}
	}

	for _, avd := range avds {
		if ctx.Err() != nil {
			break
		}
		// 进程表不可读时一律视为运行中；有模拟器进程但参数未能对应到 AVD 时，以锁文件为准
		if !procsOK || running[avd.name] || (anyEmulator && hasEmulatorLock(avd.dir)) {
			size := avdSize(avd.dir)
			add(&model.Item{
				ID:            "avd_" + avd.name,
				Title:         fmt.Sprintf(i18n.T("Android Emulator: %s (Running)", "Android 模拟器: %s（运行中）"), avd.display),
				Description:   i18n.T("The emulator is running, active safety shield applied", "模拟器正在运行，已开启安全避让保护"),
				Path:          avd.dir,
				SizeBytes:     size,
				SizeFormatted: model.FormatBytes(size),
				Risk:          model.RiskCaution,
				IsProtected:   true,
				ProtectReason: i18n.T("🟢 Running, locked by safety shield to prevent interruption", "🟢 正在运行中，已自动加锁保护，绝不中断调试"),
			})
			continue
		}

		snapshotsDir := filepath.Join(avd.dir, "snapshots")
		add(newSizedItem("avd_snapshot_"+avd.name,
			fmt.Sprintf("%s (%s)", i18n.T("Android Emulator Quick Boot Snapshot", "Android 模拟器快速启动快照"), avd.display),
			i18n.T("Saved state used by Quick Boot. The next launch cold-boots instead (slower); apps and data in the emulator are kept.",
				"快速启动 (Quick Boot) 保存的运行状态；删除后下次启动改为冷启动（稍慢），模拟器内的应用与数据都会保留"),
			[]string{filepath.Join(snapshotsDir, avdQuickBootSnapshot)}, 50*1024*1024, model.RiskRebuildable, avdSize))

		var wipe []string
		entries, _ := os.ReadDir(avd.dir)
		for _, e := range entries {
			for _, prefix := range avdWipePrefixes {
				if strings.HasPrefix(e.Name(), prefix) && !strings.HasSuffix(e.Name(), ".lock") {
					wipe = append(wipe, filepath.Join(avd.dir, e.Name()))
				}
			}
		}
		// 手动保存的快照依赖当前用户数据，重置数据后一并失效
		snaps, _ := os.ReadDir(snapshotsDir)
		for _, s := range snaps {
			if s.Name() != avdQuickBootSnapshot {
				wipe = append(wipe, filepath.Join(snapshotsDir, s.Name()))
			}
		}
		add(newSizedItem("avd_wipe_"+avd.name,
			fmt.Sprintf("%s (%s)", i18n.T("Wipe Android Emulator Data", "重置 Android 模拟器数据"), avd.display),
			i18n.T("Same as \"Wipe Data\" in Android Studio's Device Manager: apps, accounts and files inside the emulator and manually saved snapshots are removed; the device configuration is kept and it starts in its factory state. Select the Quick Boot snapshot too.",
				"等同于 Android Studio 设备管理器中的「Wipe Data」：删除模拟器内安装的应用、登录账号、文件以及手动保存的快照，保留设备配置，下次启动恢复出厂状态。建议同时勾选该模拟器的快速启动快照"),
			wipe, 100*1024*1024, model.RiskCaution, avdSize))
	}
	return group, nil
}
