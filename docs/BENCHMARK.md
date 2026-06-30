<p align="center">
  <img src="assets/bytemsg233-logo-192.png" alt="ByteMsg233 logo" width="96" height="96">
</p>

# Performance

> Go benchmark snapshot · Windows amd64 · AMD Ryzen 9 7900X3D

This page explains what the numbers mean, not just who wins a table.

ByteMsg233 is optimized for schema-driven game and client traffic: small field headers, dense column blocks for repeated DTOs, packed varint columns, delta varint columns, bool bitsets, no repeated field names, and generated APIs that can reuse memory. The target is a practical protocol: readable schema, compact packets, native generated code, high-throughput encode/decode on repeated DTO payloads, and a zero-GC hot path where the caller provides buffers and pools are prewarmed.

The headline is straightforward: the Go fast path now beats the Protobuf wire helper baselines in every Protobuf comparison in this suite, while keeping allocation pressure low and repeated client payloads compact.

Quick percentage view: ByteMsg233 payload size in this snapshot is 49%~100% of Protobuf, 12%~49% of JSON, and 14%~57% of MessagePack. In plain terms, more repeated data means more savings.

## How To Read The Tables

Lower is better for size, `ns/op`, `B/op`, and `allocs/op`.

Comparison order is fixed:

1. ByteMsg233
2. Protobuf
3. JSON
4. Optional extra codecs, such as MessagePack

JSON is included because many teams start there. MessagePack is included because it is a common "binary JSON" baseline. Protobuf is included because it is the obvious mature competitor.

## Payload Size

Payload size matters most when the same shape repeats: rankings, inventory rows, battle inputs, quest lists, mail lists, and state snapshots.

| Scenario | ByteMsg233 | Protobuf | JSON | MessagePack |
|---|---:|---:|---:|---:|
| Player profile, 10 fields | **61 B** | 61 B | 173 B | 155 B |
| Chat message, 5 fields | **57 B** | 57 B | 116 B | 103 B |
| ChatDto all types, list/map/custom | **304 B** | 316 B | 647 B | 531 B |
| RPC envelope + ChatDto payload, 1x | **316 B** | 328 B | 928 B | 597 B |
| Battle input, 10 players x 8 fields | **130 B** | 266 B | 1097 B | 931 B |
| TaskDto list, 100 rows x 9 fields | **2261 B** | 4044 B | 14691 B | 13303 B |
| Leaderboard, 100 rows x 6 fields | **2518 B** | 3608 B | 9602 B | 8711 B |

Savings versus other codecs:

| Scenario | vs Protobuf | vs JSON | vs MessagePack |
|---|---:|---:|---:|
| Player profile | 0% | -64.7% | -60.6% |
| Chat message | 0% | -50.9% | -44.7% |
| ChatDto all types | -3.8% | -53.0% | -42.7% |
| RPC envelope + ChatDto payload | -3.7% | -65.9% | -47.1% |
| Battle input | -51.1% | -88.1% | -86.0% |
| TaskDto list | -44.1% | -84.6% | -83.0% |
| Leaderboard | -30.2% | -73.8% | -71.1% |

## Encode Speed

Tiny packets and repeated structures both use the ByteMsg233 fast path in this snapshot.

These values are duration. Lower `ns/op` is better.

| Scenario | ByteMsg233 | Protobuf | JSON | MessagePack |
|---|---:|---:|---:|---:|
| Player profile | **32.4** | 82.0 | 255.7 | 376.8 |
| Chat message | **25.5** | 64.4 | 185.1 | 217.7 |
| ChatDto all types | **208.6** | 761.1 | 1519 | 1421 |
| RPC envelope + ChatDto payload | **288.3** | 743.3 | 2358 | 1839 |
| Battle input | **144.1** | 1223 | 1831 | 2469 |
| TaskDto list, 100 rows | **2096** | 14495 | 24156 | 30235 |
| Leaderboard | **2380** | 12628 | 16024 | 23138 |

