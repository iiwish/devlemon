import SwiftUI
import AppKit

// MARK: - App 代理 (接管生命周期与统一调度)
final class AppDelegate: NSObject, NSApplicationDelegate {
    func applicationDidFinishLaunching(_ notification: Notification) {
        // 启动时直接打开主窗口
        WindowManager.shared.showMainWindow()
    }

    func applicationShouldHandleReopen(_ sender: NSApplication, hasVisibleWindows flag: Bool) -> Bool {
        // 用户在 Finder/Spotlight/Launchpad 再次点击应用图标时唤起主窗口
        WindowManager.shared.showMainWindow()
        return true
    }
}

// MARK: - 窗口与程序坞显隐全局管理器
public final class WindowManager: ObservableObject {
    public static let shared = WindowManager()

    public func showMainWindow() {
        MainWindowController.shared.show()
    }

    public func showSettingsWindow() {
        SettingsWindowController.shared.show()
    }

    public func setActivationPolicy(_ policy: NSApplication.ActivationPolicy) {
        if NSApplication.shared.activationPolicy() != policy {
            NSApplication.shared.setActivationPolicy(policy)
        }
    }

    // 核心逻辑：若主窗口与设置窗口均未打开，则隐藏程序坞图标 (accessory 策略)；否则显示 (regular 策略)
    public func updateDockVisibility() {
        let isMainOpen = MainWindowController.shared.isWindowVisible
        let isSettingsOpen = SettingsWindowController.shared.isWindowVisible
        let shouldShowInDock = isMainOpen || isSettingsOpen

        let targetPolicy: NSApplication.ActivationPolicy = shouldShowInDock ? .regular : .accessory

        if NSApplication.shared.activationPolicy() != targetPolicy {
            NSApplication.shared.setActivationPolicy(targetPolicy)
            if targetPolicy == .regular {
                NSApplication.shared.activate(ignoringOtherApps: true)
            }
        }
    }
}

// MARK: - 应用主入口
@main
struct DevLemonApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) var appDelegate
    @StateObject private var appState = AppState.shared
    @ObservedObject private var settings = AppSettings.shared

    var body: some Scene {
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
