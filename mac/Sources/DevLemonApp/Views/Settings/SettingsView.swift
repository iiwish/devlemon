import SwiftUI
import ServiceManagement

public struct SettingsView: View {
    @AppStorage("showSpeedInMenuBar") private var showSpeedInMenuBar: Bool = false
    @AppStorage("autoScanOnLaunch") private var autoScanOnLaunch: Bool = false
    @State private var launchAtLogin: Bool = false
    @State private var customWorkspaces: String = ""

    public init() {}

    public var body: some View {
        Form {
            Section(header: Text("通用").font(.headline)) {
                Toggle("开机自动启动 DevLemon Lite", isOn: $launchAtLogin)
                    .onChange(of: launchAtLogin) { newValue in
                        updateLaunchAtLogin(newValue)
                    }

                Toggle("启动应用后自动执行扫描", isOn: $autoScanOnLaunch)
            }

            Section(header: Text("菜单栏托盘显示").font(.headline)) {
                Toggle("在菜单栏实时显示上传/下载网速", isOn: $showSpeedInMenuBar)
            }

            Section(header: Text("代码工作区扫描路径").font(.headline)) {
                VStack(alignment: .leading, spacing: 6) {
                    Text("默认会自动探测当前用户目录下的 Projects, Developer, Code, Workspace, Desktop 以及当前目录。")
                        .font(.system(size: 11))
                        .foregroundColor(.secondary)

                    TextField("自定义额外路径 (英文逗号隔开，例如 ~/Workspace, /Volumes/Data/repo)", text: $customWorkspaces)
                        .textFieldStyle(.roundedBorder)
                }
            }

            Section(header: Text("关于").font(.headline)) {
                HStack {
                    Text("版本")
                    Spacer()
                    Text("v0.2.4 (Native macOS)")
                        .foregroundColor(.secondary)
                }
                HStack {
                    Text("核心引擎")
                    Spacer()
                    Text("DevLemon Go Core Engine")
                        .foregroundColor(.secondary)
                }
            }
        }
        .formStyle(.grouped)
        .frame(width: 480, height: 340)
        .onAppear {
            checkLaunchAtLoginStatus()
        }
    }

    private func checkLaunchAtLoginStatus() {
        if #available(macOS 13.0, *) {
            launchAtLogin = SMAppService.mainApp.status == .enabled
        }
    }

    private func updateLaunchAtLogin(_ enable: Bool) {
        if #available(macOS 13.0, *) {
            do {
                if enable {
                    try SMAppService.mainApp.register()
                } else {
                    try SMAppService.mainApp.unregister()
                }
            } catch {
                print("Failed to toggle launch at login: \(error)")
            }
        }
    }
}
