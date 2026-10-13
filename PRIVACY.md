# DevLemon 隐私政策 / Privacy Policy

**最后更新日期：2026年10月13日 / Last Updated: October 13, 2026**

DevLemon 致力于保护用户的个人隐私与数据安全。本隐私政策旨在向您说明 DevLemon macOS 客户端及命令行工具如何处理、访问和保护您的数据。

---

## 1. 核心原则：100% 本地离线运行 (Zero Data Collection)

- **无数据上传**：DevLemon 是一个完全在您本地计算机上独立运行的开发者工具与系统维护软件。我们**不会**将您的文件列表、代码路径、项目名称、扫描报告或任何数据上传至任何远程服务器或第三方云端。
- **无分析与广告 SDK**：DevLemon 不包含任何用户追踪（Tracking）、行为埋点、遥测（Telemetry）或广告追踪代码。
- **无个人身份信息收集**：我们不收集、不要求、不存储任何用户的个人身份信息（PII）。

## 2. 系统权限与数据访问说明

为了实现缓存检测与磁盘空间释放功能，DevLemon 会读取您用户目录下的以下内容。所有读取都在本机完成，删除操作只会在您确认（勾选清理项或执行 `clean` 命令）后执行：

1. **项目构建产物与依赖**：
   - 扫描并计算指定目录下的 `node_modules`、`target`、`.build`、`build` 等临时构建中间件的大小与修改时间。
2. **应用与开发环境缓存**：
   - 检测 Xcode（`DerivedData`）、Docker/OrbStack、Homebrew、Go/Rust/Python 等包管理器以及主流桌面软件的临时网络与渲染缓存。
3. **硬件资源监测**：
   - 调用系统公开发布的 Mach / Darwin 内核性能接口（如 `host_processor_info`、`host_statistics64`）读取 CPU 与内存负载、以及通过网络接口计数器计算实时网络吞吐量，所有监控数据均仅在内存中用于界面实时渲染，不作持久化保存。
4. **运行中的应用与已安装应用**：
   - 读取运行中进程的可执行文件路径与内存占用，用于判断缓存所属应用是否正在运行，以及在菜单栏展示各应用内存占用；
   - 读取已安装应用的 `Info.plist`（Bundle ID、版本号），用于识别已卸载应用的残留数据与待安装的更新包。
   - 菜单栏中的「退出」「强制退出」只会在您点击后作用于对应应用。

## 3. 管理员权限

菜单栏内存面板的「释放」功能会调用系统自带的 `purge` 命令，需要管理员权限。密码由 macOS 系统授权对话框直接处理，DevLemon 无法获取也不会保存您的密码。除此之外 DevLemon 不请求任何管理员权限。

## 4. 软件自动更新

DevLemon 使用开源框架 [Sparkle](https://sparkle-project.org/) 进行版本更新检测。检查更新时仅会向 GitHub 请求版本元数据 (`appcast.xml`)，不会附带任何用户身份标识；可在设置中关闭自动检查。

## 5. 政策变更与联系方式

如本政策有任何更新，我们将直接更新本仓库的 `PRIVACY.md` 文档。若您对本隐私政策有任何疑问或安全建议，可通过 GitHub Issues 与我们联系：
- 官方仓库: [https://github.com/iiwish/devlemon](https://github.com/iiwish/devlemon)

---

## English Summary

- **Local Only**: DevLemon operates 100% locally on your Mac. No project files, directory paths, or scan reports are ever transmitted over the network.
- **No Tracking**: No telemetry, analytics, or advertising SDKs are included.
- **Local Access**: Reads caches, build artifacts, running process info and installed app metadata on your Mac; nothing is deleted without your confirmation.
- **Admin Rights**: Only the optional memory "Release" action runs `purge` via the macOS authorization dialog; DevLemon never sees your password.
- **Updates**: Sparkle fetches `appcast.xml` from GitHub without any user identifier.
