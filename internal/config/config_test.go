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
