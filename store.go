package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const defaultRootFolderName = "dfs"

type StoreOpts struct {
	//Root is the folder name of the root, containing all the files of the system
	Root              string
	PathTransfromFunc PathTransfromFunc
}
type Store struct {
	StoreOpts
}
type PathKey struct {
	PathName string
	FileName string
}
type PathTransfromFunc func(string) PathKey

func NewStore(opts StoreOpts) *Store {
	if opts.PathTransfromFunc == nil {
		opts.PathTransfromFunc = DefaultPathTransformFunc
	}
	if opts.Root == "" {
		opts.Root = defaultRootFolderName
	}
	return &Store{
		StoreOpts: opts,
	}
}
func (s *Store) Write(key string, r io.Reader) error {
	return s.writeStream(key, r)
}
func (s *Store) Read(key string) (io.Reader, error) {
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)
	return buf, err
}
func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathKey := s.PathTransfromFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	return os.Open(fullPathWithRoot)

}
func (s *Store) writeStream(key string, r io.Reader) error {

	pathKey := s.PathTransfromFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.PathName)

	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return err
	}

	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())

	f, err := os.Create(fullPathWithRoot)
	if err != nil {
		return err
	}
	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}
	log.Printf("written (%d) bytes to disk: %s", n, fullPathWithRoot)
	return nil
}
func (s *Store) Delete(key string) error {
	pathKey := s.PathTransfromFunc(key)
	firstPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FirstPathName())
	defer func() {
		log.Printf("deleted [%s] from disk", pathKey.FileName)
	}()
	return os.RemoveAll(firstPathWithRoot)
}
func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.FileName)
}
func (p *PathKey) FirstPathName() string {
	paths := strings.Split(p.PathName, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}
func (s *Store) Has(key string) bool {
	pathKey := s.PathTransfromFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.FullPath())
	_, err := os.Stat(fullPathWithRoot)

	return !errors.Is(err, os.ErrNotExist)

}
func (s *Store) ClearAll() error {
	return os.RemoveAll(s.Root)
}
func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLen := len(hashStr) / blockSize

	paths := make([]string, sliceLen)
	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashStr[from:to]

	}
	return PathKey{
		PathName: strings.Join(paths, "/"),
		FileName: hashStr,
	}
}

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		FileName: key,
	}
}
