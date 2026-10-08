package reporter

import (
	"fmt"
	"io"
	"strings"

	"devlemon/internal/model"
)

type TerminalReporter struct {
	w io.Writer
}

func NewTerminalReporter(w io.Writer) *TerminalReporter {
	return &TerminalReporter{w: w}
}

func (r *TerminalReporter) Render(report *model.ScanReport) {
	fmt.Fprintf(r.w, "\n🍋 \033[1;36mDevLemon\033[0m - 开发者智能磁盘扫描报告\n")
	fmt.Fprintf(r.w, "%s\n", strings.Repeat("─", 65))

	// 磁盘总体状态
	capStatus := fmt.Sprintf("%d%%", report.DiskCapacityPercent)
	if report.DiskCapacityPercent >= 90 {
		capStatus = fmt.Sprintf("\033[1;31m%d%% (空间紧张)\033[0m", report.DiskCapacityPercent)
	}

	fmt.Fprintf(r.w, "💾 磁盘总容量: \033[1m%s\033[0m  |  已用: %s  |  可用: \033[1;32m%s\033[0m [%s]\n",
		model.FormatBytes(report.DiskTotal),
		model.FormatBytes(report.DiskUsed),
		model.FormatBytes(report.DiskFree),
		capStatus,
	)

	fmt.Fprintf(r.w, "✨ 累计可回收: \033[1;33m%s\033[0m\n", model.FormatBytes(report.TotalReclaimableBytes))
	fmt.Fprintf(r.w, "%s\n\n", strings.Repeat("─", 65))

	for _, group := range report.Groups {
		fmt.Fprintf(r.w, "\033[1m%s\033[0m (组内可回收: \033[33m%s\033[0m)\n", group.Title, model.FormatBytes(group.TotalReclaimableBytes))
		for _, item := range group.Items {
			if item.IsProtected {
				fmt.Fprintf(r.w, "  🔒 \033[2m[受保护] %-30s %10s\033[0m\n", item.Title, item.SizeFormatted)
				fmt.Fprintf(r.w, "     └─ \033[32m%s\033[0m\n", item.ProtectReason)
				continue
			}

			riskTag := "🟢 绝对安全"
			switch item.Risk {
			case model.RiskRebuildable:
				riskTag = "🟡 可重构建"
			case model.RiskCaution:
				riskTag = "🟠 谨慎操作"
			}

			fmt.Fprintf(r.w, "  • %-34s \033[1m%10s\033[0m  [%s]\n", item.Title, item.SizeFormatted, riskTag)
			fmt.Fprintf(r.w, "    └─ \033[2m%s\033[0m\n", item.Description)
		}
		fmt.Fprintln(r.w)
	}

	fmt.Fprintf(r.w, "%s\n", strings.Repeat("─", 65))
	fmt.Fprintf(r.w, "💡 提示: 运行 \033[1;32mdevlemon clean --safe\033[0m 可一键无损清理所有绿色安全缓存\n\n")
}
