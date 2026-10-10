<p align="center">
  <br>
  <h1 align="center">🍋 DevLemon (柠檬清理 CLI)</h1>
  <p align="center">
    <strong>专为开发者与 AI 时代打造的极简、轻量、具备运行态感知能力的智能磁盘清理工具。</strong>
  </p>
  <p align="center">
    <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square" alt="License"></a>
    <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.24+-00ADD8.svg?style=flat-square&logo=go" alt="Go Version"></a>
    <a href="https://apple.com/macos"><img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey.svg?style=flat-square&logo=apple" alt="Platform"></a>
    <a href="https://github.com/iiwish/devlemon/releases"><img src="https://img.shields.io/badge/release-v0.2.9-emerald.svg?style=flat-square" alt="Release"></a>
    <a href="#readme"><img src="https://img.shields.io/badge/alias-dl-yellow.svg?style=flat-square" alt="Alias"></a>
    <a href="#readme"><img src="https://img.shields.io/badge/size-%3C%204MB-orange.svg?style=flat-square" alt="Binary Size"></a>
  </p>
  <p align="center">
    <a href="README.md">English</a> •
    <a href="#-快速开始">快速开始</a> •
    <a href="#-全屏交互式终端-tui">终端 TUI</a> •
    <a href="#-核心优势与痛点">核心优势</a> •
    <a href="#-杀手级特性">特性详解</a> •
    <a href="#-指令速查与实战技巧">使用技巧</a> •
    <a href="#-项目路线图">路线图</a>
  </p>
  <br>
</p>

---

## 🚀 快速开始

### 1. 通过 Homebrew 一键安装（推荐）

```bash
brew install iiwish/tap/devlemon
```

> 💡 **提示**：安装后会自动配置短别名 **`dl`**，敲击两键即可随时启动！

### 2. 启动交互式 TUI 仪表盘（无需任何参数）

在终端中直接输入 **`dl`**：

```bash
dl
```

也可以通过 `go install` 安装：
```bash
go install github.com/iiwish/devlemon/cmd/devlemon@latest
```

### 3. 常用速查命令

| 指令 | 作用 |
| :--- | :--- |
| **`dl`** 或 `devlemon` | 启动全屏交互式 Bubble Tea 终端仪表盘 (默认) |
| **`dl scan`** | 快速扫描并输出格式化摘要报告 |
| **`dl scan --json`** | 输出标准 JSON 格式 (供脚本调用或原生 GUI 渲染) |
| **`dl clean --safe`** | 一键安全清理 100% 绝对安全的缓存与包管理器文件 |
| **`dl clean --dry-run`** | 预演模式，仅模拟清理过程，不真正删除任何文件 |

### 4. 运行 macOS 原生 SwiftUI 客户端 (菜单栏浮窗 + 深度清理)

```bash
make app      # 编译并打包出独立的 build/DevLemon.app (支持直接启动与 Notarization 公证)
make run-app  # 一键启动运行原生客户端
```

原生 App 亮点：
- 🍋 **状态栏常驻监测**：超平滑贝塞尔波形网速曲线、CPU/内存/磁盘精准度量。
- ⚡️ **快捷一键安全清理**：下拉菜单常驻安全垃圾快速检测与毫秒级释放，日常维护零打扰。
- 📱 **应用垃圾与缓存深探**：覆盖 Xcode、VSCode、微信、企业微信、飞书、钉钉及主流音乐客户端临时缓存，严格物理隔离保护聊天记录与关键凭证。
- 🛡️ **保守预选策略**：默认仅勾选完全绝对安全的系统维护缓存，应用与工程依赖由用户知情按需选定，告别误删恐惧。
- 🔍 **动态搜索与体积排序**：支持关键字即时筛选，一键按占用空间倒序排列，快速擒获磁盘黑洞。
- 🔒 **100% 本地离线运行**：零遥测、零数据回传、包含 Apple Privacy Manifest，详见 [隐私政策 (PRIVACY.md)](PRIVACY.md)。

---

## ⚡ 诞生背景：为什么需要 DevLemon？

现代开发者和 AI 辅助工作流每天都在无声无息地吞噬数十乃至上百 GB 的磁盘空间：
* **Docker & OrbStack**：难以察觉的 BuildKit 构建缓存层和已销毁容器残留的匿名 Volume，动辄堆积 **50GB ~ 100GB+**。
* **现代 AI 编程助手**：Cursor、Codex、Claude Code、Ollama 等工具在本地静默生成海量缓存、浏览器测试包与临时模型。
* **语言编译与依赖**：Rust `target/`、Swift `.build/`、Next.js `.next/` 以及前端 `node_modules` 分散在几十个历史工程中，每个都是空间黑洞。
* **iOS 模拟器沙盒**：多台设备与测试沙盒随时间膨胀到数十 GB。

