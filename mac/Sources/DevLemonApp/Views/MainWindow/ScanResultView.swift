import SwiftUI
import AppKit

public struct ScanResultView: View {
    @ObservedObject var state = AppState.shared
    @State private var expandedGroupIDs: Set<String> = []

    public init() {}

    public var body: some View {
        VStack(spacing: 0) {
            // 顶部操作栏 (Header Banner)
            headerBar

            // 次级快捷工具栏 (搜索、按大小排序、全部折叠/展开)
            searchAndSortBar

            Divider()
                .opacity(0.15)

            // 主体可折叠分组列表
            if let report = state.scanReport {
                let filteredGroups = getFilteredGroups(from: report)

                if filteredGroups.isEmpty {
                    VStack(spacing: 12) {
                        Spacer()
                        Image(systemName: "magnifyingglass")
                            .font(.system(size: 36))
                            .foregroundColor(.secondary.opacity(0.6))
                        Text("未找到与 \"\(state.searchQuery)\" 匹配的清理项目")
                            .font(.system(size: 13))
                            .foregroundColor(.secondary)
                        Button("清除搜索") {
                            state.searchQuery = ""
                        }
                        .buttonStyle(.plain)
                        .font(.system(size: 12))
                        .foregroundColor(.yellow)
                        Spacer()
                    }
                } else {
                    ScrollView {
                        LazyVStack(spacing: 12) {
                            ForEach(filteredGroups) { group in
                                GroupCardView(
                                    group: group,
                                    isExpanded: isExpanded(group.id),
                                    onToggleExpand: {
                                        withAnimation(.easeInOut(duration: 0.22)) {
                                            toggleExpand(group.id)
                                        }
                                    }
                                )
                            }
                        }
                        .padding(16)
                    }
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
                expandedGroupIDs = Set(groups.map { $0.id })
            }
        }
    }

    // MARK: - 顶部信息与主操作栏
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

                Button("推荐安全项") {
                    state.selectSafeOnly()
                }
                .buttonStyle(FilterCapsuleButtonStyle(isSelected: true))
                .help("仅选中系统日志、安装包残留及浏览器网络缓存，绝不包含任何应用数据与代码依赖")

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
        .padding(.horizontal, 18)
        .padding(.vertical, 12)
        .background(
            RoundedRectangle(cornerRadius: 10)
                .fill(Color.white.opacity(0.04))
                .overlay(
                    RoundedRectangle(cornerRadius: 10)
                        .stroke(Color.white.opacity(0.06), lineWidth: 1)
                )
        )
        .padding(.horizontal, 16)
        .padding(.top, 10)
        .padding(.bottom, 4)
    }

    // MARK: - 搜索、体积排序与展开控制工具栏
    private var searchAndSortBar: some View {
        HStack(spacing: 10) {
            // 实时搜索框
            HStack(spacing: 6) {
                Image(systemName: "magnifyingglass")
                    .font(.system(size: 11))
                    .foregroundColor(.secondary)

                TextField("搜索项目名称或路径...", text: $state.searchQuery)
                    .textFieldStyle(.plain)
                    .font(.system(size: 11.5))

                if !state.searchQuery.isEmpty {
                    Button {
                        state.searchQuery = ""
                    } label: {
                        Image(systemName: "xmark.circle.fill")
                            .font(.system(size: 11))
                            .foregroundColor(.secondary)
                    }
                    .buttonStyle(.plain)
                }
            }
            .padding(.horizontal, 9)
            .padding(.vertical, 5)
            .background(Color.white.opacity(0.06))
            .cornerRadius(6)
            .frame(maxWidth: 260)

            Spacer()

            // 按体积降序排序切换
            Button {
                withAnimation(.easeInOut(duration: 0.2)) {
                    state.sortBySize.toggle()
                }
            } label: {
                HStack(spacing: 4) {
                    Image(systemName: state.sortBySize ? "arrow.down.circle.fill" : "arrow.up.arrow.down")
                        .font(.system(size: 10.5))
                    Text(state.sortBySize ? "已按体积降序" : "按体积降序")
                        .font(.system(size: 11, weight: state.sortBySize ? .semibold : .regular))
                }
                .padding(.horizontal, 8)
                .padding(.vertical, 4)
                .background(state.sortBySize ? Color.yellow.opacity(0.16) : Color.white.opacity(0.05))
                .foregroundColor(state.sortBySize ? .yellow : .secondary)
                .cornerRadius(5)
            }
            .buttonStyle(.plain)
            .help("按占用空间大小降序排列所有条目")

            // 全部折叠 / 全部展开切换
            Button {
                toggleAllGroups()
            } label: {
                HStack(spacing: 4) {
                    Image(systemName: isAllExpanded ? "chevron.up.circle" : "chevron.down.circle")
                        .font(.system(size: 10.5))
                    Text(isAllExpanded ? "全部收起" : "全部展开")
                        .font(.system(size: 11))
                }
                .padding(.horizontal, 8)
                .padding(.vertical, 4)
                .background(Color.white.opacity(0.05))
                .foregroundColor(.secondary)
                .cornerRadius(5)
            }
            .buttonStyle(.plain)
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 4)
    }

