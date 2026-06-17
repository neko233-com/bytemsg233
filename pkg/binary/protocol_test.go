package binary

import (
	"errors"
	"testing"
)

func TestProtocolHelloRoundTripAndCompatibility(t *testing.T) {
	local := ProtocolHello{Version: 7, MinCompatible: 6}
	data := AppendProtocolHello(nil, local)
	data = AppendFieldHeader(data, 99, WireTypeLengthDelimited)
	data = AppendString(data, "future")

	remote, err := ReadProtocolHello(data)
	if err != nil {
		t.Fatalf("ReadProtocolHello failed: %v", err)
	}
	if remote != local {
		t.Fatalf("hello = %#v, want %#v", remote, local)
	}
	if err := CheckProtocolHello(local, remote); err != nil {
		t.Fatalf("CheckProtocolHello failed: %v", err)
	}

	remote.MinCompatible = 8
	if !errors.Is(CheckProtocolHello(local, remote), ErrProtocolVersionMismatch) {
		t.Fatalf("expected version mismatch")
	}
}
