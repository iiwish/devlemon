import SwiftUI
import ServiceManagement

public struct SettingsView: View {
    @ObservedObject var settings = AppSettings.shared
    @ObservedObject var monitor = SystemMonitor.shared

    public init() {}

    public var body: some View {
        VStack(alignment: .leading, spacing: 20) {
            // 1. 开机自启
            VStack(alignment: .leading, spacing: 6) {
                HStack {
                    Text("开机时启动状态栏")
                        .font(.system(size: 14, weight: .semibold))
                        .foregroundColor(.primary)

                    Spacer()

                    Toggle("", isOn: $settings.launchAtLogin)
                        .toggleStyle(SwitchToggleStyle(tint: .green))
                        .onChange(of: settings.launchAtLogin) { newValue in
                            updateLaunchAtLogin(newValue)
                        }
                }

                Text("开机时状态栏将默认常驻显示实时监控。")
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)
            }
            .padding(14)
            .background(
                RoundedRectangle(cornerRadius: 10)
                    .fill(Color.white.opacity(0.04))
            )

            // 2. 状态栏展示信息设置 (多选卡片组与实时模拟预览)
            VStack(alignment: .leading, spacing: 12) {
                Text("状态栏展示信息设置")
                    .font(.system(size: 14, weight: .semibold))
                    .foregroundColor(.primary)

                // 模拟 macOS 菜单栏预览条 (类似截图 1)
                HStack(spacing: 6) {
                    // 用户勾选项目的实时模拟渲染 (双行极致紧凑)
                    HStack(spacing: 5) {
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

                    // 系统固定图标模拟 (Wi-Fi, 输入法, 时间, 搜索, 控制中心)
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
                        title: "内存占用",
                        preview: AnyView(
                            VStack(spacing: 1) {
                                Text("45%")
                                    .font(.system(size: 11, weight: .bold, design: .monospaced))
                                Text("MEM")
                                    .font(.system(size: 8, weight: .bold))
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
                                    .font(.system(size: 11, weight: .bold, design: .monospaced))
                                Text("SSD")
                                    .font(.system(size: 8, weight: .bold))
                                    .foregroundColor(.secondary)
                            }
                        ),
                        isSelected: settings.showDisk,
                        onToggle: { settings.showDisk.toggle() }
                    )

                    SettingItemCard(
                        title: "网速",
                        preview: AnyView(
                            VStack(alignment: .leading, spacing: 1) {
                                Text("↑ 2.5 K/s")
                                    .font(.system(size: 8, design: .monospaced))
                                Text("↓ 2.5 K/s")
                                    .font(.system(size: 8, design: .monospaced))
                            }
                        ),
                        isSelected: settings.showNetwork,
                        onToggle: { settings.showNetwork.toggle() }
                    )

                    SettingItemCard(
                        title: "CPU 占用",
                        preview: AnyView(
                            VStack(spacing: 1) {
                                Text("29%")
                                    .font(.system(size: 11, weight: .bold, design: .monospaced))
                                Text("CPU")
                                    .font(.system(size: 8, weight: .bold))
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

            // 3. 关于 DevLemon Lite
            HStack {
                Text("DevLemon Lite v0.2.4")
                    .font(.system(size: 12, weight: .medium))
                    .foregroundColor(.secondary)
                Spacer()
                Text("专为开发者定制的智能清理与硬件监控")
                    .font(.system(size: 11))
                    .foregroundColor(.secondary.opacity(0.7))
            }
            .padding(.horizontal, 4)

            Spacer()
        }
        .padding(20)
        .frame(width: 540, height: 350)
        .background(
            LinearGradient(
                colors: [Color(red: 0.12, green: 0.13, blue: 0.15), Color(red: 0.08, green: 0.09, blue: 0.10)],
                startPoint: .top,
                endPoint: .bottom
            )
        )
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
                // 右上角复选框
                HStack {
                    Spacer()
                    Image(systemName: isSelected ? "checkmark.square.fill" : "square")
                        .font(.system(size: 12))
                        .foregroundColor(isSelected ? .blue : .secondary.opacity(0.6))
                }

                // 预览内容
                preview
                    .frame(height: 28)

                // 标题
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
