package main

import (
	"context"

	server "github.com/kormiltsev/item-keeper/internal/server"
)

func main() {
	ctx := context.Background()
	server.StartServer(ctx)
}
