<p align="center">
  <img src="docs/assets/bytemsg233-logo-192.png" alt="ByteMsg233 logo" width="128" height="128">
</p>

<p align="center">
  <a href="#readme">English</a> | <a href="#中文">中文</a>
</p>

# ByteMsg233

## Install

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/neko233-com/bytemsg233/main/scripts/install.sh | bash

# Windows PowerShell
irm https://raw.githubusercontent.com/neko233-com/bytemsg233/main/scripts/install.ps1 | iex

# Go toolchain
go install github.com/neko233-com/bytemsg233/cmd/bytemsg233@latest
```

```bash
bytemsg233 version
```

**ByteMsg233** is a JSON-first binary schema toolchain for games, clients, and SDKs. It uses `.bmsg.json` as the protocol description DSL, generates native-feeling code, and keeps runtime libraries copyable into real projects even when package registry publishing is not available.

The short version: JSON replaces `.proto`; generated code should not feel like Protobuf.

Performance posture is deliberately practical: compact binary payloads, high-throughput generated encode/decode, very low memory churn, and 0-GC hot paths when callers use preallocated buffers plus prewarmed pools. Repeated game and business DTO workloads are the main optimization target.

## Naming Standard

The product and protocol brand is **ByteMsg233**. Documentation, generated protocol docs, benchmarks, UI labels, and release copy must use `ByteMsg233` as the display name. The lowercase `bytemsg233` spelling is reserved for the CLI command, Go module path, package names, repository names, and file paths.

## Quick Start

```bash
bytemsg233 init game

bytemsg233 compile game.bmsg.json \
  -l go,csharp,typescript,rust,java,cpp,c \
  -o ./gen

bytemsg233 export game.bmsg.json -f md,html,bmsg -o ./protocol
```

Install a runtime library by copying it into your project:

```bash
# Unity
bytemsg233 install-lib csharp --to ./Assets/Plugins/ByteMsg233

# Web / frontend tooling
bytemsg233 install-lib typescript --to ./vendor/bytemsg233

# Go / Rust / Java
bytemsg233 install-lib go --to ./third_party/bytemsg233
bytemsg233 install-lib rust --to ./vendor/bytemsg233
bytemsg233 install-lib java --to ./libs/bytemsg233

