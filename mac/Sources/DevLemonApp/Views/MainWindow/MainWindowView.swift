import SwiftUI

public struct MainWindowView: View {
    @ObservedObject var state = AppState.shared
    @ObservedObject var monitor = SystemMonitor.shared

    public init() {}

    public var body: some View {
        ZStack {
            // 背景层：沉浸式深色磨砂渐变
            LinearGradient(
                colors: [
                    Color(red: 0.11, green: 0.12, blue: 0.14),
                    Color(red: 0.07, green: 0.08, blue: 0.09)
                ],
                startPoint: .topLeading,
                endPoint: .bottomTrailing
            )
            .ignoresSafeArea()

            VStack(spacing: 0) {
                // 顶部统一导航栏 (含返回按钮与红黄绿避让)
                topNavigationBar

                Divider()
                    .opacity(0.12)

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
        .frame(minWidth: 720, minHeight: 500)
    }

    // MARK: - 顶部导航栏
    private var topNavigationBar: some View {
        HStack(spacing: 12) {
            // 左侧：为系统红黄绿交通灯留出避让空间 (macOS 默认占 ~70pt)
            Color.clear
                .frame(width: 68, height: 1)

            // 返回按钮 (在非 idle 状态下显示)
            if state.currentStage != .idle {
                Button {
                    withAnimation(.easeInOut(duration: 0.25)) {
                        state.resetToIdle()
                    }
                } label: {
                    HStack(spacing: 4) {
                        Image(systemName: "chevron.left")
                            .font(.system(size: 11, weight: .bold))
                        Text("返回首页")
                            .font(.system(size: 12, weight: .medium))
                    }
                    .padding(.horizontal, 9)
                    .padding(.vertical, 4)
                    .background(Color.white.opacity(0.08))
                    .foregroundColor(.primary)
                    .cornerRadius(6)
                }
                .buttonStyle(.plain)
                .transition(.opacity.combined(with: .move(edge: .leading)))
            }

            Spacer()

            // 居中标题
            HStack(spacing: 6) {
                Text("DevLemon Lite")
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundColor(.primary.opacity(0.9))

                if state.currentStage == .results {
                    Text("· 深度扫描结果")
                        .font(.system(size: 12))
                        .foregroundColor(.secondary)
                }
            }

            Spacer()

            // 右侧偏好设置按钮
            Button {
                WindowManager.shared.showSettingsWindow()
            } label: {
                Image(systemName: "gearshape")
                    .font(.system(size: 12))
                    .foregroundColor(.secondary)
                    .padding(6)
                    .background(Color.white.opacity(0.05))
                    .clipShape(Circle())
            }
            .buttonStyle(.plain)
            .help("打开偏好设置")
            .padding(.trailing, 12)
        }
        .frame(height: 38)
        .background(Color.black.opacity(0.18))
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

            // 硬件概况微卡片
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
