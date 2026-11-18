package core

import (
	"testing"
)

func TestGenerateID_NoPanicAndUnique(t *testing.T) {
	seen := make(map[string]struct{}, 200)
	for i := 0; i < 200; i++ {
		id := generateID()
		if id == "" {
			t.Fatal("empty id")
		}
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate id: %s", id)
		}
		seen[id] = struct{}{}
	}
}
