package binary

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/encoding/protowire"
)

// ==================== 共享数据结构 ====================

type BenchPlayer struct {
	Uid       uint64 `json:"uid" msgpack:"uid"`
	Name      string `json:"name" msgpack:"name"`
	Level     uint32 `json:"level" msgpack:"level"`
	VipLevel  uint32 `json:"vip_level" msgpack:"vip_level"`
	Diamond   uint32 `json:"diamond" msgpack:"diamond"`
	Gold      uint64 `json:"gold" msgpack:"gold"`
	Energy    uint32 `json:"energy" msgpack:"energy"`
	Avatar    uint32 `json:"avatar" msgpack:"avatar"`
	GuildId   uint32 `json:"guild_id" msgpack:"guild_id"`
	GuildName string `json:"guild_name" msgpack:"guild_name"`
}

type BenchHero struct {
	HeroId  uint32            `json:"hero_id" msgpack:"hero_id"`
	Level   uint32            `json:"level" msgpack:"level"`
	Star    uint32            `json:"star" msgpack:"star"`
	Grade   uint32            `json:"grade" msgpack:"grade"`
	Exp     uint64            `json:"exp" msgpack:"exp"`
	Skills  []BenchSkill      `json:"skills" msgpack:"skills"`
	Runes   map[uint32]uint32 `json:"runes" msgpack:"runes"`
	SkinId  uint32            `json:"skin_id" msgpack:"skin_id"`
	AwakeLv uint32            `json:"awake_lv" msgpack:"awake_lv"`
	FavorLv uint32            `json:"favor_lv" msgpack:"favor_lv"`
}

type BenchSkill struct {
	SkillId uint32 `json:"skill_id" msgpack:"skill_id"`
	Level   uint32 `json:"level" msgpack:"level"`
}

type BenchChatMsg struct {
	Channel  uint32 `json:"channel" msgpack:"channel"`
	SenderId uint32 `json:"sender_id" msgpack:"sender_id"`
	Sender   string `json:"sender" msgpack:"sender"`
	Content  string `json:"content" msgpack:"content"`
	Time     uint64 `json:"time" msgpack:"time"`
}

type BenchBattleInput struct {
	PlayerId uint32 `json:"player_id" msgpack:"player_id"`
	HeroId   uint32 `json:"hero_id" msgpack:"hero_id"`
	Action   uint32 `json:"action" msgpack:"action"`
	SkillId  uint32 `json:"skill_id" msgpack:"skill_id"`
	TargetId uint32 `json:"target_id" msgpack:"target_id"`
	X        int32  `json:"x" msgpack:"x"`
	Y        int32  `json:"y" msgpack:"y"`
	Dir      uint32 `json:"dir" msgpack:"dir"`
}

type BenchRankEntry struct {
	Rank     uint32 `json:"rank" msgpack:"rank"`
	PlayerId uint64 `json:"player_id" msgpack:"player_id"`
	Name     string `json:"name" msgpack:"name"`
	Level    uint32 `json:"level" msgpack:"level"`
	Score    uint64 `json:"score" msgpack:"score"`
	Guild    string `json:"guild" msgpack:"guild"`
}

type BenchTaskDto struct {
	TaskId      uint32 `json:"task_id" msgpack:"task_id"`
	Type        uint32 `json:"type" msgpack:"type"`
	Status      uint32 `json:"status" msgpack:"status"`
	Progress    uint32 `json:"progress" msgpack:"progress"`
	Target      uint32 `json:"target" msgpack:"target"`
	RewardId    uint32 `json:"reward_id" msgpack:"reward_id"`
	RewardCount uint32 `json:"reward_count" msgpack:"reward_count"`
	ExpireAt    uint64 `json:"expire_at" msgpack:"expire_at"`
	Title       string `json:"title" msgpack:"title"`
}

// ==================== 测试数据 ====================

func benchMakePlayer() BenchPlayer {
	return BenchPlayer{
		Uid: 100000001, Name: "绝影·暗夜猎手", Level: 65, VipLevel: 8,
		Diamond: 12580, Gold: 9876543, Energy: 85, Avatar: 1001,
		GuildId: 5001, GuildName: "苍穹之巅",
	}
}

