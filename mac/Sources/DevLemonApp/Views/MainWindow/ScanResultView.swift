import SwiftUI

public struct ScanResultView: View {
    @ObservedObject var state = AppState.shared
    @State private var expandedGroupIDs: Set<String> = []

    public init() {}

    public var body: some View {
        VStack(spacing: 0) {
            // 顶部操作栏 (Header Banner)
            headerBar

            Divider()
                .opacity(0.15)

            // 主体可折叠分组列表
            if let report = state.scanReport {
                ScrollView {
                    LazyVStack(spacing: 12) {
                        ForEach(report.groups) { group in
                            GroupCardView(
                                group: group,
                                isExpanded: isExpanded(group.id),
                                onToggleExpand: { toggleExpand(group.id) }
                            )
                        }
                    }
                    .padding(16)
                }
            } else {
                VStack(spacing: 12) {
                    Spacer()
                    Image(systemName: "tray")
                        .font(.system(size: 40))
                        .foregroundColor(.secondary)
                    Text("暂无扫描数据")
                        .font(.system(size: 14))
                        .foregroundColor(.secondary)
                    Spacer()
                }
            }
        }
        .onAppear {
            if let groups = state.scanReport?.groups {
                // 默认展开所有有内容的分组
                expandedGroupIDs = Set(groups.map { $0.id })
            }
        }
    }

    private var headerBar: some View {
        HStack(alignment: .center) {
            // 左侧：释放空间大字
            VStack(alignment: .leading, spacing: 4) {
                HStack(alignment: .firstTextBaseline, spacing: 6) {
                    Text(state.scanReport?.totalReclaimableFormatted ?? "0 B")
                        .font(.system(size: 26, weight: .bold, design: .rounded))
                        .foregroundColor(.primary)

                    Text("可清理资源")
                        .font(.system(size: 13, weight: .medium))
                        .foregroundColor(.secondary)
                }

                Text("已选 \(state.selectedItemIDs.count) 项 · 预计释放 \(state.selectedTotalFormatted)")
                    .font(.system(size: 11))
                    .foregroundColor(state.selectedItemIDs.isEmpty ? .secondary : .yellow)
            }

            Spacer()

            // 中间：快速选择过滤胶囊
            HStack(spacing: 6) {
                Button("全选") {
                    state.selectAll()
                }
                .buttonStyle(FilterCapsuleButtonStyle(isSelected: false))

                Button("仅安全项") {
                    state.selectSafeOnly()
                }
                .buttonStyle(FilterCapsuleButtonStyle(isSelected: true))

                Button("清空") {
                    state.deselectAll()
                }
                .buttonStyle(FilterCapsuleButtonStyle(isSelected: false))
            }

            Spacer()

            // 右侧：重新扫描与立即清理按钮
            HStack(spacing: 10) {
                Button {
                    state.startScan()
                } label: {
                    Image(systemName: "arrow.clockwise")
                        .font(.system(size: 12, weight: .medium))
                        .padding(8)
                        .background(Color.white.opacity(0.08))
                        .foregroundColor(.primary)
                        .clipShape(Circle())
                }
                .buttonStyle(.plain)
                .help("重新扫描")

                Button {
                    state.startClean()
                } label: {
                    HStack(spacing: 6) {
                        Image(systemName: "sparkles")
                        Text("立即清理")
                            .font(.system(size: 13, weight: .bold))
                    }
                    .padding(.horizontal, 16)
                    .padding(.vertical, 8)
                    .background(
                        state.selectedItemIDs.isEmpty
                            ? LinearGradient(colors: [Color.gray.opacity(0.3)], startPoint: .top, endPoint: .bottom)
                            : LinearGradient(
                                colors: [Color(red: 0.98, green: 0.88, blue: 0.25), Color(red: 0.90, green: 0.72, blue: 0.10)],
                                startPoint: .topLeading,
                                endPoint: .bottomTrailing
                            )
                    )
                    .foregroundColor(state.selectedItemIDs.isEmpty ? .secondary : .black)
                    .cornerRadius(8)
                    .shadow(color: state.selectedItemIDs.isEmpty ? .clear : Color.yellow.opacity(0.3), radius: 5, y: 2)
                }
                .buttonStyle(.plain)
                .disabled(state.selectedItemIDs.isEmpty)
            }
        }
        .padding(.horizontal, 20)
        .padding(.vertical, 14)
        .background(Color.black.opacity(0.12))
    }

    private func isExpanded(_ groupID: String) -> Bool {
        expandedGroupIDs.contains(groupID)
    }

    private func toggleExpand(_ groupID: String) {
        if expandedGroupIDs.contains(groupID) {
            expandedGroupIDs.remove(groupID)
        } else {
            expandedGroupIDs.insert(groupID)
        }
    }
}

