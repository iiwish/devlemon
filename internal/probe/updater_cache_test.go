package probe

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

func TestUpdaterCacheTargets(t *testing.T) {
	caches := t.TempDir()
	apps := t.TempDir()

	// Squirrel ShipIt：状态文件指向已安装应用与暂存的新版本
	installed := filepath.Join(apps, "Visual Studio Code.app")
	writePlistVersion(t, installed, "1.136.0")
	shipit := filepath.Join(caches, "com.microsoft.VSCode.ShipIt")
	staged := filepath.Join(shipit, "update.abc", "Visual Studio Code.app")
	writePlistVersion(t, staged, "1.141.0")
	state := `{"updateBundleURL":"file://` + staged + `/","targetBundleURL":"file://` + installed + `/"}`
	if err := os.WriteFile(filepath.Join(shipit, "ShipItState.plist"), []byte(state), 0o644); err != nil {
		t.Fatal(err)
	}

	// Sparkle：只认 org.sparkle-project.Sparkle 子目录，同级的其他缓存不算
	mkdir(t, filepath.Join(caches, "com.steipete.codexbar", "org.sparkle-project.Sparkle", "Installation"))
	mkfile(t, filepath.Join(caches, "com.steipete.codexbar", "other", "data"))

	// electron-updater：pending/ 必须带 update-info.json，否则可能是同名业务目录
	mkfile(t, filepath.Join(caches, "com.google.antigravity", "pending", "update-info.json"))
	mkfile(t, filepath.Join(caches, "com.example.todo", "pending", "task.json"))
	mkdir(t, filepath.Join(caches, "my-app-updater", "pending"))

	got := map[string]updaterTarget{}
	for _, tg := range updaterCacheTargets(caches) {
		got[tg.path] = tg
	}
	var paths []string
	for p := range got {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	want := []string{
		filepath.Join(caches, "com.google.antigravity", "pending"),
		shipit,
		filepath.Join(caches, "com.steipete.codexbar", "org.sparkle-project.Sparkle"),
		filepath.Join(caches, "my-app-updater"),
	}
	sort.Strings(want)
	if len(paths) != len(want) {
		t.Fatalf("targets = %v, want %v", paths, want)
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Fatalf("targets = %v, want %v", paths, want)
		}
	}

	s := got[shipit]
	if s.appPath != installed || s.installedVersion != "1.136.0" || s.stagedVersion != "1.141.0" {
		t.Errorf("ShipIt state = %+v", s)
	}
	if s.displayName() != "Visual Studio Code" {
		t.Errorf("displayName = %q", s.displayName())
	}
	if got[filepath.Join(caches, "my-app-updater")].nameHint != "my-app" {
		t.Errorf("electron-updater nameHint not derived from dir name")
	}

	claimed := updaterClaimed(caches)
	if !claimed[shipit] || claimed[filepath.Join(caches, "com.steipete.codexbar")] {
		t.Errorf("claimed = %v", claimed)
	}
}

func writePlistVersion(t *testing.T, app, version string) {
	t.Helper()
	p := filepath.Join(app, "Contents", "Info.plist")
	mkdir(t, filepath.Dir(p))
	plist := `<?xml version="1.0" encoding="UTF-8"?><plist version="1.0"><dict>
	<key>CFBundleIdentifier</key>
	<string>com.example.app</string>
	<key>CFBundleShortVersionString</key>
	<string>` + version + `</string></dict></plist>`
	if err := os.WriteFile(p, []byte(plist), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestAppRunningMatching(t *testing.T) {
	rc := &runningChecker{appsOK: true, apps: []runningApp{
		{path: "/Applications/Visual Studio Code.app", bundleID: "com.microsoft.VSCode", nameKey: appNameKey("Visual Studio Code")},
		{path: "/Applications/CodexBar.app", nameKey: appNameKey("CodexBar")},
	}}
	cases := []struct {
		bundleID, appPath, hint string
		want                    bool
	}{
		{"com.microsoft.vscode", "", "vscode", true},           // Bundle ID 忽略大小写
		{"", "/Applications/Visual Studio Code.app", "", true}, // .app 路径
		{"com.steipete.codexbar", "", "codexbar", true},        // 二进制 plist 读不到 ID 时按名称匹配
		{"com.google.antigravity", "", "antigravity", false},   // 未运行
		{"", "", "ab", false},                                  // 过短的名称不参与匹配
	}
	for _, c := range cases {
		if got := rc.appRunning(c.bundleID, c.appPath, c.hint); got != c.want {
			t.Errorf("appRunning(%q, %q, %q) = %v, want %v", c.bundleID, c.appPath, c.hint, got, c.want)
		}
	}
	if !(&runningChecker{}).appRunning("x", "", "anything") {
		t.Error("unreadable process list must be treated as running")
	}
}

func TestOuterAppBundle(t *testing.T) {
	exe := "/Applications/Visual Studio Code.app/Contents/Frameworks/Code Helper.app/Contents/MacOS/Code Helper"
	if got := outerAppBundle(exe); got != "/Applications/Visual Studio Code.app" {
		t.Errorf("outerAppBundle = %q", got)
	}
	if got := outerAppBundle("/usr/bin/ssh"); got != "" {
		t.Errorf("outerAppBundle = %q", got)
	}
}

func TestModifiedWithin(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "a", "b", "update.zip")
	mkfile(t, f)
	old := time.Now().Add(-time.Hour)
	for _, p := range []string{f, filepath.Dir(f), filepath.Join(dir, "a"), dir} {
		if err := os.Chtimes(p, old, old); err != nil {
			t.Fatal(err)
		}
	}
	if modifiedWithin(dir, 10*time.Minute) {
		t.Error("old tree reported as recently modified")
	}
	now := time.Now()
	if err := os.Chtimes(f, now, now); err != nil {
		t.Fatal(err)
	}
	if !modifiedWithin(dir, 10*time.Minute) {
		t.Error("download in progress not detected")
	}
}