func benchMakeHero() BenchHero {
	return BenchHero{
		HeroId: 10001, Level: 60, Star: 5, Grade: 3, Exp: 12345678,
		Skills: []BenchSkill{
			{101, 10}, {102, 8}, {103, 6}, {104, 4},
		},
		Runes:  map[uint32]uint32{1: 30001, 2: 30002, 3: 30003},
		SkinId: 40001, AwakeLv: 2, FavorLv: 8,
	}
}

func benchMakeChat() BenchChatMsg {
	return BenchChatMsg{
		Channel: 1, SenderId: 10001, Sender: "亚瑟",
		Content: "集合！准备打团！冲冲冲！", Time: 1718304000,
	}
}

func benchMakeBattleInputs() []BenchBattleInput {
	inputs := make([]BenchBattleInput, 10)
	for i := range inputs {
		inputs[i] = BenchBattleInput{
			PlayerId: uint32(10001 + i), HeroId: uint32(20001 + i),
			Action: uint32(i % 5), SkillId: uint32(30001 + i%3),
			TargetId: uint32(10001 + (i+5)%10),
			X:        int32(1000 + i*50), Y: int32(2000 - i*30), Dir: uint32(i * 36),
		}
	}
	return inputs
}

func benchMakeLeaderboard() []BenchRankEntry {
	entries := make([]BenchRankEntry, 100)
	guilds := []string{"苍穹之巅", "星辰大海", "龙之领域", "暗影军团", "光明圣殿", ""}
	for i := range entries {
		entries[i] = BenchRankEntry{
			Rank: uint32(i + 1), PlayerId: uint64(100000 + i),
			Name:  "玩家" + string(rune('A'+i%26)) + string(rune('0'+i/26)),
			Level: uint32(50 + i%15), Score: uint64(1000000 - i*8000),
			Guild: guilds[i%len(guilds)],
		}
	}
	return entries
}

func benchMakeTasks(count int) []BenchTaskDto {
	tasks := make([]BenchTaskDto, count)
	for i := range tasks {
		tasks[i] = BenchTaskDto{
			TaskId:      uint32(70000 + i),
			Type:        uint32(1 + i%8),
			Status:      uint32(i % 4),
			Progress:    uint32((i * 7) % 100),
			Target:      uint32(100 + i%50),
			RewardId:    uint32(90000 + i%12),
			RewardCount: uint32(10 + i%90),
			ExpireAt:    uint64(1718304000 + i*3600),
			Title:       "每日任务",
		}
	}
	return tasks
}

// ==================== ByteMsg 编码/解码 ====================

func encodePlayerBmsg(p BenchPlayer) []byte {
	buf := GetBuffer()
	defer PutBuffer(buf)
	enc := NewBufferEncoderValue(buf)
	enc.WriteFieldHeader(1, 0)
	enc.WriteVarint(p.Uid)
	enc.WriteFieldHeader(2, 2)
	enc.WriteString(p.Name)
	enc.WriteFieldHeader(3, 0)
	enc.WriteVarint(uint64(p.Level))
	enc.WriteFieldHeader(4, 0)
	enc.WriteVarint(uint64(p.VipLevel))
	enc.WriteFieldHeader(5, 0)
	enc.WriteVarint(uint64(p.Diamond))
	enc.WriteFieldHeader(6, 0)
	enc.WriteVarint(p.Gold)
	enc.WriteFieldHeader(7, 0)
	enc.WriteVarint(uint64(p.Energy))
	enc.WriteFieldHeader(8, 0)
	enc.WriteVarint(uint64(p.Avatar))
	enc.WriteFieldHeader(9, 0)
	enc.WriteVarint(uint64(p.GuildId))
	enc.WriteFieldHeader(10, 2)
	enc.WriteString(p.GuildName)
	return append([]byte(nil), buf.Bytes()...)
}

