package server

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const clientTokenName = "CLIENT_TOKEN"

// unaryInterceptor searche for userid in token
func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var err error

	// check metadata exists
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "token not found")
	}

	// check token exists
	values := md.Get(clientTokenName)
	if len(values) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "tokens list empty")
	}

	// check every of values
	for _, token := range values {
		if len(token) == 0 {
			err = status.Errorf(codes.Unauthenticated, "token empty")
		}
	}
	// all tokens are empty
	if err != nil {
		return nil, err
	}

	// check is token valid here
	//
	// ========================

	// OK
	return handler(ctx, req)
}
