// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.21.10
// source: internal/server/proto/keeper.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	ItemKeeper_RegUser_FullMethodName    = "/item_keeper.ItemKeeper/RegUser"
	ItemKeeper_LogUser_FullMethodName    = "/item_keeper.ItemKeeper/LogUser"
	ItemKeeper_AddItem_FullMethodName    = "/item_keeper.ItemKeeper/AddItem"
	ItemKeeper_DeleteItem_FullMethodName = "/item_keeper.ItemKeeper/DeleteItem"
	ItemKeeper_UpdateItem_FullMethodName = "/item_keeper.ItemKeeper/UpdateItem"
	ItemKeeper_GetCatalog_FullMethodName = "/item_keeper.ItemKeeper/GetCatalog"
	ItemKeeper_Pictures_FullMethodName   = "/item_keeper.ItemKeeper/Pictures"
)

// ItemKeeperClient is the client API for ItemKeeper service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ItemKeeperClient interface {
	RegUser(ctx context.Context, in *RegUserRequest, opts ...grpc.CallOption) (*RegUserResponse, error)
	LogUser(ctx context.Context, in *LogUserRequest, opts ...grpc.CallOption) (*LogUserResponse, error)
	AddItem(ctx context.Context, in *AddItemRequest, opts ...grpc.CallOption) (*AddItemResponse, error)
	DeleteItem(ctx context.Context, in *DeleteItemRequest, opts ...grpc.CallOption) (*DeleteItemResponse, error)
	UpdateItem(ctx context.Context, in *UpdateItemRequest, opts ...grpc.CallOption) (*UpdateItemResponse, error)
	GetCatalog(ctx context.Context, in *GetCatalogRequest, opts ...grpc.CallOption) (*GetCatalogResponse, error)
	Pictures(ctx context.Context, in *PicturesRequest, opts ...grpc.CallOption) (ItemKeeper_PicturesClient, error)
}

type itemKeeperClient struct {
	cc grpc.ClientConnInterface
}

func NewItemKeeperClient(cc grpc.ClientConnInterface) ItemKeeperClient {
	return &itemKeeperClient{cc}
}

