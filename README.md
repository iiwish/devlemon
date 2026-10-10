<p align="center">
  <br>
  <h1 align="center">🍋 DevLemon</h1>
  <p align="center">
    <strong>The intelligent, context-aware disk cleanup tool built for developers and the AI era.</strong>
  </p>
  <p align="center">
    <a href="https://github.com/iiwish/devlemon/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square" alt="License"></a>
    <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.24+-00ADD8.svg?style=flat-square&logo=go" alt="Go Version"></a>
    <a href="https://apple.com/macos"><img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey.svg?style=flat-square&logo=apple" alt="Platform"></a>
    <a href="https://github.com/iiwish/devlemon/releases"><img src="https://img.shields.io/badge/release-v0.2.7-emerald.svg?style=flat-square" alt="Release"></a>
    <a href="#readme"><img src="https://img.shields.io/badge/alias-dl-yellow.svg?style=flat-square" alt="Alias"></a>
    <a href="#readme"><img src="https://img.shields.io/badge/size-%3C%204MB-orange.svg?style=flat-square" alt="Binary Size"></a>
  </p>
  <p align="center">
    <a href="README_zh.md">简体中文</a> •
    <a href="#-quick-start">Quick Start</a> •
    <a href="#-interactive-tui-experience">Interactive TUI</a> •
    <a href="#-key-highlights">Highlights</a> •
    <a href="#-why-devlemon">Why DevLemon?</a> •
    <a href="#-usage--recipes">Usage</a> •
    <a href="#-roadmap">Roadmap</a>
  </p>
  <br>
</p>

---

## 🚀 Quick Start

### 1. Install via Homebrew (macOS & Linux)

```bash
brew install iiwish/tap/devlemon
```

> 💡 **Tip**: Both `devlemon` and the ultra-short alias **`dl`** are installed automatically!

### 2. Launch Interactive TUI (Zero Arguments)

Simply type **`dl`** in your terminal:

```bash
dl
```

Or run via `go install`:
```bash
go install github.com/iiwish/devlemon/cmd/devlemon@latest
```

### 3. Cheat Sheet

| Command | Action |
| :--- | :--- |
| **`dl`** or `devlemon` | Launch fullscreen interactive TUI dashboard (default) |
| **`dl scan`** | Run non-interactive terminal scan & print summary report |
| **`dl scan --json`** | Output machine-readable JSON (for scripts / native GUI) |
| **`dl clean --safe`** | One-click clean 100% safe compiler & package caches |
| **`dl clean --dry-run`** | Simulate cleanup without deleting any data |

### 4. Native macOS Client (MenuBar Extra + Deep Clean)

```bash
make app      # Build self-contained build/DevLemon.app (Supports Direct & Apple Notarization)
make run-app  # Launch the native macOS app
```

Features:
- 🍋 **MenuBar Live Monitor**: Real-time smooth Bézier waveform network graph, CPU, RAM and SSD gauges.
- ⚡️ **Quick Safe Clean**: Instant scan & purge safe system caches directly from the menubar popover.
- 📱 **App Caches & Garbage**: Deep detection for Xcode, VSCode, WeChat, Lark, DingTalk, and music streamers while strictly protecting chats and credentials.
- 🛡️ **Conservative Pre-selection**: Only 100% safe system caches are pre-checked by default; dependency items require explicit confirmation.
- 🔍 **Filter & Sort**: Dynamic text search and sort-by-size toggle.
- 🔒 **100% Local & Privacy-First**: Zero telemetry, zero cloud tracking, includes Apple Privacy Manifest. See [Privacy Policy](PRIVACY.md).

---

## ⚡ The Problem

Modern developers and AI workflows churn through gigabytes of disk space daily:
* **Docker & OrbStack**: Hidden BuildKit cache layers and orphaned volumes eat **50GB–100GB+**.
* **Modern AI Workflows**: Agents (Cursor, Codex, Claude Code, Ollama, UV) quietly leave massive worktrees, runtime browser dumps, and model caches.
* **Build Targets**: Rust `target/`, Swift `.build/`, Next.js `.next/`, and `node_modules` silently multiply across dozens of forgotten projects.
* **iOS Simulators**: Simulator sandboxes expand into tens of gigabytes.

