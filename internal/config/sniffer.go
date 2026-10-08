package config

import (
	"os"
	"path/filepath"
	"strings"
)

// 项目签名文件列表：只要子目录包含这些文件之一，即判定为一个代码工程
var projectMarkers = []string{
	".git",
	"package.json",
	"Cargo.toml",
	"go.mod",
	"Package.swift",
	"pom.xml",
	"build.gradle",
	"pyproject.toml",
	"CMakeLists.txt",
}

// macOS 默认的非开发用系统文件夹（跳过扫描以极大提升速度）
var ignoredHomeDirs = map[string]bool{
	"Library":      true,
	"Applications": true,
	"Desktop":      true,
	"Downloads":    true,
	"Movies":       true,
	"Music":        true,
	"Pictures":     true,
	"Public":       true,
}

// AutoDiscoverWorkspaces 智能自动嗅探用户主目录下的代码工作区
func AutoDiscoverWorkspaces() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	workspaces := make([]string, 0)
	foundSet := make(map[string]bool)

	addWorkspace := func(p string) {
		clean := filepath.Clean(p)
		if !foundSet[clean] {
			foundSet[clean] = true
			workspaces = append(workspaces, clean)
		}
	}

	// 1. 行业标准约定目录检测 (如 ~/Developer, ~/Projects 等)
	standardDirs := []string{
		"Developer",
		"Projects",
		"Workspace",
		"Code",
		"src",
		"development",
		"repos",
	}
	for _, dir := range standardDirs {
		candidate := filepath.Join(home, dir)
		if isDirectory(candidate) {
			addWorkspace(candidate)
		}
	}

	// 2. 启发式代码仓库聚类嗅探 (Repository Clustering Sniffer)
	// 遍历 ~ 下的一级非隐藏、非系统目录（如 self, daas, mycode, workspace_2026 等）
	entries, err := os.ReadDir(home)
	if err != nil {
		return workspaces
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") || ignoredHomeDirs[name] {
			continue
		}

		candidatePath := filepath.Join(home, name)
		if !entry.IsDir() {
			continue
		}

		// 对该目录做极轻量的浅层探测（深度 1）：检查是否有子目录包含项目签名
		if countProjectsInDir(candidatePath, 1) >= 1 {
			addWorkspace(candidatePath)
		}
	}

	// 3. 对 ~/Documents 做一层浅层探测（部分开发者喜欢放在 Documents/Projects 等目录）
	docsPath := filepath.Join(home, "Documents")
	if isDirectory(docsPath) {
		if docEntries, err := os.ReadDir(docsPath); err == nil {
			for _, de := range docEntries {
				if de.IsDir() && !strings.HasPrefix(de.Name(), ".") {
					subPath := filepath.Join(docsPath, de.Name())
					if countProjectsInDir(subPath, 1) >= 1 {
						addWorkspace(subPath)
					}
				}
			}
		}
	}

	return workspaces
}

// countProjectsInDir 浅层统计目录内包含的项目工程数量
func countProjectsInDir(dirPath string, maxDepth int) int {
	subEntries, err := os.ReadDir(dirPath)
	if err != nil {
		return 0
	}

	projectCount := 0
	for _, sub := range subEntries {
		if !sub.IsDir() || strings.HasPrefix(sub.Name(), ".") {
			continue
		}

		subDirPath := filepath.Join(dirPath, sub.Name())
		if isProjectDirectory(subDirPath) {
			projectCount++
		}
	}

	return projectCount
}

// isProjectDirectory 判断一个目录是否为代码工程根目录
func isProjectDirectory(dir string) bool {
	for _, marker := range projectMarkers {
		markerPath := filepath.Join(dir, marker)
		if _, err := os.Stat(markerPath); err == nil {
			return true
		}
	}
	return false
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
