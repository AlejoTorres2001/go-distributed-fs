package main

import "github.com/AlejoTorres2001/go-distributed-fs/p2p"


func main() {
	tr := p2p.NewTCPTransport(":8080")

	if err := tr.ListenAndAccept(); err != nil {
		panic(err)
	}

	select {}
}
