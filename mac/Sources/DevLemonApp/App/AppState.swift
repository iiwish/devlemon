import Foundation
import SwiftUI
import Combine

public enum AppStage {
    case idle
    case scanning
    case results
    case cleaning
    case cleaned
}

public enum QuickCleanStatus: Equatable {
    case idle
    case cleaning
    case success(String)
}

public final class AppState: ObservableObject {
    public static let shared = AppState()

    @Published public var currentStage: AppStage = .idle
    @Published public var scanReport: ScanReport?
    @Published public var lastCleanResult: CleanResult?
    @Published public var errorMessage: String?

    // 快捷清理状态 (用于下拉菜单快捷卡片)
    @Published public var quickCleanStatus: QuickCleanStatus = .idle
    @Published public var safeReclaimableBytes: Int64 = 0
    @Published public var hasCheckedSafeClean: Bool = false

    // 结果搜索与排序过滤
    @Published public var searchQuery: String = ""
    @Published public var sortBySize: Bool = false

    // 清理进度细节
    @Published public var cleaningCurrentItemTitle: String = ""
    @Published public var diskFreeBeforeClean: String = ""
    @Published public var diskFreeAfterClean: String = ""

    // 选中的清理项 ID 集合
    @Published public var selectedItemIDs: Set<String> = []

    // 扫描模拟动态反馈（用于雷达动画阶段文案与轮播路径）
    @Published public var scanPhaseText: String = "正在扫描系统与工作区..."
    @Published public var scanningCurrentPath: String = "~/Projects"
    @Published public var scanProgress: Double = 0.0

    // 是否显示偏好设置
    @Published public var isShowingSettings: Bool = false

    private var scanSimulationTimer: AnyCancellable?

    private init() {}

    /// 触发全面深度扫描
    @MainActor
    public func startScan(safeOnly: Bool = false) {
        guard currentStage != .scanning && currentStage != .cleaning else { return }

        // 若处于沙盒环境 (App Store 版) 且未授权，先引导用户选取个人主文件夹
        if SecurityBookmarkManager.shared.isSandboxed && !SecurityBookmarkManager.shared.hasAuthorization {
            SecurityBookmarkManager.shared.requestAuthorization { granted in
                if granted {
                    self.startScan(safeOnly: safeOnly)
                }
            }
            return
        }

        // 激活安全作用域访问
        _ = SecurityBookmarkManager.shared.startAccessing()

        currentStage = .scanning
        errorMessage = nil
        scanProgress = 0.0
        selectedItemIDs.removeAll()

        startScanAnimationSimulation()

        Task {
            do {
                let report = try await CLIBridge.shared.scan(safeOnly: safeOnly)
                await MainActor.run {
                    self.stopScanAnimationSimulation()
                    self.scanReport = report
                    // 默认保守预选策略（对标柠檬清理）：
                    // 1. 仅预选绝对纯净无感知的系统基础维护项与浏览器网络缓存（Category: .systemCache 且 risk == .safe）
                    // 2. 属于应用垃圾 (appCache) 或开发工具链/包管理缓存/Docker 等项目，默认一律【不勾选】！
                    // 3. 所有 rebuildable / caution 风险等级项目，默认一律【不勾选】！
                    // 4. 让用户在知情状态下自主按需勾选应用或代码依赖进行清理，彻底消除误删恐慌。
                    var defaultSelected = Set<String>()
                    for grp in report.groups {
                        for item in grp.items {
                            if !item.isProtected && item.risk == .safe && item.category == .systemCache {
                                defaultSelected.insert(item.id)
                            }
                        }
                    }
                    self.selectedItemIDs = defaultSelected
                    self.scanProgress = 1.0
                    withAnimation(.easeInOut(duration: 0.3)) {
                        self.currentStage = .results
                    }
                }
            } catch {
                await MainActor.run {
                    self.stopScanAnimationSimulation()
                    if let bridgeErr = error as? CLIBridgeError, case .cancelled = bridgeErr {
                        // 用户主动停止，无需提示错误
                        self.currentStage = .idle
                    } else {
                        self.errorMessage = error.localizedDescription
                        self.currentStage = .idle
                    }
                }
            }
        }
    }

    /// 手动终止当前深度扫描
    @MainActor
    public func stopScan() {
        CLIBridge.shared.cancelActiveOperation()
        stopScanAnimationSimulation()
        withAnimation(.easeInOut(duration: 0.25)) {
            self.currentStage = .idle
            self.scanProgress = 0.0
            self.errorMessage = nil
        }
    }

    /// 刷新下拉菜单中的安全垃圾可清理数值
    @MainActor
    public func refreshSafeReclaimableSize() {
        Task {
            do {
                let report = try await CLIBridge.shared.scan(safeOnly: true)
                var safeBytes: Int64 = 0
                for grp in report.groups {
                    for item in grp.items where !item.isProtected && item.risk == .safe {
                        safeBytes += item.sizeBytes
                    }
                }
                await MainActor.run {
                    self.safeReclaimableBytes = safeBytes
                    self.hasCheckedSafeClean = true
                }
            } catch {
                print("Fast safe scan failed: \(error)")
            }
        }
    }

