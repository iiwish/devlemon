package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"devlemon/internal/analyze"
	"devlemon/internal/cleaner"
	"devlemon/internal/config"
	"devlemon/internal/engine"
	"devlemon/internal/i18n"
	"devlemon/internal/model"
	"devlemon/internal/reporter"
	"devlemon/internal/rules"
	"devlemon/internal/tui"
	"devlemon/internal/version"
)

func main() {
	// 如果不带参数直接运行 devlemon，或显式指定 tui / -i，默认启动全屏交互式 TUI
	if len(os.Args) == 1 || (len(os.Args) >= 2 && (os.Args[1] == "tui" || os.Args[1] == "-i")) {
		cfg := loadConfig()
		cfg.DiscoverWorkspaces()
		if err := tui.RunTUI(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", i18n.T("TUI error", "TUI 启动异常"), err)
			os.Exit(1)
		}
		return
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "tui", "-i":
		cfg := loadConfig()
		cfg.DiscoverWorkspaces()
		_ = tui.RunTUI(cfg)
	case "scan":
		handleScan(os.Args[2:])
	case "clean":
		handleClean(os.Args[2:])
	case "analyze":
		handleAnalyze(os.Args[2:])
	case "rules":
		handleRules(os.Args[2:])
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
  analyze 分析目录占用，按体积列出最大的子项 (默认主目录)
  rules   列出清理规则；rules check <文件> 校验规则文件
  version 查看版本信息

通用参数 (scan / clean 均支持):
  --workspace <path>    添加自定义代码工作区路径 (例如: --workspace ~/Projects, 支持多次指定)
  --json                输出标准 JSON 格式 (供后续 SwiftUI 原生界面或脚本调用)
  --safe                仅处理 100% 绝对安全的缓存 (跳过可重构的 target/ 目录或模拟器重置)
  --dry-run             预演模式，仅模拟清理过程，不真正执行删除操作

示例:
  dl                      # 启动全屏交互 TUI
  dl scan                 # 快速扫描并输出报告
  dl analyze ~/Library    # 查看 ~/Library 下哪些目录最占空间
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
  analyze Show the largest items under a directory (default: home)
  rules   List cleanup rules; rules check <file> validates rule files
  version Show version information

Options:
  --workspace <path>    Add custom project workspace path (e.g. --workspace ~/Projects)
  --json                Output result in standard JSON format
  --safe                Only scan or clean 100% safe caches (skips rebuildable targets)
  --dry-run             Dry-run mode, simulate cleanup without deleting data

Examples:
  dl                      # Launch interactive TUI
  dl scan                 # Scan and display terminal summary
  dl analyze ~/Library    # Find what takes up space in ~/Library
  dl scan --json          # Output machine-readable JSON
  dl clean --safe         # Clean 100% safe caches
  dl clean --dry-run      # Simulate cleanup`)
	}
}

// loadConfig 默认配置 + 用户配置文件 (~/.config/devlemon/config.yaml)。配置文件有误时提示并使用默认配置
func loadConfig() *config.Config {
	cfg := config.DefaultConfig()
	if err := cfg.LoadUserFile(); err != nil {
		fmt.Fprintf(os.Stderr, "devlemon: %s: %v\n", i18n.T("ignoring config file", "配置文件有误，已忽略"), err)
	}
	return cfg
}

