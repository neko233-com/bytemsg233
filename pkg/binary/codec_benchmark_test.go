package binary

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"sort"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/encoding/protowire"
)

var benchChatArgKeys = []string{"boss_id", "map", "phase", "voice"}

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

type BenchChatDto struct {
	MsgId     uint64            `json:"msg_id" msgpack:"msg_id"`
	Channel   uint32            `json:"channel" msgpack:"channel"`
	Sender    BenchChatSender   `json:"sender" msgpack:"sender"`
	Content   string            `json:"content" msgpack:"content"`
	Lang      string            `json:"lang" msgpack:"lang"`
	CreatedAt int64             `json:"created_at" msgpack:"created_at"`
	Edited    bool              `json:"edited" msgpack:"edited"`
	Priority  int32             `json:"priority" msgpack:"priority"`
	Heat      float32           `json:"heat" msgpack:"heat"`
	Score     float64           `json:"score" msgpack:"score"`
	Raw       []byte            `json:"raw" msgpack:"raw"`
	Tags      []string          `json:"tags" msgpack:"tags"`
	Mentions  []uint64          `json:"mentions" msgpack:"mentions"`
	Args      map[string]string `json:"args" msgpack:"args"`
	Items     []BenchChatItem   `json:"items" msgpack:"items"`
	Reply     BenchChatReply    `json:"reply" msgpack:"reply"`
}

type BenchChatSender struct {
	Uid    uint64 `json:"uid" msgpack:"uid"`
	Name   string `json:"name" msgpack:"name"`
	Level  uint32 `json:"level" msgpack:"level"`
	Vip    uint32 `json:"vip" msgpack:"vip"`
	Guild  string `json:"guild" msgpack:"guild"`
	Online bool   `json:"online" msgpack:"online"`
}

type BenchChatItem struct {
	ItemId uint32 `json:"item_id" msgpack:"item_id"`
	Count  uint32 `json:"count" msgpack:"count"`
	Rare   bool   `json:"rare" msgpack:"rare"`
}

type BenchChatReply struct {
	MsgId   uint64 `json:"msg_id" msgpack:"msg_id"`
	Summary string `json:"summary" msgpack:"summary"`
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

func benchMakeChatDto() BenchChatDto {
	return BenchChatDto{
		MsgId:   8800000001,
		Channel: 3,
		Sender: BenchChatSender{
			Uid: 100000001, Name: "绝影·暗夜猎手", Level: 65, Vip: 8, Guild: "苍穹之巅", Online: true,
		},
		Content:   "集合！Boss 还剩 30%，战士开盾，奶妈留大招。",
		Lang:      "zh-CN",
		CreatedAt: 1718304000,
		Edited:    true,
		Priority:  -2,
		Heat:      0.875,
		Score:     9981.25,
		Raw:       []byte{0x08, 0x7b, 0x12, 0x04, 0x4e, 0x65, 0x6b, 0x6f},
		Tags:      []string{"raid", "boss", "guild"},
		Mentions:  []uint64{100000002, 100000003, 100000004},
		Args: map[string]string{
			"boss_id": "90001",
			"phase":   "3",
			"map":     "dragon_cave",
			"voice":   "guild",
		},
		Items: []BenchChatItem{
			{ItemId: 60001, Count: 3, Rare: true},
			{ItemId: 60002, Count: 15, Rare: false},
		},
		Reply: BenchChatReply{MsgId: 8799999999, Summary: "上一条：等人齐再开"},
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

// ==================== ByteMsg233 编码/解码 ====================

func encodePlayerBmsg(p BenchPlayer) []byte {
	buf := make([]byte, 0, 64)
	buf = AppendFieldHeader(buf, 1, 0)
	buf = AppendVarint(buf, p.Uid)
	buf = AppendFieldHeader(buf, 2, 2)
	buf = AppendString(buf, p.Name)
	buf = AppendFieldHeader(buf, 3, 0)
	buf = AppendVarint(buf, uint64(p.Level))
	buf = AppendFieldHeader(buf, 4, 0)
	buf = AppendVarint(buf, uint64(p.VipLevel))
	buf = AppendFieldHeader(buf, 5, 0)
	buf = AppendVarint(buf, uint64(p.Diamond))
	buf = AppendFieldHeader(buf, 6, 0)
	buf = AppendVarint(buf, p.Gold)
	buf = AppendFieldHeader(buf, 7, 0)
	buf = AppendVarint(buf, uint64(p.Energy))
	buf = AppendFieldHeader(buf, 8, 0)
	buf = AppendVarint(buf, uint64(p.Avatar))
	buf = AppendFieldHeader(buf, 9, 0)
	buf = AppendVarint(buf, uint64(p.GuildId))
	buf = AppendFieldHeader(buf, 10, 2)
	buf = AppendString(buf, p.GuildName)
	return buf
}

func decodePlayerBmsg(data []byte) BenchPlayer {
	dec := NewSliceDecoder(data)
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
			v, _ := dec.ReadStringView()
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
			v, _ := dec.ReadStringView()
			p.GuildName = v
		default:
			return p
		}
	}
	return p
}

func encodeChatBmsg2(c BenchChatMsg) []byte {
	buf := make([]byte, 0, 64)
	buf = AppendFieldHeader(buf, 1, 0)
	buf = AppendVarint(buf, uint64(c.Channel))
	buf = AppendFieldHeader(buf, 2, 0)
	buf = AppendVarint(buf, uint64(c.SenderId))
	buf = AppendFieldHeader(buf, 3, 2)
	buf = AppendString(buf, c.Sender)
	buf = AppendFieldHeader(buf, 4, 2)
	buf = AppendString(buf, c.Content)
	buf = AppendFieldHeader(buf, 5, 0)
	buf = AppendVarint(buf, c.Time)
	return buf
}

func decodeChatBmsg(data []byte) BenchChatMsg {
	dec := NewSliceDecoder(data)
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
			v, _ := dec.ReadStringView()
			c.Sender = v
		case tag == 4 && wt == 2:
			v, _ := dec.ReadStringView()
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

func encodeChatDtoBmsg(c BenchChatDto) []byte {
	buf := GetBuffer()
	defer PutBuffer(buf)
	enc := NewEncoder(buf)
	enc.WriteFieldHeader(1, 0)
	enc.WriteVarint(c.MsgId)
	enc.WriteFieldHeader(2, 0)
	enc.WriteVarint(uint64(c.Channel))
	enc.WriteFieldHeader(3, 2)
	enc.WriteBytes(encodeChatSenderBmsg(c.Sender))
	enc.WriteFieldHeader(4, 2)
	enc.WriteString(c.Content)
	enc.WriteFieldHeader(5, 2)
	enc.WriteString(c.Lang)
	enc.WriteFieldHeader(6, 0)
	enc.WriteZigzag(c.CreatedAt)
	if c.Edited {
		enc.WriteFieldHeader(7, 0)
		enc.WriteVarint(1)
	}
	enc.WriteFieldHeader(8, 0)
	enc.WriteZigzag(int64(c.Priority))
	enc.WriteFieldHeader(9, 5)
	enc.WriteFixed32(math.Float32bits(c.Heat))
	enc.WriteFieldHeader(10, 1)
	enc.WriteFixed64(math.Float64bits(c.Score))
	enc.WriteFieldHeader(11, 2)
	enc.WriteBytes(c.Raw)
	enc.WriteFieldHeader(12, 2)
	enc.WriteBytes(encodeStringListBmsg(c.Tags))
	enc.WriteFieldHeader(13, 2)
	enc.WriteBytes(encodeUint64ListBmsg(c.Mentions))
	enc.WriteFieldHeader(14, 2)
	enc.WriteBytes(encodeStringMapBmsg(c.Args))
	enc.WriteFieldHeader(15, 2)
	enc.WriteBytes(encodeChatItemsBmsg(c.Items))
	enc.WriteFieldHeader(16, 2)
	enc.WriteBytes(encodeChatReplyBmsg(c.Reply))
	return append([]byte(nil), buf.Bytes()...)
}

func encodeChatDtoBmsgAppend(dst []byte, c BenchChatDto) []byte {
	dst = appendHeaderRaw(dst, 1, 0)
	dst = appendVarintRaw(dst, c.MsgId)
	dst = appendHeaderRaw(dst, 2, 0)
	dst = appendVarintRaw(dst, uint64(c.Channel))
	senderSize := chatSenderBmsgSize(c.Sender)
	dst = appendHeaderRaw(dst, 3, 2)
	dst = appendVarintRaw(dst, uint64(senderSize))
	dst = appendChatSenderBmsgRaw(dst, c.Sender)
	dst = appendHeaderRaw(dst, 4, 2)
	dst = appendStringRaw(dst, c.Content)
	dst = appendHeaderRaw(dst, 5, 2)
	dst = appendStringRaw(dst, c.Lang)
	dst = appendHeaderRaw(dst, 6, 0)
	dst = appendVarintRaw(dst, ZigzagEncode(c.CreatedAt))
	if c.Edited {
		dst = appendHeaderRaw(dst, 7, 0)
		dst = appendVarintRaw(dst, 1)
	}
	dst = appendHeaderRaw(dst, 8, 0)
	dst = appendVarintRaw(dst, ZigzagEncode(int64(c.Priority)))
	dst = appendHeaderRaw(dst, 9, 5)
	dst = appendFixed32Raw(dst, math.Float32bits(c.Heat))
	dst = appendHeaderRaw(dst, 10, 1)
	dst = appendFixed64Raw(dst, math.Float64bits(c.Score))
	dst = appendHeaderRaw(dst, 11, 2)
	dst = appendBytesRaw(dst, c.Raw)
	tagsSize := stringListBmsgPayloadSize(c.Tags)
	dst = appendHeaderRaw(dst, 12, 2)
	dst = appendVarintRaw(dst, uint64(tagsSize))
	dst = appendStringListBmsgRaw(dst, c.Tags)
	mentionsSize := uint64ListBmsgPayloadSize(c.Mentions)
	dst = appendHeaderRaw(dst, 13, 2)
	dst = appendVarintRaw(dst, uint64(mentionsSize))
	dst = appendUint64ListBmsgRaw(dst, c.Mentions)
	argsSize := stringMapBmsgPayloadSize(c.Args)
	dst = appendHeaderRaw(dst, 14, 2)
	dst = appendVarintRaw(dst, uint64(argsSize))
	dst = appendStringMapBmsgRaw(dst, c.Args)
	itemsSize := chatItemsBmsgPayloadSize(c.Items)
	dst = appendHeaderRaw(dst, 15, 2)
	dst = appendVarintRaw(dst, uint64(itemsSize))
	dst = appendChatItemsBmsgRaw(dst, c.Items)
	replySize := chatReplyBmsgSize(c.Reply)
	dst = appendHeaderRaw(dst, 16, 2)
	dst = appendVarintRaw(dst, uint64(replySize))
	dst = appendChatReplyBmsgRaw(dst, c.Reply)
	return dst
}

func appendChatSenderBmsgRaw(dst []byte, s BenchChatSender) []byte {
	dst = appendHeaderRaw(dst, 1, 0)
	dst = appendVarintRaw(dst, s.Uid)
	dst = appendHeaderRaw(dst, 2, 2)
	dst = appendStringRaw(dst, s.Name)
	dst = appendHeaderRaw(dst, 3, 0)
	dst = appendVarintRaw(dst, uint64(s.Level))
	dst = appendHeaderRaw(dst, 4, 0)
	dst = appendVarintRaw(dst, uint64(s.Vip))
	dst = appendHeaderRaw(dst, 5, 2)
	dst = appendStringRaw(dst, s.Guild)
	if s.Online {
		dst = appendHeaderRaw(dst, 6, 0)
		dst = appendVarintRaw(dst, 1)
	}
	return dst
}

func appendChatReplyBmsgRaw(dst []byte, r BenchChatReply) []byte {
	dst = appendHeaderRaw(dst, 1, 0)
	dst = appendVarintRaw(dst, r.MsgId)
	dst = appendHeaderRaw(dst, 2, 2)
	dst = appendStringRaw(dst, r.Summary)
	return dst
}

func appendChatItemsBmsgRaw(dst []byte, items []BenchChatItem) []byte {
	dst = appendVarintRaw(dst, uint64(len(items)))
	for _, item := range items {
		itemSize := chatItemBmsgSize(item)
		dst = appendVarintRaw(dst, uint64(itemSize))
		dst = appendChatItemBmsgRaw(dst, item)
	}
	return dst
}

func appendChatItemBmsgRaw(dst []byte, item BenchChatItem) []byte {
	dst = appendHeaderRaw(dst, 1, 0)
	dst = appendVarintRaw(dst, uint64(item.ItemId))
	dst = appendHeaderRaw(dst, 2, 0)
	dst = appendVarintRaw(dst, uint64(item.Count))
	if item.Rare {
		dst = appendHeaderRaw(dst, 3, 0)
		dst = appendVarintRaw(dst, 1)
	}
	return dst
}

func appendStringListBmsgRaw(dst []byte, values []string) []byte {
	dst = appendVarintRaw(dst, uint64(len(values)))
	for _, value := range values {
		dst = appendStringRaw(dst, value)
	}
	return dst
}

func appendUint64ListBmsgRaw(dst []byte, values []uint64) []byte {
	dst = appendVarintRaw(dst, uint64(len(values)))
	for _, value := range values {
		dst = appendVarintRaw(dst, value)
	}
	return dst
}

func appendStringMapBmsgRaw(dst []byte, values map[string]string) []byte {
	dst = appendVarintRaw(dst, uint64(len(values)))
	for _, key := range benchChatArgKeys {
		dst = appendStringRaw(dst, key)
		dst = appendStringRaw(dst, values[key])
	}
	return dst
}

func appendHeaderRaw(dst []byte, tag int, wireType int) []byte {
	return appendVarintRaw(dst, uint64(tag<<3|wireType))
}

func appendVarintRaw(dst []byte, value uint64) []byte {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], value)
	return append(dst, buf[:n]...)
}

