package binary

import (
	"bytes"
	"encoding/binary"
	"io"
)

const ByteMsgBufferPoolLimit = 10000

// BlockKind identifies an optional length-delimited block payload layout.
// Blocks keep the normal field header outside the payload, so old row layouts
// can coexist with optimized list/column layouts in the same protocol.
type BlockKind uint8

const (
	BlockPackedVarint BlockKind = 1
	BlockPackedZigzag BlockKind = 2
	BlockDeltaVarint  BlockKind = 3
	BlockBoolBitset   BlockKind = 4
	BlockStringList   BlockKind = 5
	BlockColumnList   BlockKind = 6
)

const (
	WireTypeVarint          = 0
	WireTypeFixed64         = 1
	WireTypeLengthDelimited = 2
	WireTypeFixed32         = 5
)

var bufferPool = make([]*bytes.Buffer, 0, ByteMsgBufferPoolLimit)

// GetBuffer gets a buffer from the pool
func GetBuffer() *bytes.Buffer {
	if n := len(bufferPool); n > 0 {
		buf := bufferPool[n-1]
		bufferPool[n-1] = nil
		bufferPool = bufferPool[:n-1]
		buf.Reset()
		return buf
	}
	return new(bytes.Buffer)
}

// PutBuffer returns a buffer to the pool
func PutBuffer(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	if buf.Cap() > 64*1024 {
		return
	}
	buf.Reset()
	if len(bufferPool) < ByteMsgBufferPoolLimit {
		bufferPool = append(bufferPool, buf)
	}
}

// Encoder writes binary data
type Encoder struct {
	w   io.Writer
	buf [binary.MaxVarintLen64]byte
}

// BufferEncoder writes binary data to a bytes.Buffer without io.Writer interface overhead.
type BufferEncoder struct {
	w   *bytes.Buffer
	buf [binary.MaxVarintLen64]byte
}

// AppendEncoder appends binary data to a caller-owned byte slice.
type AppendEncoder struct {
	dst []byte
	buf [binary.MaxVarintLen64]byte
}

// NewEncoder creates a new encoder
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// NewEncoderValue creates an encoder value that can stay on the caller stack.
func NewEncoderValue(w io.Writer) Encoder {
	return Encoder{w: w}
}

// NewBufferEncoderValue creates a buffer encoder value that can stay on the caller stack.
func NewBufferEncoderValue(w *bytes.Buffer) BufferEncoder {
	return BufferEncoder{w: w}
}

// NewAppendEncoderValue creates an append encoder over a caller-owned byte slice.
func NewAppendEncoderValue(dst []byte) AppendEncoder {
	return AppendEncoder{dst: dst}
}

// WriteVarint writes a variable-length integer
func (e *Encoder) WriteVarint(value uint64) error {
	n := binary.PutUvarint(e.buf[:], value)
	_, err := e.w.Write(e.buf[:n])
	return err
}

// WriteZigzag writes a zigzag-encoded integer
func (e *Encoder) WriteZigzag(value int64) error {
	return e.WriteVarint(ZigzagEncode(value))
}

// WriteString writes a length-prefixed string
func (e *Encoder) WriteString(s string) error {
	if err := e.WriteVarint(uint64(len(s))); err != nil {
		return err
	}
	_, err := io.WriteString(e.w, s)
	return err
}

// WriteBytes writes length-prefixed bytes
func (e *Encoder) WriteBytes(data []byte) error {
	if err := e.WriteVarint(uint64(len(data))); err != nil {
		return err
	}
	_, err := e.w.Write(data)
	return err
}

// WriteFixed32 writes a fixed-width 32-bit little-endian value.
func (e *Encoder) WriteFixed32(value uint32) error {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], value)
	_, err := e.w.Write(buf[:])
	return err
}

// WriteFixed64 writes a fixed-width 64-bit little-endian value.
func (e *Encoder) WriteFixed64(value uint64) error {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], value)
	_, err := e.w.Write(buf[:])
	return err
}

// WriteFieldHeader writes a field header (tag + wire type)
func (e *Encoder) WriteFieldHeader(tag int, wireType int) error {
	return e.WriteVarint(uint64(tag<<3 | wireType))
}

// WriteVarint writes a variable-length integer.
func (e *BufferEncoder) WriteVarint(value uint64) error {
	n := binary.PutUvarint(e.buf[:], value)
	_, err := e.w.Write(e.buf[:n])
	return err
}

// WriteZigzag writes a zigzag-encoded integer.
func (e *BufferEncoder) WriteZigzag(value int64) error {
	return e.WriteVarint(ZigzagEncode(value))
}

// WriteString writes a length-prefixed string.
func (e *BufferEncoder) WriteString(s string) error {
	if err := e.WriteVarint(uint64(len(s))); err != nil {
		return err
	}
	_, err := e.w.WriteString(s)
	return err
}

// WriteBytes writes length-prefixed bytes.
func (e *BufferEncoder) WriteBytes(data []byte) error {
	if err := e.WriteVarint(uint64(len(data))); err != nil {
		return err
	}
	_, err := e.w.Write(data)
	return err
}

// WriteFixed32 writes a fixed-width 32-bit little-endian value.
func (e *BufferEncoder) WriteFixed32(value uint32) error {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], value)
	_, err := e.w.Write(buf[:])
	return err
}

// WriteFixed64 writes a fixed-width 64-bit little-endian value.
func (e *BufferEncoder) WriteFixed64(value uint64) error {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], value)
	_, err := e.w.Write(buf[:])
	return err
}

// WriteFieldHeader writes a field header (tag + wire type).
func (e *BufferEncoder) WriteFieldHeader(tag int, wireType int) error {
	return e.WriteVarint(uint64(tag<<3 | wireType))
}

