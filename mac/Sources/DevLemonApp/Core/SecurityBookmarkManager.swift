import Foundation
import AppKit

// MARK: - 安全作用域书签管理器 (对标腾讯柠檬清理 Lite，Mac App Store 严苛沙盒规范)
public final class SecurityBookmarkManager: ObservableObject {
    public static let shared = SecurityBookmarkManager()

    // 缓存文件夹书签存储 Key (针对 ~/Library/Caches，精准合规)
    private let cachesBookmarkKey = "DevLemonAuthorizedCachesFolderBookmark"
    // 用户显式添加的代码工作区工程书签字典 [Path: BookmarkData]
    private let workspacesBookmarkKey = "DevLemonAuthorizedWorkspaceBookmarks"

    // 活跃的安全作用域 URLs 集合
    private var activeSecurityScopedURLs: [URL] = []

    // 缓存目录授权状态
    @Published public var hasCachesAuthorization: Bool = false
    @Published public var authorizedCachesPath: String = ""

    // 用户已授权的项目工作区路径列表
    @Published public var authorizedWorkspaces: [String] = []

    // 兼容历史属性
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

    /// 获取系统真实的 ~/Library/Caches 路径
    public var realCachesURL: URL {
        return realHomeURL.appendingPathComponent("Library/Caches")
    }

    private init() {
        checkAuthorization()
    }

    /// 检查并恢复已有的书签授权
    public func checkAuthorization() {
        // 非沙盒环境（如 GitHub 直装版）默认拥有完整用户空间访问权限
        guard isSandboxed else {
            self.hasCachesAuthorization = true
            self.authorizedCachesPath = realCachesURL.path
            self.hasAuthorization = true
            self.authorizedPath = realHomeURL.path
            return
        }

        // 1. 恢复缓存目录书签
        if let data = UserDefaults.standard.data(forKey: cachesBookmarkKey) {
            var isStale = false
            if let url = try? URL(resolvingBookmarkData: data, options: .withSecurityScope, relativeTo: nil, bookmarkDataIsStale: &isStale) {
                if isStale, let fresh = try? url.bookmarkData(options: .withSecurityScope, includingResourceValuesForKeys: nil, relativeTo: nil) {
                    UserDefaults.standard.set(fresh, forKey: cachesBookmarkKey)
                }
                self.hasCachesAuthorization = true
                self.authorizedCachesPath = url.path
                self.hasAuthorization = true
                self.authorizedPath = url.path
            } else {
                self.hasCachesAuthorization = false
                self.authorizedCachesPath = ""
            }
        } else {
            self.hasCachesAuthorization = false
            self.authorizedCachesPath = ""
            self.hasAuthorization = false
            self.authorizedPath = ""
        }

        // 2. 恢复工作区工程书签列表
        if let dict = UserDefaults.standard.dictionary(forKey: workspacesBookmarkKey) as? [String: Data] {
            var validPaths: [String] = []
            for (path, data) in dict {
                var isStale = false
                if let url = try? URL(resolvingBookmarkData: data, options: .withSecurityScope, relativeTo: nil, bookmarkDataIsStale: &isStale) {
                    validPaths.append(url.path)
                } else {
                    validPaths.append(path)
                }
            }
            self.authorizedWorkspaces = validPaths
        } else {
            self.authorizedWorkspaces = []
        }
    }

    /// 请求用户授权公共缓存目录 ~/Library/Caches (符合 Apple 最小必要权限原则)
    public func requestCachesAuthorization(completion: @escaping (Bool) -> Void) {
        guard isSandboxed else {
            self.hasCachesAuthorization = true
            completion(true)
            return
        }

        DispatchQueue.main.async {
            let panel = NSOpenPanel()
            panel.title = "授权访问缓存文件夹"
            panel.message = "为合规扫描并释放 Google Chrome、Safari、日常开发工具等临时网络与渲染缓存，请授权 DevLemon 访问缓存文件夹 (Library/Caches)。"
            panel.prompt = "授权缓存目录"
            panel.canChooseFiles = false
            panel.canChooseDirectories = true
            panel.canCreateDirectories = false
            panel.allowsMultipleSelection = false
            panel.directoryURL = self.realCachesURL

            panel.begin { response in
                if response == .OK, let selectedURL = panel.url {
                    do {
                        let bookmarkData = try selectedURL.bookmarkData(
                            options: .withSecurityScope,
                            includingResourceValuesForKeys: nil,
                            relativeTo: nil
                        )
                        UserDefaults.standard.set(bookmarkData, forKey: self.cachesBookmarkKey)
                        self.hasCachesAuthorization = true
                        self.authorizedCachesPath = selectedURL.path
                        self.hasAuthorization = true
                        self.authorizedPath = selectedURL.path
                        print("✅ 成功获取并持久化缓存目录书签: \(selectedURL.path)")
                        completion(true)
                    } catch {
                        print("❌ 创建缓存目录书签失败: \(error)")
                        completion(false)
                    }
                } else {
                    completion(false)
                }
            }
        }
    }