# Native roadmap targets
bytemsg233 install-lib cpp --to ./third_party/bytemsg233
bytemsg233 install-lib c --to ./third_party/bytemsg233
```

## Native JSON DSL

Top-level JSON keys are message names. Reserved keys such as `schema`, `package`, and `enums` describe the protocol.

```json
{
  "schema": "bymsg/v1",
  "protocolVersion": 7,
  "package": "com.example.game",
  "enums": {
    "HeroState": ["IDLE", "MOVING", "ATTACKING", "DEAD"]
  },
  "Hero": {
    "packetId": 1001,
    "comment": "Hero profile",
    "id": {
      "type": "uint32",
      "comment": "Hero ID"
    },
    "name": {
      "type": "string",
      "comment": "Hero name"
    },
    "skill_ids": { "list": "uint32", "comment": "Skill IDs" },
    "attrs": { "map": ["string", "uint32"], "comment": "Attributes" },
    "state": "HeroState"
  }
}
```

Only three things matter for fields: field name, type, and optional comment. `tag` is optional; when omitted, ByteMsg233 assigns tags from JSON field order. `packetId` is optional and belongs on the message, which matches game protocol routing.

`protocolVersion` is the business protocol version used for server/client compatibility checks. Generated code exports `ByteMsgProtocolVersion`, so business code can read the schema/runtime version from the generated package instead of putting a version on every message. Recommended game networking flow: send a `ProtocolHello(version, minCompatible)` once when the socket/session opens, reject mismatches early, and keep per-message hot paths free of a repeated version header. If a gateway truly needs mixed protocol versions on one connection, put the version in the business envelope instead. Content fingerprints are optional business data, not a runtime-enforced ByteMsg233 default.

Generated readers and runtimes skip unknown fields for supported wire types, so adding a new field does not break older readers. Unknown fields are dropped when an old process re-encodes unless your business layer explicitly preserves them.

Binary-capable generated targets expose package-level protocol version helpers such as `GetByteMsg233ProtocolVersion()` or language-native equivalents. The Go fast path also exposes `IByteMsg233Api` with `SerializeByteMsg233()` and `DeserializeFromByteMsg233(data)` as a thin interface-friendly wrapper; performance-critical game code can still call the lower-level append/decode APIs directly.

Enums, lists, maps, and comments are first-class:

- Enums can be `["IDLE", "MOVING"]`, `{ "OK": 0, "TIMEOUT": 10 }`, or `{ "comment": "...", "values": [...] }`.
- Lists can be `"list<uint32>"` or `{ "list": "uint32", "comment": "..." }`.
- Maps can be `"map<string, uint32>"`, `{ "map": ["string", "uint32"] }`, or `{ "map": { "key": "string", "value": "uint32" } }`.
- Comments can be the short `comment` string or localized `description.zh` / `description.en`.

For complex data, reference another message class by name instead of nesting message structures inline. That keeps packets readable and keeps generated classes pool-friendly.

YAML is still supported for teams that prefer it, and legacy `.bmsg` can be exported for future tooling experiments. The main authoring path is `.bmsg.json`.

Runtime libraries and generated encode/decode hot paths are single-threaded by design. Pools use caller-owned or local stack/list storage only: no locks, no concurrent collections, no channels, no goroutines, no thread pools, and no hidden background work.

## Generated API Shape

| Language | Status | Runtime path | Pooling API | Enum style |
|---|---|---|---|---|
| Go | priority | `libs/go` | `AcquireHero()` / `ReleaseHero()` / `Reset()` | typed consts, parse helpers |
| C# / Unity | priority | `libs/csharp` | `new Hero()` / `Hero.Prewarm()` / `Hero.Rent()` / `Hero.Return()` / `Release()` | native `enum` helpers |
| TypeScript / JavaScript | priority | `libs/typescript` | `Hero.acquire()` / `release()` | `enum` plus JS runtime |
| Rust | priority | `libs/rust` | `ByteMsgPool<T>` | `enum`, `from_value()` |
| Java | priority | `libs/java` | `Hero.acquire()` / `release()` | enum instances |
| C++ | planned | `libs/cpp` | arena/object pool rent-return | scoped enum |
| C | planned | `libs/c` | caller-owned pool/context | generated constants |
| Kotlin | planned | `libs/kotlin` | companion acquire/release | enum class |
| Swift | planned | `libs/swift` | pool rent-return | enum |
| Dart / Flutter | planned | `libs/dart` | `Hero.acquire()` / `release()` | enum |
| Lua | planned | `libs/lua` | table pool acquire/release | constants/table |
| Python | supported | generated code | `Hero.acquire()` / `release()` | `IntEnum` |

Official runtime repositories are tracked in [docs/LANGUAGES.md](docs/LANGUAGES.md).

Go generated messages also support debug text output for tools, logs, and fixture inspection:

```go
text := hero.ByteMsgText()
dst := hero.AppendByteMsgText(buf[:0])
```

Debug text uses schema field names and is intentionally outside the binary protocol. Do not use it as a wire format. Hot-path gameplay/network code should stay fully binary and use caller-owned buffers.

C# / Unity generated messages are `partial class` by default. Keep custom gameplay helpers in separate partial files; regenerated protocol code can then be replaced safely. Use `Prewarm` during loading to make pool-backed `Rent` / `Return` flows allocation-free in gameplay hot paths.

Single-file exports use `ByteMsg233_Export` by default, for example `ByteMsg233_Export.go`, `ByteMsg233_Export.cs`, and `ByteMsg233_Export.ts`.

## Export Protocol Docs

```bash
bytemsg233 export game.bmsg.json -f md,html,bmsg -o ./protocol
```

This writes:

- `game.md`: human-readable protocol documentation for client/server integration.
- `game.html`: standalone user-facing protocol page.
- `game.bmsg`: legacy/extension-friendly DSL export.

Markdown is for people. HTML demo pages are standalone. Neither should depend on the other as a required reading path.

## Copy-Based Runtime Install

Package registries are convenient until they block a release. Every priority runtime is kept as a Git submodule under `libs/`, and `install-lib` copies the source into a target project.

```bash
git submodule update --init --recursive
bytemsg233 install-lib csharp --to ../MyUnityProject/Assets/Plugins/ByteMsg233
```

The copy command intentionally skips `.git`, `node_modules`, `build`, `dist`, and `target`.

## Performance Snapshot

Read this table as a practical game/client baseline, not as a tiny-object trick. It includes small packets, repeated DTOs, battle input, and 100-row ranking data. Bigger repeated structures are where binary protocols show their real shape.

| Scenario | ByteMsg233 | Protobuf | JSON payload | MessagePack |
|---|---:|---:|---:|---:|
| Player profile, 10 fields | 61 B | 61 B | 173 B | 155 B |
| Chat message, 5 fields | 57 B | 57 B | 116 B | 103 B |
| ChatDto all types | 304 B | 316 B | 647 B | 531 B |
| RPC envelope + ChatDto payload, 1x | 316 B | 328 B | 928 B | 597 B |
| Battle input, 10 players | 130 B | 266 B | 1097 B | 931 B |
| TaskDto list, 100 rows | 2261 B | 4044 B | 14691 B | 13303 B |
| Leaderboard, 100 rows | 2518 B | 3608 B | 9602 B | 8711 B |

Game-specific benchmark coverage also includes login/full-state pushes and realtime battle frames.

Run locally:

```bash
go test ./pkg/binary/... -bench="Benchmark(Encode|Decode)_" -benchmem
go test ./pkg/binary/... -bench="Benchmark(Encode|Decode)_ChatDtoAllTypes" -benchmem
go test ./pkg/binary/... -run "TestBenchmark_SizeComparison" -v
go test ./pkg/binary/... -run "TestGame_" -v
```

Run the full benchmark and runtime verification suite in one Docker image:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/bench-docker.ps1
```