func appendStringRaw(dst []byte, value string) []byte {
	dst = appendVarintRaw(dst, uint64(len(value)))
	return append(dst, value...)
}

func appendBytesRaw(dst []byte, value []byte) []byte {
	dst = appendVarintRaw(dst, uint64(len(value)))
	return append(dst, value...)
}

func appendFixed32Raw(dst []byte, value uint32) []byte {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], value)
	return append(dst, buf[:]...)
}

func appendFixed64Raw(dst []byte, value uint64) []byte {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], value)
	return append(dst, buf[:]...)
}

func chatSenderBmsgSize(s BenchChatSender) int {
	size := fieldVarintSize(1, s.Uid)
	size += fieldStringSize(2, s.Name)
	size += fieldVarintSize(3, uint64(s.Level))
	size += fieldVarintSize(4, uint64(s.Vip))
	size += fieldStringSize(5, s.Guild)
	if s.Online {
		size += fieldVarintSize(6, 1)
	}
	return size
}

func chatReplyBmsgSize(r BenchChatReply) int {
	return fieldVarintSize(1, r.MsgId) + fieldStringSize(2, r.Summary)
}

func chatItemBmsgSize(item BenchChatItem) int {
	size := fieldVarintSize(1, uint64(item.ItemId))
	size += fieldVarintSize(2, uint64(item.Count))
	if item.Rare {
		size += fieldVarintSize(3, 1)
	}
	return size
}

func chatItemsBmsgPayloadSize(items []BenchChatItem) int {
	size := benchVarintSize(uint64(len(items)))
	for _, item := range items {
		itemSize := chatItemBmsgSize(item)
		size += benchVarintSize(uint64(itemSize)) + itemSize
	}
	return size
}

func stringListBmsgPayloadSize(values []string) int {
	size := benchVarintSize(uint64(len(values)))
	for _, value := range values {
		size += stringSize(value)
	}
	return size
}

func uint64ListBmsgPayloadSize(values []uint64) int {
	size := benchVarintSize(uint64(len(values)))
	for _, value := range values {
		size += benchVarintSize(value)
	}
	return size
}

func stringMapBmsgPayloadSize(values map[string]string) int {
	size := benchVarintSize(uint64(len(values)))
	for _, key := range benchChatArgKeys {
		size += stringSize(key)
		size += stringSize(values[key])
	}
	return size
}

func fieldVarintSize(tag int, value uint64) int {
	return benchVarintSize(uint64(tag<<3|0)) + benchVarintSize(value)
}

func fieldStringSize(tag int, value string) int {
	return benchVarintSize(uint64(tag<<3|2)) + stringSize(value)
}

func stringSize(value string) int {
	return benchVarintSize(uint64(len(value))) + len(value)
}

func benchVarintSize(value uint64) int {
	size := 1
	for value >= 0x80 {
		value >>= 7
		size++
	}
	return size
}

func encodeChatSenderBmsg(s BenchChatSender) []byte {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.WriteFieldHeader(1, 0)
	enc.WriteVarint(s.Uid)
	enc.WriteFieldHeader(2, 2)
	enc.WriteString(s.Name)
	enc.WriteFieldHeader(3, 0)
	enc.WriteVarint(uint64(s.Level))
	enc.WriteFieldHeader(4, 0)
	enc.WriteVarint(uint64(s.Vip))
	enc.WriteFieldHeader(5, 2)
	enc.WriteString(s.Guild)
	if s.Online {
		enc.WriteFieldHeader(6, 0)
		enc.WriteVarint(1)
	}
	return buf.Bytes()
}

func encodeChatReplyBmsg(r BenchChatReply) []byte {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.WriteFieldHeader(1, 0)
	enc.WriteVarint(r.MsgId)
	enc.WriteFieldHeader(2, 2)
	enc.WriteString(r.Summary)
	return buf.Bytes()
}

func encodeChatItemsBmsg(items []BenchChatItem) []byte {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.WriteVarint(uint64(len(items)))
	for _, item := range items {
		enc.WriteBytes(encodeChatItemBmsg(item))
	}
	return buf.Bytes()
}

