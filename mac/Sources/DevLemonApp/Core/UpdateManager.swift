import Foundation
import AppKit
import Sparkle
import Combine

// MARK: - Sparkle 2.0 自动更新管理器
public final class UpdateManager: ObservableObject {
    public static let shared = UpdateManager()

    @Published public var canCheckForUpdates: Bool = true

    @Published public var automaticallyChecksForUpdates: Bool = true {
        didSet {
            updaterController.updater.automaticallyChecksForUpdates = automaticallyChecksForUpdates
        }
    }

    @Published public var automaticallyDownloadsUpdates: Bool = false {
        didSet {
            updaterController.updater.automaticallyDownloadsUpdates = automaticallyDownloadsUpdates
        }
    }

    public let updaterController: SPUStandardUpdaterController

    private init() {
        self.updaterController = SPUStandardUpdaterController(
            startingUpdater: true,
            updaterDelegate: nil,
            userDriverDelegate: nil
        )

        self.automaticallyChecksForUpdates = updaterController.updater.automaticallyChecksForUpdates
        self.automaticallyDownloadsUpdates = updaterController.updater.automaticallyDownloadsUpdates

        updaterController.updater.publisher(for: \.canCheckForUpdates)
            .assign(to: &$canCheckForUpdates)
    }

    public func checkForUpdates() {
        updaterController.checkForUpdates(nil)
    }
}
