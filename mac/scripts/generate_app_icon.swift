import SwiftUI
import AppKit

// MARK: - DevLemon 官方 macOS App Icon 生成器
// 忠实传承经典 Apple 🍋 Emoji 质感，结合 macOS 现代设计规范 (Squircle 824x824, 环形光晕与内高光)

struct AppIconDesign: View {
    let size: CGFloat = 1024
    let tileSide: CGFloat = 824
    let cornerRadius: CGFloat = 185

    var body: some View {
        ZStack {
            Color.clear.frame(width: size, height: size)

            // 1. 标准 macOS 图标底部弥散投影
            RoundedRectangle(cornerRadius: cornerRadius, style: .continuous)
                .fill(Color.black.opacity(0.40))
                .frame(width: tileSide, height: tileSide)
                .blur(radius: 26)
                .offset(y: 22)

            // 2. 图标底板与内容
            ZStack {
                // 深黑石墨质感 (Dark Graphite Gradient - 专业开发者工具质感)
                LinearGradient(
                    colors: [
                        Color(red: 0.17, green: 0.18, blue: 0.21),
                        Color(red: 0.09, green: 0.10, blue: 0.12)
                    ],
                    startPoint: .top,
                    endPoint: .bottom
                )

                // 弥散金黄柠檬光晕 (Citrus Ambient Glow - 衬托柠檬立体质感)
                RadialGradient(
                    gradient: Gradient(colors: [
                        Color(red: 1.0, green: 0.82, blue: 0.18).opacity(0.38),
                        Color(red: 0.98, green: 0.65, blue: 0.08).opacity(0.14),
                        Color.clear
                    ]),
                    center: UnitPoint(x: 0.52, y: 0.52),
                    startRadius: 60,
                    endRadius: 400
                )
                .frame(width: tileSide, height: tileSide)

                // 边缘高光微轮廓 (Apple HIG 1.5pt Rim Light)
                RoundedRectangle(cornerRadius: cornerRadius, style: .continuous)
                    .stroke(
                        LinearGradient(
                            stops: [
                                .init(color: Color.white.opacity(0.24), location: 0.0),
                                .init(color: Color.white.opacity(0.08), location: 0.4),
                                .init(color: Color.white.opacity(0.02), location: 1.0)
                            ],
                            startPoint: .top,
                            endPoint: .bottom
                        ),
                        lineWidth: 2
                    )

                // 3. 经典 Apple 柠檬 (高保真渲染 + 完美光学居中 + 真实环境遮蔽投影)
                Text("🍋")
                    .font(.custom("Apple Color Emoji", size: 550))
                    .shadow(color: Color(red: 0.02, green: 0.02, blue: 0.04).opacity(0.55), radius: 24, x: 0, y: 16)
                    .offset(x: 14, y: 8)
            }
            .clipShape(RoundedRectangle(cornerRadius: cornerRadius, style: .continuous))
            .frame(width: tileSide, height: tileSide)
        }
        .frame(width: size, height: size)
    }
}

@MainActor
func main() {
    print("🍋 正在渲染 DevLemon 1024x1024 Master App Icon...")
    let renderer = ImageRenderer(content: AppIconDesign())
    renderer.scale = 1.0
    guard let nsImage = renderer.nsImage,
          let tiff = nsImage.tiffRepresentation,
          let rep = NSBitmapImageRep(data: tiff),
          let png = rep.representation(using: .png, properties: [:]) else {
        print("❌ 渲染 PNG 失败")
        exit(1)
    }

    let scriptURL = URL(fileURLWithPath: #file)
    let macDir = scriptURL.deletingLastPathComponent().deletingLastPathComponent()
    let resourcesDir = macDir.appendingPathComponent("Resources")
    try? FileManager.default.createDirectory(at: resourcesDir, withIntermediateDirectories: true)

    let masterPNGPath = resourcesDir.appendingPathComponent("AppIcon_1024.png").path
    let icnsPath = resourcesDir.appendingPathComponent("AppIcon.icns").path
    let tempIconsetDir = "/tmp/DevLemon_AppIcon.iconset"

    do {
        try png.write(to: URL(fileURLWithPath: masterPNGPath))
        print("✅ 已生成 1024x1024 原图: \(masterPNGPath)")
    } catch {
        print("❌ 写入文件失败: \(error)")
        exit(1)
    }

    // 生成各个分辨率的 iconset
    try? FileManager.default.removeItem(atPath: tempIconsetDir)
    try? FileManager.default.createDirectory(atPath: tempIconsetDir, withIntermediateDirectories: true)

    let sizes: [(String, Int)] = [
        ("icon_16x16.png", 16),
        ("icon_16x16@2x.png", 32),
        ("icon_32x32.png", 32),
        ("icon_32x32@2x.png", 64),
        ("icon_128x128.png", 128),
        ("icon_128x128@2x.png", 256),
        ("icon_256x256.png", 256),
        ("icon_256x256@2x.png", 512),
        ("icon_512x512.png", 512),
        ("icon_512x512@2x.png", 1024)
    ]

    for (name, px) in sizes {
        let dest = "\(tempIconsetDir)/\(name)"
        let p = Process()
        p.executableURL = URL(fileURLWithPath: "/usr/bin/sips")
        p.arguments = ["-z", "\(px)", "\(px)", masterPNGPath, "--out", dest]
        try? p.run()
        p.waitUntilExit()
    }

    // 编译为 AppIcon.icns
    let iconutil = Process()
    iconutil.executableURL = URL(fileURLWithPath: "/usr/bin/iconutil")
    iconutil.arguments = ["-c", "icns", tempIconsetDir, "-o", icnsPath]
    do {
        try iconutil.run()
        iconutil.waitUntilExit()
        if iconutil.terminationStatus == 0 {
            print("✨ 成功生成并编译 AppIcon.icns: \(icnsPath)")
        } else {
            print("❌ iconutil 执行失败，退出码: \(iconutil.terminationStatus)")
            exit(1)
        }
    } catch {
        print("❌ 启动 iconutil 失败: \(error)")
        exit(1)
    }
}

Task { @MainActor in
    main()
    exit(0)
}
RunLoop.main.run()
