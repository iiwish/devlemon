package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"devlemon/internal/cleaner"
	"devlemon/internal/config"
	"devlemon/internal/engine"
	"devlemon/internal/i18n"
	"devlemon/internal/model"
	"devlemon/internal/reporter"
	"devlemon/internal/tui"
	"devlemon/internal/version"
)

func main() {
	// 如果不带参数直接运行 devlemon，或显式指定 tui / -i，默认启动全屏交互式 TUI
	if len(os.Args) == 1 || (len(os.Args) >= 2 && (os.Args[1] == "tui" || os.Args[1] == "-i")) {
		cfg := config.DefaultConfig()
		if err := tui.RunTUI(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", i18n.T("TUI error", "TUI 启动异常"), err)
			os.Exit(1)
		}
		return
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "tui", "-i":
		cfg := config.DefaultConfig()
		_ = tui.RunTUI(cfg)
	case "scan":
		handleScan(os.Args[2:])
	case "clean":
		handleClean(os.Args[2:])
	case "version", "-v", "--version":
		fmt.Printf("devlemon %s (%s/%s)\n", version.Version, runtime.GOOS, runtime.GOARCH)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf(i18n.T("Unknown command: %s\n", "未知指令: %s\n"), subcommand)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	if i18n.IsChinese() {
		fmt.Println(`🍋 DevLemon - 极简轻量级开发者智能磁盘清理工具

使用方式:
  devlemon <command> [options] 或 dl <command> [options]

指令列表:
  tui     启动全屏交互式终端仪表盘 (默认)
  scan    扫描当前系统的可清理缓存与构建产物
  clean   执行清理操作
  version 查看版本信息

通用参数 (scan / clean 均支持):
  --workspace <path>    添加自定义代码工作区路径 (例如: --workspace ~/Projects, 支持多次指定)
  --json                输出标准 JSON 格式 (供后续 SwiftUI 原生界面或脚本调用)
  --safe                仅处理 100% 绝对安全的缓存 (跳过可重构的 target/ 目录或模拟器重置)
  --dry-run             预演模式，仅模拟清理过程，不真正执行删除操作

示例:
  dl                      # 启动全屏交互 TUI
  dl scan                 # 快速扫描并输出报告
  dl scan --json          # 输出机器可读 JSON
  dl clean --safe         # 快速清理安全缓存
  dl clean --dry-run      # 模拟预演`)
	} else {
		fmt.Println(`🍋 DevLemon - Intelligent, Context-Aware Disk Cleanup Tool for Developers

Usage:
  devlemon <command> [options] or dl <command> [options]

Commands:
  tui     Launch interactive fullscreen terminal dashboard (default)
  scan    Scan reclaimable caches and build artifacts
  clean   Clean detected reclaimable resources
  version Show version information

Options:
  --workspace <path>    Add custom project workspace path (e.g. --workspace ~/Projects)
  --json                Output result in standard JSON format
  --safe                Only scan or clean 100% safe caches (skips rebuildable targets)
  --dry-run             Dry-run mode, simulate cleanup without deleting data

Examples:
  dl                      # Launch interactive TUI
  dl scan                 # Scan and display terminal summary
  dl scan --json          # Output machine-readable JSON
  dl clean --safe         # Clean 100% safe caches
  dl clean --dry-run      # Simulate cleanup`)
	}
}

func handleScan(args []string) {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	jsonMode := fs.Bool("json", false, "输出 JSON 格式")
	safeOnly := fs.Bool("safe", false, "仅扫描绝对安全项")
	var workspaces string
	fs.StringVar(&workspaces, "workspace", "", "指定扫描的工作区根目录 (多个目录用逗号隔开)")

	_ = fs.Parse(args)

	cfg := config.DefaultConfig()
	cfg.SafeOnly = *safeOnly
	if workspaces != "" {
		for _, w := range strings.Split(workspaces, ",") {
			cfg.AddWorkspacePath(strings.TrimSpace(w))
		}
	}

	ctx := context.Background()
	eng := engine.NewEngine(cfg)

	report, err := eng.Scan(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "扫描出错: %v\n", err)
		os.Exit(1)
	}

	if *jsonMode {
		r := reporter.NewJSONReporter(os.Stdout)
		_ = r.Render(report)
	} else {
		r := reporter.NewTerminalReporter(os.Stdout)
		r.Render(report)
	}
}

func handleClean(args []string) {
	fs := flag.NewFlagSet("clean", flag.ExitOnError)
	safeOnly := fs.Bool("safe", false, "仅清理 100% 绝对安全项")
	dryRun := fs.Bool("dry-run", false, "仅模拟演练，不真正删除")
	var workspaces string
	fs.StringVar(&workspaces, "workspace", "", "指定工作区根目录")

	_ = fs.Parse(args)

	cfg := config.DefaultConfig()
	cfg.SafeOnly = *safeOnly
	if workspaces != "" {
		for _, w := range strings.Split(workspaces, ",") {
			cfg.AddWorkspacePath(strings.TrimSpace(w))
		}
	}

	ctx := context.Background()
	eng := engine.NewEngine(cfg)

	report, err := eng.Scan(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "扫描出错: %v\n", err)
		os.Exit(1)
	}

	// 收集待清理项
	var targets []*model.Item
	for _, grp := range report.Groups {
		for _, item := range grp.Items {
			if item.IsProtected {
				continue
			}
			if *safeOnly && item.Risk != model.RiskSafe {
				continue
			}
			targets = append(targets, item)
		}
	}

	if len(targets) == 0 {
		fmt.Println("🎉 当前没有匹配的可清理项！")
		return
	}

	if *dryRun {
		fmt.Printf("🔍 【Dry-Run 预演模式】发现 %d 个清理项，预计释放 %s 空间:\n", len(targets), model.FormatBytes(sumItems(targets)))
		for _, t := range targets {
			fmt.Printf("  • [预演跳过] %-30s %10s\n", t.Title, t.SizeFormatted)
		}
		return
	}

	fmt.Printf("🚀 开始清理 %d 项资源，预计释放 %s 空间...\n", len(targets), model.FormatBytes(sumItems(targets)))
	cln := cleaner.NewCleaner(false)
	freed, count, errs := cln.CleanItems(ctx, targets)

	fmt.Printf("\n✨ 清理完毕！成功处理 %d 项，释放空间: \033[1;32m%s\033[0m\n", count, model.FormatBytes(freed))
	if len(errs) > 0 {
		fmt.Printf("⚠️ 共有 %d 项清理失败:\n", len(errs))
		for _, e := range errs {
			fmt.Printf("  - %v\n", e)
		}
	}
}

func sumItems(items []*model.Item) int64 {
	var total int64
	for _, it := range items {
		total += it.SizeBytes
	}
	return total
}
