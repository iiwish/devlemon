import Foundation
import SwiftUI
import Combine

public final class AppSettings: ObservableObject {
    public static let shared = AppSettings()

    @Published public var launchAtLogin: Bool {
        didSet {
            UserDefaults.standard.set(launchAtLogin, forKey: "launchAtLogin")
            triggerUpdate()
        }
    }

    @Published public var showLogo: Bool {
        didSet {
            UserDefaults.standard.set(showLogo, forKey: "showLogo")
            triggerUpdate()
        }
    }

    @Published public var showNetwork: Bool {
        didSet {
            UserDefaults.standard.set(showNetwork, forKey: "showNetwork")
            triggerUpdate()
        }
    }

    @Published public var showMemory: Bool {
        didSet {
            UserDefaults.standard.set(showMemory, forKey: "showMemory")
            triggerUpdate()
        }
    }

    @Published public var showDisk: Bool {
        didSet {
            UserDefaults.standard.set(showDisk, forKey: "showDisk")
            triggerUpdate()
        }
    }

    @Published public var showCPU: Bool {
        didSet {
            UserDefaults.standard.set(showCPU, forKey: "showCPU")
            triggerUpdate()
        }
    }

    @Published public var menuBarUpdateId: UUID = UUID()

    private func triggerUpdate() {
        menuBarUpdateId = UUID()
        objectWillChange.send()
        Task { @MainActor in
            MenuBarImageProvider.shared.invalidateCache()
        }
    }

    private init() {
        self.launchAtLogin = UserDefaults.standard.bool(forKey: "launchAtLogin")
        // 首次启动默认：显示单色 Logo + 实时网速
        self.showLogo = UserDefaults.standard.object(forKey: "showLogo") != nil ? UserDefaults.standard.bool(forKey: "showLogo") : true
        self.showNetwork = UserDefaults.standard.object(forKey: "showNetwork") != nil ? UserDefaults.standard.bool(forKey: "showNetwork") : true
        self.showMemory = UserDefaults.standard.bool(forKey: "showMemory")
        self.showDisk = UserDefaults.standard.bool(forKey: "showDisk")
        self.showCPU = UserDefaults.standard.bool(forKey: "showCPU")
    }
}
