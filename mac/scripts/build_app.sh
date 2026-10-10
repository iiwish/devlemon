#!/usr/bin/env bash
set -euo pipefail

TARGET="${1:-direct}" # direct (GitHub 直装版) 或 mas (Mac App Store 沙盒版)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MAC_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_DIR="$(cd "${MAC_DIR}/.." && pwd)"
OUTPUT_DIR="${REPO_DIR}/build"
VERSION="0.2.9"

if [[ "${TARGET}" == "mas" ]]; then
    APP_NAME="DevLemon-MAS.app"
    echo "🍎 启动 Mac App Store (MAS) 规范构建流水线 [沙盒 + 目录授权 + 移除 Sparkle]..."
    export DEVLEMON_BUILD_TARGET="appstore"
else
    APP_NAME="DevLemon.app"
    echo "🚀 启动 Direct (官网/GitHub) 规范构建流水线 [无沙盒 + Sparkle 2.0 + CLI 自动软链接]..."
    export DEVLEMON_BUILD_TARGET="direct"
fi

APP_BUNDLE="${OUTPUT_DIR}/${APP_NAME}"

echo "🍋 [1/5] Building devlemon Go core engine..."
cd "${REPO_DIR}"
go build -ldflags "-s -w -X devlemon/internal/version.Version=v${VERSION}" -o "${REPO_DIR}/devlemon" ./cmd/devlemon

echo "📦 [2/5] Building Swift release binary (Target: ${TARGET})..."
cd "${MAC_DIR}"
swift build -c release

SWIFT_BIN="${MAC_DIR}/.build/release/DevLemon"
if [[ ! -f "${SWIFT_BIN}" ]]; then
    echo "❌ Swift binary not found at ${SWIFT_BIN}"
    exit 1
fi

echo "📁 [3/5] Packaging ${APP_NAME} bundle..."
rm -rf "${APP_BUNDLE}"
mkdir -p "${APP_BUNDLE}/Contents/MacOS"
mkdir -p "${APP_BUNDLE}/Contents/Helpers"
mkdir -p "${APP_BUNDLE}/Contents/Resources"

cp "${SWIFT_BIN}" "${APP_BUNDLE}/Contents/MacOS/DevLemon"
cp "${REPO_DIR}/devlemon" "${APP_BUNDLE}/Contents/Helpers/devlemon"
chmod +x "${APP_BUNDLE}/Contents/MacOS/DevLemon"
chmod +x "${APP_BUNDLE}/Contents/Helpers/devlemon"

# 复制 Apple Privacy Manifest (合规必须)
if [[ -f "${MAC_DIR}/Resources/PrivacyInfo.xcprivacy" ]]; then
    cp "${MAC_DIR}/Resources/PrivacyInfo.xcprivacy" "${APP_BUNDLE}/Contents/Resources/PrivacyInfo.xcprivacy"
fi

# 仅直装版嵌入 Sparkle.framework
if [[ "${TARGET}" != "mas" ]]; then
    mkdir -p "${APP_BUNDLE}/Contents/Frameworks"
    if [[ -d "${MAC_DIR}/.build/release/Sparkle.framework" ]]; then
        cp -R "${MAC_DIR}/.build/release/Sparkle.framework" "${APP_BUNDLE}/Contents/Frameworks/"
    elif [[ -d "${MAC_DIR}/.build/artifacts/sparkle/Sparkle/Sparkle.xcframework/macos-arm64_x86_64/Sparkle.framework" ]]; then
        cp -R "${MAC_DIR}/.build/artifacts/sparkle/Sparkle/Sparkle.xcframework/macos-arm64_x86_64/Sparkle.framework" "${APP_BUNDLE}/Contents/Frameworks/"
    fi
    install_name_tool -add_rpath "@loader_path/../Frameworks" "${APP_BUNDLE}/Contents/MacOS/DevLemon" 2>/dev/null || true
