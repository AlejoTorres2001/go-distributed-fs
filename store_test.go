package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransfromFunc: CASPathTransformFunc,
	}

	s := NewStore(opts)
	key := "catsbestpictures"
	data := []byte("some bytes")
	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}
	r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}
	b, _ := io.ReadAll(r)
	fmt.Println(string(b))
	if string(b) != string(data) {
		t.Errorf("have %s want %s", string(b), string(data))
	}
	s.Delete(key)

}
func TestPathTransformFunc(t *testing.T) {
	key := "catsbestpictures"
	pathKey := CASPathTransformFunc(key)
	expectedOriginalKey := "e24fc4bc2180e4df3696836ab8ccb8ebe1b7bf9b"
	expectedPathName := "e24fc/4bc21/80e4d/f3696/836ab/8ccb8/ebe1b/7bf9b"
	if pathKey.PathName != expectedPathName {
		t.Errorf("have %s want %s", pathKey.PathName, expectedPathName)
	}
	if pathKey.FileName != expectedOriginalKey {
		t.Errorf("have %s want %s", pathKey.PathName, expectedOriginalKey)
	}
}

func TestDelete(t *testing.T) {
	opts := StoreOpts{
		PathTransfromFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	key := "catsbestpictures"
	data := []byte("some bytes")
	if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}
	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
	r, err := s.Read(key)
	if err == nil {
		t.Errorf("expected error but got %v", r)
	}
}
