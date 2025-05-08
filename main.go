package main

import (
	"log"
	"time"
	"github.com/AlejoTorres2001/go-distributed-fs/p2p"
)

func main() {
	tcpopts := p2p.TCPTransportOpts{
		ListenAddress: ":8080",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpopts)
	fileServerOpts := FileServerOpts{
		StorageRoot:       "8080_network",
		PathTransfromFunc: DefaultPathTransformFunc,
		Transport:         tcpTransport,
	}
	fileServer := NewFileServer(fileServerOpts)

	go func() {
		time.Sleep(5 * time.Second)
		fileServer.Stop()
	}()
	if err := fileServer.Start(); err != nil {
		log.Fatal(err)
	}
}