fi

# 复制 AppIcon.icns
if [[ ! -f "${MAC_DIR}/Resources/AppIcon.icns" ]]; then
    echo "🎨 Generating AppIcon.icns..."
    swift "${MAC_DIR}/scripts/generate_app_icon.swift"
fi
if [[ -f "${MAC_DIR}/Resources/AppIcon.icns" ]]; then
    cp "${MAC_DIR}/Resources/AppIcon.icns" "${APP_BUNDLE}/Contents/Resources/AppIcon.icns"
fi

echo "📝 [4/5] Generating Info.plist (${TARGET})..."
if [[ "${TARGET}" == "mas" ]]; then
cat << EOF > "${APP_BUNDLE}/Contents/Info.plist"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleName</key>
    <string>DevLemon</string>
    <key>CFBundleDisplayName</key>
    <string>DevLemon</string>
    <key>CFBundleIdentifier</key>
    <string>com.iiwish.devlemon</string>
    <key>CFBundleVersion</key>
    <string>${VERSION}</string>
    <key>CFBundleShortVersionString</key>
    <string>${VERSION}</string>
    <key>CFBundleExecutable</key>
    <string>DevLemon</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>LSMinimumSystemVersion</key>
    <string>13.0</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSPrincipalClass</key>
    <string>NSApplication</string>
    <key>LSUIElement</key>
    <false/>
    <key>ITSAppUsesNonExemptEncryption</key>
    <false/>
    <key>LSApplicationCategoryType</key>
    <string>public.app-category.developer-tools</string>
    <key>NSHumanReadableCopyright</key>
    <string>Copyright © 2025-2026 DevLemon. All rights reserved.</string>
</dict>
</plist>
EOF
else
cat << EOF > "${APP_BUNDLE}/Contents/Info.plist"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleName</key>
    <string>DevLemon</string>
    <key>CFBundleDisplayName</key>
    <string>DevLemon</string>
    <key>CFBundleIdentifier</key>
    <string>com.iiwish.devlemon</string>
    <key>CFBundleVersion</key>
    <string>${VERSION}</string>
    <key>CFBundleShortVersionString</key>
    <string>${VERSION}</string>
    <key>CFBundleExecutable</key>
    <string>DevLemon</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>LSMinimumSystemVersion</key>
    <string>13.0</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSPrincipalClass</key>
    <string>NSApplication</string>
    <key>LSUIElement</key>
    <false/>
    <key>SUFeedURL</key>
    <string>https://raw.githubusercontent.com/iiwish/devlemon/main/appcast.xml</string>
    <key>SUPublicEDKey</key>
    <string>rv9WCxT1YcV5u40DjfKkTre7ichbJHWLWmMU45YXul0=</string>
    <key>SUEnableAutomaticChecks</key>
    <true/>
    <key>SUAllowsAutomaticUpdates</key>
    <true/>
    <key>SUScheduledCheckInterval</key>
    <integer>86400</integer>
</dict>
</plist>
EOF
fi

echo "🔏 [5/5] Codesigning bundle..."
if [[ "${TARGET}" == "mas" ]]; then
    # 签署沙盒内置 helper 引擎
    codesign --force --sign - \
        --entitlements "${MAC_DIR}/Resources/devlemon-helper.entitlements" \
        "${APP_BUNDLE}/Contents/Resources/devlemon" 2>/dev/null || true

    # 签署沙盒主应用
    codesign --force --deep --sign - \
        --entitlements "${MAC_DIR}/Resources/DevLemon-MAS.entitlements" \
        "${APP_BUNDLE}" 2>/dev/null || true
    echo "✨ Successfully created Sandboxed Mac App Store bundle: ${APP_BUNDLE}"
else
    codesign --force --deep --sign - "${APP_BUNDLE}" 2>/dev/null || true
    echo "✨ Successfully created Direct release bundle: ${APP_BUNDLE}"
fi
