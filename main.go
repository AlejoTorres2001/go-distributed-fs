package main

import "github.com/AlejoTorres2001/go-distributed-fs/p2p"

func main() {
	tcpopts := p2p.TCPTransportOpts{
		ListenAddress: ":8080",
		HandshakeFunc:     p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tr := p2p.NewTCPTransport(tcpopts)

	if err := tr.ListenAndAccept(); err != nil {
		panic(err)
	}

	select {}
}
