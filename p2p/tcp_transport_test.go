package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
ops := TCPTransportOpts{
		ListenAddress: ":8080",
		HandshakeFunc:     NOPHandshakeFunc,
		Decoder:       DefaultDecoder{},
}
	tr := NewTCPTransport(ops)
	assert.Equal(t, tr.ListenAddress, ":8080")
	assert.Nil(t, tr.ListenAndAccept())
}
