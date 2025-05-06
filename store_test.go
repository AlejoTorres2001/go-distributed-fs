package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestStore(t *testing.T) {
	s := newStore()
	defer teardownStore(t, s)
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("foo_%d", i)
		
		data := []byte("some random bytes")
		if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}
		
		if ok := s.Has(key); !ok {
			t.Errorf("expected to have key  %v", key)
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
		
		if err:= s.Delete(key); err != nil {
			t.Error(err)
		}
		if ok := s.Has(key); ok {
			t.Errorf("expected to NOT have key  %v", key)
		}
	}
	
}

func newStore() *Store {
	opts := StoreOpts{
		Root:              "dfs",
		PathTransfromFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}
func teardownStore(t *testing.T, s *Store) {
	if err := s.ClearAll(); err != nil {
		t.Errorf("error clearing store %v", err)
	}
}