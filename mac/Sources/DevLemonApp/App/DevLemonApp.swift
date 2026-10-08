import SwiftUI
import AppKit

final class AppDelegate: NSObject, NSApplicationDelegate {
    func applicationDidFinishLaunching(_ notification: Notification) {
        // 设置应用不在 Dock 强占位（也可以在 Info.plist 中设 LSUIElement）
        // 允许作为辅助/常驻托盘应用，同时在需要时呼出全功能主窗口
    }
}

public final class WindowManager: ObservableObject {
    public static let shared = WindowManager()
    public var openWindowAction: (() -> Void)?

    public func showMainWindow() {
        NSApplication.shared.activate(ignoringOtherApps: true)
        if let window = NSApplication.shared.windows.first(where: { $0.title.contains("DevLemon") && !($0.className.contains("StatusBarWindow")) }) {
            window.makeKeyAndOrderFront(nil)
        } else {
            openWindowAction?()
        }
    }
}

@main
struct DevLemonApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) var appDelegate
    @StateObject private var appState = AppState.shared
    @Environment(\.openWindow) private var openWindow

    var body: some Scene {
        // 菜单栏托盘
        MenuBarExtra {
            MenuBarPopoverView {
                WindowManager.shared.showMainWindow()
            }
        } label: {
            MenuBarLabelView()
        }
        .menuBarExtraStyle(.window)

        // 主窗口
        Window("DevLemon", id: "main") {
            MainWindowView()
                .frame(minWidth: 720, minHeight: 480)
                .onAppear {
                    WindowManager.shared.openWindowAction = {
                        openWindow(id: "main")
                    }
                }
        }
        .windowResizability(.contentSize)
        .windowStyle(.hiddenTitleBar)

        // 偏好设置
        Settings {
            SettingsView()
        }
    }
}
