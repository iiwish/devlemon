#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# DevLemon - Mac App Store (MAS) 自动化签名与发布包 (.pkg) 构建脚本
# ==============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MAC_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_DIR="$(cd "${MAC_DIR}/.." && pwd)"
OUTPUT_DIR="${REPO_DIR}/build"
MAS_STAGE_DIR="${OUTPUT_DIR}/mas_stage"
FINAL_PKG="${OUTPUT_DIR}/DevLemon.pkg"

# 1. 编译 MAS 沙盒版本
echo "🍋 [1/4] 构建 Mac App Store 规范沙盒版应用程序..."
bash "${SCRIPT_DIR}/build_app.sh" mas

# 2. 准备纯净的应用包结构 (目标名称统一为 DevLemon.app)
rm -rf "${MAS_STAGE_DIR}"
mkdir -p "${MAS_STAGE_DIR}"
cp -R "${OUTPUT_DIR}/DevLemon-MAS.app" "${MAS_STAGE_DIR}/DevLemon.app"
APP_TARGET="${MAS_STAGE_DIR}/DevLemon.app"

# 3. 自动探测钥匙串中的苹果分发证书
echo "🔍 [2/4] 检索本地 Apple Distribution 证书与 Installer 证书..."

APP_CERT=$(security find-identity -v -p codesigning | grep "Apple Distribution:" | head -n 1 | awk -F'"' '{print $2}' || true)
if [[ -z "${APP_CERT}" ]]; then
    APP_CERT=$(security find-identity -v -p codesigning | grep "3rd Party Mac Developer Application:" | head -n 1 | awk -F'"' '{print $2}' || true)
fi

INSTALLER_CERT=$(security find-identity -v | grep "Mac Installer Distribution:" | head -n 1 | awk -F'"' '{print $2}' || true)
if [[ -z "${INSTALLER_CERT}" ]]; then
    INSTALLER_CERT=$(security find-identity -v | grep "3rd Party Mac Developer Installer:" | head -n 1 | awk -F'"' '{print $2}' || true)
fi

PROVISIONING_PROFILE="${1:-}"
if [[ -z "${PROVISIONING_PROFILE}" && -f "${MAC_DIR}/embedded.provisionprofile" ]]; then
    PROVISIONING_PROFILE="${MAC_DIR}/embedded.provisionprofile"
fi

if [[ -n "${PROVISIONING_PROFILE}" && -f "${PROVISIONING_PROFILE}" ]]; then
    echo "📋 嵌入 Provisioning Profile: ${PROVISIONING_PROFILE}"
    cp "${PROVISIONING_PROFILE}" "${APP_TARGET}/Contents/embedded.provisionprofile"
fi

# 4. 执行正式签名 (App Distribution)
if [[ -n "${APP_CERT}" ]]; then
    echo "✍️  [3/4] 使用证书签名应用: ${APP_CERT}"
    
    # 签署沙盒子进程 helper
    codesign --force --timestamp --options runtime \
        --entitlements "${MAC_DIR}/Resources/devlemon-helper.entitlements" \
        --sign "${APP_CERT}" \
        "${APP_TARGET}/Contents/Resources/devlemon"

    # 签署主程序与 App Bundle
    codesign --force --timestamp --options runtime \
        --entitlements "${MAC_DIR}/Resources/DevLemon-MAS.entitlements" \
        --sign "${APP_CERT}" \
        "${APP_TARGET}"

    echo "✅ 应用签名完成并通过验证："
    codesign -v "${APP_TARGET}"
else
    echo "⚠️ 未在钥匙串中找到 Apple Distribution 证书，保留 Ad-hoc 签名状态。"
    echo "   (正式提交 Mac App Store 前必须在 developer.apple.com 创建并导入分发证书)"
fi

# 5. 打包为 Mac App Store Installer (.pkg)
echo "📦 [4/4] 使用 productbuild 封装 App Store 上架安装包 (.pkg)..."
if [[ -n "${INSTALLER_CERT}" ]]; then
    productbuild --component "${APP_TARGET}" /Applications --sign "${INSTALLER_CERT}" "${FINAL_PKG}"
    echo "✨ 成功生成已签名的 Mac App Store 安装包: ${FINAL_PKG}"
    echo "🚀 您可以直接将 ${FINAL_PKG} 拖入 Transporter 应用提交审核！"
else
    productbuild --component "${APP_TARGET}" /Applications "${FINAL_PKG}"
    echo "⚠️ 已生成未签名的安装包: ${FINAL_PKG}"
    echo "ℹ️  说明: App Store Connect 上传需要此包由 [Mac Installer Distribution] 证书签名。"
    echo "   可在 Apple Developer 证书控制台申请并在安装后重新运行此脚本。"
fi

echo ""
echo "🎉 MAS 打包就绪: ${FINAL_PKG}"
