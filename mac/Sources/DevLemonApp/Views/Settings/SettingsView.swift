import SwiftUI
import ServiceManagement

public struct SettingsView: View {
    @ObservedObject var settings = AppSettings.shared
    @ObservedObject var monitor = SystemMonitor.shared
    @ObservedObject var cliInstaller = CLIInstaller.shared
    @ObservedObject var updateManager = UpdateManager.shared
    @ObservedObject var securityBookmark = SecurityBookmarkManager.shared

    @State private var cliReinstallNotice: String?
    @State private var isCopiedCLI: Bool = false

    public init() {}

    public var body: some View {
        ScrollView(.vertical, showsIndicators: true) {
            VStack(alignment: .leading, spacing: 18) {
                // 1. 状态栏展示信息设置 (多选卡片组与实时模拟预览)
                statusBarSection

                // 2. 开机自启
                launchSection

                // 3. 沙盒目录授权 (Mac App Store 沙盒规范)
                if securityBookmark.isSandboxed {
                    permissionSection
                }

                // 4. 终端命令行工具 (CLI)
                cliSection

                // 5. 软件自动更新 (Sparkle 2.0 / MAS)
                updateSection

                // 6. 关于与版本信息
                footerSection
            }
            .padding(20)
        }
        .frame(width: 560, height: 530)
        .background(
            LinearGradient(
                colors: [Color(red: 0.12, green: 0.13, blue: 0.15), Color(red: 0.08, green: 0.09, blue: 0.10)],
                startPoint: .top,
                endPoint: .bottom
            )
        )
    }

    // MARK: - 1. 状态栏展示设置
    private var statusBarSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("状态栏展示信息设置")
                .font(.system(size: 13.5, weight: .semibold))
                .foregroundColor(.primary)

            // 模拟 macOS 菜单栏预览条
            HStack(spacing: 6) {
                HStack(spacing: 6) {
                    if settings.showLogo || isAllDisabled {
                        MonochromeLemonIcon(size: 13)
                    }
                    if settings.showNetwork {
                        CompactNetBlock(upSpeed: "2.5K", downSpeed: "2.5K")
                    }
                    if settings.showMemory {
                        CompactMetricBlock(val: "45%", label: "MEM")
                    }
                    if settings.showDisk {
                        CompactMetricBlock(val: "28%", label: "SSD")
                    }
                    if settings.showCPU {
                        CompactMetricBlock(val: "29%", label: "CPU")
                    }
                }
                .padding(.horizontal, 6)

                Spacer()

                HStack(spacing: 8) {
                    Image(systemName: "wifi")
                        .font(.system(size: 11))
                    Text("拼")
                        .font(.system(size: 10, weight: .bold))
                        .padding(.horizontal, 3)
                        .padding(.vertical, 1)
                        .background(Color.white.opacity(0.15))
                        .cornerRadius(3)
                    Text("周六 下午9:41")
                        .font(.system(size: 10, weight: .medium))
                    Image(systemName: "magnifyingglass")
                        .font(.system(size: 11))
                    Image(systemName: "switch.2")
                        .font(.system(size: 11))
                }
                .foregroundColor(.secondary.opacity(0.8))
            }
            .padding(.horizontal, 12)
            .frame(height: 32)
            .background(
                RoundedRectangle(cornerRadius: 6)
                    .fill(Color.white.opacity(0.08))
                    .overlay(
                        RoundedRectangle(cornerRadius: 6)
                            .stroke(Color.white.opacity(0.12), lineWidth: 1)
                    )
            )

