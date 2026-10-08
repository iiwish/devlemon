// swift-tools-version: 5.9
import PackageDescription

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
    dependencies: [],
    targets: [
        .executableTarget(
            name: "DevLemonApp",
            dependencies: [],
            path: "Sources/DevLemonApp"
        )
    ]
)
