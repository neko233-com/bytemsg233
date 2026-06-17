package binary

import "errors"

var ErrProtocolVersionMismatch = errors.New("bytemsg233: protocol version mismatch")

type ProtocolHello struct {
	Version       uint64
	MinCompatible uint64
}

func AppendProtocolHello(dst []byte, hello ProtocolHello) []byte {
	dst = AppendFieldHeader(dst, 1, WireTypeVarint)
	dst = AppendVarint(dst, hello.Version)
	dst = AppendFieldHeader(dst, 2, WireTypeVarint)
	return AppendVarint(dst, hello.MinCompatible)
}

func ReadProtocolHello(data []byte) (ProtocolHello, error) {
	dec := NewSliceDecoder(data)
	var hello ProtocolHello
	for !dec.EOF() {
		tag, wireType, err := dec.ReadFieldHeader()
		if err != nil {
			return hello, err
		}
		switch tag {
		case 1:
			hello.Version, err = dec.ReadVarint()
		case 2:
			hello.MinCompatible, err = dec.ReadVarint()
		default:
			err = dec.SkipField(wireType)
		}
		if err != nil {
			return hello, err
		}
	}
	return hello, nil
}

func CheckProtocolHello(local ProtocolHello, remote ProtocolHello) error {
	if remote.Version < local.MinCompatible || local.Version < remote.MinCompatible {
		return ErrProtocolVersionMismatch
	}
	return nil
}
