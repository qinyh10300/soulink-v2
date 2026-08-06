package chat

import (
	"io"
	"log/slog"
	"testing"
)

func TestHubDeliversToEveryConnectionForUser(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := NewHub(logger)
	first := &Client{userID: 7, send: make(chan WSEnvelope, 2), logger: logger}
	second := &Client{userID: 7, send: make(chan WSEnvelope, 2), logger: logger}
	hub.Register(first)
	hub.Register(second)

	want := WSEnvelope{Type: "message"}
	if !hub.SendToUser(7, want) {
		t.Fatal("SendToUser() reported user offline")
	}
	for name, client := range map[string]*Client{"first": first, "second": second} {
		select {
		case got := <-client.send:
			if got.Type != want.Type {
				t.Fatalf("%s connection received %q, want %q", name, got.Type, want.Type)
			}
		default:
			t.Fatalf("%s connection did not receive the event", name)
		}
	}
}

func TestHubUnregisterKeepsOtherConnectionsOnline(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := NewHub(logger)
	first := &Client{userID: 9, send: make(chan WSEnvelope, 1), logger: logger}
	second := &Client{userID: 9, send: make(chan WSEnvelope, 1), logger: logger}
	hub.Register(first)
	hub.Register(second)
	hub.Unregister(first)

	if !hub.SendToUser(9, WSEnvelope{Type: "group_message"}) {
		t.Fatal("SendToUser() reported remaining connection offline")
	}
	select {
	case <-first.send:
		t.Fatal("unregistered connection received an event")
	default:
	}
	select {
	case got := <-second.send:
		if got.Type != "group_message" {
			t.Fatalf("remaining connection received %q", got.Type)
		}
	default:
		t.Fatal("remaining connection did not receive the event")
	}
}