Traditional cleanup utilities (CleanMyMac, legacy cleaners) are **context-blind**:
> 💥 *They either miss developer caches entirely, or blindly wipe active Xcode simulators and kill ongoing container workflows!*

**DevLemon** is built from the ground up with **Active State Shielding**—it understands what is running, what is dormant, and what can be safely reclaimed without breaking your flow.

---

## 🖥️ Interactive TUI Experience

DevLemon includes a rich, responsive terminal interface powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea):

```text
┌────────────────────────────────────────────────────────────────────────┐
│ 🍋 DevLemon v0.2.2                                                     │
│ Total Disk: 460.38 GB  |  Used: 418.51 GB  |  Free: 41.88 GB [90%]     │
│ Selected: 12 items / 34.20 GB  |  Item: [1 / 48]  |  Press [Tab] Jump  │
└────────────────────────────────────────────────────────────────────────┘

 🐳 Docker / OrbStack Container Resources
  [✓] Docker BuildKit Build Cache              12.40 GB  [Safe]
  [✓] Unused Docker Images                      4.10 GB  [Safe]
  [ ] Docker Dangling Volumes                   8.50 GB  [Caution]

 🛠️ iOS Simulator Environment (Active Shield Protected)
  🔒  Simulator: iPhone 16 Pro (Booted)         3.80 GB  [Protected]
  [ ] Shutdown Device: iPhone 15 Pro            2.75 GB  [Caution]

 📦 Package Manager & Developer Caches
  [✓] Go Build Cache                            1.91 GB  [Safe]
  [✓] npm Package Cache                         3.55 GB  [Safe]
  [✓] Python uv Package Cache                   2.40 GB  [Safe]

 📂 Workspace Builds & Dependencies (Auto-Clustered)
  [✓] my-rust-app / target                      8.02 GB  [Rebuild]
  [✓] frontend-web / node_modules               4.12 GB  [Rebuild]
    ▼ ... Scroll down for more items (38 remaining, press Tab to jump) ...

 ────────────────────────────────────────────────────────────────────────
 [Space] Toggle | [Tab] Jump Group | [A] All Safe | [Enter] Clean | [Q] Quit
```

* **Dynamic Viewport**: Automatically adjusts to your terminal height; includes scroll markers (`▲` / `▼`) and item counters.
* **Instant Category Jumping**: Hit **`Tab`** / **`Shift+Tab`** to jump across major categories in milliseconds.
* **Active Spinner**: Visual feedback during scanning and cleaning.
* **Auto i18n**: English by default; automatically switches to Chinese on Chinese systems (override anytime with `DEVLEMON_LANG=en` or `DEVLEMON_LANG=zh`).

---

## 🌟 Key Highlights

### 🛡️ 1. Active State Shield (Zero Interruption)
* **iOS Simulators**: Inspects `xcrun simctl` in real-time. Any device in `Booted` state is **strictly shielded**—active app tests and login sessions are never destroyed.
* **Containers**: Protects images and volumes associated with running (`Up`) containers.

### 🧠 2. Zero-Config Workspace Clustering Sniffer
You don't need to configure `--workspace ~/your-folder`.
DevLemon's heuristic sniffer evaluates directory clusters in `$HOME` (excluding OS folders). If a folder contains modern project signatures (`.git`, `Cargo.toml`, `package.json`, `go.mod`, `Package.swift`), DevLemon catalogs it as a workspace root in **5 milliseconds**.

### ⚡ 3. 16-Worker Concurrent Sizing Pool
Directory scanning and disk usage calculation are completely decoupled. Sizing thousands of subdirectories (including giant `node_modules` and Rust `target/` folders) is handled by a 16-goroutine parallel worker pool.

### ⏳ 4. Project Dormancy Scoring
Not all build targets should be wiped. DevLemon calculates the dormancy threshold:
* Projects untouched for $\ge 7$ days $\rightarrow$ Classified as **Safe Reclaimable targets**.
* Active projects touched today $\rightarrow$ Flagged with a warning to avoid unwanted rebuild lag.

### 🍏 5. APFS Local Snapshot Thinning
On macOS, deleting files often doesn't free physical blocks because Time Machine local APFS snapshots lock them. DevLemon automatically triggers safe snapshot thinning after cleanup so your disk space actually returns.

---

## 🆚 Why DevLemon?

