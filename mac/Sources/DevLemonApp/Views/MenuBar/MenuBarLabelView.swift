import SwiftUI

public enum MenuBarDisplayStyle: String, CaseIterable, Identifiable {
    case networkDual = "networkDual"          // 🍋 紧凑双行网速 (↑ 12K / ↓ 85K)
    case networkInline = "networkInline"      // 🍋 ↑ 12K  ↓ 85K (单行)
    case networkAndDisk = "networkAndDisk"    // 🍋 ↓ 85K · 28%
    case cpuAndMemory = "cpuAndMemory"        // 🍋 C:12% M:65%
    case diskOnly = "diskOnly"                // 🍋 28%

    public var id: String { rawValue }

    public var title: String {
        switch self {
        case .networkDual: return "实时上下行网速 (双行紧凑 · 推荐)"
        case .networkInline: return "实时上下行网速 (单行平铺)"
        case .networkAndDisk: return "实时下载网速 + 主盘容量"
        case .cpuAndMemory: return "CPU + 内存实时负载"
        case .diskOnly: return "仅主盘容量百分比"
        }
    }
}

public struct MenuBarLabelView: View {
    @ObservedObject var monitor = SystemMonitor.shared
    @AppStorage("menuBarDisplayStyle") private var displayStyle: String = MenuBarDisplayStyle.networkDual.rawValue

    public init() {}

    public var body: some View {
        HStack(spacing: 4) {
            Text("🍋")
                .font(.system(size: 12))

            switch MenuBarDisplayStyle(rawValue: displayStyle) ?? .networkDual {
            case .networkDual:
                // 紧凑双行网速 (经典柠檬样式)
                VStack(alignment: .leading, spacing: -1) {
                    HStack(spacing: 2) {
                        Image(systemName: "arrow.up")
                            .font(.system(size: 7, weight: .bold))
                            .foregroundColor(.green)
                        Text(monitor.uploadSpeedFormatted)
                            .font(.system(size: 8.5, weight: .medium, design: .monospaced))
                    }
                    HStack(spacing: 2) {
                        Image(systemName: "arrow.down")
                            .font(.system(size: 7, weight: .bold))
                            .foregroundColor(.cyan)
                        Text(monitor.downloadSpeedFormatted)
                            .font(.system(size: 8.5, weight: .medium, design: .monospaced))
                    }
                }

            case .networkInline:
                HStack(spacing: 3) {
                    Image(systemName: "arrow.up")
                        .font(.system(size: 8, weight: .bold))
                        .foregroundColor(.green)
                    Text(monitor.uploadSpeedShort)
                        .font(.system(size: 10, weight: .medium, design: .monospaced))
                    Image(systemName: "arrow.down")
                        .font(.system(size: 8, weight: .bold))
                        .foregroundColor(.cyan)
                    Text(monitor.downloadSpeedShort)
                        .font(.system(size: 10, weight: .medium, design: .monospaced))
                }

            case .networkAndDisk:
                HStack(spacing: 4) {
                    Image(systemName: "arrow.down")
                        .font(.system(size: 8, weight: .bold))
                        .foregroundColor(.cyan)
                    Text(monitor.downloadSpeedShort)
                        .font(.system(size: 10, weight: .medium, design: .monospaced))
                    Text("·")
                        .foregroundColor(.secondary)
                    Text("\(Int(monitor.diskUsagePercent))%")
                        .font(.system(size: 10, weight: .medium, design: .monospaced))
                }

            case .cpuAndMemory:
                Text("C:\(Int(monitor.cpuUsage))% M:\(Int(monitor.memoryUsage))%")
                    .font(.system(size: 10, weight: .medium, design: .monospaced))

            case .diskOnly:
                Text("\(Int(monitor.diskUsagePercent))%")
                    .font(.system(size: 10, weight: .medium, design: .monospaced))
            }
        }
    }
}
