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
            }

            // 底部 4 个阶段雷达徽章
            HStack(spacing: 24) {
                PhaseBadge(icon: "shippingbox.fill", title: "容器环境", isActive: true)
                PhaseBadge(icon: "hammer.fill", title: "工程构建", isActive: true)
                PhaseBadge(icon: "iphone", title: "模拟器", isActive: true)
                PhaseBadge(icon: "trash.fill", title: "系统维护", isActive: true)
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
}

fileprivate struct PhaseBadge: View {
    let icon: String
    let title: String
    let isActive: Bool

    var body: some View {
        HStack(spacing: 6) {
            Image(systemName: icon)
                .font(.system(size: 11))
                .foregroundColor(isActive ? .yellow : .secondary)
            Text(title)
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(isActive ? .primary : .secondary)
        }
        .padding(.horizontal, 10)
        .padding(.vertical, 5)
        .background(
            Capsule()
                .fill(Color.white.opacity(0.04))
                .overlay(
                    Capsule()
                        .stroke(Color.white.opacity(0.08), lineWidth: 1)
                )
        )
    }
}

fileprivate struct HexagonShape: Shape {
    func path(in rect: CGRect) -> Path {
        var path = Path()
        let width = rect.width
        let height = rect.height
        let xCenter = width / 2
        let yCenter = height / 2
        let radius = min(width, height) / 2

        for i in 0..<6 {
            let angle = CGFloat(i) * (CGFloat.pi / 3.0)
            let x = xCenter + radius * cos(angle)
            let y = yCenter + radius * sin(angle)
            if i == 0 {
                path.move(to: CGPoint(x: x, y: y))
            } else {
                path.addLine(to: CGPoint(x: x, y: y))
            }
        }
        path.closeSubpath()
        return path
    }
}
