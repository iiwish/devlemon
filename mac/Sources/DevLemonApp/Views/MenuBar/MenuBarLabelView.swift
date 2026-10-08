import SwiftUI

public struct MenuBarLabelView: View {
    @ObservedObject var settings = AppSettings.shared
    @ObservedObject var monitor = SystemMonitor.shared

    public init() {}

    public var body: some View {
        HStack(spacing: 5) {
            // 1. 单色 Logo (允许关闭)
            if settings.showLogo || isAllDisabled {
                MonochromeLemonView()
            }

            // 2. 实时网速 (双行紧凑)
            if settings.showNetwork {
                VStack(alignment: .leading, spacing: -1) {
                    HStack(spacing: 2) {
                        Image(systemName: "arrow.up")
                            .font(.system(size: 7, weight: .bold))
                            .foregroundColor(.primary.opacity(0.8))
                        Text(monitor.uploadSpeedFormatted)
                            .font(.system(size: 8.5, weight: .medium, design: .monospaced))
                    }
                    HStack(spacing: 2) {
                        Image(systemName: "arrow.down")
                            .font(.system(size: 7, weight: .bold))
                            .foregroundColor(.primary.opacity(0.8))
                        Text(monitor.downloadSpeedFormatted)
                            .font(.system(size: 8.5, weight: .medium, design: .monospaced))
                    }
                }
            }

            // 3. 内存占用
            if settings.showMemory {
                Text("\(Int(monitor.memoryUsage))% MEM")
                    .font(.system(size: 9.5, weight: .medium, design: .monospaced))
            }

            // 4. 磁盘占用
            if settings.showDisk {
                Text("\(Int(monitor.diskUsagePercent))% SSD")
                    .font(.system(size: 9.5, weight: .medium, design: .monospaced))
            }

            // 5. CPU 占用
            if settings.showCPU {
                Text("\(Int(monitor.cpuUsage))% CPU")
                    .font(.system(size: 9.5, weight: .medium, design: .monospaced))
            }
        }
        .foregroundColor(.primary)
    }

    private var isAllDisabled: Bool {
        !settings.showLogo && !settings.showNetwork && !settings.showMemory && !settings.showDisk && !settings.showCPU
    }
}

public struct MonochromeLemonView: View {
    public init() {}

    public var body: some View {
        MonochromeLemonIcon(size: 14)
    }
}
