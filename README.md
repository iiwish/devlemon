<p align="center">
  <br>
  <h1 align="center">🍋 DevLemon</h1>
  <p align="center">
    <strong>The intelligent, context-aware disk cleanup tool built for developers and the AI era.</strong>
  </p>
  <p align="center">
    <a href="https://github.com/iiwish/devlemon/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square" alt="License"></a>
    <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.26+-00ADD8.svg?style=flat-square&logo=go" alt="Go Version"></a>
    <a href="https://apple.com/macos"><img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey.svg?style=flat-square&logo=apple" alt="Platform"></a>
    <a href="https://github.com/iiwish/devlemon/releases"><img src="https://img.shields.io/badge/release-v0.1.0-emerald.svg?style=flat-square" alt="Release"></a>
    <a href="#readme"><img src="https://img.shields.io/badge/size-%3C%204MB-orange.svg?style=flat-square" alt="Binary Size"></a>
  </p>
  <p align="center">
    <a href="README_zh.md">简体中文</a> •
    <a href="#-quick-start">Quick Start</a> •
    <a href="#-why-devlemon">Why DevLemon?</a> •
    <a href="#-features">Features</a> •
    <a href="#-architecture">Architecture</a> •
    <a href="#-roadmap">Roadmap</a>
  </p>
  <br>
</p>

---

## ⚡ The Problem

Modern developers and AI workflows churn through gigabytes of disk space daily:
* **Docker & OrbStack**: Hidden BuildKit cache layers and orphaned volumes eat **50GB–100GB+**.
* **Modern AI Workflows**: Agents (Cursor, Codex, Claude Code, Ollama, UV) quietly leave massive worktrees, runtime browser dumps, and model caches.
* **Build Targets**: Rust `target/`, Swift `.build/`, Next.js `.next/`, and `node_modules` silently multiply across dozens of forgotten projects.
* **iOS Simulators**: Simulator sandboxes expand into tens of gigabytes.

Traditional cleanup utilities (CleanMyMac, legacy cleaners) are **context-blind**:
> 💥 *They either miss developer caches entirely, or blindly wipe active Xcode simulators and kill ongoing container workflows!*

**DevLemon** is different. Built from the ground up with **Active State Shielding**, it understands what is running, what is dormant, and what can be safely reclaimed without breaking your flow.

---

## 🖥️ Terminal Experience

```text
🍋 DevLemon - Developer Context-Aware Disk Scanner
─────────────────────────────────────────────────────────────────
💾 Disk Capacity: 460.38 GB  |  Used: 418.51 GB  |  Avail: 41.88 GB [90% (Warning)]
✨ Total Reclaimable: 30.80 GB
─────────────────────────────────────────────────────────────────

🐳 Docker / OrbStack (Reclaimable: 8.19 GB)
  • Unused Docker Images                        1.06 GB  [🟢 Safe]
    └─ Stale images unreferenced by any running container (Active containers protected)
  • Dangling Local Volumes                      6.91 GB  [🟠 Caution]
    └─ Orphaned anonymous persistent volumes from deleted test containers
  • Docker BuildKit Cache                     233.00 MB  [🟢 Safe]
    └─ Intermediate build cache layers (Rebuilt automatically on demand)

🛠️ iOS Simulator Environments (Reclaimable: 11.00 GB)
  🔒 [PROTECTED] Simulator: HICANG Development UI    2.79 GB
     └─ 🟢 Currently Booted & Active. Locked to prevent test interruption!
  • Inactive Device: iPhone 17 Pro              2.75 GB  [🟠 Caution]
    └─ Dormant test sandbox & app caches. Safe to erase to factory state.
  • Inactive Device: Voimory Regression         2.95 GB  [🟠 Caution]
    └─ Dormant test sandbox & app caches. Safe to erase to factory state.

📦 Package Managers & Toolchains (Reclaimable: 11.61 GB)
  • Go Build Cache                              1.91 GB  [🟢 Safe]
  • Go Mod Cache (pkg/mod)                      2.84 GB  [🟢 Safe]
  • npm Offline Package Cache                   3.55 GB  [🟢 Safe]
  • Playwright Headless Browsers                3.10 GB  [🟢 Safe]

📂 Project Workspaces & Build Artifacts (Reclaimable: 37.40 GB)
  • sugrid-maco / target                        8.02 GB  [🟡 Rebuildable]
    └─ Rust target directory. Dormant for 29 days (Safe to clean)
  • kirje / target                              2.58 GB  [🟡 Rebuildable]
    └─ Rust target directory. Dormant for 28 days (Safe to clean)
  • Semlia / build                              2.43 GB  [🟡 Rebuildable]
    └─ Build output. Dormant for 10 days (Safe to clean)
  • sugrid / target                             3.11 GB  [🟡 Rebuildable]
    └─ ⚠️ Active project (modified 0 days ago, rebuild will take time)
```