func decodePlayerBmsg(data []byte) BenchPlayer {
	dec := NewDecoder(bytes.NewReader(data))
	var p BenchPlayer
	for {
		tag, wt, err := dec.ReadFieldHeader()
		if err != nil {
			break
		}
		switch {
		case tag == 1 && wt == 0:
			v, _ := dec.ReadVarint()
			p.Uid = v
		case tag == 2 && wt == 2:
			v, _ := dec.ReadString()
			p.Name = v
		case tag == 3 && wt == 0:
			v, _ := dec.ReadVarint()
			p.Level = uint32(v)
		case tag == 4 && wt == 0:
			v, _ := dec.ReadVarint()
			p.VipLevel = uint32(v)
		case tag == 5 && wt == 0:
			v, _ := dec.ReadVarint()
			p.Diamond = uint32(v)
		case tag == 6 && wt == 0:
			v, _ := dec.ReadVarint()
			p.Gold = v
		case tag == 7 && wt == 0:
			v, _ := dec.ReadVarint()
			p.Energy = uint32(v)
		case tag == 8 && wt == 0:
			v, _ := dec.ReadVarint()
			p.Avatar = uint32(v)
		case tag == 9 && wt == 0:
			v, _ := dec.ReadVarint()
			p.GuildId = uint32(v)
		case tag == 10 && wt == 2:
			v, _ := dec.ReadString()
			p.GuildName = v
		default:
			return p
		}
	}
	return p
}

func encodeChatBmsg2(c BenchChatMsg) []byte {
	buf := GetBuffer()
	defer PutBuffer(buf)
	enc := NewEncoder(buf)
	enc.WriteFieldHeader(1, 0)
	enc.WriteVarint(uint64(c.Channel))
	enc.WriteFieldHeader(2, 0)
	enc.WriteVarint(uint64(c.SenderId))
	enc.WriteFieldHeader(3, 2)
	enc.WriteString(c.Sender)
	enc.WriteFieldHeader(4, 2)
	enc.WriteString(c.Content)
	enc.WriteFieldHeader(5, 0)
	enc.WriteVarint(c.Time)
	return append([]byte(nil), buf.Bytes()...)
}

func decodeChatBmsg(data []byte) BenchChatMsg {
	dec := NewDecoder(bytes.NewReader(data))
	var c BenchChatMsg
	for {
		tag, wt, err := dec.ReadFieldHeader()
		if err != nil {
			break
		}
		switch {
		case tag == 1 && wt == 0:
			v, _ := dec.ReadVarint()
			c.Channel = uint32(v)
		case tag == 2 && wt == 0:
			v, _ := dec.ReadVarint()
			c.SenderId = uint32(v)
		case tag == 3 && wt == 2:
			v, _ := dec.ReadString()
			c.Sender = v
		case tag == 4 && wt == 2:
			v, _ := dec.ReadString()
			c.Content = v
		case tag == 5 && wt == 0:
			v, _ := dec.ReadVarint()
			c.Time = v
		default:
			return c
		}
	}
	return c
}

func encodeInputsBmsg(inputs []BenchBattleInput) []byte {
	buf := GetBuffer()
	defer PutBuffer(buf)
	enc := NewEncoder(buf)
	enc.WriteVarint(uint64(len(inputs)))
	for _, in := range inputs {
		enc.WriteFieldHeader(1, 0)
		enc.WriteVarint(uint64(in.PlayerId))
		enc.WriteFieldHeader(2, 0)
		enc.WriteVarint(uint64(in.HeroId))
		enc.WriteFieldHeader(3, 0)
		enc.WriteVarint(uint64(in.Action))
		enc.WriteFieldHeader(4, 0)
		enc.WriteVarint(uint64(in.SkillId))
		enc.WriteFieldHeader(5, 0)
		enc.WriteVarint(uint64(in.TargetId))
		enc.WriteFieldHeader(6, 0)
		enc.WriteZigzag(int64(in.X))
		enc.WriteFieldHeader(7, 0)
		enc.WriteZigzag(int64(in.Y))
		enc.WriteFieldHeader(8, 0)
		enc.WriteVarint(uint64(in.Dir))
	}
	return append([]byte(nil), buf.Bytes()...)
}

