import SwiftUI
import AppKit

/// 菜单栏浮窗中的应用内存占用面板：按应用聚合的内存排行，悬停可退出应用，顶部可释放系统内存
public struct AppMemoryPanel: View {
    @ObservedObject var memory = AppMemoryMonitor.shared
    @ObservedObject var monitor = SystemMonitor.shared
    var onCollapse: () -> Void

    @State private var hoveredID: String?

    public init(onCollapse: @escaping () -> Void) {
        self.onCollapse = onCollapse
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            header

            if memory.isRestricted {
                VStack(alignment: .leading, spacing: 6) {
                    Text("无法读取其他应用的内存占用。")
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)
                        .fixedSize(horizontal: false, vertical: true)
                    Button("在活动监视器中查看") { Self.openActivityMonitor() }
                        .buttonStyle(.link)
                        .font(.system(size: 11))
                }
                .padding(.vertical, 6)
            } else if memory.items.isEmpty {
                HStack {
                    Spacer()
                    ProgressView().scaleEffect(0.7)
                    Spacer()
                }
                .frame(height: 60)
            } else {
                VStack(spacing: 4) {
                    ForEach(memory.items) { item in
                        row(item)
                    }
                }
            }
        }
        .padding(10)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color.white.opacity(0.045))
                .overlay(RoundedRectangle(cornerRadius: 10).stroke(Color.white.opacity(0.08), lineWidth: 1))
        )
        .onAppear { memory.start() }
        .onDisappear { memory.stop() }
    }

    static func openActivityMonitor() {
        if let url = NSWorkspace.shared.urlForApplication(withBundleIdentifier: "com.apple.ActivityMonitor") {
            NSWorkspace.shared.open(url)
        }
    }

    // MARK: - 标题栏：内存占用率 + 释放按钮 + 收起

    private var header: some View {
        HStack(spacing: 6) {
            Image(systemName: "memorychip")
                .font(.system(size: 12))
                .foregroundColor(.secondary)
            Text("内存占用")
                .font(.system(size: 12, weight: .semibold))
            Text("\(Int(monitor.memoryUsage.rounded()))%")
                .font(.system(size: 12, weight: .bold, design: .rounded))
                .foregroundColor(memoryColor)

            Spacer()

            purgeButton

            Button(action: onCollapse) {
                Image(systemName: "chevron.up")
                    .font(.system(size: 10, weight: .semibold))
                    .foregroundColor(.secondary)
                    .frame(width: 20, height: 20)
                    .contentShape(Rectangle())
            }
            .buttonStyle(.plain)
            .help("收起")
        }
    }

    private var memoryColor: Color {
        switch monitor.memoryUsage {
        case ..<60: return .green
        case ..<85: return .orange
        default: return .red
        }
    }

    @ViewBuilder
    private var purgeButton: some View {
        if memory.canPurge {
            switch memory.purgeState {
            case .running:
                ProgressView().scaleEffect(0.5).frame(width: 36, height: 18)
            case .done(let freed):
                Text(freed > 0 ? "已释放 \(ByteFormatter.format(freed))" : "已是最佳")
                    .font(.system(size: 11, weight: .medium))
                    .foregroundColor(.green)
            case .failed:
                Text("释放失败")
                    .font(.system(size: 11, weight: .medium))
                    .foregroundColor(.red)
            case .idle:
                Button {
                    memory.purge()
                } label: {
                    Text("释放")
                        .font(.system(size: 12, weight: .semibold))
                        .foregroundColor(.accentColor)
                }
                .buttonStyle(.plain)
                .help("清空系统可回收的文件缓存以腾出空闲内存（需要管理员密码）。macOS 会自动管理内存，通常只在内存紧张时才需要")
            }
        }
    }

    // MARK: - 应用行

    private func row(_ item: AppMemoryItem) -> some View {
        let maxBytes = max(memory.items.first?.footprintBytes ?? 1, 1)
        let ratio = CGFloat(Double(item.footprintBytes) / Double(maxBytes))
        let hovered = hoveredID == item.id

        return HStack(spacing: 8) {
            Image(nsImage: AppIconCache.shared.icon(for: item.bundlePath))
                .resizable()
                .frame(width: 18, height: 18)
            Text(item.name)
                .font(.system(size: 12))
                .lineLimit(1)
                .truncationMode(.tail)
            Spacer(minLength: 6)
            if hovered {
                quitButton(item)
            } else {
                Text(ByteFormatter.format(item.footprintBytes))
                    .font(.system(size: 12, weight: .medium, design: .rounded))
                    .foregroundColor(.primary)
            }
        }
        .padding(.horizontal, 8)
        .frame(height: 26)
        .background(
            GeometryReader { geo in
                ZStack(alignment: .leading) {
                    RoundedRectangle(cornerRadius: 5)
                        .fill(Color.white.opacity(hovered ? 0.06 : 0))
                    RoundedRectangle(cornerRadius: 5)
                        .fill(Color.white.opacity(0.09))
                        .frame(width: max(4, geo.size.width * ratio))
                }
            }
        )
        .contentShape(Rectangle())
        .onHover { inside in
            if inside {
                hoveredID = item.id
            } else if hoveredID == item.id {
                hoveredID = nil
            }
        }
        .help(item.processCount > 1 ? "\(item.name)：共 \(item.processCount) 个进程" : item.name)
    }

    @ViewBuilder
    private func quitButton(_ item: AppMemoryItem) -> some View {
        switch memory.quitStates[item.id] {
        case .quitting:
            Text("退出中…")
                .font(.system(size: 11))
                .foregroundColor(.secondary)
        case .unresponsive:
            Button {
                memory.forceQuit(item)
            } label: {
                pill("强制退出", color: .red)
            }
            .buttonStyle(.plain)
            .help("应用未响应退出请求。强制退出会丢失未保存的内容")
        case nil:
            Button {
                memory.quit(item)
            } label: {
                pill("退出", color: .orange)
            }
            .buttonStyle(.plain)
            .help("退出 \(item.name)（与在应用中按 ⌘Q 相同，有未保存内容时应用会询问）")
        }
    }

    private func pill(_ text: String, color: Color) -> some View {
        Text(text)
            .font(.system(size: 11, weight: .semibold))
            .foregroundColor(color)
            .padding(.horizontal, 8)
            .padding(.vertical, 2)
            .background(color.opacity(0.15))
            .cornerRadius(4)
    }
}

/// 应用图标缓存，避免每次刷新重复读取
final class AppIconCache {
    static let shared = AppIconCache()
    private var cache: [String: NSImage] = [:]

    func icon(for bundlePath: String) -> NSImage {
        if let img = cache[bundlePath] { return img }
        let img = NSWorkspace.shared.icon(forFile: bundlePath)
        img.size = NSSize(width: 36, height: 36)
        cache[bundlePath] = img
        return img
    }
}
