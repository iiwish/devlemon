<p align="center">
  <br>
  <h1 align="center">🍋 DevLemon (柠檬清理 CLI)</h1>
  <p align="center">
    <strong>专为开发者与 AI 时代打造的极简、轻量、具备运行态感知能力的智能磁盘清理工具。</strong>
  </p>
  <p align="center">
    <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square" alt="License"></a>
    <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.26+-00ADD8.svg?style=flat-square&logo=go" alt="Go Version"></a>
    <a href="https://apple.com/macos"><img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey.svg?style=flat-square&logo=apple" alt="Platform"></a>
    <a href="#readme"><img src="https://img.shields.io/badge/size-%3C%204MB-orange.svg?style=flat-square" alt="Binary Size"></a>
  </p>
  <p align="center">
    <a href="README.md">English</a> •
    <a href="#-快速开始">快速开始</a> •
    <a href="#-核心优势与痛点">核心优势</a> •
    <a href="#-功能特性">功能特性</a> •
    <a href="#-架构与路线图">架构与路线图</a>
  </p>
  <br>
</p>

---

## ⚡ 诞生背景：为什么需要 DevLemon？

现代开发者和 AI 辅助工作流每天都在无声无息地吞噬数十乃至上百 GB 的磁盘空间：
* **Docker & OrbStack**：难以察觉的 BuildKit 构建缓存层和已销毁容器残留的匿名 Volume，动辄堆积 **50GB ~ 100GB+**。
* **现代 AI 编程助手**：Cursor、Codex、Claude Code、Ollama 等工具，在本地静默生成海量的 git worktrees、浏览器驱动包与临时模型缓存。
* **语言编译与依赖**：Rust `target/`、Swift `.build/`、Next.js `.next/` 以及前端 `node_modules` 分散在几十个历史工程中，每个都是空间黑洞。
* **iOS 模拟器沙盒**：多台设备与测试沙盒随时间膨胀到数十 GB。

而传统清理工具（如 CleanMyMac 或普通系统清理器）存在**严重的“上下文盲区”**：
> 💥 *它们要么根本识别不出容器和编译缓存，要么粗暴抹掉所有 iOS 模拟器，强行掐死您正在运行的测试进程与登录态！*

**DevLemon** 应运而生。它具备**运行态避让保护机制（Active State Shield）**，真正理解开发者的工作栈：知道什么正在运行、什么已经休眠、什么是可以随时安全回收的。

---

## 🖥️ 终端交互体验

```text
🍋 DevLemon - 开发者智能磁盘扫描报告
─────────────────────────────────────────────────────────────────
💾 磁盘总容量: 460.38 GB  |  已用: 418.51 GB  |  可用: 41.88 GB [90% (空间紧张)]
✨ 累计可回收: 30.80 GB
─────────────────────────────────────────────────────────────────

🐳 Docker / OrbStack 容器资源 (组内可回收: 8.19 GB)
  • 未使用的 Docker 镜像                        1.06 GB  [🟢 绝对安全]
    └─ 未被任何运行中容器引用的旧版本镜像（已自动保护活跃容器镜像）
  • 未挂载的 Docker 孤立数据卷                     6.91 GB  [🟠 谨慎操作]
    └─ 容器已销毁但残留的历史匿名数据卷
  • Docker BuildKit 构建缓存                233.00 MB  [🟢 绝对安全]
    └─ 中间编译层与 Dockerfile 构建缓存，清空后按需自动重新构建

🛠️ iOS 模拟器环境（带运行态避让保护） (组内可回收: 11.00 GB)
  🔒 [受保护] 模拟器: HICANG Development UI        2.79 GB
     └─ 🟢 正在运行中 (Booted)，已自动加锁保护，绝不中断测试！
  • 已关机设备: iPhone 17 Pro                  2.75 GB  [🟠 谨慎操作]
    └─ 历史测试残留数据与应用沙盒缓存，重置后恢复初始洁净状态
  • 已关机设备: Voimory Regression             2.95 GB  [🟠 谨慎操作]
    └─ 历史测试残留数据与应用沙盒缓存，重置后恢复初始洁净状态

📦 开发语言与包管理器缓存 (组内可回收: 11.61 GB)
  • Go 编译中间缓存                             1.91 GB  [🟢 绝对安全]
  • Go 模块下载缓存 (pkg/mod)                   2.84 GB  [🟢 绝对安全]
  • npm 离线安装包缓存                           3.55 GB  [🟢 绝对安全]
  • Playwright 自动化无头浏览器包                  3.10 GB  [🟢 绝对安全]

📂 项目构建产物与依赖（target/、.build、node_modules） (组内可回收: 37.40 GB)
  • sugrid-maco / target                        8.02 GB  [🟡 可重构建]
    └─ Rust target 目录，距最后修改约 29 天（已休眠，可安全清理）
  • kirje / target                              2.58 GB  [🟡 可重构建]
    └─ Rust target 目录，距最后修改约 28 天（已休眠，可安全清理）
  • Semlia / build                              2.43 GB  [🟡 可重构建]
    └─ 构建产物，距最后修改约 10 天（已休眠，可安全清理）
  • sugrid / target                             3.11 GB  [🟡 可重构建]
    └─ ⚠️ 近期活跃项目（0 天前刚修改，重新编译将耗时）
```

