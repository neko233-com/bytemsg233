package binary

import (
	"bytes"
	"encoding/binary"
	"io"
)

const ByteMsgBufferPoolLimit = 10000

var bufferPool = make(chan *bytes.Buffer, ByteMsgBufferPoolLimit)

// GetBuffer gets a buffer from the pool
func GetBuffer() *bytes.Buffer {
	select {
	case buf := <-bufferPool:
		buf.Reset()
		return buf
	default:
		return new(bytes.Buffer)
	}
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
	select {
	case bufferPool <- buf:
	default:
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

// ZigzagEncode converts int64 to uint64 using zigzag encoding
func ZigzagEncode(value int64) uint64 {
	return uint64((value << 1) ^ (value >> 63))
}

// ZigzagDecode converts uint64 to int64 using zigzag decoding
func ZigzagDecode(value uint64) int64 {
	return int64((value >> 1) ^ -(value & 1))
}
