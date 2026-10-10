package model

import (
	"fmt"
	"time"
)

// RiskLevel 风险等级定义
type RiskLevel string

const (
	// RiskSafe 100% 安全：纯临时下载或编译缓存，删除后无任何副作用，按需自动重新下载生成
	RiskSafe RiskLevel = "safe"
	// RiskRebuildable 可重构产物：如本地代码的 target/、.build/、node_modules，需重新编译但绝不丢失代码
	RiskRebuildable RiskLevel = "rebuildable"
	// RiskCaution 需谨慎确认：如重置已关机的旧模拟器沙盒（应用数据将被抹掉）
	RiskCaution RiskLevel = "caution"
)

// Category 类别定义
type Category string

const (
	CategoryDocker         Category = "docker"
	CategorySimulator      Category = "simulator"
	CategoryPackageCache   Category = "package_cache"
	CategoryWorkspaceBuild Category = "workspace_build"
	CategorySystemCache    Category = "system_cache"
	CategoryAppCache       Category = "app_cache"
	CategoryAIEraCache     Category = "ai_cache"
)

// CleanType 清理方式
type CleanType string

const (
	CleanTypeRemovePath CleanType = "remove_path"
	CleanTypeCommand    CleanType = "command"
	// CleanTypeRemovePaths 逐个删除 CleanPaths 中的路径（不经过 shell，规避参数长度上限）
	CleanTypeRemovePaths CleanType = "remove_paths"
)

// Item 可清理项或受保护项的最小单元
type Item struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Path          string    `json:"path,omitempty"`
	SizeBytes     int64     `json:"size_bytes"`
	SizeFormatted string    `json:"size_formatted"`
	Risk          RiskLevel `json:"risk"`
	Category      Category  `json:"category"`

	// 运行态保护机制
	IsProtected   bool   `json:"is_protected"`
	ProtectReason string `json:"protect_reason,omitempty"`

	// 清理动作定义
	CleanType    CleanType `json:"clean_type"`
	CleanPath    string    `json:"clean_path,omitempty"`
	CleanCommand []string  `json:"clean_command,omitempty"`
	CleanPaths   []string  `json:"clean_paths,omitempty"`
}

// Group 归类组
type Group struct {
	ID                    string   `json:"id"`
	Title                 string   `json:"title"`
	Category              Category `json:"category"`
	TotalSizeBytes        int64    `json:"total_size_bytes"`
	TotalReclaimableBytes int64    `json:"total_reclaimable_bytes"`
	Items                 []*Item  `json:"items"`
}

// ScanReport 总体扫描报告（也是后续向 SwiftUI 暴露的标准数据模型）
type ScanReport struct {
	DiskTotal             int64     `json:"disk_total_bytes"`
	DiskFree              int64     `json:"disk_free_bytes"`
	DiskUsed              int64     `json:"disk_used_bytes"`
	DiskCapacityPercent   int       `json:"disk_capacity_percent"`
	TotalReclaimableBytes int64     `json:"total_reclaimable_bytes"`
	Groups                []*Group  `json:"groups"`
	ScannedAt             time.Time `json:"scanned_at"`
}

// FormatBytes 字节格式化工具函数
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
