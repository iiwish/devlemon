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
    public var openSettingsAction: (() -> Void)?

    public func showMainWindow() {
        NSApplication.shared.activate(ignoringOtherApps: true)
        if let window = NSApplication.shared.windows.first(where: {
            let title = $0.title
            return (title.contains("DevLemon") || title.isEmpty) && !($0.className.contains("StatusBarWindow")) && !title.contains("偏好设置")
        }) {
            window.makeKeyAndOrderFront(nil)
        } else {
            openWindowAction?()
        }
    }

    public func showSettingsWindow() {
        NSApplication.shared.activate(ignoringOtherApps: true)
        if let window = NSApplication.shared.windows.first(where: {
            let title = $0.title
            return title.contains("偏好设置") || title.contains("Settings")
        }) {
            window.makeKeyAndOrderFront(nil)
        } else {
            openSettingsAction?()
        }
    }
}

@main
struct DevLemonApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) var appDelegate
    @StateObject private var appState = AppState.shared
    @ObservedObject private var settings = AppSettings.shared
    @Environment(\.openWindow) private var openWindow

    var body: some Scene {
        // 主窗口（启动时呈现）
        Window("DevLemon", id: "main") {
            MainWindowView()
                .frame(minWidth: 720, minHeight: 480)
                .onAppear {
                    WindowManager.shared.openWindowAction = {
                        openWindow(id: "main")
                    }
                    WindowManager.shared.openSettingsAction = {
                        openWindow(id: "preferences")
                    }
                }
        }
        .windowResizability(.contentSize)
        .windowStyle(.hiddenTitleBar)

        // 偏好设置窗口
        Window("偏好设置", id: "preferences") {
            SettingsView()
                .onAppear {
                    WindowManager.shared.openSettingsAction = {
                        openWindow(id: "preferences")
                    }
                }
        }
        .windowResizability(.contentSize)

        // 菜单栏托盘
        MenuBarExtra {
            MenuBarPopoverView {
                WindowManager.shared.showMainWindow()
            }
        } label: {
            MenuBarLabelView()
                .id(settings.menuBarUpdateId)
        }
        .menuBarExtraStyle(.window)

        // 原生 Cmd+, 偏好设置兜底
        Settings {
            SettingsView()
        }
    }
}
