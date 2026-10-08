import SwiftUI

public struct MainWindowView: View {
    @ObservedObject var state = AppState.shared
    @ObservedObject var monitor = SystemMonitor.shared

    public init() {}

    public var body: some View {
        ZStack {
            // 背景层：沉浸式深色磨砂玻璃背景
            LinearGradient(
                colors: [
                    Color(red: 0.10, green: 0.11, blue: 0.13),
                    Color(red: 0.07, green: 0.08, blue: 0.09)
                ],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
            .ignoresSafeArea()

            VStack(spacing: 0) {
                // 顶部极简拖动栏与红黄绿避让区
                HStack {
                    Spacer()
                    Text("DevLemon Lite")
                        .font(.system(size: 12, weight: .medium))
                        .foregroundColor(.secondary.opacity(0.8))
                    Spacer()
                }
                .frame(height: 32)
                .background(Color.black.opacity(0.15))

                // 内容切换区
                switch state.currentStage {
                case .idle:
                    idleView
                case .scanning:
                    ScanningView()
                case .results:
                    ScanResultView()
                case .cleaning:
                    cleaningView
                case .cleaned:
                    CleanSuccessView()
                }
            }

            // 错误浮层
            if let error = state.errorMessage {
                VStack {
                    HStack(spacing: 8) {
                        Image(systemName: "exclamationmark.triangle.fill")
                            .foregroundColor(.orange)
                        Text(error)
                            .font(.system(size: 12))
                            .foregroundColor(.primary)
                        Spacer()
                        Button {
                            state.errorMessage = nil
                        } label: {
                            Image(systemName: "xmark")
                                .font(.system(size: 10))
                                .foregroundColor(.secondary)
                        }
                        .buttonStyle(.plain)
                    }
                    .padding(12)
                    .background(
                        RoundedRectangle(cornerRadius: 8)
                            .fill(Color.red.opacity(0.18))
                            .overlay(
                                RoundedRectangle(cornerRadius: 8)
                                    .stroke(Color.red.opacity(0.3), lineWidth: 1)
                            )
                    )
                    .padding(16)
                    Spacer()
                }
                .transition(.move(edge: .top).combined(with: .opacity))
            }
        }
        .frame(minWidth: 680, minHeight: 460)
    }

    // MARK: - 就绪首页 (Idle)
    private var idleView: some View {
        VStack(spacing: 24) {
            Spacer()

            // 柠檬大标与微光
            ZStack {
                Circle()
                    .fill(Color.yellow.opacity(0.08))
                    .frame(width: 130, height: 130)

                Text("🍋")
                    .font(.system(size: 64))
                    .shadow(color: Color.yellow.opacity(0.4), radius: 12)
            }

            VStack(spacing: 8) {
                Text("DevLemon")
                    .font(.system(size: 26, weight: .bold, design: .rounded))
                    .foregroundColor(.primary)

                Text("专为开发者打造的轻量级智能磁盘清理工具")
                    .font(.system(size: 13))
                    .foregroundColor(.secondary)
            }

            // 硬件概况轻量卡片
            HStack(spacing: 16) {
                MetricChip(icon: "internaldrive", label: "系统主盘", val: "\(Int(monitor.diskUsagePercent))% 已用 (剩 \(monitor.diskFreeFormatted))")
                MetricChip(icon: "cpu", label: "CPU", val: "\(Int(monitor.cpuUsage))%")
                MetricChip(icon: "memorychip", label: "内存", val: "\(Int(monitor.memoryUsage))% (\(monitor.memoryUsedFormatted))")
            }
            .padding(.top, 4)

            // 开始扫描大按钮
            Button {
                state.startScan()
            } label: {
                HStack(spacing: 8) {
                    Image(systemName: "magnifyingglass")
                        .font(.system(size: 14, weight: .bold))
                    Text("开始深度扫描")
                        .font(.system(size: 14, weight: .bold))
                }
                .frame(width: 200)
                .padding(.vertical, 11)
                .background(
                    LinearGradient(
                        colors: [Color(red: 0.98, green: 0.88, blue: 0.25), Color(red: 0.90, green: 0.72, blue: 0.10)],
                        startPoint: .topLeading,
                        endPoint: .bottomTrailing
                    )
                )
                .foregroundColor(.black)
                .cornerRadius(10)
                .shadow(color: Color.yellow.opacity(0.35), radius: 8, y: 3)
            }
            .buttonStyle(.plain)
            .padding(.top, 12)

            Spacer()
        }
    }

    // MARK: - 清理中视图 (Cleaning)
    private var cleaningView: some View {
        VStack(spacing: 20) {
            Spacer()
            ProgressView()
                .scaleEffect(1.4)
                .progressViewStyle(CircularProgressViewStyle(tint: .yellow))

            Text("正在释放选中的项目与缓存...")
                .font(.system(size: 15, weight: .medium))
                .foregroundColor(.primary)

            Text("正在根据规则安全清理，请稍候")
                .font(.system(size: 12))
                .foregroundColor(.secondary)
            Spacer()
        }
    }
}

fileprivate struct MetricChip: View {
    let icon: String
    let label: String
    let val: String

    var body: some View {
        HStack(spacing: 6) {
            Image(systemName: icon)
                .font(.system(size: 11))
                .foregroundColor(.yellow)
            Text(label)
                .font(.system(size: 11))
                .foregroundColor(.secondary)
            Text(val)
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(.primary)
        }
        .padding(.horizontal, 10)
        .padding(.vertical, 6)
        .background(
            RoundedRectangle(cornerRadius: 6)
                .fill(Color.white.opacity(0.04))
                .overlay(
                    RoundedRectangle(cornerRadius: 6)
                        .stroke(Color.white.opacity(0.06), lineWidth: 1)
                )
        )
    }
}
