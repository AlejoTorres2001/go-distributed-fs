package p2p

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
    // Configurar canal para recibir mensajes
    messageReceived := make(chan RPC, 1)
    
    // Definir OnPeer similar al main.go
    onPeerCalled := false
    onPeerFunc := func(p Peer) error {
        onPeerCalled = true
        fmt.Println("doing something outside tcp transport")
        return nil
    }
    
    // Configurar transporte con puerto dinámico para evitar conflictos en CI/CD
    // En uso manual puedes usar ":8080" como en main.go
    opts := TCPTransportOpts{
        ListenAddress: ":0", // Puerto dinámico para pruebas automatizadas
        HandshakeFunc: NOPHandshakeFunc,
        Decoder:       DefaultDecoder{},
        OnPeer:        onPeerFunc,
    }
    
    tr := NewTCPTransport(opts)
    
    // Iniciar goroutine para consumir mensajes (como en main.go)
    go func() {
        rpc := <-tr.Consume()
        messageReceived <- rpc
    }()
    
    // Iniciar el transporte
    err := tr.ListenAndAccept()
    assert.Nil(t, err)
    defer tr.listener.Close()
    
    // Obtener el puerto asignado para conectarnos
    addr := tr.listener.Addr().String()
    t.Logf("Test server running on %s", addr)
    
    // Simular un cliente telnet
    telnetConn, err := net.Dial("tcp", addr)
    assert.Nil(t, err)
    defer telnetConn.Close()
    
    // Verificar que OnPeer fue llamado
    time.Sleep(100 * time.Millisecond)
    assert.True(t, onPeerCalled, "OnPeer debería haber sido llamado")
    
    // Enviar un mensaje como lo harías con telnet
    testMessage := "Hola desde telnet\n"
    _, err = telnetConn.Write([]byte(testMessage))
    assert.Nil(t, err)
    
    // Esperar y verificar que se recibió el mensaje
    select {
    case rpc := <-messageReceived:
        t.Logf("Mensaje recibido: %s", rpc.Payload)
        assert.Contains(t, string(rpc.Payload), "Hola desde telnet")
    case <-time.After(1 * time.Second):
        t.Fatal("Timeout esperando el mensaje")
    }
}
