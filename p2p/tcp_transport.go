package p2p

import (
	"fmt"
	"net"
	"strings"
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
	OnPeer 				func(Peer) error

}

type TCPTransport struct {
	TCPTransportOpts 	
	listener      		net.Listener
	rpcchan 					chan RPC
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
	var err error
	defer func() {
		fmt.Printf("TCP Error dropping peer: %s\n", err)
		conn.Close()

	}()
	peer := NewTCPPeer(conn, true)
	
	if err := t.HandshakeFunc(peer); err != nil {
		return
	}
	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return 
		}
	}
	rpc := RPC{}
	//Read loop
	for {
		err = t.Decoder.Decode(conn,&rpc) 
		if err == net.ErrClosed || (err != nil && strings.Contains(err.Error(), "use of closed network connection")) {
			return 
	}
	if err != nil {
		fmt.Printf("TCP Error decoding the message: %+v\n", err)
		continue // Skip to next iteration without breaking the loop
}
		rpc.From = conn.RemoteAddr()
		t.rpcchan <- rpc
		fmt.Printf("TCP Received message: %+v\n", rpc)
	}
}