func encodeChatItemBmsg(item BenchChatItem) []byte {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.WriteFieldHeader(1, 0)
	enc.WriteVarint(uint64(item.ItemId))
	enc.WriteFieldHeader(2, 0)
	enc.WriteVarint(uint64(item.Count))
	if item.Rare {
		enc.WriteFieldHeader(3, 0)
		enc.WriteVarint(1)
	}
	return buf.Bytes()
}

func encodeStringListBmsg(values []string) []byte {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.WriteVarint(uint64(len(values)))
	for _, value := range values {
		enc.WriteString(value)
	}
	return buf.Bytes()
}

func encodeUint64ListBmsg(values []uint64) []byte {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.WriteVarint(uint64(len(values)))
	for _, value := range values {
		enc.WriteVarint(value)
	}
	return buf.Bytes()
}

func encodeStringMapBmsg(values map[string]string) []byte {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	enc.WriteVarint(uint64(len(keys)))
	for _, key := range keys {
		enc.WriteString(key)
		enc.WriteString(values[key])
	}
	return buf.Bytes()
}

func decodeChatDtoBmsg(data []byte) BenchChatDto {
	dec := NewSliceDecoder(data)
	var c BenchChatDto
	for {
		tag, wt, err := dec.ReadFieldHeader()
		if err != nil {
			break
		}
		switch {
		case tag == 1 && wt == 0:
			c.MsgId, _ = dec.ReadVarint()
		case tag == 2 && wt == 0:
			v, _ := dec.ReadVarint()
			c.Channel = uint32(v)
		case tag == 3 && wt == 2:
			v, _ := dec.ReadBytesView()
			c.Sender = decodeChatSenderBmsg(v)
		case tag == 4 && wt == 2:
			c.Content, _ = dec.ReadStringView()
		case tag == 5 && wt == 2:
			c.Lang, _ = dec.ReadStringView()
		case tag == 6 && wt == 0:
			c.CreatedAt, _ = dec.ReadZigzag()
		case tag == 7 && wt == 0:
			v, _ := dec.ReadVarint()
			c.Edited = v != 0
		case tag == 8 && wt == 0:
			v, _ := dec.ReadZigzag()
			c.Priority = int32(v)
		case tag == 9 && wt == 5:
			v, _ := dec.ReadFixed32()
			c.Heat = math.Float32frombits(v)
		case tag == 10 && wt == 1:
			v, _ := dec.ReadFixed64()
			c.Score = math.Float64frombits(v)
		case tag == 11 && wt == 2:
			c.Raw, _ = dec.ReadBytesView()
		case tag == 12 && wt == 2:
			v, _ := dec.ReadBytesView()
			c.Tags = decodeStringListBmsg(v)
		case tag == 13 && wt == 2:
			v, _ := dec.ReadBytesView()
			c.Mentions = decodeUint64ListBmsg(v)
		case tag == 14 && wt == 2:
			v, _ := dec.ReadBytesView()
			c.Args = decodeStringMapBmsg(v)
		case tag == 15 && wt == 2:
			v, _ := dec.ReadBytesView()
			c.Items = decodeChatItemsBmsg(v)
		case tag == 16 && wt == 2:
			v, _ := dec.ReadBytesView()
			c.Reply = decodeChatReplyBmsg(v)
		default:
			return c
		}
	}
	return c
}

func decodeChatSenderBmsg(data []byte) BenchChatSender {
	dec := NewSliceDecoder(data)
	var s BenchChatSender
	for {
		tag, wt, err := dec.ReadFieldHeader()
		if err != nil {
			break
		}
		switch {
		case tag == 1 && wt == 0:
			s.Uid, _ = dec.ReadVarint()
		case tag == 2 && wt == 2:
			s.Name, _ = dec.ReadStringView()
		case tag == 3 && wt == 0:
			v, _ := dec.ReadVarint()
			s.Level = uint32(v)
		case tag == 4 && wt == 0:
			v, _ := dec.ReadVarint()
			s.Vip = uint32(v)
		case tag == 5 && wt == 2:
			s.Guild, _ = dec.ReadStringView()
		case tag == 6 && wt == 0:
			v, _ := dec.ReadVarint()
			s.Online = v != 0
		default:
			return s
		}
	}
	return s
}

func decodeChatReplyBmsg(data []byte) BenchChatReply {
	dec := NewSliceDecoder(data)
	var r BenchChatReply
	for {
		tag, wt, err := dec.ReadFieldHeader()
		if err != nil {
			break
		}
		switch {
		case tag == 1 && wt == 0:
			r.MsgId, _ = dec.ReadVarint()
		case tag == 2 && wt == 2:
			r.Summary, _ = dec.ReadStringView()
		default:
			return r
		}
	}
	return r
}

func decodeChatItemsBmsg(data []byte) []BenchChatItem {
	dec := NewSliceDecoder(data)
	count, _ := dec.ReadVarint()
	items := make([]BenchChatItem, 0, count)
	for i := uint64(0); i < count; i++ {
		raw, _ := dec.ReadBytesView()
		items = append(items, decodeChatItemBmsg(raw))
	}
	return items
}

func decodeChatItemBmsg(data []byte) BenchChatItem {
	dec := NewSliceDecoder(data)
	var item BenchChatItem
	for {
		tag, wt, err := dec.ReadFieldHeader()
		if err != nil {
			break
		}
		switch {
		case tag == 1 && wt == 0:
			v, _ := dec.ReadVarint()
			item.ItemId = uint32(v)
		case tag == 2 && wt == 0:
			v, _ := dec.ReadVarint()
			item.Count = uint32(v)
		case tag == 3 && wt == 0:
			v, _ := dec.ReadVarint()
			item.Rare = v != 0
		default:
			return item
		}
	}
	return item
}

func decodeStringListBmsg(data []byte) []string {
	dec := NewSliceDecoder(data)
	count, _ := dec.ReadVarint()
	values := make([]string, 0, count)
	for i := uint64(0); i < count; i++ {
		value, _ := dec.ReadStringView()
		values = append(values, value)
	}
	return values
}

func decodeUint64ListBmsg(data []byte) []uint64 {
	dec := NewSliceDecoder(data)
	count, _ := dec.ReadVarint()
	values := make([]uint64, 0, count)
	for i := uint64(0); i < count; i++ {
		value, _ := dec.ReadVarint()
		values = append(values, value)
	}
	return values
}

func decodeStringMapBmsg(data []byte) map[string]string {
	dec := NewSliceDecoder(data)
	count, _ := dec.ReadVarint()
	values := make(map[string]string, count)
	for i := uint64(0); i < count; i++ {
		key, _ := dec.ReadStringView()
		value, _ := dec.ReadStringView()
		values[key] = value
	}
	return values
}

func encodeInputsBmsg(inputs []BenchBattleInput) []byte {
	buf := make([]byte, 0, 160)
	buf = AppendVarint(buf, uint64(len(inputs)))
	buf = appendBattlePlayerIDDeltaColumn(buf, inputs)
	buf = appendBattleHeroIDDeltaColumn(buf, inputs)
	buf = appendBattleActionColumn(buf, inputs)
	buf = appendBattleSkillDeltaColumn(buf, inputs)
	buf = appendBattleTargetColumn(buf, inputs)
	buf = appendBattleXColumn(buf, inputs)
	buf = appendBattleYColumn(buf, inputs)
	buf = appendBattleDirColumn(buf, inputs)
	return buf
}

func appendBattlePlayerIDDeltaColumn(buf []byte, inputs []BenchBattleInput) []byte {
	buf = AppendVarint(buf, uint64(len(inputs)))
	if len(inputs) == 0 {
		return buf
	}
	prev := uint64(inputs[0].PlayerId)
	buf = AppendVarint(buf, prev)
	for _, input := range inputs[1:] {
		value := uint64(input.PlayerId)
		buf = AppendZigzag(buf, int64(value)-int64(prev))
		prev = value
	}
	return buf
}

func appendBattleHeroIDDeltaColumn(buf []byte, inputs []BenchBattleInput) []byte {
	buf = AppendVarint(buf, uint64(len(inputs)))
	if len(inputs) == 0 {
		return buf
	}
	prev := uint64(inputs[0].HeroId)
	buf = AppendVarint(buf, prev)
	for _, input := range inputs[1:] {
		value := uint64(input.HeroId)
		buf = AppendZigzag(buf, int64(value)-int64(prev))
		prev = value
	}
	return buf
}

func appendBattleActionColumn(buf []byte, inputs []BenchBattleInput) []byte {
	buf = AppendVarint(buf, uint64(len(inputs)))
	for _, input := range inputs {
		buf = AppendVarint(buf, uint64(input.Action))
	}
	return buf
}

func appendBattleSkillDeltaColumn(buf []byte, inputs []BenchBattleInput) []byte {
	buf = AppendVarint(buf, uint64(len(inputs)))
	if len(inputs) == 0 {
		return buf
	}
	prev := uint64(inputs[0].SkillId)
	buf = AppendVarint(buf, prev)
	for _, input := range inputs[1:] {
		value := uint64(input.SkillId)
		buf = AppendZigzag(buf, int64(value)-int64(prev))
		prev = value
	}
	return buf
}

