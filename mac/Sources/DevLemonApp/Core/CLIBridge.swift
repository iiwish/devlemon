import Foundation

public enum CLIBridgeError: LocalizedError {
    case binaryNotFound
    case executionFailed(String)
    case decodingFailed(Error)
    case cancelled

    public var errorDescription: String? {
        switch self {
        case .binaryNotFound:
            return "未找到 devlemon 核心引擎二进制可执行文件"
        case .executionFailed(let msg):
            return "执行核心引擎命令失败: \(msg)"
        case .decodingFailed(let err):
            return "解析数据失败: \(err.localizedDescription)"
        case .cancelled:
            return "操作已由用户手动停止"
        }
    }
}

public final class CLIBridge: @unchecked Sendable {
    public static let shared = CLIBridge()

    private var activeProcess: Process?
    private let processLock = NSLock()

    private init() {}

    /// 取消当前正在执行的进程任务
    public func cancelActiveOperation() {
        processLock.lock()
        defer { processLock.unlock() }
        if let proc = activeProcess, proc.isRunning {
            proc.terminate()
        }
        activeProcess = nil
    }

    /// 寻找 devlemon 引擎可执行文件的绝对路径
    public func resolveBinaryPath() -> String? {
        let execURL = Bundle.main.executableURL

        // 1. App Bundle Contents/Helpers (规范标准路径，规避与 Contents/MacOS/DevLemon 的大小写冲突)
        if let bundleURL = Bundle.main.bundleURL as URL? {
            let helperInHelpers = bundleURL.appendingPathComponent("Contents/Helpers/devlemon").path
            if FileManager.default.isExecutableFile(atPath: helperInHelpers) {
                return helperInHelpers
            }
        }
        if let execURL = execURL {
            let macosDir = execURL.deletingLastPathComponent()
            let helpersDir = macosDir.deletingLastPathComponent().appendingPathComponent("Helpers")
            let helperInHelpers = helpersDir.appendingPathComponent("devlemon").path
            if FileManager.default.isExecutableFile(atPath: helperInHelpers) {
                return helperInHelpers
            }
        }

        // 2. App Bundle Resources
        if let bundlePath = Bundle.main.path(forResource: "devlemon", ofType: nil) {
            if FileManager.default.isExecutableFile(atPath: bundlePath) {
                return bundlePath
            }
        }

        // 3. 检查系统标准路径
        let standardPaths = [
            "/usr/local/bin/devlemon",
            "/opt/homebrew/bin/devlemon"
        ]

        for path in standardPaths {
            if FileManager.default.isExecutableFile(atPath: path) {
                return path
            }
        }

        // 4. 环境变量 PATH 查找 (排除当前主程序自身)
        if let pathEnv = ProcessInfo.processInfo.environment["PATH"] {
            for dir in pathEnv.split(separator: ":") {
                let candidate = String(dir) + "/devlemon"
                if FileManager.default.isExecutableFile(atPath: candidate) {
                    if let execPath = execURL?.path, (candidate as NSString).standardizingPath == (execPath as NSString).standardizingPath {
                        continue
                    }
                    return candidate
                }
            }
        }

        return nil
    }

    /// 执行扫描任务
    public func scan(safeOnly: Bool = false, workspaces: [String] = []) async throws -> ScanReport {
        guard let binaryPath = resolveBinaryPath() else {
            throw CLIBridgeError.binaryNotFound
        }

        let isSandboxed = SecurityBookmarkManager.shared.isSandboxed
        var targetWorkspaces = workspaces
        if isSandboxed && targetWorkspaces.isEmpty {
            targetWorkspaces = SecurityBookmarkManager.shared.authorizedWorkspaces
        }

        var arguments = ["scan", "--json"]
        if isSandboxed {
            arguments.append("--sandbox")
        }
        if safeOnly {
            arguments.append("--safe")
        }
        if !targetWorkspaces.isEmpty {
            arguments.append(contentsOf: ["--workspace", targetWorkspaces.joined(separator: ",")])
        }

        SecurityBookmarkManager.shared.startAccessing()
        defer { SecurityBookmarkManager.shared.stopAccessing() }

        let outputData = try await runProcess(binaryPath: binaryPath, arguments: arguments)
        do {
            let decoder = JSONDecoder()
            decoder.dateDecodingStrategy = .iso8601
            let report = try decoder.decode(ScanReport.self, from: outputData)
            return report
        } catch {
            let rawStr = String(data: outputData, encoding: .utf8) ?? ""
            print("Decode ScanReport failed. Raw output: \(rawStr)")
            throw CLIBridgeError.decodingFailed(error)
        }
    }

