import SwiftUI

// MARK: - 1. 状态栏与偏好设置专用：极简单色矢量柠檬图标 (Monochrome Lemon Icon)
public struct MonochromeLemonIcon: View {
    public var size: CGFloat = 14

    public init(size: CGFloat = 14) {
        self.size = size
    }

    public var body: some View {
        Canvas { context, canvasSize in
            let w = canvasSize.width
            let h = canvasSize.height

            // 绘制倾斜的柠檬轮廓
            var lemonPath = Path()
            // 稍稍倾斜 -15 度
            let cx = w * 0.48
            let cy = h * 0.54
            let rx = w * 0.38
            let ry = h * 0.32

            // 绘制带有两端微尖特征的柠檬椭圆轮廓
            lemonPath.move(to: CGPoint(x: cx - rx * 1.15, y: cy))
            // 左上到右尖
            lemonPath.addCurve(
                to: CGPoint(x: cx + rx * 1.15, y: cy),
                control1: CGPoint(x: cx - rx * 0.6, y: cy - ry * 1.25),
                control2: CGPoint(x: cx + rx * 0.6, y: cy - ry * 1.25)
            )
            // 右尖到左尖
            lemonPath.addCurve(
                to: CGPoint(x: cx - rx * 1.15, y: cy),
                control1: CGPoint(x: cx + rx * 0.6, y: cy + ry * 1.25),
                control2: CGPoint(x: cx - rx * 0.6, y: cy + ry * 1.25)
            )
            lemonPath.closeSubpath()

            // 绘制顶部嫩叶
            var leafPath = Path()
            let leafBase = CGPoint(x: cx + rx * 0.2, y: cy - ry * 0.9)
            let leafTip = CGPoint(x: cx + rx * 0.8, y: cy - ry * 1.7)
            leafPath.move(to: leafBase)
            leafPath.addQuadCurve(to: leafTip, control: CGPoint(x: cx + rx * 0.1, y: cy - ry * 1.6))
            leafPath.addQuadCurve(to: leafBase, control: CGPoint(x: cx + rx * 0.75, y: cy - ry * 1.0))
            leafPath.closeSubpath()

            // 渲染单色轮廓与填充
            context.fill(leafPath, with: .color(.primary.opacity(0.9)))
            context.stroke(lemonPath, with: .color(.primary), style: StrokeStyle(lineWidth: w * 0.10, lineCap: .round, lineJoin: .round))
        }
        .frame(width: size, height: size)
    }
}

// MARK: - 2. 首页中央发光大图标：高保真拟物立体渐变柠檬 (Hero Lemon Icon)
public struct HeroLemonIcon: View {
    public var size: CGFloat = 110
    @State private var breathing: Bool = false

    public init(size: CGFloat = 110) {
        self.size = size
    }

