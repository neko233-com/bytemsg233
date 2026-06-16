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
