import SwiftUI

// MARK: - 1. 状态栏单色矢量柠檬图标 (精细匹配 Apple 🍋 剪影轮廓)
public struct MonochromeLemonIcon: View {
    public var size: CGFloat = 14

    public init(size: CGFloat = 14) {
        self.size = size
    }

    public var body: some View {
        Canvas { context, canvasSize in
            let w = canvasSize.width
            let h = canvasSize.height

            // 果蒂基部与右下尖端坐标 (与 Apple 🍋 倾角完全一致)
            let stem = CGPoint(x: w * 0.38, y: h * 0.30)
            let tip = CGPoint(x: w * 0.78, y: h * 0.78)

            // 1. 柠檬本体
            var bodyPath = Path()
            bodyPath.move(to: stem)
            bodyPath.addCurve(
                to: tip,
                control1: CGPoint(x: w * 0.92, y: h * 0.28),
                control2: CGPoint(x: w * 0.95, y: h * 0.65)
            )
            bodyPath.addCurve(
                to: stem,
                control1: CGPoint(x: w * 0.56, y: h * 0.96),
                control2: CGPoint(x: w * 0.18, y: h * 0.62)
            )
            bodyPath.closeSubpath()

            // 2. 左侧大绿叶 (向左上方舒展)
            var bigLeaf = Path()
            bigLeaf.move(to: stem)
            bigLeaf.addQuadCurve(
                to: CGPoint(x: w * 0.08, y: h * 0.14),
                control: CGPoint(x: w * 0.08, y: h * 0.36)
            )
            bigLeaf.addQuadCurve(
                to: stem,
                control: CGPoint(x: w * 0.32, y: h * 0.12)
            )
            bigLeaf.closeSubpath()

            // 3. 右侧小绿叶 (向上方小巧挺立)
            var smallLeaf = Path()
            smallLeaf.move(to: stem)
            smallLeaf.addQuadCurve(
                to: CGPoint(x: w * 0.56, y: h * 0.12),
                control: CGPoint(x: w * 0.42, y: h * 0.14)
            )
            smallLeaf.addQuadCurve(
                to: stem,
                control: CGPoint(x: w * 0.60, y: h * 0.24)
            )
            smallLeaf.closeSubpath()

            // 纯色模板填充
            context.fill(bodyPath, with: .color(.primary))
            context.fill(bigLeaf, with: .color(.primary))
            context.fill(smallLeaf, with: .color(.primary))
        }
        .frame(width: size, height: size)
    }
}

// MARK: - 2. 经典 🍋 风格高保真主图标 (Hero Lemon Icon)
public struct HeroLemonIcon: View {
    public var size: CGFloat = 110
    @State private var pulse: Bool = false

    public init(size: CGFloat = 110) {
        self.size = size
    }

    public var body: some View {
        ZStack {
            // 1. 金黄弥散漫反射微光 (极具质感的大范围柔和环境光)
            Circle()
                .fill(
                    RadialGradient(
                        gradient: Gradient(colors: [
                            Color(red: 1.0, green: 0.85, blue: 0.20).opacity(0.35),
                            Color(red: 0.98, green: 0.70, blue: 0.10).opacity(0.12),
                            Color.clear
                        ]),
                        center: .center,
                        startRadius: size * 0.15,
                        endRadius: size * 0.70
                    )
                )
                .frame(width: size * 1.5, height: size * 1.5)
                .scaleEffect(pulse ? 1.05 : 0.95)
                .animation(Animation.easeInOut(duration: 2.4).repeatForever(autoreverses: true), value: pulse)

            // 2. 原生 Apple 🍋 Emoji 经典 3D 渲染 (最高画质原生矢量字体输出)
            Text("🍋")
                .font(.custom("Apple Color Emoji", size: size * 0.78))
                .shadow(color: Color.black.opacity(0.32), radius: size * 0.10, x: 0, y: size * 0.05)
                .scaleEffect(pulse ? 1.02 : 0.98)
                .animation(Animation.easeInOut(duration: 2.4).repeatForever(autoreverses: true), value: pulse)
        }
        .frame(width: size, height: size)
        .onAppear {
            pulse = true
        }
    }
}

