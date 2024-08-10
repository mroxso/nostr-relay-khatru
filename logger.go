package main

import (
	"context"
	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
	"log"
)

func EventLogger(logtype string) func(ctx context.Context, event *nostr.Event) error {
	return func(ctx context.Context, event *nostr.Event) error {
		fromIP := khatru.GetIP(ctx)
		log.Printf("EVENT %s : %s kind %d with event ID %s", logtype, fromIP, event.Kind, event.ID)
		return nil
	}
}
