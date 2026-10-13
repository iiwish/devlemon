#!/usr/bin/env bash
# 公证已用 Developer ID 签名的 DevLemon.app，并将公证票据钉入 App，最后输出可分发的 zip
# 用法: mac/scripts/notarize.sh <DevLemon.app> <输出 zip 路径>
#
# 凭据（任选其一）：
#   1. App Store Connect API 密钥：NOTARY_KEY_PATH（.p8 文件路径）、NOTARY_KEY_ID、NOTARY_ISSUER_ID
#   2. Apple ID + App 专用密码：NOTARY_APPLE_ID、NOTARY_PASSWORD、NOTARY_TEAM_ID
#   3. 本地钥匙串配置：NOTARY_KEYCHAIN_PROFILE（先执行 xcrun notarytool store-credentials 创建）
set -euo pipefail

APP="${1:?usage: notarize.sh <app> <output.zip>}"
OUT_ZIP="${2:?usage: notarize.sh <app> <output.zip>}"

if [[ -n "${NOTARY_KEYCHAIN_PROFILE:-}" ]]; then
    AUTH=(--keychain-profile "${NOTARY_KEYCHAIN_PROFILE}")
elif [[ -n "${NOTARY_KEY_PATH:-}" && -n "${NOTARY_KEY_ID:-}" && -n "${NOTARY_ISSUER_ID:-}" ]]; then
    AUTH=(--key "${NOTARY_KEY_PATH}" --key-id "${NOTARY_KEY_ID}" --issuer "${NOTARY_ISSUER_ID}")
elif [[ -n "${NOTARY_APPLE_ID:-}" && -n "${NOTARY_PASSWORD:-}" && -n "${NOTARY_TEAM_ID:-}" ]]; then
    AUTH=(--apple-id "${NOTARY_APPLE_ID}" --password "${NOTARY_PASSWORD}" --team-id "${NOTARY_TEAM_ID}")
else
    echo "❌ 缺少公证凭据：请设置 NOTARY_KEYCHAIN_PROFILE、NOTARY_KEY_PATH / NOTARY_KEY_ID / NOTARY_ISSUER_ID，或 NOTARY_APPLE_ID / NOTARY_PASSWORD / NOTARY_TEAM_ID" >&2
    exit 1
fi

# 临时签名的 App 无法公证，提前给出明确错误
if ! codesign -dv "${APP}" 2>&1 | grep -q "^Authority=Developer ID Application"; then
    echo "❌ ${APP} 未使用 Developer ID Application 证书签名" >&2
    exit 1
fi

WORK="$(mktemp -d)"
trap 'rm -rf "${WORK}"' EXIT
SUBMIT_ZIP="${WORK}/submit.zip"

echo "📤 提交公证..."
ditto -c -k --sequesterRsrc --keepParent "${APP}" "${SUBMIT_ZIP}"
xcrun notarytool submit "${SUBMIT_ZIP}" "${AUTH[@]}" --wait --timeout 30m --output-format json > "${WORK}/result.json"
cat "${WORK}/result.json"

STATUS=$(/usr/bin/plutil -extract status raw -o - "${WORK}/result.json" 2>/dev/null || true)
SUBMISSION_ID=$(/usr/bin/plutil -extract id raw -o - "${WORK}/result.json" 2>/dev/null || true)
if [[ "${STATUS}" != "Accepted" ]]; then
    echo "❌ 公证未通过 (status=${STATUS:-unknown})，详细日志：" >&2
    if [[ -n "${SUBMISSION_ID}" ]]; then
        xcrun notarytool log "${SUBMISSION_ID}" "${AUTH[@]}" >&2 || true
    fi
    exit 1
fi

echo "📎 钉入公证票据..."
xcrun stapler staple "${APP}"
xcrun stapler validate "${APP}"
spctl --assess --type execute --verbose=2 "${APP}"

# 钉入票据后重新打包，供 GitHub Release 与 Sparkle 更新使用
rm -f "${OUT_ZIP}"
ditto -c -k --sequesterRsrc --keepParent "${APP}" "${OUT_ZIP}"
echo "✨ 公证完成: ${OUT_ZIP}"
