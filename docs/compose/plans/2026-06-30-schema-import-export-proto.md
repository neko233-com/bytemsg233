# ByteMsg233 Schema Import/Export 与 Proto 支持技术方案

## 1. 背景

ByteMsg233 当前以 `.bmsg.json` 作为主要协议描述 DSL，通过统一的 `schema.Schema` 生成不同目标语言代码。现有 CLI 使用方式已经接近三段式管线：

```text
schema file -> schema.Schema -> codegen -> target language files
schema file -> schema.Schema -> exporter -> md/html/bmsg files
```

但是当前实现里，导入和导出仍然偏函数分支：

- `compiler.Compile` 调用 `schema.ParseFile`，再调用 `codegen.Get(lang)` 和 `Generate`。
- `schema.ParseFile` 直接根据文件扩展名分支处理 `.bmsg`、`.yaml`、`.json`、`.toml`。
- `cmd/bytemsg233 export` 直接根据 `format` switch 调用 `exporter.Markdown`、`exporter.HTML`、`exporter.Bmsg`。

为了支持 `.proto` 作为元数据导入/导出格式，需要把元数据格式解析与目标语言代码生成解耦，形成明确的 `schema.Import`、`exporter.Export`、`codegen` 三层结构。

## 2. 现状分析

### 2.1 已有能力

当前导入能力：

| 输入格式 | 当前状态 | 入口 |
|---|---|---|
| `.bmsg.json` / `.json` | 已支持 | `schema.ParseJSON` |
| `.yaml` / `.yml` | 已支持 | `schema.ParseYAML` |
| `.bmsg` | 已支持 legacy DSL，并兼容 JSON/YAML 内容 | `schema.ParseBmsg` 及 `ParseFile` fallback |
| `.toml` | 有占位，不可用 | `schema.ParseTOML` |
| `.proto` | 未支持 | 无 |

当前导出能力：

| 输出格式 | 当前状态 | 入口 |
|---|---|---|
| `md` / `markdown` | 已支持 | `exporter.Markdown` |
| `html` | 已支持 | `exporter.HTML` |
| `bmsg` | 已支持 | `exporter.Bmsg` |
| `proto` | 未支持 | 无 |

当前代码生成能力：

- Go、C#、TypeScript、Rust、Java、Python 通过 `pkg/codegen` registry 生成。
- Go 生成器已基于 `schema.Schema` 生成 struct、enum、marshal/unmarshal、packet registry、协议版本 helper。
- `compiler.Compile` 外部 API 可以保持不变。

### 2.2 主要问题

1. 导入格式路由写在 `schema.ParseFile` 内部，新增格式会继续扩大分支。
2. 导出格式路由写在 CLI command 内部，新增格式会让 CLI 关心 exporter 细节。
3. `.proto` 作为行业常见协议描述格式，当前不能作为 ByteMsg233 元数据交换格式。
4. 反向解析目标语言实体类到 `schema.Schema` 成本较高，短期不纳入。

## 3. 目标与非目标

### 3.1 目标

1. 建立 `schema.Import` 层：不同元数据格式统一解析成 `schema.Schema`。
2. 建立 `exporter.Export` 层：`schema.Schema` 统一导出为不同元数据/文档格式。
3. 保持 `compiler.Compile` 外部调用不变。
4. 保持 `schema.ParseFile` 兼容入口，但内部委托给新的 import 路由。
5. 新增 `.proto` exporter：`schema.Schema -> .proto`。
6. 新增 `.proto` importer：`.proto -> schema.Schema`。
7. 支持 `.proto -> schema.Schema -> Go` 的 compile 链路。

### 3.2 非目标

1. 不做 Go/C#/TypeScript/Java 实体类反向解析到 `schema.Schema`。
2. 不调整各语言 codegen 的实体池化策略。
3. 不优化 `pkg/binary` buffer pool 的并发问题。
4. 不引入每条消息重复 protocol version 字段。
5. 不支持复杂 Protobuf 全量语法作为第一阶段目标，例如 `service`、`rpc`、`oneof`、跨文件 import 解析、复杂 custom option。

