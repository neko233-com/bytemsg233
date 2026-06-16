package binary

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
)

// ==================== 贴合业务的数据结构 ====================

// 登录推送 — 玩家登录时服务端一次性下发的全量数据
type LoginPush struct {
	Player     PlayerInfo     `json:"player" msgpack:"player"`
	Heroes     []HeroData     `json:"heroes" msgpack:"heroes"`
	Items      []ItemData     `json:"items" msgpack:"items"`
	Mail       []MailData     `json:"mail" msgpack:"mail"`
	Quests     []QuestData    `json:"quests" msgpack:"quests"`
	Settings   PlayerSettings `json:"settings" msgpack:"settings"`
	ServerTime int64          `json:"server_time" msgpack:"server_time"`
}

type PlayerInfo struct {
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

type HeroData struct {
	HeroId   uint32            `json:"hero_id" msgpack:"hero_id"`
	Level    uint32            `json:"level" msgpack:"level"`
	Star     uint32            `json:"star" msgpack:"star"`
	Grade    uint32            `json:"grade" msgpack:"grade"`
	Exp      uint64            `json:"exp" msgpack:"exp"`
	Skills   []SkillData       `json:"skills" msgpack:"skills"`
	Runes    map[uint32]uint32 `json:"runes" msgpack:"runes"`
	SkinId   uint32            `json:"skin_id" msgpack:"skin_id"`
	AwakeLv  uint32            `json:"awake_lv" msgpack:"awake_lv"`
	FavorLv  uint32            `json:"favor_lv" msgpack:"favor_lv"`
	FavorExp uint32            `json:"favor_exp" msgpack:"favor_exp"`
}

type SkillData struct {
	SkillId uint32 `json:"skill_id" msgpack:"skill_id"`
	Level   uint32 `json:"level" msgpack:"level"`
}

type ItemData struct {
	ItemId uint32 `json:"item_id" msgpack:"item_id"`
	Count  uint32 `json:"count" msgpack:"count"`
	Level  uint32 `json:"level" msgpack:"level"`
	Locked bool   `json:"locked" msgpack:"locked"`
	Exp    uint32 `json:"exp" msgpack:"exp"`
}

type MailData struct {
	MailId   uint64 `json:"mail_id" msgpack:"mail_id"`
	Type     uint32 `json:"type" msgpack:"type"`
	Title    string `json:"title" msgpack:"title"`
	Read     bool   `json:"read" msgpack:"read"`
	Taken    bool   `json:"taken" msgpack:"taken"`
	SendTime int64  `json:"send_time" msgpack:"send_time"`
	HasItems bool   `json:"has_items" msgpack:"has_items"`
}

type QuestData struct {
	QuestId  uint32 `json:"quest_id" msgpack:"quest_id"`
	Progress uint32 `json:"progress" msgpack:"progress"`
	Status   uint32 `json:"status" msgpack:"status"` // 0=进行中 1=可领取 2=已完成
}

type PlayerSettings struct {
	MusicOn    bool   `json:"music_on" msgpack:"music_on"`
	SfxOn      bool   `json:"sfx_on" msgpack:"sfx_on"`
	Language   uint32 `json:"language" msgpack:"language"`
	Quality    uint32 `json:"quality" msgpack:"quality"`
	AutoBattle bool   `json:"auto_battle" msgpack:"auto_battle"`
}

// 战斗帧同步 — 每帧广播给所有客户端
type BattleFrame struct {
	BattleId   uint64        `json:"battle_id" msgpack:"battle_id"`
	Frame      uint32        `json:"frame" msgpack:"frame"`
	Timestamp  int64         `json:"timestamp" msgpack:"timestamp"`
	Inputs     []PlayerInput `json:"inputs" msgpack:"inputs"`
	RandomSeed uint32        `json:"random_seed" msgpack:"random_seed"`
}

type PlayerInput struct {
	PlayerId uint32 `json:"player_id" msgpack:"player_id"`
	HeroId   uint32 `json:"hero_id" msgpack:"hero_id"`
	Action   uint32 `json:"action" msgpack:"action"` // 0=idle 1=move 2=attack 3=skill 4=item
	SkillId  uint32 `json:"skill_id" msgpack:"skill_id"`
	TargetId uint32 `json:"target_id" msgpack:"target_id"`
	X        int32  `json:"x" msgpack:"x"` // 定点数 * 1000
	Y        int32  `json:"y" msgpack:"y"`
	Dir      uint32 `json:"dir" msgpack:"dir"` // 朝向 0-359
}

// 公会战状态 — 大规模多人同步
type GuildWarState struct {
	WarId      uint64       `json:"war_id" msgpack:"war_id"`
	Phase      uint32       `json:"phase" msgpack:"phase"` // 0=准备 1=战斗 2=结算
	RemainSec  uint32       `json:"remain_sec" msgpack:"remain_sec"`
	Guilds     []GuildInfo  `json:"guilds" msgpack:"guilds"`
	Towers     []TowerState `json:"towers" msgpack:"towers"`
	TopPlayers []PlayerRank `json:"top_players" msgpack:"top_players"`
}

type GuildInfo struct {
	GuildId  uint32 `json:"guild_id" msgpack:"guild_id"`
	Name     string `json:"name" msgpack:"name"`
	Score    uint32 `json:"score" msgpack:"score"`
	Members  uint32 `json:"members" msgpack:"members"`
	AliveCnt uint32 `json:"alive_cnt" msgpack:"alive_cnt"`
}

type TowerState struct {
	TowerId uint32 `json:"tower_id" msgpack:"tower_id"`
	Owner   uint32 `json:"owner" msgpack:"owner"` // guild_id
	Hp      uint32 `json:"hp" msgpack:"hp"`
	MaxHp   uint32 `json:"max_hp" msgpack:"max_hp"`
	Level   uint32 `json:"level" msgpack:"level"`
}

type PlayerRank struct {
	PlayerId uint32 `json:"player_id" msgpack:"player_id"`
	Name     string `json:"name" msgpack:"name"`
	Score    uint32 `json:"score" msgpack:"score"`
	KillCnt  uint32 `json:"kill_cnt" msgpack:"kill_cnt"`
}

// 排行榜 — 大量重复结构
type Leaderboard struct {
	Type      uint32      `json:"type" msgpack:"type"`
	Season    uint32      `json:"season" msgpack:"season"`
	UpdatedAt int64       `json:"updated_at" msgpack:"updated_at"`
	Entries   []RankEntry `json:"entries" msgpack:"entries"`
	MyRank    uint32      `json:"my_rank" msgpack:"my_rank"`
	MyScore   uint64      `json:"my_score" msgpack:"my_score"`
}

type RankEntry struct {
	Rank     uint32 `json:"rank" msgpack:"rank"`
	PlayerId uint64 `json:"player_id" msgpack:"player_id"`
	Name     string `json:"name" msgpack:"name"`
	Level    uint32 `json:"level" msgpack:"level"`
	Score    uint64 `json:"score" msgpack:"score"`
	Avatar   uint32 `json:"avatar" msgpack:"avatar"`
	Guild    string `json:"guild" msgpack:"guild"`
}

// ==================== ByteMsg 编码 ====================

func encodeLoginPushBmsg(lp LoginPush) []byte {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	// player
	enc.WriteFieldHeader(1, 2)
	var pBuf bytes.Buffer
	p := NewEncoder(&pBuf)
	p.WriteFieldHeader(1, 0)
	p.WriteVarint(lp.Player.Uid)
	p.WriteFieldHeader(2, 2)
	p.WriteString(lp.Player.Name)
	p.WriteFieldHeader(3, 0)
	p.WriteVarint(uint64(lp.Player.Level))
	p.WriteFieldHeader(4, 0)
	p.WriteVarint(uint64(lp.Player.VipLevel))
	p.WriteFieldHeader(5, 0)
	p.WriteVarint(uint64(lp.Player.Diamond))
	p.WriteFieldHeader(6, 0)
	p.WriteVarint(lp.Player.Gold)
	p.WriteFieldHeader(7, 0)
	p.WriteVarint(uint64(lp.Player.Energy))
	p.WriteFieldHeader(8, 0)
	p.WriteVarint(uint64(lp.Player.Avatar))
	p.WriteFieldHeader(9, 0)
	p.WriteVarint(uint64(lp.Player.GuildId))
	p.WriteFieldHeader(10, 2)
	p.WriteString(lp.Player.GuildName)
	enc.WriteBytes(pBuf.Bytes())
	// heroes
	enc.WriteFieldHeader(2, 2)
	var hBuf bytes.Buffer
	h := NewEncoder(&hBuf)
	h.WriteVarint(uint64(len(lp.Heroes)))
	for _, hero := range lp.Heroes {
		h.WriteFieldHeader(1, 0)
		h.WriteVarint(uint64(hero.HeroId))
		h.WriteFieldHeader(2, 0)
		h.WriteVarint(uint64(hero.Level))
		h.WriteFieldHeader(3, 0)
		h.WriteVarint(uint64(hero.Star))
		h.WriteFieldHeader(4, 0)
		h.WriteVarint(uint64(hero.Grade))
		h.WriteFieldHeader(5, 0)
		h.WriteVarint(hero.Exp)
		h.WriteFieldHeader(6, 2)
		var sBuf bytes.Buffer
		s := NewEncoder(&sBuf)
		s.WriteVarint(uint64(len(hero.Skills)))
		for _, sk := range hero.Skills {
			s.WriteFieldHeader(1, 0)
			s.WriteVarint(uint64(sk.SkillId))
			s.WriteFieldHeader(2, 0)
			s.WriteVarint(uint64(sk.Level))
		}
		h.WriteBytes(sBuf.Bytes())
		h.WriteFieldHeader(7, 2)
		var rBuf bytes.Buffer
		r := NewEncoder(&rBuf)
		r.WriteVarint(uint64(len(hero.Runes)))
		for k, v := range hero.Runes {
			r.WriteVarint(uint64(k))
			r.WriteVarint(uint64(v))
		}
		h.WriteBytes(rBuf.Bytes())
		h.WriteFieldHeader(8, 0)
		h.WriteVarint(uint64(hero.SkinId))
		h.WriteFieldHeader(9, 0)
		h.WriteVarint(uint64(hero.AwakeLv))
		h.WriteFieldHeader(10, 0)
		h.WriteVarint(uint64(hero.FavorLv))
		h.WriteFieldHeader(11, 0)
		h.WriteVarint(uint64(hero.FavorExp))
	}
	enc.WriteBytes(hBuf.Bytes())
	// items
	enc.WriteFieldHeader(3, 2)
	var iBuf bytes.Buffer
	it := NewEncoder(&iBuf)
	it.WriteVarint(uint64(len(lp.Items)))
	for _, item := range lp.Items {
		it.WriteFieldHeader(1, 0)
		it.WriteVarint(uint64(item.ItemId))
		it.WriteFieldHeader(2, 0)
		it.WriteVarint(uint64(item.Count))
		it.WriteFieldHeader(3, 0)
		it.WriteVarint(uint64(item.Level))
		if item.Locked {
			it.WriteFieldHeader(4, 0)
			it.WriteVarint(1)
		}
		it.WriteFieldHeader(5, 0)
		it.WriteVarint(uint64(item.Exp))
	}
	enc.WriteBytes(iBuf.Bytes())
	// mail
	enc.WriteFieldHeader(4, 2)
	var mBuf bytes.Buffer
	m := NewEncoder(&mBuf)
	m.WriteVarint(uint64(len(lp.Mail)))
	for _, mail := range lp.Mail {
		m.WriteFieldHeader(1, 0)
		m.WriteVarint(mail.MailId)
		m.WriteFieldHeader(2, 0)
		m.WriteVarint(uint64(mail.Type))
		m.WriteFieldHeader(3, 2)
		m.WriteString(mail.Title)
		if mail.Read {
			m.WriteFieldHeader(4, 0)
			m.WriteVarint(1)
		}
		if mail.Taken {
			m.WriteFieldHeader(5, 0)
			m.WriteVarint(1)
		}
		m.WriteFieldHeader(6, 0)
		m.WriteVarint(uint64(mail.SendTime))
		if mail.HasItems {
			m.WriteFieldHeader(7, 0)
			m.WriteVarint(1)
		}
	}
	enc.WriteBytes(mBuf.Bytes())
	// quests
	enc.WriteFieldHeader(5, 2)
	var qBuf bytes.Buffer
	q := NewEncoder(&qBuf)
	q.WriteVarint(uint64(len(lp.Quests)))
	for _, quest := range lp.Quests {
		q.WriteFieldHeader(1, 0)
		q.WriteVarint(uint64(quest.QuestId))
		q.WriteFieldHeader(2, 0)
		q.WriteVarint(uint64(quest.Progress))
		q.WriteFieldHeader(3, 0)
		q.WriteVarint(uint64(quest.Status))
	}
	enc.WriteBytes(qBuf.Bytes())
	// settings
	enc.WriteFieldHeader(6, 2)
	var stBuf bytes.Buffer
	st := NewEncoder(&stBuf)
	if lp.Settings.MusicOn {
		st.WriteFieldHeader(1, 0)
		st.WriteVarint(1)
	}
	if lp.Settings.SfxOn {
		st.WriteFieldHeader(2, 0)
		st.WriteVarint(1)
	}
	st.WriteFieldHeader(3, 0)
	st.WriteVarint(uint64(lp.Settings.Language))
	st.WriteFieldHeader(4, 0)
	st.WriteVarint(uint64(lp.Settings.Quality))
	if lp.Settings.AutoBattle {
		st.WriteFieldHeader(5, 0)
		st.WriteVarint(1)
	}
	enc.WriteBytes(stBuf.Bytes())
	// server_time
	enc.WriteFieldHeader(7, 0)
	enc.WriteVarint(uint64(lp.ServerTime))
	return buf.Bytes()
}

func encodeBattleFrameBmsg(f BattleFrame) []byte {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.WriteFieldHeader(1, 0)
	enc.WriteVarint(f.BattleId)
	enc.WriteFieldHeader(2, 0)
	enc.WriteVarint(uint64(f.Frame))
	enc.WriteFieldHeader(3, 0)
	enc.WriteVarint(uint64(f.Timestamp))
	enc.WriteFieldHeader(4, 2)
	var inBuf bytes.Buffer
	in := NewEncoder(&inBuf)
	in.WriteVarint(uint64(len(f.Inputs)))
	for _, input := range f.Inputs {
		in.WriteFieldHeader(1, 0)
		in.WriteVarint(uint64(input.PlayerId))
		in.WriteFieldHeader(2, 0)
		in.WriteVarint(uint64(input.HeroId))
		in.WriteFieldHeader(3, 0)
		in.WriteVarint(uint64(input.Action))
		in.WriteFieldHeader(4, 0)
		in.WriteVarint(uint64(input.SkillId))
		in.WriteFieldHeader(5, 0)
		in.WriteVarint(uint64(input.TargetId))
		in.WriteFieldHeader(6, 0)
		in.WriteZigzag(int64(input.X))
		in.WriteFieldHeader(7, 0)
		in.WriteZigzag(int64(input.Y))
		in.WriteFieldHeader(8, 0)
		in.WriteVarint(uint64(input.Dir))
	}
	enc.WriteBytes(inBuf.Bytes())
	enc.WriteFieldHeader(5, 0)
	enc.WriteVarint(uint64(f.RandomSeed))
	return buf.Bytes()
}

func encodeLeaderboardBmsg(lb Leaderboard) []byte {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.WriteFieldHeader(1, 0)
	enc.WriteVarint(uint64(lb.Type))
	enc.WriteFieldHeader(2, 0)
	enc.WriteVarint(uint64(lb.Season))
	enc.WriteFieldHeader(3, 0)
	enc.WriteVarint(uint64(lb.UpdatedAt))
	enc.WriteFieldHeader(4, 2)
	var eBuf bytes.Buffer
	e := NewEncoder(&eBuf)
	e.WriteVarint(uint64(len(lb.Entries)))
	for _, entry := range lb.Entries {
		e.WriteFieldHeader(1, 0)
		e.WriteVarint(uint64(entry.Rank))
		e.WriteFieldHeader(2, 0)
		e.WriteVarint(entry.PlayerId)
		e.WriteFieldHeader(3, 2)
		e.WriteString(entry.Name)
		e.WriteFieldHeader(4, 0)
		e.WriteVarint(uint64(entry.Level))
		e.WriteFieldHeader(5, 0)
		e.WriteVarint(entry.Score)
		e.WriteFieldHeader(6, 0)
		e.WriteVarint(uint64(entry.Avatar))
		e.WriteFieldHeader(7, 2)
		e.WriteString(entry.Guild)
	}
	enc.WriteBytes(eBuf.Bytes())
	enc.WriteFieldHeader(5, 0)
	enc.WriteVarint(uint64(lb.MyRank))
	enc.WriteFieldHeader(6, 0)
	enc.WriteVarint(lb.MyScore)
	return buf.Bytes()
}

// ==================== 测试数据生成 ====================

func makeLoginPush() LoginPush {
	heroes := make([]HeroData, 30)
	for i := range heroes {
		heroes[i] = HeroData{
			HeroId: uint32(10001 + i), Level: uint32(40 + i%20),
			Star: uint32(3 + i%4), Grade: uint32(i % 5),
			Exp: uint64(100000 * (i + 1)),
			Skills: []SkillData{
				{uint32(20001 + i*3), uint32(5 + i%6)},
				{uint32(20002 + i*3), uint32(3 + i%4)},
				{uint32(20003 + i*3), uint32(1 + i%3)},
			},
			Runes:   map[uint32]uint32{1: uint32(30001 + i), 2: uint32(30002 + i), 3: uint32(30003 + i)},
			SkinId:  uint32(40001 + i%5),
			AwakeLv: uint32(i % 3), FavorLv: uint32(1 + i%10), FavorExp: uint32(i * 100),
		}
	}
	items := make([]ItemData, 80)
	for i := range items {
		items[i] = ItemData{
			ItemId: uint32(60001 + i), Count: uint32(1 + i%99),
			Level: uint32(i % 15), Locked: i%11 == 0, Exp: uint32(i * 50),
		}
	}
	mails := make([]MailData, 15)
	for i := range mails {
		mails[i] = MailData{
			MailId: uint64(900001 + i), Type: uint32(i%4 + 1),
			Title: fmt.Sprintf("系统邮件 #%d — 恭喜获得奖励", i+1),
			Read:  i < 8, Taken: i < 5, SendTime: 1718304000 - int64(i*3600),
			HasItems: i%3 == 0,
		}
	}
	quests := make([]QuestData, 20)
	for i := range quests {
		quests[i] = QuestData{
			QuestId: uint32(70001 + i), Progress: uint32(i * 5), Status: uint32(i % 3),
		}
	}
	return LoginPush{
		Player: PlayerInfo{
			Uid: 100000001, Name: "绝影·暗夜猎手", Level: 65, VipLevel: 8,
			Diamond: 12580, Gold: 9876543, Energy: 85, Avatar: 1001,
			GuildId: 5001, GuildName: "苍穹之巅",
		},
		Heroes: heroes, Items: items, Mail: mails, Quests: quests,
		ServerTime: 1718304000,
	}
}

func makeBattleFrame() BattleFrame {
	inputs := make([]PlayerInput, 10)
	for i := range inputs {
		inputs[i] = PlayerInput{
			PlayerId: uint32(10001 + i), HeroId: uint32(20001 + i),
			Action: uint32(i % 5), SkillId: uint32(30001 + i%3),
			TargetId: uint32(10001 + (i+5)%10),
			X:        int32(1000 + i*50), Y: int32(2000 - i*30), Dir: uint32(i * 36),
		}
	}
	return BattleFrame{
		BattleId: 888001, Frame: 1234, Timestamp: 1718304000,
		Inputs: inputs, RandomSeed: 42,
	}
}

func makeLeaderboard() Leaderboard {
	entries := make([]RankEntry, 100)
	guilds := []string{"苍穹之巅", "星辰大海", "龙之领域", "暗影军团", "光明圣殿", "风暴骑士", "永恒之火", ""}
	for i := range entries {
		entries[i] = RankEntry{
			Rank: uint32(i + 1), PlayerId: uint64(100000 + i),
			Name: fmt.Sprintf("玩家%04d", i+1), Level: uint32(50 + i%15),
			Score: uint64(1000000 - i*8000), Avatar: uint32(1001 + i%10),
			Guild: guilds[i%len(guilds)],
		}
	}
	return Leaderboard{
		Type: 1, Season: 5, UpdatedAt: 1718304000,
		Entries: entries, MyRank: 42, MyScore: 654321,
	}
}

// ==================== 优化版编码 (buffer pool) ====================

func encodeLoginPushOptimized(lp LoginPush) []byte {
	buf := GetBuffer()
	defer PutBuffer(buf)
	enc := NewEncoder(buf)

	pBuf := GetBuffer()
	p := NewEncoder(pBuf)
	p.WriteFieldHeader(1, 0)
	p.WriteVarint(lp.Player.Uid)
	p.WriteFieldHeader(2, 2)
	p.WriteString(lp.Player.Name)
	p.WriteFieldHeader(3, 0)
	p.WriteVarint(uint64(lp.Player.Level))
	p.WriteFieldHeader(4, 0)
	p.WriteVarint(uint64(lp.Player.VipLevel))
	p.WriteFieldHeader(5, 0)
	p.WriteVarint(uint64(lp.Player.Diamond))
	p.WriteFieldHeader(6, 0)
	p.WriteVarint(lp.Player.Gold)
	p.WriteFieldHeader(7, 0)
	p.WriteVarint(uint64(lp.Player.Energy))
	p.WriteFieldHeader(8, 0)
	p.WriteVarint(uint64(lp.Player.Avatar))
	p.WriteFieldHeader(9, 0)
	p.WriteVarint(uint64(lp.Player.GuildId))
	p.WriteFieldHeader(10, 2)
	p.WriteString(lp.Player.GuildName)
	enc.WriteFieldHeader(1, 2)
	enc.WriteBytes(pBuf.Bytes())
	PutBuffer(pBuf)

	hBuf := GetBuffer()
	h := NewEncoder(hBuf)
	h.WriteVarint(uint64(len(lp.Heroes)))
	for _, hero := range lp.Heroes {
		h.WriteFieldHeader(1, 0)
		h.WriteVarint(uint64(hero.HeroId))
		h.WriteFieldHeader(2, 0)
		h.WriteVarint(uint64(hero.Level))
		h.WriteFieldHeader(3, 0)
		h.WriteVarint(uint64(hero.Star))
		h.WriteFieldHeader(4, 0)
		h.WriteVarint(uint64(hero.Grade))
		h.WriteFieldHeader(5, 0)
		h.WriteVarint(hero.Exp)
		sBuf := GetBuffer()
		s := NewEncoder(sBuf)
		s.WriteVarint(uint64(len(hero.Skills)))
		for _, sk := range hero.Skills {
			s.WriteFieldHeader(1, 0)
			s.WriteVarint(uint64(sk.SkillId))
			s.WriteFieldHeader(2, 0)
			s.WriteVarint(uint64(sk.Level))
		}
		h.WriteFieldHeader(6, 2)
		h.WriteBytes(sBuf.Bytes())
		PutBuffer(sBuf)
		rBuf := GetBuffer()
		r := NewEncoder(rBuf)
		r.WriteVarint(uint64(len(hero.Runes)))
		for k, v := range hero.Runes {
			r.WriteVarint(uint64(k))
			r.WriteVarint(uint64(v))
		}
		h.WriteFieldHeader(7, 2)
		h.WriteBytes(rBuf.Bytes())
		PutBuffer(rBuf)
		h.WriteFieldHeader(8, 0)
		h.WriteVarint(uint64(hero.SkinId))
		h.WriteFieldHeader(9, 0)
		h.WriteVarint(uint64(hero.AwakeLv))
		h.WriteFieldHeader(10, 0)
		h.WriteVarint(uint64(hero.FavorLv))
		h.WriteFieldHeader(11, 0)
		h.WriteVarint(uint64(hero.FavorExp))
	}
	enc.WriteFieldHeader(2, 2)
	enc.WriteBytes(hBuf.Bytes())
	PutBuffer(hBuf)

	iBuf := GetBuffer()
	it := NewEncoder(iBuf)
	it.WriteVarint(uint64(len(lp.Items)))
	for _, item := range lp.Items {
		it.WriteFieldHeader(1, 0)
		it.WriteVarint(uint64(item.ItemId))
		it.WriteFieldHeader(2, 0)
		it.WriteVarint(uint64(item.Count))
		it.WriteFieldHeader(3, 0)
		it.WriteVarint(uint64(item.Level))
		if item.Locked {
			it.WriteFieldHeader(4, 0)
			it.WriteVarint(1)
		}
		it.WriteFieldHeader(5, 0)
		it.WriteVarint(uint64(item.Exp))
	}
	enc.WriteFieldHeader(3, 2)
	enc.WriteBytes(iBuf.Bytes())
	PutBuffer(iBuf)

	mBuf := GetBuffer()
	m := NewEncoder(mBuf)
	m.WriteVarint(uint64(len(lp.Mail)))
	for _, mail := range lp.Mail {
		m.WriteFieldHeader(1, 0)
		m.WriteVarint(mail.MailId)
		m.WriteFieldHeader(2, 0)
		m.WriteVarint(uint64(mail.Type))
		m.WriteFieldHeader(3, 2)
		m.WriteString(mail.Title)
		if mail.Read {
			m.WriteFieldHeader(4, 0)
			m.WriteVarint(1)
		}
		if mail.Taken {
			m.WriteFieldHeader(5, 0)
			m.WriteVarint(1)
		}
		m.WriteFieldHeader(6, 0)
		m.WriteVarint(uint64(mail.SendTime))
		if mail.HasItems {
			m.WriteFieldHeader(7, 0)
			m.WriteVarint(1)
		}
	}
	enc.WriteFieldHeader(4, 2)
	enc.WriteBytes(mBuf.Bytes())
	PutBuffer(mBuf)

	qBuf := GetBuffer()
	q := NewEncoder(qBuf)
	q.WriteVarint(uint64(len(lp.Quests)))
	for _, quest := range lp.Quests {
		q.WriteFieldHeader(1, 0)
		q.WriteVarint(uint64(quest.QuestId))
		q.WriteFieldHeader(2, 0)
		q.WriteVarint(uint64(quest.Progress))
		q.WriteFieldHeader(3, 0)
		q.WriteVarint(uint64(quest.Status))
	}
	enc.WriteFieldHeader(5, 2)
	enc.WriteBytes(qBuf.Bytes())
	PutBuffer(qBuf)

	enc.WriteFieldHeader(7, 0)
	enc.WriteVarint(uint64(lp.ServerTime))
	return append([]byte(nil), buf.Bytes()...)
}

func encodeBattleFrameOptimized(f BattleFrame) []byte {
	buf := GetBuffer()
	defer PutBuffer(buf)
	enc := NewEncoder(buf)
	enc.WriteFieldHeader(1, 0)
	enc.WriteVarint(f.BattleId)
	enc.WriteFieldHeader(2, 0)
	enc.WriteVarint(uint64(f.Frame))
	enc.WriteFieldHeader(3, 0)
	enc.WriteVarint(uint64(f.Timestamp))
	inBuf := GetBuffer()
	in := NewEncoder(inBuf)
	in.WriteVarint(uint64(len(f.Inputs)))
	for _, input := range f.Inputs {
		in.WriteFieldHeader(1, 0)
		in.WriteVarint(uint64(input.PlayerId))
		in.WriteFieldHeader(2, 0)
		in.WriteVarint(uint64(input.HeroId))
		in.WriteFieldHeader(3, 0)
		in.WriteVarint(uint64(input.Action))
		in.WriteFieldHeader(4, 0)
		in.WriteVarint(uint64(input.SkillId))
		in.WriteFieldHeader(5, 0)
		in.WriteVarint(uint64(input.TargetId))
		in.WriteFieldHeader(6, 0)
		in.WriteZigzag(int64(input.X))
		in.WriteFieldHeader(7, 0)
		in.WriteZigzag(int64(input.Y))
		in.WriteFieldHeader(8, 0)
		in.WriteVarint(uint64(input.Dir))
	}
	enc.WriteFieldHeader(4, 2)
	enc.WriteBytes(inBuf.Bytes())
	PutBuffer(inBuf)
	enc.WriteFieldHeader(5, 0)
	enc.WriteVarint(uint64(f.RandomSeed))
	return append([]byte(nil), buf.Bytes()...)
}

func encodeLeaderboardOptimized(lb Leaderboard) []byte {
	buf := GetBuffer()
	defer PutBuffer(buf)
	enc := NewEncoder(buf)
	enc.WriteFieldHeader(1, 0)
	enc.WriteVarint(uint64(lb.Type))
	enc.WriteFieldHeader(2, 0)
	enc.WriteVarint(uint64(lb.Season))
	enc.WriteFieldHeader(3, 0)
	enc.WriteVarint(uint64(lb.UpdatedAt))
	eBuf := GetBuffer()
	e := NewEncoder(eBuf)
	e.WriteVarint(uint64(len(lb.Entries)))
	for _, entry := range lb.Entries {
		e.WriteFieldHeader(1, 0)
		e.WriteVarint(uint64(entry.Rank))
		e.WriteFieldHeader(2, 0)
		e.WriteVarint(entry.PlayerId)
		e.WriteFieldHeader(3, 2)
		e.WriteString(entry.Name)
		e.WriteFieldHeader(4, 0)
		e.WriteVarint(uint64(entry.Level))
		e.WriteFieldHeader(5, 0)
		e.WriteVarint(entry.Score)
		e.WriteFieldHeader(6, 0)
		e.WriteVarint(uint64(entry.Avatar))
		e.WriteFieldHeader(7, 2)
		e.WriteString(entry.Guild)
	}
	enc.WriteFieldHeader(4, 2)
	enc.WriteBytes(eBuf.Bytes())
	PutBuffer(eBuf)
	enc.WriteFieldHeader(5, 0)
	enc.WriteVarint(uint64(lb.MyRank))
	enc.WriteFieldHeader(6, 0)
	enc.WriteVarint(lb.MyScore)
	return append([]byte(nil), buf.Bytes()...)
}

// ==================== 场景测试 ====================

func TestGame_LoginPush(t *testing.T) {
	lp := makeLoginPush()
	bmsg := encodeLoginPushBmsg(lp)
	jsonData, _ := json.Marshal(lp)
	mpData, _ := msgpack.Marshal(lp)

	t.Logf("╔══════════════════════════════════════════════╗")
	t.Logf("║  场景: 登录推送 (Login Push)                 ║")
	t.Logf("║  30 英雄 · 80 背包 · 15 邮件 · 20 任务      ║")
	t.Logf("╠══════════════════════════════════════════════╣")
	t.Logf("║  ByteMsg:     %5d bytes                    ║", len(bmsg))
	t.Logf("║  JSON:        %5d bytes                    ║", len(jsonData))
	t.Logf("║  MessagePack: %5d bytes                    ║", len(mpData))
	t.Logf("╠══════════════════════════════════════════════╣")
	t.Logf("║  ByteMsg / JSON    = %5.1f%%                ║", float64(len(bmsg))/float64(len(jsonData))*100)
	t.Logf("║  ByteMsg / MsgPack = %5.1f%%                ║", float64(len(bmsg))/float64(len(mpData))*100)
	t.Logf("║  节省 vs JSON      = %5.1f%%                ║", (1-float64(len(bmsg))/float64(len(jsonData)))*100)
	t.Logf("╚══════════════════════════════════════════════╝")
}

func TestGame_BattleFrame(t *testing.T) {
	frame := makeBattleFrame()
	bmsg := encodeBattleFrameBmsg(frame)
	jsonData, _ := json.Marshal(frame)
	mpData, _ := msgpack.Marshal(frame)

	t.Logf("╔══════════════════════════════════════════════╗")
	t.Logf("║  场景: 战斗帧同步 (Battle Frame Sync)        ║")
	t.Logf("║  10 玩家输入 · 每秒 30 帧                    ║")
	t.Logf("╠══════════════════════════════════════════════╣")
	t.Logf("║  ByteMsg:     %4d bytes                     ║", len(bmsg))
	t.Logf("║  JSON:        %4d bytes                     ║", len(jsonData))
	t.Logf("║  MessagePack: %4d bytes                     ║", len(mpData))
	t.Logf("╠══════════════════════════════════════════════╣")
	t.Logf("║  ByteMsg / JSON    = %5.1f%%                ║", float64(len(bmsg))/float64(len(jsonData))*100)
	t.Logf("║  ByteMsg / MsgPack = %5.1f%%                ║", float64(len(bmsg))/float64(len(mpData))*100)
	t.Logf("╚══════════════════════════════════════════════╝")

	// 30fps 带宽计算
	bpsBmsg := float64(len(bmsg)) * 30 * 8 / 1000
	bpsJson := float64(len(jsonData)) * 30 * 8 / 1000
	t.Logf("")
	t.Logf("  30fps 带宽需求:")
	t.Logf("    ByteMsg: %.1f kbps", bpsBmsg)
	t.Logf("    JSON:    %.1f kbps", bpsJson)
	t.Logf("    节省:    %.1f kbps", bpsJson-bpsBmsg)
}

func TestGame_Leaderboard(t *testing.T) {
	lb := makeLeaderboard()
	bmsg := encodeLeaderboardBmsg(lb)
	jsonData, _ := json.Marshal(lb)
	mpData, _ := msgpack.Marshal(lb)

	t.Logf("╔══════════════════════════════════════════════╗")
	t.Logf("║  场景: 排行榜 (Leaderboard — 100 players)    ║")
	t.Logf("╠══════════════════════════════════════════════╣")
	t.Logf("║  ByteMsg:     %5d bytes                    ║", len(bmsg))
	t.Logf("║  JSON:        %5d bytes                    ║", len(jsonData))
	t.Logf("║  MessagePack: %5d bytes                    ║", len(mpData))
	t.Logf("╠══════════════════════════════════════════════╣")
	t.Logf("║  ByteMsg / JSON    = %5.1f%%                ║", float64(len(bmsg))/float64(len(jsonData))*100)
	t.Logf("║  ByteMsg / MsgPack = %5.1f%%                ║", float64(len(bmsg))/float64(len(mpData))*100)
	t.Logf("╚══════════════════════════════════════════════╝")
}

// ==================== 编码性能 Benchmark ====================

func BenchmarkGame_LoginPush_ByteMsg(b *testing.B) {
	lp := makeLoginPush()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeLoginPushBmsg(lp)
	}
}

