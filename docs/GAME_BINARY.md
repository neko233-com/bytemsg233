# Game Binary Message Design

ByteMsg233 is designed around a very normal game problem: the server keeps sending small packets all day, and sometimes sends one big state package when the player logs in, reconnects, opens a ranking page, or enters a battle.

JSON is comfortable to read, but it repeats field names like `"player_id"` and `"score"` again and again. That is fine for tools. It is wasteful for gameplay traffic. ByteMsg233 keeps the schema readable in JSON, then sends compact binary on the wire.

## The Mental Model

Think of every field as:

```text
field tag + wire type + value
```

The packet does not send the field name. The generated code already knows that tag `1` means `player_id`, tag `2` means `hero_id`, and so on.

For one object this saves a little. For 100 inventory items, 100 ranking rows, or 30 battle frames, it saves a lot because the field names never repeat on the network.

## Game Packet Families

| Packet family | Real example | What ByteMsg233 optimizes |
|---|---|---|
| Tiny hot packets | input, ping, damage tick, skill cast | very small bytes, simple branches, no hidden allocations |
| Repeated lists | leaderboard, inventory, quests, mail | one schema, many rows, no repeated field names |
| Full-state payloads | login push, reconnect state, battle snapshot | nested messages, stable structure, pool-friendly reset |
| Frame batches | battle frame, replay segment, rollback window | append buffers, fixed numeric fields, predictable encode/decode |

If you are new to serialization, this is the important part: a protocol format should match the shape of the traffic. A login payload and a joystick input packet should not be judged with the same tiny toy object.

## Field Encoding Rules

| Field kind | Encoding rule | Beginner note |
|---|---|---|
| `uint32`, `uint64` | varint | small numbers use fewer bytes |
| `int32`, `int64` | zigzag varint | negative values stay compact |
| `bool` | varint, often omitted when false | false does not need to cost a byte in many schemas |
| `enum` | integer value | generated code exposes native enum names |
| `string`, `bytes` | length + bytes | no extra field names inside the value |
| `list<T>` | count + repeated values | ideal for items, rankings, inputs |
| `map<K,V>` | repeated key/value entries | useful, but avoid maps in the hottest frame loops when arrays work |
| nested message | length + message body | keeps complex packets structured |

## Optimized Game Blocks

The current game-first layout is allowed to use dense schema blocks instead of preserving older row-style wire layouts. The goal is simple: keep schemas general, but let repeated game traffic use the shape that CPUs and networks like.

| Block | Best for | Why it helps |
|---|---|---|
| packed varint list | enum, level, status, count, quality columns | one count, then contiguous small integers |
| packed zigzag list | signed coordinates and deltas | negative numbers stay compact |
| delta varint list | rank, id, frame, timestamp, score movement | stores base + small deltas instead of full values |
| bool bitset | flags and repeated booleans | eight booleans per byte |
| string list | names, guilds, labels | one count, then length-prefixed strings |
| dense column list | repeated DTOs such as tasks, inventory, leaderboard rows, battle inputs | avoids repeating field tags for every row |

For example, a 100-row `TaskDto` list is encoded as schema-ordered columns:

```text
count
task_id delta column
type packed column
status packed column
progress packed column
target packed column
reward_id delta column
reward_count packed column
expire_at delta column
title string column
```

This keeps the data general enough for generated code in every language, while avoiding the per-row tag cost that hurts large game lists.

## Deep Nesting Strategy

Complex packets are expected: login state, heroes with skills, inventory with item attrs, battle frames with input lists, and replay segments with frame batches. Deep nesting should be optimized by generated code, not avoided by users.

- Decode nested length-delimited data through slice/span/view readers where the language supports it.
- Decode into existing objects when possible, especially after pool prewarm.
- Reuse list/map/nested-message storage during reset instead of throwing it away.
- Use dense column layout independently at each repeated-message layer; do not force the whole packet into one global layout.
- Keep a reasonable maximum nesting depth in runtime readers to protect servers from malformed input.

