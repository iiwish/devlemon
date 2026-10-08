package probe

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"devlemon/internal/config"
	"devlemon/internal/model"
)

type DockerProbe struct{}

func NewDockerProbe() *DockerProbe {
	return &DockerProbe{}
}

func (p *DockerProbe) Name() string {
	return "Docker & Container Environment Probe"
}

func (p *DockerProbe) Category() model.Category {
	return model.CategoryDocker
}

type dockerDFOutput struct {
	Type        string `json:"Type"`
	Total       string `json:"Total"`
	Active      string `json:"Active"`
	Size        string `json:"Size"`
	Reclaimable string `json:"Reclaimable"`
}

func (p *DockerProbe) Scan(ctx context.Context, cfg *config.Config) (*model.Group, error) {
	group := &model.Group{
		ID:       "docker_resources",
		Title:    "🐳 Docker / OrbStack 容器资源",
		Category: model.CategoryDocker,
		Items:    make([]*model.Item, 0),
	}

	// 1. 检查 Docker 命令与守护进程是否可用
	if _, err := exec.LookPath("docker"); err != nil {
		return group, nil
	}

	cmd := exec.CommandContext(ctx, "docker", "system", "df", "--format", "{{json .}}")
	out, err := cmd.Output()
	if err != nil {
		// Docker 未启动（例如 daemon 未运行）
		return group, nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var df dockerDFOutput
		if err := json.Unmarshal([]byte(line), &df); err != nil {
			continue
		}

		sizeBytes := parseDockerSize(df.Size)
		reclaimBytes := parseDockerReclaimable(df.Reclaimable)

		switch df.Type {
		case "Build Cache":
			if sizeBytes > 0 {
				group.Items = append(group.Items, &model.Item{
					ID:            "docker_build_cache",
					Title:         "Docker BuildKit 构建缓存",
					Description:   "中间编译层与 Dockerfile 构建缓存，清空后按需自动重新构建",
					SizeBytes:     sizeBytes,
					SizeFormatted: model.FormatBytes(sizeBytes),
					Risk:          model.RiskSafe,
					Category:      model.CategoryDocker,
					CleanType:     model.CleanTypeCommand,
					CleanCommand:  []string{"docker", "builder", "prune", "-a", "-f"},
				})
				group.TotalSizeBytes += sizeBytes
				group.TotalReclaimableBytes += sizeBytes
			}

		case "Images":
			if reclaimBytes > 0 {
				group.Items = append(group.Items, &model.Item{
					ID:            "docker_unused_images",
					Title:         "未使用的 Docker 镜像",
					Description:   "未被任何运行中容器引用的旧版本镜像（已自动保护活跃容器镜像）",
					SizeBytes:     reclaimBytes,
					SizeFormatted: model.FormatBytes(reclaimBytes),
					Risk:          model.RiskSafe,
					Category:      model.CategoryDocker,
					CleanType:     model.CleanTypeCommand,
					CleanCommand:  []string{"docker", "image", "prune", "-a", "-f"},
				})
				group.TotalSizeBytes += reclaimBytes
				group.TotalReclaimableBytes += reclaimBytes
			}

		case "Local Volumes":
			if reclaimBytes > 0 {
				group.Items = append(group.Items, &model.Item{
					ID:            "docker_dangling_volumes",
					Title:         "未挂载的 Docker 孤立数据卷",
					Description:   "容器已销毁但残留的历史匿名数据卷",
					SizeBytes:     reclaimBytes,
					SizeFormatted: model.FormatBytes(reclaimBytes),
					Risk:          model.RiskCaution,
					Category:      model.CategoryDocker,
					CleanType:     model.CleanTypeCommand,
					CleanCommand:  []string{"docker", "volume", "prune", "-f"},
				})
				group.TotalSizeBytes += reclaimBytes
				group.TotalReclaimableBytes += reclaimBytes
			}

		case "Containers":
			if reclaimBytes > 0 {
				group.Items = append(group.Items, &model.Item{
					ID:            "docker_stopped_containers",
					Title:         "已退出的历史测试容器",
					Description:   "处于 Exited 状态的旧容器",
					SizeBytes:     reclaimBytes,
					SizeFormatted: model.FormatBytes(reclaimBytes),
					Risk:          model.RiskSafe,
					Category:      model.CategoryDocker,
					CleanType:     model.CleanTypeCommand,
					CleanCommand:  []string{"docker", "container", "prune", "-f"},
				})
				group.TotalSizeBytes += reclaimBytes
				group.TotalReclaimableBytes += reclaimBytes
			}
		}
	}

	// 2. 检测 OrbStack 宿主机目录
	home, _ := os.UserHomeDir()
	if home != "" {
		orbstackPath := filepath.Join(home, "Library", "Group Containers", "HUAQ24HBR6.dev.orbstack")
		if info, err := os.Stat(orbstackPath); err == nil && info.IsDir() {
			// 仅做空间信息展示或检测
		}
	}

	return group, nil
}

// 辅助解析 Docker 输出的大小字符串，如 "15.34GB", "500MB", "12.5kB"
func parseDockerSize(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "0B" {
		return 0
	}

	multiplier := int64(1)
	cleanStr := s

	switch {
	case strings.HasSuffix(s, "GB") || strings.HasSuffix(s, "GiB"):
		multiplier = 1024 * 1024 * 1024
		cleanStr = strings.TrimSuffix(strings.TrimSuffix(s, "GiB"), "GB")
	case strings.HasSuffix(s, "MB") || strings.HasSuffix(s, "MiB"):
		multiplier = 1024 * 1024
		cleanStr = strings.TrimSuffix(strings.TrimSuffix(s, "MiB"), "MB")
	case strings.HasSuffix(s, "kB") || strings.HasSuffix(s, "KB") || strings.HasSuffix(s, "KiB"):
		multiplier = 1024
		cleanStr = strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(s, "KiB"), "kB"), "KB")
	case strings.HasSuffix(s, "B"):
		cleanStr = strings.TrimSuffix(s, "B")
	}

	val, err := strconv.ParseFloat(strings.TrimSpace(cleanStr), 64)
	if err != nil {
		return 0
	}
	return int64(val * float64(multiplier))
}

// 解析类似 "12.69GB (26%)" 中的可回收大小
func parseDockerReclaimable(s string) int64 {
	s = strings.TrimSpace(s)
	if idx := strings.Index(s, "("); idx != -1 {
		s = strings.TrimSpace(s[:idx])
	}
	return parseDockerSize(s)
}
