package binary

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestVarintEncoding(t *testing.T) {
	tests := []struct {
		name  string
		value uint64
	}{
		{"zero", 0},
		{"one", 1},
		{"127", 127},
		{"128", 128},
		{"300", 300},
		{"16384", 16384},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			if err := enc.WriteVarint(tt.value); err != nil {
				t.Fatalf("WriteVarint failed: %v", err)
			}

			dec := NewDecoder(bytes.NewReader(buf.Bytes()))
			result, err := dec.ReadVarint()
			if err != nil {
				t.Fatalf("ReadVarint failed: %v", err)
			}

			if result != tt.value {
				t.Errorf("roundtrip: got %d, want %d", result, tt.value)
			}
		})
	}
}

func TestZigzagEncoding(t *testing.T) {
	tests := []struct {
		name  string
		value int64
	}{
		{"zero", 0},
		{"positive 1", 1},
		{"negative 1", -1},
		{"positive 2", 2},
		{"negative 2", -2},
		{"max int32", 2147483647},
		{"min int32", -2147483648},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := ZigzagEncode(tt.value)
			decoded := ZigzagDecode(encoded)
			if decoded != tt.value {
				t.Errorf("roundtrip: got %d, want %d", decoded, tt.value)
			}
		})
	}
}

func TestStringEncoding(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	testStr := "Hello, 世界!"
	if err := enc.WriteString(testStr); err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}

	dec := NewDecoder(bytes.NewReader(buf.Bytes()))
	result, err := dec.ReadString()
	if err != nil {
		t.Fatalf("ReadString failed: %v", err)
	}

	if result != testStr {
		t.Errorf("ReadString() = %q, want %q", result, testStr)
	}
}

func TestBytesEncoding(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	testData := []byte{0x01, 0x02, 0x03, 0xFF}
	if err := enc.WriteBytes(testData); err != nil {
		t.Fatalf("WriteBytes failed: %v", err)
	}

	dec := NewDecoder(bytes.NewReader(buf.Bytes()))
	result, err := dec.ReadBytes()
	if err != nil {
		t.Fatalf("ReadBytes failed: %v", err)
	}

	if !bytes.Equal(result, testData) {
		t.Errorf("ReadBytes() = %v, want %v", result, testData)
	}
}

func TestFieldHeaderEncoding(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	if err := enc.WriteFieldHeader(1, 0); err != nil {
		t.Fatalf("WriteFieldHeader failed: %v", err)
	}

	dec := NewDecoder(bytes.NewReader(buf.Bytes()))
	tag, wireType, err := dec.ReadFieldHeader()
	if err != nil {
		t.Fatalf("ReadFieldHeader failed: %v", err)
	}

	if tag != 1 {
		t.Errorf("Expected tag 1, got %d", tag)
	}
	if wireType != 0 {
		t.Errorf("Expected wireType 0, got %d", wireType)
	}
}

func TestFixedEncoding(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoderValue(&buf)

	if err := enc.WriteFixed32(0x12345678); err != nil {
		t.Fatalf("WriteFixed32 failed: %v", err)
	}
	if err := enc.WriteFixed64(0x0102030405060708); err != nil {
		t.Fatalf("WriteFixed64 failed: %v", err)
	}

	dec := NewDecoder(bytes.NewReader(buf.Bytes()))
	value32, err := dec.ReadFixed32()
	if err != nil {
		t.Fatalf("ReadFixed32 failed: %v", err)
	}
	if value32 != 0x12345678 {
		t.Fatalf("ReadFixed32() = %#x, want %#x", value32, uint32(0x12345678))
	}

	value64, err := dec.ReadFixed64()
	if err != nil {
		t.Fatalf("ReadFixed64 failed: %v", err)
	}
	if value64 != 0x0102030405060708 {
		t.Fatalf("ReadFixed64() = %#x, want %#x", value64, uint64(0x0102030405060708))
	}
}

