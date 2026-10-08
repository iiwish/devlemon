import SwiftUI

public struct MenuBarLabelView: View {
    @ObservedObject var settings = AppSettings.shared
    @ObservedObject var monitor = SystemMonitor.shared

    public init() {}

    public var body: some View {
        HStack(alignment: .center, spacing: 5) {
            // 1. 单色 Logo
            if settings.showLogo || isAllDisabled {
                MonochromeLemonIcon(size: 13)
                    .frame(height: 22)
            }

            // 2. 实时上下行网速 (双行紧凑)
            if settings.showNetwork {
                CompactNetBlock(
                    upSpeed: monitor.uploadSpeedShort,
                    downSpeed: monitor.downloadSpeedShort
                )
                .frame(height: 22)
            }

            // 3. 内存占用 (双行：上百分比，下 MEM)
            if settings.showMemory {
                CompactMetricBlock(
                    val: "\(Int(monitor.memoryUsage))%",
                    label: "MEM"
                )
                .frame(height: 22)
            }

            // 4. 磁盘占用 (双行：上百分比，下 SSD)
            if settings.showDisk {
                CompactMetricBlock(
                    val: "\(Int(monitor.diskUsagePercent))%",
                    label: "SSD"
                )
                .frame(height: 22)
            }

            // 5. CPU 占用 (双行：上百分比，下 CPU)
            if settings.showCPU {
                CompactMetricBlock(
                    val: "\(Int(monitor.cpuUsage))%",
                    label: "CPU"
                )
                .frame(height: 22)
            }
        }
        .foregroundColor(.primary)
        .fixedSize()
    }

    private var isAllDisabled: Bool {
        !settings.showLogo && !settings.showNetwork && !settings.showMemory && !settings.showDisk && !settings.showCPU
    }
}

// 腾讯柠檬同款：上下双行极致紧凑微指标卡片 (高度 22pt，宽度仅 22~25pt)
public struct CompactMetricBlock: View {
    public let val: String
    public let label: String

    public init(val: String, label: String) {
        self.val = val
        self.label = label
    }

    public var body: some View {
        VStack(alignment: .center, spacing: -2) {
            Text(val)
                .font(.system(size: 8.5, weight: .bold, design: .rounded))
                .lineLimit(1)

            Text(label)
                .font(.system(size: 6.5, weight: .semibold, design: .rounded))
                .foregroundColor(.primary.opacity(0.7))
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