    /// 请求用户添加特定代码工作区工程目录
    public func requestAddWorkspace(completion: @escaping (Bool) -> Void) {
        guard isSandboxed else {
            completion(true)
            return
        }

        DispatchQueue.main.async {
            let panel = NSOpenPanel()
            panel.title = "添加代码工程工作区"
            panel.message = "选取您希望扫描并清理依赖或构建产物 (node_modules, target, .build) 的项目文件夹："
            panel.prompt = "添加项目"
            panel.canChooseFiles = false
            panel.canChooseDirectories = true
            panel.canCreateDirectories = false
            panel.allowsMultipleSelection = true
            panel.directoryURL = self.realHomeURL

            panel.begin { response in
                if response == .OK && !panel.urls.isEmpty {
                    var dict = UserDefaults.standard.dictionary(forKey: self.workspacesBookmarkKey) as? [String: Data] ?? [:]
                    for url in panel.urls {
                        if let bookmark = try? url.bookmarkData(options: .withSecurityScope, includingResourceValuesForKeys: nil, relativeTo: nil) {
                            dict[url.path] = bookmark
                        }
                    }
                    UserDefaults.standard.set(dict, forKey: self.workspacesBookmarkKey)
                    self.checkAuthorization()
                    completion(true)
                } else {
                    completion(false)
                }
            }
        }
    }

    /// 移除指定工程目录授权
    public func removeWorkspace(path: String) {
        var dict = UserDefaults.standard.dictionary(forKey: workspacesBookmarkKey) as? [String: Data] ?? [:]
        dict.removeValue(forKey: path)
        UserDefaults.standard.set(dict, forKey: workspacesBookmarkKey)
        checkAuthorization()
    }

    /// 兼容旧方法
    public func requestAuthorization(completion: @escaping (Bool) -> Void) {
        requestCachesAuthorization(completion: completion)
    }

    /// 开始使用全部已授权的安全作用域资源（在扫描与清理执行前调用）
    @discardableResult
    public func startAccessing() -> Bool {
        guard isSandboxed else { return true }

        var successCount = 0

        // 1. 激活 Caches 书签
        if let data = UserDefaults.standard.data(forKey: cachesBookmarkKey) {
            var isStale = false
            if let url = try? URL(resolvingBookmarkData: data, options: .withSecurityScope, relativeTo: nil, bookmarkDataIsStale: &isStale) {
                if url.startAccessingSecurityScopedResource() {
                    activeSecurityScopedURLs.append(url)
                    successCount += 1
                }
            }
        }

        // 2. 激活 Workspaces 书签
        if let dict = UserDefaults.standard.dictionary(forKey: workspacesBookmarkKey) as? [String: Data] {
            for (_, data) in dict {
                var isStale = false
                if let url = try? URL(resolvingBookmarkData: data, options: .withSecurityScope, relativeTo: nil, bookmarkDataIsStale: &isStale) {
                    if url.startAccessingSecurityScopedResource() {
                        activeSecurityScopedURLs.append(url)
                        successCount += 1
                    }
                }
            }
        }

        return successCount > 0
    }

    /// 停止访问全部安全作用域资源
    public func stopAccessing() {
        guard isSandboxed else { return }
        for url in activeSecurityScopedURLs {
            url.stopAccessingSecurityScopedResource()
        }
        activeSecurityScopedURLs.removeAll()
    }

    /// 清除全部授权
    public func clearAuthorization() {
        stopAccessing()
        UserDefaults.standard.removeObject(forKey: cachesBookmarkKey)
        UserDefaults.standard.removeObject(forKey: workspacesBookmarkKey)
        checkAuthorization()
    }
}
