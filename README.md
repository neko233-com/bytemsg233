# bytemsg233

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

`bytemsg233` is a JSON-first binary schema toolchain. It uses `.bmsg.json` as the protocol description DSL, generates native-feeling code, and keeps runtime libraries copyable into real projects even when package registry publishing is not available.

The short version: JSON replaces `.proto`; generated code should not feel like Protobuf.

## Quick Start

```bash
bytemsg233 init game

bytemsg233 compile game.bmsg.json \
  -l go,csharp,typescript,rust,java \
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

Only three things matter for fields: field name, type, and optional comment. `tag` is optional; when omitted, bytemsg233 assigns tags from JSON field order. `packetId` is optional and belongs on the message, which matches game protocol routing.

Enums, lists, maps, and comments are first-class:

- Enums can be `["IDLE", "MOVING"]`, `{ "OK": 0, "TIMEOUT": 10 }`, or `{ "comment": "...", "values": [...] }`.
- Lists can be `"list<uint32>"` or `{ "list": "uint32", "comment": "..." }`.
- Maps can be `"map<string, uint32>"`, `{ "map": ["string", "uint32"] }`, or `{ "map": { "key": "string", "value": "uint32" } }`.
- Comments can be the short `comment` string or localized `description.zh` / `description.en`.

For complex data, reference another message class by name instead of nesting message structures inline. That keeps packets readable and keeps generated classes pool-friendly.

YAML is still supported for teams that prefer it, and legacy `.bmsg` can be exported for future tooling experiments. The main authoring path is `.bmsg.json`.

## Generated API Shape

| Language | Status | Runtime path | Pooling API | Enum style |
|---|---|---|---|---|
| Go | priority | `libs/go` | `AcquireHero()` / `ReleaseHero()` / `Reset()` | typed consts, parse helpers |
| C# / Unity | priority | `libs/csharp` | `Hero.Rent()` / `Hero.Return()` / `Release()` | native `enum` helpers |
| TypeScript / JavaScript | priority | `libs/typescript` | `Hero.acquire()` / `release()` | `enum` plus JS runtime |
| Rust | priority | `libs/rust` | `ByteMsgPool<T>` | `enum`, `from_value()` |
| Java | priority | `libs/java` | `Hero.acquire()` / `release()` | enum instances |
| Python | supported | generated code | `Hero.acquire()` / `release()` | `IntEnum` |

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

| Scenario | bytemsg233 | Protobuf | MessagePack | JSON payload |
|---|---:|---:|---:|---:|
| Player profile, 10 fields | 61 B | 61 B | 155 B | 173 B |
| Chat message, 5 fields | 57 B | 57 B | 103 B | 116 B |
| Battle input, 10 players | 247 B | 266 B | 931 B | 1,097 B |
| Leaderboard, 100 rows | 3,409 B | 3,608 B | 8,711 B | 9,602 B |

Run locally:

```bash
go test ./pkg/binary/... -bench="Benchmark(Encode|Decode)_" -benchmem
go test ./pkg/binary/... -run "TestBenchmark_SizeComparison" -v
```

Full notes: [docs/BENCHMARK.md](docs/BENCHMARK.md).

## Repositories

| Path | Repository |
|---|---|
| `libs/go` | https://github.com/neko233-com/bytemsg233-lib-go |
| `libs/csharp` | https://github.com/neko233-com/bytemsg233-lib-csharp |
| `libs/typescript` | https://github.com/neko233-com/bytemsg233-lib-typescript |
| `libs/rust` | https://github.com/neko233-com/bytemsg233-lib-rust |
| `libs/java` | https://github.com/neko233-com/bytemsg233-lib-java |
| `editors/vscode` | https://github.com/neko233-com/bytemsg233-plugin-vscode |
| `editors/jetbrains` | https://github.com/neko233-com/bytemsg233-plugin-jetbrains |

## Develop

```bash
go test ./...
```

MIT License.
