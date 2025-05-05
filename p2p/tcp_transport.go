package p2p

import (
	"fmt"
	"net"
	"sync"
)

// TCPPeer represents a remote node over a TCP connection.
type TCPPeer struct {
	// underlying connection to the peer
	conn net.Conn
	// if we dial and retrieve a conn -> outbound == true
	// if we accepts and retrieve a conn -> outbound == false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}
type TCPTransportOpts struct {
	ListenAddress string 
	HandshakeFunc HandshakeFunc
	Decoder 			Decoder

}

type TCPTransport struct {
	TCPTransportOpts 	
	listener      		net.Listener
	mu    						sync.RWMutex
	peers 						map[net.Addr]Peer
}


func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
	}
}


func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.listenAddress)
	if err != nil {
		return err
	}
	go t.startAcceptLoop()
	return nil
}

func (t *TCPTransport) startAcceptLoop() error {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP Error accepting connection: %s\n", err)
		}
		fmt.Printf(" new incoming connection %+v\n", conn)
		go t.handleConn(conn)
	}
}
type Temp struct {}
func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)

	if err := t.shakeHands(peer); err != nil {
		fmt.Printf("Error during handshake: %s\n", err)
		conn.Close()
		return
	}

	msg := &Temp{}
	for {
		if err := t.decoder.Decode(conn,msg) ; err != nil {
			fmt.Printf("TCP Error decoding message: %s\n", err)
			continue
		}
	}
}