func BenchmarkGame_LoginPush_JSON(b *testing.B) {
	lp := makeLoginPush()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(lp)
	}
}

func BenchmarkGame_LoginPush_Msgpack(b *testing.B) {
	lp := makeLoginPush()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msgpack.Marshal(lp)
	}
}

func BenchmarkGame_BattleFrame_ByteMsg(b *testing.B) {
	f := makeBattleFrame()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeBattleFrameBmsg(f)
	}
}

func BenchmarkGame_BattleFrame_JSON(b *testing.B) {
	f := makeBattleFrame()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(f)
	}
}

func BenchmarkGame_BattleFrame_Msgpack(b *testing.B) {
	f := makeBattleFrame()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msgpack.Marshal(f)
	}
}

func BenchmarkGame_Leaderboard_ByteMsg(b *testing.B) {
	lb := makeLeaderboard()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeLeaderboardBmsg(lb)
	}
}

func BenchmarkGame_Leaderboard_JSON(b *testing.B) {
	lb := makeLeaderboard()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(lb)
	}
}

func BenchmarkGame_Leaderboard_Msgpack(b *testing.B) {
	lb := makeLeaderboard()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msgpack.Marshal(lb)
	}
}

func BenchmarkGame_LoginPush_Optimized(b *testing.B) {
	lp := makeLoginPush()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeLoginPushOptimized(lp)
	}
}

func BenchmarkGame_BattleFrame_Optimized(b *testing.B) {
	f := makeBattleFrame()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeBattleFrameOptimized(f)
	}
}

func BenchmarkGame_Leaderboard_Optimized(b *testing.B) {
	lb := makeLeaderboard()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeLeaderboardOptimized(lb)
	}
}