    private var isAllExpanded: Bool {
        guard let groups = state.scanReport?.groups, !groups.isEmpty else { return false }
        return expandedGroupIDs.count >= groups.count
    }

    private func toggleAllGroups() {
        guard let groups = state.scanReport?.groups else { return }
        withAnimation(.easeInOut(duration: 0.22)) {
            if isAllExpanded {
                expandedGroupIDs.removeAll()
            } else {
                expandedGroupIDs = Set(groups.map { $0.id })
            }
        }
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

    // 过滤与排序处理
    private func getFilteredGroups(from report: ScanReport) -> [Group] {
        var result: [Group] = []

        for group in report.groups {
            var items = group.items

            // 1. 搜索词匹配
            let query = state.searchQuery.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
            if !query.isEmpty {
                items = items.filter { item in
                    item.title.lowercased().contains(query) ||
                    (item.path?.lowercased().contains(query) ?? false) ||
                    item.description.lowercased().contains(query)
                }
            }

            // 2. 排序
            if state.sortBySize {
                items.sort { $0.sizeBytes > $1.sizeBytes }
            }

            if !items.isEmpty {
                var newGroup = group
                newGroup.items = items
                result.append(newGroup)
            }
        }

        return result
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
    @State private var isHovered: Bool = false

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

            // 在访达中定位快捷按钮 (悬浮或常驻)
            if let path = item.path, !path.isEmpty {
                Button {
                    let url = URL(fileURLWithPath: path)
                    NSWorkspace.shared.activateFileViewerSelecting([url])
                } label: {
                    Image(systemName: "folder")
                        .font(.system(size: 11))
                        .foregroundColor(isHovered ? .yellow : .secondary.opacity(0.4))
                        .padding(5)
                        .background(
                            Circle()
                                .fill(isHovered ? Color.white.opacity(0.12) : Color.clear)
                        )
                }
                .buttonStyle(.plain)
                .help("在访达中显示此项目路径")
            }

            // Size
            Text(item.sizeFormatted)
                .font(.system(size: 12, weight: .medium, design: .monospaced))
                .foregroundColor(item.isProtected ? .secondary.opacity(0.5) : .primary)
        }
        .padding(.horizontal, 14)
        .padding(.vertical, 8)
        .background(isHovered ? Color.white.opacity(0.04) : Color.clear)
        .onHover { isHovered = $0 }
        .contextMenu {
            if let path = item.path, !path.isEmpty {
                Button {
                    let url = URL(fileURLWithPath: path)
                    NSWorkspace.shared.activateFileViewerSelecting([url])
                } label: {
                    Label("在访达中定位", systemImage: "folder")
                }

                Button {
                    NSPasteboard.general.clearContents()
                    NSPasteboard.general.setString(path, forType: .string)
                } label: {
                    Label("拷贝文件路径", systemImage: "doc.on.doc")
                }
            }
        }
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