// Reset changes the destination slice.
func (e *AppendEncoder) Reset(dst []byte) {
	e.dst = dst
}

// Bytes returns encoded bytes.
func (e *AppendEncoder) Bytes() []byte {
	return e.dst
}

// WriteVarint writes a variable-length integer.
func (e *AppendEncoder) WriteVarint(value uint64) error {
	n := binary.PutUvarint(e.buf[:], value)
	e.dst = append(e.dst, e.buf[:n]...)
	return nil
}

// WriteZigzag writes a zigzag-encoded integer.
func (e *AppendEncoder) WriteZigzag(value int64) error {
	return e.WriteVarint(ZigzagEncode(value))
}

// WriteString writes a length-prefixed string.
func (e *AppendEncoder) WriteString(s string) error {
	if err := e.WriteVarint(uint64(len(s))); err != nil {
		return err
	}
	e.dst = append(e.dst, s...)
	return nil
}

// WriteBytes writes length-prefixed bytes.
func (e *AppendEncoder) WriteBytes(data []byte) error {
	if err := e.WriteVarint(uint64(len(data))); err != nil {
		return err
	}
	e.dst = append(e.dst, data...)
	return nil
}

// WriteFixed32 writes a fixed-width 32-bit little-endian value.
func (e *AppendEncoder) WriteFixed32(value uint32) error {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], value)
	e.dst = append(e.dst, buf[:]...)
	return nil
}

// WriteFixed64 writes a fixed-width 64-bit little-endian value.
func (e *AppendEncoder) WriteFixed64(value uint64) error {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], value)
	e.dst = append(e.dst, buf[:]...)
	return nil
}

// WriteFieldHeader writes a field header (tag + wire type).
func (e *AppendEncoder) WriteFieldHeader(tag int, wireType int) error {
	return e.WriteVarint(uint64(tag<<3 | wireType))
}

// AppendVarint appends a variable-length integer to dst.
func AppendVarint(dst []byte, value uint64) []byte {
	for value >= 0x80 {
		dst = append(dst, byte(value)|0x80)
		value >>= 7
	}
	return append(dst, byte(value))
}

// AppendZigzag appends a zigzag-encoded integer to dst.
func AppendZigzag(dst []byte, value int64) []byte {
	return AppendVarint(dst, ZigzagEncode(value))
}

// AppendString appends a length-prefixed string to dst.
func AppendString(dst []byte, value string) []byte {
	dst = AppendVarint(dst, uint64(len(value)))
	return append(dst, value...)
}

// AppendBytes appends length-prefixed bytes to dst.
func AppendBytes(dst []byte, value []byte) []byte {
	dst = AppendVarint(dst, uint64(len(value)))
	return append(dst, value...)
}

// AppendFieldHeader appends a field header (tag + wire type) to dst.
func AppendFieldHeader(dst []byte, tag int, wireType int) []byte {
	return AppendVarint(dst, uint64(tag<<3|wireType))
}

// VarintLen returns the encoded byte length of a uint64 varint.
func VarintLen(value uint64) int {
	n := 1
	for value >= 0x80 {
		n++
		value >>= 7
	}
	return n
}

// AppendBlockHeader appends the tag, block kind, and payload length for an
// optimized length-delimited block. The payload itself follows immediately.
func AppendBlockHeader(dst []byte, tag int, kind BlockKind, payloadLen int) []byte {
	dst = AppendFieldHeader(dst, tag, 2)
	dst = AppendVarint(dst, uint64(payloadLen+1))
	return append(dst, byte(kind))
}

// AppendPackedVarints appends count-prefixed unsigned varints.
func AppendPackedVarints(dst []byte, values []uint64) []byte {
	dst = AppendVarint(dst, uint64(len(values)))
	for _, value := range values {
		dst = AppendVarint(dst, value)
	}
	return dst
}

// AppendPackedZigzags appends count-prefixed zigzag varints.
func AppendPackedZigzags(dst []byte, values []int64) []byte {
	dst = AppendVarint(dst, uint64(len(values)))
	for _, value := range values {
		dst = AppendZigzag(dst, value)
	}
	return dst
}

// AppendDeltaVarints appends count-prefixed unsigned varints as base + deltas.
// It is useful for ranks, ids, frames, timestamps, and other monotonic lists.
func AppendDeltaVarints(dst []byte, values []uint64) []byte {
	dst = AppendVarint(dst, uint64(len(values)))
	if len(values) == 0 {
		return dst
	}
	prev := values[0]
	dst = AppendVarint(dst, prev)
	for _, value := range values[1:] {
		dst = AppendZigzag(dst, int64(value)-int64(prev))
		prev = value
	}
	return dst
}

// AppendBoolBitset appends count-prefixed bool values packed into bits.
func AppendBoolBitset(dst []byte, values []bool) []byte {
	dst = AppendVarint(dst, uint64(len(values)))
	var current byte
	for i, value := range values {
		if value {
			current |= 1 << uint(i&7)
		}
		if i&7 == 7 {
			dst = append(dst, current)
			current = 0
		}
	}
	if len(values)&7 != 0 {
		dst = append(dst, current)
	}
	return dst
}

// AppendStringList appends count-prefixed strings.
func AppendStringList(dst []byte, values []string) []byte {
	dst = AppendVarint(dst, uint64(len(values)))
	for _, value := range values {
		dst = AppendString(dst, value)
	}
	return dst
}

// ZigzagEncode converts int64 to uint64 using zigzag encoding
func ZigzagEncode(value int64) uint64 {
	return uint64((value << 1) ^ (value >> 63))
}

// ZigzagDecode converts uint64 to int64 using zigzag decoding
func ZigzagDecode(value uint64) int64 {
	return int64((value >> 1) ^ -(value & 1))
}