func encodeLeaderboardBmsg2(entries []BenchRankEntry) []byte {
	buf := GetBuffer()
	defer PutBuffer(buf)
	enc := NewEncoder(buf)
	enc.WriteVarint(uint64(len(entries)))
	for _, e := range entries {
		enc.WriteFieldHeader(1, 0)
		enc.WriteVarint(uint64(e.Rank))
		enc.WriteFieldHeader(2, 0)
		enc.WriteVarint(e.PlayerId)
		enc.WriteFieldHeader(3, 2)
		enc.WriteString(e.Name)
		enc.WriteFieldHeader(4, 0)
		enc.WriteVarint(uint64(e.Level))
		enc.WriteFieldHeader(5, 0)
		enc.WriteVarint(e.Score)
		enc.WriteFieldHeader(6, 2)
		enc.WriteString(e.Guild)
	}
	return append([]byte(nil), buf.Bytes()...)
}

func encodeTasksBmsg(tasks []BenchTaskDto) []byte {
	buf := GetBuffer()
	defer PutBuffer(buf)
	encodeTasksBmsgTo(buf, tasks)
	return append([]byte(nil), buf.Bytes()...)
}

func encodeTasksBmsgTo(buf *bytes.Buffer, tasks []BenchTaskDto) {
	buf.Reset()
	enc := NewEncoder(buf)
	enc.WriteVarint(uint64(len(tasks)))
	for _, task := range tasks {
		enc.WriteFieldHeader(1, 0)
		enc.WriteVarint(uint64(task.TaskId))
		enc.WriteFieldHeader(2, 0)
		enc.WriteVarint(uint64(task.Type))
		enc.WriteFieldHeader(3, 0)
		enc.WriteVarint(uint64(task.Status))
		enc.WriteFieldHeader(4, 0)
		enc.WriteVarint(uint64(task.Progress))
		enc.WriteFieldHeader(5, 0)
		enc.WriteVarint(uint64(task.Target))
		enc.WriteFieldHeader(6, 0)
		enc.WriteVarint(uint64(task.RewardId))
		enc.WriteFieldHeader(7, 0)
		enc.WriteVarint(uint64(task.RewardCount))
		enc.WriteFieldHeader(8, 0)
		enc.WriteVarint(task.ExpireAt)
		enc.WriteFieldHeader(9, 2)
		enc.WriteString(task.Title)
	}
}

// ==================== Protobuf 编码/解码 (手动 wire format) ====================

func encodePlayerProto(p BenchPlayer) []byte {
	var buf []byte
	buf = protowire.AppendTag(buf, 1, protowire.VarintType)
	buf = protowire.AppendVarint(buf, p.Uid)
	buf = protowire.AppendTag(buf, 2, protowire.BytesType)
	buf = protowire.AppendString(buf, p.Name)
	buf = protowire.AppendTag(buf, 3, protowire.VarintType)
	buf = protowire.AppendVarint(buf, uint64(p.Level))
	buf = protowire.AppendTag(buf, 4, protowire.VarintType)
	buf = protowire.AppendVarint(buf, uint64(p.VipLevel))
	buf = protowire.AppendTag(buf, 5, protowire.VarintType)
	buf = protowire.AppendVarint(buf, uint64(p.Diamond))
	buf = protowire.AppendTag(buf, 6, protowire.VarintType)
	buf = protowire.AppendVarint(buf, p.Gold)
	buf = protowire.AppendTag(buf, 7, protowire.VarintType)
	buf = protowire.AppendVarint(buf, uint64(p.Energy))
	buf = protowire.AppendTag(buf, 8, protowire.VarintType)
	buf = protowire.AppendVarint(buf, uint64(p.Avatar))
	buf = protowire.AppendTag(buf, 9, protowire.VarintType)
	buf = protowire.AppendVarint(buf, uint64(p.GuildId))
	buf = protowire.AppendTag(buf, 10, protowire.BytesType)
	buf = protowire.AppendString(buf, p.GuildName)
	return buf
}