## 4. 总体设计

### 4.1 分层结构

```text
                +----------------+
 .json/.yaml -> | schema.Import  |
 .bmsg/.proto ->|                |
                +-------+--------+
                        |
                        v
                +---------------+
                | schema.Schema |
                +-------+-------+
                        |
        +---------------+----------------+
        |                                |
        v                                v
+---------------+                +----------------+
| exporter      |                | codegen         |
| md/html/bmsg  |                | go/csharp/ts    |
| proto         |                | java/rust/py    |
+---------------+                +----------------+
```

### 4.2 Import 层接口建议

新增 `pkg/schema/importer.go`：

```go
type ImportOptions struct {
    Format string
}

type Importer interface {
    Name() string
    Extensions() []string
    Import(data []byte, options *ImportOptions) (*Schema, error)
}

func RegisterImporter(importer Importer)
func Import(format string, data []byte, options *ImportOptions) (*Schema, error)
func ImportFile(path string, options *ImportOptions) (*Schema, error)
```

兼容入口：

```go
func ParseFile(path string) (*Schema, error) {
    return ImportFile(path, nil)
}
```

路由规则：

1. `ImportOptions.Format` 非空时优先按显式 format 选择 importer。
2. format 为空时按文件扩展名选择 importer。
3. `.bmsg` 保留兼容行为：先尝试 JSON，再尝试 YAML，最后尝试 legacy BMSG DSL。
4. 未识别扩展名时，可以保留当前行为：尝试 JSON、YAML、BMSG fallback。

### 4.3 Export 层接口建议

新增 `pkg/exporter/exporter.go` 或拆分 registry 文件：

```go
type ExportOptions struct {
    Format string
    Name   string
}

type Exporter interface {
    Name() string
    Extensions() []string
    Export(s *schema.Schema, options *ExportOptions) ([]byte, error)
}

func RegisterExporter(exporter Exporter)
func Export(format string, s *schema.Schema, options *ExportOptions) ([]byte, error)
```

迁移现有函数：

- `Markdown(s)` 保留，注册为 `md` / `markdown`。
- `HTML(s)` 保留，注册为 `html`。
- `Bmsg(s)` 保留，注册为 `bmsg`。
- 新增 `Proto(s)`，注册为 `proto`。

CLI `export` 不再 switch 具体格式，只负责：

1. `schema.ImportFile(input)`。
2. 遍历 `--format`。
3. 调用 `exporter.Export(format, s, options)`。
4. 根据 exporter 扩展名写文件。

### 4.4 Codegen 层保持现状

`codegen` 已经是 registry 模式，本次不做架构改动。`compiler.Compile` 保持：

```text
ImportFile(input) -> codegen.Get(language) -> Generate(schema)
```

唯一变化是 `compiler.Compile` 内部的 `schema.ParseFile` 可以继续调用，也可以改成 `schema.ImportFile`。为了最小改动，建议先让 `ParseFile` 委托 `ImportFile`，`compiler.Compile` 不必立即改签名。

## 5. Proto Exporter 设计

### 5.1 输出范围

第一阶段输出 proto3 子集：

```proto
syntax = "proto3";

package example.game;

// ByteMsg233 schema: bymsg/v1
// ByteMsg233 protocolVersion: 7

enum PlayerState {
  PLAYER_STATE_UNKNOWN = 0;
  PLAYER_STATE_ACTIVE = 1;
}

// ByteMsg233 packetId: 1001
message Player {
  uint64 id = 1;
  string name = 2;
  repeated string tags = 3;
  map<string, uint32> attrs = 4;
}
```

### 5.2 类型映射

