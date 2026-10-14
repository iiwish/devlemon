package probe

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"devlemon/internal/config"
)

// makeAVD 按 Android 模拟器的真实布局构造一个 AVD：<avdHome>/<name>.ini 通过 path= 指向数据目录
func makeAVD(t *testing.T, avdHome, dataDir, name string, files map[string]int64) {
	t.Helper()
	mkdir(t, avdHome)
	ini := "avd.ini.encoding=UTF-8\npath=" + dataDir + "\npath.rel=avd/" + name + ".avd\ntarget=android-36\n"
	if err := os.WriteFile(filepath.Join(avdHome, name+".ini"), []byte(ini), 0o644); err != nil {
		t.Fatal(err)
	}
	mkdir(t, dataDir)
	cfgIni := "avd.ini.displayname=" + strings.ReplaceAll(name, "_", " ") + " Pro\nimage.sysdir.1=system-images/android-36/google_apis/arm64-v8a/\n"
	if err := os.WriteFile(filepath.Join(dataDir, "config.ini"), []byte(cfgIni), 0o644); err != nil {
		t.Fatal(err)
	}
	for f, size := range files {
		sparse(t, filepath.Join(dataDir, f), size)
	}
}

// scanAndroid 用稀疏文件构造的大文件按逻辑长度统计，避免测试真的写入数 GB 数据
func scanAndroid(t *testing.T, home string) map[string][]string {
	t.Helper()
	orig := avdSize
	avdSize = logicalSize
	t.Cleanup(func() { avdSize = orig })
	t.Setenv("HOME", home)
	t.Setenv("ANDROID_AVD_HOME", "")
	t.Setenv("ANDROID_EMULATOR_HOME", "")
	g, err := NewAndroidEmulatorProbe().Scan(context.Background(), &config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	out := map[string][]string{}
	for _, it := range g.Items {
		key := it.ID
		if it.IsProtected {
			key += " [protected]"
		}
		out[key] = append([]string{string(it.Risk)}, it.CleanPaths...)
	}
	return out
}

func TestAndroidEmulatorItems(t *testing.T) {
	home := t.TempDir()
	avdHome := filepath.Join(home, ".android", "avd")
	// 数据目录在 AVD 注册目录之外（通过 .ini 的 path= 指定）
	external := filepath.Join(home, "AVDs", "Pixel_9.avd")
	makeAVD(t, avdHome, external, "Pixel_9", map[string]int64{
		"userdata-qemu.img":              800 * mb,
		"userdata-qemu.img.qcow2":        200 * mb,
		"cache.img":                      60 * mb,
		"encryptionkey.img":              18 * mb,
		"sdcard.img":                     512 * mb, // 共享存储，不属于 Wipe Data
		"snapshots/default_boot/ram.bin": 900 * mb,
		"snapshots/before_login/ram.bin": 700 * mb, // 手动保存的快照
		"hardware-qemu.ini":              1,
	})
	// 没有 .ini 的孤立 AVD 目录
	makeAVD(t, filepath.Join(home, "tmp-unused"), filepath.Join(avdHome, "Orphan.avd"), "Orphan",
		map[string]int64{"userdata-qemu.img": 300 * mb})

	items := scanAndroid(t, home)
	wipe := items["avd_wipe_Pixel_9"]
	if wipe == nil || wipe[0] != "caution" {
		t.Fatalf("wipe item missing or wrong risk: %v", items)
	}
	gotWipe := map[string]bool{}
	for _, p := range wipe[1:] {
		gotWipe[filepath.Base(p)] = true
	}
	for _, want := range []string{"userdata-qemu.img", "userdata-qemu.img.qcow2", "cache.img", "encryptionkey.img", "before_login"} {
		if !gotWipe[want] {
			t.Errorf("wipe item should remove %s: %v", want, wipe)
		}
	}
	for _, keep := range []string{"sdcard.img", "config.ini", "default_boot", "hardware-qemu.ini"} {
		if gotWipe[keep] {
			t.Errorf("wipe item must not remove %s", keep)
		}
	}
	snap := items["avd_snapshot_Pixel_9"]
	if len(snap) != 2 || snap[0] != "rebuildable" || filepath.Base(snap[1]) != "default_boot" {
		t.Errorf("snapshot item = %v", snap)
	}
	if items["avd_wipe_Orphan"] == nil {
		t.Errorf("orphan .avd directory should be listed: %v", items)
	}
}

// TestHelperEmulatorProcess 被下面的测试以 qemu-system-aarch64 的名字重新启动，模拟正在运行的模拟器
func TestHelperEmulatorProcess(t *testing.T) {
	if os.Getenv("DEVLEMON_FAKE_EMULATOR") != "1" {
		t.Skip("helper process")
	}
	time.Sleep(30 * time.Second)
}

// startFakeEmulator 把测试二进制复制为 qemu-system-aarch64 并带上 -avd 参数启动，
// 让运行检测走真实的系统进程表（macOS kern.procargs2 / Linux /proc）
func startFakeEmulator(t *testing.T, args ...string) {
	t.Helper()
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("process table not supported")
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(self)
	if err != nil {
		t.Fatal(err)
	}
	fake := filepath.Join(t.TempDir(), "qemu-system-aarch64")
	if err := os.WriteFile(fake, data, 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(fake, append([]string{"-test.run=^TestHelperEmulatorProcess$", "--"}, args...)...)
	cmd.Env = append(os.Environ(), "DEVLEMON_FAKE_EMULATOR=1")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	// 等待进程出现在进程表中
	for i := 0; i < 50; i++ {
		if procs, _ := processArgs(isEmulatorExe); len(procs) > 0 {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("fake emulator process not visible")
}

func TestRunningEmulatorIsProtected(t *testing.T) {
	home := t.TempDir()
	avdHome := filepath.Join(home, ".android", "avd")
	files := map[string]int64{"userdata-qemu.img": 800 * mb, "snapshots/default_boot/ram.bin": 900 * mb}
	makeAVD(t, avdHome, filepath.Join(avdHome, "Running_Phone.avd"), "Running_Phone", files)
	makeAVD(t, avdHome, filepath.Join(avdHome, "Idle_Phone.avd"), "Idle_Phone", files)

	startFakeEmulator(t, "-avd", "Running_Phone", "-no-window")

	items := scanAndroid(t, home)
	if items["avd_Running_Phone [protected]"] == nil {
		t.Fatalf("running emulator must be protected: %v", items)
	}
	if items["avd_wipe_Running_Phone"] != nil || items["avd_snapshot_Running_Phone"] != nil {
		t.Errorf("running emulator must not offer cleanable items: %v", items)
	}
	if items["avd_wipe_Idle_Phone"] == nil || items["avd_snapshot_Idle_Phone"] == nil {
		t.Errorf("idle emulator should still be cleanable: %v", items)
	}
}

func TestEmulatorLockFallback(t *testing.T) {
	home := t.TempDir()
	avdHome := filepath.Join(home, ".android", "avd")
	files := map[string]int64{"userdata-qemu.img": 800 * mb, "hardware-qemu.ini.lock": 1}
	makeAVD(t, avdHome, filepath.Join(avdHome, "Locked.avd"), "Locked", files)

	orig := emulatorProcesses
	t.Cleanup(func() { emulatorProcesses = orig })
	stub := func(procs [][]string, ok bool) {
		emulatorProcesses = func() ([][]string, bool) { return procs, ok }
	}

	// 没有任何模拟器进程时，残留的锁文件不应让 AVD 永远处于保护状态
	stub(nil, true)
	if items := scanAndroid(t, home); items["avd_wipe_Locked"] == nil {
		t.Errorf("stale lock without emulator process should not protect: %v", items)
	}
	// 有模拟器进程但参数未对应到 AVD 时，以锁文件为准
	stub([][]string{{"qemu-system-aarch64", "-some-other-flag"}}, true)
	if items := scanAndroid(t, home); items["avd_Locked [protected]"] == nil {
		t.Errorf("locked AVD with a running emulator must be protected: %v", items)
	}
	// 进程表不可读时一律视为运行中
	stub(nil, false)
	if items := scanAndroid(t, home); items["avd_Locked [protected]"] == nil {
		t.Errorf("unreadable process table must protect all AVDs: %v", items)
	}
}

func TestParseProcArgs2(t *testing.T) {
	// argc=3 | exe\0 | 填充 \0\0 | argv... | env
	buf := []byte{3, 0, 0, 0}
	buf = append(buf, []byte("/sdk/emulator/qemu/darwin-aarch64/qemu-system-aarch64\x00\x00\x00")...)
	buf = append(buf, []byte("qemu-system-aarch64\x00@Pixel_9\x00-no-window\x00HOME=/Users/me\x00")...)
	exe, args := parseProcArgs2(buf)
	if filepath.Base(exe) != "qemu-system-aarch64" || len(args) != 3 || args[1] != "@Pixel_9" {
		t.Errorf("parseProcArgs2 = %q %q", exe, args)
	}
	if exe, args := parseProcArgs2([]byte{1, 0}); exe != "" || args != nil {
		t.Error("short buffer must be rejected")
	}
}

// 实机验证中发现：模拟器的 userdata-qemu.img 逻辑长度 10GB，实际只占用写入过的部分。
// 统计必须按实际占用的磁盘块，否则可释放空间会被夸大数十倍
func TestAVDSizeUsesAllocatedBlocks(t *testing.T) {
	home := t.TempDir()
	avdHome := filepath.Join(home, ".android", "avd")
	dir := filepath.Join(avdHome, "Sparse.avd")
	makeAVD(t, avdHome, dir, "Sparse", map[string]int64{"userdata-qemu.img": 10 << 30})
	// 真实写入 120MB 数据
	f, err := os.OpenFile(filepath.Join(dir, "userdata-qemu.img"), os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	chunk := make([]byte, 1<<20)
	for i := range chunk {
		chunk[i] = byte(i)
	}
	for i := 0; i < 120; i++ {
		if _, err := f.Write(chunk); err != nil {
			t.Fatal(err)
		}
	}
	f.Close()

	t.Setenv("HOME", home)
	t.Setenv("ANDROID_AVD_HOME", "")
	t.Setenv("ANDROID_EMULATOR_HOME", "")
	g, err := NewAndroidEmulatorProbe().Scan(context.Background(), &config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range g.Items {
		if it.ID == "avd_wipe_Sparse" {
			if it.SizeBytes < 120<<20 || it.SizeBytes > 200<<20 {
				t.Fatalf("wipe size = %s, want about 120MB (allocated), not the 10GB logical length", it.SizeFormatted)
			}
			return
		}
	}
	t.Fatalf("wipe item missing: %+v", g.Items)
}
