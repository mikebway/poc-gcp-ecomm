// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.9
// source: mikebway/order/order_api.proto

package order

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

// OrderAPIClient is the client API for OrderAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type OrderAPIClient interface {
	// Get a specified order
	GetOrderByID(ctx context.Context, in *GetOrderByIDRequest, opts ...grpc.CallOption) (*GetOrderByIDResponse, error)
	// Get a list of orders matching some criteria
	GetOrders(ctx context.Context, in *GetOrdersRequest, opts ...grpc.CallOption) (*GetOrdersResponse, error)
}

type orderAPIClient struct {
	cc grpc.ClientConnInterface
}

func NewOrderAPIClient(cc grpc.ClientConnInterface) OrderAPIClient {
	return &orderAPIClient{cc}
}

func (c *orderAPIClient) GetOrderByID(ctx context.Context, in *GetOrderByIDRequest, opts ...grpc.CallOption) (*GetOrderByIDResponse, error) {
	out := new(GetOrderByIDResponse)
	err := c.cc.Invoke(ctx, "/mikebway.order.OrderAPI/GetOrderByID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *orderAPIClient) GetOrders(ctx context.Context, in *GetOrdersRequest, opts ...grpc.CallOption) (*GetOrdersResponse, error) {
	out := new(GetOrdersResponse)
	err := c.cc.Invoke(ctx, "/mikebway.order.OrderAPI/GetOrders", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OrderAPIServer is the server API for OrderAPI service.
// All implementations must embed UnimplementedOrderAPIServer
// for forward compatibility
type OrderAPIServer interface {
	// Get a specified order
	GetOrderByID(context.Context, *GetOrderByIDRequest) (*GetOrderByIDResponse, error)
	// Get a list of orders matching some criteria
	GetOrders(context.Context, *GetOrdersRequest) (*GetOrdersResponse, error)
	mustEmbedUnimplementedOrderAPIServer()
}

// UnimplementedOrderAPIServer must be embedded to have forward compatible implementations.
type UnimplementedOrderAPIServer struct {
}

func (UnimplementedOrderAPIServer) GetOrderByID(context.Context, *GetOrderByIDRequest) (*GetOrderByIDResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetOrderByID not implemented")
}
func (UnimplementedOrderAPIServer) GetOrders(context.Context, *GetOrdersRequest) (*GetOrdersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetOrders not implemented")
}
func (UnimplementedOrderAPIServer) mustEmbedUnimplementedOrderAPIServer() {}

// UnsafeOrderAPIServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to OrderAPIServer will
// result in compilation errors.
type UnsafeOrderAPIServer interface {
	mustEmbedUnimplementedOrderAPIServer()
}

func RegisterOrderAPIServer(s grpc.ServiceRegistrar, srv OrderAPIServer) {
	s.RegisterService(&OrderAPI_ServiceDesc, srv)
}

func _OrderAPI_GetOrderByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetOrderByIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderAPIServer).GetOrderByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mikebway.order.OrderAPI/GetOrderByID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderAPIServer).GetOrderByID(ctx, req.(*GetOrderByIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _OrderAPI_GetOrders_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetOrdersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderAPIServer).GetOrders(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mikebway.order.OrderAPI/GetOrders",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OrderAPIServer).GetOrders(ctx, req.(*GetOrdersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// OrderAPI_ServiceDesc is the grpc.ServiceDesc for OrderAPI service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var OrderAPI_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mikebway.order.OrderAPI",
	HandlerType: (*OrderAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetOrderByID",
			Handler:    _OrderAPI_GetOrderByID_Handler,
		},
		{
			MethodName: "GetOrders",
			Handler:    _OrderAPI_GetOrders_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "mikebway/order/order_api.proto",
}