func decodePlayerProto(data []byte) BenchPlayer {
	var p BenchPlayer
	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			break
		}
		data = data[n:]
		switch typ {
		case protowire.VarintType:
			v, n := protowire.ConsumeVarint(data)
			if n < 0 {
				break
			}
			data = data[n:]
			switch num {
			case 1:
				p.Uid = v
			case 3:
				p.Level = uint32(v)
			case 4:
				p.VipLevel = uint32(v)
			case 5:
				p.Diamond = uint32(v)
			case 6:
				p.Gold = v
			case 7:
				p.Energy = uint32(v)
			case 8:
				p.Avatar = uint32(v)
			case 9:
				p.GuildId = uint32(v)
			}
		case protowire.BytesType:
			v, n := protowire.ConsumeBytes(data)
			if n < 0 {
				break
			}
			data = data[n:]
			switch num {
			case 2:
				p.Name = string(v)
			case 10:
				p.GuildName = string(v)
			}
		default:
			n = protowire.ConsumeFieldValue(num, typ, data)
			if n < 0 {
				break
			}
			data = data[n:]
		}
	}
	return p
}

func encodeChatProto(c BenchChatMsg) []byte {
	var buf []byte
	buf = protowire.AppendTag(buf, 1, protowire.VarintType)
	buf = protowire.AppendVarint(buf, uint64(c.Channel))
	buf = protowire.AppendTag(buf, 2, protowire.VarintType)
	buf = protowire.AppendVarint(buf, uint64(c.SenderId))
	buf = protowire.AppendTag(buf, 3, protowire.BytesType)
	buf = protowire.AppendString(buf, c.Sender)
	buf = protowire.AppendTag(buf, 4, protowire.BytesType)
	buf = protowire.AppendString(buf, c.Content)
	buf = protowire.AppendTag(buf, 5, protowire.VarintType)
	buf = protowire.AppendVarint(buf, c.Time)
	return buf
}

func decodeChatProto(data []byte) BenchChatMsg {
	var c BenchChatMsg
	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			break
		}
		data = data[n:]
		switch typ {
		case protowire.VarintType:
			v, n := protowire.ConsumeVarint(data)
			if n < 0 {
				break
			}
			data = data[n:]
			switch num {
			case 1:
				c.Channel = uint32(v)
			case 2:
				c.SenderId = uint32(v)
			case 5:
				c.Time = v
			}
		case protowire.BytesType:
			v, n := protowire.ConsumeBytes(data)
			if n < 0 {
				break
			}
			data = data[n:]
			switch num {
			case 3:
				c.Sender = string(v)
			case 4:
				c.Content = string(v)
			}
		default:
			n = protowire.ConsumeFieldValue(num, typ, data)
			if n < 0 {
				break
			}
			data = data[n:]
		}
	}
	return c
}

func encodeInputsProto(inputs []BenchBattleInput) []byte {
	var buf []byte
	for _, in := range inputs {
		msg := protowire.AppendTag(nil, 1, protowire.VarintType)
		msg = protowire.AppendVarint(msg, uint64(in.PlayerId))
		msg = protowire.AppendTag(msg, 2, protowire.VarintType)
		msg = protowire.AppendVarint(msg, uint64(in.HeroId))
		msg = protowire.AppendTag(msg, 3, protowire.VarintType)
		msg = protowire.AppendVarint(msg, uint64(in.Action))
		msg = protowire.AppendTag(msg, 4, protowire.VarintType)
		msg = protowire.AppendVarint(msg, uint64(in.SkillId))
		msg = protowire.AppendTag(msg, 5, protowire.VarintType)
		msg = protowire.AppendVarint(msg, uint64(in.TargetId))
		msg = protowire.AppendTag(msg, 6, protowire.VarintType)
		msg = protowire.AppendVarint(msg, uint64(protowire.EncodeZigZag(int64(in.X))))
		msg = protowire.AppendTag(msg, 7, protowire.VarintType)
		msg = protowire.AppendVarint(msg, uint64(protowire.EncodeZigZag(int64(in.Y))))
		msg = protowire.AppendTag(msg, 8, protowire.VarintType)
		msg = protowire.AppendVarint(msg, uint64(in.Dir))
		buf = protowire.AppendTag(buf, 1, protowire.BytesType)
		buf = protowire.AppendBytes(buf, msg)
	}
	return buf
}