func appendBattleTargetColumn(buf []byte, inputs []BenchBattleInput) []byte {
	buf = AppendVarint(buf, uint64(len(inputs)))
	for _, input := range inputs {
		buf = AppendVarint(buf, uint64(input.TargetId))
	}
	return buf
}

func appendBattleXColumn(buf []byte, inputs []BenchBattleInput) []byte {
	buf = AppendVarint(buf, uint64(len(inputs)))
	for _, input := range inputs {
		buf = AppendZigzag(buf, int64(input.X))
	}
	return buf
}

func appendBattleYColumn(buf []byte, inputs []BenchBattleInput) []byte {
	buf = AppendVarint(buf, uint64(len(inputs)))
	for _, input := range inputs {
		buf = AppendZigzag(buf, int64(input.Y))
	}
	return buf
}

func appendBattleDirColumn(buf []byte, inputs []BenchBattleInput) []byte {
	buf = AppendVarint(buf, uint64(len(inputs)))
	for _, input := range inputs {
		buf = AppendVarint(buf, uint64(input.Dir))
	}
	return buf
}

func encodeLeaderboardBmsg2(entries []BenchRankEntry) []byte {
	buf := make([]byte, 0, 2048)
	buf = AppendVarint(buf, uint64(len(entries)))
	buf = appendRankDeltaColumn(buf, entries)
	buf = appendRankPlayerIDDeltaColumn(buf, entries)
	buf = appendRankNameColumn(buf, entries)
	buf = appendRankLevelColumn(buf, entries)
	buf = appendRankScoreDeltaColumn(buf, entries)
	buf = appendRankGuildColumn(buf, entries)
	return buf
}

func encodeTasksBmsg(tasks []BenchTaskDto) []byte {
	return encodeTasksBmsgColumn(tasks)
}

func encodeTasksBmsgTo(buf *bytes.Buffer, tasks []BenchTaskDto) {
	buf.Reset()
	_, _ = buf.Write(encodeTasksBmsgColumn(tasks))
}

func encodeTasksBmsgColumn(tasks []BenchTaskDto) []byte {
	buf := make([]byte, 0, len(tasks)*24)
	buf = AppendVarint(buf, uint64(len(tasks)))
	buf = appendTaskIDDeltaColumn(buf, tasks)
	buf = appendTaskTypeColumn(buf, tasks)
	buf = appendTaskStatusColumn(buf, tasks)
	buf = appendTaskProgressColumn(buf, tasks)
	buf = appendTaskTargetColumn(buf, tasks)
	buf = appendTaskRewardIDDeltaColumn(buf, tasks)
	buf = appendTaskRewardCountColumn(buf, tasks)
	buf = appendTaskExpireDeltaColumn(buf, tasks)
	buf = appendTaskTitleColumn(buf, tasks)
	return buf
}

func appendTaskIDDeltaColumn(buf []byte, tasks []BenchTaskDto) []byte {
	buf = AppendVarint(buf, uint64(len(tasks)))
	if len(tasks) == 0 {
		return buf
	}
	prev := uint64(tasks[0].TaskId)
	buf = AppendVarint(buf, prev)
	for _, task := range tasks[1:] {
		value := uint64(task.TaskId)
		buf = AppendZigzag(buf, int64(value)-int64(prev))
		prev = value
	}
	return buf
}

func appendTaskTypeColumn(buf []byte, tasks []BenchTaskDto) []byte {
	buf = AppendVarint(buf, uint64(len(tasks)))
	for _, task := range tasks {
		buf = AppendVarint(buf, uint64(task.Type))
	}
	return buf
}

func appendTaskStatusColumn(buf []byte, tasks []BenchTaskDto) []byte {
	buf = AppendVarint(buf, uint64(len(tasks)))
	for _, task := range tasks {
		buf = AppendVarint(buf, uint64(task.Status))
	}
	return buf
}

func appendTaskProgressColumn(buf []byte, tasks []BenchTaskDto) []byte {
	buf = AppendVarint(buf, uint64(len(tasks)))
	for _, task := range tasks {
		buf = AppendVarint(buf, uint64(task.Progress))
	}
	return buf
}

func appendTaskTargetColumn(buf []byte, tasks []BenchTaskDto) []byte {
	buf = AppendVarint(buf, uint64(len(tasks)))
	for _, task := range tasks {
		buf = AppendVarint(buf, uint64(task.Target))
	}
	return buf
}

func appendTaskRewardIDDeltaColumn(buf []byte, tasks []BenchTaskDto) []byte {
	buf = AppendVarint(buf, uint64(len(tasks)))
	if len(tasks) == 0 {
		return buf
	}
	prev := uint64(tasks[0].RewardId)
	buf = AppendVarint(buf, prev)
	for _, task := range tasks[1:] {
		value := uint64(task.RewardId)
		buf = AppendZigzag(buf, int64(value)-int64(prev))
		prev = value
	}
	return buf
}

func appendTaskRewardCountColumn(buf []byte, tasks []BenchTaskDto) []byte {
	buf = AppendVarint(buf, uint64(len(tasks)))
	for _, task := range tasks {
		buf = AppendVarint(buf, uint64(task.RewardCount))
	}
	return buf
}

func appendTaskExpireDeltaColumn(buf []byte, tasks []BenchTaskDto) []byte {
	buf = AppendVarint(buf, uint64(len(tasks)))
	if len(tasks) == 0 {
		return buf
	}
	prev := tasks[0].ExpireAt
	buf = AppendVarint(buf, prev)
	for _, task := range tasks[1:] {
		value := task.ExpireAt
		buf = AppendZigzag(buf, int64(value)-int64(prev))
		prev = value
	}
	return buf
}

func appendTaskTitleColumn(buf []byte, tasks []BenchTaskDto) []byte {
	buf = AppendVarint(buf, uint64(len(tasks)))
	for _, task := range tasks {
		buf = AppendString(buf, task.Title)
	}
	return buf
}

func appendRankDeltaColumn(buf []byte, entries []BenchRankEntry) []byte {
	buf = AppendVarint(buf, uint64(len(entries)))
	if len(entries) == 0 {
		return buf
	}
	prev := uint64(entries[0].Rank)
	buf = AppendVarint(buf, prev)
	for _, entry := range entries[1:] {
		value := uint64(entry.Rank)
		buf = AppendZigzag(buf, int64(value)-int64(prev))
		prev = value
	}
	return buf
}

func appendRankPlayerIDDeltaColumn(buf []byte, entries []BenchRankEntry) []byte {
	buf = AppendVarint(buf, uint64(len(entries)))
	if len(entries) == 0 {
		return buf
	}
	prev := entries[0].PlayerId
	buf = AppendVarint(buf, prev)
	for _, entry := range entries[1:] {
		value := entry.PlayerId
		buf = AppendZigzag(buf, int64(value)-int64(prev))
		prev = value
	}
	return buf
}

func appendRankNameColumn(buf []byte, entries []BenchRankEntry) []byte {
	buf = AppendVarint(buf, uint64(len(entries)))
	for _, entry := range entries {
		buf = AppendString(buf, entry.Name)
	}
	return buf
}

func appendRankLevelColumn(buf []byte, entries []BenchRankEntry) []byte {
	buf = AppendVarint(buf, uint64(len(entries)))
	for _, entry := range entries {
		buf = AppendVarint(buf, uint64(entry.Level))
	}
	return buf
}

func appendRankScoreDeltaColumn(buf []byte, entries []BenchRankEntry) []byte {
	buf = AppendVarint(buf, uint64(len(entries)))
	if len(entries) == 0 {
		return buf
	}
	prev := entries[0].Score
	buf = AppendVarint(buf, prev)
	for _, entry := range entries[1:] {
		value := entry.Score
		buf = AppendZigzag(buf, int64(value)-int64(prev))
		prev = value
	}
	return buf
}

func appendRankGuildColumn(buf []byte, entries []BenchRankEntry) []byte {
	buf = AppendVarint(buf, uint64(len(entries)))
	for _, entry := range entries {
		buf = AppendString(buf, entry.Guild)
	}
	return buf
}

func decodeTasksBmsgColumn(data []byte) []BenchTaskDto {
	var state benchTaskColumnDecodeState
	return state.decode(data)
}

type benchBattleColumnDecodeState struct {
	inputs    []BenchBattleInput
	playerIds []uint64
	heroIds   []uint64
	actions   []uint64
	skillIds  []uint64
	targetIds []uint64
	xs        []int64
	ys        []int64
	dirs      []uint64
}

