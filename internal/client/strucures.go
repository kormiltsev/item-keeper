package client

import pb "github.com/kormiltsev/item-keeper/internal/server/proto"

type ClientConnector struct {
	UserID     string
	Items      []*pb.Uitem
	LastUpdate int64
	User       pb.LogUserRequest
	Client     *pb.ItemKeeperClient
	Error      error
}