            // 5 个信息选择卡片
            HStack(spacing: 10) {
                SettingItemCard(
                    title: "Logo",
                    preview: AnyView(MonochromeLemonIcon(size: 22)),
                    isSelected: settings.showLogo,
                    onToggle: { settings.showLogo.toggle() }
                )

                SettingItemCard(
                    title: "网速",
                    preview: AnyView(
                        VStack(alignment: .leading, spacing: 1) {
                            Text("↑ 2.5K")
                                .font(.system(size: 8, weight: .semibold, design: .monospaced))
                            Text("↓ 2.5K")
                                .font(.system(size: 8, weight: .semibold, design: .monospaced))
                        }
                    ),
                    isSelected: settings.showNetwork,
                    onToggle: { settings.showNetwork.toggle() }
                )

                SettingItemCard(
                    title: "内存占用",
                    preview: AnyView(
                        VStack(spacing: 1) {
                            Text("45%")
                                .font(.system(size: 11, weight: .bold))
                            Text("MEM")
                                .font(.system(size: 7.5, weight: .bold))
                                .foregroundColor(.secondary)
                        }
                    ),
                    isSelected: settings.showMemory,
                    onToggle: { settings.showMemory.toggle() }
                )

                SettingItemCard(
                    title: "磁盘占用",
                    preview: AnyView(
                        VStack(spacing: 1) {
                            Text("28%")
                                .font(.system(size: 11, weight: .bold))
                            Text("SSD")
                                .font(.system(size: 7.5, weight: .bold))
                                .foregroundColor(.secondary)
                        }
                    ),
                    isSelected: settings.showDisk,
                    onToggle: { settings.showDisk.toggle() }
                )

                SettingItemCard(
                    title: "CPU 占用",
                    preview: AnyView(
                        VStack(spacing: 1) {
                            Text("29%")
                                .font(.system(size: 11, weight: .bold))
                            Text("CPU")
                                .font(.system(size: 7.5, weight: .bold))
                                .foregroundColor(.secondary)
                        }
                    ),
                    isSelected: settings.showCPU,
                    onToggle: { settings.showCPU.toggle() }
                )
            }
        }
        .padding(14)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color.white.opacity(0.04))
        )
    }

    // MARK: - 2. 开机自启
    private var launchSection: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack {
                Text("开机时启动状态栏")
                    .font(.system(size: 13.5, weight: .semibold))
                    .foregroundColor(.primary)

                Spacer()

                Toggle("", isOn: $settings.launchAtLogin)
                    .toggleStyle(SwitchToggleStyle(tint: .green))
                    .onChange(of: settings.launchAtLogin) { newValue in
                        updateLaunchAtLogin(newValue)
                    }
            }

            Text("登录系统时默认常驻菜单栏，实时监控硬件性能与释放内存。")
                .font(.system(size: 11))
                .foregroundColor(.secondary)
        }
        .padding(14)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color.white.opacity(0.04))
        )
    }

    // MARK: - 3. 沙盒目录授权 (Mac App Store 规范)
    private var permissionSection: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack {
                Text("系统目录访问授权")
                    .font(.system(size: 13.5, weight: .semibold))
                    .foregroundColor(.primary)

                Spacer()

                HStack(spacing: 5) {
                    Circle()
                        .fill(securityBookmark.hasAuthorization ? Color.green : Color.orange)
                        .frame(width: 7, height: 7)
                    Text(securityBookmark.hasAuthorization ? "已获得个人主目录授权" : "未授权")
                        .font(.system(size: 11, weight: .medium))
                        .foregroundColor(securityBookmark.hasAuthorization ? .green : .orange)
                }
                .padding(.horizontal, 8)
                .padding(.vertical, 3)
                .background(Color.white.opacity(0.06))
                .cornerRadius(12)
            }

            Text("Mac App Store 沙盒版本遵循苹果隐私合规规范。DevLemon 需要获得个人主目录授权，以扫描并释放开发缓存（如 Xcode DerivedData、node_modules、Docker 等）。")
                .font(.system(size: 11.5))
                .foregroundColor(.secondary)

            HStack {
                Text(securityBookmark.hasAuthorization ? securityBookmark.authorizedPath : "未授予主目录访问权限")
                    .font(.system(size: 11, design: .monospaced))
                    .foregroundColor(.secondary)
                    .lineLimit(1)
                    .truncationMode(.middle)

                Spacer()

                if securityBookmark.hasAuthorization {
                    Button("重置授权") {
                        securityBookmark.clearAuthorization()
                    }
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
                    .buttonStyle(.plain)
                    .padding(.horizontal, 8)
                    .padding(.vertical, 4)
                    .background(Color.white.opacity(0.08))
                    .cornerRadius(5)
                }

                Button {
                    securityBookmark.requestAuthorization { granted in
                        if granted {
                            cliInstaller.autoInstallIfNeeded()
                        }
                    }
                } label: {
                    HStack(spacing: 4) {
                        Image(systemName: "folder.badge.gearshape")
                            .font(.system(size: 11))
                        Text(securityBookmark.hasAuthorization ? "更改授权" : "立即授权个人主目录")
                            .font(.system(size: 11, weight: .medium))
                    }
                    .padding(.horizontal, 10)
                    .padding(.vertical, 4)
                    .background(Color.blue.opacity(0.8))
                    .foregroundColor(.white)
                    .cornerRadius(6)
                }
                .buttonStyle(.plain)
            }
            .padding(10)
            .background(Color.black.opacity(0.25))
            .cornerRadius(6)
        }
        .padding(14)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color.white.opacity(0.04))
        )
    }

    // MARK: - 4. 命令行终端工具 (CLI)
    private var cliSection: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack {
                Text("终端命令行工具 (CLI)")
                    .font(.system(size: 13.5, weight: .semibold))
                    .foregroundColor(.primary)

                Spacer()

                // 状态徽标
                HStack(spacing: 5) {
                    Circle()
                        .fill(cliInstaller.isInstalled ? Color.green : Color.orange)
                        .frame(width: 7, height: 7)
                    Text(cliInstaller.isInstalled ? "已安装 (\(cliInstaller.installedPath))" : "未检测到")
                        .font(.system(size: 11, weight: .medium))
                        .foregroundColor(cliInstaller.isInstalled ? .green : .orange)
                }
                .padding(.horizontal, 8)
                .padding(.vertical, 3)
                .background(Color.white.opacity(0.06))
                .cornerRadius(12)
            }

            Text("应用启动时已默认自动将核心引擎安装至系统终端 PATH，您可随时在 Terminal 中运行命令：")
                .font(.system(size: 11.5))
                .foregroundColor(.secondary)

            // 示例命令行代码块
            HStack {
                Text("$ dl scan")
                    .font(.system(size: 12, weight: .medium, design: .monospaced))
                    .foregroundColor(Color(red: 1.0, green: 0.85, blue: 0.3))
                Text("或")
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
                Text("$ devlemon clean")
                    .font(.system(size: 12, weight: .medium, design: .monospaced))
                    .foregroundColor(Color(red: 1.0, green: 0.85, blue: 0.3))
                Spacer()

                Button {
                    NSPasteboard.general.clearContents()
                    NSPasteboard.general.setString("dl scan", forType: .string)
                    isCopiedCLI = true
                    DispatchQueue.main.asyncAfter(deadline: .now() + 2.0) {
                        isCopiedCLI = false
                    }
                } label: {
                    HStack(spacing: 3) {
                        Image(systemName: isCopiedCLI ? "checkmark" : "doc.on.doc")
                            .font(.system(size: 10))
                        Text(isCopiedCLI ? "已复制" : "复制")
                            .font(.system(size: 10.5))
                    }
                    .foregroundColor(isCopiedCLI ? .green : .secondary)
                    .padding(.horizontal, 6)
                    .padding(.vertical, 3)
                    .background(Color.white.opacity(0.08))
                    .cornerRadius(4)
                }
                .buttonStyle(.plain)
                .help("复制终端命令至剪贴板")

                Button {
                    let result = cliInstaller.reinstall()
                    cliReinstallNotice = result.message
                } label: {
                    HStack(spacing: 4) {
                        Image(systemName: "arrow.triangle.2.circlepath")
                            .font(.system(size: 11))
                        Text("重新链接")
                            .font(.system(size: 11, weight: .medium))
                    }
                    .padding(.horizontal, 10)
                    .padding(.vertical, 4)
                    .background(Color.white.opacity(0.12))
                    .cornerRadius(6)
                }
                .buttonStyle(.plain)
            }
            .padding(10)
            .background(Color.black.opacity(0.25))
            .cornerRadius(6)

            if let notice = cliReinstallNotice {
                Text(notice)
                    .font(.system(size: 11))
                    .foregroundColor(.green)
            }
        }
        .padding(14)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color.white.opacity(0.04))
        )
    }

    // MARK: - 4. 软件自动更新 (Sparkle 2.0)
    private var updateSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("软件更新")
                    .font(.system(size: 13.5, weight: .semibold))
                    .foregroundColor(.primary)

                Spacer()

                if updateManager.isAppStoreBuild {
                    Text("由 Mac App Store 自动管理")
                        .font(.system(size: 11.5, weight: .medium))
                        .foregroundColor(.secondary)
                } else {
                    Button {
                        updateManager.checkForUpdates()
                    } label: {
                        HStack(spacing: 5) {
                            Image(systemName: "arrow.clockwise")
                                .font(.system(size: 11, weight: .bold))
                            Text("立即检查更新")
                                .font(.system(size: 12, weight: .medium))
                        }
                        .padding(.horizontal, 12)
                        .padding(.vertical, 5)
                        .background(
                            LinearGradient(
                                colors: [Color(red: 0.98, green: 0.88, blue: 0.25), Color(red: 0.90, green: 0.72, blue: 0.10)],
                                startPoint: .topLeading,
                                endPoint: .bottomTrailing
                            )
                        )
                        .foregroundColor(.black)
                        .cornerRadius(6)
                        .shadow(color: Color.yellow.opacity(0.25), radius: 3, y: 1)
                    }
                    .buttonStyle(.plain)
                }
            }

            if !updateManager.isAppStoreBuild {
                VStack(alignment: .leading, spacing: 8) {
                    Toggle(isOn: $updateManager.automaticallyChecksForUpdates) {
                        VStack(alignment: .leading, spacing: 2) {
                            Text("自动检查新版本")
                                .font(.system(size: 12, weight: .medium))
                                .foregroundColor(.primary)
                            Text("定期在后台检索最新版本并在发布时通知。")
                                .font(.system(size: 10.5))
                                .foregroundColor(.secondary)
                        }
                    }
                    .toggleStyle(.checkbox)

                    Toggle(isOn: $updateManager.automaticallyDownloadsUpdates) {
                        VStack(alignment: .leading, spacing: 2) {
                            Text("在后台自动下载新版本")
                                .font(.system(size: 12, weight: .medium))
                                .foregroundColor(.primary)
                            Text("检测到更新后自动下载并在下次启动时准备就绪。")
                                .font(.system(size: 10.5))
                                .foregroundColor(.secondary)
                        }
                    }
                    .toggleStyle(.checkbox)
                }
                .padding(.top, 2)
            } else {
                Text("此版本来自 Mac App Store，所有软件更新与补丁均由系统商店统一安全推送。")
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
            }
        }
        .padding(14)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color.white.opacity(0.04))
        )
    }

    // MARK: - 5. 底部版本信息
    private var footerSection: some View {
        HStack {
            Text("DevLemon v0.2.5")
                .font(.system(size: 12, weight: .semibold, design: .rounded))
                .foregroundColor(.secondary)
            Spacer()
            if updateManager.isAppStoreBuild {
                Text("Mac App Store 正式版")
                    .font(.system(size: 11))
                    .foregroundColor(.secondary.opacity(0.7))
            } else {
                Text("基于 Sparkle 2.0 安全签名更新机制")
                    .font(.system(size: 11))
                    .foregroundColor(.secondary.opacity(0.7))
            }
        }
        .padding(.horizontal, 4)
    }

    private var isAllDisabled: Bool {
        !settings.showLogo && !settings.showNetwork && !settings.showMemory && !settings.showDisk && !settings.showCPU
    }

    private func updateLaunchAtLogin(_ enable: Bool) {
        if #available(macOS 13.0, *) {
            do {
                if enable {
                    try SMAppService.mainApp.register()
                } else {
                    try SMAppService.mainApp.unregister()
                }
            } catch {
                print("Failed to toggle launch at login: \(error)")
            }
        }
    }
}

fileprivate struct SettingItemCard: View {
    let title: String
    let preview: AnyView
    let isSelected: Bool
    let onToggle: () -> Void

    var body: some View {
        Button(action: onToggle) {
            VStack(spacing: 8) {
                HStack {
                    Spacer()
                    Image(systemName: isSelected ? "checkmark.square.fill" : "square")
                        .font(.system(size: 12))
                        .foregroundColor(isSelected ? .blue : .secondary.opacity(0.6))
                }

                preview
                    .frame(height: 28)

                Text(title)
                    .font(.system(size: 11, weight: .medium))
                    .foregroundColor(isSelected ? .primary : .secondary)
                    .lineLimit(1)
            }
            .padding(8)
            .frame(maxWidth: .infinity)
            .frame(height: 86)
            .background(
                RoundedRectangle(cornerRadius: 8)
                    .fill(isSelected ? Color.blue.opacity(0.12) : Color.white.opacity(0.04))
                    .overlay(
                        RoundedRectangle(cornerRadius: 8)
                            .stroke(isSelected ? Color.blue.opacity(0.5) : Color.white.opacity(0.08), lineWidth: 1)
                    )
            )
        }
        .buttonStyle(.plain)
    }
}