func TestBufferEncoder(t *testing.T) {
	var buf bytes.Buffer
	enc := NewBufferEncoderValue(&buf)

	if err := enc.WriteFieldHeader(1, 0); err != nil {
		t.Fatalf("WriteFieldHeader failed: %v", err)
	}
	if err := enc.WriteVarint(42); err != nil {
		t.Fatalf("WriteVarint failed: %v", err)
	}
	if err := enc.WriteFieldHeader(2, 2); err != nil {
		t.Fatalf("WriteFieldHeader failed: %v", err)
	}
	if err := enc.WriteString("debug"); err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}

	dec := NewDecoder(bytes.NewReader(buf.Bytes()))
	tag, wireType, err := dec.ReadFieldHeader()
	if err != nil {
		t.Fatalf("ReadFieldHeader failed: %v", err)
	}
	if tag != 1 || wireType != 0 {
		t.Fatalf("field header = (%d, %d), want (1, 0)", tag, wireType)
	}
	value, err := dec.ReadVarint()
	if err != nil {
		t.Fatalf("ReadVarint failed: %v", err)
	}
	if value != 42 {
		t.Fatalf("value = %d, want 42", value)
	}
	tag, wireType, err = dec.ReadFieldHeader()
	if err != nil {
		t.Fatalf("ReadFieldHeader failed: %v", err)
	}
	if tag != 2 || wireType != 2 {
		t.Fatalf("field header = (%d, %d), want (2, 2)", tag, wireType)
	}
	text, err := dec.ReadString()
	if err != nil {
		t.Fatalf("ReadString failed: %v", err)
	}
	if text != "debug" {
		t.Fatalf("text = %q, want debug", text)
	}
}

func TestAppendEncoder(t *testing.T) {
	enc := NewAppendEncoderValue(make([]byte, 0, 32))
	if err := enc.WriteFieldHeader(1, 0); err != nil {
		t.Fatalf("WriteFieldHeader failed: %v", err)
	}
	if err := enc.WriteVarint(42); err != nil {
		t.Fatalf("WriteVarint failed: %v", err)
	}
	if err := enc.WriteFieldHeader(2, 2); err != nil {
		t.Fatalf("WriteFieldHeader failed: %v", err)
	}
	if err := enc.WriteString("debug"); err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}

	dec := NewDecoder(bytes.NewReader(enc.Bytes()))
	tag, wireType, err := dec.ReadFieldHeader()
	if err != nil {
		t.Fatalf("ReadFieldHeader failed: %v", err)
	}
	if tag != 1 || wireType != 0 {
		t.Fatalf("field header = (%d, %d), want (1, 0)", tag, wireType)
	}
	value, err := dec.ReadVarint()
	if err != nil {
		t.Fatalf("ReadVarint failed: %v", err)
	}
	if value != 42 {
		t.Fatalf("value = %d, want 42", value)
	}
}

func TestAppendAndSliceDecoder(t *testing.T) {
	var data []byte
	data = AppendFieldHeader(data, 1, 0)
	data = AppendVarint(data, 300)
	data = AppendFieldHeader(data, 2, 2)
	data = AppendString(data, "Hello, 世界!")
	data = AppendFieldHeader(data, 3, 0)
	data = AppendZigzag(data, -42)
	data = AppendFieldHeader(data, 4, 2)
	data = AppendBytes(data, []byte{0x01, 0x02, 0xff})

	dec := NewSliceDecoder(data)
	tag, wt, err := dec.ReadFieldHeader()
	if err != nil || tag != 1 || wt != 0 {
		t.Fatalf("field 1 header = (%d, %d, %v), want (1, 0, nil)", tag, wt, err)
	}
	u, err := dec.ReadVarint()
	if err != nil || u != 300 {
		t.Fatalf("field 1 value = (%d, %v), want (300, nil)", u, err)
	}

	tag, wt, err = dec.ReadFieldHeader()
	if err != nil || tag != 2 || wt != 2 {
		t.Fatalf("field 2 header = (%d, %d, %v), want (2, 2, nil)", tag, wt, err)
	}
	s, err := dec.ReadString()
	if err != nil || s != "Hello, 世界!" {
		t.Fatalf("field 2 value = (%q, %v), want string", s, err)
	}

	tag, wt, err = dec.ReadFieldHeader()
	if err != nil || tag != 3 || wt != 0 {
		t.Fatalf("field 3 header = (%d, %d, %v), want (3, 0, nil)", tag, wt, err)
	}
	i, err := dec.ReadZigzag()
	if err != nil || i != -42 {
		t.Fatalf("field 3 value = (%d, %v), want (-42, nil)", i, err)
	}

	tag, wt, err = dec.ReadFieldHeader()
	if err != nil || tag != 4 || wt != 2 {
		t.Fatalf("field 4 header = (%d, %d, %v), want (4, 2, nil)", tag, wt, err)
	}
	b, err := dec.ReadBytes()
	if err != nil || !bytes.Equal(b, []byte{0x01, 0x02, 0xff}) {
		t.Fatalf("field 4 value = (%v, %v), want bytes", b, err)
	}
	if !dec.EOF() {
		t.Fatalf("decoder has %d bytes remaining", dec.Remaining())
	}
}