## RPC And Socket Usage

ByteMsg233 intentionally stays transport-neutral. TCP length prefixes, WebSocket frames, UDP/datagram envelopes, encryption flags, compression flags, retry sequence numbers, and gateway routing are business or transport concerns. The binary runtime should encode the message body quickly and let the caller choose the socket frame.

For game RPC, use generated message `packetId` values for routing, or define a normal ByteMsg233 envelope message with fields such as `packetId`, `sequence`, `kind`, `flags`, and `payload bytes`. Because it is still a normal message, it keeps the same schema evolution rules and the same benchmarkable hot path.

Do not add a protocol version integer to every gameplay packet by default. The recommended shape is:

1. Client and server open the socket/session.
2. Both sides exchange `ProtocolHello(version, minCompatible)`.
3. If `ByteMsgProtocolVersion` or the compatibility range does not match, reject before entering gameplay traffic.
4. After the handshake passes, gameplay packets stay small and do not repeat the version.

Only put a version field in every packet when a gateway intentionally multiplexes multiple protocol versions on the same connection.

Generated package metadata can include `packetId` and other schema metadata for business templates, registries, or RPC glue, but ByteMsg233 runtime should remain focused on binary encode/decode. Binary-capable generated targets should expose a light interface shape for generic tooling, such as `IByteMsg233Api` with `SerializeByteMsg233` and `DeserializeFromByteMsg233`, while keeping append-style APIs available for the fastest game loops.

## Forward Compatibility

Adding a new field should not break older generated readers. Unknown fields are skipped by wire type:

| Wire type | Meaning | Skip behavior |
|---|---|---|
| `0` | varint | read and discard one varint |
| `1` | fixed64 | skip 8 bytes |
| `2` | length-delimited | read length, skip that many bytes |
| `5` | fixed32 | skip 4 bytes |

Unknown fields are not preserved when an old service re-encodes the message unless business code stores them explicitly. For normal game client/server traffic, the usual flow is to either rely on unknown-field skip for additive rollout or reject incompatible protocol versions during the session handshake. If a team wants content fingerprints, keep them as business-defined handshake fields or envelope fields.

## Hot-Path Rules

Gameplay code should be boring in the best way:

- Encode into a caller-owned buffer when possible.
- Reuse decoder state between packets.
- Reuse lists, arrays, maps, and nested messages when returning objects to a pool.
- Keep debug text, JSON export, logs, and debug views out of the gameplay hot path.
- Avoid reflection and dynamic map-shaped objects in generated encode/decode.

Normal allocation is still supported. You can use `new Hero()` / `Hero{}` / `Hero()` when that is the clearest code. The pool path exists for update loops, battle sync, and other places where GC spikes hurt.

## What We Benchmark

ByteMsg233 benchmarks must cover real game traffic, not only a tiny object:

| Scenario | Why it is included |
|---|---|
| Player profile | small account packet with strings and numbers |
| Chat message | small mixed string/integer packet |
| Battle input | hot numeric input batch for many clients |
| Battle frame | realtime frame sync at 30/60 FPS |
| Login push | full-state payload: player, heroes, items, mail, quests, settings |
| Leaderboard | 100+ repeated rows with player names and guild names |
| TaskDto list | business-style repeated DTO baseline |
| Guild war / large state | multi-list state snapshot with towers, guilds, and rankings |

Run locally:

```bash
go test ./pkg/binary -run "TestBenchmark_SizeComparison|TestGame_" -v
go test ./pkg/binary -run ^$ -bench "Benchmark(Encode_|Game_)" -benchmem
powershell -ExecutionPolicy Bypass -File scripts/test-java.ps1
```

Comparison order is always:

1. ByteMsg233
2. Protobuf
3. JSON
4. Optional codecs, such as MessagePack

## Practical Advice

Use ByteMsg233 when you want readable schema files and compact runtime traffic at the same time.

Use debug text only when inspecting a bad packet.

Use JSON export when talking to tools, logs, or designers.

Use the binary path for the actual client/server protocol.
