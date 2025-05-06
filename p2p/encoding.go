package p2p

import (
	"encoding/gob"
	"fmt"
	"io"
)


type Decoder interface {
	Decode(io.Reader,*RPC)	error
}

type GOBDecoder struct {}

func (dec GOBDecoder) Decode(r io.Reader, rpc *RPC) error {
	return gob.NewDecoder(r).Decode(rpc)
}

type DefaultDecoder struct {}
func (dec DefaultDecoder) Decode(r io.Reader, rpc *RPC) error {
	if r == nil {
		return fmt.Errorf("reader cannot be nil")
}
	buf := make([]byte, 1028)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	rpc.Payload = buf[:n]


	return nil
}