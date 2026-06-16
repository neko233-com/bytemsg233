# AGENTS.md

## Communication

- caveman full: 精简、专业但完整。

## Performance Policy

- bytemsg233 runtime and generated hot-path code must target ultra-high performance.
- Serialization/deserialization hot paths should aim for zero GC / zero allocation where practical.
- Prefer caller-provided buffers, append-style APIs, object pools, stack values, and reusable decoder state.
- Prefer `AppendEncoder` for zero-GC byte-slice hot paths and `BufferEncoder` for concrete `bytes.Buffer` hot paths.
- Do not add reflection, maps, JSON, fmt-heavy formatting, or heap-heavy helpers to hot paths.
- Debug/tooling APIs may trade allocations for readability, but must be clearly outside hot path.
- Benchmarks must include `-benchmem`; allocation regressions are first-class failures.

## Codegen Policy

- Generated code should prioritize native extension points in each target language.
- Every target language must support normal allocation/constructor usage, such as `new Message()` or the language-native equivalent.
- Every target language must support pool usage with a zero-GC hot-path goal after prewarm or equivalent setup.
- Pool reset should reuse existing lists, dictionaries, arrays, nested messages, and other reusable storage where practical.
- If a target language supports native extension mechanisms, generated code must expose them so users can add custom behavior without editing generated files.
- Extension examples: C# `partial class`, Kotlin/Swift extension methods, TypeScript declaration merging or prototype-safe helpers, Rust extension traits, Go sidecar methods in the same package, Java subclass/companion/helper hooks where practical.
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
  - `[]TaskDto` with at least 100 entries.
  - leaderboard / ranking lists with 100+ entries.
  - battle frame batches.
  - login/full-state game payloads.
- Size comparisons should use realistic game/business payloads and repeated structures, not only one small object.
- Prefer real encoded bytes over theoretical sizes. If theoretical values are used, label them clearly.
