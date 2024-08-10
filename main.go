package main

import (
	"context"
	"fmt"
	"time"
	"net/http"
	"slices"
	"strings"

	"github.com/fiatjaf/khatru"
	"github.com/fiatjaf/khatru/policies"

	"github.com/fiatjaf/eventstore/sqlite3"
	"github.com/nbd-wtf/go-nostr"
)

var allowedKinds = []uint16{0,1,3,5,6,7}
var page string = "Hello.\n\nUse me in your nostr client.\n\nThanks."
var whiteListedIPs = []string{"127.0.0.1", "::1"}

func servepage(w http.ResponseWriter) {
	fmt.Fprint(w, page)
}

func main() {
	relay := khatru.NewRelay()

	relay.Info.Name = "Nostr Relay"
	relay.Info.PubKey = "0000000000000000000000000000000000000000000000000000000000000000"
	relay.Info.Description = "Nostr relay written in khatru."
	relay.Info.Icon = "https://example.com/logo.png"
	relay.Info.Contact = "mailto:nobody@example.com"

	db := sqlite3.SQLite3Backend{DatabaseURL: "/home/nostr/nostr.db"}
	if err := db.Init(); err != nil {
		panic(err)
	}

	relay.StoreEvent = append(relay.StoreEvent, db.SaveEvent)
	relay.QueryEvents = append(relay.QueryEvents, db.QueryEvents)
	relay.CountEvents = append(relay.CountEvents, db.CountEvents)
	relay.DeleteEvent = append(relay.DeleteEvent, db.DeleteEvent)

	relay.RejectEvent = append(relay.RejectEvent,
		func(ctx context.Context, _ *nostr.Event) (reject bool, msg string) {
			fromIP := khatru.GetIP(ctx)

			if slices.Contains(whiteListedIPs, fromIP) {
				return false, ""
			} else {
				return policies.EventIPRateLimiter(2, time.Minute*3, 5)(ctx, nil)
			}
		},
		policies.PreventLargeTags(70),
		policies.RejectEventsWithBase64Media,
		policies.RestrictToSpecifiedKinds(allowedKinds...),
	)

	relay.RejectFilter = append(relay.RejectFilter,
		func(ctx context.Context, filter nostr.Filter) (reject bool, msg string) {
			fromIP := khatru.GetIP(ctx)

			if slices.Contains(whiteListedIPs, fromIP) {
				return false, ""
			} else {
				return policies.FilterIPRateLimiter(20, time.Minute, 100)(ctx, filter)
			}
		},
	)

	relay.RejectConnection = append(relay.RejectConnection,
		func (r *http.Request) bool {
			fromIP := khatru.GetIPFromRequest(r)
			if slices.Contains(whiteListedIPs, fromIP) {
				return false
			} else {
				return policies.ConnectionRateLimiter(1, time.Minute*5, 3)(r)
			}
		},
	)

	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.Header.Get("Upgrade") == "websocket" {
			relay.HandleWebsocket(w, r)
		} else {
			accept := r.Header.Get("Accept")
			if strings.Contains(accept, "application/nostr+json") {
				relay.HandleNIP11(w, r)
			} else {
				servepage(w)
			}
		}
	})

	fmt.Println("Blowing up on localhost:7777")
	http.ListenAndServe("localhost:7777", nil)
}
