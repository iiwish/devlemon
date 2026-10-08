import AppKit
import SwiftUI

public final class SettingsWindowController: NSObject, NSWindowDelegate {
    public static let shared = SettingsWindowController()

    private var window: NSWindow?

    private override init() {
        super.init()
    }

    public var isWindowVisible: Bool {
        guard let window = window else { return false }
        return window.isVisible || window.isMiniaturized
    }

    public func show() {
        WindowManager.shared.setActivationPolicy(.regular)
        NSApplication.shared.activate(ignoringOtherApps: true)

        if let existing = window {
            existing.makeKeyAndOrderFront(nil)
            existing.orderFrontRegardless()
            WindowManager.shared.updateDockVisibility()
            return
        }

        // 创建原生专属偏好设置窗口
        let hostingController = NSHostingController(rootView: SettingsView())
        let newWindow = NSWindow(contentViewController: hostingController)
        newWindow.title = "偏好设置"
        newWindow.styleMask = [.titled, .closable]
        newWindow.isReleasedWhenClosed = false
        newWindow.delegate = self
        newWindow.minSize = NSSize(width: 560, height: 480)
        newWindow.setContentSize(NSSize(width: 560, height: 530))
        newWindow.center()
        newWindow.titleVisibility = .visible
        newWindow.titlebarAppearsTransparent = false

        self.window = newWindow
        newWindow.makeKeyAndOrderFront(nil)
        newWindow.orderFrontRegardless()
        WindowManager.shared.updateDockVisibility()
    }

    public func close() {
        window?.orderOut(nil)
        DispatchQueue.main.async {
            WindowManager.shared.updateDockVisibility()
        }
    }

    public func windowWillClose(_ notification: Notification) {
        DispatchQueue.main.async {
            WindowManager.shared.updateDockVisibility()
        }
    }
}