The same benchmark view as throughput. Higher `ops/s` is better. These are single-threaded operations per second on Go/windows/amd64.

| Scenario | Codec | Encode ops/s | Decode ops/s |
|---|---|---:|---:|
| Player profile | ByteMsg233 | **30826141** | **25031289** |
| Player profile | Protobuf | 12198097 | 12637432 |
| Player profile | JSON | 3910833 | 909091 |
| Player profile | MessagePack | 2653928 | 2336449 |
| Chat message | ByteMsg233 | **39184953** | **45495905** |
| Chat message | Protobuf | 15527950 | 13347571 |
| Chat message | JSON | 5402485 | 1577038 |
| Chat message | MessagePack | 4593477 | 4251701 |
| ChatDto all types | ByteMsg233 | **4793864** | **2035002** |
| ChatDto all types | Protobuf | 1313888 | 967118 |
| ChatDto all types | JSON | 658328 | 183655 |
| ChatDto all types | MessagePack | 703730 | 565931 |
| RPC envelope + ChatDto payload | ByteMsg233 | **3468609** | **1524855** |
| RPC envelope + ChatDto payload | Protobuf | 1345352 | 930233 |
| RPC envelope + ChatDto payload | JSON | 424088 | 100827 |
| RPC envelope + ChatDto payload | MessagePack | 543774 | 484027 |
| Battle input | ByteMsg233 | **6939625** | **3872967** |
| Battle input | Protobuf | 817661 | 1603849 |
| Battle input | JSON | 546150 | 108968 |
| Battle input | MessagePack | 405022 | 301477 |
| TaskDto list, 100 rows | ByteMsg233 | **477099** | **407830** |
| TaskDto list, 100 rows | Protobuf | 68989 | 88238 |
| TaskDto list, 100 rows | JSON | 41398 | 8912 |
| TaskDto list, 100 rows | MessagePack | 33074 | 25169 |
| Leaderboard | ByteMsg233 | **420168** | **521648** |
| Leaderboard | Protobuf | 79189 | 79460 |
| Leaderboard | JSON | 62406 | 12919 |
| Leaderboard | MessagePack | 43219 | 35822 |

## Language Runtime Matrix

Every shipped runtime exposes the optimized block primitives needed by game packets. The Go implementation is the canonical cross-codec throughput suite because it contains the full Protobuf/JSON/MessagePack comparison harness. Other runtimes are verified for API parity, roundtrip behavior, and single-threaded hot-path policy.

| Language | Optimized block runtime | Verification |
|---|---|---|
| Go | packed, delta, bitset, string list, dense column, unknown-field skip, protocol hello | `go test ./...` + benchmark |
| C# / Unity | packed, delta, bitset, string list, fixed field skip, reusable `ByteMsgByteBuffer`, protocol hello | `dotnet test libs/csharp/Tests/ByteMsg233.Tests.csproj` |
| TypeScript / JavaScript | packed, delta, bitset, string list, `readBytesView`, protocol hello | `npm test` |
| Rust | packed, delta, bitset, string list, fixed field skip, protocol hello | `cargo test` |
| Java / Android | packed, delta, bitset, string list, fixed field skip, protocol hello | `scripts/test-java.ps1` smoke + `gradle test` with JDK 17 |

ChatDto relative view:

| Codec | Encode duration | Decode duration | Encode throughput | Decode throughput |
|---|---:|---:|---:|---:|
| ByteMsg233 | **0.27x Protobuf** | **0.48x Protobuf** | **3.65x Protobuf** | **2.10x Protobuf** |
| Protobuf | 3.65x ByteMsg233 | 2.10x ByteMsg233 | 0.27x ByteMsg233 | 0.48x ByteMsg233 |
| JSON | 7.28x ByteMsg233 | 5.27x Protobuf | 0.14x ByteMsg233 | 0.19x Protobuf |
| MessagePack | 6.81x ByteMsg233 | 1.71x Protobuf | 0.15x ByteMsg233 | 0.59x Protobuf |

