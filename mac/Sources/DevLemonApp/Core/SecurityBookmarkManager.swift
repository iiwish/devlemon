import Foundation
import AppKit

// MARK: - 安全作用域书签管理器 (Mac App Store 沙盒授权核心机制)
public final class SecurityBookmarkManager: ObservableObject {
    public static let shared = SecurityBookmarkManager()

    private let bookmarkKey = "DevLemonAuthorizedHomeFolderBookmark"
    private var activeSecurityScopedURL: URL?

    @Published public var hasAuthorization: Bool = false
    @Published public var authorizedPath: String = ""

    /// 判断当前运行实例是否处于 macOS App 沙盒环境中
    public var isSandboxed: Bool {
        return ProcessInfo.processInfo.environment["APP_SANDBOX_CONTAINER_ID"] != nil
    }

    /// 获取系统真实的当前用户主目录（突破沙盒容器虚拟路径 ~/Library/Containers/...）
    public var realHomeURL: URL {
        if let pw = getpwuid(getuid()), let home = pw.pointee.pw_dir {
            return URL(fileURLWithPath: String(cString: home))
        }
        return FileManager.default.homeDirectoryForCurrentUser
    }

    private init() {
        checkAuthorization()
    }

    /// 检查并恢复已有的书签授权
    public func checkAuthorization() {
        // 非沙盒环境（如 GitHub 直装版）默认拥有完整用户空间访问权限
        guard isSandboxed else {
            self.hasAuthorization = true
            self.authorizedPath = realHomeURL.path
            return
        }

        guard let bookmarkData = UserDefaults.standard.data(forKey: bookmarkKey) else {
            self.hasAuthorization = false
            self.authorizedPath = ""
            return
        }

        var isStale = false
        do {
            let url = try URL(
                resolvingBookmarkData: bookmarkData,
                options: .withSecurityScope,
                relativeTo: nil,
                bookmarkDataIsStale: &isStale
            )

            if isStale {
                // 书签数据陈旧，尝试更新
                if let newBookmark = try? url.bookmarkData(
                    options: .withSecurityScope,
                    includingResourceValuesForKeys: nil,
                    relativeTo: nil
                ) {
                    UserDefaults.standard.set(newBookmark, forKey: bookmarkKey)
                }
            }

            self.hasAuthorization = true
            self.authorizedPath = url.path
        } catch {
            print("⚠️ 解析安全作用域书签失败: \(error)")
            self.hasAuthorization = false
            self.authorizedPath = ""
        }
    }

    /// 请求用户选取主文件夹以获得沙盒穿透授权
    public func requestAuthorization(completion: @escaping (Bool) -> Void) {
        guard isSandboxed else {
            self.hasAuthorization = true
            completion(true)
            return
        }

        DispatchQueue.main.async {
            let panel = NSOpenPanel()
            panel.title = "授权访问个人主目录"
            panel.message = "为合规扫描并释放项目缓存（如 node_modules、构建缓存等），请授权 DevLemon 访问您的个人主文件夹。"
            panel.prompt = "授权访问"
            panel.canChooseFiles = false
            panel.canChooseDirectories = true
            panel.canCreateDirectories = false
            panel.allowsMultipleSelection = false
            panel.directoryURL = self.realHomeURL

            panel.begin { response in
                if response == .OK, let selectedURL = panel.url {
                    do {
                        let bookmarkData = try selectedURL.bookmarkData(
                            options: .withSecurityScope,
                            includingResourceValuesForKeys: nil,
                            relativeTo: nil
                        )
                        UserDefaults.standard.set(bookmarkData, forKey: self.bookmarkKey)
                        self.hasAuthorization = true
                        self.authorizedPath = selectedURL.path
                        print("✅ 成功获取并持久化安全作用域书签: \(selectedURL.path)")
                        completion(true)
                    } catch {
                        print("❌ 创建安全作用域书签失败: \(error)")
                        completion(false)
                    }
                } else {
                    completion(false)
                }
            }
        }
    }

    /// 开始使用安全作用域资源（在扫描与清理执行前调用）
    @discardableResult
    public func startAccessing() -> Bool {
        guard isSandboxed else { return true }
        guard let bookmarkData = UserDefaults.standard.data(forKey: bookmarkKey) else { return false }

        var isStale = false
        guard let url = try? URL(
            resolvingBookmarkData: bookmarkData,
            options: .withSecurityScope,
            relativeTo: nil,
            bookmarkDataIsStale: &isStale
        ) else { return false }

        if url.startAccessingSecurityScopedResource() {
            self.activeSecurityScopedURL = url
            return true
        }
        return false
    }

    /// 停止访问安全作用域资源
    public func stopAccessing() {
        guard isSandboxed else { return }
        activeSecurityScopedURL?.stopAccessingSecurityScopedResource()
        activeSecurityScopedURL = nil
    }

    /// 清除授权
    public func clearAuthorization() {
        stopAccessing()
        UserDefaults.standard.removeObject(forKey: bookmarkKey)
        checkAuthorization()
    }
}