    /// 执行清理任务
    public func clean(items: [String]? = nil, safeOnly: Bool = false, dryRun: Bool = false) async throws -> CleanResult {
        guard let binaryPath = resolveBinaryPath() else {
            throw CLIBridgeError.binaryNotFound
        }

        let isSandboxed = SecurityBookmarkManager.shared.isSandboxed
        var arguments = ["clean", "--json"]
        if isSandboxed {
            arguments.append("--sandbox")
            let ws = SecurityBookmarkManager.shared.authorizedWorkspaces
            if !ws.isEmpty {
                arguments.append(contentsOf: ["--workspace", ws.joined(separator: ",")])
            }
        }
        if let items = items, !items.isEmpty {
            arguments.append(contentsOf: ["--items", items.joined(separator: ",")])
        } else if safeOnly {
            arguments.append("--safe")
        }

        if dryRun {
            arguments.append("--dry-run")
        }

        SecurityBookmarkManager.shared.startAccessing()
        defer { SecurityBookmarkManager.shared.stopAccessing() }

        let outputData = try await runProcess(binaryPath: binaryPath, arguments: arguments)
        do {
            let decoder = JSONDecoder()
            let result = try decoder.decode(CleanResult.self, from: outputData)
            return result
        } catch {
            let rawStr = String(data: outputData, encoding: .utf8) ?? ""
            print("Decode CleanResult failed. Raw output: \(rawStr)")
            throw CLIBridgeError.decodingFailed(error)
        }
    }

    private func runProcess(binaryPath: String, arguments: [String]) async throws -> Data {
        try await withCheckedThrowingContinuation { continuation in
            DispatchQueue.global(qos: .userInitiated).async {
                let process = Process()
                process.executableURL = URL(fileURLWithPath: binaryPath)
                process.arguments = arguments

                var env = ProcessInfo.processInfo.environment
                if !SecurityBookmarkManager.shared.authorizedPath.isEmpty {
                    env["HOME"] = SecurityBookmarkManager.shared.authorizedPath
                } else {
                    env["HOME"] = SecurityBookmarkManager.shared.realHomeURL.path
                }
                process.environment = env

                let stdoutPipe = Pipe()
                let stderrPipe = Pipe()
                process.standardOutput = stdoutPipe
                process.standardError = stderrPipe

                self.processLock.lock()
                self.activeProcess = process
                self.processLock.unlock()

                var wasTerminatedByUser = false

                do {
                    try process.run()
                    let data = stdoutPipe.fileHandleForReading.readDataToEndOfFile()
                    let errData = stderrPipe.fileHandleForReading.readDataToEndOfFile()
                    process.waitUntilExit()

                    self.processLock.lock()
                    if self.activeProcess == nil {
                        wasTerminatedByUser = true
                    }
                    self.activeProcess = nil
                    self.processLock.unlock()

                    if wasTerminatedByUser {
                        continuation.resume(throwing: CLIBridgeError.cancelled)
                    } else if process.terminationStatus == 0 {
                        continuation.resume(returning: data)
                    } else {
                        let errStr = String(data: errData, encoding: .utf8) ?? "Unknown error"
                        continuation.resume(throwing: CLIBridgeError.executionFailed(errStr))
                    }
                } catch {
                    self.processLock.lock()
                    self.activeProcess = nil
                    self.processLock.unlock()
                    continuation.resume(throwing: CLIBridgeError.executionFailed(error.localizedDescription))
                }
            }
        }
    }
}
