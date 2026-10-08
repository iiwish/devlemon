import SwiftUI

public struct MenuBarPopoverView: View {
    @ObservedObject var state = AppState.shared
    @ObservedObject var monitor = SystemMonitor.shared
    var onOpenMainWindow: () -> Void

    public init(onOpenMainWindow: @escaping () -> Void) {
        self.onOpenMainWindow = onOpenMainWindow
    }

    public var body: some View {
        VStack(spacing: 14) {
            // Header
            HStack(alignment: .center) {
                HStack(spacing: 6) {
                    Text("🍋")
                        .font(.system(size: 16))
                    Text("DevLemon Lite")
                        .font(.system(size: 13, weight: .bold))
                        .foregroundColor(.primary)
                }

                Spacer()

                HStack(spacing: 8) {
                    Button {
                        WindowManager.shared.showSettingsWindow()
                    } label: {
                        Image(systemName: "gearshape")
                            .font(.system(size: 12))
                            .foregroundColor(.secondary)
                    }
                    .buttonStyle(.plain)
                    .help("偏好设置")

                    Button {
                        NSApplication.shared.terminate(nil)
                    } label: {
                        Image(systemName: "power")
                            .font(.system(size: 12))
                            .foregroundColor(.secondary)
                    }
                    .buttonStyle(.plain)
                    .help("退出 DevLemon")
                }
            }

            // Status Banner Card
            HStack(spacing: 12) {
                ZStack {
                    Circle()
                        .fill(Color.green.opacity(0.15))
                        .frame(width: 36, height: 36)
                    Image(systemName: "checkmark.shield.fill")
                        .font(.system(size: 18))
                        .foregroundColor(.green)
                }

                VStack(alignment: .leading, spacing: 2) {
                    Text("开发者环境状态良好")
                        .font(.system(size: 13, weight: .semibold))
                    Text("实时监控系统硬件负载与网络波形")
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)
                }

                Spacer()
            }
            .padding(10)
            .background(
                RoundedRectangle(cornerRadius: 10)
                    .fill(Color.white.opacity(0.04))
            )

            // Hardware Gauges (三大等宽微卡片)
            HStack(spacing: 10) {
                CircularGaugeView(
                    title: "CPU",
                    percent: monitor.cpuUsage,
                    subtitle: monitor.cpuSubtitle,
                    tintColor: .blue
                )

                CircularGaugeView(
                    title: "内存",
                    percent: monitor.memoryUsage,
                    subtitle: monitor.memorySubtitle,
                    tintColor: .orange
                )

                CircularGaugeView(
                    title: "主盘",
                    percent: monitor.diskUsagePercent,
                    subtitle: monitor.diskSubtitle,
                    tintColor: .green
                )
            }
            .frame(maxWidth: .infinity)
            .padding(.vertical, 2)

            // Network Waveform
            NetworkSpeedGraph()

            // Action Buttons
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
                            colors: [Color(red: 0.95, green: 0.85, blue: 0.25), Color(red: 0.85, green: 0.70, blue: 0.10)],
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
    }
}
