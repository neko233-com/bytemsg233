# Changelog

All notable changes to ByteMsg233 will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [1.1.1] - 2026-07-07

### Added

- Root-level CHANGELOG.md

## [1.1.0] - 2026-07-06

### Added

- `bytemsg233 update` self-update command with optional `--mirror` fallback
- Schema import/export with Proto interop (`import-proto`, `export-proto`)
- Python 3 and Verse (UEFN) runtime libs as submodules
- PHP, Erlang, Objective-C runtime libs as submodules
- PR-first contributing guide (`CONTRIBUTING.md`)

### Changed

- Renamed submodules to `-lib-` convention; updated all refs to include full runtime source
- CI: keep GitHub Pages on legacy static publish workflow
- README: bilingual support with updated install instructions

## [1.0.7] - 2026-06-24

### Fixed

- Go buffer pools for concurrent use

## [1.0.6] - 2026-06-17

### Added

- ByteMsg233 logo assets

## [1.0.5] - 2026-06-17

### Changed

- Optimized game binary protocol flow

## [1.0.4] - 2026-06-17

### Added

- Dense column game blocks for optimized repeated DTO encoding
- Docker benchmark suite

## [1.0.3] - 2026-06-16

### Changed

- Codec and documentation optimizations
- Added GitHub Docs HTML policy

## [1.0.2] - 2026-06-16

### Changed

- Enforced single-threaded runtime hot paths (no concurrency primitives in encode/decode)
- Benchmarks: added relative multiplier display and all-type ChatDto comparisons
- Docs: expanded language roadmap and game benchmarks, bilingual demo view

## [1.0.1] - 2026-06-16

### Added

- Go codegen: wire codecs for `Marshal`/`Unmarshal` with `IByteMsg233Api` shape
- Go codegen: zero-alloc debug text generation
- Improved generated debug and pooling APIs
- Enum/list/map comments as first-class schema citizens

### Changed

- Bounded bytemsg object pools for memory safety

## [1.0.0] - 2026-06-16

### Added

- JSON-first binary schema DSL (`.bmsg.json`) as primary protocol description
- Go codegen with hot-path encode/decode, pool support, and `AppendEncoder`
- CLI: `init`, `compile`, `export` commands
- Protocol doc export to Markdown and HTML
- GoReleaser-based multi-platform release pipeline

## [0.2.0] - 2026-06-14

### Added

- Initial public release
- Binary schema compilation and code generation
- Basic runtime library structure
