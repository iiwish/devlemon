import AppKit
import SwiftUI

// MARK: - 主窗口控制器 (AppKit 级别精准接管生命周期与程序坞交互)
public final class MainWindowController: NSObject, NSWindowDelegate {
    public static let shared = MainWindowController()

    private var window: NSWindow?

    private override init() {
        super.init()
    }

    public var isWindowVisible: Bool {
        guard let window = window else { return false }
        return window.isVisible || window.isMiniaturized
    }

    public func show() {
        // 先确保切换为常规应用策略以显示 Dock 图标
        WindowManager.shared.setActivationPolicy(.regular)
        NSApplication.shared.activate(ignoringOtherApps: true)

        if let existing = window {
            existing.makeKeyAndOrderFront(nil)
            existing.orderFrontRegardless()
            WindowManager.shared.updateDockVisibility()
            SystemMonitor.shared.isDetailViewActive = true
            return
        }

        // 创建原生专属主窗口
        let hostingController = NSHostingController(rootView: MainWindowView())
        let newWindow = NSWindow(contentViewController: hostingController)
        newWindow.title = "DevLemon"
        newWindow.styleMask = [.titled, .closable, .miniaturizable, .resizable, .fullSizeContentView]
        newWindow.isReleasedWhenClosed = false
        newWindow.delegate = self
        newWindow.minSize = NSSize(width: 720, height: 480)
        newWindow.setContentSize(NSSize(width: 760, height: 520))
        newWindow.center()
        newWindow.titleVisibility = .hidden
        newWindow.titlebarAppearsTransparent = true
        newWindow.setFrameAutosaveName("DevLemonMainWindow")

        self.window = newWindow
        newWindow.makeKeyAndOrderFront(nil)
        newWindow.orderFrontRegardless()
        WindowManager.shared.updateDockVisibility()
        SystemMonitor.shared.isDetailViewActive = true
    }

    public func close() {
        window?.orderOut(nil)
        SystemMonitor.shared.isDetailViewActive = false
        DispatchQueue.main.async {
            WindowManager.shared.updateDockVisibility()
        }
    }

    // 窗口即将/已经关闭时，延迟刷新程序坞显示状态
    public func windowWillClose(_ notification: Notification) {
        SystemMonitor.shared.isDetailViewActive = false
        DispatchQueue.main.async {
            WindowManager.shared.updateDockVisibility()
        }
    }

    public func windowDidMiniaturize(_ notification: Notification) {
        WindowManager.shared.updateDockVisibility()
    }

    public func windowDidDeminiaturize(_ notification: Notification) {
        WindowManager.shared.updateDockVisibility()
    }
}