func TestSliceDecoderReadStringView(t *testing.T) {
	var data []byte
	data = AppendString(data, "zero-copy")

	dec := NewSliceDecoder(data)
	value, err := dec.ReadStringView()
	if err != nil {
		t.Fatalf("ReadStringView failed: %v", err)
	}
	if value != "zero-copy" {
		t.Fatalf("ReadStringView() = %q, want zero-copy", value)
	}
	if !dec.EOF() {
		t.Fatalf("decoder has %d bytes remaining", dec.Remaining())
	}
}

func TestOptimizedBlocksRoundTrip(t *testing.T) {
	unsigned := []uint64{100, 101, 105, 120, 121}
	var buf []byte
	buf = AppendDeltaVarints(buf, unsigned)
	dec := NewSliceDecoder(buf)
	gotUnsigned, err := dec.ReadDeltaVarints(nil)
	if err != nil {
		t.Fatalf("ReadDeltaVarints failed: %v", err)
	}
	if len(gotUnsigned) != len(unsigned) {
		t.Fatalf("delta len = %d, want %d", len(gotUnsigned), len(unsigned))
	}
	for i := range unsigned {
		if gotUnsigned[i] != unsigned[i] {
			t.Fatalf("delta[%d] = %d, want %d", i, gotUnsigned[i], unsigned[i])
		}
	}

	flags := []bool{true, false, true, true, false, false, true, false, true}
	buf = buf[:0]
	buf = AppendBoolBitset(buf, flags)
	dec.Reset(buf)
	gotFlags, err := dec.ReadBoolBitset(nil)
	if err != nil {
		t.Fatalf("ReadBoolBitset failed: %v", err)
	}
	for i := range flags {
		if gotFlags[i] != flags[i] {
			t.Fatalf("flag[%d] = %v, want %v", i, gotFlags[i], flags[i])
		}
	}

	strings := []string{"rank", "inventory", "battle"}
	buf = buf[:0]
	buf = AppendStringList(buf, strings)
	dec.Reset(buf)
	gotStrings, err := dec.ReadStringList(nil)
	if err != nil {
		t.Fatalf("ReadStringList failed: %v", err)
	}
	for i := range strings {
		if gotStrings[i] != strings[i] {
			t.Fatalf("string[%d] = %q, want %q", i, gotStrings[i], strings[i])
		}
	}
}

func TestBufferPoolLimit(t *testing.T) {
	bufferPool = bufferPool[:0]
	for i := 0; i < ByteMsgBufferPoolLimit+1; i++ {
		PutBuffer(new(bytes.Buffer))
	}
	if len(bufferPool) != ByteMsgBufferPoolLimit {
		t.Fatalf("buffer pool len = %d, want %d", len(bufferPool), ByteMsgBufferPoolLimit)
	}

	bufferPool = bufferPool[:0]
}

func TestVarintCompactness(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	enc.WriteVarint(1)
	enc.WriteVarint(300)
	enc.WriteVarint(16384)

	data := buf.Bytes()
	if len(data) != 6 {
		t.Errorf("Expected 6 bytes total (1+2+3), got %d bytes", len(data))
	}
}

func TestSliceDecoderSkipUnknownFields(t *testing.T) {
	var data []byte
	data = AppendFieldHeader(data, 99, WireTypeVarint)
	data = AppendVarint(data, 9001)
	data = AppendFieldHeader(data, 1, WireTypeVarint)
	data = AppendVarint(data, 42)
	data = AppendFieldHeader(data, 100, WireTypeLengthDelimited)
	data = AppendString(data, "future")
	data = AppendFieldHeader(data, 2, WireTypeLengthDelimited)
	data = AppendString(data, "stable")
	data = AppendFieldHeader(data, 101, WireTypeFixed32)
	var fixed32 [4]byte
	binary.LittleEndian.PutUint32(fixed32[:], 0x12345678)
	data = append(data, fixed32[:]...)
	data = AppendFieldHeader(data, 102, WireTypeFixed64)
	var fixed64 [8]byte
	binary.LittleEndian.PutUint64(fixed64[:], 0x0102030405060708)
	data = append(data, fixed64[:]...)

	dec := NewSliceDecoder(data)
	var id uint64
	var name string
	for !dec.EOF() {
		tag, wireType, err := dec.ReadFieldHeader()
		if err != nil {
			t.Fatalf("ReadFieldHeader failed: %v", err)
		}
		switch tag {
		case 1:
			id, err = dec.ReadVarint()
		case 2:
			name, err = dec.ReadStringView()
		default:
			err = dec.SkipField(wireType)
		}
		if err != nil {
			t.Fatalf("read tag %d failed: %v", tag, err)
		}
	}
	if id != 42 || name != "stable" {
		t.Fatalf("decoded old fields = (%d, %q), want (42, stable)", id, name)
	}
}