| Feature | CleanMyMac | ncdu / dust | npkill / cargo-sweep | 🍋 **DevLemon (`dl`)** |
| :--- | :---: | :---: | :---: | :---: |
| **Developer-First Design** | ❌ No | ❌ Generic | 🟡 Language-specific | 🟢 **Native** |
| **Docker BuildKit & Volume Intelligence** | ❌ Blind | ❌ Blind | ❌ Blind | 🟢 **Deep Inspection** |
| **Active Simulator Protection (Shield)** | ❌ Danger | ❌ Blind | ❌ No | 🟢 **Zero Downtime** |
| **Zero-Config Workspace Discovery** | ❌ No | ❌ Manual | ❌ Manual | 🟢 **Automatic (5ms)** |
| **Project Dormancy Scoring** | ❌ No | ❌ No | 🟡 Partial | 🟢 **Git & ModTime Aware** |
| **Short Command Alias** | ❌ None | ❌ None | ❌ None | 🟢 **`dl`** |
| **Interactive TUI Dashboard** | ❌ GUI only | 🟡 Monochromatic | 🟡 Single-tree | 🟢 **Rich Bubble Tea TUI** |
| **Automatic Internationalization** | ❌ App Store | ❌ English | ❌ English | 🟢 **En / Zh Auto Detect** |
| **Binary Footprint** | ~200MB App | <5MB | ~100MB Node | **< 4MB Static Binary** |
| **Background Daemons / Telemetry** | ❌ Heavy | 🟢 None | 🟢 None | 🟢 **Zero Daemons** |
| **APFS Snapshot Thinning** | ❌ No | ❌ No | ❌ No | 🟢 **Automatic** |

---

## 📖 Usage & Recipes

### 1. Interactive Dashboard (Recommended)
```bash
dl
```

### 2. Fast Terminal Scan
```bash
dl scan
```

### 3. One-Click Safe Cleanup
Only clean 100% safe compiler and package caches (never touches simulators or project targets):
```bash
dl clean --safe
```

### 4. Dry-Run Mode
Simulate what would be cleaned without touching anything:
```bash
dl clean --dry-run
```

### 5. Custom Workspace Directory
```bash
dl scan --workspace ~/Projects,~/Developer
```

### 6. JSON Output (For Automation & GUI)
Pipe structured scan results into your scripts, CI, or native macOS apps:
```bash
dl scan --json | jq '.total_reclaimable_bytes'
```

---

## ⚙️ Configuration (Optional)

DevLemon works **zero-config out of the box**. If you want custom behavior, place `~/.config/devlemon/config.yaml`:

```yaml
# Add custom project workspace roots
workspace_paths:
  - ~/Developer
  - ~/Projects

# Threshold for dormant projects (default: 7 days)
dormancy_days: 14

# Directories recognized as safe rebuildable artifacts
target_dir_names:
  - target         # Rust / Cargo
  - .build         # Swift Package Manager
  - node_modules   # Node.js
  - build          # C++ / CMake / Gradle
  - .next          # Next.js
  - DerivedData    # Xcode

# Max directory depth to search
max_depth: 3
```

---

## 🗺️ Roadmap

- [x] High-performance Go core (< 4MB single binary)
- [x] Docker BuildKit & Dangling Volume cleaner
- [x] iOS Simulator Active State Shield
- [x] Zero-config project workspace clustering sniffer
- [x] 16-worker parallel sizing pool
- [x] Dynamic Bubble Tea TUI Dashboard with Tab jump navigation
- [x] Official Homebrew tap with `dl` short command
- [x] Automatic system language detection (EN / ZH)
- [x] **Native macOS Menu Bar App** (SwiftUI status bar monitor & popover inspired by Lemon Lite)
- [ ] AI-Era artifact probe (Ollama weights, HuggingFace cache, Claude Code/Cursor worktrees)
- [ ] Linux system cache probes (systemd journal, pacman/apt cache, podman)

---

## 🤝 Contributing

Contributions are warmly welcomed! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingProbe`)
3. Commit your Changes (`git commit -m 'Add some AmazingProbe'`)
4. Push to the Branch (`git push origin feature/AmazingProbe`)
5. Open a Pull Request

---

## 📄 License

Distributed under the **MIT License**. See [`LICENSE`](LICENSE) for more information.

<p align="center">
  Crafted with care for developers who love clean machines and fast code. 🍋
</p>