而传统清理工具（如 CleanMyMac 或普通系统清理器）存在**严重的“上下文盲区”**：
> 💥 *它们要么根本识别不出容器和编译缓存，要么粗暴抹掉所有 iOS 模拟器，强行掐死您正在运行的测试进程与登录态！*

**DevLemon** 具备**运行态避让保护机制（Active State Shield）**，真正理解开发者的工作栈：知道什么正在运行、什么已经休眠、什么是可以随时安全回收的。

---

## 🖥️ 全屏交互式终端 TUI

基于 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 深度打造，兼顾美观与极客操控感：

```text
┌────────────────────────────────────────────────────────────────────────┐
│ 🍋 DevLemon v0.2.2                                                     │
│ 磁盘总容量: 460.38 GB  |  已用: 418.51 GB  |  可用: 41.88 GB [90%]      │
│ 已选: 12 项 / 34.20 GB  |  条目位置: [1 / 48]  |  按 [Tab] 在分类间跳转  │
└────────────────────────────────────────────────────────────────────────┘

 🐳 Docker / OrbStack 容器资源
  [✓] Docker BuildKit 构建缓存                 12.40 GB  [安全]
  [✓] 未使用的 Docker 镜像                      4.10 GB  [安全]
  [ ] 未挂载的 Docker 孤立数据卷                 8.50 GB  [谨慎]

 🛠️ iOS 模拟器环境（带运行态避让保护）
  🔒  模拟器: iPhone 16 Pro (Booted)            3.80 GB  [保护]
  [ ] 已关机设备: iPhone 15 Pro                 2.75 GB  [谨慎]

 📦 开发语言与包管理器缓存
  [✓] Go 编译中间缓存                           1.91 GB  [安全]
  [✓] npm 离线安装包缓存                        3.55 GB  [安全]
  [✓] Python uv 高速包缓存                      2.40 GB  [安全]

 📂 项目构建产物与依赖（自动聚类嗅探）
  [✓] my-rust-app / target                      8.02 GB  [重构]
  [✓] frontend-web / node_modules               4.12 GB  [重构]
    ▼ ... 向下滚动查看更多项目 (还有 38 项，按 Tab 直接跳转) ...

 ────────────────────────────────────────────────────────────────────────
 [Space] 切换选中 | [Tab] 分类跳转 | [A] 全选安全项 | [Enter] 一键清理 | [Q] 退出
```

* **自适应动态视口**：根据当前终端窗口高度动态计算展示条目，包含滚动指示（`▲` / `▼`）与实时位置指示。
* **`Tab` 跨分类极速跳跃**：按下 **`Tab`** / **`Shift+Tab`** 可在各大类（Docker、模拟器、系统缓存、代码构建）之间瞬间跳跃。
* **高响应动态动效**：扫描与清理全程带呼吸式点阵 Spinner 动态反馈。
* **中英双语智能自适应**：默认使用英文；在中文系统环境下自动切换为全中文（亦可通过 `DEVLEMON_LANG=zh` / `DEVLEMON_LANG=en` 显式指定）。

---

## 🌟 杀手级特性

### 🛡️ 1. 运行态避让保护（Active State Shield）
在执行清理前感知系统运行状态：
* **iOS 模拟器**：实时解析 `xcrun simctl`，对任何处于 `Booted`（运行中）的测试机**强制加锁保护**，绝不中断正在进行的 App 测试；仅精准清理处于 `Shutdown` 状态的历史设备。
* **容器服务**：保护正在运行中容器所引用的镜像与持久化卷。

### 🧠 2. 零配置工作区聚类嗅探（Repository Clustering Sniffer）
您无需繁琐配置 `--workspace ~/your-folder`。
DevLemon 能够在 5 毫秒内浅层扫描用户目录，基于现代工程签名（`.git`, `Cargo.toml`, `package.json`, `go.mod`, `Package.swift` 等）自动将代码根目录聚类识别出来。

### ⚡ 3. 16 并发协程体积计算池
目录收集与磁盘大小统计完全异步解耦。面对数十个包含数万微小文件的深度目录（如前端巨型 `node_modules` 与 Rust `target/`），16 协程并发在数秒内极速统计完毕。

### ⏳ 4. 项目休眠度评分（Project Dormancy Scoring）
不是所有的构建输出都应该直接抹掉。DevLemon 自动计算项目的最后更新天数：
* 休眠 $\ge 7$ 天的项目 $\rightarrow$ 标记为 **安全可重构项**。
* 当天仍在活跃开发的项目 $\rightarrow$ 增加友好提示，避免无意义的重新全量编译耗时。

### 🍏 5. APFS 本地快照联动释放
macOS 上删除了大文件后经常发现剩余空间没变，原因是 Time Machine 本地快照锁死了磁盘块。DevLemon 在清理后会自动向系统发起安全的快照 Thinning 请求，真正把物理磁盘空间还给系统。

---

## 🆚 为什么选择 DevLemon？

