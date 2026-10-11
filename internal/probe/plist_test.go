package probe

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

const xmlPlistFixture = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>CFBundleIdentifier</key><string>com.example.Demo</string>
<key>Name</key><string>柠檬 Demo</string>
<key>Count</key><integer>300</integer>
<key>ProgramArguments</key><array><string>/Applications/Demo.app/Contents/MacOS/demo</string><string>--flag</string></array>
<key>RunAtLoad</key><true/>
<key>Ratio</key><real>1.5</real>
</dict></plist>`

// 同一内容经 `plutil -convert binary1` 生成的二进制 plist（含 UTF-16 字符串与 1 字节 / 2 字节整数）
const binaryPlistFixture = "62706c6973743030d6010203040506070a0b0c0d0e5f101050726f6772616d417267756d656e7473544e616d655952756e41744c6f61645f1012434642756e646c654964656e74696669657255436f756e7455526174696fa208095f102a2f4170706c69636174696f6e732f44656d6f2e6170702f436f6e74656e74732f4d61634f532f64656d6f562d2d666c61676767e06aac002000440065006d006f095f1010636f6d2e6578616d706c652e44656d6f11012c233ff80000000000000815282d374c52585b888f9e9fb2b50000000000000101000000000000000f000000000000000000000000000000be"

func TestDecodePlist(t *testing.T) {
	want := map[string]any{
		"CFBundleIdentifier": "com.example.Demo",
		"Name":               "柠檬 Demo",
		"Count":              int64(300),
		"ProgramArguments":   []any{"/Applications/Demo.app/Contents/MacOS/demo", "--flag"},
		"RunAtLoad":          true,
		"Ratio":              1.5,
	}
	bin, err := hex.DecodeString(binaryPlistFixture)
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{"xml": []byte(xmlPlistFixture), "binary": bin} {
		got, err := decodePlist(data)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %#v", name, got)
		}
	}
}

func TestDecodePlistRejectsGarbage(t *testing.T) {
	bin, _ := hex.DecodeString(binaryPlistFixture)
	// 截断与篡改偏移表都必须返回错误而不是 panic
	cases := [][]byte{
		[]byte("bplist00"),
		bin[:len(bin)-10],
		append(append([]byte{}, bin[:len(bin)-8]...), 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff),
		[]byte("<plist><dict><key>a</key>"),
	}
	for i, c := range cases {
		if _, err := decodePlist(c); err == nil {
			t.Errorf("case %d: expected error", i)
		}
	}
}

func TestDecodePlistSelfReference(t *testing.T) {
	// 对象 0 为含两个元素的数组，两个元素都引用自身：无预算限制时会展开 2^64 次
	data := []byte("bplist00")
	data = append(data, 0xA2, 0x00, 0x00)
	table := len(data)
	data = append(data, 0x08)
	trailer := make([]byte, 32)
	trailer[6], trailer[7] = 1, 1
	trailer[15] = 1 // numObjects
	trailer[31] = byte(table)
	data = append(data, trailer...)

	done := make(chan error, 1)
	go func() { _, err := decodeBinaryPlist(data); done <- err }()
	select {
	case err := <-done:
		if err == nil {
			t.Error("self-referencing plist must be rejected")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("decoder did not terminate on self-referencing plist")
	}
}

func FuzzDecodeBinaryPlist(f *testing.F) {
	bin, _ := hex.DecodeString(binaryPlistFixture)
	f.Add(bin)
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = decodeBinaryPlist(data)
	})
}

func TestInstalledAppIndex(t *testing.T) {
	root := t.TempDir()
	writeInfoPlist(t, filepath.Join(root, "Demo.app"), "com.acme.Demo")
	writeInfoPlist(t, filepath.Join(root, "Demo.app", "Contents", "Library", "LoginItems", "Launcher.app"), "com.acmehelper.launcher")
	writeInfoPlist(t, filepath.Join(root, "Tools", "Nested.app"), "org.nested.App")
	writeInfoPlist(t, filepath.Join(root, "Gone.app"), "com.github.someone.gone")
	if err := os.RemoveAll(filepath.Join(root, "Gone.app")); err != nil {
		t.Fatal(err)
	}

	// Mac 上运行的 iPad 应用：Info.plist 位于 Wrapper/<名称>.app/ 根目录
	ipad := filepath.Join(root, "Pad.app", "Wrapper", "Pad.app", "Info.plist")
	mkdir(t, filepath.Dir(ipad))
	if err := os.WriteFile(ipad, []byte(`<?xml version="1.0"?><plist version="1.0"><dict><key>CFBundleIdentifier</key><string>com.ipadmaker.Pad</string></dict></plist>`), 0o644); err != nil {
		t.Fatal(err)
	}

	idx := buildAppIndex([]string{root})
	cases := map[string]bool{
		"com.acme.Demo":                true,  // 精确匹配（忽略大小写）
		"com.acme.demo.ShareExtension": true,  // 扩展以主应用 ID 为前缀
		"com.acme.OtherProduct":        true,  // 同厂商仍有应用安装
		"com.acmehelper.launcher":      true,  // 内嵌登录项
		"org.nested.App":               true,  // 子文件夹中的应用
		"com.ipadmaker.Pad":            true,  // Mac 上的 iPad 应用
		"com.github.someone.gone":      false, // 泛用前缀不按厂商匹配
		"com.removed.Thing":            false,
	}
	for id, want := range cases {
		if got := idx.hasApp(id); got != want {
			t.Errorf("hasApp(%q) = %v, want %v", id, got, want)
		}
	}
}

func writeInfoPlist(t *testing.T, app, id string) {
	t.Helper()
	p := filepath.Join(app, "Contents", "Info.plist")
	mkdir(t, filepath.Dir(p))
	body := `<?xml version="1.0"?><plist version="1.0"><dict><key>CFBundleIdentifier</key><string>` + id + `</string></dict></plist>`
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLooksLikeBundleID(t *testing.T) {
	for name, want := range map[string]bool{
		"com.example.App": true, "app.kirje.desktop": true, "com.baidu.BaiduNetdisk-mac": true,
		"Google": false, "com.example": false, "1.2.3": false, "my app.v2.x": false, "a..b": false,
	} {
		if got := looksLikeBundleID(name); got != want {
			t.Errorf("looksLikeBundleID(%q) = %v", name, got)
		}
	}
}
