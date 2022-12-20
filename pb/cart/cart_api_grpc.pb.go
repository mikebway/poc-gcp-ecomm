// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.9
// source: mikebway/cart/cart_api.proto

package cart

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

// CartAPIClient is the client API for CartAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CartAPIClient interface {
	// Create a new shopping cart
	CreateShoppingCart(ctx context.Context, in *CreateShoppingCartRequest, opts ...grpc.CallOption) (*CreateShoppingCartResponse, error)
	// Retrieve a cart by UUID ID
	GetShoppingCartByID(ctx context.Context, in *GetShoppingCartByIDRequest, opts ...grpc.CallOption) (*GetShoppingCartByIDResponse, error)
	// Add an item to a cart
	AddItemToShoppingCart(ctx context.Context, in *AddItemToShoppingCartRequest, opts ...grpc.CallOption) (*AddItemToShoppingCartResponse, error)
	// Remove an item from the cart
	RemoveItemFromShoppingCart(ctx context.Context, in *RemoveItemFromShoppingCartRequest, opts ...grpc.CallOption) (*RemoveItemFromShoppingCartResponse, error)
	// Set the delivery address for physical cart items
	SetDeliveryAddress(ctx context.Context, in *SetDeliveryAddressRequest, opts ...grpc.CallOption) (*SetDeliveryAddressResponse, error)
	// Submit the order / checkout the shopping cart
	CheckoutShoppingCart(ctx context.Context, in *CheckoutShoppingCartRequest, opts ...grpc.CallOption) (*CheckoutShoppingCartResponse, error)
	// Explicitly abandon a shopping cart in response to a user request.
	AbandonShoppingCart(ctx context.Context, in *AbandonShoppingCartRequest, opts ...grpc.CallOption) (*AbandonShoppingCartResponse, error)
}

type cartAPIClient struct {
	cc grpc.ClientConnInterface
}

func NewCartAPIClient(cc grpc.ClientConnInterface) CartAPIClient {
	return &cartAPIClient{cc}
}

