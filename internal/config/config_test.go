package config

import (
	"os"
	"path/filepath"
	"testing"
)

// 默认配置不得遍历主目录：快速清理依赖这一点避免触发「文稿」等受保护目录的系统授权弹窗
func TestDefaultConfigDoesNotDiscover(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, "Projects"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := DefaultConfig()
	if len(cfg.WorkspacePaths) != 0 {
		t.Fatalf("DefaultConfig discovered workspaces: %v", cfg.WorkspacePaths)
	}

	cfg.DiscoverWorkspaces()
	cfg.DiscoverWorkspaces() // 重复调用不应产生重复路径
	if len(cfg.WorkspacePaths) != 1 || cfg.WorkspacePaths[0] != filepath.Join(home, "Projects") {
		t.Fatalf("DiscoverWorkspaces = %v", cfg.WorkspacePaths)
	}
}

func TestApplyFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := DefaultConfig()
	err := cfg.applyFile([]byte(`
workspace_paths:
  - ~/Projects
  - /opt/work
dormancy_days: 30
target_dir_names: [target, node_modules]
max_depth: 4
`), "config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{filepath.Join(home, "Projects"), "/opt/work"}
	if len(cfg.WorkspacePaths) != 2 || cfg.WorkspacePaths[0] != want[0] || cfg.WorkspacePaths[1] != want[1] {
		t.Errorf("workspace paths = %v, want %v", cfg.WorkspacePaths, want)
	}
	if cfg.DormancyDays != 30 || cfg.MaxDepth != 4 || len(cfg.TargetDirNames) != 2 {
		t.Errorf("config not applied: %+v", cfg)
	}

	// 有误的文件整体不生效
	for _, bad := range []string{
		"dormancy_days: 0",
		"max_depth: 99",
		"target_dir_names: [\"../x\"]",
		"workspace_path: [~/typo]",
		"dormancy_days: [",
	} {
		c := DefaultConfig()
		if err := c.applyFile([]byte(bad), "bad.yaml"); err == nil {
			t.Errorf("%q should be rejected", bad)
		}
		if c.DormancyDays != 7 || c.MaxDepth != 3 || len(c.WorkspacePaths) != 0 {
			t.Errorf("%q partially applied: %+v", bad, c)
		}
	}
}

func TestLoadUserFileMissing(t *testing.T) {
	t.Setenv("DEVLEMON_CONFIG", filepath.Join(t.TempDir(), "none.yaml"))
	if err := DefaultConfig().LoadUserFile(); err != nil {
		t.Errorf("missing config file must not be an error: %v", err)
	}
}
