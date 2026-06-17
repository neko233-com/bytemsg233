# AGENTS.md

## Communication

- caveman full: 精简、专业但完整。

## Docs Policy

- Documentation updates that affect GitHub Docs must update every checked-in HTML page under `docs/`, including `docs/index.html` and `docs/demo/index.html`.
- Markdown docs and HTML docs must stay consistent on product naming, benchmark numbers, install commands, and release-facing claims.
- HTML docs are for human readers only. Do not include agent instructions, implementation plans, prompts, or internal workflow guidance in `docs/**/*.html`.
- Put agent-facing guidance in `AGENTS.md` or developer-only Markdown docs, not public HTML.

## Performance Policy

- bytemsg233 runtime and generated hot-path code must target ultra-high performance.
- Serialization/deserialization hot paths should aim for zero GC / zero allocation where practical.
- Target language runtime libraries and generated encode/decode paths must be single-threaded and lock-free.
- Do not use concurrency primitives in target runtime or generated hot paths: no goroutines, channels, `sync.Pool`, mutexes, locks, `Concurrent*` collections, `synchronized`, thread pools, atomics, or background workers.
- Prefer caller-provided buffers, append-style APIs, object pools, stack values, and reusable decoder state.
- Prefer `AppendEncoder` for zero-GC byte-slice hot paths and `BufferEncoder` for concrete `bytes.Buffer` hot paths.
- Runtime and generated code must not force a built-in socket frame. Packet id, seq, kind, flags, encryption, compression, and length framing belong to the transport/business envelope chosen by the caller.
- Protocol compatibility should be checked once per connection/session with generated `ByteMsgProtocolVersion` and `ProtocolHello(version, minCompatible)`. Do not add a repeated per-message version integer to the hot path unless a gateway intentionally multiplexes protocol versions on one connection. Content fingerprints are business-owned and should not be generated or enforced by the runtime by default.
- Binary-capable generated targets should expose a lightweight `IByteMsg233Api` shape where practical, with `SerializeByteMsg233` and `DeserializeFromByteMsg233` wrappers. These wrappers must not remove lower-level zero-GC append/decode APIs used by game hot paths.
- Do not add reflection, maps, JSON, fmt-heavy formatting, or heap-heavy helpers to hot paths.
- Debug/tooling APIs may trade allocations for readability, but must be clearly outside hot path.
- Benchmarks must include `-benchmem`; allocation regressions are first-class failures.

## Codegen Policy

- Generated code should prioritize native extension points in each target language.
- Common target languages are part of the product roadmap, not third-party afterthoughts. Official targets should cover Go, C# / Unity, TypeScript / JavaScript, Java / Android, Rust, C++, C, Kotlin, Swift, Dart / Flutter, Lua, and Python.
- Every official target language must support the optimized game binary block layouts: packed varint lists, packed zigzag lists, delta varint lists, bool bitsets, string lists, and schema-driven dense column lists.
- New generated message-list hot paths should prefer schema-driven dense column layout for fixed-schema repeated DTOs, leaderboards, inventories, task lists, battle inputs, and frame batches.
- Deeply nested generated decode must avoid avoidable copies: prefer slice/span/view subreaders, decode-into APIs, and reusable list/map/nested-message storage.
- Generated readers and runtime readers must skip unknown fields for supported wire types `0` varint, `1` fixed64, `2` length-delimited, and `5` fixed32 so adding fields remains forward-compatible.
- Optimized binary layouts are allowed to supersede older row layouts. Do not preserve old wire formats when doing so would block game hot-path performance, unless a release note explicitly requires compatibility.
- Every target language must support normal allocation/constructor usage, such as `new Message()` or the language-native equivalent.
- Every target language must support pool usage with a zero-GC hot-path goal after prewarm or equivalent setup.
- Pool reset should reuse existing lists, dictionaries, arrays, nested messages, and other reusable storage where practical.
- Bytes fields should expose a reusable-buffer path where the target language can support it without making the normal API awkward.
- Generated packages must expose a protocol-version helper, such as `GetByteMsg233ProtocolVersion`, so business code can verify schema/runtime version without putting version data on every message.
- If a target language supports native extension mechanisms, generated code must expose them so users can add custom behavior without editing generated files.
- Extension examples: C# `partial class`, Kotlin/Swift extension methods, TypeScript declaration merging or prototype-safe helpers, Rust extension traits, C++ free functions / wrapper types, C opaque context hooks, Go sidecar methods in the same package, Java subclass/companion/helper hooks where practical.
- Single-file language exports must default to `ByteMsg233_Export` plus the target extension, such as `ByteMsg233_Export.go` or `ByteMsg233_Export.cs`.
- Multi-file targets that require class-name-matching file names, such as Java, may keep per-type file names.
- C# / Unity generated messages must be `partial class` so users can add custom methods in separate files.
- C# / Unity generated messages must support both `new Message()` and pool flow: `Message.Prewarm(count)`, `Message.Rent()`, `Message.Return(value)`, `value.Release()`.
- C# / Unity pool hot paths should avoid GC after prewarm; reset should reuse existing lists, dictionaries, and nested messages where practical.

## Benchmark Policy

- Benchmark comparison order must be:
  1. bytemsg233
  2. Protobuf
  3. JSON
- MessagePack or other codecs may appear after those three only.
- Avoid tiny-only comparisons. Include large repeated DTO scenarios, especially:
  - `ChatDto` all-type encode/decode coverage with bool, signed/unsigned integers, float, double, string, bytes, list, map key/value parameters, and nested custom message types.
  - `[]TaskDto` with at least 100 entries.
  - leaderboard / ranking lists with 100+ entries.
  - battle frame batches.
  - login/full-state game payloads.
- Size comparisons should use realistic game/business payloads and repeated structures, not only one small object.
- Prefer real encoded bytes over theoretical sizes. If theoretical values are used, label them clearly.
- Performance docs should show both duration (`ns/op`) and throughput (`ops/s` or `次/s`) when presenting encode/decode speed.
