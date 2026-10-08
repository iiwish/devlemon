import SwiftUI
import AppKit

final class AppDelegate: NSObject, NSApplicationDelegate {
    func applicationDidFinishLaunching(_ notification: Notification) {
        NSApplication.shared.setActivationPolicy(.regular)
    }

    func applicationShouldHandleReopen(_ sender: NSApplication, hasVisibleWindows flag: Bool) -> Bool {
        if !flag {
            WindowManager.shared.showMainWindow()
        }
        return true
    }
}

public final class WindowManager: ObservableObject {
    public static let shared = WindowManager()
    public var openWindowAction: (() -> Void)?

    public func showMainWindow() {
        NSApplication.shared.activate(ignoringOtherApps: true)
        if let window = NSApplication.shared.windows.first(where: {
            let title = $0.title
            return (title.contains("DevLemon") || title.isEmpty) && !($0.className.contains("StatusBarWindow"))
        }) {
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
        // 主窗口（放在第一个 Scene，启动时自动呈现）
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

        // 菜单栏托盘
        MenuBarExtra {
            MenuBarPopoverView {
                WindowManager.shared.showMainWindow()
            }
        } label: {
            MenuBarLabelView()
        }
        .menuBarExtraStyle(.window)

        // 偏好设置
        Settings {
            SettingsView()
        }
    }
}