This runs Go codec benchmarks against Protobuf/JSON/MessagePack, game benchmarks, TypeScript tests, Rust tests, C# runtime tests, and Java 17 `javac` checks for runtime plus generated code. Logs are written under `bench-results/`.

Java runtime compile and smoke checks are intentionally easy:

```bash
powershell -ExecutionPolicy Bypass -File scripts/test-java.ps1
```

The script uses local JDK 17 when available, or a pre-pulled `eclipse-temurin:17-jdk` Docker image.

Full notes: [docs/BENCHMARK.md](docs/BENCHMARK.md). Game packet design: [docs/GAME_BINARY.md](docs/GAME_BINARY.md).

## Repositories

| Path | Repository |
|---|---|
| `libs/go` | https://github.com/neko233-com/bytemsg233-lib-go |
| `libs/csharp` | https://github.com/neko233-com/bytemsg233-lib-csharp |
| `libs/typescript` | https://github.com/neko233-com/bytemsg233-lib-typescript |
| `libs/rust` | https://github.com/neko233-com/bytemsg233-lib-rust |
| `libs/java` | https://github.com/neko233-com/bytemsg233-lib-java |
| `libs/cpp` | https://github.com/neko233-com/bytemsg233-lib-cpp |
| `libs/c` | https://github.com/neko233-com/bytemsg233-lib-c |
| `libs/kotlin` | https://github.com/neko233-com/bytemsg233-lib-kotlin |
| `libs/swift` | https://github.com/neko233-com/bytemsg233-lib-swift |
| `libs/dart` | https://github.com/neko233-com/bytemsg233-lib-dart |
| `libs/lua` | https://github.com/neko233-com/bytemsg233-lib-lua |
| `editors/vscode` | https://github.com/neko233-com/bytemsg233-plugin-vscode |
| `editors/jetbrains` | https://github.com/neko233-com/bytemsg233-plugin-jetbrains |

## Develop

```bash
go test ./...
```

MIT License.

---

# 中文

## 安装

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/neko233-com/bytemsg233/main/scripts/install.sh | bash

# Windows PowerShell
irm https://raw.githubusercontent.com/neko233-com/bytemsg233/main/scripts/install.ps1 | iex