| 功能维度 | CleanMyMac | ncdu / dust | npkill / cargo-sweep | 🍋 **DevLemon (`dl`)** |
| :--- | :---: | :---: | :---: | :---: |
| **专为开发者设计** | ❌ 否 | ❌ 通用工具 | 🟡 单一语言工具 | 🟢 **原生深度支持** |
| **Docker BuildKit 与匿名卷透视** | ❌ 盲区 | ❌ 盲区 | ❌ 盲区 | 🟢 **深度解析与释放** |
| **iOS 模拟器运行态保护** | ❌ 粗暴/危险 | ❌ 盲区 | ❌ 无 | 🟢 **零中断测试保护** |
| **零配置代码工作区发现** | ❌ 否 | ❌ 需手动指定 | ❌ 需手动指定 | 🟢 **5ms 自动聚类识别** |
| **项目休眠度天数评分** | ❌ 否 | ❌ 否 | 🟡 部分支持 | 🟢 **智能冷热区分** |
| **极简短命令** | ❌ 无 | ❌ 无 | ❌ 无 | 🟢 **`dl` 随处可用** |
| **全屏现代终端 TUI** | ❌ 纯 GUI | 🟡 单色黑白 | 🟡 单一树形 | 🟢 **Bubble Tea 炫酷 TUI** |
| **中英语言自动自适应** | ❌ 商店绑定 | ❌ 仅英文 | ❌ 仅英文 | 🟢 **系统环境智能识别** |
| **二进制体积** | ~200MB App | <5MB | ~100MB Node | **< 4MB 静态二进制** |
| **后台驻留 / 隐私上报** | ❌ 臃肿常驻 | 🟢 无 | 🟢 无 | 🟢 **零常驻守护进程** |
| **APFS 快照自动释放联动** | ❌ 否 | ❌ 否 | ❌ 否 | 🟢 **自动释放** |

---

## 📖 指令速查与实战技巧

### 1. 启动全屏交互终端（最推荐）
```bash
dl
```

### 2. 命令行快速扫描
```bash
dl scan
```

### 3. 一键安全清理
仅清理 100% 绝对安全的编译下载缓存，跳过本地可重构代码构建与模拟器重置：
```bash
dl clean --safe
```

### 4. 模拟预演（Dry-Run）
预先查看可释放空间与清理条目，不改动任何实际数据：
```bash
dl clean --dry-run
```

### 5. 添加自定义工作区根目录
```bash
dl scan --workspace ~/Projects,~/Developer
```

### 6. 输出标准 JSON 格式
无缝对接 CI 流程、自动化脚本或 macOS 菜单栏原生应用：
```bash
dl scan --json | jq '.total_reclaimable_bytes'
```

---

## ⚙️ 进阶配置（可选）

DevLemon 支持**开箱即用零配置**。若您有特殊深度或白名单需求，可在 `~/.config/devlemon/config.yaml` 中配置：

```yaml
# 自定义工作区扫描根目录
workspace_paths:
  - ~/Developer
  - ~/Projects

# 项目休眠判定阈值（单位：天，默认 7 天）
dormancy_days: 14

# 目标清理目录特征名称
target_dir_names:
  - target         # Rust / Cargo
  - .build         # Swift Package Manager
  - node_modules   # Node.js
  - build          # C++ / CMake / Gradle
  - .next          # Next.js
  - DerivedData    # Xcode

# 目录探测最大递归深度
max_depth: 3
```

---

## 🗺️ 项目路线图

- [x] 超高性能 Go 核心引擎（< 4MB 独立二进制）
- [x] Docker BuildKit 编译缓存与孤立数据卷分析
- [x] iOS 模拟器运行态避让保护（Active State Shield）
- [x] 零配置代码工程自动聚类嗅探器
- [x] 16-Worker 异步并发目录测速池
- [x] 交互式终端 TUI 仪表盘（支持 Tab 快速跳跃与动态视口）
- [x] 官方 Homebrew Tap 支持与 `dl` 极速短命令别名
- [x] 系统环境自动语言识别（中英文自适应）
- [x] **macOS 原生菜单栏 App**（参考柠檬 Lite 设计的 SwiftUI 状态栏常驻浮窗 + 深度清理窗口）
- [ ] AI 时代专项大模型缓存嗅探（Ollama 模型权重、HuggingFace 断点缓存、Claude Code/Cursor 临时文件）
- [ ] Linux 系统与工具链支持（systemd 日志、pacman/apt 缓存、podman）

---

## 🤝 参与贡献

热烈欢迎提交 Issue 和 Pull Request！详情请参阅 [CONTRIBUTING.md](CONTRIBUTING.md)。

---

## 📄 开源协议

本项目采用 **MIT 开源协议**。详情请参阅 [`LICENSE`](LICENSE) 文件。

<p align="center">
  献给所有热爱清爽电脑与快速代码的开发者。🍋
</p>
