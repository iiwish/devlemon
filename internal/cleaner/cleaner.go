package cleaner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"devlemon/internal/model"
)

type Cleaner struct {
	DryRun bool
}

func NewCleaner(dryRun bool) *Cleaner {
	return &Cleaner{DryRun: dryRun}
}

// CleanItem 执行单项清理
func (c *Cleaner) CleanItem(ctx context.Context, item *model.Item) error {
	// 运行态保护熔断机制
	if item.IsProtected {
		return fmt.Errorf("操作被拒绝: 该项受到安全保护 (%s)", item.ProtectReason)
	}

	if c.DryRun {
		return nil
	}

	switch item.CleanType {
	case model.CleanTypeRemovePath:
		if item.CleanPath == "" || item.CleanPath == "/" {
			return fmt.Errorf("非法删除路径: %s", item.CleanPath)
		}
		return os.RemoveAll(item.CleanPath)

	case model.CleanTypeRemovePaths:
		if len(item.CleanPaths) == 0 {
			return fmt.Errorf("空清理路径列表")
		}
		// 先整体校验再删除，任一路径非法则整项拒绝执行
		for _, p := range item.CleanPaths {
			if err := validateRemovePath(p); err != nil {
				return err
			}
		}
		var firstErr error
		for _, p := range item.CleanPaths {
			if err := os.RemoveAll(p); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return firstErr

	case model.CleanTypeCommand:
		if len(item.CleanCommand) == 0 {
			return fmt.Errorf("空清理指令")
		}
		cmd := exec.CommandContext(ctx, item.CleanCommand[0], item.CleanCommand[1:]...)
		return cmd.Run()

	default:
		return fmt.Errorf("未知的清理模式: %s", item.CleanType)
	}
}

// validateRemovePath 删除前的最后一道防线：必须是绝对路径，且不能是根目录、用户主目录或其直接上级
func validateRemovePath(p string) error {
	clean := filepath.Clean(p)
	if p == "" || !filepath.IsAbs(clean) || clean == "/" {
		return fmt.Errorf("非法删除路径: %q", p)
	}
	if home, err := os.UserHomeDir(); err == nil {
		home = filepath.Clean(home)
		if clean == home || strings.HasPrefix(home, clean+string(filepath.Separator)) {
			return fmt.Errorf("拒绝删除用户主目录或其上级: %s", clean)
		}
	}
	// 至少位于三级目录之下，杜绝 /Users、/var/folders 等顶层目录被误删
	if strings.Count(clean, string(filepath.Separator)) < 3 {
		return fmt.Errorf("删除路径层级过浅: %s", clean)
	}
	return nil
}

// CleanItems 批量清理
func (c *Cleaner) CleanItems(ctx context.Context, items []*model.Item) (int64, int, []error) {
	var totalFreed int64
	var successCount int
	var errs []error

	for _, item := range items {
		if item.IsProtected {
			continue
		}
		err := c.CleanItem(ctx, item)
		if err != nil {
			errs = append(errs, fmt.Errorf("[%s] %w", item.Title, err))
		} else {
			successCount++
			totalFreed += item.SizeBytes
		}
	}

	// 清理完毕后，如在 macOS 上，安全触发一次 APFS 本地快照释放通知（避免快照锁死空间）
	if !c.DryRun && successCount > 0 {
		_ = exec.CommandContext(ctx, "tmutil", "thinlocalsnapshots", "/", "999999999999", "4").Run()
	}

	return totalFreed, successCount, errs
}