# Go 工具链
go install github.com/neko233-com/bytemsg233/cmd/bytemsg233@latest
```

```bash
bytemsg233 version
```

**ByteMsg233** 是面向游戏、客户端和 SDK 的 JSON 优先二进制协议工具链。使用 `.bmsg.json` 作为协议描述 DSL，生成具有原生风格的代码，运行时库可直接复制到项目中使用（无需依赖包管理器发布）。

简而言之：JSON 替代 `.proto`，生成的代码不应像 Protobuf 那样生硬。

性能设计注重实用性：紧凑的二进制载荷、高吞吐量的编解码、极低的内存抖动，当调用者使用预分配缓冲区和预热对象池时可实现 0-GC 热路径。重复的游戏和业务 DTO 负载是主要优化目标。

## 命名规范

产品和协议品牌为 **ByteMsg233**。文档、协议文档、基准测试、UI 标签和发布文案必须使用 `ByteMsg233` 作为显示名称。小写 `bytemsg233` 保留给 CLI 命令、Go 模块路径、包名、仓库名和文件路径。

## 快速开始

```bash
bytemsg233 init game

bytemsg233 compile game.bmsg.json \
  -l go,csharp,typescript,rust,java,cpp,c \
  -o ./gen

bytemsg233 export game.bmsg.json -f md,html,bmsg -o ./protocol
```

通过复制安装运行时库：

```bash
# Unity
bytemsg233 install-lib csharp --to ./Assets/Plugins/ByteMsg233

# Web / 前端工具
bytemsg233 install-lib typescript --to ./vendor/bytemsg233

# Go / Rust / Java
bytemsg233 install-lib go --to ./third_party/bytemsg233
bytemsg233 install-lib rust --to ./vendor/bytemsg233
bytemsg233 install-lib java --to ./libs/bytemsg233

