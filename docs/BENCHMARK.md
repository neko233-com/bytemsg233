# Performance

> Go benchmark snapshot · Windows amd64 · AMD Ryzen 9 7900X3D

This page explains what the numbers mean, not just who wins a table.

ByteMsg233 is optimized for schema-driven game and client traffic: small field headers, dense column blocks for repeated DTOs, packed varint columns, delta varint columns, bool bitsets, no repeated field names, and generated APIs that can reuse memory. The target is a practical protocol: readable schema, compact packets, native generated code, high-throughput encode/decode on repeated DTO payloads, and a zero-GC hot path where the caller provides buffers and pools are prewarmed.

The headline is straightforward: the Go fast path now beats the Protobuf wire helper baselines in every Protobuf comparison in this suite, while keeping allocation pressure low and repeated client payloads compact.

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
| Battle input, 10 players x 8 fields | **247 B** | 266 B | 1097 B | 931 B |
| TaskDto list, 100 rows x 9 fields | **2261 B** | 4044 B | 14691 B | 13303 B |
| Leaderboard, 100 rows x 6 fields | **2518 B** | 3608 B | 9602 B | 8711 B |

Savings versus other codecs:

| Scenario | vs Protobuf | vs JSON | vs MessagePack |
|---|---:|---:|---:|
| Player profile | 0% | -64.7% | -60.6% |
| Chat message | 0% | -50.9% | -44.7% |
| ChatDto all types | -3.8% | -53.0% | -42.7% |
| Battle input | -7.1% | -77.5% | -73.5% |
| TaskDto list | -44.1% | -84.6% | -83.0% |
| Leaderboard | -30.2% | -73.8% | -71.1% |

## Encode Speed

Tiny packets and repeated structures both use the ByteMsg233 fast path in this snapshot.

These values are duration. Lower `ns/op` is better.

| Scenario | ByteMsg233 | Protobuf | JSON | MessagePack |
|---|---:|---:|---:|---:|
| Player profile | **39.9** | 86.7 | 259.4 | 373.0 |
| Chat message | **32.9** | 63.0 | 189.9 | 226.3 |
| ChatDto all types | **214.4** | 752.3 | 1440 | 1487 |
| Battle input | **169.1** | 952.5 | 2056 | 2500 |
| TaskDto list, 100 rows | **2191** | 14451 | 22726 | 46997 |
| Leaderboard | **4158** | 22715 | 20665 | 37416 |

The same benchmark view as throughput. Higher `ops/s` is better. These are single-threaded operations per second on Go/windows/amd64.

| Scenario | Codec | Encode ops/s | Decode ops/s |
|---|---|---:|---:|
| Player profile | ByteMsg233 | **25050100** | **17132088** |
| Player profile | Protobuf | 11531365 | 9523810 |
| Player profile | JSON | 3855050 | 656168 |
| Player profile | MessagePack | 2680965 | 1665834 |
| Chat message | ByteMsg233 | **30376671** | **33852404** |
| Chat message | Protobuf | 15865461 | 12286521 |
| Chat message | JSON | 5266456 | 1181892 |
| Chat message | MessagePack | 4419797 | 3109453 |
| ChatDto all types | ByteMsg233 | **4664179** | **1695778** |
| ChatDto all types | Protobuf | 1329257 | 829876 |
| ChatDto all types | JSON | 694444 | 130787 |
| ChatDto all types | MessagePack | 672495 | 418936 |
| Battle input | ByteMsg233 | **5913661** | 2608923 |
| Battle input | Protobuf | 1049869 | - |
| Battle input | JSON | 486381 | 5790388 |
| Battle input | MessagePack | 400000 | 11169440 |
| TaskDto list, 100 rows | ByteMsg233 | **456413** | **307125** |
| TaskDto list, 100 rows | Protobuf | 69199 | 94913 |
| TaskDto list, 100 rows | JSON | 44002 | 7137 |
| TaskDto list, 100 rows | MessagePack | 21278 | 18996 |
| Leaderboard | ByteMsg233 | **240500** | - |
| Leaderboard | Protobuf | 44024 | - |
| Leaderboard | JSON | 48391 | - |
| Leaderboard | MessagePack | 26727 | - |

## Language Throughput Matrix

Every official runtime must expose the same optimized block capabilities. Cross-language benchmark harnesses must report one-second throughput against Protobuf, JSON, and MessagePack equivalents before a language-specific performance claim is published.

