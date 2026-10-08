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

    /// 应用启动时默认自动检测并安装/同步 CLI 软链接
    public func autoInstallIfNeeded() {
        guard let binaryPath = resolveSourceBinaryPath() else {
            print("⚠️ 未找到应用内置 devlemon 二进制，跳过自动链接")
            return
        }

        let home = FileManager.default.homeDirectoryForCurrentUser.path
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
    }

    /// 手动重新安装或修复
    @discardableResult
    public func reinstall() -> (success: Bool, message: String) {
        guard let binaryPath = resolveSourceBinaryPath() else {
            let msg = "未找到应用内置的 devlemon 引擎"
            self.lastMessage = msg
            return (false, msg)
        }

        let home = FileManager.default.homeDirectoryForCurrentUser.path
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
        let home = FileManager.default.homeDirectoryForCurrentUser.path
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
        if let bundlePath = Bundle.main.path(forResource: "devlemon", ofType: nil),
           FileManager.default.isExecutableFile(atPath: bundlePath) {
            return bundlePath
        }

        // 开发与调试目录备选
        let candidates = [
            "/Users/iiwish/self/devlemon/build/DevLemon.app/Contents/Resources/devlemon",
            "/Users/iiwish/self/devlemon/devlemon"
        ]
        for path in candidates {
            if FileManager.default.isExecutableFile(atPath: path) {
                return path
            }
        }
        return nil
    }
}