    /// 下拉菜单一键极速安全清理 (只清理绝对安全的临时缓存与垃圾)
    @MainActor
    public func startQuickClean() {
        guard quickCleanStatus != .cleaning else { return }
        _ = SecurityBookmarkManager.shared.startAccessing()
        quickCleanStatus = .cleaning

        Task {
            do {
                let result = try await CLIBridge.shared.clean(safeOnly: true)
                await MainActor.run {
                    self.lastCleanResult = result
                    self.safeReclaimableBytes = 0
                    self.quickCleanStatus = .success(result.freedFormatted)
                    SystemMonitor.shared.refresh()

                    DispatchQueue.main.asyncAfter(deadline: .now() + 3.5) {
                        if case .success = self.quickCleanStatus {
                            self.quickCleanStatus = .idle
                        }
                    }
                }
            } catch {
                await MainActor.run {
                    self.quickCleanStatus = .idle
                    self.errorMessage = error.localizedDescription
                }
            }
        }
    }

    /// 触发清理
    @MainActor
    public func startClean() {
        guard currentStage == .results else { return }
        guard !selectedItemIDs.isEmpty else { return }

        _ = SecurityBookmarkManager.shared.startAccessing()

        diskFreeBeforeClean = SystemMonitor.shared.diskFreeFormatted
        currentStage = .cleaning
        errorMessage = nil

        let targetIDs = Array(selectedItemIDs)

        Task {
            do {
                let result = try await CLIBridge.shared.clean(items: targetIDs)
                await MainActor.run {
                    SystemMonitor.shared.refresh()
                    self.diskFreeAfterClean = SystemMonitor.shared.diskFreeFormatted
                    self.lastCleanResult = result
                    withAnimation(.spring(response: 0.5, dampingFraction: 0.7)) {
                        self.currentStage = .cleaned
                    }
                }
            } catch {
                await MainActor.run {
                    self.errorMessage = error.localizedDescription
                    self.currentStage = .results
                }
            }
        }
    }

    /// 重置状态返回主页
    @MainActor
    public func resetToIdle() {
        withAnimation {
            self.currentStage = .idle
            self.scanReport = nil
            self.lastCleanResult = nil
            self.selectedItemIDs.removeAll()
            self.searchQuery = ""
        }
    }

    // MARK: - 选中逻辑
    public func toggleItemSelection(_ item: Item) {
        if selectedItemIDs.contains(item.id) {
            selectedItemIDs.remove(item.id)
        } else {
            selectedItemIDs.insert(item.id)
        }
    }

    public func toggleGroupSelection(_ group: Group) {
        let groupItemIDs = group.items.filter { !$0.isProtected }.map { $0.id }
        let allSelected = groupItemIDs.allSatisfy { selectedItemIDs.contains($0) }

        if allSelected {
            for id in groupItemIDs {
                selectedItemIDs.remove(id)
            }
        } else {
            for id in groupItemIDs {
                selectedItemIDs.insert(id)
            }
        }
    }

    public func isGroupAllSelected(_ group: Group) -> Bool {
        let validItems = group.items.filter { !$0.isProtected }
        if validItems.isEmpty { return false }
        return validItems.allSatisfy { selectedItemIDs.contains($0.id) }
    }

    public func isGroupPartiallySelected(_ group: Group) -> Bool {
        let validItems = group.items.filter { !$0.isProtected }
        let selectedCount = validItems.filter { selectedItemIDs.contains($0.id) }.count
        return selectedCount > 0 && selectedCount < validItems.count
    }

    public func selectAll() {
        guard let report = scanReport else { return }
        var ids = Set<String>()
        for grp in report.groups {
            for item in grp.items where !item.isProtected {
                ids.insert(item.id)
            }
        }
        selectedItemIDs = ids
    }

    public func deselectAll() {
        selectedItemIDs.removeAll()
    }

    public func selectSafeOnly() {
        guard let report = scanReport else { return }
        var ids = Set<String>()
        for grp in report.groups {
            for item in grp.items where !item.isProtected && item.risk == .safe && item.category == .systemCache {
                ids.insert(item.id)
            }
        }
        selectedItemIDs = ids
    }

    public var selectedTotalBytes: Int64 {
        guard let report = scanReport else { return 0 }
        var total: Int64 = 0
        for grp in report.groups {
            for item in grp.items where selectedItemIDs.contains(item.id) {
                total += item.sizeBytes
            }
        }
        return total
    }

    public var selectedTotalFormatted: String {
        ByteFormatter.format(selectedTotalBytes)
    }

    // MARK: - 扫描动画模拟
    private func startScanAnimationSimulation() {
        let samplePaths = [
            "~/Library/Developer/Xcode/DerivedData",
            "~/Library/Caches/Homebrew",
            "~/.docker/daemon.json",
            "~/Projects/workspace/target",
            "~/Projects/frontend/node_modules",
            "~/Library/Developer/CoreSimulator/Devices",
            "~/.cargo/registry/cache",
            "~/.gradle/caches",
            "~/.Trash",
            "~/Library/Logs"
        ]

        let samplePhases = [
            "正在探测 Docker 与 OrbStack 容器资源...",
            "正在分析 Xcode 派生数据与 iOS 模拟器...",
            "正在扫描本地代码工作区与构建输出...",
            "正在统计 Homebrew, Cargo, npm 缓存...",
            "正在排查系统废纸篓与更新临时文件..."
        ]

        var counter = 0
        scanSimulationTimer = Timer.publish(every: 0.45, on: .main, in: .common)
            .autoconnect()
            .sink { [weak self] _ in
                guard let self = self else { return }
                counter += 1
                self.scanningCurrentPath = samplePaths[counter % samplePaths.count]
                self.scanPhaseText = samplePhases[(counter / 2) % samplePhases.count]
                if self.scanProgress < 0.90 {
                    self.scanProgress += 0.05
                }
            }
    }

    private func stopScanAnimationSimulation() {
        scanSimulationTimer?.cancel()
        scanSimulationTimer = nil
    }
}
