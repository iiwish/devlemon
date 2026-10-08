import SwiftUI

public struct MenuBarLabelView: View {
    @ObservedObject var monitor = SystemMonitor.shared

    public init() {}

    public var body: some View {
        HStack(spacing: 4) {
            Image(systemName: "shield.lefthalf.filled")
                .foregroundColor(.yellow)

            Text("DevLemon")
                .font(.system(size: 11, weight: .semibold, design: .rounded))

            Text("\(Int(monitor.diskUsagePercent))%")
                .font(.system(size: 10, weight: .medium, design: .monospaced))
                .foregroundColor(.secondary)
        }
    }
}