func encodeLeaderboardProto(entries []BenchRankEntry) []byte {
	var buf []byte
	for _, e := range entries {
		msg := protowire.AppendTag(nil, 1, protowire.VarintType)
		msg = protowire.AppendVarint(msg, uint64(e.Rank))
		msg = protowire.AppendTag(msg, 2, protowire.VarintType)
		msg = protowire.AppendVarint(msg, e.PlayerId)
		msg = protowire.AppendTag(msg, 3, protowire.BytesType)
		msg = protowire.AppendString(msg, e.Name)
		msg = protowire.AppendTag(msg, 4, protowire.VarintType)
		msg = protowire.AppendVarint(msg, uint64(e.Level))
		msg = protowire.AppendTag(msg, 5, protowire.VarintType)
		msg = protowire.AppendVarint(msg, e.Score)
		msg = protowire.AppendTag(msg, 6, protowire.BytesType)
		msg = protowire.AppendString(msg, e.Guild)
		buf = protowire.AppendTag(buf, 1, protowire.BytesType)
		buf = protowire.AppendBytes(buf, msg)
	}
	return buf
}

func encodeTasksProto(tasks []BenchTaskDto) []byte {
	var buf []byte
	for _, task := range tasks {
		msg := protowire.AppendTag(nil, 1, protowire.VarintType)
		msg = protowire.AppendVarint(msg, uint64(task.TaskId))
		msg = protowire.AppendTag(msg, 2, protowire.VarintType)
		msg = protowire.AppendVarint(msg, uint64(task.Type))
		msg = protowire.AppendTag(msg, 3, protowire.VarintType)
		msg = protowire.AppendVarint(msg, uint64(task.Status))
		msg = protowire.AppendTag(msg, 4, protowire.VarintType)
		msg = protowire.AppendVarint(msg, uint64(task.Progress))
		msg = protowire.AppendTag(msg, 5, protowire.VarintType)
		msg = protowire.AppendVarint(msg, uint64(task.Target))
		msg = protowire.AppendTag(msg, 6, protowire.VarintType)
		msg = protowire.AppendVarint(msg, uint64(task.RewardId))
		msg = protowire.AppendTag(msg, 7, protowire.VarintType)
		msg = protowire.AppendVarint(msg, uint64(task.RewardCount))
		msg = protowire.AppendTag(msg, 8, protowire.VarintType)
		msg = protowire.AppendVarint(msg, task.ExpireAt)
		msg = protowire.AppendTag(msg, 9, protowire.BytesType)
		msg = protowire.AppendString(msg, task.Title)
		buf = protowire.AppendTag(buf, 1, protowire.BytesType)
		buf = protowire.AppendBytes(buf, msg)
	}
	return buf
}

// ==================== 通用编解码函数 ====================

func encodeJSON(v any) []byte    { d, _ := json.Marshal(v); return d }
func encodeMsgpack(v any) []byte { d, _ := msgpack.Marshal(v); return d }

// ==================== 测试: 体积对比 ====================

