package probe

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func mkfile(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mkdir(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestIsChromiumCacheDir(t *testing.T) {
	root := t.TempDir()

	// 现代 HTTP 缓存：Cache/Cache_Data 为 Simple Cache 结构
	simple := filepath.Join(root, "a", "Cache")
	mkfile(t, filepath.Join(simple, "Cache_Data", "index"))
	mkdir(t, filepath.Join(simple, "Cache_Data", "index-dir"))

	// V8 字节码缓存
	code := filepath.Join(root, "a", "Code Cache")
	mkfile(t, filepath.Join(code, "js", "index"))
	mkdir(t, filepath.Join(code, "js", "index-dir"))

	// GPU 缓存：Blockfile 结构
	gpu := filepath.Join(root, "a", "GPUCache")
	mkfile(t, filepath.Join(gpu, "index"))
	mkfile(t, filepath.Join(gpu, "data_0"))

	// 恰好同名但并非 Chromium 缓存的业务数据目录
	fakeCache := filepath.Join(root, "b", "Cache")
	mkfile(t, filepath.Join(fakeCache, "index"))
	mkfile(t, filepath.Join(fakeCache, "user-notes.db"))
	fakeCode := filepath.Join(root, "b", "Code Cache")
	mkdir(t, filepath.Join(fakeCode, "js"))
	fakeGPU := filepath.Join(root, "b", "GPUCache")
	mkfile(t, filepath.Join(fakeGPU, "index"))

	for _, c := range []struct {
		dir  string
		want bool
	}{
		{simple, true}, {code, true}, {gpu, true},
		{fakeCache, false}, {fakeCode, false}, {fakeGPU, false},
	} {
		if got := isChromiumCacheDir(c.dir); got != c.want {
			t.Errorf("isChromiumCacheDir(%s) = %v, want %v", c.dir, got, c.want)
		}
	}

	if got := collectChromiumCaches(filepath.Join(root, "b")); len(got) != 0 {
		t.Errorf("collectChromiumCaches on fake dirs = %v, want none", got)
	}
}

func TestCollectChromiumCachesSkipsSymlinks(t *testing.T) {
	root := t.TempDir()
	real := filepath.Join(root, "elsewhere")
	mkfile(t, filepath.Join(real, "index"))
	mkfile(t, filepath.Join(real, "data_0"))
	app := filepath.Join(root, "app")
	mkdir(t, app)
	if err := os.Symlink(real, filepath.Join(app, "GPUCache")); err != nil {
		t.Fatal(err)
	}
	if got := collectChromiumCaches(app); len(got) != 0 {
		t.Errorf("symlinked cache dir must be ignored, got %v", got)
	}
}

func TestChromiumLockHeld(t *testing.T) {
	root := t.TempDir()
	if chromiumLockHeld(root) {
		t.Fatal("no lock should mean not running")
	}

	lock := filepath.Join(root, "SingletonLock")
	link := func(target string) {
		_ = os.Remove(lock)
		if err := os.Symlink(target, lock); err != nil {
			t.Fatal(err)
		}
	}

	link(fmt.Sprintf("host.local-%d", os.Getpid()))
	if !chromiumLockHeld(root) {
		t.Error("lock pointing to a live pid should mean running")
	}

	link("host.local-garbage")
	if !chromiumLockHeld(root) {
		t.Error("unparsable lock must conservatively mean running")
	}

	link("host.local-99999999")
	if chromiumLockHeld(root) {
		t.Error("lock pointing to a dead pid should mean not running")
	}
}

func TestIsStale(t *testing.T) {
	root := t.TempDir()
	old := time.Now().Add(-30 * 24 * time.Hour)
	cutoff := time.Now().Add(-tempFileMinAge)

	oldDir := filepath.Join(root, "old")
	mkfile(t, filepath.Join(oldDir, "a.tmp"))
	_ = os.Chtimes(filepath.Join(oldDir, "a.tmp"), old, old)
	_ = os.Chtimes(oldDir, old, old)
	if !isStale(oldDir, cutoff) {
		t.Error("fully old dir should be stale")
	}

	// 顶层目录很旧，但内部有近期修改的文件：必须跳过
	mixed := filepath.Join(root, "mixed")
	mkfile(t, filepath.Join(mixed, "fresh.tmp"))
	_ = os.Chtimes(mixed, old, old)
	if isStale(mixed, cutoff) {
		t.Error("dir containing a recently modified file must not be stale")
	}
}

// TestHelperHoldLock 子进程辅助：对指定文件加 fcntl 写锁并保持，直到 stdin 关闭
func TestHelperHoldLock(t *testing.T) {
	path := os.Getenv("DEVLEMON_HOLD_LOCK")
	if path == "" {
		t.Skip("helper process only")
	}
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		os.Exit(2)
	}
	lk := syscall.Flock_t{Type: syscall.F_WRLCK, Whence: 0}
	if err := syscall.FcntlFlock(f.Fd(), syscall.F_SETLK, &lk); err != nil {
		os.Exit(3)
	}
	fmt.Println("locked")
	_, _ = io.ReadAll(os.Stdin)
	os.Exit(0)
}

func TestChromiumRunningByLevelDBLock(t *testing.T) {
	root := t.TempDir()
	lockFile := filepath.Join(root, "Default", "Local Storage", "leveldb", "LOCK")
	mkfile(t, lockFile)

	rc := newRunningChecker()
	if rc.chromiumRunning(root) {
		t.Fatal("unlocked LevelDB LOCK should mean not running")
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperHoldLock$")
	cmd.Env = append(os.Environ(), "DEVLEMON_HOLD_LOCK="+lockFile)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { stdin.Close(); _ = cmd.Wait() }()
	if line, _ := bufio.NewReader(stdout).ReadString('\n'); !strings.HasPrefix(line, "locked") {
		t.Fatalf("helper failed to lock: %q", line)
	}

	if !rc.chromiumRunning(root) {
		t.Error("LevelDB LOCK held by another process should mean running")
	}
	// F_GETLK 只查询不加锁：检测之后锁仍属于子进程，本进程无法再加锁
	f, _ := os.OpenFile(lockFile, os.O_RDWR, 0)
	defer f.Close()
	lk := syscall.Flock_t{Type: syscall.F_WRLCK, Whence: 0}
	if err := syscall.FcntlFlock(f.Fd(), syscall.F_SETLK, &lk); err == nil {
		t.Error("lock should still be held by the helper process")
	}
}

func TestProcessRunning(t *testing.T) {
	rc := newRunningChecker()
	if !rc.procsOK {
		t.Skip("process table unavailable")
	}
	if rc.processRunning("definitely-not-a-process-xyz") {
		t.Error("unknown process should not be running")
	}
	if !(&runningChecker{}).processRunning("Safari") {
		t.Error("unreadable process table must conservatively mean running")
	}
}
