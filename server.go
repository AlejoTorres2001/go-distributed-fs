package main

import (
	"fmt"
	"log"

	"github.com/AlejoTorres2001/go-distributed-fs/p2p"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransfromFunc PathTransfromFunc
	Transport				p2p.Transport
}
type FileServer struct {
	FileServerOpts
	store *Store
	quitch chan struct{}
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
		quitch: 			 make(chan struct{}),
	}
}
func (s *FileServer) Stop()  {
	close(s.quitch)
}
func (s *FileServer) loop() error {
	defer func() {
		log.Println("FileServer stopped due to user quit action")
		s.Transport.Close()
	}()
	for  {
		select {
		case msg := <-s.Transport.Consume():
			fmt.Println(msg)

		case <- s.quitch:
			return nil
		}
		
	}
}
func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	s.loop()
	return nil
}
