import Foundation
import AppKit
import Darwin

/// 单个应用的内存占用（主进程与其包内所有辅助进程之和）
public struct AppMemoryItem: Identifiable, Equatable {
    public let id: String          // 应用包路径
    public let name: String
    public let bundlePath: String
    public let pid: pid_t          // 主应用进程，用于退出
    public let footprintBytes: Int64
    public let processCount: Int

    public static func == (lhs: AppMemoryItem, rhs: AppMemoryItem) -> Bool {
        lhs.id == rhs.id && lhs.footprintBytes == rhs.footprintBytes && lhs.pid == rhs.pid
    }
}

/// 按应用聚合各进程内存（与活动监视器「内存」列同口径的 phys_footprint），支持退出应用与释放系统内存。
/// 只在面板展开时采样，收起后停止，不产生后台开销。
public final class AppMemoryMonitor: ObservableObject {
    public static let shared = AppMemoryMonitor()

    public enum QuitState: Equatable {
        case quitting             // 已请求正常退出，等待应用自行关闭（可能在询问是否保存）
        case unresponsive         // 超时仍未退出，可强制退出
    }

    public enum PurgeState: Equatable {
        case idle
        case running
        case done(freedBytes: Int64)
        case failed(String)
    }

    @Published public private(set) var items: [AppMemoryItem] = []
    /// 无法读取其他进程信息
    @Published public private(set) var isRestricted = false
    @Published public private(set) var quitStates: [String: QuitState] = [:]
    @Published public private(set) var purgeState: PurgeState = .idle

    /// 展示的应用数量上限
    public var limit = 8

    private var timer: Timer?
    private var activeObservers = 0
    private let sampleQueue = DispatchQueue(label: "devlemon.app-memory", qos: .utility)
    private var quitTimers: [String: Timer] = [:]

    /// 是否支持释放系统内存（依赖系统自带的 purge）
    public var canPurge: Bool {
        FileManager.default.isExecutableFile(atPath: "/usr/sbin/purge")
    }

    private init() {}

    // MARK: - 采样生命周期

    public func start() {
        activeObservers += 1
        guard timer == nil else { return }
        refresh()
        let t = Timer(timeInterval: 2.0, repeats: true) { [weak self] _ in self?.refresh() }
        RunLoop.main.add(t, forMode: .common)
        timer = t
    }

    public func stop() {
        activeObservers = max(0, activeObservers - 1)
        guard activeObservers == 0 else { return }
        timer?.invalidate()
        timer = nil
    }

    public func refresh() {
        // NSWorkspace 需在主线程读取；逐进程采样放到后台队列
        let apps = NSWorkspace.shared.runningApplications
        var appsByPath: [String: NSRunningApplication] = [:]
        for app in apps {
            if let url = app.bundleURL {
                appsByPath[Self.canonicalPath(url.path)] = app
            }
        }
        let ownPath = Self.canonicalPath(Bundle.main.bundlePath)
        let limit = self.limit

        sampleQueue.async { [weak self] in
            let (result, readable) = Self.sample(appsByPath: appsByPath, excluding: ownPath)
            let top = Array(result.prefix(limit))
            DispatchQueue.main.async {
                guard let self else { return }
                self.isRestricted = !readable
                if self.items != top {
                    self.items = top
                }
                // 已退出的应用清理其退出状态
                let alive = Set(appsByPath.keys)
                for key in self.quitStates.keys where !alive.contains(key) {
                    self.clearQuitState(key)
                }
            }
        }
    }

    // MARK: - 进程采样

    /// 枚举所有进程，按可执行文件所在的最外层 .app 聚合 phys_footprint。
    /// 只统计以 LaunchServices 应用身份运行的包（有对应 NSRunningApplication），并排除系统核心组件。
    static func sample(appsByPath: [String: NSRunningApplication], excluding ownPath: String) -> ([AppMemoryItem], Bool) {
        let count = proc_listallpids(nil, 0)
        guard count > 0 else { return ([], false) }
        var pids = [pid_t](repeating: 0, count: Int(count) + 64)
        let filled = pids.withUnsafeMutableBufferPointer { buf in
            proc_listallpids(buf.baseAddress, Int32(buf.count * MemoryLayout<pid_t>.size))
        }
        guard filled > 0 else { return ([], false) }

        let ownPID = ProcessInfo.processInfo.processIdentifier
        var totals: [String: (bytes: Int64, procs: Int)] = [:]
        var readOthers = false
        var pathBuf = [CChar](repeating: 0, count: Int(MAXPATHLEN) * 4)

        for pid in pids.prefix(Int(filled)) where pid > 0 {
            let len = proc_pidpath(pid, &pathBuf, UInt32(pathBuf.count))
            guard len > 0 else { continue }
            let exe = String(cString: pathBuf)
            guard let bundle = outerAppBundle(exe), bundle != ownPath, isUserFacing(bundle) else { continue }
            guard appsByPath[bundle] != nil else { continue }
            guard let footprint = physFootprint(pid) else { continue }
            if pid != ownPID { readOthers = true }
            totals[bundle, default: (0, 0)].bytes += footprint
            totals[bundle, default: (0, 0)].procs += 1
        }

        let items = totals.compactMap { path, total -> AppMemoryItem? in
            guard let app = appsByPath[path] else { return nil }
            let name = app.localizedName ?? (path as NSString).lastPathComponent.replacingOccurrences(of: ".app", with: "")
            return AppMemoryItem(id: path, name: name, bundlePath: path, pid: app.processIdentifier,
                                 footprintBytes: total.bytes, processCount: total.procs)
        }
        .sorted { $0.footprintBytes > $1.footprintBytes }
        return (items, readOthers)
    }

