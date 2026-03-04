// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "SeocheonSDK",
    platforms: [
        .macOS(.v13),
        .iOS(.v16),
    ],
    products: [
        .library(name: "SeocheonSDK", targets: ["SeocheonSDK"]),
    ],
    dependencies: [
        .package(url: "https://github.com/GigaBitcoin/secp256k1.swift", exact: "0.18.0"),
    ],
    targets: [
        .target(
            name: "SeocheonSDK",
            dependencies: [
                .product(name: "secp256k1", package: "secp256k1.swift"),
            ],
            path: "Sources/SeocheonSDK"
        ),
        .testTarget(
            name: "SeocheonSDKTests",
            dependencies: ["SeocheonSDK"],
            path: "Tests/SeocheonSDKTests"
        ),
        .testTarget(
            name: "SeocheonSDKE2ETests",
            dependencies: ["SeocheonSDK"],
            path: "Tests/SeocheonSDKE2ETests"
        ),
    ]
)