func handleScan(args []string) {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	jsonMode := fs.Bool("json", false, "输出 JSON 格式")
	safeOnly := fs.Bool("safe", false, "仅扫描绝对安全项")
	var workspaces string
	fs.StringVar(&workspaces, "workspace", "", "指定扫描的工作区根目录 (多个目录用逗号隔开)")

	_ = fs.Parse(args)

	cfg := loadConfig()
	cfg.SafeOnly = *safeOnly
	// 快速清理只扫描系统缓存，不遍历主目录寻找工作区
	if !cfg.SafeOnly {
		cfg.DiscoverWorkspaces()
	}
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
	jsonMode := fs.Bool("json", false, "输出 JSON 格式")
	var itemsFilter string
	fs.StringVar(&itemsFilter, "items", "", "指定要清理的 Item ID 列表 (英文逗号隔开)")
	var workspaces string
	fs.StringVar(&workspaces, "workspace", "", "指定工作区根目录")

	_ = fs.Parse(args)

	cfg := loadConfig()
	cfg.SafeOnly = *safeOnly
	// 快速清理只扫描系统缓存，不遍历主目录寻找工作区
	if !cfg.SafeOnly {
		cfg.DiscoverWorkspaces()
	}
	if workspaces != "" {
		for _, w := range strings.Split(workspaces, ",") {
			cfg.AddWorkspacePath(strings.TrimSpace(w))
		}
	}

	ctx := context.Background()
	eng := engine.NewEngine(cfg)

	report, err := eng.Scan(ctx)
	if err != nil {
		if *jsonMode {
			fmt.Printf(`{"error": %q}`+"\n", err.Error())
		} else {
			fmt.Fprintf(os.Stderr, "扫描出错: %v\n", err)
		}
		os.Exit(1)
	}

	var allowedIDs map[string]bool
	if itemsFilter != "" {
		allowedIDs = make(map[string]bool)
		for _, id := range strings.Split(itemsFilter, ",") {
			allowedIDs[strings.TrimSpace(id)] = true
		}
	}

	// 收集待清理项
	var targets []*model.Item
	for _, grp := range report.Groups {
		for _, item := range grp.Items {
			if item.IsProtected {
				continue
			}
			if allowedIDs != nil {
				if !allowedIDs[item.ID] {
					continue
				}
			} else if *safeOnly && item.Risk != model.RiskSafe {
				continue
			}
			targets = append(targets, item)
		}
	}

	if len(targets) == 0 {
		if *jsonMode {
			fmt.Println(`{"freed_bytes": 0, "freed_formatted": "0 B", "cleaned_count": 0, "errors": []}`)
		} else {
			fmt.Println("🎉 当前没有匹配的可清理项！")
		}
		return
	}

	if *dryRun {
		total := sumItems(targets)
		if *jsonMode {
			fmt.Printf(`{"freed_bytes": %d, "freed_formatted": %q, "cleaned_count": %d, "dry_run": true, "errors": []}`+"\n",
				total, model.FormatBytes(total), len(targets))
		} else {
			fmt.Printf("🔍 【Dry-Run 预演模式】发现 %d 个清理项，预计释放 %s 空间:\n", len(targets), model.FormatBytes(total))
			for _, t := range targets {
				fmt.Printf("  • [预演跳过] %-30s %10s\n", t.Title, t.SizeFormatted)
			}
		}
		return
	}

	if !*jsonMode {
		fmt.Printf("🚀 开始清理 %d 项资源，预计释放 %s 空间...\n", len(targets), model.FormatBytes(sumItems(targets)))
	}
	cln := cleaner.NewCleaner(false)
	freed, count, errs := cln.CleanItems(ctx, targets)

	if *jsonMode {
		var errStrs []string
		for _, e := range errs {
			errStrs = append(errStrs, fmt.Sprintf("%q", e.Error()))
		}
		fmt.Printf(`{"freed_bytes": %d, "freed_formatted": %q, "cleaned_count": %d, "errors": [%s]}`+"\n",
			freed, model.FormatBytes(freed), count, strings.Join(errStrs, ","))
	} else {
		fmt.Printf("\n✨ 清理完毕！成功处理 %d 项，释放空间: \033[1;32m%s\033[0m\n", count, model.FormatBytes(freed))
		if len(errs) > 0 {
			fmt.Printf("⚠️ 共有 %d 项清理失败:\n", len(errs))
			for _, e := range errs {
				fmt.Printf("  - %v\n", e)
			}
		}
	}
}

