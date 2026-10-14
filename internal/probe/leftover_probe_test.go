package probe

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

func ageTree(t *testing.T, root string, age time.Duration) {
	t.Helper()
	old := time.Now().Add(-age)
	_ = filepath.Walk(root, func(p string, _ os.FileInfo, err error) error {
		if err == nil {
			_ = os.Chtimes(p, old, old)
		}
		return nil
	})
}

func TestFindLeftovers(t *testing.T) {
	home := t.TempDir()
	lib := filepath.Join(home, "Library")
	apps := filepath.Join(home, "Applications")
	writeInfoPlist(t, filepath.Join(apps, "Keep.app"), "com.keep.App")

	// 已卸载的 GUI 应用：偏好设置 + 数据目录 + 内含 Chromium 缓存（由安全缓存规则单独统计）
	mkfile(t, filepath.Join(lib, "Preferences", "com.gone.App.plist"))
	mkfile(t, filepath.Join(lib, "Application Support", "com.gone.App", "db.sqlite"))
	mkfile(t, filepath.Join(lib, "Application Support", "com.gone.App", "GPUCache", "index"))
	mkfile(t, filepath.Join(lib, "Application Support", "com.gone.App", "GPUCache", "data_0"))
	// 仍安装应用的数据、Apple 自带、命令行工具（无 GUI 痕迹）、非 Bundle ID 名称
	mkfile(t, filepath.Join(lib, "Preferences", "com.keep.App.plist"))
	mkfile(t, filepath.Join(lib, "Preferences", "com.apple.finder.plist"))
	mkfile(t, filepath.Join(lib, "Application Support", "com.vercel.cli", "auth.json"))
	mkfile(t, filepath.Join(lib, "Application Support", "Gone App", "data"))
	// 已卸载但近期仍有写入：可能装在非标准位置
	mkfile(t, filepath.Join(lib, "Preferences", "com.active.App.plist"))
	ageTree(t, home, 60*24*time.Hour)
	mkfile(t, filepath.Join(lib, "Application Support", "com.active.App", "fresh"))

	got := findLeftovers(home, buildAppIndex([]string{apps}), nil, nil)
	if len(got) != 1 {
		t.Fatalf("leftovers = %v", got)
	}
	paths := got["com.gone.app"]
	sort.Strings(paths)
	want := []string{
		filepath.Join(lib, "Application Support", "com.gone.App", "db.sqlite"),
		filepath.Join(lib, "Preferences", "com.gone.App.plist"),
	}
	if len(paths) != len(want) || paths[0] != want[0] || paths[1] != want[1] {
		t.Errorf("paths = %v, want %v", paths, want)
	}

	// /Applications 不可读时一律不判定
	if got := findLeftovers(home, buildAppIndex([]string{filepath.Join(home, "missing")}), nil, nil); len(got) != 0 {
		t.Errorf("unreadable app index must yield no leftovers, got %v", got)
	}
}

func TestBrokenLaunchAgents(t *testing.T) {
	dir := t.TempDir()
	exists := filepath.Join(dir, "bin")
	mkfile(t, exists)
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	agent := func(prog string) string {
		return `<?xml version="1.0"?><plist version="1.0"><dict><key>ProgramArguments</key><array><string>` + prog + `</string></array></dict></plist>`
	}
	write("broken.plist", agent("/Applications/Gone.app/Contents/MacOS/gone"))
	write("ok.plist", agent(exists))
	write("relative.plist", agent("gone"))
	write("volume.plist", agent("/Volumes/External/tool"))
	write("com.apple.x.plist", agent("/nope"))

	items := brokenLaunchAgents(dir)
	if len(items) != 1 || items[0].ID != "broken_agent_broken" {
		t.Fatalf("items = %+v", items)
	}
}

func TestInstallerAppInstalled(t *testing.T) {
	idx := &installedAppIndex{nameKeys: map[string]bool{"docker": true, "visualstudiocode": true}}
	for name, want := range map[string]bool{
		"Docker-4.30.0.dmg":         true,
		"Visual Studio Code.dmg":    true,
		"NTFS-Helper-Installer.dmg": false,
		"123.pkg":                   false,
	} {
		if got := installerAppInstalled(name, idx); got != want {
			t.Errorf("installerAppInstalled(%q) = %v", name, got)
		}
	}
}
