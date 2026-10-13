package probe

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"devlemon/internal/model"
)

// sparse 创建指定逻辑大小的稀疏文件（目录体积按逻辑大小统计，不实际占用磁盘）
func sparse(t *testing.T, p string, size int64) {
	t.Helper()
	mkdir(t, filepath.Dir(p))
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(size); err != nil {
		t.Fatal(err)
	}
	f.Close()
}

func itemIDs(items []*model.Item) []string {
	var ids []string
	for _, it := range items {
		ids = append(ids, it.ID)
	}
	sort.Strings(ids)
	return ids
}

func assertIDs(t *testing.T, got []*model.Item, want ...string) {
	t.Helper()
	ids := itemIDs(got)
	sort.Strings(want)
	if len(ids) != len(want) {
		t.Fatalf("ids = %v, want %v", ids, want)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("ids = %v, want %v", ids, want)
		}
	}
}

const mb = 1024 * 1024

func TestPnpmOldStores(t *testing.T) {
	home := t.TempDir()
	root := filepath.Join(home, "Library", "pnpm", "store")
	sparse(t, filepath.Join(root, "v3", "files", "00", "a"), 80*mb)
	sparse(t, filepath.Join(root, "v10", "files", "00", "a"), 80*mb)
	sparse(t, filepath.Join(root, "v11", "files", "00", "a"), 80*mb)
	ageTree(t, home, 60*24*time.Hour)
	// v10 最近仍有写入：仍在被旧版 pnpm 使用，必须保留
	sparse(t, filepath.Join(root, "v10", "files", "01", "b"), 1)

	assertIDs(t, pnpmOldStores(home), "pnpm_store_v3")
}

func TestRustupToolchains(t *testing.T) {
	home := t.TempDir()
	t.Setenv("RUSTUP_HOME", "")
	root := filepath.Join(home, ".rustup")
	settings := `version = "12"
default_toolchain = "stable-aarch64-apple-darwin"

[overrides]
"/Users/me/legacy" = "1.88.0-aarch64-apple-darwin"
`
	mkdir(t, root)
	if err := os.WriteFile(filepath.Join(root, "settings.toml"), []byte(settings), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []string{"stable-aarch64-apple-darwin", "nightly-aarch64-apple-darwin",
		"1.88.0-aarch64-apple-darwin", "1.95.0-aarch64-apple-darwin", "1.96.1-aarch64-apple-darwin"} {
		sparse(t, filepath.Join(root, "toolchains", tc, "lib", "librustc.dylib"), 60*mb)
	}
	sparse(t, filepath.Join(root, "update-hashes", "1.95.0-aarch64-apple-darwin"), 40)
	ageTree(t, home, 60*24*time.Hour)
	// 1.96.1 刚安装：30 天内的不列出
	sparse(t, filepath.Join(root, "toolchains", "1.96.1-aarch64-apple-darwin", "fresh"), 1)

	items := rustupToolchains(home)
	assertIDs(t, items, "rustup_1.95.0-aarch64-apple-darwin")
	if len(items[0].CleanPaths) != 2 {
		t.Errorf("update-hashes entry should be removed together: %v", items[0].CleanPaths)
	}
}

func TestAndroidSDKItems(t *testing.T) {
	home := t.TempDir()
	t.Setenv("ANDROID_HOME", "")
	t.Setenv("ANDROID_SDK_ROOT", "")
	t.Setenv("ANDROID_AVD_HOME", "")
	sdk := filepath.Join(home, "Library", "Android", "sdk")
	sparse(t, filepath.Join(sdk, "system-images", "android-34", "google_apis", "arm64-v8a", "system.img"), 200*mb)
	sparse(t, filepath.Join(sdk, "system-images", "android-36", "google_apis", "arm64-v8a", "system.img"), 200*mb)
	for _, v := range []string{"25.2.9519653", "26.1.10909125", "27.0.12077973"} {
		sparse(t, filepath.Join(sdk, "ndk", v, "toolchains.bin"), 50*mb)
	}
	for _, v := range []string{"34.0.0", "35.0.1"} {
		sparse(t, filepath.Join(sdk, "build-tools", v, "aapt2"), 30*mb)
	}
	// 一个 AVD 使用 android-34 镜像
	avd := filepath.Join(home, ".android", "avd", "Pixel.avd")
	mkdir(t, avd)
	if err := os.WriteFile(filepath.Join(avd, "config.ini"),
		[]byte("hw.lcd.density=420\nimage.sysdir.1=system-images/android-34/google_apis/arm64-v8a/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ageTree(t, home, 60*24*time.Hour)

	assertIDs(t, androidSDKItems(home),
		"android_sysimg_system-images_android-36_google_apis_arm64-v8a",
		"android_ndk_25.2.9519653", "android_ndk_26.1.10909125",
		"android_build_tools_34.0.0")
}

func TestGradleWrapperDists(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GRADLE_USER_HOME", "")
	dists := filepath.Join(home, ".gradle", "wrapper", "dists")
	for _, d := range []string{"gradle-8.10-bin", "gradle-9.4.1-bin", "gradle-9.5.0-bin", "gradle-9.5.0-all"} {
		sparse(t, filepath.Join(dists, d, "hash", "gradle.zip"), 30*mb)
	}
	ageTree(t, home, 60*24*time.Hour)

	assertIDs(t, gradleWrapperDists(home), "gradle_dist_gradle-8.10-bin", "gradle_dist_gradle-9.4.1-bin")
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"9.10", "9.9", 1},
		{"27.0.12077973", "26.1.10909125", 1},
		{"1.2", "1.2.0", 0},
		{"34.0.0", "35.0.1", -1},
	}
	for _, c := range cases {
		if got := compareVersions(c.a, c.b); got != c.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
	if parseVersion("android-36") != nil || parseVersion("latest") != nil {
		t.Error("non-version names must not parse")
	}
}
