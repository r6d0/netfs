package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	netfs "netfs/internal"
)

func main() {
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	server, err := netfs.New(netfs.Config{BufferSize: 100 * 1024 * 1024, TaskCount: 10, Server: netfs.ServerConfig{Port: 49153, Protocol: "http"}, Database: netfs.DatabaseConfig{Path: "./tmp"}})
	if err == nil {
		server.Listen()
	}
	panic(err.Error())
}
