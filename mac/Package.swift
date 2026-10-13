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
    dependencies: [
        .package(url: "https://github.com/sparkle-project/Sparkle", from: "2.6.4")
    ],
    targets: [
        .executableTarget(
            name: "DevLemonApp",
            dependencies: [
                .product(name: "Sparkle", package: "Sparkle")
            ],
            path: "Sources/DevLemonApp"
        )
    ]
)
