import SwiftUI

public struct ScanningView: View {
    @ObservedObject var state = AppState.shared
    @State private var rotationDegree: Double = 0.0
    @State private var pulseScale: CGFloat = 1.0

    public init() {}

    public var body: some View {
        VStack(spacing: 32) {
            Spacer()

            // 核心旋转发光动画 (类似截图 2 的科技雷达感)
            ZStack {
                // 外层脉冲光环
                Circle()
                    .stroke(
                        RadialGradient(
                            gradient: Gradient(colors: [Color.yellow.opacity(0.35), Color.clear]),
                            center: .center,
                            startRadius: 50,
                            endRadius: 110
                        ),
                        lineWidth: 25
                    )
                    .frame(width: 220, height: 220)
                    .scaleEffect(pulseScale)
                    .animation(
                        Animation.easeInOut(duration: 1.6)
                            .repeatForever(autoreverses: true),
                        value: pulseScale
                    )

                // 旋转多边形/六边形光环
                HexagonShape()
                    .stroke(
                        AngularGradient(
                            gradient: Gradient(colors: [
                                Color.yellow.opacity(0.1),
                                Color.yellow.opacity(0.8),
                                Color.orange,
                                Color.yellow.opacity(0.1)
                            ]),
                            center: .center
                        ),
                        lineWidth: 3
                    )
                    .frame(width: 140, height: 140)
                    .rotationEffect(.degrees(rotationDegree))

                // 反向旋转虚线内环
                Circle()
                    .stroke(
                        Color.white.opacity(0.2),
                        style: StrokeStyle(lineWidth: 1.5, dash: [4, 6])
                    )
                    .frame(width: 90, height: 90)
                    .rotationEffect(.degrees(-rotationDegree * 0.7))

                // 中心发光微标 (高保真立体柠檬)
                HeroLemonIcon(size: 48)
            }
            .frame(height: 240)

            // 进度与状态文案
            VStack(spacing: 12) {
                Text(state.scanPhaseText)
                    .font(.system(size: 16, weight: .semibold))
                    .foregroundColor(.primary)

                // 动态扫描路径文本（中段省略效果）
                Text(state.scanningCurrentPath)
                    .font(.system(size: 12, design: .monospaced))
                    .foregroundColor(.secondary)
                    .lineLimit(1)
                    .frame(maxWidth: 420)
                    .padding(.horizontal, 16)
                    .padding(.vertical, 6)
                    .background(
                        RoundedRectangle(cornerRadius: 6)
                            .fill(Color.white.opacity(0.05))
                    )

                // 进度条
                ProgressView(value: state.scanProgress, total: 1.0)
                    .progressViewStyle(LinearProgressViewStyle(tint: Color.yellow))
                    .frame(width: 320)
                    .padding(.top, 6)

                // 停止扫描按钮 (优雅胶囊微晶造型)
                Button {
                    state.stopScan()
                } label: {
                    HStack(spacing: 5) {
                        Image(systemName: "stop.fill")
                            .font(.system(size: 9))
                        Text("停止扫描")
                            .font(.system(size: 11.5, weight: .medium))
                    }
                    .padding(.horizontal, 14)
                    .padding(.vertical, 5.5)
                    .background(
                        Capsule()
                            .fill(Color.white.opacity(0.08))
                            .overlay(
                                Capsule()
                                    .stroke(Color.white.opacity(0.14), lineWidth: 0.8)
                            )
                    )
                    .foregroundColor(.secondary)
                }
                .buttonStyle(.plain)
                .help("中止当前扫描并返回首页")
                .padding(.top, 4)
            }

            // 底部 4 个阶段雷达徽章 (动态点亮与脉冲)
            HStack(spacing: 18) {
                PhaseBadge(
                    icon: "trash.fill",
                    title: "系统维护",
                    state: phaseState(forIndex: 0)
                )
                PhaseBadge(
                    icon: "square.grid.2x2.fill",
                    title: "应用垃圾",
                    state: phaseState(forIndex: 1)
                )
                PhaseBadge(
                    icon: "shippingbox.fill",
                    title: "容器环境",
                    state: phaseState(forIndex: 2)
                )
                PhaseBadge(
                    icon: "hammer.fill",
                    title: "工程构建",
                    state: phaseState(forIndex: 3)
                )
            }
            .padding(.top, 8)

            Spacer()
        }
        .onAppear {
            withAnimation(Animation.linear(duration: 4.0).repeatForever(autoreverses: false)) {
                rotationDegree = 360.0
            }
            pulseScale = 1.15
        }
    }

    private func phaseState(forIndex index: Int) -> PhaseState {
        let p = state.scanProgress
        // 4 个阶段对应的进度区间
        let thresholds: [Double] = [0.25, 0.50, 0.75, 1.0]
        let currentTarget = thresholds[index]
        let previousTarget = index == 0 ? 0.0 : thresholds[index - 1]

        if p >= currentTarget {
            return .completed
        } else if p >= previousTarget {
            return .active
        } else {
            return .pending
        }
    }
}

fileprivate enum PhaseState {
    case pending
    case active
    case completed
}

fileprivate struct PhaseBadge: View {
    let icon: String
    let title: String
    let state: PhaseState

    var body: some View {
        HStack(spacing: 6) {
            badgeIcon

            Text(title)
                .font(.system(size: 11, weight: state == .active ? .semibold : .medium))
                .foregroundColor(textColor)
        }
        .padding(.horizontal, 10)
        .padding(.vertical, 5)
        .background(
            Capsule()
                .fill(state == .active ? Color.yellow.opacity(0.12) : Color.white.opacity(0.04))
                .overlay(
                    Capsule()
                        .stroke(
                            state == .active ? Color.yellow.opacity(0.35) : Color.white.opacity(0.06),
                            lineWidth: 1
                        )
                )
        )
    }

    @ViewBuilder
    private var badgeIcon: some View {
        switch state {
        case .completed:
            Image(systemName: "checkmark.circle.fill")
                .font(.system(size: 11))
                .foregroundColor(.green)
        case .active:
            Image(systemName: icon)
                .font(.system(size: 11))
                .foregroundColor(.yellow)
        case .pending:
            Image(systemName: icon)
                .font(.system(size: 11))
                .foregroundColor(.secondary.opacity(0.4))
        }
    }

    private var textColor: Color {
        switch state {
        case .active: return .primary
        case .completed: return .primary.opacity(0.85)
        case .pending: return .secondary.opacity(0.5)
        }
    }
}