---

## 🆚 Why DevLemon?

| Feature | CleanMyMac | ncdu / dust | npkill / cargo-sweep | 🍋 **DevLemon** |
| :--- | :---: | :---: | :---: | :---: |
| **Developer-First Design** | ❌ No | ❌ Generic | 🟡 Language-specific | 🟢 **Native** |
| **Docker BuildKit & Volume Intelligence** | ❌ Blind | ❌ Blind | ❌ Blind | 🟢 **Deep Inspection** |
| **Active Simulator Protection (Shield)** | ❌ Danger | ❌ Blind | ❌ No | 🟢 **Zero Downtime** |
| **Zero-Config Workspace Discovery** | ❌ No | ❌ Manual | ❌ Manual | 🟢 **Automatic (5ms)** |
| **Project Dormancy Scoring** | ❌ No | ❌ No | 🟡 Partial | 🟢 **Git & Timestamp Aware** |
| **Binary Footprint** | ~200MB App | <5MB | ~100MB Node | **< 4MB Static Binary** |
| **Background Daemons / Telemetry** | ❌ Heavy | 🟢 None | 🟢 None | 🟢 **Zero Daemons** |
| **APFS Snapshot Thinning** | ❌ No | ❌ No | ❌ No | 🟢 **Automatic** |

---

## 🌟 Key Highlights

### 🛡️ 1. Active State Shield (Zero Interruption)
DevLemon inspects running state before suggesting actions:
* **iOS Simulators**: Inspects `xcrun simctl` in real-time. Any device in `Booted` state is **strictly shielded**—your active app tests and login sessions are never destroyed.
* **Containers**: Protects images and volumes associated with running (`Up`) containers.

### 🧠 2. Zero-Config Workspace Clustering
You don't need to configure `--workspace ~/your-folder`.
DevLemon's heuristic sniffer evaluates directory clusters in `$HOME` (excluding OS folders). If a folder contains modern project signatures (`.git`, `Cargo.toml`, `package.json`, `go.mod`, `Package.swift`), DevLemon automatically catalogs it as a workspace root in milliseconds.

### ⏳ 3. Project Dormancy Scoring
Not all build targets should be wiped. DevLemon calculates the dormancy threshold:
* Projects untouched for $\ge 7$ days $\rightarrow$ Classified as **Safe Reclaimable targets**.
* Active projects touched today $\rightarrow$ Flagged with a warning to avoid unwanted rebuild lag.

### 🍏 4. APFS Snapshot Awareness
On macOS, deleting files often doesn't free space because Time Machine local APFS snapshots lock the blocks. DevLemon automatically triggers safe snapshot thinning after cleanup so your disk space actually returns.

---

## 🚀 Quick Start

### Homebrew (Recommended)
```bash
brew install iiwish/tap/devlemon
```

### Go Install
```bash
go install github.com/iiwish/devlemon/cmd/devlemon@latest
```

### From Source
```bash
git clone https://github.com/iiwish/devlemon.git
cd devlemon
go build -o devlemon ./cmd/devlemon
sudo mv devlemon /usr/local/bin/
```

---

## 📖 Usage & Recipes

### 1. Zero-Config Scan
```bash
devlemon scan
```

### 2. One-Click Safe Cleanup
Only clean 100% safe compiler and package caches (never touches simulators or project targets):
```bash
devlemon clean --safe
```

### 3. Dry-Run Mode
Inspect what would be cleaned without touching anything:
```bash
devlemon clean --dry-run
```

### 4. JSON Output (For Automation & GUI)
Pipe structured scan results into your scripts, CI, or native macOS apps:
```bash
devlemon scan --json | jq '.total_reclaimable_bytes'
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
- [x] APFS Snapshot thinning integration
- [ ] **Interactive TUI Dashboard** (Powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea))
- [ ] **Native macOS Menu Bar App** (Lightweight SwiftUI shell over CLI JSON stream)
- [ ] AI-Era artifact probe (Ollama weights, HuggingFace cache, Claude Code/Cursor worktrees)
- [ ] Linux support (systemd journal, apt/pacman caches, podman)

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

---

<p align="center">
  Crafted with care for developers who love clean machines and fast code. 🍋
</p>
