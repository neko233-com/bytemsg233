# 编辑器插件

ByteMsg233 提供两个独立 public Git 子仓库：

| IDE | 路径 | 仓库 |
| --- | --- | --- |
| JetBrains / GoLand / Rider | `editors/jetbrains` | https://github.com/neko233-com/plugin-jetbrains-bytemsg233 |
| VS Code | `editors/vscode` | https://github.com/neko233-com/plugin-vscode-bytemsg233 |

支持 `.bmsg.json`、`.bmsg.yaml`、`.bmsg.yml` 和 legacy `.bmsg` 的高亮、补全、原生格式化、校验、当前文件类型导航、结构视图和语言导出；支持 Markdown / HTML / BMSG / Proto 导出。安装方式和设置详见各插件 README。ZIP / VSIX 可从各仓库 GitHub Releases 下载；插件市场发布状态以实际市场页面为准。

使用 CLI v1.1.3+：`go install github.com/neko233-com/bytemsg233/cmd/bytemsg233@v1.1.3`。可执行文件路径可在 IDE 设置。语言菜单使用 `bytemsg233 languages` 返回的实际生成器（v1.1.3：Go、C#、TypeScript、Java、Rust、Python）。

- `bytemsg233 check hero.bmsg.json`：只校验，不生成文件。
- `bytemsg233 format hero.bmsg.json -w`：原地格式化。默认不带 `-w` 输出到 stdout。
- `bytemsg233 format hero.bmsg.json --stdin`：读取未保存的缓冲区；文件名决定格式。
- `bytemsg233 compile hero.bmsg.json -l go,csharp -o generated --locale zh`：生成语言代码。
- `bytemsg233 export hero.bmsg.json -f md,html,bmsg,proto -o generated/docs`：导出文档和协议。

JSON 格式化保留字段顺序和大整数原文；YAML 保留注释。生成前会保存当前文件，保存失败则停止。未受信任项目禁止执行 CLI。普通 JSON/YAML 文件保持原类型；`.proto` 提供校验 / 导出入口，语法高亮和格式化交给 Proto 插件。