# Native 路线图目标
bytemsg233 install-lib cpp --to ./third_party/bytemsg233
bytemsg233 install-lib c --to ./third_party/bytemsg233
```

## 原生 JSON DSL

顶层 JSON 键为消息名称。保留键如 `schema`、`package` 和 `enums` 用于描述协议。

```json
{
  "schema": "bymsg/v1",
  "protocolVersion": 7,
  "package": "com.example.game",
  "enums": {
    "HeroState": ["IDLE", "MOVING", "ATTACKING", "DEAD"]
  },
  "Hero": {
    "packetId": 1001,
    "comment": "英雄档案",
    "id": {
      "type": "uint32",
      "comment": "英雄 ID"
    },
    "name": {
      "type": "string",
      "comment": "英雄名称"
    },
    "skill_ids": { "list": "uint32", "comment": "技能 ID" },
    "attrs": { "map": ["string", "uint32"], "comment": "属性" },
    "state": "HeroState"
  }
}
```

字段只需关注三件事：字段名、类型和可选注释。`tag` 是可选的；省略时 ByteMsg233 按 JSON 字段顺序分配 tag。`packetId` 是可选的，属于消息级别，符合游戏协议路由习惯。

`protocolVersion` 是用于服务端/客户端兼容性检查的业务协议版本。生成的代码导出 `ByteMsgProtocolVersion`，业务代码可从生成的包中读取 schema/运行时版本，无需在每条消息上附加版本号。推荐的游戏网络流程：在 socket/会话打开时发送一次 `ProtocolHello(version, minCompatible)`，尽早拒绝不兼容连接，保持每条消息热路径无重复版本头。如果网关确实需要在一条连接上混合协议版本，应将版本放在业务信封中。内容指纹是可选的业务数据，不是运行时强制的 ByteMsg233 默认行为。

生成的读取器和运行时会跳过支持的线类型的未知字段，因此添加新字段不会破坏旧读取器。旧进程重新编码时会丢弃未知字段，除非业务层显式保留它们。

支持二进制的生成目标会导出包级别的协议版本辅助函数，如 `GetByteMsg233ProtocolVersion()` 或语言原生等效物。Go 快速路径还导出 `IByteMsg233Api`，包含 `SerializeByteMsg233()` 和 `DeserializeFromByteMsg233(data)` 作为轻量接口封装；性能关键的游戏代码仍可直接调用底层 append/decode API。

枚举、列表、映射和注释都是一等公民：

- 枚举可以是 `["IDLE", "MOVING"]`、`{ "OK": 0, "TIMEOUT": 10 }` 或 `{ "comment": "...", "values": [...] }`。
- 列表可以是 `"list<uint32>"` 或 `{ "list": "uint32", "comment": "..." }`。
- 映射可以是 `"map<string, uint32>"`、`{ "map": ["string", "uint32"] }` 或 `{ "map": { "key": "string", "value": "uint32" } }`。
- 注释可以是简短的 `comment` 字符串或本地化的 `description.zh` / `description.en`。

对于复杂数据，通过名称引用另一个消息类，而不是内联嵌套消息结构。这使数据包可读且生成的类对对象池友好。

仍支持 YAML（适合偏好 YAML 的团队），旧版 `.bmsg` 可导出用于未来工具实验。主要创作路径是 `.bmsg.json`。

运行时库和生成的编解码热路径在设计上是单线程的。对象池仅使用调用者拥有或本地栈/列表存储：无锁、无并发集合、无 channel、无 goroutine、无线程池、无隐藏后台任务。

## 生成 API 形状

| 语言 | 状态 | 运行时路径 | 对象池 API | 枚举风格 |
|---|---|---|---|---|
| Go | 优先 | `libs/go` | `AcquireHero()` / `ReleaseHero()` / `Reset()` | 类型化 const，解析辅助 |
| C# / Unity | 优先 | `libs/csharp` | `new Hero()` / `Hero.Prewarm()` / `Hero.Rent()` / `Hero.Return()` / `Release()` | 原生 `enum` 辅助 |
| TypeScript / JavaScript | 优先 | `libs/typescript` | `Hero.acquire()` / `release()` | `enum` + JS 运行时 |
| Rust | 优先 | `libs/rust` | `ByteMsgPool<T>` | `enum`，`from_value()` |
| Java | 优先 | `libs/java` | `Hero.acquire()` / `release()` | 枚举实例 |
| C++ | 计划中 | `libs/cpp` | arena/对象池 rent-return | 作用域枚举 |
| C | 计划中 | `libs/c` | 调用者持有 pool/context | 生成常量 |
| Kotlin | 计划中 | `libs/kotlin` | companion acquire/release | enum class |
| Swift | 计划中 | `libs/swift` | pool rent-return | enum |
| Dart / Flutter | 计划中 | `libs/dart` | `Hero.acquire()` / `release()` | enum |
| Lua | 计划中 | `libs/lua` | table pool acquire/release | 常量/table |
| Python | 已支持 | 生成代码 | `Hero.acquire()` / `release()` | `IntEnum` |

官方运行时仓库列表见 [docs/LANGUAGES.md](docs/LANGUAGES.md)。

Go 生成消息还支持调试文本输出，用于工具、日志和测试检查：

```go
text := hero.ByteMsgText()
dst := hero.AppendByteMsgText(buf[:0])
```

调试文本使用 schema 字段名，故意在二进制协议之外。不要将其用作线路格式。热路径游戏/网络代码应保持完全二进制并使用调用者拥有的缓冲区。

C# / Unity 生成消息默认为 `partial class`。在单独的 partial 文件中保持自定义游戏辅助方法；重新生成的协议代码可安全替换。在加载期间使用 `Prewarm` 使基于对象池的 `Rent` / `Return` 流程在游戏热路径中无分配。

单文件导出默认使用 `ByteMsg233_Export`，例如 `ByteMsg233_Export.go`、`ByteMsg233_Export.cs` 和 `ByteMsg233_Export.ts`。

## 导出协议文档

```bash
bytemsg233 export game.bmsg.json -f md,html,bmsg -o ./protocol
```

输出内容：

- `game.md`：人类可读的协议文档，用于客户端/服务端集成。
- `game.html`：独立的用户协议页面。
- `game.bmsg`：旧版/扩展友好的 DSL 导出。

Markdown 面向开发者。HTML 演示页面独立运行。两者不依赖对方作为必读路径。

## 复制式运行时安装

包管理器在阻塞发布时会很不便。每个优先运行时都作为 Git 子模块保存在 `libs/` 下，`install-lib` 可将源码复制到目标项目中。

```bash
git submodule update --init --recursive
bytemsg233 install-lib csharp --to ../MyUnityProject/Assets/Plugins/ByteMsg233
```

复制命令会跳过 `.git`、`node_modules`、`build`、`dist` 和 `target`。

## 性能快照

将此表作为实际游戏/客户端基线，而非微小对象测试。包含小数据包、重复 DTO、战斗输入和 100 行排行数据。更大的重复结构才能体现二进制协议的真正优势。

| 场景 | ByteMsg233 | Protobuf | JSON 载荷 | MessagePack |
|---|---:|---:|---:|---:|
| 玩家档案，10 字段 | 61 B | 61 B | 173 B | 155 B |
| 聊天消息，5 字段 | 57 B | 57 B | 116 B | 103 B |
| ChatDto 全类型 | 304 B | 316 B | 647 B | 531 B |
| RPC 信封 + ChatDto 载荷，1x | 316 B | 328 B | 928 B | 597 B |
| 战斗输入，10 玩家 | 130 B | 266 B | 1097 B | 931 B |
| 任务列表，100 行 | 2261 B | 4044 B | 14691 B | 13303 B |
| 排行榜，100 行 | 2518 B | 3608 B | 9602 B | 8711 B |

游戏基准测试还包括登录/全状态推送和实时战斗帧。

本地运行：

```bash
go test ./pkg/binary/... -bench="Benchmark(Encode|Decode)_" -benchmem
go test ./pkg/binary/... -bench="Benchmark(Encode|Decode)_ChatDtoAllTypes" -benchmem
go test ./pkg/binary/... -run "TestBenchmark_SizeComparison" -v
go test ./pkg/binary/... -run "TestGame_" -v
```

在单个 Docker 镜像中运行完整基准测试和运行时验证套件：

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/bench-docker.ps1
```

