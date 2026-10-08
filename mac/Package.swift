// swift-tools-version: 5.9
import PackageDescription
import Foundation

// 双轨构建检测：通过环境变量 DEVLEMON_BUILD_TARGET 切换直装版 (Direct) 与商店版 (App Store)
let isAppStore = ProcessInfo.processInfo.environment["DEVLEMON_BUILD_TARGET"] == "appstore"

var packageDependencies: [Package.Dependency] = []
var targetDependencies: [Target.Dependency] = []
var swiftSettings: [SwiftSetting] = []

if isAppStore {
    swiftSettings.append(.define("APP_STORE"))
} else {
    packageDependencies.append(.package(url: "https://github.com/sparkle-project/Sparkle", from: "2.6.4"))
    targetDependencies.append(.product(name: "Sparkle", package: "Sparkle"))
}

let package = Package(
    name: "DevLemon",
    platforms: [
        .macOS(.v13)
    ],
    products: [
        .executable(
            name: "DevLemon",
            targets: ["DevLemonApp"]
        )
    ],
    dependencies: packageDependencies,
    targets: [
        .executableTarget(
            name: "DevLemonApp",
            dependencies: targetDependencies,
            path: "Sources/DevLemonApp",
            swiftSettings: swiftSettings
        )
    ]
)
