import SwiftUI

public struct NetworkSpeedGraph: View {
    @ObservedObject var monitor = SystemMonitor.shared

    public init() {}

    public var body: some View {
        VStack(spacing: 8) {
            HStack {
                HStack(spacing: 4) {
                    Image(systemName: "arrow.down")
                        .font(.system(size: 10, weight: .bold))
                        .foregroundColor(.cyan)
                    Text("下载: \(monitor.downloadSpeedFormatted)")
                        .font(.system(size: 11, weight: .medium, design: .monospaced))
                        .foregroundColor(.primary)
                }

                Spacer()

                HStack(spacing: 4) {
                    Image(systemName: "arrow.up")
                        .font(.system(size: 10, weight: .bold))
                        .foregroundColor(.green)
                    Text("上传: \(monitor.uploadSpeedFormatted)")
                        .font(.system(size: 11, weight: .medium, design: .monospaced))
                        .foregroundColor(.primary)
                }
            }

            // 动态网速波形图
            ZStack {
                RoundedRectangle(cornerRadius: 8)
                    .fill(Color.black.opacity(0.30))

                // 微弱示波器背景网格线
                VStack(spacing: 0) {
                    Spacer()
                    Divider().opacity(0.08)
                    Spacer()
                    Divider().opacity(0.08)
                    Spacer()
                }

                // 下载波形
                WaveformShape(points: monitor.downloadHistory)
                    .stroke(
                        LinearGradient(
                            colors: [.cyan.opacity(0.85), .cyan],
                            startPoint: .leading,
                            endPoint: .trailing
                        ),
                        style: StrokeStyle(lineWidth: 1.6, lineCap: .round, lineJoin: .round)
                    )

                WaveformShape(points: monitor.downloadHistory, closed: true)
                    .fill(
                        LinearGradient(
                            colors: [.cyan.opacity(0.22), .clear],
                            startPoint: .top,
                            endPoint: .bottom
                        )
                    )

                // 上传波形
                WaveformShape(points: monitor.uploadHistory)
                    .stroke(
                        LinearGradient(
                            colors: [.green.opacity(0.85), .green],
                            startPoint: .leading,
                            endPoint: .trailing
                        ),
                        style: StrokeStyle(lineWidth: 1.3, lineCap: .round, lineJoin: .round)
                    )
            }
            .frame(height: 48)
            .clipShape(RoundedRectangle(cornerRadius: 8))
        }
        .padding(10)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color.white.opacity(0.04))
                .overlay(
                    RoundedRectangle(cornerRadius: 10)
                        .stroke(Color.white.opacity(0.08), lineWidth: 1)
                )
        )
    }
}

fileprivate struct WaveformShape: Shape {
    let points: [Double]
    var closed: Bool = false

    func path(in rect: CGRect) -> Path {
        var path = Path()
        guard points.count > 1 else { return path }

        let maxVal = max(1024 * 10, points.max() ?? 1)
        let stepX = rect.width / CGFloat(points.count - 1)

        let scaledY: (Double) -> CGFloat = { val in
            let ratio = CGFloat(val / maxVal)
            return rect.height - (ratio * (rect.height - 4)) - 2
        }

        path.move(to: CGPoint(x: 0, y: scaledY(points[0])))

        // 平滑贝塞尔曲线拟合
        for i in 1..<points.count {
            let prevX = CGFloat(i - 1) * stepX
            let prevY = scaledY(points[i - 1])
            let curX = CGFloat(i) * stepX
            let curY = scaledY(points[i])
            let midX = (prevX + curX) / 2
            let midY = (prevY + curY) / 2

            if i == 1 {
                path.addLine(to: CGPoint(x: midX, y: midY))
            } else {
                path.addQuadCurve(to: CGPoint(x: midX, y: midY), control: CGPoint(x: prevX, y: prevY))
            }

            if i == points.count - 1 {
                path.addLine(to: CGPoint(x: curX, y: curY))
            }
        }

        if closed {
            path.addLine(to: CGPoint(x: rect.width, y: rect.height))
            path.addLine(to: CGPoint(x: 0, y: rect.height))
            path.closeSubpath()
        }

        return path
    }
}
