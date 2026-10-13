import Foundation
import AppKit
import Combine
import Sparkle

// MARK: - 软件自动更新管理器 (基于 Sparkle 2.0)
public final class UpdateManager: ObservableObject {
    public static let shared = UpdateManager()

    @Published public var canCheckForUpdates: Bool = true
    @Published public var automaticallyChecksForUpdates: Bool = true
    @Published public var automaticallyDownloadsUpdates: Bool = false

    public let updaterController: SPUStandardUpdaterController

    private init() {
        let controller = SPUStandardUpdaterController(
            startingUpdater: true,
            updaterDelegate: nil,
            userDriverDelegate: nil
        )
        self.updaterController = controller
        self.automaticallyChecksForUpdates = controller.updater.automaticallyChecksForUpdates
        self.automaticallyDownloadsUpdates = controller.updater.automaticallyDownloadsUpdates

        controller.updater.publisher(for: \.canCheckForUpdates)
            .assign(to: &$canCheckForUpdates)
    }

    public func checkForUpdates() {
        updaterController.checkForUpdates(nil)
    }

    public func setAutomaticallyChecksForUpdates(_ enable: Bool) {
        self.automaticallyChecksForUpdates = enable
        updaterController.updater.automaticallyChecksForUpdates = enable
    }

    public func setAutomaticallyDownloadsUpdates(_ enable: Bool) {
        self.automaticallyDownloadsUpdates = enable
        updaterController.updater.automaticallyDownloadsUpdates = enable
    }
}
