import Foundation

public enum CLIBridgeError: LocalizedError {
    case binaryNotFound
    case executionFailed(String)
    case decodingFailed(Error)

    public var errorDescription: String? {
        switch self {
        case .binaryNotFound:
            return "未找到 devlemon 核心引擎二进制可执行文件"
        case .executionFailed(let msg):
            return "执行核心引擎命令失败: \(msg)"
        case .decodingFailed(let err):
            return "解析数据失败: \(err.localizedDescription)"
        }
    }
}

public final class CLIBridge {
    public static let shared = CLIBridge()

    private init() {}

    /// 寻找 devlemon 引擎可执行文件的绝对路径
    public func resolveBinaryPath() -> String? {
        // 1. App Bundle Resources
        if let bundlePath = Bundle.main.path(forResource: "devlemon", ofType: nil) {
            if FileManager.default.isExecutableFile(atPath: bundlePath) {
                return bundlePath
            }
        }

        // 2. 检查应用同级目录或父级目录（开发运行状态）
        let candidates = [
            "/Users/iiwish/self/devlemon/devlemon",
            "/usr/local/bin/devlemon",
            "/opt/homebrew/bin/devlemon"
        ]

        for path in candidates {
            if FileManager.default.isExecutableFile(atPath: path) {
                return path
            }
        }

        // 3. 环境变量 PATH 查找
        if let pathEnv = ProcessInfo.processInfo.environment["PATH"] {
            for dir in pathEnv.split(separator: ":") {
                let candidate = String(dir) + "/devlemon"
                if FileManager.default.isExecutableFile(atPath: candidate) {
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

        var arguments = ["scan", "--json"]
        if safeOnly {
            arguments.append("--safe")
        }
        if !workspaces.isEmpty {
            arguments.append(contentsOf: ["--workspace", workspaces.joined(separator: ",")])
        }

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

        var arguments = ["clean", "--json"]
        if let items = items, !items.isEmpty {
            arguments.append(contentsOf: ["--items", items.joined(separator: ",")])
        } else if safeOnly {
            arguments.append("--safe")
        }

        if dryRun {
            arguments.append("--dry-run")
        }

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

                do {
                    try process.run()
                    let data = stdoutPipe.fileHandleForReading.readDataToEndOfFile()
                    let errData = stderrPipe.fileHandleForReading.readDataToEndOfFile()
                    process.waitUntilExit()

                    if process.terminationStatus == 0 {
                        continuation.resume(returning: data)
                    } else {
                        let errStr = String(data: errData, encoding: .utf8) ?? "Unknown error"
                        continuation.resume(throwing: CLIBridgeError.executionFailed(errStr))
                    }
                } catch {
                    continuation.resume(throwing: CLIBridgeError.executionFailed(error.localizedDescription))
                }
            }
        }
    }
}