该脚本会对 Protobuf/JSON/MessagePack 运行 Go 编解码基准测试、游戏基准测试、TypeScript 测试、Rust 测试、C# 运行时测试和 Java 17 `javac` 检查（运行时加生成代码）。日志写入 `bench-results/`。

Java 运行时编译和冒烟测试设计简单：

```bash
powershell -ExecutionPolicy Bypass -File scripts/test-java.ps1
```

脚本优先使用本地 JDK 17，否则使用预拉取的 `eclipse-temurin:17-jdk` Docker 镜像。

完整文档：[docs/BENCHMARK.md](docs/BENCHMARK.md)。游戏包设计：[docs/GAME_BINARY.md](docs/GAME_BINARY.md)。

## 仓库列表

| 路径 | 仓库 |
|---|---|
| `libs/go` | https://github.com/neko233-com/bytemsg233-lib-go |
| `libs/csharp` | https://github.com/neko233-com/bytemsg233-lib-csharp |
| `libs/typescript` | https://github.com/neko233-com/bytemsg233-lib-typescript |
| `libs/rust` | https://github.com/neko233-com/bytemsg233-lib-rust |
| `libs/java` | https://github.com/neko233-com/bytemsg233-lib-java |
| `libs/cpp` | https://github.com/neko233-com/bytemsg233-lib-cpp |
| `libs/c` | https://github.com/neko233-com/bytemsg233-lib-c |
| `libs/kotlin` | https://github.com/neko233-com/bytemsg233-lib-kotlin |
| `libs/swift` | https://github.com/neko233-com/bytemsg233-lib-swift |
| `libs/dart` | https://github.com/neko233-com/bytemsg233-lib-dart |
| `libs/lua` | https://github.com/neko233-com/bytemsg233-lib-lua |
| `editors/vscode` | https://github.com/neko233-com/bytemsg233-plugin-vscode |
| `editors/jetbrains` | https://github.com/neko233-com/bytemsg233-plugin-jetbrains |

## 开发

```bash
go test ./...
```

MIT 许可证。