| ByteMsg233 type | Proto type | 说明 |
|---|---|---|
| `bool` | `bool` | 直接映射 |
| `int32` | `sint32` | ByteMsg233 signed 当前使用 zigzag |
| `int64` | `sint64` | ByteMsg233 signed 当前使用 zigzag |
| `uint32` | `uint32` | 直接映射 |
| `uint64` | `uint64` | 直接映射 |
| `float32` | `float` | proto3 名称 |
| `float64` | `double` | proto3 名称 |
| `string` | `string` | 直接映射 |
| `bytes` | `bytes` | 直接映射 |
| `list<T>` | `repeated T` | 列表 |
| `map<K,V>` | `map<K, V>` | 受 proto map key 限制 |
| enum | enum name | 直接引用 |
| message | message name | 直接引用 |

Proto map key 第一阶段只允许：

```text
int32, int64, uint32, uint64, bool, string
```

如果 ByteMsg233 schema 中出现 proto 不支持的 map key 类型，应返回明确错误。

### 5.3 元数据保留

`schema.Version`、`ProtocolVersion`、`Message.PacketID` 是 ByteMsg233 元数据，不应强制进入运行时热路径。第一阶段用注释保留：

```proto
// ByteMsg233 schema: bymsg/v1
// ByteMsg233 protocolVersion: 7
// ByteMsg233 packetId: 1001
```

不建议第一阶段引入 custom option，原因：

1. 需要额外 `.proto` option 定义文件。
2. 会增加跨工具兼容成本。
3. 当前目标是元数据互导，不是 Protobuf runtime 兼容增强。

## 6. Proto Importer 设计

### 6.1 支持范围

第一阶段支持：

- `syntax = "proto3";`
- `package`
- `enum`
- `message`
- scalar field
- `repeated T`
- `map<K, V>`
- ByteMsg233 专用注释：
  - `// ByteMsg233 schema: ...`
  - `// ByteMsg233 protocolVersion: ...`
  - `// ByteMsg233 packetId: ...`

第一阶段不支持：

- `service` / `rpc`
- `oneof`
- `reserved`
- `extensions`
- 跨文件 `import` 解析
- custom option
- proto2 required/optional 语义

遇到不支持的语法时返回明确错误，不静默丢字段。

### 6.2 解析策略

优先实现轻量 parser，而不是直接引入完整 protoc 依赖：

1. 先做词法扫描，识别 identifier、number、string、symbol、comment。
2. 保留字段前连续注释，用于识别 ByteMsg233 元数据。
3. 解析 `syntax`、`package`、`enum`、`message`。
4. 解析字段：
   - `type name = tag;`
   - `repeated type name = tag;`
   - `map<key, value> name = tag;`
5. 生成统一 `schema.Schema` 后调用现有 validate。

理由：

- 当前只需要 proto3 子集。
- 项目已经有 `.bmsg` 轻量 parser 风格。
- 避免为了第一阶段 `.proto` 互导引入复杂 protoc 编译链。

后续如果要支持完整 proto 生态，再评估 `protocompile` 或 descriptor 路线。

## 7. CLI 行为

### 7.1 编译 proto 到目标语言

保持当前 compile 命令形态：

```bash
bytemsg233 compile protocol.proto -l go -o ./gen/go
```

内部流程：

```text
protocol.proto -> schema.ImportFile -> schema.Schema -> codegen.Go -> ByteMsg233_Export.go
```

### 7.2 导出 proto

扩展 export 格式：

```bash
bytemsg233 export protocol.bmsg.json -f proto -o ./protocol
```

多格式导出：

```bash
bytemsg233 export protocol.bmsg.json -f md,html,bmsg,proto -o ./protocol
```

输出文件：

```text
protocol.md
protocol.html
protocol.bmsg
protocol.proto
```

## 8. 实施计划

### 阶段一：Import registry

1. 新增 importer 接口和 registry。
2. 将 JSON/YAML/BMSG/TOML 包装为 importer。
3. `ParseFile` 改为调用 `ImportFile`。
4. 保持现有 schema 测试全部通过。