func (s *benchBattleColumnDecodeState) decode(data []byte) []BenchBattleInput {
	dec := NewSliceDecoder(data)
	count, _ := dec.ReadVarint()
	s.playerIds, _ = dec.ReadDeltaVarints(s.playerIds)
	s.heroIds, _ = dec.ReadDeltaVarints(s.heroIds)
	s.actions, _ = dec.ReadPackedVarints(s.actions)
	s.skillIds, _ = dec.ReadDeltaVarints(s.skillIds)
	s.targetIds, _ = dec.ReadPackedVarints(s.targetIds)
	s.xs, _ = dec.ReadPackedZigzags(s.xs)
	s.ys, _ = dec.ReadPackedZigzags(s.ys)
	s.dirs, _ = dec.ReadPackedVarints(s.dirs)
	if uint64(cap(s.inputs)) < count {
		s.inputs = make([]BenchBattleInput, int(count))
	} else {
		s.inputs = s.inputs[:int(count)]
	}
	for i := range s.inputs {
		s.inputs[i].PlayerId = uint32(s.playerIds[i])
		s.inputs[i].HeroId = uint32(s.heroIds[i])
		s.inputs[i].Action = uint32(s.actions[i])
		s.inputs[i].SkillId = uint32(s.skillIds[i])
		s.inputs[i].TargetId = uint32(s.targetIds[i])
		s.inputs[i].X = int32(s.xs[i])
		s.inputs[i].Y = int32(s.ys[i])
		s.inputs[i].Dir = uint32(s.dirs[i])
	}
	return s.inputs
}

func (s *benchBattleColumnDecodeState) prewarm(count int) {
	s.inputs = make([]BenchBattleInput, count)
	s.playerIds = make([]uint64, 0, count)
	s.heroIds = make([]uint64, 0, count)
	s.actions = make([]uint64, 0, count)
	s.skillIds = make([]uint64, 0, count)
	s.targetIds = make([]uint64, 0, count)
	s.xs = make([]int64, 0, count)
	s.ys = make([]int64, 0, count)
	s.dirs = make([]uint64, 0, count)
}

type benchLeaderboardColumnDecodeState struct {
	entries   []BenchRankEntry
	ranks     []uint64
	playerIds []uint64
	names     []string
	levels    []uint64
	scores    []uint64
	guilds    []string
}

func (s *benchLeaderboardColumnDecodeState) decode(data []byte) []BenchRankEntry {
	dec := NewSliceDecoder(data)
	count, _ := dec.ReadVarint()
	s.ranks, _ = dec.ReadDeltaVarints(s.ranks)
	s.playerIds, _ = dec.ReadDeltaVarints(s.playerIds)
	s.names, _ = dec.ReadStringList(s.names)
	s.levels, _ = dec.ReadPackedVarints(s.levels)
	s.scores, _ = dec.ReadDeltaVarints(s.scores)
	s.guilds, _ = dec.ReadStringList(s.guilds)
	if uint64(cap(s.entries)) < count {
		s.entries = make([]BenchRankEntry, int(count))
	} else {
		s.entries = s.entries[:int(count)]
	}
	for i := range s.entries {
		s.entries[i].Rank = uint32(s.ranks[i])
		s.entries[i].PlayerId = s.playerIds[i]
		s.entries[i].Name = s.names[i]
		s.entries[i].Level = uint32(s.levels[i])
		s.entries[i].Score = s.scores[i]
		s.entries[i].Guild = s.guilds[i]
	}
	return s.entries
}

func (s *benchLeaderboardColumnDecodeState) prewarm(count int) {
	s.entries = make([]BenchRankEntry, count)
	s.ranks = make([]uint64, 0, count)
	s.playerIds = make([]uint64, 0, count)
	s.names = make([]string, 0, count)
	s.levels = make([]uint64, 0, count)
	s.scores = make([]uint64, 0, count)
	s.guilds = make([]string, 0, count)
}

type benchTaskColumnDecodeState struct {
	tasks        []BenchTaskDto
	taskIds      []uint64
	types        []uint64
	statuses     []uint64
	progresses   []uint64
	targets      []uint64
	rewardIds    []uint64
	rewardCounts []uint64
	expireAts    []uint64
	titles       []string
}

func (s *benchTaskColumnDecodeState) decode(data []byte) []BenchTaskDto {
	dec := NewSliceDecoder(data)
	count, _ := dec.ReadVarint()
	s.taskIds, _ = dec.ReadDeltaVarints(s.taskIds)
	s.types, _ = dec.ReadPackedVarints(s.types)
	s.statuses, _ = dec.ReadPackedVarints(s.statuses)
	s.progresses, _ = dec.ReadPackedVarints(s.progresses)
	s.targets, _ = dec.ReadPackedVarints(s.targets)
	s.rewardIds, _ = dec.ReadDeltaVarints(s.rewardIds)
	s.rewardCounts, _ = dec.ReadPackedVarints(s.rewardCounts)
	s.expireAts, _ = dec.ReadDeltaVarints(s.expireAts)
	s.titles, _ = dec.ReadStringList(s.titles)
	if uint64(cap(s.tasks)) < count {
		s.tasks = make([]BenchTaskDto, int(count))
	} else {
		s.tasks = s.tasks[:int(count)]
	}
	for i := range s.tasks {
		s.tasks[i].TaskId = uint32(s.taskIds[i])
		s.tasks[i].Type = uint32(s.types[i])
		s.tasks[i].Status = uint32(s.statuses[i])
		s.tasks[i].Progress = uint32(s.progresses[i])
		s.tasks[i].Target = uint32(s.targets[i])
		s.tasks[i].RewardId = uint32(s.rewardIds[i])
		s.tasks[i].RewardCount = uint32(s.rewardCounts[i])
		s.tasks[i].ExpireAt = s.expireAts[i]
		s.tasks[i].Title = s.titles[i]
	}
	return s.tasks
}

func (s *benchTaskColumnDecodeState) prewarm(count int) {
	s.tasks = make([]BenchTaskDto, count)
	s.taskIds = make([]uint64, 0, count)
	s.types = make([]uint64, 0, count)
	s.statuses = make([]uint64, 0, count)
	s.progresses = make([]uint64, 0, count)
	s.targets = make([]uint64, 0, count)
	s.rewardIds = make([]uint64, 0, count)
	s.rewardCounts = make([]uint64, 0, count)
	s.expireAts = make([]uint64, 0, count)
	s.titles = make([]string, 0, count)
}

func decodeInputsProto(data []byte) []BenchBattleInput {
	inputs := make([]BenchBattleInput, 0, 10)
	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			break
		}
		data = data[n:]
		if num != 1 || typ != protowire.BytesType {
			n = protowire.ConsumeFieldValue(num, typ, data)
			if n < 0 {
				break
			}
			data = data[n:]
			continue
		}
		msg, n := protowire.ConsumeBytes(data)
		if n < 0 {
			break
		}
		data = data[n:]
		var input BenchBattleInput
		for len(msg) > 0 {
			field, fieldType, consumed := protowire.ConsumeTag(msg)
			if consumed < 0 {
				break
			}
			msg = msg[consumed:]
			if fieldType != protowire.VarintType {
				consumed = protowire.ConsumeFieldValue(field, fieldType, msg)
				if consumed < 0 {
					break
				}
				msg = msg[consumed:]
				continue
			}
			v, consumed := protowire.ConsumeVarint(msg)
			if consumed < 0 {
				break
			}
			msg = msg[consumed:]
			switch field {
			case 1:
				input.PlayerId = uint32(v)
			case 2:
				input.HeroId = uint32(v)
			case 3:
				input.Action = uint32(v)
			case 4:
				input.SkillId = uint32(v)
			case 5:
				input.TargetId = uint32(v)
			case 6:
				input.X = int32(protowire.DecodeZigZag(v))
			case 7:
				input.Y = int32(protowire.DecodeZigZag(v))
			case 8:
				input.Dir = uint32(v)
			}
		}
		inputs = append(inputs, input)
	}
	return inputs
}

func decodeLeaderboardProto(data []byte) []BenchRankEntry {
	entries := make([]BenchRankEntry, 0, 100)
	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			break
		}
		data = data[n:]
		if num != 1 || typ != protowire.BytesType {
			n = protowire.ConsumeFieldValue(num, typ, data)
			if n < 0 {
				break
			}
			data = data[n:]
			continue
		}
		msg, n := protowire.ConsumeBytes(data)
		if n < 0 {
			break
		}
		data = data[n:]
		var entry BenchRankEntry
		for len(msg) > 0 {
			field, fieldType, consumed := protowire.ConsumeTag(msg)
			if consumed < 0 {
				break
			}
			msg = msg[consumed:]
			switch fieldType {
			case protowire.VarintType:
				v, consumed := protowire.ConsumeVarint(msg)
				if consumed < 0 {
					break
				}
				msg = msg[consumed:]
				switch field {
				case 1:
					entry.Rank = uint32(v)
				case 2:
					entry.PlayerId = v
				case 4:
					entry.Level = uint32(v)
				case 5:
					entry.Score = v
				}
			case protowire.BytesType:
				v, consumed := protowire.ConsumeBytes(msg)
				if consumed < 0 {
					break
				}
				msg = msg[consumed:]
				switch field {
				case 3:
					entry.Name = string(v)
				case 6:
					entry.Guild = string(v)
				}
			default:
				consumed = protowire.ConsumeFieldValue(field, fieldType, msg)
				if consumed < 0 {
					break
				}
				msg = msg[consumed:]
			}
		}
		entries = append(entries, entry)
	}
	return entries
}