    public var body: some View {
        ZStack {
            // 1. 弥散呼吸光晕
            Circle()
                .fill(
                    RadialGradient(
                        gradient: Gradient(colors: [
                            Color(red: 1.0, green: 0.85, blue: 0.2).opacity(0.28),
                            Color(red: 1.0, green: 0.75, blue: 0.1).opacity(0.10),
                            Color.clear
                        ]),
                        center: .center,
                        startRadius: size * 0.2,
                        endRadius: size * 0.75
                    )
                )
                .frame(width: size * 1.6, height: size * 1.6)
                .scaleEffect(breathing ? 1.08 : 0.95)
                .animation(Animation.easeInOut(duration: 2.8).repeatForever(autoreverses: true), value: breathing)

            // 2. 柠檬主体与光影
            ZStack {
                // 柠檬本体图形
                LemonShape()
                    .fill(
                        LinearGradient(
                            colors: [
                                Color(red: 1.0, green: 0.93, blue: 0.38), // 亮柠檬黄
                                Color(red: 0.98, green: 0.82, blue: 0.18), // 中调柠黄
                                Color(red: 0.88, green: 0.65, blue: 0.10)  // 金黄底调阴影
                            ],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        )
                    )
                    .frame(width: size * 0.85, height: size * 0.72)
                    .rotationEffect(.degrees(-20))
                    .shadow(color: Color.black.opacity(0.25), radius: 10, x: 2, y: 6)

                // 柠檬高光斑 (左上方立体光泽)
                Ellipse()
                    .fill(
                        LinearGradient(
                            colors: [Color.white.opacity(0.55), Color.white.opacity(0.0)],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        )
                    )
                    .frame(width: size * 0.42, height: size * 0.24)
                    .rotationEffect(.degrees(-28))
                    .offset(x: -size * 0.10, y: -size * 0.10)

                // 3. 顶部青绿立体嫩叶 (叶柄 + 叶片)
                ZStack {
                    // 叶片
                    LeafShape()
                        .fill(
                            LinearGradient(
                                colors: [
                                    Color(red: 0.35, green: 0.82, blue: 0.38),
                                    Color(red: 0.18, green: 0.65, blue: 0.24)
                                ],
                                startPoint: .topLeading,
                                endPoint: .bottomTrailing
                            )
                        )
                        .frame(width: size * 0.38, height: size * 0.28)
                        .rotationEffect(.degrees(-15))
                        .offset(x: -size * 0.12, y: -size * 0.34)
                        .shadow(color: Color.black.opacity(0.18), radius: 3, x: 1, y: 2)

                    // 叶片高光叶脉
                    Path { path in
                        path.move(to: CGPoint(x: size * 0.05, y: size * 0.15))
                        path.addQuadCurve(to: CGPoint(x: size * 0.32, y: size * 0.02), control: CGPoint(x: size * 0.18, y: size * 0.06))
                    }
                    .stroke(Color.white.opacity(0.35), lineWidth: 1.5)
                    .offset(x: -size * 0.14, y: -size * 0.36)
                }
            }
        }
        .frame(width: size, height: size)
        .onAppear {
            breathing = true
        }
    }
}

// 拟物柠檬轮廓 Shape
fileprivate struct LemonShape: Shape {
    func path(in rect: CGRect) -> Path {
        var path = Path()
        let w = rect.width
        let h = rect.height
        let cx = w / 2
        let cy = h / 2

        // 左尖端
        path.move(to: CGPoint(x: 0, y: cy))

        // 上半圆弧
        path.addCurve(
            to: CGPoint(x: w, y: cy),
            control1: CGPoint(x: cx * 0.35, y: -h * 0.08),
            control2: CGPoint(x: cx * 1.65, y: -h * 0.08)
        )

        // 下半圆弧
        path.addCurve(
            to: CGPoint(x: 0, y: cy),
            control1: CGPoint(x: cx * 1.65, y: h * 1.08),
            control2: CGPoint(x: cx * 0.35, y: h * 1.08)
        )

        path.closeSubpath()
        return path
    }
}

// 嫩叶 Shape
fileprivate struct LeafShape: Shape {
    func path(in rect: CGRect) -> Path {
        var path = Path()
        let w = rect.width
        let h = rect.height

        path.move(to: CGPoint(x: 0, y: h * 0.8))
        path.addQuadCurve(to: CGPoint(x: w, y: 0), control: CGPoint(x: w * 0.2, y: -h * 0.1))
        path.addQuadCurve(to: CGPoint(x: 0, y: h * 0.8), control: CGPoint(x: w * 0.85, y: h * 0.9))
        path.closeSubpath()
        return path
    }
}

// MARK: - 3. 腾讯柠檬同款：上下双行极致紧凑微指标卡片
public struct CompactMetricBlock: View {
    public let val: String
    public let label: String

    public init(val: String, label: String) {
        self.val = val
        self.label = label
    }

    public var body: some View {
        VStack(alignment: .center, spacing: -2.5) {
            Text(val)
                .font(.system(size: 9, weight: .bold, design: .rounded))
                .lineLimit(1)
            Text(label)
                .font(.system(size: 6.5, weight: .bold, design: .rounded))
                .opacity(0.85)
                .lineLimit(1)
        }
        .frame(minWidth: 20)
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
        VStack(alignment: .leading, spacing: -2) {
            HStack(spacing: 1.5) {
                Image(systemName: "arrow.up")
                    .font(.system(size: 6, weight: .bold))
                Text(upSpeed)
                    .font(.system(size: 7.5, weight: .medium, design: .monospaced))
                    .lineLimit(1)
            }
            HStack(spacing: 1.5) {
                Image(systemName: "arrow.down")
                    .font(.system(size: 6, weight: .bold))
                Text(downSpeed)
                    .font(.system(size: 7.5, weight: .medium, design: .monospaced))
                    .lineLimit(1)
            }
        }
    }
}
