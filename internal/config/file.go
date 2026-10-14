package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// fileConfig 用户配置文件中可设置的字段；未出现的字段保持默认值
type fileConfig struct {
	WorkspacePaths []string `yaml:"workspace_paths"`
	DormancyDays   *int     `yaml:"dormancy_days"`
	TargetDirNames []string `yaml:"target_dir_names"`
	MaxDepth       *int     `yaml:"max_depth"`
}

// UserConfigPath 配置文件路径：$DEVLEMON_CONFIG，否则 $XDG_CONFIG_HOME/devlemon/config.yaml，否则 ~/.config/devlemon/config.yaml
func UserConfigPath() string {
	if p := os.Getenv("DEVLEMON_CONFIG"); p != "" {
		return p
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "devlemon", "config.yaml")
}

// LoadUserFile 读取用户配置文件并合并到 c 中。文件不存在时什么也不做；
// 内容有误时返回错误且不修改 c（整个文件要么全部生效，要么全部不生效）
func (c *Config) LoadUserFile() error {
	path := UserConfigPath()
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return c.applyFile(data, path)
}

func (c *Config) applyFile(data []byte, path string) error {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var f fileConfig
	if err := dec.Decode(&f); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return fmt.Errorf("%s: %w", path, err)
	}
	if f.DormancyDays != nil && *f.DormancyDays < 1 {
		return fmt.Errorf("%s: dormancy_days must be at least 1", path)
	}
	if f.MaxDepth != nil && (*f.MaxDepth < 1 || *f.MaxDepth > 10) {
		return fmt.Errorf("%s: max_depth must be between 1 and 10", path)
	}
	for _, n := range f.TargetDirNames {
		if n == "" || n == "." || n == ".." || strings.ContainsAny(n, `/\`) {
			return fmt.Errorf("%s: target_dir_names must be plain directory names (got %q)", path, n)
		}
	}

	for _, w := range f.WorkspacePaths {
		c.AddWorkspacePath(w)
	}
	if f.DormancyDays != nil {
		c.DormancyDays = *f.DormancyDays
	}
	if len(f.TargetDirNames) > 0 {
		c.TargetDirNames = f.TargetDirNames
	}
	if f.MaxDepth != nil {
		c.MaxDepth = *f.MaxDepth
	}
	return nil
}
