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
| Battle input, 10 players | 247 B | 266 B | 1097 B | 931 B |
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
| `libs/go` | https://github.com/neko233-com/bytemsg233-go |
| `libs/csharp` | https://github.com/neko233-com/bytemsg233-csharp |
| `libs/typescript` | https://github.com/neko233-com/bytemsg233-typescript |
| `libs/rust` | https://github.com/neko233-com/bytemsg233-rust |
| `libs/java` | https://github.com/neko233-com/bytemsg233-java |
| `libs/cpp` | https://github.com/neko233-com/bytemsg233-cpp |
| `libs/c` | https://github.com/neko233-com/bytemsg233-c |
| `libs/kotlin` | https://github.com/neko233-com/bytemsg233-kotlin |
| `libs/swift` | https://github.com/neko233-com/bytemsg233-swift |
| `libs/dart` | https://github.com/neko233-com/bytemsg233-dart |
| `libs/lua` | https://github.com/neko233-com/bytemsg233-lua |
| `editors/vscode` | https://github.com/neko233-com/bytemsg233-plugin-vscode |
| `editors/jetbrains` | https://github.com/neko233-com/bytemsg233-plugin-jetbrains |

## Develop

```bash
go test ./...
```

MIT License.
