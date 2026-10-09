import SwiftUI

public struct MainWindowView: View {
    @ObservedObject var state = AppState.shared
    @ObservedObject var monitor = SystemMonitor.shared

    @State private var isBackHovered: Bool = false
    @State private var isSettingsHovered: Bool = false

    public init() {}

    public var body: some View {
        ZStack {
            // 背景层：沉浸式深色微渐变背景，全屏铺满到顶
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
                // 与系统红黄绿交通灯处于同一水平行的单行标题栏 (彻底消除多余空行)
                topInlineBar

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
            .ignoresSafeArea(edges: .top)

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
                    .padding(10)
                    .background(
                        RoundedRectangle(cornerRadius: 8)
                            .fill(Color.red.opacity(0.18))
                            .overlay(
                                RoundedRectangle(cornerRadius: 8)
                                    .stroke(Color.red.opacity(0.3), lineWidth: 1)
                            )
                    )
                    .padding(14)
                    Spacer()
                }
                .transition(.move(edge: .top).combined(with: .opacity))
            }
        }
        .frame(minWidth: 720, minHeight: 480)
    }

    // MARK: - 与系统红黄绿处于同一行的统一标题栏
    private var topInlineBar: some View {
        HStack(spacing: 8) {
            // 左侧：为系统红黄绿交通灯预留位置 (宽 72pt，完全避让，同行排列)
            Color.clear
                .frame(width: 72, height: 1)

            // 返回按钮 (在非首页且非扫描状态下显示，优雅 macOS 原生胶囊按钮)
            if state.currentStage != .idle && state.currentStage != .scanning && state.currentStage != .cleaning {
                Button {
                    withAnimation(.spring(response: 0.35, dampingFraction: 0.8)) {
                        state.resetToIdle()
                    }
                } label: {
                    HStack(spacing: 4) {
                        Image(systemName: "chevron.backward")
                            .font(.system(size: 10.5, weight: .semibold))
                        Text("首页")
                            .font(.system(size: 11, weight: .medium))
                    }
                    .foregroundColor(isBackHovered ? .primary : .secondary)
                    .padding(.horizontal, 9)
                    .padding(.vertical, 4.5)
                    .background(
                        Capsule()
                            .fill(isBackHovered ? Color.white.opacity(0.12) : Color.white.opacity(0.05))
                            .overlay(
                                Capsule()
                                    .stroke(Color.white.opacity(isBackHovered ? 0.18 : 0.08), lineWidth: 0.8)
                            )
                    )
                }
                .buttonStyle(.plain)
                .onHover { isBackHovered = $0 }
                .help("返回 DevLemon 首页")
                .transition(.opacity.combined(with: .move(edge: .leading)))
            }

            Spacer()

            // 居中单行标题与状态面包屑
            HStack(spacing: 5) {
                Text("DevLemon")
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundColor(.secondary.opacity(0.9))

                if state.currentStage == .results {
                    Text("›")
                        .font(.system(size: 11))
                        .foregroundColor(.secondary.opacity(0.5))
                    Text("扫描结果")
                        .font(.system(size: 11.5, weight: .medium))
                        .foregroundColor(.primary.opacity(0.85))
                } else if state.currentStage == .scanning {
                    Text("›")
                        .font(.system(size: 11))
                        .foregroundColor(.secondary.opacity(0.5))
                    Text("雷达扫描")
                        .font(.system(size: 11.5, weight: .medium))
                        .foregroundColor(.yellow.opacity(0.9))
                } else if state.currentStage == .cleaned {
                    Text("›")
                        .font(.system(size: 11))
                        .foregroundColor(.secondary.opacity(0.5))
                    Text("清理完成")
                        .font(.system(size: 11.5, weight: .medium))
                        .foregroundColor(.green.opacity(0.9))
                }
            }

            Spacer()

            // 右侧偏好设置小图标
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
            .help("打开偏好设置")
            .padding(.trailing, 12)
        }
        .frame(height: 34)
        .padding(.top, 4)
    }

    // MARK: - 就绪首页 (Idle)
    private var idleView: some View {
        VStack(spacing: 22) {
            Spacer()

            // 全新高保真立体渐变柠檬 Hero 图标
            HeroLemonIcon(size: 105)

            VStack(spacing: 6) {
                Text("DevLemon")
                    .font(.system(size: 26, weight: .bold, design: .rounded))
                    .foregroundColor(.primary)

                Text("专为开发者打造的轻量级智能磁盘清理工具")
                    .font(.system(size: 13))
                    .foregroundColor(.secondary)
            }

            // 硬件概况微卡片
            HStack(spacing: 14) {
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
                .frame(width: 190)
                .padding(.vertical, 10)
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
            .padding(.top, 10)

            Spacer()
        }
    }

    // MARK: - 清理中视图 (Cleaning)
    private var cleaningView: some View {
        VStack(spacing: 24) {
            Spacer()

            ZStack {
                Circle()
                    .stroke(Color.yellow.opacity(0.2), lineWidth: 4)
                    .frame(width: 72, height: 72)

                ProgressView()
                    .scaleEffect(1.6)
                    .progressViewStyle(CircularProgressViewStyle(tint: .yellow))

                HeroLemonIcon(size: 26)
            }

            VStack(spacing: 8) {
                Text("正在释放选中的项目与缓存...")
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundColor(.primary)

                Text("已开启运行态保护机制，安全清理中，请稍候")
                    .font(.system(size: 12))
                    .foregroundColor(.secondary)
            }

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
        .padding(.vertical, 5)
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
