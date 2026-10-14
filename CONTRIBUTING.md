# 贡献指南

感谢你愿意为 DevLemon 出力！DevLemon 是一个**面向开发者**的开源磁盘清理工具，专注于通用清理工具看不懂的东西：构建产物、包管理器缓存、工具链旧版本、模拟器、容器与 AI 模型缓存。

[English summary](#english-summary) is at the end.

## 范围

**欢迎：**

- 新的开发工具、语言、SDK、AI 工具的缓存规则（最常见、最容易上手的贡献）
- 更准确的「是否仍在使用」判断，减少误报和漏报
- Bug 修复、性能优化、文档与翻译

**不在范围内：**

- 与开发无关的通用清理（相册、聊天记录、隐私痕迹等）和「内存加速」类功能
- 需要 root 权限或常驻后台服务的功能
- 从 GPL 等不兼容许可证的项目中复制的代码或规则数据。DevLemon 采用 MIT 许可证，可以参考其他项目的思路，但请自行实现

不确定是否合适时，先开一个 Issue 讨论。

## 安全原则

清理工具出错的代价是用户丢数据，所以所有贡献都要遵守：

1. **宁可漏扫，不可误删**：无法确认是缓存的目录不要列出；判断条件拿不准时选更保守的一侧。
2. **扫描只读**：探针（`internal/probe`）在扫描阶段绝不修改或删除任何文件，删除只发生在用户确认后的清理阶段。
3. **正确标注风险等级：**

   | 等级 | 含义 | 示例 |
   |---|---|---|
   | `safe` | 删除后无任何可感知影响 | 应用日志、已安装完毕的更新包残留 |
   | `rebuildable` | 会自动重建，但需要重新下载或编译 | npm 缓存、DerivedData、旧版工具链 |
   | `caution` | 可能包含用户数据或需要人工判断 | 设备备份、本地大模型、已卸载应用残留 |

   新规则默认用 `rebuildable`。只有完全没有副作用时才用 `safe`：`safe` 项可能在菜单栏的一键快速清理中被删除。
4. **避让正在使用的内容**：应用运行中、最近仍有写入（项目里通常以 30 天为界）的内容，不应该进入快速清理。
5. **不触碰用户数据**：聊天记录、登录凭据、文档、数据库、`~/.ssh`、`~/.config` 等一律不碰。

## 贡献方式一：新增缓存规则（只需 YAML）

如果某个工具的缓存就是一个（或几个）目录，删掉后工具会自动恢复，那么只需要在 [`internal/rules/builtin/`](internal/rules/builtin/) 中加一条规则：

- 开发语言、包管理器、构建工具的缓存 → `package_caches.yaml`
- 桌面应用的缓存 → `app_caches.yaml`

```yaml
  - id: my_tool_cache
    title:
      en: "My Tool Download Cache"
      zh: "My Tool 下载缓存"
    description:
      en: "Packages downloaded by my-tool (~/.my-tool/cache); downloaded again on next use"
      zh: "my-tool 下载的依赖包 (~/.my-tool/cache)，下次使用时重新下载"
    paths:
      - "~/.my-tool/cache"
    risk: rebuildable
    min_size: 10MB
```

要点：

- 中英文标题与描述都要填写；描述写清楚**删除后会发生什么**（重新下载？重新编译？丢失什么？）
- 如果工具自带清理命令（如 `go clean -cache`），用 `clean_command` 优先调用它
- 只适用于某个系统的路径加上 `os: [darwin]` 或 `os: [linux]`
- 路径会经过安全校验（只能在主目录下、不能是 `~/Library` 这类整体目录等），完整格式见 [docs/RULES.md](docs/RULES.md)

校验与测试：

```bash
go run ./cmd/devlemon rules check internal/rules/builtin/*.yaml
```

```bash
go test ./internal/rules/
```

## 贡献方式二：Go 探针

需要判断逻辑的清理项写成探针，例如应用是否在运行、哪个版本仍被项目使用、目录结构是否符合预期。可以参考：

- [`toolchain_probe.go`](internal/probe/toolchain_probe.go)：工具链旧版本（版本比较、30 天未改动、保留默认版本）
- [`updater_cache.go`](internal/probe/updater_cache.go)：更新安装包（运行中检测）
- [`chromium_cache.go`](internal/probe/chromium_cache.go)：严格校验缓存目录结构，防止误删同名目录

约定：

- 新探针实现 `probe.Probe` 接口，在 [`internal/engine/engine.go`](internal/engine/engine.go) 注册
- 用户可见的文字都用 `i18n.T(英文, 中文)` 同时提供中英文
- 能用系统调用就不要执行外部命令；需要外部命令时处理好命令不存在、超时的情况
- 和已有规则或探针覆盖同一目录时要去重，避免同一份空间被统计两次
- 测试用 `t.TempDir()` 构造目录结构，大文件用稀疏文件模拟（参考 `toolchain_probe_test.go`），不要依赖本机真实环境

## 开发环境

- Go：版本见 `go.mod`，较旧的 Go 会自动下载匹配的工具链
- macOS 菜单栏应用（可选）：Xcode 15+，macOS 13+
- Linux 上可以开发和测试命令行部分

常用命令：

```bash
go build -o devlemon ./cmd/devlemon
```

```bash
go vet ./... && go test -race ./...
```

```bash
./devlemon scan
```

```bash
DEVLEMON_LANG=en ./devlemon scan --json
```

```bash
make app
```

在不影响自己机器的情况下验证规则，可以把 `HOME` 指向一个临时目录，并在里面按规则路径放几个文件：

```bash
HOME=/tmp/fakehome ./devlemon scan --json
```

清理相关的改动请先用 `--dry-run` 预演：

```bash
./devlemon clean --dry-run
```

## 提交与 Pull Request

- 从 `main` 新建分支，一个 PR 只做一件事
- 提交信息使用 [Conventional Commits](https://www.conventionalcommits.org/)，与仓库历史保持一致，例如 `feat(rules): add Flutter pub cache`、`fix(probe): skip running Android emulators`
- 提交前运行 `gofmt`、`go vet ./...` 与 `go test -race ./...`；CI 会在 macOS 与 Linux 上运行同样的检查
- 不要修改版本号，也不要提交 `appcast.xml`，它们由维护者在发版时处理（见 [docs/RELEASING.md](docs/RELEASING.md)）
- 填写 PR 模板中的检查清单，新增清理项时请说明你在什么系统、什么工具版本上验证过

## 报告问题

- **误删或误报（最重要）**：请提供清理项的 ID、路径和你认为它不该被清理的原因。可以用 `devlemon scan --json` 找到对应条目，提交前请去掉其中的隐私信息
- **漏报**：告诉我们工具名称、缓存目录位置，以及删除后工具会如何恢复
- 安全相关问题（例如可能导致删除用户数据的缺陷）请在 Issue 中明确标注

## English Summary

DevLemon is a developer-focused, MIT-licensed disk cleaner. Contributions of cache rules for more developer tools, languages, SDKs and AI tools are especially welcome.

- **Safety first**: when unsure, don't list it. Probes are read-only. Use `rebuildable` for anything that needs a re-download or rebuild, and `safe` only when there is no noticeable side effect.
- **New cache rule (YAML only)**: add an entry to [`internal/rules/builtin/`](internal/rules/builtin/) with English and Chinese title/description, then run `go run ./cmd/devlemon rules check internal/rules/builtin/*.yaml` and `go test ./internal/rules/`. Format: [docs/RULES.md](docs/RULES.md).
- **Probes with logic**: implement `probe.Probe`, register it in `internal/engine/engine.go`, provide all user-facing text via `i18n.T(en, zh)`, and test with `t.TempDir()` and sparse files.
- **Checks**: `gofmt`, `go vet ./...`, `go test -race ./...`. Use Conventional Commits. Don't bump versions or edit `appcast.xml`.
- **Out of scope**: general-purpose (non-developer) cleaning, memory "boosters", anything requiring root, and code or rule data copied from GPL-licensed projects.