    /// 与内核 proc_pidpath 一致的真实路径。不能用 standardizedFileURL：它会把 /private/var、/private/tmp
    /// 缩写为 /var、/tmp，导致从磁盘映像直接打开（App Translocation）的应用匹配不上
    static func canonicalPath(_ path: String) -> String {
        guard let resolved = realpath(path, nil) else { return path }
        defer { free(resolved) }
        return String(cString: resolved)
    }

    /// 可执行文件路径中最外层的 .app 包，Helper 进程归属到主应用
    static func outerAppBundle(_ exe: String) -> String? {
        guard let range = exe.range(of: ".app/") else { return nil }
        return String(exe[..<range.lowerBound]) + ".app"
    }

    /// 排除系统核心组件（程序坞、访达、登录窗口等），它们退出会影响系统使用或被立即重启
    static func isUserFacing(_ bundle: String) -> Bool {
        !(bundle.hasPrefix("/System/Library/") || bundle.hasPrefix("/usr/") || bundle.hasPrefix("/Library/Apple/"))
    }

    static func physFootprint(_ pid: pid_t) -> Int64? {
        var info = rusage_info_v4()
        let ret = withUnsafeMutablePointer(to: &info) { ptr in
            ptr.withMemoryRebound(to: rusage_info_t?.self, capacity: 1) {
                proc_pid_rusage(pid, RUSAGE_INFO_V4, $0)
            }
        }
        guard ret == 0 else { return nil }
        return Int64(info.ri_phys_footprint)
    }

    // MARK: - 退出应用

    /// 正常退出：等同于在应用中按 ⌘Q，应用可询问是否保存；3 秒后仍未退出则允许强制退出
    public func quit(_ item: AppMemoryItem) {
        guard let app = NSRunningApplication(processIdentifier: item.pid) else { return }
        if !app.terminate() {
            quitStates[item.id] = .unresponsive
            return
        }
        quitStates[item.id] = .quitting
        quitTimers[item.id]?.invalidate()
        let timer = Timer(timeInterval: 3.0, repeats: false) { [weak self] _ in
            guard let self else { return }
            if app.isTerminated {
                self.clearQuitState(item.id)
            } else {
                self.quitStates[item.id] = .unresponsive
            }
            self.refresh()
        }
        RunLoop.main.add(timer, forMode: .common)
        quitTimers[item.id] = timer
    }

    /// 强制退出：未保存的内容会丢失，仅在正常退出无响应后提供
    public func forceQuit(_ item: AppMemoryItem) {
        guard let app = NSRunningApplication(processIdentifier: item.pid) else { return }
        _ = app.forceTerminate()
        clearQuitState(item.id)
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.5) { [weak self] in self?.refresh() }
    }

    private func clearQuitState(_ id: String) {
        quitTimers[id]?.invalidate()
        quitTimers[id] = nil
        quitStates[id] = nil
    }

    // MARK: - 释放系统内存

    /// 运行系统自带的 purge 清空可回收的文件缓存（需要管理员密码），对比前后可用内存得出释放量
    public func purge() {
        guard canPurge, purgeState != .running else { return }
        purgeState = .running
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            let before = Self.availableMemoryBytes()
            var error: NSDictionary?
            let script = NSAppleScript(source: "do shell script \"/usr/sbin/purge\" with administrator privileges")
            script?.executeAndReturnError(&error)
            let after = Self.availableMemoryBytes()
            DispatchQueue.main.async {
                guard let self else { return }
                if let error {
                    // -128：用户取消了授权对话框
                    let code = error[NSAppleScript.errorNumber] as? Int ?? 0
                    self.purgeState = code == -128 ? .idle : .failed(error[NSAppleScript.errorMessage] as? String ?? "释放失败")
                } else {
                    self.purgeState = .done(freedBytes: max(0, after - before))
                }
                SystemMonitor.shared.refresh()
                self.refresh()
                DispatchQueue.main.asyncAfter(deadline: .now() + 4) {
                    if case .done = self.purgeState { self.purgeState = .idle }
                    if case .failed = self.purgeState { self.purgeState = .idle }
                }
            }
        }
    }

    /// 空闲内存（空闲页 + 预读页）：purge 会把文件缓存归还为空闲页，前后差值即释放量
    static func availableMemoryBytes() -> Int64 {
        var size = mach_msg_type_number_t(MemoryLayout<vm_statistics64_data_t>.size / MemoryLayout<integer_t>.size)
        var stats = vm_statistics64()
        let ret = withUnsafeMutablePointer(to: &stats) {
            $0.withMemoryRebound(to: integer_t.self, capacity: Int(size)) {
                host_statistics64(mach_host_self(), HOST_VM_INFO64, $0, &size)
            }
        }
        guard ret == KERN_SUCCESS else { return 0 }
        let page = Int64(vm_kernel_page_size)
        return (Int64(stats.free_count) + Int64(stats.speculative_count)) * page
    }
}
