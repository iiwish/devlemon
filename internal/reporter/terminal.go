package reporter

import (
	"fmt"
	"io"
	"strings"

	"devlemon/internal/i18n"
	"devlemon/internal/model"
)

type TerminalReporter struct {
	w io.Writer
}

func NewTerminalReporter(w io.Writer) *TerminalReporter {
	return &TerminalReporter{w: w}
}

func (r *TerminalReporter) Render(report *model.ScanReport) {
	fmt.Fprintf(r.w, "\n🍋 \033[1;36mDevLemon\033[0m - %s\n", i18n.T("Developer Disk Scan Report", "开发者智能磁盘扫描报告"))
	fmt.Fprintf(r.w, "%s\n", strings.Repeat("─", 65))

	// 磁盘总体状态
	capStatus := fmt.Sprintf("%d%%", report.DiskCapacityPercent)
	if report.DiskCapacityPercent >= 90 {
		capStatus = fmt.Sprintf("\033[1;31m%d%% (%s)\033[0m", report.DiskCapacityPercent, i18n.T("Low Space", "空间紧张"))
	}

	fmt.Fprintf(r.w, "%s: \033[1m%s\033[0m  |  %s: %s  |  %s: \033[1;32m%s\033[0m [%s]\n",
		i18n.T("💾 Total Disk", "💾 磁盘总容量"),
		model.FormatBytes(report.DiskTotal),
		i18n.T("Used", "已用"),
		model.FormatBytes(report.DiskUsed),
		i18n.T("Free", "可用"),
		model.FormatBytes(report.DiskFree),
		capStatus,
	)

	fmt.Fprintf(r.w, "%s: \033[1;33m%s\033[0m\n", i18n.T("✨ Total Reclaimable", "✨ 累计可回收"), model.FormatBytes(report.TotalReclaimableBytes))
	fmt.Fprintf(r.w, "%s\n\n", strings.Repeat("─", 65))

	for _, group := range report.Groups {
		fmt.Fprintf(r.w, "\033[1m%s\033[0m (%s: \033[33m%s\033[0m)\n", group.Title, i18n.T("Reclaimable in group", "组内可回收"), model.FormatBytes(group.TotalReclaimableBytes))
		for _, item := range group.Items {
			if item.IsProtected {
				fmt.Fprintf(r.w, "  🔒 \033[2m[%s] %-30s %10s\033[0m\n", i18n.T("Protected", "受保护"), item.Title, item.SizeFormatted)
				fmt.Fprintf(r.w, "     └─ \033[32m%s\033[0m\n", item.ProtectReason)
				continue
			}

			riskTag := i18n.T("🟢 Safe", "🟢 绝对安全")
			switch item.Risk {
			case model.RiskRebuildable:
				riskTag = i18n.T("🟡 Rebuildable", "🟡 可重构建")
			case model.RiskCaution:
				riskTag = i18n.T("🟠 Caution", "🟠 谨慎操作")
			}

			fmt.Fprintf(r.w, "  • %-34s \033[1m%10s\033[0m  [%s]\n", item.Title, item.SizeFormatted, riskTag)
			fmt.Fprintf(r.w, "    └─ \033[2m%s\033[0m\n", item.Description)
		}
		fmt.Fprintln(r.w)
	}

	fmt.Fprintf(r.w, "%s\n", strings.Repeat("─", 65))
	fmt.Fprintf(r.w, "💡 提示: 运行 \033[1;32mdevlemon clean --safe\033[0m 可一键无损清理所有绿色安全缓存\n\n")
}