---

## 🆚 为什么选择 DevLemon？

| 功能维度 | CleanMyMac | ncdu / dust | npkill / cargo-sweep | 🍋 **DevLemon** |
| :--- | :---: | :---: | :---: | :---: |
| **专为开发者设计** | ❌ 否 | ❌ 通用工具 | 🟡 单一语言工具 | 🟢 **原生深度支持** |
| **Docker BuildKit 与匿名卷透视** | ❌ 盲区 | ❌ 盲区 | ❌ 盲区 | 🟢 **深度解析与释放** |
| **iOS 模拟器运行态保护** | ❌ 粗暴/危险 | ❌ 盲区 | ❌ 无 | 🟢 **零中断测试保护** |
| **零配置代码工作区发现** | ❌ 否 | ❌ 需手动指定 | ❌ 需手动指定 | 🟢 **5ms 自动聚类识别** |
| **项目休眠度天数评分** | ❌ 否 | ❌ 否 | 🟡 部分支持 | 🟢 **智能冷热区分** |
| **二进制体积** | ~200MB App | <5MB | ~100MB Node | **< 4MB 静态二进制** |
| **后台驻留 / 隐私上报** | ❌ 臃肿常驻 | 🟢 无 | 🟢 无 | 🟢 **零常驻守护进程** |
| **APFS 快照自动释放联动** | ❌ 否 | ❌ 否 | ❌ 否 | 🟢 **自动释放** |

---

## 🌟 核心杀手级特性

### 🛡️ 1. 运行态避让保护（Active State Shield）
在执行清理前感知系统运行状态：
* **iOS 模拟器**：实时解析 `xcrun simctl`，对任何处于 `Booted`（运行中）的测试机**强制加锁保护**，绝不中断正在进行的 App 测试；仅精准清理处于 `Shutdown` 状态的历史设备。
* **容器服务**：保护正在运行中容器所引用的镜像与持久化卷。

### 🧠 2. 零配置工作区聚类嗅探（Repository Clustering Sniffer）
您无需繁琐配置 `--workspace ~/your-folder`。
DevLemon 能够在 5 毫秒内浅层扫描用户目录，基于现代工程签名（`.git`, `Cargo.toml`, `package.json`, `go.mod`, `Package.swift` 等）自动将代码根目录聚类识别出来。

### ⏳ 3. 项目休眠度评分（Project Dormancy Scoring）
不是所有的构建输出都应该直接抹掉。DevLemon 自动计算项目的最后更新天数：
* 休眠 $\ge 7$ 天的项目 $\rightarrow$ 标记为 **安全可重构项**。
* 当天仍在活跃开发的项目 $\rightarrow$ 增加友好提示，避免无意义的重新全量编译耗时。

### 🍏 4. APFS 本地快照联动释放
macOS 上删除了大文件后经常发现剩余空间没变，原因是 Time Machine 本地快照锁死了磁盘块。DevLemon 在清理后会自动向系统发起安全的快照 Thinning 请求，真正把磁盘空间还给系统。

---

## 🚀 快速开始

### 方式 1：Homebrew 安装（推荐）
```bash
brew install iiwish/tap/devlemon
```

### 方式 2：Go Install
```bash
go install github.com/iiwish/devlemon/cmd/devlemon@latest
```

### 方式 3：源码极速编译
```bash
git clone https://github.com/iiwish/devlemon.git
cd devlemon
go build -o devlemon ./cmd/devlemon
sudo mv devlemon /usr/local/bin/
```

---

## 📖 常用指令

```bash
# 1. 全自动零配置扫描（自动发现系统缓存、Docker 与所有代码工作区）
devlemon scan

# 2. 一键安全清理（仅清理 100% 绝对安全的绿色缓存，不碰模拟器与项目 target）
devlemon clean --safe

# 3. 预演模式（Dry-Run：仅查看可清理项，不实际删除任何文件）
devlemon clean --dry-run

# 4. JSON 流式输出（专供脚本自动化或上层 SwiftUI 界面消费）
devlemon scan --json | jq '.total_reclaimable_bytes'
```

---

## 🗺️ 架构与路线图

- [x] 高性能 Go 极简核心（< 4MB 单二进制）
- [x] Docker BuildKit 缓存与孤立 Volume 深度回收
- [x] iOS 模拟器运行态避让保护机制
- [x] 零配置代码工程聚类嗅探器
- [x] APFS 本地快照薄化联动
- [ ] **交互式终端 TUI 仪表盘**（基于 [Bubble Tea](https://github.com/charmbracelet/bubbletea)）
- [ ] **原生 macOS 菜单栏状态栏应用**（基于 SwiftUI 原生渲染，通过 JSON 管道驱动 Go 核心）
- [ ] 更多 AI 时代产物支持（Ollama 本地模型、HuggingFace 权重、Cursor/Claude Code 历史工作区）
- [ ] Linux 开发者环境支持（systemd 日志、Podman、包管理器缓存）

---

## 📄 开源许可证

本项目基于 **MIT License** 开源，详情请参阅 [`LICENSE`](LICENSE)。

---

<p align="center">
  专为热爱干净机器与飞快代码的开发者精心雕琢。🍋
</p>