验证：

```bash
go test ./pkg/schema
go test ./pkg/compiler
```

### 阶段二：Export registry

1. 新增 exporter 接口和 registry。
2. 将 md/html/bmsg 包装为 exporter。
3. CLI `export` 改为 exporter registry 路由。
4. 保持现有导出行为不变。

验证：

```bash
go test ./pkg/exporter ./cmd/bytemsg233
```

如果 `cmd/bytemsg233` 当前没有测试，可补最小 exporter registry 单测，不强行做 CLI 端到端测试。

### 阶段三：Proto exporter

1. 新增 `exporter.Proto(s)`。
2. 实现类型映射、message、enum、package、ByteMsg233 元数据注释。
3. 补稳定输出测试。
4. 补不支持 map key 类型的错误测试。

验证：

```bash
go test ./pkg/exporter
```

### 阶段四：Proto importer

1. 新增 `pkg/schema/proto_parser.go`。
2. 支持 proto3 子集解析。
3. 支持 ByteMsg233 元数据注释恢复。
4. 补 `.proto` fixture。
5. 补 `ParseFile` / `ImportFile` 测试。

验证：

```bash
go test ./pkg/schema
```

### 阶段五：端到端链路

1. 补 `protocol.proto -> schema.Schema -> Go` 编译测试。
2. 补 `schema -> proto -> schema` round-trip 测试。
3. 运行相关包测试。

验证：

```bash
go test ./pkg/schema ./pkg/exporter ./pkg/compiler ./pkg/codegen/go
```

## 9. 兼容性与迁移策略

1. `schema.ParseFile` 保留，外部调用不破坏。
2. `exporter.Markdown`、`exporter.HTML`、`exporter.Bmsg` 保留，避免破坏现有包级调用。
3. CLI 默认行为不变：
   - `compile` 默认语言仍为 Go。
   - `export` 默认格式仍为 `md,html,bmsg`，是否把 `proto` 加入默认值需要单独决策。
4. `.bmsg` 兼容解析顺序保持不变。
5. `.proto` 第一阶段作为元数据交换格式，不承诺完整 Protobuf 语义兼容。

## 10. 风险与处理

| 风险 | 影响 | 处理 |
|---|---|---|
| proto importer 子集过小 | 用户导入复杂 proto 失败 | 错误信息明确列出不支持语法 |
| proto exporter 类型映射有歧义 | round-trip 后类型变化 | signed 类型导出为 `sint32/sint64`，测试覆盖 |
| 注释承载 ByteMsg233 元数据不够强 | 第三方工具可能删除注释 | 第一阶段接受；后续再评估 custom option |
| exporter registry 改动影响 CLI 输出 | 现有 export 行为变化 | 保留默认格式和文件名规则，补测试 |
| `.bmsg` fallback 行为变化 | 老文件解析失败 | `.bmsg` importer 明确保留 JSON/YAML/BMSG 顺序 |

## 11. 验证清单

必须通过：

```bash
go test ./pkg/schema ./pkg/exporter ./pkg/compiler ./pkg/codegen/go
```

建议补充：

```bash
go test ./...
```

如果本机 Go cache 权限异常，可使用仓库内临时 cache：

```bash
$env:GOCACHE="D:\Projects\Goland\bytemsg233\.cache\go-build"
go test ./pkg/schema ./pkg/exporter ./pkg/compiler ./pkg/codegen/go
```

## 12. 结论

本方案建议把 ByteMsg233 元数据处理正式拆成三层：

```text
schema.Import -> schema.Schema -> exporter.Export / codegen
```

短期最小闭环是先补 `.proto` importer/exporter，让 `.proto` 成为 ByteMsg233 的元数据交换格式之一，同时保持 `compiler.Compile` 外部调用和现有 `.bmsg.json` 主路径稳定。实体类反向解析、完整 Protobuf 语义兼容、binary buffer pool 并发优化均不进入本阶段。
