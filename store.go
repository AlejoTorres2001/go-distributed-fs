package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strings"
)
func CASPathTransformFunc(key string) string {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLen := len(hashStr) / blockSize

	paths := make([]string,sliceLen)
	for i := 0; i < sliceLen; i++ {
		from,to := i * blockSize, (i * blockSize) + blockSize
		paths[i] = hashStr[from:to]

	}
	return strings.Join(paths,"/")
}
type PathTransfromFunc func(string) string
var DefaultPathTransformFunc = func(key string) string {return key}

type StoreOpts struct {
	PathTransfromFunc PathTransfromFunc
}
type Store struct {
	StoreOpts
}
func NewStore(opts StoreOpts) *Store {
	return &Store{
		StoreOpts: opts,
	}
}
func (s *Store) writeStream(key string, r io.Reader) error {
	pathName :=  s.PathTransfromFunc(key)
	if err := os.MkdirAll(pathName,os.ModePerm); err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	io.Copy(buf,r)

	fileNameBytes := md5.Sum(buf.Bytes())
	filename := hex.EncodeToString(fileNameBytes[:])

	pathAndFilename := pathName + "/" + filename
	f,err := os.Create(pathAndFilename)
	if err != nil {
		return err
	}
	n,err := io.Copy(f,buf)
	if err != nil {
		return err
	}
	log.Printf("written (%d) bytes to disk: %s",n,pathAndFilename)
	return nil
}