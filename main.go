package main

import (
	"fmt"

	"github.com/AlejoTorres2001/go-distributed-fs/p2p"
)

func OnPeer(p2p.Peer) error { 
	fmt.Println("doing something outside tcp transport")
	return nil 
}
func main() {
	tcpopts := p2p.TCPTransportOpts{
		ListenAddress: ":8080",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}
	tr := p2p.NewTCPTransport(tcpopts)
	go func() {
		for {
			rpc := <-tr.Consume()
			fmt.Printf("%+v\n", rpc)
		}
	}()

	if err := tr.ListenAndAccept(); err != nil {
		panic(err)
	}

	select {}
}