// MARK: - 3. 官方 macOS 拟态 AppIcon 卡片组件 (可用于关于页或主面板展示)
public struct AppIconSquircleView: View {
    public var size: CGFloat = 120

    public init(size: CGFloat = 120) {
        self.size = size
    }

    public var body: some View {
        let tileSide = size * 0.80
        let cornerRadius = tileSide * (185.0 / 824.0)

        ZStack {
            // 底部柔和投影
            RoundedRectangle(cornerRadius: cornerRadius, style: .continuous)
                .fill(Color.black.opacity(0.35))
                .frame(width: tileSide, height: tileSide)
                .blur(radius: tileSide * 0.08)
                .offset(y: tileSide * 0.04)

            // 实体卡片
            ZStack {
                // 深黑石墨背景
                LinearGradient(
                    colors: [
                        Color(red: 0.17, green: 0.18, blue: 0.21),
                        Color(red: 0.09, green: 0.10, blue: 0.12)
                    ],
                    startPoint: .top,
                    endPoint: .bottom
                )

                // 弥散金黄柠檬光晕
                RadialGradient(
                    gradient: Gradient(colors: [
                        Color(red: 1.0, green: 0.82, blue: 0.18).opacity(0.38),
                        Color(red: 0.98, green: 0.65, blue: 0.08).opacity(0.14),
                        Color.clear
                    ]),
                    center: UnitPoint(x: 0.52, y: 0.52),
                    startRadius: tileSide * 0.08,
                    endRadius: tileSide * 0.50
                )

                // 边缘高光微轮廓
                RoundedRectangle(cornerRadius: cornerRadius, style: .continuous)
                    .stroke(
                        LinearGradient(
                            stops: [
                                .init(color: Color.white.opacity(0.24), location: 0.0),
                                .init(color: Color.white.opacity(0.08), location: 0.4),
                                .init(color: Color.white.opacity(0.02), location: 1.0)
                            ],
                            startPoint: .top,
                            endPoint: .bottom
                        ),
                        lineWidth: 1.5
                    )

                // 经典 Apple 柠檬
                Text("🍋")
                    .font(.custom("Apple Color Emoji", size: tileSide * 0.67))
                    .shadow(color: Color.black.opacity(0.50), radius: tileSide * 0.04, x: 0, y: tileSide * 0.02)
                    .offset(x: tileSide * 0.015, y: tileSide * 0.01)
            }
            .clipShape(RoundedRectangle(cornerRadius: cornerRadius, style: .continuous))
            .frame(width: tileSide, height: tileSide)
        }
        .frame(width: size, height: size)
    }
}

// MARK: - 4. 腾讯柠檬同款：上下双行紧凑微指标卡片 (字号与字重完全对齐柠檬清理规范)
public struct CompactMetricBlock: View {
    public let val: String
    public let label: String

    public init(val: String, label: String) {
        self.val = val
        self.label = label
    }

    public var body: some View {
        VStack(alignment: .center, spacing: -1.5) {
            Text(val)
                .font(.system(size: 11, weight: .bold))
                .lineLimit(1)
                .fixedSize()
            Text(label)
                .font(.system(size: 7.5, weight: .bold))
                .opacity(0.85)
                .lineLimit(1)
                .fixedSize()
        }
    }
}

// 紧凑双行网速
public struct CompactNetBlock: View {
    public let upSpeed: String
    public let downSpeed: String

    public init(upSpeed: String, downSpeed: String) {
        self.upSpeed = upSpeed
        self.downSpeed = downSpeed
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: -1.5) {
            HStack(spacing: 2) {
                Image(systemName: "arrow.up")
                    .font(.system(size: 7, weight: .bold))
                Text(upSpeed)
                    .font(.system(size: 8, weight: .semibold, design: .monospaced))
                    .lineLimit(1)
                    .fixedSize()
            }
            HStack(spacing: 2) {
                Image(systemName: "arrow.down")
                    .font(.system(size: 7, weight: .bold))
                Text(downSpeed)
                    .font(.system(size: 8, weight: .semibold, design: .monospaced))
                    .lineLimit(1)
                    .fixedSize()
            }
        }
    }
}