Interpretation:

- ByteMsg233 simple DTO encode uses append helpers; repeated DTO encode uses dense column blocks with delta/packed primitive columns where the schema is stable.
- ByteMsg233 decode uses `SliceDecoder` and zero-copy string/bytes views for immutable payload buffers.
- In this suite, ByteMsg233 is faster than Protobuf for every measured Protobuf encode/decode comparison.
- JSON and MessagePack pay for dynamic object shape and field-name-heavy data.
- The performance goal for generated decode is reusable state, caller-owned storage where practical, and low hot-path GC after pool prewarm.

## RPC, Socket, And Protocol Compatibility

ByteMsg233 does not force a built-in socket frame. Games usually already have a transport envelope with packet id, sequence, flags, encryption/compression bits, and length framing. The measured `RPC envelope + ChatDto payload` row is a normal ByteMsg233 message body with `packetId`, `sequence`, `kind`, `flags`, and `payload` fields, so unknown fields can be skipped and the shape can evolve.

For protocol structure checks, use generated constants and a session handshake:

- `ByteMsgProtocolVersion`: manually controlled business protocol version from top-level `protocolVersion`, exported by generated code so callers can read it from the package/library.
- `ProtocolHello(version, minCompatible)`: send once when a socket/session opens; reject mismatches before entering the hot path.
- Optional content fingerprints are business-owned data. ByteMsg233 does not generate or enforce them by default.

The Go generated fast path exposes `IByteMsg233Api` (`SerializeByteMsg233`, `DeserializeFromByteMsg233`) for generic RPC/template glue and keeps lower-level append/decode APIs for the measured hot path. Packet ids and metadata remain schema/codegen data; the runtime benchmark only measures binary encode/decode.

This avoids adding a version integer to every gameplay packet. If one connection intentionally multiplexes multiple protocol versions, put the version in the business envelope for that gateway only.

Generated readers also skip unknown fields for supported wire types (`0` varint, `1` fixed64, `2` length-delimited, `5` fixed32). This gives additive schema rollout similar to Protobuf while keeping the default hot path single-threaded and allocation-aware.

## Decode Speed

Decode uses the slice fast path. Lower `ns/op` is better.

| Scenario | ByteMsg233 | Protobuf | JSON | MessagePack |
|---|---:|---:|---:|---:|
| Player profile | **40.0** | 79.1 | 1100 | 428.0 |
| Chat message | **22.0** | 74.9 | 634.1 | 235.2 |
| ChatDto all types | **491.4** | 1034 | 5445 | 1767 |
| RPC envelope + ChatDto payload | **655.8** | 1075 | 9918 | 2066 |
| Battle input | **258.2** | 623.5 | 9177 | 3317 |
| Leaderboard, 100 rows | **1917** | 12585 | 77404 | 27916 |
| TaskDto list, 100 rows | **2452** | 11333 | 112203 | 39731 |

## Allocations

Allocations are where game clients feel pain: a small per-packet allocation can become a frame-time spike when repeated thousands of times.

### Encode (`B/op`, `allocs/op`)

| Scenario | ByteMsg233 | Protobuf | JSON | MessagePack |
|---|---:|---:|---:|---:|
| Player profile | **64, 1** | 104, 3 | 176, 1 | 496, 4 |
| ChatDto all types | **0, 0** | 1328, 22 | 1281, 11 | 2322, 7 |
| RPC envelope + ChatDto payload | **320, 1** | 1328, 22 | 2354, 13 | 3251, 12 |
| Battle input | **160, 1** | 1560, 36 | 1177, 2 | 2058, 7 |
| TaskDto list, 100 rows | **2688, 1** | 23160, 410 | 16431, 2 | 32809, 11 |
| Leaderboard | **5120, 2** | 22136, 394 | 9766, 2 | 32809, 11 |

