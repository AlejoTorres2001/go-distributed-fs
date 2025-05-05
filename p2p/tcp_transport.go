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
//Close implements the Peer interface, which closes the connection to the peer.
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}
type TCPTransportOpts struct {
	ListenAddress string 
	HandshakeFunc HandshakeFunc
	Decoder 			Decoder

}

type TCPTransport struct {
	TCPTransportOpts 	
	listener      		net.Listener
	rpcchan 					chan RPC

	mu    						sync.RWMutex
	peers 						map[net.Addr]Peer
}


func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcchan: 			 make(chan RPC),
	}
}
// Consume implements the Transport interface, which returns a read only channel for the incoming RPCs received from another peer in the network.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcchan
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddress)
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

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)

	if err := t.HandshakeFunc(peer); err != nil {
		fmt.Printf("TCP Error during handshake: %s\n", err)
		conn.Close()
		return
	}

	rpc := RPC{}
	//Read loop
	for {
		if err := t.Decoder.Decode(conn,&rpc) ; err != nil {
			fmt.Printf("TCP Error decoding message: %s\n", err)
			continue
		}
		rpc.From = conn.RemoteAddr()
		t.rpcchan <- rpc
		fmt.Printf("TCP Received message: %+v\n", rpc)
	}
}
