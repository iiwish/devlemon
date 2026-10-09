# DevLemon 隐私政策 / Privacy Policy

**最后更新日期：2026年10月9日 / Last Updated: October 9, 2026**

DevLemon 致力于保护用户的个人隐私与数据安全。本隐私政策旨在向您说明 DevLemon macOS 客户端及命令行工具如何处理、访问和保护您的数据。

---

## 1. 核心原则：100% 本地离线运行 (Zero Data Collection)

- **无数据上传**：DevLemon 是一个完全在您本地计算机上独立运行的开发者工具与系统维护软件。我们**不会**将您的文件列表、代码路径、项目名称、扫描报告或任何数据上传至任何远程服务器或第三方云端。
- **无分析与广告 SDK**：DevLemon 不包含任何用户追踪（Tracking）、行为埋点、遥测（Telemetry）或广告追踪代码。
- **无个人身份信息收集**：我们不收集、不要求、不存储任何用户的个人身份信息（PII）。

## 2. 系统权限与数据访问说明

为了实现缓存检测与磁盘空间释放功能，DevLemon 会在获得您明确授权的前提下访问本地特定文件目录：

1. **项目构建产物与依赖**：
   - 扫描并计算指定目录下的 `node_modules`、`target`、`.build`、`build` 等临时构建中间件的大小与修改时间。
2. **应用与开发环境缓存**：
   - 检测 Xcode（`DerivedData`）、Docker/OrbStack、Homebrew、Go/Rust/Python 等包管理器以及主流桌面软件的临时网络与渲染缓存。
3. **硬件资源监测**：
   - 调用系统公开发布的 Mach / Darwin 内核性能接口（如 `host_processor_info`、`host_statistics64`）读取 CPU 与内存负载、以及通过网络接口计数器计算实时网络吞吐量，所有监控数据均仅在内存中用于界面实时渲染，不作持久化保存。

## 3. 安全作用域与沙盒合规 (Security-Scoped Bookmarks)

在开启沙盒的 Mac App Store 版本中，DevLemon 严格遵循 Apple 沙盒设计准则。DevLemon 仅会在您通过系统原生文件选取器（`NSOpenPanel`）显式选定并授权目录后，利用安全作用域书签（Security-Scoped Bookmark）临时访问该特定目录，并且所有操作均在本地受限沙盒内完成。

## 4. 软件自动更新 (仅限直装版)

在 GitHub / 官网独立分发版中，DevLemon 使用开源框架 [Sparkle](https://sparkle-project.org/) 进行版本更新检测。检查更新时仅会向 GitHub 静态资源库请求版本的版本号元数据 (`appcast.xml`)，不会附带任何用户身份标识。Mac App Store 版本则完全由苹果系统应用商店统一推送更新。

## 5. 政策变更与联系方式

如本政策有任何更新，我们将直接更新本仓库的 `PRIVACY.md` 文档。若您对本隐私政策有任何疑问或安全建议，可通过 GitHub Issues 与我们联系：
- 官方仓库: [https://github.com/iiwish/devlemon](https://github.com/iiwish/devlemon)

---

## English Summary

- **Local Only**: DevLemon operates 100% locally on your Mac. No project files, directory paths, or scan reports are ever transmitted over the network.
- **No Tracking**: No telemetry, analytics, or advertising SDKs are included.
- **Data Privacy**: Strictly complies with Apple App Store Review Guidelines and Privacy Manifest requirements.