func decodeTasksProto(data []byte) []BenchTaskDto {
	tasks := make([]BenchTaskDto, 0, 100)
	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			break
		}
		data = data[n:]
		if num != 1 || typ != protowire.BytesType {
			n = protowire.ConsumeFieldValue(num, typ, data)
			if n < 0 {
				break
			}
			data = data[n:]
			continue
		}
		msg, n := protowire.ConsumeBytes(data)
		if n < 0 {
			break
		}
		data = data[n:]
		var task BenchTaskDto
		for len(msg) > 0 {
			field, fieldType, consumed := protowire.ConsumeTag(msg)
			if consumed < 0 {
				break
			}
			msg = msg[consumed:]
			switch fieldType {
			case protowire.VarintType:
				v, consumed := protowire.ConsumeVarint(msg)
				if consumed < 0 {
					break
				}
				msg = msg[consumed:]
				switch field {
				case 1:
					task.TaskId = uint32(v)
				case 2:
					task.Type = uint32(v)
				case 3:
					task.Status = uint32(v)
				case 4:
					task.Progress = uint32(v)
				case 5:
					task.Target = uint32(v)
				case 6:
					task.RewardId = uint32(v)
				case 7:
					task.RewardCount = uint32(v)
				case 8:
					task.ExpireAt = v
				}
			case protowire.BytesType:
				v, consumed := protowire.ConsumeBytes(msg)
				if consumed < 0 {
					break
				}
				msg = msg[consumed:]
				if field == 9 {
					task.Title = string(v)
				}
			default:
				consumed = protowire.ConsumeFieldValue(field, fieldType, msg)
				if consumed < 0 {
					break
				}
				msg = msg[consumed:]
			}
		}
		tasks = append(tasks, task)
	}
	return tasks
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

func encodeChatDtoProto(c BenchChatDto) []byte {
	var buf []byte
	buf = protowire.AppendTag(buf, 1, protowire.VarintType)
	buf = protowire.AppendVarint(buf, c.MsgId)
	buf = protowire.AppendTag(buf, 2, protowire.VarintType)
	buf = protowire.AppendVarint(buf, uint64(c.Channel))
	buf = protowire.AppendTag(buf, 3, protowire.BytesType)
	buf = protowire.AppendBytes(buf, encodeChatSenderProto(c.Sender))
	buf = protowire.AppendTag(buf, 4, protowire.BytesType)
	buf = protowire.AppendString(buf, c.Content)
	buf = protowire.AppendTag(buf, 5, protowire.BytesType)
	buf = protowire.AppendString(buf, c.Lang)
	buf = protowire.AppendTag(buf, 6, protowire.VarintType)
	buf = protowire.AppendVarint(buf, uint64(protowire.EncodeZigZag(c.CreatedAt)))
	if c.Edited {
		buf = protowire.AppendTag(buf, 7, protowire.VarintType)
		buf = protowire.AppendVarint(buf, 1)
	}
	buf = protowire.AppendTag(buf, 8, protowire.VarintType)
	buf = protowire.AppendVarint(buf, uint64(protowire.EncodeZigZag(int64(c.Priority))))
	buf = protowire.AppendTag(buf, 9, protowire.Fixed32Type)
	buf = protowire.AppendFixed32(buf, math.Float32bits(c.Heat))
	buf = protowire.AppendTag(buf, 10, protowire.Fixed64Type)
	buf = protowire.AppendFixed64(buf, math.Float64bits(c.Score))
	buf = protowire.AppendTag(buf, 11, protowire.BytesType)
	buf = protowire.AppendBytes(buf, c.Raw)
	for _, tag := range c.Tags {
		buf = protowire.AppendTag(buf, 12, protowire.BytesType)
		buf = protowire.AppendString(buf, tag)
	}
	for _, mention := range c.Mentions {
		buf = protowire.AppendTag(buf, 13, protowire.VarintType)
		buf = protowire.AppendVarint(buf, mention)
	}
	keys := make([]string, 0, len(c.Args))
	for key := range c.Args {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		entry := protowire.AppendTag(nil, 1, protowire.BytesType)
		entry = protowire.AppendString(entry, key)
		entry = protowire.AppendTag(entry, 2, protowire.BytesType)
		entry = protowire.AppendString(entry, c.Args[key])
		buf = protowire.AppendTag(buf, 14, protowire.BytesType)
		buf = protowire.AppendBytes(buf, entry)
	}
	for _, item := range c.Items {
		buf = protowire.AppendTag(buf, 15, protowire.BytesType)
		buf = protowire.AppendBytes(buf, encodeChatItemProto(item))
	}
	buf = protowire.AppendTag(buf, 16, protowire.BytesType)
	buf = protowire.AppendBytes(buf, encodeChatReplyProto(c.Reply))
	return buf
}

func encodeChatSenderProto(s BenchChatSender) []byte {
	var buf []byte
	buf = protowire.AppendTag(buf, 1, protowire.VarintType)
	buf = protowire.AppendVarint(buf, s.Uid)
	buf = protowire.AppendTag(buf, 2, protowire.BytesType)
	buf = protowire.AppendString(buf, s.Name)
	buf = protowire.AppendTag(buf, 3, protowire.VarintType)
	buf = protowire.AppendVarint(buf, uint64(s.Level))
	buf = protowire.AppendTag(buf, 4, protowire.VarintType)
	buf = protowire.AppendVarint(buf, uint64(s.Vip))
	buf = protowire.AppendTag(buf, 5, protowire.BytesType)
	buf = protowire.AppendString(buf, s.Guild)
	if s.Online {
		buf = protowire.AppendTag(buf, 6, protowire.VarintType)
		buf = protowire.AppendVarint(buf, 1)
	}
	return buf
}

func encodeChatItemProto(item BenchChatItem) []byte {
	var buf []byte
	buf = protowire.AppendTag(buf, 1, protowire.VarintType)
	buf = protowire.AppendVarint(buf, uint64(item.ItemId))
	buf = protowire.AppendTag(buf, 2, protowire.VarintType)
	buf = protowire.AppendVarint(buf, uint64(item.Count))
	if item.Rare {
		buf = protowire.AppendTag(buf, 3, protowire.VarintType)
		buf = protowire.AppendVarint(buf, 1)
	}
	return buf
}

func encodeChatReplyProto(r BenchChatReply) []byte {
	var buf []byte
	buf = protowire.AppendTag(buf, 1, protowire.VarintType)
	buf = protowire.AppendVarint(buf, r.MsgId)
	buf = protowire.AppendTag(buf, 2, protowire.BytesType)
	buf = protowire.AppendString(buf, r.Summary)
	return buf
}

func decodeChatDtoProto(data []byte) BenchChatDto {
	var c BenchChatDto
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
				return c
			}
			data = data[n:]
			switch num {
			case 1:
				c.MsgId = v
			case 2:
				c.Channel = uint32(v)
			case 6:
				c.CreatedAt = protowire.DecodeZigZag(v)
			case 7:
				c.Edited = v != 0
			case 8:
				c.Priority = int32(protowire.DecodeZigZag(v))
			case 13:
				c.Mentions = append(c.Mentions, v)
			}
		case protowire.Fixed32Type:
			v, n := protowire.ConsumeFixed32(data)
			if n < 0 {
				return c
			}
			data = data[n:]
			if num == 9 {
				c.Heat = math.Float32frombits(v)
			}
		case protowire.Fixed64Type:
			v, n := protowire.ConsumeFixed64(data)
			if n < 0 {
				return c
			}
			data = data[n:]
			if num == 10 {
				c.Score = math.Float64frombits(v)
			}
		case protowire.BytesType:
			v, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return c
			}
			data = data[n:]
			switch num {
			case 3:
				c.Sender = decodeChatSenderProto(v)
			case 4:
				c.Content = string(v)
			case 5:
				c.Lang = string(v)
			case 11:
				c.Raw = append(c.Raw[:0], v...)
			case 12:
				c.Tags = append(c.Tags, string(v))
			case 14:
				if c.Args == nil {
					c.Args = make(map[string]string)
				}
				key, value := decodeStringMapEntryProto(v)
				c.Args[key] = value
			case 15:
				c.Items = append(c.Items, decodeChatItemProto(v))
			case 16:
				c.Reply = decodeChatReplyProto(v)
			}
		default:
			n = protowire.ConsumeFieldValue(num, typ, data)
			if n < 0 {
				return c
			}
			data = data[n:]
		}
	}
	return c
}

func decodeChatSenderProto(data []byte) BenchChatSender {
	var s BenchChatSender
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
				return s
			}
			data = data[n:]
			switch num {
			case 1:
				s.Uid = v
			case 3:
				s.Level = uint32(v)
			case 4:
				s.Vip = uint32(v)
			case 6:
				s.Online = v != 0
			}
		case protowire.BytesType:
			v, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return s
			}
			data = data[n:]
			switch num {
			case 2:
				s.Name = string(v)
			case 5:
				s.Guild = string(v)
			}
		}
	}
	return s
}

func decodeChatItemProto(data []byte) BenchChatItem {
	var item BenchChatItem
	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			break
		}
		data = data[n:]
		if typ != protowire.VarintType {
			continue
		}
		v, n := protowire.ConsumeVarint(data)
		if n < 0 {
			return item
		}
		data = data[n:]
		switch num {
		case 1:
			item.ItemId = uint32(v)
		case 2:
			item.Count = uint32(v)
		case 3:
			item.Rare = v != 0
		}
	}
	return item
}

