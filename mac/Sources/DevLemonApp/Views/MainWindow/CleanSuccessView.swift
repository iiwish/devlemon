import SwiftUI

public struct CleanSuccessView: View {
    @ObservedObject var state = AppState.shared
    @State private var checkmarkScale: CGFloat = 0.5
    @State private var opacity: Double = 0.0

    public init() {}

    public var body: some View {
        VStack(spacing: 24) {
            Spacer()

            // 成功徽章动画
            ZStack {
                Circle()
                    .fill(Color.green.opacity(0.12))
                    .frame(width: 120, height: 120)

                Circle()
                    .stroke(Color.green.opacity(0.4), lineWidth: 2)
                    .frame(width: 120, height: 120)

                Image(systemName: "checkmark")
                    .font(.system(size: 52, weight: .bold))
                    .foregroundColor(.green)
            }
            .scaleEffect(checkmarkScale)

            // 文案展示
            VStack(spacing: 8) {
                Text("清理完毕")
                    .font(.system(size: 22, weight: .bold))
                    .foregroundColor(.primary)

                if let result = state.lastCleanResult {
                    HStack(alignment: .firstTextBaseline, spacing: 4) {
                        Text("共释放")
                            .font(.system(size: 14))
                            .foregroundColor(.secondary)

                        Text(result.freedFormatted)
                            .font(.system(size: 26, weight: .bold, design: .rounded))
                            .foregroundColor(.green)

                        Text("磁盘空间")
                            .font(.system(size: 14))
                            .foregroundColor(.secondary)
                    }

                    Text("成功处理 \(result.cleanedCount) 个项目")
                        .font(.system(size: 12))
                        .foregroundColor(.secondary)

                    if let errs = result.errors, !errs.isEmpty {
                        Text("\(errs.count) 项由于系统权限未能移除")
                            .font(.system(size: 11))
                            .foregroundColor(.orange)
                            .padding(.top, 4)
                    }
                }
            }
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
                    .background(Color.white.opacity(0.1))
                    .foregroundColor(.primary)
                    .cornerRadius(8)
            }
            .buttonStyle(.plain)
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
