package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"devlemon/internal/cleaner"
	"devlemon/internal/config"
	"devlemon/internal/engine"
	"devlemon/internal/model"
)

type state int

const (
	stateScanning state = iota
	stateDashboard
	stateCleaning
	stateDone
)

type scanDoneMsg struct {
	report *model.ScanReport
	err    error
}

type cleanDoneMsg struct {
	freedBytes   int64
	successCount int
	errs         []error
	isDryRun     bool
}

// FlatItem 用于列表扁平化展示的条目
type FlatItem struct {
	IsHeader  bool
	Group     *model.Group
	Item      *model.Item
	IsChecked bool
}

type Model struct {
	cfg       *config.Config
	state     state
	report    *model.ScanReport
	flatItems []FlatItem
	cursor    int
	width     int
	height    int
	statusMsg string

	// 动态旋转动画 Spinner
	spinner spinner.Model

	// 清理结果统计
	cleanedBytes int64
	cleanedCount int
	isDryRun     bool
}

func NewModel(cfg *config.Config) Model {
	s := spinner.NewModel()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FACC15")) // 经典柠檬黄

	return Model{
		cfg:       cfg,
		state:     stateScanning,
		cursor:    0,
		flatItems: make([]FlatItem, 0),
		spinner:   s,
		height:    24, // 默认初始高度
		width:     80,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		spinner.Tick,
		m.startScanCmd(),
	)
}

func (m Model) startScanCmd() tea.Cmd {
	return func() tea.Msg {
		eng := engine.NewEngine(m.cfg)
		report, err := eng.Scan(context.Background())
		return scanDoneMsg{report: report, err: err}
	}
}