| Language | Optimized block runtime | Local verification in this release | Cross-codec throughput table |
|---|---|---|---|
| Go | packed, delta, bitset, string list, dense column benchmark | `go test ./...` + benchmark | measured above |
| C# / Unity | packed, delta, bitset, string list | `dotnet test libs/csharp/Tests/ByteMsg233.Tests.csproj` | pending harness |
| TypeScript / JavaScript | packed, delta, bitset, string list | `npm test` | pending harness |
| Rust | packed, delta, bitset, string list | `cargo test` | pending harness |
| Java / Android | packed, delta, bitset, string list | `scripts/test-java.ps1` smoke + `gradle test` with JDK 17 | pending harness |
| C++, C, Kotlin, Swift, Dart, Lua, Python | required by policy | pending runtime parity work | pending harness |

ChatDto relative view:

| Codec | Encode duration | Decode duration | Encode throughput | Decode throughput |
|---|---:|---:|---:|---:|
| ByteMsg233 | **0.28x Protobuf** | **0.49x Protobuf** | **3.51x Protobuf** | **2.04x Protobuf** |
| Protobuf | 3.51x ByteMsg233 | 2.04x ByteMsg233 | 0.28x ByteMsg233 | 0.49x ByteMsg233 |
| JSON | 6.72x ByteMsg233 | 6.35x Protobuf | 0.15x ByteMsg233 | 0.16x Protobuf |
| MessagePack | 6.94x ByteMsg233 | 1.98x Protobuf | 0.14x ByteMsg233 | 0.49x Protobuf |

Interpretation:

- ByteMsg233 simple DTO encode uses append helpers; repeated DTO encode uses dense column blocks with delta/packed primitive columns where the schema is stable.
- ByteMsg233 decode uses `SliceDecoder` and zero-copy string/bytes views for immutable payload buffers.
- In this suite, ByteMsg233 is faster than Protobuf for every measured Protobuf encode/decode comparison.
- JSON and MessagePack pay for dynamic object shape and field-name-heavy data.
- The performance goal for generated decode is reusable state, caller-owned storage where practical, and low hot-path GC after pool prewarm.

## Decode Speed

Decode uses the slice fast path. Lower `ns/op` is better.

| Scenario | ByteMsg233 | Protobuf | JSON | MessagePack |
|---|---:|---:|---:|---:|
| Player profile | **58.4** | 105.0 | 1524 | 600.3 |
| Chat message | **29.5** | 81.4 | 846.1 | 321.6 |
| ChatDto all types | **589.7** | 1205 | 7646 | 2387 |
| Battle input | 383.3 | - | 172.7 | **89.5** |
| TaskDto list, 100 rows | **3256** | 10536 | 140119 | 52642 |

## Allocations

Allocations are where game clients feel pain: a small per-packet allocation can become a frame-time spike when repeated thousands of times.

### Encode (`B/op`, `allocs/op`)

| Scenario | ByteMsg233 | Protobuf | JSON | MessagePack |
|---|---:|---:|---:|---:|
| Player profile | **64, 1** | 104, 3 | 176, 1 | 496, 4 |
| ChatDto all types | **0, 0** | 1328, 22 | 1281, 11 | 2322, 7 |
| Battle input | **256, 1** | 1560, 36 | 1177, 2 | 2058, 7 |
| TaskDto list, 100 rows | **2688, 1** | 23160, 410 | 16431, 2 | 32808, 11 |
| Leaderboard | **5120, 2** | 22136, 394 | 9765, 2 | 32809, 11 |

### Decode (`B/op`, `allocs/op`)

| Scenario | ByteMsg233 | Protobuf | JSON | MessagePack |
|---|---:|---:|---:|---:|
| Player profile | **0, 0** | 40, 2 | 216, 4 | 48, 1 |
| Chat message | **0, 0** | 56, 2 | 216, 4 | 48, 1 |
| ChatDto all types | **432, 5** | 752, 26 | 600, 28 | 296, 18 |
| Battle input | **0, 0** | - | 144, 1 | 48, 1 |
| TaskDto list, 100 rows | **0, 0** | 7744, 101 | 1833, 105 | 1674, 102 |

Generated object pools are separate from these raw codec benchmark numbers. They reduce application-level churn after code generation, especially in Unity-style gameplay code and client update loops. Runtime pools are single-threaded and lock-free by policy so hot-path memory reuse stays predictable.

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
