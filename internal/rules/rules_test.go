package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuiltinRulesAreValid(t *testing.T) {
	rs, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	if len(rs) < 30 {
		t.Fatalf("expected the built-in rule set, got %d rules", len(rs))
	}
	for _, r := range rs {
		if r.Title.Zh == "" || r.Description.En == "" || r.Description.Zh == "" {
			t.Errorf("%s: built-in rules must have en/zh title and description", r.ID)
		}
	}
}

func rule(paths ...string) Rule {
	return Rule{ID: "demo", Group: GroupPackageCaches, Title: Text{En: "Demo"}, Risk: "rebuildable", Paths: paths}
}

func TestValidatePaths(t *testing.T) {
	ok := []string{
		"~/.npm",
		"~/Library/Caches/go-build",
		"~/Library/Developer/Xcode/DerivedData",
		"~/.cache/huggingface/hub",
		"{darwin_user_cache}/clang/ModuleCache",
	}
	for _, p := range ok {
		if err := rule(p).Validate(); err != nil {
			t.Errorf("%q should be valid: %v", p, err)
		}
	}
	bad := []string{
		"/Users/me/.npm",           // 绝对路径
		"~",                        // 主目录本身
		"~/",                       // 主目录本身
		"~/Library",                // 受保护目录
		"~/Library/Caches",         // 受保护目录
		"~/Documents",              // 用户数据
		"~/.ssh",                   // 凭据
		"~/.config",                // 所有工具的配置
		"~/.cache/../Documents",    // 越界
		"~/Library/Caches/../../x", // 越界
		"~/Library/Caches/*",       // 通配符
		"~/Library//Caches/x",      // 非规范路径
		"~/Library/Caches/./x",     // 非规范路径
		"{darwin_user_cache}",      // 缓存根目录本身
		"{darwin_user_cache}/../T", // 越界
		"$HOME/.npm",               // 不支持环境变量
		"~\\Library\\Caches\\foo",  // 反斜杠
		"{other_var}/x",            // 未知变量
	}
	for _, p := range bad {
		if err := rule(p).Validate(); err == nil {
			t.Errorf("%q should be rejected", p)
		}
	}
}

func TestValidateFields(t *testing.T) {
	base := rule("~/.npm")
	cases := map[string]func(r *Rule){
		"bad id":        func(r *Rule) { r.ID = "Bad-ID" },
		"no title":      func(r *Rule) { r.Title.En = "" },
		"bad risk":      func(r *Rule) { r.Risk = "yolo" },
		"bad size":      func(r *Rule) { r.MinSize = "ten" },
		"bad os":        func(r *Rule) { r.OS = []string{"plan9"} },
		"no paths":      func(r *Rule) { r.Paths = nil },
		"no group":      func(r *Rule) { r.Group = "" },
		"empty command": func(r *Rule) { r.CleanCommand = []string{""} },
	}
	for name, mutate := range cases {
		r := base
		mutate(&r)
		if err := r.Validate(); err == nil {
			t.Errorf("%s: expected validation error", name)
		}
	}
}

func TestParseSize(t *testing.T) {
	cases := map[string]int64{
		"": 10 << 20, "0": 0, "512KB": 512 << 10, "10MB": 10 << 20, "1.5GB": 3 << 29, "100 mb": 100 << 20, "2048": 2048,
	}
	for in, want := range cases {
		got, err := ParseSize(in)
		if err != nil || got != want {
			t.Errorf("ParseSize(%q) = %d, %v; want %d", in, got, err, want)
		}
	}
	for _, in := range []string{"ten", "10TB", "-1MB"} {
		if _, err := ParseSize(in); err == nil {
			t.Errorf("ParseSize(%q) should fail", in)
		}
	}
}

func TestParseFileRejectsUnknownFields(t *testing.T) {
	_, err := ParseFile([]byte("group: package_caches\nrules:\n  - id: x\n    pathz: [\"~/.x\"]\n"), "typo.yaml")
	if err == nil || !strings.Contains(err.Error(), "pathz") {
		t.Fatalf("typo must be reported, got %v", err)
	}
	if _, err := ParseFile([]byte("group: nope\nrules: []\n"), "g.yaml"); err == nil {
		t.Error("unknown group must be rejected")
	}
	dup := "group: app_caches\nrules:\n  - id: a\n  - id: a\n"
	if _, err := ParseFile([]byte(dup), "dup.yaml"); err == nil {
		t.Error("duplicate ids must be rejected")
	}
}

func TestLoadUserRules(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DEVLEMON_RULES_DIR", dir)
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("10-add.yaml", `group: package_caches
rules:
  - id: my_tool_cache
    title: {en: "My Tool Cache"}
    paths: ["~/.my-tool/cache"]
    risk: rebuildable
  - id: go_build_cache          # 覆盖内置规则
    title: {en: "Go Build Cache (custom)"}
    paths: ["~/Library/Caches/go-build"]
    risk: caution
  - id: dangerous
    title: {en: "Bad"}
    paths: ["~/Documents"]
    risk: rebuildable
`)
	write("20-disable.yaml", `rules:
  - id: npm_cache
    enabled: false
  - id: no_such_rule
    enabled: false
`)
	write("30-broken.yml", "group: package_caches\nrules: [\n")

	set, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]Rule{}
	for _, r := range set.Rules {
		byID[r.ID] = r
	}
	if _, ok := byID["my_tool_cache"]; !ok {
		t.Error("user rule not added")
	}
	if r := byID["go_build_cache"]; r.Risk != "caution" || len(r.CleanCommand) != 0 {
		t.Errorf("user rule should replace the built-in one entirely: %+v", r)
	}
	if _, ok := byID["npm_cache"]; ok {
		t.Error("disabled built-in rule still loaded")
	}
	if _, ok := byID["dangerous"]; ok {
		t.Error("invalid user rule must be skipped")
	}
	warnings := strings.Join(set.Warnings, "\n")
	for _, want := range []string{"dangerous", "no_such_rule", "30-broken.yml"} {
		if !strings.Contains(warnings, want) {
			t.Errorf("missing warning about %s in:\n%s", want, warnings)
		}
	}
	// 覆盖后的规则仍只出现一次
	n := 0
	for _, r := range set.Rules {
		if r.ID == "go_build_cache" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("go_build_cache appears %d times", n)
	}
}

func TestExpandPaths(t *testing.T) {
	r := rule("~/Library/Caches/go-build", "{darwin_user_cache}/clang/ModuleCache")
	got := r.ExpandPaths("/Users/me")
	if got[0] != "/Users/me/Library/Caches/go-build" {
		t.Errorf("ExpandPaths = %v", got)
	}
	if dir := DarwinUserCacheDir(); dir != "" {
		if len(got) != 2 || got[1] != filepath.Join(dir, "clang", "ModuleCache") {
			t.Errorf("ExpandPaths = %v", got)
		}
	} else if len(got) != 1 {
		t.Errorf("unavailable variable must skip the path: %v", got)
	}
}

func TestTextFallback(t *testing.T) {
	if (Text{En: "Only English"}).String() != "Only English" {
		t.Error("missing zh must fall back to en")
	}
}
