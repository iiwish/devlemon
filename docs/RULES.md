# 清理规则

DevLemon 的「包管理器缓存」「应用缓存」两组清理项由 YAML 规则描述，一条规则把一个或多个目录描述为一个清理项。

- **内置规则**：[`internal/rules/builtin/`](../internal/rules/builtin/)，随程序一起发布。新增对某个工具缓存的支持，只需在这里加一条规则并提交 PR，不需要写 Go 代码。
- **用户规则**：`~/.config/devlemon/rules.d/*.yaml`（或 `$XDG_CONFIG_HOME/devlemon/rules.d/`、`$DEVLEMON_RULES_DIR`），可以新增规则、覆盖或禁用内置规则，只对本机生效。

需要判断逻辑的清理项（应用运行检测、工具链版本比较、已卸载应用识别、浏览器缓存结构校验等）仍由 Go 实现，不在规则文件中。

## 示例

```yaml
group: package_caches          # 归入哪一组：package_caches 或 app_caches
rules:
  - id: my_tool_cache          # 唯一 ID：小写字母、数字、下划线
    title:
      en: "My Tool Cache"
      zh: "My Tool 下载缓存"     # 可省略，缺少时显示英文
    description:
      en: "Downloads cached by my-tool; fetched again on next use"
      zh: "my-tool 下载的缓存文件，下次使用时重新下载"
    paths:                     # 一个或多个目录，聚合为一个清理项
      - "~/.my-tool/cache"
      - "~/Library/Caches/my-tool"
    risk: rebuildable          # safe / rebuildable / caution
    min_size: 10MB             # 小于该体积不显示，默认 10MB
    clean_command: ["my-tool", "cache", "clean"]   # 可选：用命令代替直接删除目录
    os: [darwin]               # 可选：仅在指定系统上生效（darwin / linux）
```

## 字段

| 字段 | 必填 | 说明 |
|---|---|---|
| `id` | ✅ | 唯一标识，小写字母、数字、下划线。与内置规则同名时覆盖内置规则 |
| `title.en` | ✅ | 标题；`title.zh` 可选 |
| `description` | | 说明清理后的影响，建议写清「会重新下载」还是「会丢失数据」 |
| `paths` | ✅ | 目录列表，见下方路径规则；不存在的目录自动忽略 |
| `risk` | ✅ | `safe`：无副作用；`rebuildable`：可自动重建，但需要重新下载或编译；`caution`：可能含用户数据，需逐项确认 |
| `min_size` | | 低于该体积不列出，如 `512KB`、`10MB`、`1GB`，默认 `10MB` |
| `clean_command` | | 清理时执行的命令（不经过 shell），如 `["go", "clean", "-cache"]` |
| `os` | | 适用的系统，省略表示所有系统 |
| `enabled` | | 设为 `false` 可禁用同 ID 的规则 |

只有 `system_maintenance` 组的安全项会进入菜单栏的快速清理；规则文件中的 `safe` 项只在深度清理中默认勾选。

## 路径规则

为避免误删，加载时会强制校验，不合规的规则会被跳过并提示：

- 必须以 `~/`（主目录）或 `{darwin_user_cache}/`（macOS 每用户缓存目录，`$TMPDIR` 同级的 `C/`）开头
- 不支持通配符、环境变量、`.` 与 `..`
- 不能是受保护的目录本身，例如 `~/Library`、`~/Library/Caches`、`~/Documents`、`~/.ssh`、`~/.config`；它们的子目录可以
- 删除前清理器还会再次校验路径层级

## 禁用或覆盖内置规则

```yaml
# ~/.config/devlemon/rules.d/overrides.yaml
rules:
  - id: npm_cache
    enabled: false             # 只禁用时可以省略 group 与其他字段
```

覆盖内置规则时需要写完整规则（同一 ID 的用户规则会整体替换内置规则）。

## 查看与校验

```bash
devlemon rules
```

```bash
devlemon rules check ~/.config/devlemon/rules.d/my-rules.yaml
```

`devlemon rules` 列出当前生效的全部规则及其来源，并显示被跳过的用户规则；`rules check` 校验规则文件，有错误时退出码非 0，可用于 CI。

## 提交内置规则

1. 在 `internal/rules/builtin/` 对应文件中添加规则，中英文标题与描述都需要填写
2. 运行 `go test ./internal/rules/`（会校验全部内置规则）
3. 在 PR 中说明该目录的用途，以及删除后工具会如何恢复（重新下载、重新编译等）
