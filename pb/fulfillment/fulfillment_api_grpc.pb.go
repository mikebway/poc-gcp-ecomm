// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.9
// source: mikebway/fulfillment/fulfillment_api.proto

package fulfillment

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

// FulfillmentAPIClient is the client API for FulfillmentAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FulfillmentAPIClient interface {
	// Get a specified task
	GetTaskByID(ctx context.Context, in *GetTaskByIDRequest, opts ...grpc.CallOption) (*GetTaskByIDResponse, error)
	// Get a list of Tasks matching some criteria
	GetTasks(ctx context.Context, in *GetTasksRequest, opts ...grpc.CallOption) (*GetTasksResponse, error)
	// Update the status of a task
	UpdateTaskStatus(ctx context.Context, in *UpdateTaskStatusRequest, opts ...grpc.CallOption) (*UpdateTaskStatusResponse, error)
}

type fulfillmentAPIClient struct {
	cc grpc.ClientConnInterface
}

func NewFulfillmentAPIClient(cc grpc.ClientConnInterface) FulfillmentAPIClient {
	return &fulfillmentAPIClient{cc}
}

func (c *fulfillmentAPIClient) GetTaskByID(ctx context.Context, in *GetTaskByIDRequest, opts ...grpc.CallOption) (*GetTaskByIDResponse, error) {
	out := new(GetTaskByIDResponse)
	err := c.cc.Invoke(ctx, "/mikebway.fulfillment.FulfillmentAPI/GetTaskByID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fulfillmentAPIClient) GetTasks(ctx context.Context, in *GetTasksRequest, opts ...grpc.CallOption) (*GetTasksResponse, error) {
	out := new(GetTasksResponse)
	err := c.cc.Invoke(ctx, "/mikebway.fulfillment.FulfillmentAPI/GetTasks", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fulfillmentAPIClient) UpdateTaskStatus(ctx context.Context, in *UpdateTaskStatusRequest, opts ...grpc.CallOption) (*UpdateTaskStatusResponse, error) {
	out := new(UpdateTaskStatusResponse)
	err := c.cc.Invoke(ctx, "/mikebway.fulfillment.FulfillmentAPI/UpdateTaskStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FulfillmentAPIServer is the server API for FulfillmentAPI service.
// All implementations must embed UnimplementedFulfillmentAPIServer
// for forward compatibility
type FulfillmentAPIServer interface {
	// Get a specified task
	GetTaskByID(context.Context, *GetTaskByIDRequest) (*GetTaskByIDResponse, error)
	// Get a list of Tasks matching some criteria
	GetTasks(context.Context, *GetTasksRequest) (*GetTasksResponse, error)
	// Update the status of a task
	UpdateTaskStatus(context.Context, *UpdateTaskStatusRequest) (*UpdateTaskStatusResponse, error)
	mustEmbedUnimplementedFulfillmentAPIServer()
}

// UnimplementedFulfillmentAPIServer must be embedded to have forward compatible implementations.
type UnimplementedFulfillmentAPIServer struct {
}

func (UnimplementedFulfillmentAPIServer) GetTaskByID(context.Context, *GetTaskByIDRequest) (*GetTaskByIDResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTaskByID not implemented")
}
func (UnimplementedFulfillmentAPIServer) GetTasks(context.Context, *GetTasksRequest) (*GetTasksResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTasks not implemented")
}
func (UnimplementedFulfillmentAPIServer) UpdateTaskStatus(context.Context, *UpdateTaskStatusRequest) (*UpdateTaskStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateTaskStatus not implemented")
}
func (UnimplementedFulfillmentAPIServer) mustEmbedUnimplementedFulfillmentAPIServer() {}

// UnsafeFulfillmentAPIServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FulfillmentAPIServer will
// result in compilation errors.
type UnsafeFulfillmentAPIServer interface {
	mustEmbedUnimplementedFulfillmentAPIServer()
}

func RegisterFulfillmentAPIServer(s grpc.ServiceRegistrar, srv FulfillmentAPIServer) {
	s.RegisterService(&FulfillmentAPI_ServiceDesc, srv)
}

func _FulfillmentAPI_GetTaskByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTaskByIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FulfillmentAPIServer).GetTaskByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mikebway.fulfillment.FulfillmentAPI/GetTaskByID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FulfillmentAPIServer).GetTaskByID(ctx, req.(*GetTaskByIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FulfillmentAPI_GetTasks_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTasksRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FulfillmentAPIServer).GetTasks(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mikebway.fulfillment.FulfillmentAPI/GetTasks",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FulfillmentAPIServer).GetTasks(ctx, req.(*GetTasksRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FulfillmentAPI_UpdateTaskStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateTaskStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FulfillmentAPIServer).UpdateTaskStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mikebway.fulfillment.FulfillmentAPI/UpdateTaskStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FulfillmentAPIServer).UpdateTaskStatus(ctx, req.(*UpdateTaskStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// FulfillmentAPI_ServiceDesc is the grpc.ServiceDesc for FulfillmentAPI service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FulfillmentAPI_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mikebway.fulfillment.FulfillmentAPI",
	HandlerType: (*FulfillmentAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetTaskByID",
			Handler:    _FulfillmentAPI_GetTaskByID_Handler,
		},
		{
			MethodName: "GetTasks",
			Handler:    _FulfillmentAPI_GetTasks_Handler,
		},
		{
			MethodName: "UpdateTaskStatus",
			Handler:    _FulfillmentAPI_UpdateTaskStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "mikebway/fulfillment/fulfillment_api.proto",
}