func decodeChatReplyProto(data []byte) BenchChatReply {
	var r BenchChatReply
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
				return r
			}
			data = data[n:]
			if num == 1 {
				r.MsgId = v
			}
		case protowire.BytesType:
			v, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return r
			}
			data = data[n:]
			if num == 2 {
				r.Summary = string(v)
			}
		}
	}
	return r
}

func decodeStringMapEntryProto(data []byte) (string, string) {
	var key, value string
	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			break
		}
		data = data[n:]
		if typ != protowire.BytesType {
			continue
		}
		v, n := protowire.ConsumeBytes(data)
		if n < 0 {
			return key, value
		}
		data = data[n:]
		if num == 1 {
			key = string(v)
		} else if num == 2 {
			value = string(v)
		}
	}
	return key, value
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

type benchRPCEnvelope struct {
	PacketID uint32 `json:"packet_id" msgpack:"packet_id"`
	Sequence uint64 `json:"sequence" msgpack:"sequence"`
	Kind     uint32 `json:"kind" msgpack:"kind"`
	Flags    uint32 `json:"flags" msgpack:"flags"`
	Payload  []byte `json:"payload" msgpack:"payload"`
}

func benchMakeRPCEnvelope(payload []byte) benchRPCEnvelope {
	return benchRPCEnvelope{
		PacketID: 2001,
		Sequence: 99,
		Kind:     1,
		Flags:    1,
		Payload:  payload,
	}
}

func appendRPCEnvelopeBmsg(dst []byte, payload []byte) []byte {
	dst = AppendFieldHeader(dst, 1, WireTypeVarint)
	dst = AppendVarint(dst, 2001)
	dst = AppendFieldHeader(dst, 2, WireTypeVarint)
	dst = AppendVarint(dst, 99)
	dst = AppendFieldHeader(dst, 3, WireTypeVarint)
	dst = AppendVarint(dst, 1)
	dst = AppendFieldHeader(dst, 4, WireTypeVarint)
	dst = AppendVarint(dst, 1)
	dst = AppendFieldHeader(dst, 5, WireTypeLengthDelimited)
	return AppendBytes(dst, payload)
}

func encodeRPCEnvelopeBmsg(payload []byte) []byte {
	return appendRPCEnvelopeBmsg(make([]byte, 0, len(payload)+16), payload)
}

func encodeRPCEnvelopeProto(payload []byte) []byte {
	return appendRPCEnvelopeProto(make([]byte, 0, len(payload)+16), payload)
}

func appendRPCEnvelopeProto(dst []byte, payload []byte) []byte {
	dst = protowire.AppendTag(dst, 1, protowire.VarintType)
	dst = protowire.AppendVarint(dst, 2001)
	dst = protowire.AppendTag(dst, 2, protowire.VarintType)
	dst = protowire.AppendVarint(dst, 99)
	dst = protowire.AppendTag(dst, 3, protowire.VarintType)
	dst = protowire.AppendVarint(dst, 1)
	dst = protowire.AppendTag(dst, 4, protowire.VarintType)
	dst = protowire.AppendVarint(dst, 1)
	dst = protowire.AppendTag(dst, 5, protowire.BytesType)
	return protowire.AppendBytes(dst, payload)
}

func decodeRPCEnvelopeBmsgPayload(data []byte) []byte {
	dec := NewSliceDecoder(data)
	for !dec.EOF() {
		tag, wt, err := dec.ReadFieldHeader()
		if err != nil {
			return nil
		}
		if tag == 5 && wt == WireTypeLengthDelimited {
			payload, _ := dec.ReadBytesView()
			return payload
		}
		if err := dec.SkipField(wt); err != nil {
			return nil
		}
	}
	return nil
}

func decodeRPCEnvelopeProtoPayload(data []byte) []byte {
	for len(data) > 0 {
		num, typ, n := protowire.ConsumeTag(data)
		if n < 0 {
			return nil
		}
		data = data[n:]
		if num == 5 && typ == protowire.BytesType {
			payload, n := protowire.ConsumeBytes(data)
			if n < 0 {
				return nil
			}
			return payload
		}
		n = protowire.ConsumeFieldValue(num, typ, data)
		if n < 0 {
			return nil
		}
		data = data[n:]
	}
	return nil
}

// ==================== 测试: 体积对比 ====================

func TestBenchmark_SizeComparison(t *testing.T) {
	player := benchMakePlayer()
	chat := benchMakeChat()
	chatDto := benchMakeChatDto()
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
		{"ChatDto 全类型 (list/map/custom)", len(encodeChatDtoBmsg(chatDto)), len(encodeChatDtoProto(chatDto)), len(encodeJSON(chatDto)), len(encodeMsgpack(chatDto))},
		{"RPC envelope + ChatDto payload (1x)", len(encodeRPCEnvelopeBmsg(encodeChatDtoBmsg(chatDto))), len(encodeRPCEnvelopeProto(encodeChatDtoProto(chatDto))), len(encodeJSON(benchMakeRPCEnvelope(encodeJSON(chatDto)))), len(encodeMsgpack(benchMakeRPCEnvelope(encodeMsgpack(chatDto))))},
		{"战斗输入 (10人×8 fields)", len(encodeInputsBmsg(inputs)), len(encodeInputsProto(inputs)), len(encodeJSON(inputs)), len(encodeMsgpack(inputs))},
		{"任务列表 (100 TaskDto×9 fields)", len(encodeTasksBmsg(tasks)), len(encodeTasksProto(tasks)), len(encodeJSON(tasks)), len(encodeMsgpack(tasks))},
		{"排行榜 (100人×6 fields)", len(encodeLeaderboardBmsg2(lb)), len(encodeLeaderboardProto(lb)), len(encodeJSON(lb)), len(encodeMsgpack(lb))},
	}

	t.Logf("")
	t.Logf("╔══════════════════════════════════════════════════════════════════════╗")
	t.Logf("║                        体积对比 (bytes)                             ║")
	t.Logf("╠══════════════════════════════════════════════════════════════════════╣")
	t.Logf("║  %-30s │ %6s │ %6s │ %6s │ %6s ║", "场景", "ByteMsg233", "Proto", "JSON", "MsgPk")
	t.Logf("╠══════════════════════════════════════════════════════════════════════╣")
	for _, r := range rows {
		t.Logf("║  %-30s │ %6d │ %6d │ %6d │ %6d ║", r.Name, r.Bmsg, r.Proto, r.JSON, r.Mp)
	}
	t.Logf("╚══════════════════════════════════════════════════════════════════════╝")
	t.Logf("")
	t.Logf("  ByteMsg233 vs Protobuf:")
	for _, r := range rows {
		ratio := float64(r.Bmsg) / float64(r.Proto) * 100
		t.Logf("    %-24s  %.1f%%", r.Name, ratio)
	}
	t.Logf("")
	t.Logf("  ByteMsg233 vs JSON:")
	for _, r := range rows {
		saved := (1 - float64(r.Bmsg)/float64(r.JSON)) * 100
		t.Logf("    %-24s  -%.1f%%", r.Name, saved)
	}
}

func TestBenchmark_ChatDtoAllTypesRoundTrip(t *testing.T) {
	chat := benchMakeChatDto()
	bmsgBytes := encodeChatDtoBmsg(chat)
	appendBytes := encodeChatDtoBmsgAppend(make([]byte, 0, len(bmsgBytes)), chat)
	if !bytes.Equal(bmsgBytes, appendBytes) {
		t.Fatalf("ByteMsg233 append ChatDto bytes mismatch: %d != %d", len(appendBytes), len(bmsgBytes))
	}
	bmsg := decodeChatDtoBmsg(appendBytes)
	proto := decodeChatDtoProto(encodeChatDtoProto(chat))
	if bmsg.MsgId != chat.MsgId || bmsg.Sender.Uid != chat.Sender.Uid || len(bmsg.Tags) != len(chat.Tags) || len(bmsg.Args) != len(chat.Args) || len(bmsg.Items) != len(chat.Items) {
		t.Fatalf("ByteMsg233 ChatDto all-types roundtrip mismatch: %#v", bmsg)
	}
	if proto.MsgId != chat.MsgId || proto.Sender.Uid != chat.Sender.Uid || len(proto.Tags) != len(chat.Tags) || len(proto.Args) != len(chat.Args) || len(proto.Items) != len(chat.Items) {
		t.Fatalf("Protobuf ChatDto all-types roundtrip mismatch: %#v", proto)
	}
}

func TestBenchmark_RPCEnvelopeRoundTrip(t *testing.T) {
	payload := encodeChatDtoBmsg(benchMakeChatDto())
	envelope := encodeRPCEnvelopeBmsg(payload)
	got := decodeRPCEnvelopeBmsgPayload(envelope)
	if !bytes.Equal(got, payload) {
		t.Fatalf("RPC envelope payload mismatch")
	}
}

