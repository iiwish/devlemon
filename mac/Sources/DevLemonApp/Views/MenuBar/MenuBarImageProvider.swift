import SwiftUI
import AppKit
import Combine

// MARK: - 状态栏纯图像模板提供器 (解决 macOS 吞 Text 与字体过大的系统问题)
@MainActor
public final class MenuBarImageProvider: ObservableObject {
    public static let shared = MenuBarImageProvider()

    @Published public var currentImage: NSImage?

    private init() {
        regenerateImage()
    }

    public func regenerateImage() {
        let view = MenuBarCompositeContent(
            settings: AppSettings.shared,
            monitor: SystemMonitor.shared
        )

        let renderer = ImageRenderer(content: view)
        renderer.scale = NSScreen.main?.backingScaleFactor ?? 2.0
        if let img = renderer.nsImage {
            img.isTemplate = true
            self.currentImage = img
        }
    }
}

// MARK: - 渲染为单张模板图像的复合内容视图
public struct MenuBarCompositeContent: View {
    @ObservedObject var settings: AppSettings
    @ObservedObject var monitor: SystemMonitor

    public init(settings: AppSettings, monitor: SystemMonitor) {
        self.settings = settings
        self.monitor = monitor
    }

    public var body: some View {
        HStack(alignment: .center, spacing: 6) {
            // 1. 单色 Logo (允许关闭)
            if settings.showLogo || isAllDisabled {
                MonochromeLemonIcon(size: 13)
                    .foregroundColor(.black)
                    .frame(height: 18)
            }

            // 2. 实时上下行网速 (双行紧凑)
            if settings.showNetwork {
                CompactNetBlock(
                    upSpeed: monitor.uploadSpeedShort,
                    downSpeed: monitor.downloadSpeedShort
                )
                .foregroundColor(.black)
                .frame(height: 18)
            }

            // 3. 内存占用 (双行紧凑：上百分比，下 MEM)
            if settings.showMemory {
                CompactMetricBlock(val: "\(Int(monitor.memoryUsage))%", label: "MEM")
                    .foregroundColor(.black)
                    .frame(height: 18)
            }

            // 4. 磁盘占用 (双行紧凑：上百分比，下 SSD)
            if settings.showDisk {
                CompactMetricBlock(val: "\(Int(monitor.diskUsagePercent))%", label: "SSD")
                    .foregroundColor(.black)
                    .frame(height: 18)
            }

            // 5. CPU 占用 (双行紧凑：上百分比，下 CPU)
            if settings.showCPU {
                CompactMetricBlock(val: "\(Int(monitor.cpuUsage))%", label: "CPU")
                    .foregroundColor(.black)
                    .frame(height: 18)
            }
        }
        .padding(.horizontal, 2)
        .padding(.vertical, 1)
    }

    private var isAllDisabled: Bool {
        !settings.showLogo && !settings.showNetwork && !settings.showMemory && !settings.showDisk && !settings.showCPU
    }
}
