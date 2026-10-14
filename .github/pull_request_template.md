<!-- 感谢贡献！请先阅读 CONTRIBUTING.md。Thanks for contributing! Please read CONTRIBUTING.md first. -->

## 改动内容 / What

<!-- 这个 PR 做了什么、为什么。关联的 Issue 写在这里，如 Closes #123 -->

## 改动类型 / Type

- [ ] 新增或修改清理规则 (YAML) / New or changed cleanup rule
- [ ] 新增或修改探针 (Go) / New or changed probe
- [ ] Bug 修复 / Bug fix
- [ ] macOS 菜单栏应用 / macOS menu bar app
- [ ] 文档 / Docs
- [ ] 其他 / Other

## 清理项说明（新增或修改清理项时填写）/ Cleanup items

| 项目 | 说明 |
|---|---|
| 规则 ID / Rule ID | |
| 路径 / Paths | |
| 风险等级 / Risk | safe / rebuildable / caution |
| 这个目录是什么 / What the directory holds | |
| 删除后工具如何恢复 / How the tool recovers | <!-- 自动重新下载？重新编译？会丢失什么？ --> |
| 验证环境 / Tested on | <!-- 如 macOS 15 + pnpm 11.2 --> |

## 检查清单 / Checklist

- [ ] `go vet ./...` 与 `go test -race ./...` 通过 / pass
- [ ] 代码已 `gofmt` / code is gofmt-ed
- [ ] 用户可见的文字同时提供中英文 / user-facing text has both English and Chinese
- [ ] 扫描阶段只读，不删除或修改任何文件 / scanning is read-only
- [ ] 风险等级按 CONTRIBUTING.md 标注；`safe` 仅用于完全无副作用的内容 / risk level follows CONTRIBUTING.md
- [ ] 不与已有规则或探针重复统计同一目录 / no double counting with existing rules or probes
- [ ] 新增的规则文件通过 `devlemon rules check` / new rule files pass `devlemon rules check`
- [ ] 未修改版本号与 `appcast.xml` / no version bump or `appcast.xml` change
- [ ] 不包含从 GPL 等不兼容许可证项目复制的代码或规则数据 / no code or rule data copied from GPL-licensed projects
