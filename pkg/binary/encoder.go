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

// NewEncoder creates a new encoder
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// NewEncoderValue creates an encoder value that can stay on the caller stack.
func NewEncoderValue(w io.Writer) Encoder {
	return Encoder{w: w}
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

// ZigzagEncode converts int64 to uint64 using zigzag encoding
func ZigzagEncode(value int64) uint64 {
	return uint64((value << 1) ^ (value >> 63))
}

// ZigzagDecode converts uint64 to int64 using zigzag decoding
func ZigzagDecode(value uint64) int64 {
	return int64((value >> 1) ^ -(value & 1))
}
