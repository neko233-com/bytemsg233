package binary

import (
	"bytes"
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

func TestBufferPoolLimit(t *testing.T) {
	for {
		select {
		case <-bufferPool:
		default:
			goto drained
		}
	}

drained:
	for i := 0; i < ByteMsgBufferPoolLimit+1; i++ {
		PutBuffer(new(bytes.Buffer))
	}
	if len(bufferPool) != ByteMsgBufferPoolLimit {
		t.Fatalf("buffer pool len = %d, want %d", len(bufferPool), ByteMsgBufferPoolLimit)
	}

	for {
		select {
		case <-bufferPool:
		default:
			return
		}
	}
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
