package webfetch

import (
	"context"
	"testing"
)

func TestFetchRejectsBadURL(t *testing.T) {
	_, err := Fetch(context.Background(), nil, "u1", "not-a-url", EngineBasic)
	if err == nil {
		t.Fatal("expected error for invalid url")
	}
}

func TestFetchRejectsUnknownEngine(t *testing.T) {
	_, err := Fetch(context.Background(), nil, "u1", "https://example.com", "selenium")
	if err == nil {
		t.Fatal("expected error for unknown engine")
	}
}
