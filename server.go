package main

import (
	"fmt"
	"log"
	"sync"

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
		case msg := <-s.Transport.Consume():
			fmt.Println(msg)

		case <-s.quitch:
			return nil
		}

	}
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