func (c *itemKeeperClient) RegUser(ctx context.Context, in *RegUserRequest, opts ...grpc.CallOption) (*RegUserResponse, error) {
	out := new(RegUserResponse)
	err := c.cc.Invoke(ctx, ItemKeeper_RegUser_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *itemKeeperClient) LogUser(ctx context.Context, in *LogUserRequest, opts ...grpc.CallOption) (*LogUserResponse, error) {
	out := new(LogUserResponse)
	err := c.cc.Invoke(ctx, ItemKeeper_LogUser_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *itemKeeperClient) AddItem(ctx context.Context, in *AddItemRequest, opts ...grpc.CallOption) (*AddItemResponse, error) {
	out := new(AddItemResponse)
	err := c.cc.Invoke(ctx, ItemKeeper_AddItem_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *itemKeeperClient) DeleteItem(ctx context.Context, in *DeleteItemRequest, opts ...grpc.CallOption) (*DeleteItemResponse, error) {
	out := new(DeleteItemResponse)
	err := c.cc.Invoke(ctx, ItemKeeper_DeleteItem_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *itemKeeperClient) UpdateItem(ctx context.Context, in *UpdateItemRequest, opts ...grpc.CallOption) (*UpdateItemResponse, error) {
	out := new(UpdateItemResponse)
	err := c.cc.Invoke(ctx, ItemKeeper_UpdateItem_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *itemKeeperClient) GetCatalog(ctx context.Context, in *GetCatalogRequest, opts ...grpc.CallOption) (*GetCatalogResponse, error) {
	out := new(GetCatalogResponse)
	err := c.cc.Invoke(ctx, ItemKeeper_GetCatalog_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *itemKeeperClient) Pictures(ctx context.Context, in *PicturesRequest, opts ...grpc.CallOption) (ItemKeeper_PicturesClient, error) {
	stream, err := c.cc.NewStream(ctx, &ItemKeeper_ServiceDesc.Streams[0], ItemKeeper_Pictures_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &itemKeeperPicturesClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ItemKeeper_PicturesClient interface {
	Recv() (*PicturesResponse, error)
	grpc.ClientStream
}

type itemKeeperPicturesClient struct {
	grpc.ClientStream
}

func (x *itemKeeperPicturesClient) Recv() (*PicturesResponse, error) {
	m := new(PicturesResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ItemKeeperServer is the server API for ItemKeeper service.
// All implementations must embed UnimplementedItemKeeperServer
// for forward compatibility
type ItemKeeperServer interface {
	RegUser(context.Context, *RegUserRequest) (*RegUserResponse, error)
	LogUser(context.Context, *LogUserRequest) (*LogUserResponse, error)
	AddItem(context.Context, *AddItemRequest) (*AddItemResponse, error)
	DeleteItem(context.Context, *DeleteItemRequest) (*DeleteItemResponse, error)
	UpdateItem(context.Context, *UpdateItemRequest) (*UpdateItemResponse, error)
	GetCatalog(context.Context, *GetCatalogRequest) (*GetCatalogResponse, error)
	Pictures(*PicturesRequest, ItemKeeper_PicturesServer) error
	mustEmbedUnimplementedItemKeeperServer()
}

// UnimplementedItemKeeperServer must be embedded to have forward compatible implementations.
type UnimplementedItemKeeperServer struct {
}

func (UnimplementedItemKeeperServer) RegUser(context.Context, *RegUserRequest) (*RegUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegUser not implemented")
}
func (UnimplementedItemKeeperServer) LogUser(context.Context, *LogUserRequest) (*LogUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LogUser not implemented")
}
func (UnimplementedItemKeeperServer) AddItem(context.Context, *AddItemRequest) (*AddItemResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddItem not implemented")
}
func (UnimplementedItemKeeperServer) DeleteItem(context.Context, *DeleteItemRequest) (*DeleteItemResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteItem not implemented")
}
func (UnimplementedItemKeeperServer) UpdateItem(context.Context, *UpdateItemRequest) (*UpdateItemResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateItem not implemented")
}
func (UnimplementedItemKeeperServer) GetCatalog(context.Context, *GetCatalogRequest) (*GetCatalogResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCatalog not implemented")
}
func (UnimplementedItemKeeperServer) Pictures(*PicturesRequest, ItemKeeper_PicturesServer) error {
	return status.Errorf(codes.Unimplemented, "method Pictures not implemented")
}
func (UnimplementedItemKeeperServer) mustEmbedUnimplementedItemKeeperServer() {}

// UnsafeItemKeeperServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ItemKeeperServer will
// result in compilation errors.
type UnsafeItemKeeperServer interface {
	mustEmbedUnimplementedItemKeeperServer()
}

func RegisterItemKeeperServer(s grpc.ServiceRegistrar, srv ItemKeeperServer) {
	s.RegisterService(&ItemKeeper_ServiceDesc, srv)
}

func _ItemKeeper_RegUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ItemKeeperServer).RegUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ItemKeeper_RegUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ItemKeeperServer).RegUser(ctx, req.(*RegUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ItemKeeper_LogUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ItemKeeperServer).LogUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ItemKeeper_LogUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ItemKeeperServer).LogUser(ctx, req.(*LogUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ItemKeeper_AddItem_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddItemRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ItemKeeperServer).AddItem(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ItemKeeper_AddItem_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ItemKeeperServer).AddItem(ctx, req.(*AddItemRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ItemKeeper_DeleteItem_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteItemRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ItemKeeperServer).DeleteItem(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ItemKeeper_DeleteItem_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ItemKeeperServer).DeleteItem(ctx, req.(*DeleteItemRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ItemKeeper_UpdateItem_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateItemRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ItemKeeperServer).UpdateItem(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ItemKeeper_UpdateItem_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ItemKeeperServer).UpdateItem(ctx, req.(*UpdateItemRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ItemKeeper_GetCatalog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCatalogRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ItemKeeperServer).GetCatalog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ItemKeeper_GetCatalog_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ItemKeeperServer).GetCatalog(ctx, req.(*GetCatalogRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ItemKeeper_Pictures_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(PicturesRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ItemKeeperServer).Pictures(m, &itemKeeperPicturesServer{stream})
}

type ItemKeeper_PicturesServer interface {
	Send(*PicturesResponse) error
	grpc.ServerStream
}

type itemKeeperPicturesServer struct {
	grpc.ServerStream
}

func (x *itemKeeperPicturesServer) Send(m *PicturesResponse) error {
	return x.ServerStream.SendMsg(m)
}

// ItemKeeper_ServiceDesc is the grpc.ServiceDesc for ItemKeeper service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ItemKeeper_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "item_keeper.ItemKeeper",
	HandlerType: (*ItemKeeperServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegUser",
			Handler:    _ItemKeeper_RegUser_Handler,
		},
		{
			MethodName: "LogUser",
			Handler:    _ItemKeeper_LogUser_Handler,
		},
		{
			MethodName: "AddItem",
			Handler:    _ItemKeeper_AddItem_Handler,
		},
		{
			MethodName: "DeleteItem",
			Handler:    _ItemKeeper_DeleteItem_Handler,
		},
		{
			MethodName: "UpdateItem",
			Handler:    _ItemKeeper_UpdateItem_Handler,
		},
		{
			MethodName: "GetCatalog",
			Handler:    _ItemKeeper_GetCatalog_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Pictures",
			Handler:       _ItemKeeper_Pictures_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "internal/server/proto/keeper.proto",
}