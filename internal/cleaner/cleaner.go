package cleaner

import (
	"context"
	"fmt"
	"os"
	"os/exec"

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