func TestBenchmark_SizeComparison(t *testing.T) {
	player := benchMakePlayer()
	chat := benchMakeChat()
	inputs := benchMakeBattleInputs()
	tasks := benchMakeTasks(100)
	lb := benchMakeLeaderboard()

	type row struct {
		Name  string
		Bmsg  int
		Proto int
		JSON  int
		Mp    int
	}

	rows := []row{
		{"玩家信息 (10 fields)", len(encodePlayerBmsg(player)), len(encodePlayerProto(player)), len(encodeJSON(player)), len(encodeMsgpack(player))},
		{"聊天消息 (5 fields)", len(encodeChatBmsg2(chat)), len(encodeChatProto(chat)), len(encodeJSON(chat)), len(encodeMsgpack(chat))},
		{"战斗输入 (10人×8 fields)", len(encodeInputsBmsg(inputs)), len(encodeInputsProto(inputs)), len(encodeJSON(inputs)), len(encodeMsgpack(inputs))},
		{"任务列表 (100 TaskDto×9 fields)", len(encodeTasksBmsg(tasks)), len(encodeTasksProto(tasks)), len(encodeJSON(tasks)), len(encodeMsgpack(tasks))},
		{"排行榜 (100人×6 fields)", len(encodeLeaderboardBmsg2(lb)), len(encodeLeaderboardProto(lb)), len(encodeJSON(lb)), len(encodeMsgpack(lb))},
	}

	t.Logf("")
	t.Logf("╔══════════════════════════════════════════════════════════════════════╗")
	t.Logf("║                        体积对比 (bytes)                             ║")
	t.Logf("╠══════════════════════════════════════════════════════════════════════╣")
	t.Logf("║  %-30s │ %6s │ %6s │ %6s │ %6s ║", "场景", "ByteMsg", "Proto", "JSON", "MsgPk")
	t.Logf("╠══════════════════════════════════════════════════════════════════════╣")
	for _, r := range rows {
		t.Logf("║  %-30s │ %6d │ %6d │ %6d │ %6d ║", r.Name, r.Bmsg, r.Proto, r.JSON, r.Mp)
	}
	t.Logf("╚══════════════════════════════════════════════════════════════════════╝")
	t.Logf("")
	t.Logf("  ByteMsg vs Protobuf:")
	for _, r := range rows {
		ratio := float64(r.Bmsg) / float64(r.Proto) * 100
		t.Logf("    %-24s  %.1f%%", r.Name, ratio)
	}
	t.Logf("")
	t.Logf("  ByteMsg vs JSON:")
	for _, r := range rows {
		saved := (1 - float64(r.Bmsg)/float64(r.JSON)) * 100
		t.Logf("    %-24s  -%.1f%%", r.Name, saved)
	}
}

// ==================== 编码 Benchmark ====================

func BenchmarkEncode_Player_ByteMsg(b *testing.B) {
	p := benchMakePlayer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodePlayerBmsg(p)
	}
}
func BenchmarkEncode_Player_Proto(b *testing.B) {
	p := benchMakePlayer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodePlayerProto(p)
	}
}
func BenchmarkEncode_Player_JSON(b *testing.B) {
	p := benchMakePlayer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeJSON(p)
	}
}
func BenchmarkEncode_Player_Msgpack(b *testing.B) {
	p := benchMakePlayer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeMsgpack(p)
	}
}

func BenchmarkEncode_Chat_ByteMsg(b *testing.B) {
	c := benchMakeChat()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeChatBmsg2(c)
	}
}
func BenchmarkEncode_Chat_Proto(b *testing.B) {
	c := benchMakeChat()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeChatProto(c)
	}
}
func BenchmarkEncode_Chat_JSON(b *testing.B) {
	c := benchMakeChat()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeJSON(c)
	}
}
func BenchmarkEncode_Chat_Msgpack(b *testing.B) {
	c := benchMakeChat()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeMsgpack(c)
	}
}

func BenchmarkEncode_Battle_ByteMsg(b *testing.B) {
	in := benchMakeBattleInputs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeInputsBmsg(in)
	}
}
func BenchmarkEncode_Battle_Proto(b *testing.B) {
	in := benchMakeBattleInputs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeInputsProto(in)
	}
}
func BenchmarkEncode_Battle_JSON(b *testing.B) {
	in := benchMakeBattleInputs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeJSON(in)
	}
}
func BenchmarkEncode_Battle_Msgpack(b *testing.B) {
	in := benchMakeBattleInputs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeMsgpack(in)
	}
}

