nostr relay (with khatru)
written by: yonle

it's using github.com/fiatjaf/khatru for relay implementation.

- default database is sqlite3, storing at /home/nostr/nostr.db.
- modify some values in main.go before running.

after modifying, build by running the following command:
  $ go build .

it will make an executable under filename "relay" in current dir.
to start the relay, simply run:
  $ ./relay
