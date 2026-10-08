import SwiftUI

public struct CircularGaugeView: View {
    public let title: String
    public let percent: Double
    public let subtitle: String
    public let tintColor: Color

    public init(title: String, percent: Double, subtitle: String, tintColor: Color = .blue) {
        self.title = title
        self.percent = max(0, min(100, percent))
        self.subtitle = subtitle
        self.tintColor = tintColor
    }

    public var body: some View {
        VStack(spacing: 8) {
            // 环形进度
            ZStack {
                // 背景环
                Circle()
                    .stroke(Color.white.opacity(0.10), lineWidth: 5.5)
                    .frame(width: 54, height: 54)

                // 进度环
                Circle()
                    .trim(from: 0, to: CGFloat(percent / 100.0))
                    .stroke(
                        AngularGradient(
                            gradient: Gradient(colors: [
                                tintColor.opacity(0.5),
                                tintColor,
                                tintColor.opacity(0.9)
                            ]),
                            center: .center,
                            startAngle: .degrees(-90),
                            endAngle: .degrees(270)
                        ),
                        style: StrokeStyle(lineWidth: 5.5, lineCap: .round)
                    )
                    .rotationEffect(.degrees(-90))
                    .frame(width: 54, height: 54)
                    .shadow(color: tintColor.opacity(0.3), radius: 3)
                    .animation(.spring(response: 0.45, dampingFraction: 0.8), value: percent)

                // 中心数值
                Text("\(Int(percent))%")
                    .font(.system(size: 13, weight: .bold, design: .rounded))
                    .foregroundColor(.primary)
            }
            .frame(width: 56, height: 56)

            // 标题
            Text(title)
                .font(.system(size: 12, weight: .semibold))
                .foregroundColor(.primary)

            // 固定高度的副标题 (确保三列严格等高对齐)
            Text(subtitle)
                .font(.system(size: 10, weight: .regular, design: .rounded))
                .foregroundColor(.secondary)
                .lineLimit(1)
                .minimumScaleFactor(0.85)
                .frame(height: 14)
        }
        .padding(.vertical, 10)
        .padding(.horizontal, 4)
        .frame(maxWidth: .infinity)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color.white.opacity(0.04))
                .overlay(
                    RoundedRectangle(cornerRadius: 10)
                        .stroke(Color.white.opacity(0.06), lineWidth: 1)
                )
        )
    }
}