func BenchmarkEncode_TaskList_ByteMsg(b *testing.B) {
	tasks := benchMakeTasks(100)
	dst := make([]byte, 0, len(encodeTasksBmsg(tasks)))
	enc := NewAppendEncoderValue(dst)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc.Reset(dst[:0])
		enc.WriteVarint(uint64(len(tasks)))
		for _, task := range tasks {
			enc.WriteFieldHeader(1, 0)
			enc.WriteVarint(uint64(task.TaskId))
			enc.WriteFieldHeader(2, 0)
			enc.WriteVarint(uint64(task.Type))
			enc.WriteFieldHeader(3, 0)
			enc.WriteVarint(uint64(task.Status))
			enc.WriteFieldHeader(4, 0)
			enc.WriteVarint(uint64(task.Progress))
			enc.WriteFieldHeader(5, 0)
			enc.WriteVarint(uint64(task.Target))
			enc.WriteFieldHeader(6, 0)
			enc.WriteVarint(uint64(task.RewardId))
			enc.WriteFieldHeader(7, 0)
			enc.WriteVarint(uint64(task.RewardCount))
			enc.WriteFieldHeader(8, 0)
			enc.WriteVarint(task.ExpireAt)
			enc.WriteFieldHeader(9, 2)
			enc.WriteString(task.Title)
		}
	}
}
func BenchmarkEncode_TaskList_Proto(b *testing.B) {
	tasks := benchMakeTasks(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeTasksProto(tasks)
	}
}
func BenchmarkEncode_TaskList_JSON(b *testing.B) {
	tasks := benchMakeTasks(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeJSON(tasks)
	}
}
func BenchmarkEncode_TaskList_Msgpack(b *testing.B) {
	tasks := benchMakeTasks(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeMsgpack(tasks)
	}
}

func BenchmarkEncode_Leaderboard_ByteMsg(b *testing.B) {
	lb := benchMakeLeaderboard()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeLeaderboardBmsg2(lb)
	}
}
func BenchmarkEncode_Leaderboard_Proto(b *testing.B) {
	lb := benchMakeLeaderboard()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeLeaderboardProto(lb)
	}
}
func BenchmarkEncode_Leaderboard_JSON(b *testing.B) {
	lb := benchMakeLeaderboard()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeJSON(lb)
	}
}
func BenchmarkEncode_Leaderboard_Msgpack(b *testing.B) {
	lb := benchMakeLeaderboard()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeMsgpack(lb)
	}
}

// ==================== 解码 Benchmark ====================

func BenchmarkDecode_Player_ByteMsg(b *testing.B) {
	data := encodePlayerBmsg(benchMakePlayer())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decodePlayerBmsg(data)
	}
}
func BenchmarkDecode_Player_Proto(b *testing.B) {
	data := encodePlayerProto(benchMakePlayer())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decodePlayerProto(data)
	}
}
func BenchmarkDecode_Player_JSON(b *testing.B) {
	var p BenchPlayer
	data := encodeJSON(p)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal(data, &p)
	}
}
func BenchmarkDecode_Player_Msgpack(b *testing.B) {
	var p BenchPlayer
	data := encodeMsgpack(p)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msgpack.Unmarshal(data, &p)
	}
}

func BenchmarkDecode_Chat_ByteMsg(b *testing.B) {
	data := encodeChatBmsg2(benchMakeChat())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decodeChatBmsg(data)
	}
}
func BenchmarkDecode_Chat_Proto(b *testing.B) {
	data := encodeChatProto(benchMakeChat())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decodeChatProto(data)
	}
}
func BenchmarkDecode_Chat_JSON(b *testing.B) {
	var c BenchChatMsg
	data := encodeJSON(c)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal(data, &c)
	}
}
func BenchmarkDecode_Chat_Msgpack(b *testing.B) {
	var c BenchChatMsg
	data := encodeMsgpack(c)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msgpack.Unmarshal(data, &c)
	}
}

func BenchmarkDecode_Battle_ByteMsg(b *testing.B) {
	data := encodeInputsBmsg(benchMakeBattleInputs())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(bytes.NewReader(data))
		cnt, _ := dec.ReadVarint()
		for j := uint64(0); j < cnt; j++ {
			for {
				tag, _, err := dec.ReadFieldHeader()
				if err != nil || tag == 0 {
					break
				}
				dec.ReadVarint()
			}
		}
	}
}
func BenchmarkDecode_Battle_Msgpack(b *testing.B) {
	var in []BenchBattleInput
	data := encodeMsgpack(in)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msgpack.Unmarshal(data, &in)
	}
}
func BenchmarkDecode_Battle_JSON(b *testing.B) {
	var in []BenchBattleInput
	data := encodeJSON(in)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal(data, &in)
	}
}
