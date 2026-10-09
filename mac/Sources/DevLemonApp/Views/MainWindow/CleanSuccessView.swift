import SwiftUI

public struct CleanSuccessView: View {
    @ObservedObject var state = AppState.shared
    @ObservedObject var monitor = SystemMonitor.shared

    @State private var checkmarkScale: CGFloat = 0.5
    @State private var opacity: Double = 0.0
    @State private var isDoneHovered: Bool = false

    public init() {}

    public var body: some View {
        VStack(spacing: 22) {
            Spacer()

            // 成功徽章动画
            ZStack {
                Circle()
                    .fill(Color.green.opacity(0.12))
                    .frame(width: 110, height: 110)

                Circle()
                    .stroke(
                        LinearGradient(
                            colors: [Color.green.opacity(0.6), Color.green.opacity(0.2)],
                            startPoint: .topLeading,
                            endPoint: .bottomTrailing
                        ),
                        lineWidth: 2.5
                    )
                    .frame(width: 110, height: 110)

                Image(systemName: "checkmark")
                    .font(.system(size: 48, weight: .bold))
                    .foregroundColor(.green)
            }
            .scaleEffect(checkmarkScale)

            // 文案与数据高光
            VStack(spacing: 8) {
                Text("清理完毕")
                    .font(.system(size: 24, weight: .bold, design: .rounded))
                    .foregroundColor(.primary)

                if let result = state.lastCleanResult {
                    HStack(alignment: .firstTextBaseline, spacing: 4) {
                        Text("共释放")
                            .font(.system(size: 14))
                            .foregroundColor(.secondary)

                        Text(result.freedFormatted)
                            .font(.system(size: 28, weight: .bold, design: .rounded))
                            .foregroundColor(.green)

                        Text("磁盘空间")
                            .font(.system(size: 14))
                            .foregroundColor(.secondary)
                    }

                    Text("成功处理 \(result.cleanedCount) 个项目")
                        .font(.system(size: 12))
                        .foregroundColor(.secondary)

                    if let errs = result.errors, !errs.isEmpty {
                        Text("\(errs.count) 项由于系统权限保护未能移除")
                            .font(.system(size: 11))
                            .foregroundColor(.orange)
                            .padding(.top, 2)
                    }
                }
            }
            .opacity(opacity)

            // 前后对比指示微卡片 (增强清理完成的获得感)
            HStack(spacing: 16) {
                VStack(spacing: 3) {
                    Text("清理前剩余")
                        .font(.system(size: 10.5))
                        .foregroundColor(.secondary)
                    Text(state.diskFreeBeforeClean.isEmpty ? monitor.diskFreeFormatted : state.diskFreeBeforeClean)
                        .font(.system(size: 13, weight: .semibold, design: .monospaced))
                        .foregroundColor(.secondary)
                }

                Image(systemName: "arrow.right")
                    .font(.system(size: 12, weight: .bold))
                    .foregroundColor(.green.opacity(0.8))

                VStack(spacing: 3) {
                    Text("清理后剩余")
                        .font(.system(size: 10.5))
                        .foregroundColor(.secondary)
                    Text(state.diskFreeAfterClean.isEmpty ? monitor.diskFreeFormatted : state.diskFreeAfterClean)
                        .font(.system(size: 14, weight: .bold, design: .monospaced))
                        .foregroundColor(.green)
                }
            }
            .padding(.horizontal, 20)
            .padding(.vertical, 10)
            .background(
                RoundedRectangle(cornerRadius: 10)
                    .fill(Color.white.opacity(0.04))
                    .overlay(
                        RoundedRectangle(cornerRadius: 10)
                            .stroke(Color.white.opacity(0.08), lineWidth: 1)
                    )
            )
            .opacity(opacity)

            Spacer()

            // 完成按钮
            Button {
                state.resetToIdle()
            } label: {
                Text("完成")
                    .font(.system(size: 14, weight: .semibold))
                    .frame(width: 160)
                    .padding(.vertical, 9)
                    .background(
                        RoundedRectangle(cornerRadius: 8)
                            .fill(isDoneHovered ? Color.white.opacity(0.16) : Color.white.opacity(0.09))
                            .overlay(
                                RoundedRectangle(cornerRadius: 8)
                                    .stroke(Color.white.opacity(isDoneHovered ? 0.22 : 0.12), lineWidth: 1)
                            )
                    )
                    .foregroundColor(.primary)
            }
            .buttonStyle(.plain)
            .onHover { isDoneHovered = $0 }
            .padding(.bottom, 24)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .onAppear {
            withAnimation(.spring(response: 0.5, dampingFraction: 0.6)) {
                checkmarkScale = 1.0
            }
            withAnimation(.easeIn(duration: 0.4).delay(0.2)) {
                opacity = 1.0
            }
        }
    }
}
