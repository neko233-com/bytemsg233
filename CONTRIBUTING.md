# Contributing

This repository uses a PR-first workflow.

## Access

- Invite `Deqiying` as a GitHub collaborator on the repository.
- Use branch-based development for every change.
- Do not push directly to `main` unless the change is an emergency hotfix.

## Workflow

1. Create a topic branch from `main`.
2. Make the change locally and keep the diff focused.
3. Run the relevant tests or benchmarks before opening a PR.
4. Open a PR and wait for review.
5. Merge only after checks pass and the review is approved.

## Branch Naming

- `codex/<short-topic>`
- `fix/<short-topic>`
- `feat/<short-topic>`

## PR Expectations

- Explain what changed and why.
- Call out any behavior, benchmark, or compatibility impact.
- Link related issues or context when available.
- Keep generated docs and public docs consistent when the change affects release-facing content.

## Project Rules

- Preserve the high-performance hot-path design of ByteMsg233.
- Avoid adding allocations, locks, concurrency primitives, or reflection to runtime hot paths.
- Keep benchmark claims backed by real outputs and `-benchmem` where relevant.

