package client

import "testing"

func TestNewAPIClient(t *testing.T) {
	c := NewAPIClient("http://127.0.0.1:3000", "test-key")
	if c.baseURL != "http://127.0.0.1:3000" {
		t.Errorf("baseURL mismatch")
	}
	if c.accessKey != "test-key" {
		t.Errorf("accessKey mismatch")
	}
}

func TestBuildURL(t *testing.T) {
	c := NewAPIClient("http://127.0.0.1:3000", "key")
	u := c.buildURL("/api/messages/channels")
	expected := "http://127.0.0.1:3000/api/messages/channels"
	if u != expected {
		t.Errorf("got %s, want %s", u, expected)
	}
}

func TestChannelTypeValues(t *testing.T) {
	if ChannelTypeMessages != "messages" {
		t.Error("ChannelTypeMessages mismatch")
	}
	if len(AllChannelTypes) != 4 {
		t.Errorf("expected 4 channel types, got %d", len(AllChannelTypes))
	}
}