func (c *cartAPIClient) CreateShoppingCart(ctx context.Context, in *CreateShoppingCartRequest, opts ...grpc.CallOption) (*CreateShoppingCartResponse, error) {
	out := new(CreateShoppingCartResponse)
	err := c.cc.Invoke(ctx, "/mikebway.cart.CartAPI/CreateShoppingCart", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cartAPIClient) GetShoppingCartByID(ctx context.Context, in *GetShoppingCartByIDRequest, opts ...grpc.CallOption) (*GetShoppingCartByIDResponse, error) {
	out := new(GetShoppingCartByIDResponse)
	err := c.cc.Invoke(ctx, "/mikebway.cart.CartAPI/GetShoppingCartByID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cartAPIClient) AddItemToShoppingCart(ctx context.Context, in *AddItemToShoppingCartRequest, opts ...grpc.CallOption) (*AddItemToShoppingCartResponse, error) {
	out := new(AddItemToShoppingCartResponse)
	err := c.cc.Invoke(ctx, "/mikebway.cart.CartAPI/AddItemToShoppingCart", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cartAPIClient) RemoveItemFromShoppingCart(ctx context.Context, in *RemoveItemFromShoppingCartRequest, opts ...grpc.CallOption) (*RemoveItemFromShoppingCartResponse, error) {
	out := new(RemoveItemFromShoppingCartResponse)
	err := c.cc.Invoke(ctx, "/mikebway.cart.CartAPI/RemoveItemFromShoppingCart", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cartAPIClient) SetDeliveryAddress(ctx context.Context, in *SetDeliveryAddressRequest, opts ...grpc.CallOption) (*SetDeliveryAddressResponse, error) {
	out := new(SetDeliveryAddressResponse)
	err := c.cc.Invoke(ctx, "/mikebway.cart.CartAPI/SetDeliveryAddress", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cartAPIClient) CheckoutShoppingCart(ctx context.Context, in *CheckoutShoppingCartRequest, opts ...grpc.CallOption) (*CheckoutShoppingCartResponse, error) {
	out := new(CheckoutShoppingCartResponse)
	err := c.cc.Invoke(ctx, "/mikebway.cart.CartAPI/CheckoutShoppingCart", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cartAPIClient) AbandonShoppingCart(ctx context.Context, in *AbandonShoppingCartRequest, opts ...grpc.CallOption) (*AbandonShoppingCartResponse, error) {
	out := new(AbandonShoppingCartResponse)
	err := c.cc.Invoke(ctx, "/mikebway.cart.CartAPI/AbandonShoppingCart", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CartAPIServer is the server API for CartAPI service.
// All implementations must embed UnimplementedCartAPIServer
// for forward compatibility
type CartAPIServer interface {
	// Create a new shopping cart
	CreateShoppingCart(context.Context, *CreateShoppingCartRequest) (*CreateShoppingCartResponse, error)
	// Retrieve a cart by UUID ID
	GetShoppingCartByID(context.Context, *GetShoppingCartByIDRequest) (*GetShoppingCartByIDResponse, error)
	// Add an item to a cart
	AddItemToShoppingCart(context.Context, *AddItemToShoppingCartRequest) (*AddItemToShoppingCartResponse, error)
	// Remove an item from the cart
	RemoveItemFromShoppingCart(context.Context, *RemoveItemFromShoppingCartRequest) (*RemoveItemFromShoppingCartResponse, error)
	// Set the delivery address for physical cart items
	SetDeliveryAddress(context.Context, *SetDeliveryAddressRequest) (*SetDeliveryAddressResponse, error)
	// Submit the order / checkout the shopping cart
	CheckoutShoppingCart(context.Context, *CheckoutShoppingCartRequest) (*CheckoutShoppingCartResponse, error)
	// Explicitly abandon a shopping cart in response to a user request.
	AbandonShoppingCart(context.Context, *AbandonShoppingCartRequest) (*AbandonShoppingCartResponse, error)
	mustEmbedUnimplementedCartAPIServer()
}

// UnimplementedCartAPIServer must be embedded to have forward compatible implementations.
type UnimplementedCartAPIServer struct {
}

func (UnimplementedCartAPIServer) CreateShoppingCart(context.Context, *CreateShoppingCartRequest) (*CreateShoppingCartResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateShoppingCart not implemented")
}
func (UnimplementedCartAPIServer) GetShoppingCartByID(context.Context, *GetShoppingCartByIDRequest) (*GetShoppingCartByIDResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetShoppingCartByID not implemented")
}
func (UnimplementedCartAPIServer) AddItemToShoppingCart(context.Context, *AddItemToShoppingCartRequest) (*AddItemToShoppingCartResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddItemToShoppingCart not implemented")
}
func (UnimplementedCartAPIServer) RemoveItemFromShoppingCart(context.Context, *RemoveItemFromShoppingCartRequest) (*RemoveItemFromShoppingCartResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveItemFromShoppingCart not implemented")
}
func (UnimplementedCartAPIServer) SetDeliveryAddress(context.Context, *SetDeliveryAddressRequest) (*SetDeliveryAddressResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetDeliveryAddress not implemented")
}
func (UnimplementedCartAPIServer) CheckoutShoppingCart(context.Context, *CheckoutShoppingCartRequest) (*CheckoutShoppingCartResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckoutShoppingCart not implemented")
}
func (UnimplementedCartAPIServer) AbandonShoppingCart(context.Context, *AbandonShoppingCartRequest) (*AbandonShoppingCartResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AbandonShoppingCart not implemented")
}
func (UnimplementedCartAPIServer) mustEmbedUnimplementedCartAPIServer() {}

// UnsafeCartAPIServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CartAPIServer will
// result in compilation errors.
type UnsafeCartAPIServer interface {
	mustEmbedUnimplementedCartAPIServer()
}

func RegisterCartAPIServer(s grpc.ServiceRegistrar, srv CartAPIServer) {
	s.RegisterService(&CartAPI_ServiceDesc, srv)
}

func _CartAPI_CreateShoppingCart_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateShoppingCartRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CartAPIServer).CreateShoppingCart(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mikebway.cart.CartAPI/CreateShoppingCart",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CartAPIServer).CreateShoppingCart(ctx, req.(*CreateShoppingCartRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CartAPI_GetShoppingCartByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetShoppingCartByIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CartAPIServer).GetShoppingCartByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mikebway.cart.CartAPI/GetShoppingCartByID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CartAPIServer).GetShoppingCartByID(ctx, req.(*GetShoppingCartByIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CartAPI_AddItemToShoppingCart_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddItemToShoppingCartRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CartAPIServer).AddItemToShoppingCart(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mikebway.cart.CartAPI/AddItemToShoppingCart",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CartAPIServer).AddItemToShoppingCart(ctx, req.(*AddItemToShoppingCartRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CartAPI_RemoveItemFromShoppingCart_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveItemFromShoppingCartRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CartAPIServer).RemoveItemFromShoppingCart(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mikebway.cart.CartAPI/RemoveItemFromShoppingCart",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CartAPIServer).RemoveItemFromShoppingCart(ctx, req.(*RemoveItemFromShoppingCartRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CartAPI_SetDeliveryAddress_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetDeliveryAddressRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CartAPIServer).SetDeliveryAddress(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mikebway.cart.CartAPI/SetDeliveryAddress",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CartAPIServer).SetDeliveryAddress(ctx, req.(*SetDeliveryAddressRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CartAPI_CheckoutShoppingCart_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckoutShoppingCartRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CartAPIServer).CheckoutShoppingCart(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mikebway.cart.CartAPI/CheckoutShoppingCart",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CartAPIServer).CheckoutShoppingCart(ctx, req.(*CheckoutShoppingCartRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CartAPI_AbandonShoppingCart_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AbandonShoppingCartRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CartAPIServer).AbandonShoppingCart(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mikebway.cart.CartAPI/AbandonShoppingCart",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CartAPIServer).AbandonShoppingCart(ctx, req.(*AbandonShoppingCartRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CartAPI_ServiceDesc is the grpc.ServiceDesc for CartAPI service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CartAPI_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mikebway.cart.CartAPI",
	HandlerType: (*CartAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateShoppingCart",
			Handler:    _CartAPI_CreateShoppingCart_Handler,
		},
		{
			MethodName: "GetShoppingCartByID",
			Handler:    _CartAPI_GetShoppingCartByID_Handler,
		},
		{
			MethodName: "AddItemToShoppingCart",
			Handler:    _CartAPI_AddItemToShoppingCart_Handler,
		},
		{
			MethodName: "RemoveItemFromShoppingCart",
			Handler:    _CartAPI_RemoveItemFromShoppingCart_Handler,
		},
		{
			MethodName: "SetDeliveryAddress",
			Handler:    _CartAPI_SetDeliveryAddress_Handler,
		},
		{
			MethodName: "CheckoutShoppingCart",
			Handler:    _CartAPI_CheckoutShoppingCart_Handler,
		},
		{
			MethodName: "AbandonShoppingCart",
			Handler:    _CartAPI_AbandonShoppingCart_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "mikebway/cart/cart_api.proto",
}
