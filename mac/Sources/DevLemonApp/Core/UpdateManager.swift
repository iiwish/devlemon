import Foundation
import AppKit
import Combine

#if !APP_STORE && canImport(Sparkle)
import Sparkle
#endif

// MARK: - 软件自动更新管理器 (双轨架构：Direct 版基于 Sparkle 2.0，App Store 版由 MAS 接管)
public final class UpdateManager: ObservableObject {
    public static let shared = UpdateManager()

    @Published public var isAppStoreBuild: Bool = false
    @Published public var canCheckForUpdates: Bool = true
    @Published public var automaticallyChecksForUpdates: Bool = true
    @Published public var automaticallyDownloadsUpdates: Bool = false

    #if !APP_STORE && canImport(Sparkle)
    public let updaterController: SPUStandardUpdaterController?
    #endif

    private init() {
        #if !APP_STORE && canImport(Sparkle)
        let controller = SPUStandardUpdaterController(
            startingUpdater: true,
            updaterDelegate: nil,
            userDriverDelegate: nil
        )
        self.updaterController = controller
        self.automaticallyChecksForUpdates = controller.updater.automaticallyChecksForUpdates
        self.automaticallyDownloadsUpdates = controller.updater.automaticallyDownloadsUpdates
        self.isAppStoreBuild = false

        controller.updater.publisher(for: \.canCheckForUpdates)
            .assign(to: &$canCheckForUpdates)
        #else
        self.isAppStoreBuild = true
        self.canCheckForUpdates = false
        #endif
    }

    public func checkForUpdates() {
        #if !APP_STORE && canImport(Sparkle)
        updaterController?.checkForUpdates(nil)
        #else
        // 在 App Store 版中打开 App Store 产品页 (预留 MAS URL)
        if let url = URL(string: "macappstore://apps.apple.com/app/devlemon/id000000000") {
            NSWorkspace.shared.open(url)
        }
        #endif
    }

    public func setAutomaticallyChecksForUpdates(_ enable: Bool) {
        self.automaticallyChecksForUpdates = enable
        #if !APP_STORE && canImport(Sparkle)
        updaterController?.updater.automaticallyChecksForUpdates = enable
        #endif
    }

    public func setAutomaticallyDownloadsUpdates(_ enable: Bool) {
        self.automaticallyDownloadsUpdates = enable
        #if !APP_STORE && canImport(Sparkle)
        updaterController?.updater.automaticallyDownloadsUpdates = enable
        #endif
    }
}
