package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/AlejoTorres2001/go-distributed-fs/p2p"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransfromFunc PathTransfromFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}
type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer
	store    *Store
	quitch   chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	if opts.StorageRoot == "" {
		opts.StorageRoot = defaultRootFolderName
	}
	if opts.PathTransfromFunc == nil {
		opts.PathTransfromFunc = DefaultPathTransformFunc
	}
	StoreOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransfromFunc: opts.PathTransfromFunc,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(StoreOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

type Message struct {
	Payload any
}


func (s *FileServer) broadcast(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}
func (s *FileServer) StoreData(key string, r io.Reader) error {
	buf := new(bytes.Buffer)
	msg := Message{
			Payload:[]byte("storagekey"), 
	}
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	for _, peer := range s.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}
	time.Sleep(time.Second * 3)
	payload := []byte("THIS LARGE FILE")
	for _, peer := range s.peers {
		if err := peer.Send(payload); err != nil {
			return err
		}
	}
	return nil
	// buf := new(bytes.Buffer)
	// tee := io.TeeReader(r, buf)
	// if err := s.store.Write(key, tee); err != nil {
	// 	return err
	// }
	// p := &DataMessage{
	// 	Key:  key,
	// 	Data: buf.Bytes(),
	// }
	// return s.broadcast(&Message{
	// 	From:    "todo",
	// 	payload: p,
	// })
}
func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p
	log.Printf("Connected to Remote: %s\n", p.RemoteAddr())
	return nil

}
func (s *FileServer) Stop() {
	close(s.quitch)
}
func (s *FileServer) loop() error {
	defer func() {
		log.Println("FileServer stopped due to user quit action")
		s.Transport.Close()
	}()
	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Received message: %s\n", string(msg.Payload.([]byte)))


			peer, ok := s.peers[rpc.From]
			if !ok {
				fmt.Printf("Peer not found in peers map: %s\n", peer)
				panic("peer not found in peers map")
			}
			buf := make([]byte, 1024)
			_, err := peer.Read(buf); if err != nil {
				panic(err)
			}
			fmt.Printf(" %s\n", string(buf))
			peer.(*p2p.TCPPeer).Wg.Done()

			// if err := s.handleMessage(&m); err != nil {
			// 	log.Println("Error handling message:", err)
			// }

		case <-s.quitch:
			return nil
		}

	}
}
// func (s *FileServer) handleMessage(msg *Message) error {
// 	switch v := msg.payload.(type) {
// 	case *DataMessage:
// 		fmt.Printf("Received data message: %+v\n", v)
// 	}
// 	return nil
// }
func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			log.Println("Attempting to connect to bootstrap node:", addr)
			if err := s.Transport.Dial(addr); err != nil {
				log.Println("Dial error:", err)
			}
		}(addr)
	}
	return nil
}
func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	s.bootstrapNetwork()
	s.loop()
	return nil
}
