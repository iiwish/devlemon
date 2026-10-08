import AppKit
import SwiftUI

public final class SettingsWindowController: NSObject, NSWindowDelegate {
    public static let shared = SettingsWindowController()

    private var window: NSWindow?

    private override init() {
        super.init()
    }

    public func show() {
        NSApplication.shared.activate(ignoringOtherApps: true)

        if let existing = window {
            existing.makeKeyAndOrderFront(nil)
            existing.orderFrontRegardless()
            return
        }

        // 创建原生专属偏好设置窗口
        let hostingController = NSHostingController(rootView: SettingsView())
        let newWindow = NSWindow(contentViewController: hostingController)
        newWindow.title = "偏好设置"
        newWindow.styleMask = [.titled, .closable]
        newWindow.isReleasedWhenClosed = false
        newWindow.delegate = self
        newWindow.center()
        newWindow.titleVisibility = .visible
        newWindow.titlebarAppearsTransparent = false

        self.window = newWindow
        newWindow.makeKeyAndOrderFront(nil)
        newWindow.orderFrontRegardless()
    }

    public func windowWillClose(_ notification: Notification) {
        // 保持引用，关闭时仅隐藏
    }
}
