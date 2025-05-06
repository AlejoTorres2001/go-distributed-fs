package main

import (
	"bytes"
	"fmt"
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
	pathname := CASPathTransformFunc(key)
	expectedPathName := "e24fc/4bc21/80e4d/f3696/836ab/8ccb8/ebe1b/7bf9b"
	if pathname != expectedPathName {
		t.Errorf("have %s want %s",pathname,expectedPathName)
	}
	fmt.Println(pathname)
}