func (m Model) startCleanCmd(isDryRun bool) tea.Cmd {
	return func() tea.Msg {
		var targets []*model.Item
		for _, fi := range m.flatItems {
			if !fi.IsHeader && fi.IsChecked && fi.Item != nil && !fi.Item.IsProtected {
				targets = append(targets, fi.Item)
			}
		}

		cln := cleaner.NewCleaner(isDryRun)
		freed, count, errs := cln.CleanItems(context.Background(), targets)
		return cleanDoneMsg{
			freedBytes:   freed,
			successCount: count,
			errs:         errs,
			isDryRun:     isDryRun,
		}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		if m.state == stateScanning || m.state == stateCleaning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case scanDoneMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("扫描失败: %v", msg.err)
			m.state = stateDashboard
			return m, nil
		}
		m.report = msg.report
		m.flatItems = buildFlatItems(m.report)
		m.state = stateDashboard
		m.cursor = 0
		m.skipHeadersDown()
		return m, nil

	case cleanDoneMsg:
		m.cleanedBytes = msg.freedBytes
		m.cleanedCount = msg.successCount
		m.isDryRun = msg.isDryRun
		m.state = stateDone
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.state == stateDashboard {
				m.moveCursorUp()
			}

		case "down", "j":
			if m.state == stateDashboard {
				m.moveCursorDown()
			}

		// Tab 键：支持在各大分类组之间秒级快速跳转！
		case "tab":
			if m.state == stateDashboard {
				m.jumpToNextGroup()
			}

		case "shift+tab":
			if m.state == stateDashboard {
				m.jumpToPrevGroup()
			}

		case " ":
			if m.state == stateDashboard && len(m.flatItems) > 0 {
				fi := &m.flatItems[m.cursor]
				if !fi.IsHeader && fi.Item != nil && !fi.Item.IsProtected {
					fi.IsChecked = !fi.IsChecked
				}
			}

		case "a", "A":
			if m.state == stateDashboard {
				m.toggleSelectAllSafe()
			}

		case "enter":
			if m.state == stateDashboard {
				if m.selectedCount() > 0 {
					m.state = stateCleaning
					cmds = append(cmds, spinner.Tick, m.startCleanCmd(false))
				}
			} else if m.state == stateDone {
				m.state = stateScanning
				cmds = append(cmds, spinner.Tick, m.startScanCmd())
			}

		case "d", "D":
			if m.state == stateDashboard && m.selectedCount() > 0 {
				m.state = stateCleaning
				cmds = append(cmds, spinner.Tick, m.startCleanCmd(true))
			}

		case "r", "R":
			if m.state == stateDashboard || m.state == stateDone {
				m.state = stateScanning
				cmds = append(cmds, spinner.Tick, m.startScanCmd())
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) moveCursorUp() {
	if m.cursor > 0 {
		m.cursor--
		for m.cursor > 0 && m.flatItems[m.cursor].IsHeader {
			m.cursor--
		}
	}
}

func (m *Model) moveCursorDown() {
	if m.cursor < len(m.flatItems)-1 {
		m.cursor++
		for m.cursor < len(m.flatItems)-1 && m.flatItems[m.cursor].IsHeader {
			m.cursor++
		}
	}
}

func (m *Model) skipHeadersDown() {
	for m.cursor < len(m.flatItems) && m.flatItems[m.cursor].IsHeader {
		m.cursor++
	}
}

func (m *Model) jumpToNextGroup() {
	foundNext := false
	for i := m.cursor + 1; i < len(m.flatItems); i++ {
		if m.flatItems[i].IsHeader {
			// 跳到该 header 后的第一个有效条目
			target := i + 1
			for target < len(m.flatItems) && m.flatItems[target].IsHeader {
				target++
			}
			if target < len(m.flatItems) {
				m.cursor = target
				foundNext = true
				break
			}
		}
	}
	if !foundNext {
		// 循环回顶部
		m.cursor = 0
		m.skipHeadersDown()
	}
}

func (m *Model) jumpToPrevGroup() {
	for i := m.cursor - 2; i >= 0; i-- {
		if m.flatItems[i].IsHeader {
			target := i + 1
			for target < len(m.flatItems) && m.flatItems[target].IsHeader {
				target++
			}
			if target < len(m.flatItems) {
				m.cursor = target
				return
			}
		}
	}
}

func (m *Model) selectedCount() int {
	count := 0
	for _, fi := range m.flatItems {
		if !fi.IsHeader && fi.IsChecked {
			count++
		}
	}
	return count
}

func (m *Model) selectedBytes() int64 {
	var total int64
	for _, fi := range m.flatItems {
		if !fi.IsHeader && fi.IsChecked && fi.Item != nil {
			total += fi.Item.SizeBytes
		}
	}
	return total
}

func (m *Model) toggleSelectAllSafe() {
	allSafeChecked := true
	for _, fi := range m.flatItems {
		if !fi.IsHeader && fi.Item != nil && fi.Item.Risk == model.RiskSafe && !fi.Item.IsProtected {
			if !fi.IsChecked {
				allSafeChecked = false
				break
			}
		}
	}

	for i := range m.flatItems {
		fi := &m.flatItems[i]
		if !fi.IsHeader && fi.Item != nil && fi.Item.Risk == model.RiskSafe && !fi.Item.IsProtected {
			fi.IsChecked = !allSafeChecked
		}
	}
}

func buildFlatItems(r *model.ScanReport) []FlatItem {
	var list []FlatItem
	if r == nil {
		return list
	}

	for _, g := range r.Groups {
		list = append(list, FlatItem{
			IsHeader: true,
			Group:    g,
		})
		for _, item := range g.Items {
			defaultChecked := item.Risk == model.RiskSafe && !item.IsProtected
			list = append(list, FlatItem{
				IsHeader:  false,
				Group:     g,
				Item:      item,
				IsChecked: defaultChecked,
			})
		}
	}
	return list
}

// ======================= 界面视图渲染 (Lip Gloss) =======================

var (
	subtleColor    = lipgloss.AdaptiveColor{Light: "#999999", Dark: "#666666"}
	highlightColor = lipgloss.AdaptiveColor{Light: "#06B6D4", Dark: "#22D3EE"} // Cyan
	lemonColor     = lipgloss.AdaptiveColor{Light: "#EAB308", Dark: "#FACC15"} // Lemon Yellow
	safeColor      = lipgloss.AdaptiveColor{Light: "#16A34A", Dark: "#4ADE80"} // Green
	rebuildColor   = lipgloss.AdaptiveColor{Light: "#CA8A04", Dark: "#FDE047"} // Yellow
	cautionColor   = lipgloss.AdaptiveColor{Light: "#EA580C", Dark: "#FB923C"} // Orange

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#0284C7")).
			Padding(0, 2)

	headerBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(subtleColor).
			Padding(0, 1)

	groupHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#38BDF8")).
				MarginTop(1)

	cursorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(highlightColor)

	detailBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#0284C7")).
			Padding(0, 1)

	keyHelpStyle = lipgloss.NewStyle().
			Foreground(subtleColor)

	keyBadgeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(highlightColor)

	scrollCueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FACC15")).
			Italic(true)
)

func (m Model) View() string {
	switch m.state {
	case stateScanning:
		return m.viewScanning()
	case stateDashboard:
		return m.viewDashboard()
	case stateCleaning:
		return m.viewCleaning()
	case stateDone:
		return m.viewDone()
	default:
		return ""
	}
}

func (m Model) viewScanning() string {
	var b strings.Builder
	b.WriteString("\n\n")
	b.WriteString(titleStyle.Render("🍋 DevLemon") + "\n\n")
	b.WriteString(fmt.Sprintf("  %s \033[1;36m正在全方位深度扫描系统与开发环境...\033[0m\n\n", m.spinner.View()))
	b.WriteString("  ⚡ 并发排查 Docker BuildKit 缓存与孤立数据卷\n")
	b.WriteString("  ⚡ 感知 iOS 模拟器运行状态（运行态避让保护已加锁）\n")
	b.WriteString("  ⚡ 启发式聚类嗅探代码工作区 (~/self, ~/daas 等)\n")
	b.WriteString("  ⚡ 16 线程并行统计 target/、.build、node_modules 体积\n")
	b.WriteString("  ⚡ 分析 Go、npm、uv、Homebrew 等包管理器下载层\n\n")
	b.WriteString(keyHelpStyle.Render("  扫描正在毫秒级多线程推进中，请稍候..."))
	return b.String()
}

func (m Model) viewDashboard() string {
	var b strings.Builder

	// 1. 顶部标题与磁盘容量状态
	title := titleStyle.Render("🍋 DevLemon v0.2.0")

	diskInfo := ""
	if m.report != nil {
		diskInfo = fmt.Sprintf("磁盘总容量: %s  |  已用: %s  |  可用: %s [%d%%]",
			model.FormatBytes(m.report.DiskTotal),
			model.FormatBytes(m.report.DiskUsed),
			model.FormatBytes(m.report.DiskFree),
			m.report.DiskCapacityPercent,
		)
	}

	selectedSummary := fmt.Sprintf("已选: %d 项 / \033[1;33m%s\033[0m", m.selectedCount(), model.FormatBytes(m.selectedBytes()))

	// 当前光标位置与条目总数提示
	navInfo := fmt.Sprintf("条目位置: [%d / %d]  |  按 [Tab] 在分类间极速跳转", m.cursor+1, len(m.flatItems))

	headerContent := fmt.Sprintf("%s\n%s\n%s  |  \033[36m%s\033[0m", title, diskInfo, selectedSummary, navInfo)
	b.WriteString(headerBoxStyle.Render(headerContent))
	b.WriteString("\n")

	// 2. 列表区域与动态高度计算
	if len(m.flatItems) == 0 {
		b.WriteString("\n  🎉 太棒了！未发现可清理的冗余项目，磁盘非常干净！\n\n")
	} else {
		// 根据终端高度动态计算可视行数（保证大屏幕看到更多，小屏幕不截断）
		maxVisible := m.height - 15
		if maxVisible < 10 {
			maxVisible = 10
		}
		if maxVisible > 25 {
			maxVisible = 25
		}

		startIdx := 0
		if m.cursor > maxVisible/2 {
			startIdx = m.cursor - maxVisible/2
		}
		endIdx := startIdx + maxVisible
		if endIdx > len(m.flatItems) {
			endIdx = len(m.flatItems)
			startIdx = endIdx - maxVisible
			if startIdx < 0 {
				startIdx = 0
			}
		}

		// 向上滚动提示
		if startIdx > 0 {
			b.WriteString(scrollCueStyle.Render(fmt.Sprintf("    ▲ ... 向上滚动查看上方 %d 项 ...", startIdx)))
			b.WriteString("\n")
		}

		for i := startIdx; i < endIdx; i++ {
			fi := m.flatItems[i]
			if fi.IsHeader {
				b.WriteString(groupHeaderStyle.Render(fmt.Sprintf(" %s", fi.Group.Title)))
				b.WriteString("\n")
				continue
			}

			pointer := "  "
			if i == m.cursor {
				pointer = cursorStyle.Render("❯ ")
			}

			check := "[ ]"
			if fi.Item.IsProtected {
				check = "🔒 "
			} else if fi.IsChecked {
				check = "[\033[32m✓\033[0m]"
			}

			riskBadge := "\033[32m[安全]\033[0m"
			switch fi.Item.Risk {
			case model.RiskRebuildable:
				riskBadge = "\033[33m[重构]\033[0m"
			case model.RiskCaution:
				riskBadge = "\033[31m[谨慎]\033[0m"
			}
			if fi.Item.IsProtected {
				riskBadge = "\033[36m[保护]\033[0m"
			}

			line := fmt.Sprintf("%s%s %-32s %10s  %s", pointer, check, truncate(fi.Item.Title, 32), fi.Item.SizeFormatted, riskBadge)
			if i == m.cursor {
				line = lipgloss.NewStyle().Bold(true).Render(line)
			}
			b.WriteString(line)
			b.WriteString("\n")
		}

		// 向下滚动提示
		if endIdx < len(m.flatItems) {
			remaining := len(m.flatItems) - endIdx
			b.WriteString(scrollCueStyle.Render(fmt.Sprintf("    ▼ ... 向下滚动查看更多项目 (还有 %d 项，按 Tab 直接跳转) ...", remaining)))
			b.WriteString("\n")
		}

		// 3. 当前选中项的详细说明面板
		if m.cursor < len(m.flatItems) {
			curr := m.flatItems[m.cursor]
			if !curr.IsHeader && curr.Item != nil {
				var d strings.Builder
				d.WriteString(fmt.Sprintf("\033[1m%s\033[0m (%s)\n", curr.Item.Title, curr.Item.SizeFormatted))
				d.WriteString(fmt.Sprintf("说明: %s\n", curr.Item.Description))
				if curr.Item.Path != "" {
					d.WriteString(fmt.Sprintf("路径: %s\n", curr.Item.Path))
				}
				if curr.Item.IsProtected {
					d.WriteString(fmt.Sprintf("\033[32m保护理由: %s\033[0m\n", curr.Item.ProtectReason))
				}
				b.WriteString(detailBoxStyle.Render(d.String()))
				b.WriteString("\n")
			}
		}
	}

	// 4. 底部快捷键指南
	helpBar := fmt.Sprintf(
		" %s 切换选中 | %s 分类跳转 | %s 全选安全项 | %s 一键清理 | %s 预演 | %s 重扫 | %s 退出",
		keyBadgeStyle.Render("[Space]"),
		keyBadgeStyle.Render("[Tab]"),
		keyBadgeStyle.Render("[A]"),
		keyBadgeStyle.Render("[Enter]"),
		keyBadgeStyle.Render("[D]"),
		keyBadgeStyle.Render("[R]"),
		keyBadgeStyle.Render("[Q]"),
	)
	b.WriteString(keyHelpStyle.Render(helpBar))
	b.WriteString("\n")

	return b.String()
}

func (m Model) viewCleaning() string {
	var b strings.Builder
	b.WriteString("\n\n")
	b.WriteString(titleStyle.Render("🍋 DevLemon - 正在执行清理") + "\n\n")
	b.WriteString(fmt.Sprintf("  %s \033[1;33m正在安全清理已选资源，请稍候...\033[0m\n\n", m.spinner.View()))
	b.WriteString("  • 释放无用缓存与项目构建产物\n")
	b.WriteString("  • 联动请求 macOS APFS 本地快照薄化释放空间\n\n")
	return b.String()
}

func (m Model) viewDone() string {
	var b strings.Builder
	b.WriteString("\n\n")

	if m.isDryRun {
		b.WriteString(titleStyle.Render("🔍 DevLemon - 模拟预演完成") + "\n\n")
		b.WriteString(fmt.Sprintf("  • 预演跳过项数: %d 项\n", m.cleanedCount))
		b.WriteString(fmt.Sprintf("  • 预计释放空间: \033[1;32m%s\033[0m\n\n", model.FormatBytes(m.cleanedBytes)))
		b.WriteString("  (预演模式未删除任何实际数据)\n\n")
	} else {
		b.WriteString(titleStyle.Render("🎉 DevLemon - 清理圆满完成！") + "\n\n")
		b.WriteString(fmt.Sprintf("  • 成功处理项目: %d 项\n", m.cleanedCount))
		b.WriteString(fmt.Sprintf("  • 本次释放空间: \033[1;32m%s\033[0m\n", model.FormatBytes(m.cleanedBytes)))
		b.WriteString("  • APFS 本地快照释放请求已成功触发。\n\n")
	}

	b.WriteString(keyHelpStyle.Render("  [Enter] 重新扫描  |  [Q] 退出程序\n"))
	return b.String()
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}

// RunTUI 启动 Bubble Tea TUI 主程序
func RunTUI(cfg *config.Config) error {
	p := tea.NewProgram(NewModel(cfg), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