fileprivate struct GroupCardView: View {
    let group: Group
    let isExpanded: Bool
    let onToggleExpand: () -> Void
    @ObservedObject var state = AppState.shared

    var body: some View {
        VStack(spacing: 0) {
            // Group Header
            HStack(spacing: 12) {
                // Group Checkbox
                Button {
                    state.toggleGroupSelection(group)
                } label: {
                    Image(systemName: groupCheckboxIcon)
                        .font(.system(size: 14))
                        .foregroundColor(state.isGroupAllSelected(group) || state.isGroupPartiallySelected(group) ? .yellow : .secondary)
                }
                .buttonStyle(.plain)

                // Category Icon
                Image(systemName: group.category.iconName)
                    .font(.system(size: 14))
                    .foregroundColor(.yellow)
                    .frame(width: 20)

                // Title
                Text(group.title)
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundColor(.primary)

                Text("(\(group.items.count))")
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)

                Spacer()

                // Total Group Size
                Text(ByteFormatter.format(group.totalReclaimableBytes))
                    .font(.system(size: 13, weight: .bold, design: .rounded))
                    .foregroundColor(.primary)

                // Expand/Collapse Chevron
                Button(action: onToggleExpand) {
                    Image(systemName: isExpanded ? "chevron.up" : "chevron.down")
                        .font(.system(size: 11, weight: .medium))
                        .foregroundColor(.secondary)
                        .padding(4)
                }
                .buttonStyle(.plain)
            }
            .padding(.horizontal, 14)
            .padding(.vertical, 10)
            .background(Color.white.opacity(0.04))
            .contentShape(Rectangle())
            .onTapGesture {
                onToggleExpand()
            }

            // Group Items
            if isExpanded {
                VStack(spacing: 0) {
                    ForEach(group.items) { item in
                        ItemRowView(item: item)
                        if item.id != group.items.last?.id {
                            Divider()
                                .opacity(0.08)
                                .padding(.leading, 42)
                        }
                    }
                }
                .background(Color.black.opacity(0.18))
            }
        }
        .clipShape(RoundedRectangle(cornerRadius: 10))
        .overlay(
            RoundedRectangle(cornerRadius: 10)
                .stroke(Color.white.opacity(0.08), lineWidth: 1)
        )
    }

    private var groupCheckboxIcon: String {
        if state.isGroupAllSelected(group) {
            return "checkmark.square.fill"
        } else if state.isGroupPartiallySelected(group) {
            return "minus.square.fill"
        } else {
            return "square"
        }
    }
}

fileprivate struct ItemRowView: View {
    let item: Item
    @ObservedObject var state = AppState.shared

    var body: some View {
        HStack(spacing: 12) {
            // Checkbox
            if item.isProtected {
                Image(systemName: "lock.fill")
                    .font(.system(size: 12))
                    .foregroundColor(.secondary.opacity(0.6))
                    .help(item.protectReason ?? "运行中受保护资源")
            } else {
                Button {
                    state.toggleItemSelection(item)
                } label: {
                    Image(systemName: state.selectedItemIDs.contains(item.id) ? "checkmark.square.fill" : "square")
                        .font(.system(size: 13))
                        .foregroundColor(state.selectedItemIDs.contains(item.id) ? .yellow : .secondary)
                }
                .buttonStyle(.plain)
            }

            // Details
            VStack(alignment: .leading, spacing: 2) {
                HStack(spacing: 8) {
                    Text(item.title)
                        .font(.system(size: 12, weight: .medium))
                        .foregroundColor(item.isProtected ? .secondary : .primary)
                        .lineLimit(1)

                    RiskBadge(risk: item.risk)

                    if item.isProtected {
                        Text("运行保护")
                            .font(.system(size: 9, weight: .medium))
                            .foregroundColor(.secondary)
                            .padding(.horizontal, 4)
                            .padding(.vertical, 1)
                            .background(Color.gray.opacity(0.2))
                            .cornerRadius(3)
                    }
                }

                if let path = item.path, !path.isEmpty {
                    Text(path)
                        .font(.system(size: 10, design: .monospaced))
                        .foregroundColor(.secondary.opacity(0.7))
                        .lineLimit(1)
                }
            }

            Spacer()

            // Size
            Text(item.sizeFormatted)
                .font(.system(size: 12, weight: .medium, design: .monospaced))
                .foregroundColor(item.isProtected ? .secondary.opacity(0.5) : .primary)
        }
        .padding(.horizontal, 14)
        .padding(.vertical, 8)
        .help(item.path ?? item.description)
    }
}

fileprivate struct RiskBadge: View {
    let risk: RiskLevel

    var body: some View {
        Text(risk.title)
            .font(.system(size: 9, weight: .bold))
            .foregroundColor(risk.color)
            .padding(.horizontal, 5)
            .padding(.vertical, 1.5)
            .background(risk.color.opacity(0.15))
            .cornerRadius(4)
    }
}

fileprivate struct FilterCapsuleButtonStyle: ButtonStyle {
    let isSelected: Bool

    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .font(.system(size: 11, weight: .medium))
            .padding(.horizontal, 10)
            .padding(.vertical, 4)
            .background(
                Capsule()
                    .fill(isSelected ? Color.yellow.opacity(0.18) : Color.white.opacity(0.06))
            )
            .foregroundColor(isSelected ? .yellow : .secondary)
            .scaleEffect(configuration.isPressed ? 0.96 : 1.0)
    }
}
