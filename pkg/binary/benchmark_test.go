package binary

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
)

type UserProfile struct {
	Id       uint32            `json:"id" msgpack:"id" bytemsg:"1"`
	Name     string            `json:"name" msgpack:"name" bytemsg:"2"`
	Email    string            `json:"email" msgpack:"email" bytemsg:"3"`
	Tags     []string          `json:"tags" msgpack:"tags" bytemsg:"4"`
	Metadata map[string]string `json:"metadata" msgpack:"metadata" bytemsg:"5"`
}

func encodeBytemsg(user UserProfile) []byte {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.WriteFieldHeader(1, 0)
	enc.WriteVarint(uint64(user.Id))
	enc.WriteFieldHeader(2, 2)
	enc.WriteString(user.Name)
	enc.WriteFieldHeader(3, 2)
	enc.WriteString(user.Email)
	enc.WriteFieldHeader(4, 2)
	var tagsBuf bytes.Buffer
	tagsEnc := NewEncoder(&tagsBuf)
	tagsEnc.WriteVarint(uint64(len(user.Tags)))
	for _, tag := range user.Tags {
		tagsEnc.WriteString(tag)
	}
	enc.WriteBytes(tagsBuf.Bytes())
	enc.WriteFieldHeader(5, 2)
	var metaBuf bytes.Buffer
	metaEnc := NewEncoder(&metaBuf)
	metaEnc.WriteVarint(uint64(len(user.Metadata)))
	for k, v := range user.Metadata {
		metaEnc.WriteString(k)
		metaEnc.WriteString(v)
	}
	enc.WriteBytes(metaBuf.Bytes())
	return buf.Bytes()
}

func calcProtobufSize(user UserProfile) int {
	size := 0
	size += 1 + varintSize(uint64(user.Id))
	nameBytes := []byte(user.Name)
	size += 1 + varintSize(uint64(len(nameBytes))) + len(nameBytes)
	emailBytes := []byte(user.Email)
	size += 1 + varintSize(uint64(len(emailBytes))) + len(emailBytes)
	for _, tag := range user.Tags {
		tagBytes := []byte(tag)
		size += 1 + varintSize(uint64(len(tagBytes))) + len(tagBytes)
	}
	for k, v := range user.Metadata {
		kBytes := []byte(k)
		vBytes := []byte(v)
		entrySize := 1 + varintSize(uint64(len(kBytes))) + len(kBytes)
		entrySize += 1 + varintSize(uint64(len(vBytes))) + len(vBytes)
		size += 1 + varintSize(uint64(entrySize)) + entrySize
	}
	return size
}

func varintSize(v uint64) int {
	n := 0
	for v >= 0x80 {
		n++
		v >>= 7
	}
	return n + 1
}

func TestSizeComparison(t *testing.T) {
	user := UserProfile{
		Id:    12345,
		Name:  "张三",
		Email: "zhangsan@example.com",
		Tags:  []string{"admin", "user", "developer"},
		Metadata: map[string]string{
			"department": "engineering",
			"level":      "senior",
		},
	}

	bmsgData := encodeBytemsg(user)
	jsonData, _ := json.Marshal(user)
	msgpackData, _ := msgpack.Marshal(user)
	protoSize := calcProtobufSize(user)

	t.Logf("========================================")
	t.Logf("  单条记录体积对比 (Single Record)")
	t.Logf("========================================")
	t.Logf("  ByteMsg:     %3d bytes", len(bmsgData))
	t.Logf("  Protobuf v3: %3d bytes (理论值)", protoSize)
	t.Logf("  JSON:        %3d bytes", len(jsonData))
	t.Logf("  MessagePack: %3d bytes", len(msgpackData))
	t.Logf("========================================")
	t.Logf("  ByteMsg / Protobuf = %.1f%%", float64(len(bmsgData))/float64(protoSize)*100)
	t.Logf("  ByteMsg / JSON     = %.1f%%", float64(len(bmsgData))/float64(len(jsonData))*100)
	t.Logf("  ByteMsg / MsgPack  = %.1f%%", float64(len(bmsgData))/float64(len(msgpackData))*100)
	t.Logf("========================================")

	// Verify roundtrip
	dec := NewDecoder(bytes.NewReader(bmsgData))
	tag, wt, _ := dec.ReadFieldHeader()
	if tag != 1 || wt != 0 {
		t.Errorf("Expected field 1 varint, got tag=%d wt=%d", tag, wt)
	}
	id, _ := dec.ReadVarint()
	if uint32(id) != user.Id {
		t.Errorf("Expected id=%d, got %d", user.Id, id)
	}
	tag, wt, _ = dec.ReadFieldHeader()
	if tag != 2 || wt != 2 {
		t.Errorf("Expected field 2 string, got tag=%d wt=%d", tag, wt)
	}
	name, _ := dec.ReadString()
	if name != user.Name {
		t.Errorf("Expected name=%s, got %s", user.Name, name)
	}
}