func handleAnalyze(args []string) {
	fs := flag.NewFlagSet("analyze", flag.ExitOnError)
	jsonMode := fs.Bool("json", false, "输出 JSON 格式")
	top := fs.Int("top", 20, "显示最大的前 N 项")
	// 允许路径写在参数前面：dl analyze ~/Library --top 10
	var dir string
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		dir, args = args[0], args[1:]
	}
	_ = fs.Parse(args)
	if dir == "" {
		dir = fs.Arg(0)
	}
	if dir == "" {
		dir = "~"
	}
	dir = config.ExpandHome(dir)

	res, err := analyze.Dir(context.Background(), dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", i18n.T("Analyze failed", "分析失败"), err)
		os.Exit(1)
	}
	if *top > 0 && len(res.Entries) > *top {
		res.Entries = res.Entries[:*top]
	}
	if *jsonMode {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
		return
	}

	fmt.Printf("\n📊 %s  \033[1m%s\033[0m\n", res.Path, model.FormatBytes(res.SizeBytes))
	fmt.Println(strings.Repeat("─", 65))
	for _, e := range res.Entries {
		pct := 0.0
		if res.SizeBytes > 0 {
			pct = float64(e.SizeBytes) / float64(res.SizeBytes) * 100
		}
		bar := strings.Repeat("█", int(pct/5+0.5))
		name := e.Name
		if e.IsDir {
			name += "/"
		}
		fmt.Printf("%10s %5.1f%%  %-20s %s\n", model.FormatBytes(e.SizeBytes), pct, bar, name)
	}
	fmt.Println()
}

func handleRules(args []string) {
	if len(args) > 0 && args[0] == "check" {
		os.Exit(checkRuleFiles(args[1:]))
	}

	set, err := rules.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", i18n.T("Failed to load built-in rules", "内置规则加载失败"), err)
		os.Exit(1)
	}
	fmt.Printf("\n📜 %s: %s\n", i18n.T("User rules directory", "用户规则目录"), rules.UserRulesDir())
	fmt.Println(strings.Repeat("─", 65))
	for _, group := range []string{rules.GroupAppCaches, rules.GroupPackageCaches} {
		fmt.Printf("\n[%s]\n", group)
		for _, r := range set.ForGroup(group) {
			source := "builtin"
			if !strings.HasPrefix(r.Source, "builtin/") {
				source = "user"
			}
			fmt.Printf("  %-28s %-11s %-7s %s\n", r.ID, r.Risk, source, r.Title.String())
		}
	}
	if len(set.Warnings) > 0 {
		fmt.Printf("\n⚠️  %s\n", i18n.T("Skipped user rules:", "已跳过的用户规则："))
		for _, w := range set.Warnings {
			fmt.Println("  - " + w)
		}
	}
	fmt.Println()
}

// checkRuleFiles 校验规则文件（供贡献者与 CI 使用），返回进程退出码
func checkRuleFiles(paths []string) int {
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, i18n.T("Usage: devlemon rules check <file.yaml>...", "用法: devlemon rules check <规则文件.yaml>..."))
		return 2
	}
	failed := false
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err == nil {
			var rs []rules.Rule
			if rs, err = rules.ParseFile(data, path); err == nil {
				var errs []string
				for _, r := range rs {
					if !r.IsEnabled() {
						continue
					}
					if verr := r.Validate(); verr != nil {
						errs = append(errs, verr.Error())
					}
				}
				if len(errs) == 0 {
					fmt.Printf("✅ %s (%d %s)\n", path, len(rs), i18n.T("rules", "条规则"))
					continue
				}
				err = fmt.Errorf("%s", strings.Join(errs, "\n   "))
			}
		}
		failed = true
		fmt.Printf("❌ %s\n   %v\n", path, err)
	}
	if failed {
		return 1
	}
	return 0
}

func sumItems(items []*model.Item) int64 {
	var total int64
	for _, it := range items {
		total += it.SizeBytes
	}
	return total
}
