#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MAC_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_DIR="$(cd "${MAC_DIR}/.." && pwd)"
OUTPUT_DIR="${REPO_DIR}/build"
APP_NAME="DevLemon.app"
APP_BUNDLE="${OUTPUT_DIR}/${APP_NAME}"
VERSION="0.2.5"

echo "🍋 [1/5] Building devlemon Go core engine..."
cd "${REPO_DIR}"
go build -ldflags "-s -w -X devlemon/internal/version.Version=v${VERSION}" -o "${REPO_DIR}/devlemon" ./cmd/devlemon

echo "📦 [2/5] Building Swift release binary with Sparkle..."
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
mkdir -p "${APP_BUNDLE}/Contents/Resources"
mkdir -p "${APP_BUNDLE}/Contents/Frameworks"

cp "${SWIFT_BIN}" "${APP_BUNDLE}/Contents/MacOS/DevLemon"
cp "${REPO_DIR}/devlemon" "${APP_BUNDLE}/Contents/Resources/devlemon"
chmod +x "${APP_BUNDLE}/Contents/MacOS/DevLemon"
chmod +x "${APP_BUNDLE}/Contents/Resources/devlemon"

# 嵌入 Sparkle.framework
if [[ -d "${MAC_DIR}/.build/release/Sparkle.framework" ]]; then
    cp -R "${MAC_DIR}/.build/release/Sparkle.framework" "${APP_BUNDLE}/Contents/Frameworks/"
elif [[ -d "${MAC_DIR}/.build/artifacts/sparkle/Sparkle/Sparkle.xcframework/macos-arm64_x86_64/Sparkle.framework" ]]; then
    cp -R "${MAC_DIR}/.build/artifacts/sparkle/Sparkle/Sparkle.xcframework/macos-arm64_x86_64/Sparkle.framework" "${APP_BUNDLE}/Contents/Frameworks/"
fi
install_name_tool -add_rpath "@loader_path/../Frameworks" "${APP_BUNDLE}/Contents/MacOS/DevLemon" 2>/dev/null || true

# 复制 AppIcon.icns
if [[ ! -f "${MAC_DIR}/Resources/AppIcon.icns" ]]; then
    echo "🎨 Generating AppIcon.icns..."
    swift "${MAC_DIR}/scripts/generate_app_icon.swift"
fi
if [[ -f "${MAC_DIR}/Resources/AppIcon.icns" ]]; then
    cp "${MAC_DIR}/Resources/AppIcon.icns" "${APP_BUNDLE}/Contents/Resources/AppIcon.icns"
fi

echo "📝 [4/5] Generating Info.plist (v${VERSION})..."
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

echo "🔏 [5/5] Ad-hoc codesigning bundle..."
codesign --force --deep --sign - "${APP_BUNDLE}" 2>/dev/null || true

echo "✨ Successfully created ${APP_BUNDLE} (v${VERSION})"
