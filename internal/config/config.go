package config

import (
	"os"
	"path/filepath"
)

// Config 全局通用配置
type Config struct {
	// WorkspacePaths 用户自定义的工作区根目录列表（例如：~/Projects, ~/Developer 等）
	// 注意：绝不硬编码任何个人私有路径，纯动态探测或由用户配置
	WorkspacePaths []string `json:"workspace_paths" yaml:"workspace_paths"`

	// DormancyDays 休眠天数阈值：多少天未修改的项目构建目录可被列为休眠可清理项（默认 7 天）
	DormancyDays int `json:"dormancy_days" yaml:"dormancy_days"`

	// TargetDirNames 识别为可清理的构建目录名称
	TargetDirNames []string `json:"target_dir_names" yaml:"target_dir_names"`

	// MaxDepth 递归查找项目的最大深度，避免无休止扫描深层目录
	MaxDepth int `json:"max_depth" yaml:"max_depth"`

	// SafeOnly 是否仅扫描/执行 100% 绝对安全的项目
	SafeOnly bool `json:"safe_only" yaml:"safe_only"`

	// SandboxMode 是否处于 macOS App Sandbox 沙盒环境 (符合 Mac App Store 规范)
	SandboxMode bool `json:"sandbox_mode" yaml:"sandbox_mode"`
}

// DefaultConfig 获取默认配置
func DefaultConfig() *Config {
	isSandboxed := os.Getenv("APP_SANDBOX_CONTAINER_ID") != ""
	cfg := &Config{
		WorkspacePaths: []string{},
		DormancyDays:   7,
		TargetDirNames: []string{
			"target",       // Rust / Cargo, Maven, Scala
			".build",       // Swift Package Manager
			"node_modules", // Node.js / 前端项目
			"build",        // C++/CMake/Java/Gradle
			".next",        // Next.js 编译产物
			"DerivedData",  // Xcode 编译产物
		},
		MaxDepth:    3,
		SafeOnly:    false,
		SandboxMode: isSandboxed,
	}

	// 非沙盒环境才自动嗅探整个个人主目录下的工作区，沙盒环境由用户显式授权指定
	if !isSandboxed {
		discovered := AutoDiscoverWorkspaces()
		cfg.WorkspacePaths = append(cfg.WorkspacePaths, discovered...)
	}

	return cfg
}

// AddWorkspacePath 动态添加工作区路径（支持展开 ~）
func (c *Config) AddWorkspacePath(p string) {
	if p == "" {
		return
	}
	expanded := ExpandHome(p)
	// 避免重复添加
	for _, existing := range c.WorkspacePaths {
		if existing == expanded {
			return
		}
	}
	c.WorkspacePaths = append(c.WorkspacePaths, expanded)
}

// ExpandHome 展开波浪号 ~ 路径
func ExpandHome(path string) string {
	if path == "" {
		return ""
	}
	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			if len(path) == 1 {
				return home
			}
			if path[1] == '/' || path[1] == filepath.Separator {
				return filepath.Join(home, path[2:])
			}
		}
	}
	return path
}