func TestBenchmark_TaskColumnRoundTrip(t *testing.T) {
	tasks := benchMakeTasks(100)
	data := encodeTasksBmsgColumn(tasks)
	got := decodeTasksBmsgColumn(data)
	if len(got) != len(tasks) {
		t.Fatalf("decoded task count = %d, want %d", len(got), len(tasks))
	}
	for i := range tasks {
		if got[i] != tasks[i] {
			t.Fatalf("task[%d] = %#v, want %#v", i, got[i], tasks[i])
		}
	}
}

func TestBenchmark_BattleColumnRoundTrip(t *testing.T) {
	inputs := benchMakeBattleInputs()
	data := encodeInputsBmsg(inputs)
	var state benchBattleColumnDecodeState
	got := state.decode(data)
	if len(got) != len(inputs) {
		t.Fatalf("decoded input count = %d, want %d", len(got), len(inputs))
	}
	for i := range inputs {
		if got[i] != inputs[i] {
			t.Fatalf("input[%d] = %#v, want %#v", i, got[i], inputs[i])
		}
	}
}

func TestBenchmark_LeaderboardColumnRoundTrip(t *testing.T) {
	entries := benchMakeLeaderboard()
	data := encodeLeaderboardBmsg2(entries)
	var state benchLeaderboardColumnDecodeState
	got := state.decode(data)
	if len(got) != len(entries) {
		t.Fatalf("decoded leaderboard count = %d, want %d", len(got), len(entries))
	}
	for i := range entries {
		if got[i] != entries[i] {
			t.Fatalf("entry[%d] = %#v, want %#v", i, got[i], entries[i])
		}
	}
}

// ==================== 编码 Benchmark ====================

func BenchmarkEncode_Player_ByteMsg233(b *testing.B) {
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

func BenchmarkEncode_Chat_ByteMsg233(b *testing.B) {
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

func BenchmarkEncode_ChatDtoAllTypes_ByteMsg233(b *testing.B) {
	c := benchMakeChatDto()
	dst := make([]byte, 0, len(encodeChatDtoBmsg(c)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = encodeChatDtoBmsgAppend(dst[:0], c)
	}
}
func BenchmarkEncode_ChatDtoAllTypes_Proto(b *testing.B) {
	c := benchMakeChatDto()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeChatDtoProto(c)
	}
}
func BenchmarkEncode_ChatDtoAllTypes_JSON(b *testing.B) {
	c := benchMakeChatDto()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeJSON(c)
	}
}
func BenchmarkEncode_ChatDtoAllTypes_Msgpack(b *testing.B) {
	c := benchMakeChatDto()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeMsgpack(c)
	}
}

func BenchmarkEncode_RPCEnvelope_ByteMsg233(b *testing.B) {
	c := benchMakeChatDto()
	payload := make([]byte, 0, len(encodeChatDtoBmsg(c)))
	envelope := make([]byte, 0, len(payload)+16)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		payload = encodeChatDtoBmsgAppend(payload[:0], c)
		_ = appendRPCEnvelopeBmsg(envelope[:0], payload)
	}
}

func BenchmarkEncode_RPCEnvelope_Proto(b *testing.B) {
	c := benchMakeChatDto()
	envelope := make([]byte, 0, len(encodeChatDtoProto(c))+16)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		payload := encodeChatDtoProto(c)
		_ = appendRPCEnvelopeProto(envelope[:0], payload)
	}
}

func BenchmarkEncode_RPCEnvelope_JSON(b *testing.B) {
	c := benchMakeChatDto()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		payload := encodeJSON(c)
		encodeJSON(benchMakeRPCEnvelope(payload))
	}
}

func BenchmarkEncode_RPCEnvelope_Msgpack(b *testing.B) {
	c := benchMakeChatDto()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		payload := encodeMsgpack(c)
		encodeMsgpack(benchMakeRPCEnvelope(payload))
	}
}

func BenchmarkEncode_Battle_ByteMsg233(b *testing.B) {
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

func BenchmarkEncode_TaskList_ByteMsg233(b *testing.B) {
	tasks := benchMakeTasks(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeTasksBmsgColumn(tasks)
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

func BenchmarkEncode_Leaderboard_ByteMsg233(b *testing.B) {
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

func BenchmarkDecode_Player_ByteMsg233(b *testing.B) {
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

func BenchmarkDecode_Chat_ByteMsg233(b *testing.B) {
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

func BenchmarkDecode_ChatDtoAllTypes_ByteMsg233(b *testing.B) {
	data := encodeChatDtoBmsg(benchMakeChatDto())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decodeChatDtoBmsg(data)
	}
}
func BenchmarkDecode_ChatDtoAllTypes_Proto(b *testing.B) {
	data := encodeChatDtoProto(benchMakeChatDto())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decodeChatDtoProto(data)
	}
}
func BenchmarkDecode_ChatDtoAllTypes_JSON(b *testing.B) {
	var c BenchChatDto
	data := encodeJSON(benchMakeChatDto())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal(data, &c)
	}
}
func BenchmarkDecode_ChatDtoAllTypes_Msgpack(b *testing.B) {
	var c BenchChatDto
	data := encodeMsgpack(benchMakeChatDto())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msgpack.Unmarshal(data, &c)
	}
}

func BenchmarkDecode_RPCEnvelope_ByteMsg233(b *testing.B) {
	data := encodeRPCEnvelopeBmsg(encodeChatDtoBmsg(benchMakeChatDto()))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		payload := decodeRPCEnvelopeBmsgPayload(data)
		decodeChatDtoBmsg(payload)
	}
}

func BenchmarkDecode_RPCEnvelope_Proto(b *testing.B) {
	data := encodeRPCEnvelopeProto(encodeChatDtoProto(benchMakeChatDto()))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		payload := decodeRPCEnvelopeProtoPayload(data)
		decodeChatDtoProto(payload)
	}
}

func BenchmarkDecode_RPCEnvelope_JSON(b *testing.B) {
	var envelope benchRPCEnvelope
	var c BenchChatDto
	data := encodeJSON(benchMakeRPCEnvelope(encodeJSON(benchMakeChatDto())))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal(data, &envelope)
		json.Unmarshal(envelope.Payload, &c)
	}
}

func BenchmarkDecode_RPCEnvelope_Msgpack(b *testing.B) {
	var envelope benchRPCEnvelope
	var c BenchChatDto
	data := encodeMsgpack(benchMakeRPCEnvelope(encodeMsgpack(benchMakeChatDto())))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msgpack.Unmarshal(data, &envelope)
		msgpack.Unmarshal(envelope.Payload, &c)
	}
}

func BenchmarkDecode_Battle_ByteMsg233(b *testing.B) {
	inputs := benchMakeBattleInputs()
	data := encodeInputsBmsg(inputs)
	var state benchBattleColumnDecodeState
	state.prewarm(len(inputs))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state.decode(data)
	}
}
func BenchmarkDecode_Battle_Proto(b *testing.B) {
	data := encodeInputsProto(benchMakeBattleInputs())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decodeInputsProto(data)
	}
}
func BenchmarkDecode_Battle_Msgpack(b *testing.B) {
	var in []BenchBattleInput
	data := encodeMsgpack(benchMakeBattleInputs())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msgpack.Unmarshal(data, &in)
	}
}
func BenchmarkDecode_Battle_JSON(b *testing.B) {
	var in []BenchBattleInput
	data := encodeJSON(benchMakeBattleInputs())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal(data, &in)
	}
}

func BenchmarkDecode_Leaderboard_ByteMsg233(b *testing.B) {
	entries := benchMakeLeaderboard()
	data := encodeLeaderboardBmsg2(entries)
	var state benchLeaderboardColumnDecodeState
	state.prewarm(len(entries))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state.decode(data)
	}
}

func BenchmarkDecode_Leaderboard_Proto(b *testing.B) {
	data := encodeLeaderboardProto(benchMakeLeaderboard())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decodeLeaderboardProto(data)
	}
}

func BenchmarkDecode_Leaderboard_JSON(b *testing.B) {
	var entries []BenchRankEntry
	data := encodeJSON(benchMakeLeaderboard())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal(data, &entries)
	}
}

func BenchmarkDecode_Leaderboard_Msgpack(b *testing.B) {
	var entries []BenchRankEntry
	data := encodeMsgpack(benchMakeLeaderboard())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msgpack.Unmarshal(data, &entries)
	}
}

func BenchmarkDecode_TaskList_ByteMsg233(b *testing.B) {
	tasks := benchMakeTasks(100)
	data := encodeTasksBmsgColumn(tasks)
	var state benchTaskColumnDecodeState
	state.prewarm(len(tasks))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		state.decode(data)
	}
}

func BenchmarkDecode_TaskList_Proto(b *testing.B) {
	data := encodeTasksProto(benchMakeTasks(100))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decodeTasksProto(data)
	}
}

func BenchmarkDecode_TaskList_JSON(b *testing.B) {
	var tasks []BenchTaskDto
	data := encodeJSON(benchMakeTasks(100))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Unmarshal(data, &tasks)
	}
}

func BenchmarkDecode_TaskList_Msgpack(b *testing.B) {
	var tasks []BenchTaskDto
	data := encodeMsgpack(benchMakeTasks(100))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msgpack.Unmarshal(data, &tasks)
	}
}