func TestSizeComparisonBatch(t *testing.T) {
	users := []UserProfile{
		{Id: 1, Name: "Alice", Email: "alice@test.com", Tags: []string{"admin"}, Metadata: map[string]string{"role": "admin"}},
		{Id: 2, Name: "Bob", Email: "bob@test.com", Tags: []string{"user", "dev"}, Metadata: map[string]string{"role": "user"}},
		{Id: 100, Name: "Charlie", Email: "charlie@longdomain.com", Tags: []string{"moderator", "reviewer", "tester"}, Metadata: map[string]string{"department": "engineering", "level": "senior", "team": "backend"}},
		{Id: 99999, Name: "大卫·张", Email: "david.zhang@verylongdomain.example.com", Tags: []string{"vip", "premium", "enterprise", "beta"}, Metadata: map[string]string{"region": "apac", "tier": "gold", "since": "2020-01-15"}},
		{Id: 42, Name: "Eve", Email: "e@x.co", Tags: nil, Metadata: nil},
	}

	// ByteMsg
	var bmsgBuf bytes.Buffer
	enc := NewEncoder(&bmsgBuf)
	for _, user := range users {
		enc.WriteFieldHeader(1, 0)
		enc.WriteVarint(uint64(user.Id))
		enc.WriteFieldHeader(2, 2)
		enc.WriteString(user.Name)
		enc.WriteFieldHeader(3, 2)
		enc.WriteString(user.Email)
		enc.WriteFieldHeader(4, 2)
		var tagsBuf bytes.Buffer
		tagsEnc := NewEncoder(&tagsBuf)
		tagsEnc.WriteVarint(uint64(len(user.Tags)))
		for _, tag := range user.Tags {
			tagsEnc.WriteString(tag)
		}
		enc.WriteBytes(tagsBuf.Bytes())
		enc.WriteFieldHeader(5, 2)
		var metaBuf bytes.Buffer
		metaEnc := NewEncoder(&metaBuf)
		metaEnc.WriteVarint(uint64(len(user.Metadata)))
		for k, v := range user.Metadata {
			metaEnc.WriteString(k)
			metaEnc.WriteString(v)
		}
		enc.WriteBytes(metaBuf.Bytes())
	}

	jsonData, _ := json.Marshal(users)
	msgpackData, _ := msgpack.Marshal(users)
	protoSize := 0
	for _, u := range users {
		protoSize += calcProtobufSize(u)
	}

	t.Logf("========================================")
	t.Logf("  批量记录体积对比 (5 Records)")
	t.Logf("========================================")
	t.Logf("  ByteMsg:     %3d bytes", bmsgBuf.Len())
	t.Logf("  Protobuf v3: %3d bytes (理论值)", protoSize)
	t.Logf("  JSON:        %3d bytes", len(jsonData))
	t.Logf("  MessagePack: %3d bytes", len(msgpackData))
	t.Logf("========================================")
	t.Logf("  ByteMsg / Protobuf = %.1f%%", float64(bmsgBuf.Len())/float64(protoSize)*100)
	t.Logf("  ByteMsg / JSON     = %.1f%%", float64(bmsgBuf.Len())/float64(len(jsonData))*100)
	t.Logf("  ByteMsg / MsgPack  = %.1f%%", float64(bmsgBuf.Len())/float64(len(msgpackData))*100)
	t.Logf("========================================")
	t.Logf("  节省 vs JSON:      %.1f%%", (1-float64(bmsgBuf.Len())/float64(len(jsonData)))*100)
	t.Logf("  节省 vs MsgPack:   %.1f%%", (1-float64(bmsgBuf.Len())/float64(len(msgpackData)))*100)
}

