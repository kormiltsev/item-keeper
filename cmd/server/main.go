package main

import (
	"context"

	server "github.com/kormiltsev/item-keeper/internal/server"
)

func main() {
	// chan close signal
	var chclose = make(chan struct{})
	// start server
	ctx := context.Background()
	go server.StartServer(ctx, chclose)

	server.StartServerGRPC(":3333")

	close(chclose)
}
