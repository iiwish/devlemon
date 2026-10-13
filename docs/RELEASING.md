# 发版指南

推送 `v*` 标签后，[release.yml](../.github/workflows/release.yml) 会自动完成：

1. 构建四个平台的 CLI 压缩包与 `checksums.txt`
2. 使用 **Developer ID Application** 证书签名 `DevLemon.app`（强化运行时 + 时间戳）
3. 提交 Apple 公证并钉入公证票据，生成 `DevLemon-vX.Y.Z-macOS.zip`
4. 生成 Sparkle 签名更新源 `appcast.xml` 并推送到 main
5. 发布 GitHub Release，并更新 Homebrew tap (`iiwish/homebrew-tap`)

签名与公证后，用户首次打开不会再出现「无法验证开发者」提示；并且后续每次更新的签名身份保持一致，用户已授予的系统权限（如访问「文稿」、控制访达）不会因升级而失效。

## 一次性配置

### 1. 申请 Developer ID Application 证书

需要 Apple 开发者账号的「账户持有人」角色。

- **Xcode**：设置 → Accounts → 选中团队 → Manage Certificates… → 左下角 `+` → **Developer ID Application**
- 或在 [developer.apple.com](https://developer.apple.com/account/resources/certificates/list) → Certificates → `+` → Developer ID Application

> 请选择 **G2 Sub-CA**：选旧版 Sub-CA 或由 Xcode 自动创建时，证书可能由旧中间证书签发，有效期会被截断到 2027-02-01。网页创建需要先用「钥匙串访问 → 证书助理 → 从证书颁发机构请求证书」生成 CSR。

证书会进入本机钥匙串，可用下面命令确认：

```bash
security find-identity -v -p codesigning | grep "Developer ID Application"
```

### 2. 导出证书并配置 GitHub Secrets

打开「钥匙串访问」→ 左侧选「登录」→ 顶部选「我的证书」→ 右键 `Developer ID Application: Beijing Wish Technology Co., Ltd (38S76J82P4)` → 导出，格式选「个人信息交换 (.p12)」，设置一个导出密码。只导出这一张证书，不要把其他证书的私钥一起上传。钥匙串里如果有多张同名证书，按「过期时间」区分，选最新的那张。

上传前确认导出的证书（会提示输入导出密码），输出中应包含 `Developer ID Application`，签发方为 `OU=G2`：

```bash
openssl pkcs12 -in DeveloperID.p12 -nokeys | grep -E "subject|issuer"
```

```bash
base64 -i DeveloperID.p12 | gh secret set DEVELOPER_ID_CERT_P12
```

```bash
gh secret set DEVELOPER_ID_CERT_PASSWORD
```

第二条命令会提示输入导出 `.p12` 时设置的密码。完成后请删除导出的 `.p12` 文件。

### 3. 创建公证用的 App Store Connect API 密钥

[App Store Connect](https://appstoreconnect.apple.com/access/integrations/api) → 用户和访问 → 集成 → App Store Connect API → 团队密钥 → 生成，角色选择 **Developer**。

记下 **Issuer ID** 与 **密钥 ID**，下载 `AuthKey_XXXXXXXXXX.p8`（只能下载一次）。

```bash
gh secret set NOTARY_KEY_P8 < AuthKey_XXXXXXXXXX.p8
```

```bash
gh secret set NOTARY_KEY_ID
```

```bash
gh secret set NOTARY_ISSUER_ID
```

> 生成团队密钥前，需要账户持有人先在该页面开通 App Store Connect API 访问。暂时无法开通时，可改用下面的 Apple ID 方式。

**备选：Apple ID + App 专用密码**。在 [account.apple.com](https://account.apple.com) → 登录与安全 → App 专用密码 中生成一个密码，然后：

```bash
gh secret set NOTARY_APPLE_ID
```

```bash
gh secret set NOTARY_PASSWORD
```

两组凭据都配置时优先使用 API 密钥。

### 4. 已有的 Secrets

| Secret | 用途 |
|---|---|
| `SPARKLE_PRIVATE_KEY` | Sparkle 更新包 EdDSA 签名 |
| `HOMEBREW_TAP_TOKEN` | 推送 `iiwish/homebrew-tap`（仅需该仓库 Contents 读写权限） |

任一 Secret 缺失时对应步骤会直接失败，而不是悄悄发布未签名或未同步的版本。

## 发版步骤

1. 在以下位置把版本号改为新版本：`internal/version/version.go`、`mac/scripts/build_app.sh`、`mac/Sources/DevLemonApp/Views/Settings/SettingsView.swift`、`README.md` / `README_zh.md` 徽章
2. 合并到 main，等待 CI 通过
3. 打标签并推送：

```bash
git tag -a v0.2.14 -m "v0.2.14" && git push origin v0.2.14
```

## 本地签名与公证（可选）

```bash
DEVLEMON_SIGN_IDENTITY="Developer ID Application: Beijing Wish Technology Co., Ltd (38S76J82P4)" bash mac/scripts/build_app.sh
```

首次使用前把公证凭据存入钥匙串（之后只需引用配置名）：

```bash
xcrun notarytool store-credentials devlemon-notary --key AuthKey_XXXXXXXXXX.p8 --key-id XXXXXXXXXX --issuer xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

```bash
NOTARY_KEYCHAIN_PROFILE=devlemon-notary bash mac/scripts/notarize.sh build/DevLemon.app build/DevLemon-macOS.zip
```

钥匙串中有多张同名证书时，`codesign` 会报身份不唯一，此时把 `DEVLEMON_SIGN_IDENTITY` 设为证书的 SHA-1 指纹（`security find-identity -v -p codesigning` 输出的第一列）。

不设置 `DEVLEMON_SIGN_IDENTITY` 时 `build_app.sh` 使用临时签名，适合本地开发调试。