func TestSizeComparisonIntensive(t *testing.T) {
	// Heavy integer data - where varint shines
	type IntRecord struct {
		A uint32 `json:"a" msgpack:"a"`
		B uint32 `json:"b" msgpack:"b"`
		C uint32 `json:"c" msgpack:"c"`
		D uint64 `json:"d" msgpack:"d"`
	}

	records := []IntRecord{
		{1, 2, 3, 4},
		{100, 200, 300, 400},
		{10000, 20000, 30000, 40000},
		{1000000, 2000000, 3000000, 4000000},
	}

	// ByteMsg
	var bmsgBuf bytes.Buffer
	enc := NewEncoder(&bmsgBuf)
	for _, r := range records {
		enc.WriteFieldHeader(1, 0)
		enc.WriteVarint(uint64(r.A))
		enc.WriteFieldHeader(2, 0)
		enc.WriteVarint(uint64(r.B))
		enc.WriteFieldHeader(3, 0)
		enc.WriteVarint(uint64(r.C))
		enc.WriteFieldHeader(4, 0)
		enc.WriteVarint(r.D)
	}

	jsonData, _ := json.Marshal(records)
	msgpackData, _ := msgpack.Marshal(records)

	// Protobuf: each field is 1 byte tag + varint
	protoSize := 0
	for _, r := range records {
		protoSize += 1 + varintSize(uint64(r.A))
		protoSize += 1 + varintSize(uint64(r.B))
		protoSize += 1 + varintSize(uint64(r.C))
		protoSize += 1 + varintSize(r.D)
	}

	t.Logf("========================================")
	t.Logf("  整数密集型体积对比 (Int-Heavy)")
	t.Logf("========================================")
	t.Logf("  ByteMsg:     %3d bytes", bmsgBuf.Len())
	t.Logf("  Protobuf v3: %3d bytes (理论值)", protoSize)
	t.Logf("  JSON:        %3d bytes", len(jsonData))
	t.Logf("  MessagePack: %3d bytes", len(msgpackData))
	t.Logf("========================================")
	t.Logf("  ByteMsg / Protobuf = %.1f%%", float64(bmsgBuf.Len())/float64(protoSize)*100)
	t.Logf("  ByteMsg / JSON     = %.1f%%", float64(bmsgBuf.Len())/float64(len(jsonData))*100)
	t.Logf("  ByteMsg / MsgPack  = %.1f%%", float64(bmsgBuf.Len())/float64(len(msgpackData))*100)
}

func BenchmarkBytemsgEncode(b *testing.B) {
	user := UserProfile{
		Id: 12345, Name: "张三", Email: "zhangsan@example.com",
		Tags:     []string{"admin", "user", "developer"},
		Metadata: map[string]string{"department": "engineering", "level": "senior"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encodeBytemsg(user)
	}
}

func BenchmarkJSONEncode(b *testing.B) {
	user := UserProfile{
		Id: 12345, Name: "张三", Email: "zhangsan@example.com",
		Tags:     []string{"admin", "user", "developer"},
		Metadata: map[string]string{"department": "engineering", "level": "senior"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(user)
	}
}

func BenchmarkMsgpackEncode(b *testing.B) {
	user := UserProfile{
		Id: 12345, Name: "张三", Email: "zhangsan@example.com",
		Tags:     []string{"admin", "user", "developer"},
		Metadata: map[string]string{"department": "engineering", "level": "senior"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msgpack.Marshal(user)
	}
}
