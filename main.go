package main

import (
	"bytes"
	"log"
	"time"

	"github.com/AlejoTorres2001/go-distributed-fs/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpopts := p2p.TCPTransportOpts{
		ListenAddress: listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpopts)
	fileServerOpts := FileServerOpts{
		StorageRoot:       listenAddr + "_network",
		PathTransfromFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}
	s := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s
}
func main() {
	s1 := makeServer(":8080", "")
	s2 := makeServer(":8081", ":8080")

	go func() {
		log.Fatal(s1.Start())
	}()
	time.Sleep(time.Second * 3)

	go s2.Start()
	time.Sleep(time.Second * 3)

	data := bytes.NewReader([]byte("personal data"))
	s2.StoreData("mydata",data)

	select {}
}
