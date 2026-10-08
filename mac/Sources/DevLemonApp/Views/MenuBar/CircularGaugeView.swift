import SwiftUI

public struct CircularGaugeView: View {
    public let title: String
    public let percent: Double
    public let subtitle: String?
    public let tintColor: Color

    public init(title: String, percent: Double, subtitle: String? = nil, tintColor: Color = .blue) {
        self.title = title
        self.percent = max(0, min(100, percent))
        self.subtitle = subtitle
        self.tintColor = tintColor
    }

    public var body: some View {
        VStack(spacing: 6) {
            ZStack {
                // 背景环
                Circle()
                    .stroke(Color.white.opacity(0.12), lineWidth: 5)
                    .frame(width: 52, height: 52)

                // 进度环
                Circle()
                    .trim(from: 0, to: CGFloat(percent / 100.0))
                    .stroke(
                        AngularGradient(
                            gradient: Gradient(colors: [tintColor.opacity(0.6), tintColor]),
                            center: .center,
                            startAngle: .degrees(-90),
                            endAngle: .degrees(270)
                        ),
                        style: StrokeStyle(lineWidth: 5, lineCap: .round)
                    )
                    .rotationEffect(.degrees(-90))
                    .frame(width: 52, height: 52)
                    .animation(.spring(response: 0.4, dampingFraction: 0.8), value: percent)

                // 中心数值
                VStack(spacing: 0) {
                    Text("\(Int(percent))%")
                        .font(.system(size: 13, weight: .bold, design: .rounded))
                        .foregroundColor(.primary)
                }
            }

            Text(title)
                .font(.system(size: 11, weight: .medium))
                .foregroundColor(.secondary)

            if let sub = subtitle {
                Text(sub)
                    .font(.system(size: 9))
                    .foregroundColor(.secondary.opacity(0.8))
                    .lineLimit(1)
            }
        }
    }
}
