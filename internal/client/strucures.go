package client

import (
	"context"

	pb "github.com/kormiltsev/item-keeper/internal/server/proto"
)

// ClientConnector store gRPC connection
type ClientConnector struct {
	Ctx    context.Context
	Client *pb.ItemKeeperClient
}
