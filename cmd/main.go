package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	netfs "netfs/internal"

	"github.com/dgraph-io/badger/v4"
)

func main() {
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	db, err := badger.Open(badger.DefaultOptions("./tmp/badger"))
	if err != nil {
		log.Fatal(err)
	}

	server := netfs.Server{DB: db, Context: context.Background(), Config: netfs.Config{BufferSize: 100 * 1024 * 1024, TaskCount: 10}}
	server.Listen()

	defer db.Close()
}
