import SwiftUI
import AppKit

public struct MenuBarPopoverView: View {
    @ObservedObject var state = AppState.shared
    @ObservedObject var monitor = SystemMonitor.shared
    var onOpenMainWindow: () -> Void

    @State private var isSettingsHovered = false
    @State private var isPowerHovered = false

    public init(onOpenMainWindow: @escaping () -> Void) {
        self.onOpenMainWindow = onOpenMainWindow
    }

    public var body: some View {
        VStack(spacing: 14) {
            // MARK: - 1. 顶栏标题与快捷设置/退出按钮
            HStack(alignment: .center) {
                HStack(spacing: 6) {
                    HeroLemonIcon(size: 18)
                    Text("DevLemon")
                        .font(.system(size: 13, weight: .bold))
                        .foregroundColor(.primary)
                }

                Spacer()

                HStack(spacing: 6) {
                    Button {
                        SettingsWindowController.shared.show()
                    } label: {
                        Image(systemName: "gearshape")
                            .font(.system(size: 12.5))
                            .foregroundColor(isSettingsHovered ? .primary : .secondary)
                            .frame(width: 24, height: 24)
                            .background(
                                Circle()
                                    .fill(isSettingsHovered ? Color.white.opacity(0.12) : Color.clear)
                            )
                    }
                    .buttonStyle(.plain)
                    .onHover { isSettingsHovered = $0 }
                    .help("偏好设置")

                    Button {
                        NSApplication.shared.terminate(nil)
                    } label: {
                        Image(systemName: "power")
                            .font(.system(size: 11.5))
                            .foregroundColor(isPowerHovered ? .red : .secondary)
                            .frame(width: 24, height: 24)
                            .background(
                                Circle()
                                    .fill(isPowerHovered ? Color.red.opacity(0.16) : Color.clear)
                            )
                    }
                    .buttonStyle(.plain)
                    .onHover { isPowerHovered = $0 }
                    .help("退出 DevLemon")
                }
            }

            // MARK: - 2. 核心快速清理看板 (仿腾讯柠檬清理设计)
            quickCleanCard

            // MARK: - 3. 三大硬件指标微卡片 (点击可联动系统活动监视器)
            HStack(spacing: 10) {
                CircularGaugeView(
                    title: "CPU",
                    percent: monitor.cpuUsage,
                    subtitle: monitor.cpuSubtitle,
                    tintColor: .blue,
                    onTap: {
                        openActivityMonitor()
                    }
                )
                .help("点击打开系统活动监视器")

                CircularGaugeView(
                    title: "内存",
                    percent: monitor.memoryUsage,
                    subtitle: monitor.memorySubtitle,
                    tintColor: .orange,
                    onTap: {
                        openActivityMonitor()
                    }
                )
                .help("点击打开系统活动监视器")

                CircularGaugeView(
                    title: "主盘",
                    percent: monitor.diskUsagePercent,
                    subtitle: monitor.diskSubtitle,
                    tintColor: .green,
                    onTap: {
                        onOpenMainWindow()
                    }
                )
                .help("点击进入深度清理")
            }
            .frame(maxWidth: .infinity)
            .padding(.vertical, 2)

            // MARK: - 4. 实时网络波形图
            NetworkSpeedGraph()

            // MARK: - 5. 底部主操作区 (深度清理 & 打开主面板)
            HStack(spacing: 10) {
                Button {
                    onOpenMainWindow()
                    state.startScan()
                } label: {
                    HStack(spacing: 6) {
                        Image(systemName: "sparkles")
                        Text("深度清理")
                            .font(.system(size: 13, weight: .bold))
                    }
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 8)
                    .background(
                        LinearGradient(
                            colors: [Color(red: 0.98, green: 0.88, blue: 0.25), Color(red: 0.90, green: 0.72, blue: 0.10)],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        )
                    )
                    .foregroundColor(.black)
                    .cornerRadius(8)
                    .shadow(color: Color.yellow.opacity(0.3), radius: 4, y: 2)
                }
                .buttonStyle(.plain)

                Button {
                    onOpenMainWindow()
                } label: {
                    Text("打开主面板")
                        .font(.system(size: 13, weight: .medium))
                        .padding(.horizontal, 12)
                        .padding(.vertical, 8)
                        .background(Color.white.opacity(0.08))
                        .foregroundColor(.primary)
                        .cornerRadius(8)
                }
                .buttonStyle(.plain)
            }
        }
        .padding(16)
        .frame(width: 320)
        .onAppear {
            state.refreshSafeReclaimableSize()
        }
    }

    // MARK: - 快速清理顶部卡片
    private var quickCleanCard: some View {
        HStack(spacing: 12) {
            // 左侧：科技柠檬微晶六边形
            HexagonLemonBadge(size: 42)

            // 中间：数值与状态描述
            VStack(alignment: .leading, spacing: 3) {
                switch state.quickCleanStatus {
                case .cleaning:
                    Text("清理中...")
                        .font(.system(size: 18, weight: .bold, design: .rounded))
                        .foregroundColor(.primary)
                    Text("正在安全释放缓存")
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)

                case .success(let freed):
                    Text(freed)
                        .font(.system(size: 21, weight: .bold, design: .rounded))
                        .foregroundColor(.green)
                    Text("已成功释放空间")
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)

                case .idle:
                    if state.safeReclaimableBytes > 0 {
                        Text(ByteFormatter.format(state.safeReclaimableBytes))
                            .font(.system(size: 22, weight: .bold, design: .rounded))
                            .foregroundColor(.primary)
                        Text("安全垃圾可清理")
                            .font(.system(size: 11))
                            .foregroundColor(.secondary)
                    } else if state.hasCheckedSafeClean {
                        Text("0 B")
                            .font(.system(size: 22, weight: .bold, design: .rounded))
                            .foregroundColor(.primary)
                        Text("系统保持最佳状态")
                            .font(.system(size: 11))
                            .foregroundColor(.secondary)
                    } else {
                        Text("检测中...")
                            .font(.system(size: 18, weight: .bold, design: .rounded))
                            .foregroundColor(.primary)
                        Text("正在快速排查缓存")
                            .font(.system(size: 11))
                            .foregroundColor(.secondary)
                    }
                }
            }

            Spacer()

            // 右侧：一键快速清理按钮
            quickCleanButton
        }
        .padding(12)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color.white.opacity(0.045))
                .overlay(
                    RoundedRectangle(cornerRadius: 10)
                        .stroke(Color.white.opacity(0.08), lineWidth: 1)
                )
        )
    }

    // MARK: - 快速清理按钮
    private var quickCleanButton: some View {
        SwiftUI.Group {
            switch state.quickCleanStatus {
            case .cleaning:
                ProgressView()
                    .scaleEffect(0.85)
                    .frame(width: 60, height: 28)
                    .background(Color.white.opacity(0.08))
                    .cornerRadius(6)

            case .success:
                HStack(spacing: 3) {
                    Image(systemName: "checkmark")
                        .font(.system(size: 11, weight: .bold))
                    Text("已完成")
                        .font(.system(size: 12, weight: .semibold))
                }
                .foregroundColor(.green)
                .padding(.horizontal, 10)
                .padding(.vertical, 5)
                .background(Color.green.opacity(0.12))
                .cornerRadius(6)

            case .idle:
                let canClean = state.safeReclaimableBytes > 0
                Button {
                    if canClean {
                        state.startQuickClean()
                    } else {
                        state.refreshSafeReclaimableSize()
                    }
                } label: {
                    Text(canClean ? "清理" : "很干净")
                        .font(.system(size: 13, weight: .bold))
                        .foregroundColor(canClean ? .white : .secondary)
                        .padding(.horizontal, 16)
                        .padding(.vertical, 6)
                        .background(
                            canClean
                                ? LinearGradient(
                                    colors: [
                                        Color(red: 0.18, green: 0.82, blue: 0.48),
                                        Color(red: 0.12, green: 0.70, blue: 0.38)
                                    ],
                                    startPoint: .top,
                                    endPoint: .bottom
                                )
                                : LinearGradient(
                                    colors: [Color.white.opacity(0.08)],
                                    startPoint: .top,
                                    endPoint: .bottom
                                )
                        )
                        .cornerRadius(6)
                        .shadow(color: canClean ? Color.green.opacity(0.3) : Color.clear, radius: 4, y: 1.5)
                }
                .buttonStyle(.plain)
                .disabled(!canClean && state.hasCheckedSafeClean)
                .help(canClean ? "一键无损释放全部安全缓存" : "系统当前暂无冗余安全垃圾")
            }
        }
    }

    private func openActivityMonitor() {
        if let url = NSWorkspace.shared.urlForApplication(withBundleIdentifier: "com.apple.ActivityMonitor") {
            NSWorkspace.shared.open(url)
        } else {
            let path = "/System/Applications/Utilities/Activity Monitor.app"
            NSWorkspace.shared.open(URL(fileURLWithPath: path))
        }
    }
}
