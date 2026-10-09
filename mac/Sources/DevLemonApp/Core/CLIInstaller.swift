import Foundation
import AppKit

// MARK: - 终端命令行工具 (CLI) 自动安装与软链接管理器
public final class CLIInstaller: ObservableObject {
    public static let shared = CLIInstaller()

    @Published public var isInstalled: Bool = false
    @Published public var installedPath: String = ""
    @Published public var lastMessage: String?

    private init() {
        checkStatus()
    }

    /// 应用启动时默认自动检测并安装/同步 CLI 软链接 (非 App Store 直装版启用)
    public func autoInstallIfNeeded() {
        #if APP_STORE
        // Mac App Store 规范：严禁沙盒应用启动时向用户 PATH 静默植入软链接
        return
        #else
        guard let binaryPath = resolveSourceBinaryPath() else {
            print("⚠️ 未找到应用内置 devlemon 二进制，跳过自动链接")
            return
        }

        var shouldStopAccess = false
        if SecurityBookmarkManager.shared.isSandboxed {
            guard SecurityBookmarkManager.shared.hasAuthorization else {
                checkStatus()
                return
            }
            guard SecurityBookmarkManager.shared.startAccessing() else {
                checkStatus()
                return
            }
            shouldStopAccess = true
        }
        defer {
            if shouldStopAccess {
                SecurityBookmarkManager.shared.stopAccessing()
            }
        }

        let home = SecurityBookmarkManager.shared.realHomeURL.path
        let localBin = (home as NSString).appendingPathComponent(".local/bin")

        // 确保 ~/.local/bin 目录存在
        if !FileManager.default.fileExists(atPath: localBin) {
            try? FileManager.default.createDirectory(atPath: localBin, withIntermediateDirectories: true)
        }

        let targets = [
            (localBin as NSString).appendingPathComponent("devlemon"),
            (localBin as NSString).appendingPathComponent("dl")
        ]

        var needsUpdate = false
        for target in targets {
            if let dest = try? FileManager.default.destinationOfSymbolicLink(atPath: target) {
                if dest != binaryPath {
                    needsUpdate = true
                }
            } else {
                needsUpdate = true
            }
        }

        if needsUpdate {
            _ = installTo(targetDir: localBin, sourceBinary: binaryPath)
        }

        // 尝试检查 /usr/local/bin (若当前用户拥有写权限则同步安装)
        let usrLocalBin = "/usr/local/bin"
        if FileManager.default.isWritableFile(atPath: usrLocalBin) {
            _ = installTo(targetDir: usrLocalBin, sourceBinary: binaryPath)
        }

        checkStatus()
        #endif
    }

    /// 手动重新安装或修复
    @discardableResult
    public func reinstall() -> (success: Bool, message: String) {
        guard let binaryPath = resolveSourceBinaryPath() else {
            let msg = "未找到应用内置的 devlemon 引擎"
            self.lastMessage = msg
            return (false, msg)
        }

        var shouldStopAccess = false
        if SecurityBookmarkManager.shared.isSandboxed {
            guard SecurityBookmarkManager.shared.startAccessing() else {
                let msg = "沙盒环境下请先在设置中授权访问个人主目录"
                self.lastMessage = msg
                return (false, msg)
            }
            shouldStopAccess = true
        }
        defer {
            if shouldStopAccess {
                SecurityBookmarkManager.shared.stopAccessing()
            }
        }

        let home = SecurityBookmarkManager.shared.realHomeURL.path
        let localBin = (home as NSString).appendingPathComponent(".local/bin")
        if !FileManager.default.fileExists(atPath: localBin) {
            try? FileManager.default.createDirectory(atPath: localBin, withIntermediateDirectories: true)
        }

        let res = installTo(targetDir: localBin, sourceBinary: binaryPath)
        // 尝试 /usr/local/bin
        if FileManager.default.isWritableFile(atPath: "/usr/local/bin") {
            _ = installTo(targetDir: "/usr/local/bin", sourceBinary: binaryPath)
        }

        checkStatus()
        self.lastMessage = res ? "已成功链接至 ~/.local/bin" : "软链接创建失败"
        return (res, self.lastMessage ?? "")
    }

    public func checkStatus() {
        let home = SecurityBookmarkManager.shared.realHomeURL.path
        let localBin = (home as NSString).appendingPathComponent(".local/bin")
        let devlemonLink = (localBin as NSString).appendingPathComponent("devlemon")
        let dlLink = (localBin as NSString).appendingPathComponent("dl")

        if FileManager.default.fileExists(atPath: devlemonLink) || FileManager.default.fileExists(atPath: dlLink) {
            self.isInstalled = true
            self.installedPath = "~/.local/bin"
            return
        }

        if FileManager.default.fileExists(atPath: "/usr/local/bin/devlemon") || FileManager.default.fileExists(atPath: "/usr/local/bin/dl") {
            self.isInstalled = true
            self.installedPath = "/usr/local/bin"
            return
        }

        self.isInstalled = false
        self.installedPath = ""
    }

    private func installTo(targetDir: String, sourceBinary: String) -> Bool {
        let targets = ["devlemon", "dl"]
        var allOk = true

        for name in targets {
            let linkPath = (targetDir as NSString).appendingPathComponent(name)
            try? FileManager.default.removeItem(atPath: linkPath)
            do {
                try FileManager.default.createSymbolicLink(atPath: linkPath, withDestinationPath: sourceBinary)
            } catch {
                print("❌ 创建软链接失败 \(linkPath) -> \(sourceBinary): \(error)")
                allOk = false
            }
        }
        return allOk
    }

    private func resolveSourceBinaryPath() -> String? {
        return CLIBridge.shared.resolveBinaryPath()
    }
}
