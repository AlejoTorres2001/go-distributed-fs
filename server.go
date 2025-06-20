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
type MessageStoreFile struct {
	Key string
	Size int64

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
	tee := io.TeeReader(r, buf)
	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}

	msg := Message{
			Payload:MessageStoreFile{
				Key: key,
				Size: size,
			},
	}
	msgBuf := new(bytes.Buffer)
	if err := gob.NewEncoder(msgBuf).Encode(msg); err != nil {
		return err
	}
	for _, peer := range s.peers {
		if err := peer.Send(msgBuf.Bytes()); err != nil {
			return err
		}
	}
	time.Sleep(time.Second * 3)
	for _, peer := range s.peers {
		n,err := io.Copy(peer, buf)
		if err != nil {
			return err
		}
		println("received and written bytes to disk",n)
	}
	return nil
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
func (s *FileServer) loop() {
	defer func() {
		log.Println("FileServer stopped due to user quit action")
		s.Transport.Close()
	}()
	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("Error decoding message:", err)
				return 
			}
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println("Error handling message:", err)
			}

		case <-s.quitch:
			return
		}

	}
}
func (s *FileServer) handleMessage(from string ,msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	}
	return nil
}
func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer ,ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer not found in the peer list: %s", from)
	}
	_, err := s.store.Write(msg.Key, io.LimitReader(peer,msg.Size))
	if err != nil {
		return err
	}
	peer.(*p2p.TCPPeer).Wg.Done()
	return nil
}
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
func init() {
	gob.Register(Message{})
	gob.Register(MessageStoreFile{})
}