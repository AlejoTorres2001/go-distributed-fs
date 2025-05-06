package main

import (
	"bytes"
	"testing"
)


func TestStore(t *testing.T){
	opts:= StoreOpts{

		PathTransfromFunc:CASPathTransformFunc,
	}
	
	s := NewStore(opts)
	data := bytes.NewReader([]byte("some bytes"))
	if err := s.writeStream("mypicture",data); err != nil {
		t.Error(err)
	}
}
func TestPathTransformFunc(t *testing.T){
	key:="catsbestpictures"
	pathKey := CASPathTransformFunc(key)
	expectedOriginalKey := "e24fc4bc2180e4df3696836ab8ccb8ebe1b7bf9b"
	expectedPathName := "e24fc/4bc21/80e4d/f3696/836ab/8ccb8/ebe1b/7bf9b"
	if pathKey.PathName != expectedPathName {
		t.Errorf("have %s want %s",pathKey.PathName,expectedPathName)
	}
	if pathKey.Original != expectedOriginalKey {
		t.Errorf("have %s want %s",pathKey.PathName,expectedOriginalKey)
	}
}