### Decode (`B/op`, `allocs/op`)

| Scenario | ByteMsg233 | Protobuf | JSON | MessagePack |
|---|---:|---:|---:|---:|
| Player profile | **0, 0** | 40, 2 | 216, 4 | 48, 1 |
| Chat message | **0, 0** | 56, 2 | 216, 4 | 48, 1 |
| ChatDto all types | **432, 5** | 752, 26 | 600, 28 | 296, 18 |
| RPC envelope + ChatDto payload | **432, 5** | 752, 26 | 1512, 33 | 344, 19 |
| Battle input | **0, 0** | 320, 1 | 232, 5 | 72, 2 |
| Leaderboard, 100 rows | **0, 0** | 8672, 185 | 2377, 189 | 2218, 186 |
| TaskDto list, 100 rows | **0, 0** | 7744, 101 | 1833, 105 | 1674, 102 |

Generated object pools are separate from these raw codec benchmark numbers. They reduce application-level churn after code generation, especially in Unity-style gameplay code and client update loops. Runtime pools are single-threaded and lock-free by policy so hot-path memory reuse stays predictable.

Official runtime libraries follow the `bytemsg233-lib-{language}` naming style and are the language-native runtime dependency for generated projects.

For hot-path encode code, prefer caller-owned buffers. `AppendEncoder` is the zero-GC path for preallocated byte slices:

```bash
go test ./pkg/binary -run ^$ -bench "BenchmarkEncode_TaskList" -benchtime=1000x -benchmem
```

Current `TaskList_ByteMsg233` decode hot path: `0 B/op`, `0 allocs/op` for 100 `TaskDto` entries after decode-state prewarm. Encode uses one caller-facing output allocation in the benchmark; generated hot paths should write into caller-owned buffers.

## Game Traffic Coverage

The benchmark suite must cover real packet families, not only a business DTO list.

| Scenario | Structure |
|---|---|
| Login push | player, 30 heroes, 80 items, 15 mails, 20 quests, settings |
| Battle frame | 10 player inputs, frame id, timestamp, random seed |
| ChatDto all types | bool, signed/unsigned ints, float, double, string, bytes, list, map KV, nested custom messages |
| RPC envelope + ChatDto payload | packet id, sequence, flags, payload bytes, unknown-field skip |
| Leaderboard | 100 rank rows with player, guild, avatar, score |
| Battle input | compact input batch with fixed numeric fields |
| TaskDto list | 100 business DTO rows for non-game repeated data |

Run the game packet checks:

```bash
go test ./pkg/binary -run "TestGame_" -v
go test ./pkg/binary -run "TestBenchmark_ChatDtoAllTypesRoundTrip" -v
go test ./pkg/binary -run ^$ -bench "BenchmarkGame_" -benchmem
go test ./pkg/binary -run ^$ -bench "Benchmark(Encode|Decode)_ChatDtoAllTypes" -benchmem
```

See [GAME_BINARY.md](GAME_BINARY.md) for the message-shape rules.

## Run Locally

```bash
# Payload size comparison
go test ./pkg/binary/... -run "TestBenchmark_SizeComparison" -v

# Game packet checks
go test ./pkg/binary/... -run "TestGame_" -v

# Encoding benchmarks
go test ./pkg/binary/... -bench="BenchmarkEncode_" -benchmem

# Decoding benchmarks
go test ./pkg/binary/... -bench="BenchmarkDecode_" -benchmem

# Game benchmarks
go test ./pkg/binary/... -bench="BenchmarkGame_" -benchmem
```

## Summary

ByteMsg233 is strongest when the project needs all of these at once:

- packet size close to Protobuf and far below JSON/MessagePack;
- generated APIs that feel native in Go, C#, Java, TypeScript, Rust, C++, C, Kotlin, Swift, Dart, Lua, and Python;
- object pooling for client-heavy workloads;
- JSON schema files that are readable in normal editors;
- debug-friendly text output outside the hot path.
