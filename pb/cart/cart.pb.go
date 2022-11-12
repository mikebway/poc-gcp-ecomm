// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.8
// source: mikebway/cart/cart.proto

package cart

import (
	types "github.com/mikebway/poc-gcp-ecomm/pb/types"
	money "google.golang.org/genproto/googleapis/type/money"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// An enumeration of shopping cart states
type ShoppingCartStatus int32

const (
	ShoppingCartStatus_SCS_UNSPECIFIED          ShoppingCartStatus = 0
	ShoppingCartStatus_SCS_OPEN                 ShoppingCartStatus = 1
	ShoppingCartStatus_SCS_ORDER_SUBMITTED      ShoppingCartStatus = 2
	ShoppingCartStatus_SCS_ABANDONED_BY_USER    ShoppingCartStatus = 3
	ShoppingCartStatus_SCS_ABANDONED_BY_TIMEOUT ShoppingCartStatus = 4
)

// Enum value maps for ShoppingCartStatus.
var (
	ShoppingCartStatus_name = map[int32]string{
		0: "SCS_UNSPECIFIED",
		1: "SCS_OPEN",
		2: "SCS_ORDER_SUBMITTED",
		3: "SCS_ABANDONED_BY_USER",
		4: "SCS_ABANDONED_BY_TIMEOUT",
	}
	ShoppingCartStatus_value = map[string]int32{
		"SCS_UNSPECIFIED":          0,
		"SCS_OPEN":                 1,
		"SCS_ORDER_SUBMITTED":      2,
		"SCS_ABANDONED_BY_USER":    3,
		"SCS_ABANDONED_BY_TIMEOUT": 4,
	}
)

func (x ShoppingCartStatus) Enum() *ShoppingCartStatus {
	p := new(ShoppingCartStatus)
	*p = x
	return p
}

func (x ShoppingCartStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ShoppingCartStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_mikebway_cart_cart_proto_enumTypes[0].Descriptor()
}

func (ShoppingCartStatus) Type() protoreflect.EnumType {
	return &file_mikebway_cart_cart_proto_enumTypes[0]
}

func (x ShoppingCartStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ShoppingCartStatus.Descriptor instead.
func (ShoppingCartStatus) EnumDescriptor() ([]byte, []int) {
	return file_mikebway_cart_cart_proto_rawDescGZIP(), []int{0}
}

// A shopping cart collects the cart items that a shopper is considering purchasing
// or has purchased. A cart should be considered immutable once purchase has been
// processed.
//
// It is persisted in the cart datastore kind.
type ShoppingCart struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A UUID ID in hexadecimal string form - a unique ID for this cart.
	// This will be set by the cart service when the cart is first created.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// The time at which shopping cart was first instantiated
	CreationTime *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=creation_time,json=creationTime,proto3" json:"creation_time,omitempty"`
	// Optional. The time at which shopping cart was closed, either  as
	// abandoned or submitted / checked out. See the Status to determine which
	ClosedTime *timestamppb.Timestamp `protobuf:"bytes,3,opt,name=closed_time,json=closedTime,proto3" json:"closed_time,omitempty"`
	// The state of the shopping cart as an enumerated value
	Status ShoppingCartStatus `protobuf:"varint,4,opt,name=status,proto3,enum=cart.ShoppingCartStatus" json:"status,omitempty"`
	// The person who opened the shopping cart
	Shopper *types.Person `protobuf:"bytes,5,opt,name=shopper,proto3" json:"shopper,omitempty"`
	// The delivery address for the order
	DeliveryAddress *types.PostalAddress `protobuf:"bytes,6,opt,name=delivery_address,json=deliveryAddress,proto3" json:"delivery_address,omitempty"`
	// Cart items is the list of one to many items that make up the potential order
	CartItems []*CartItem `protobuf:"bytes,7,rep,name=cart_items,json=cartItems,proto3" json:"cart_items,omitempty"`
}

func (x *ShoppingCart) Reset() {
	*x = ShoppingCart{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mikebway_cart_cart_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShoppingCart) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShoppingCart) ProtoMessage() {}

func (x *ShoppingCart) ProtoReflect() protoreflect.Message {
	mi := &file_mikebway_cart_cart_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShoppingCart.ProtoReflect.Descriptor instead.
func (*ShoppingCart) Descriptor() ([]byte, []int) {
	return file_mikebway_cart_cart_proto_rawDescGZIP(), []int{0}
}

func (x *ShoppingCart) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ShoppingCart) GetCreationTime() *timestamppb.Timestamp {
	if x != nil {
		return x.CreationTime
	}
	return nil
}

func (x *ShoppingCart) GetClosedTime() *timestamppb.Timestamp {
	if x != nil {
		return x.ClosedTime
	}
	return nil
}

func (x *ShoppingCart) GetStatus() ShoppingCartStatus {
	if x != nil {
		return x.Status
	}
	return ShoppingCartStatus_SCS_UNSPECIFIED
}

func (x *ShoppingCart) GetShopper() *types.Person {
	if x != nil {
		return x.Shopper
	}
	return nil
}

func (x *ShoppingCart) GetDeliveryAddress() *types.PostalAddress {
	if x != nil {
		return x.DeliveryAddress
	}
	return nil
}

func (x *ShoppingCart) GetCartItems() []*CartItem {
	if x != nil {
		return x.CartItems
	}
	return nil
}

// CartItem represents a single entry in an order. An order will contain one
// to many order items.
type CartItem struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A UUID ID in hexadecimal string form - a unique ID for this item.
	// This will be set by the cart when the item is added to the shopper's cart
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// The UUID ID (as a hexadecimal string) of the shopping cart that this item belongs to
	CartId string `protobuf:"bytes,2,opt,name=cart_id,json=cartId,proto3" json:"cart_id,omitempty"`
	// Product code is the equivalent of a SKU code identifying the type of
	// product or service being ordered.
	ProductCode string `protobuf:"bytes,3,opt,name=product_code,json=productCode,proto3" json:"product_code,omitempty"`
	// Quantity is the number of this item type that is being ordered.
	Quantity int32 `protobuf:"varint,4,opt,name=quantity,proto3" json:"quantity,omitempty"`
	// The unit price is the price that the customer was shown for a single item
	// when they selected the item for their cart
	UnitPrice *money.Money `protobuf:"bytes,5,opt,name=unit_price,json=unitPrice,proto3" json:"unit_price,omitempty"`
}

func (x *CartItem) Reset() {
	*x = CartItem{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mikebway_cart_cart_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CartItem) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CartItem) ProtoMessage() {}

func (x *CartItem) ProtoReflect() protoreflect.Message {
	mi := &file_mikebway_cart_cart_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CartItem.ProtoReflect.Descriptor instead.
func (*CartItem) Descriptor() ([]byte, []int) {
	return file_mikebway_cart_cart_proto_rawDescGZIP(), []int{1}
}

func (x *CartItem) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *CartItem) GetCartId() string {
	if x != nil {
		return x.CartId
	}
	return ""
}

func (x *CartItem) GetProductCode() string {
	if x != nil {
		return x.ProductCode
	}
	return ""
}

func (x *CartItem) GetQuantity() int32 {
	if x != nil {
		return x.Quantity
	}
	return 0
}

func (x *CartItem) GetUnitPrice() *money.Money {
	if x != nil {
		return x.UnitPrice
	}
	return nil
}

var File_mikebway_cart_cart_proto protoreflect.FileDescriptor

var file_mikebway_cart_cart_proto_rawDesc = []byte{
	0x0a, 0x18, 0x6d, 0x69, 0x6b, 0x65, 0x62, 0x77, 0x61, 0x79, 0x2f, 0x63, 0x61, 0x72, 0x74, 0x2f,
	0x63, 0x61, 0x72, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x63, 0x61, 0x72, 0x74,
	0x1a, 0x17, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x2f, 0x6d, 0x6f,
	0x6e, 0x65, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x6d, 0x69, 0x6b, 0x65, 0x62, 0x77, 0x61, 0x79,
	0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x6d, 0x69, 0x6b, 0x65, 0x62, 0x77, 0x61, 0x79, 0x2f, 0x74,
	0x79, 0x70, 0x65, 0x73, 0x2f, 0x70, 0x65, 0x72, 0x73, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0xf9, 0x02, 0x0a, 0x0c, 0x53, 0x68, 0x6f, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x43, 0x61,
	0x72, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x3f, 0x0a, 0x0d, 0x63, 0x72, 0x65, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x74,
	0x69, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0c, 0x63, 0x72, 0x65, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x54,
	0x69, 0x6d, 0x65, 0x12, 0x3b, 0x0a, 0x0b, 0x63, 0x6c, 0x6f, 0x73, 0x65, 0x64, 0x5f, 0x74, 0x69,
	0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x52, 0x0a, 0x63, 0x6c, 0x6f, 0x73, 0x65, 0x64, 0x54, 0x69, 0x6d, 0x65,
	0x12, 0x30, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x18, 0x2e, 0x63, 0x61, 0x72, 0x74, 0x2e, 0x53, 0x68, 0x6f, 0x70, 0x70, 0x69, 0x6e, 0x67,
	0x43, 0x61, 0x72, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x12, 0x30, 0x0a, 0x07, 0x73, 0x68, 0x6f, 0x70, 0x70, 0x65, 0x72, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x6d, 0x69, 0x6b, 0x65, 0x62, 0x77, 0x61, 0x79, 0x2e, 0x74,
	0x79, 0x70, 0x65, 0x73, 0x2e, 0x50, 0x65, 0x72, 0x73, 0x6f, 0x6e, 0x52, 0x07, 0x73, 0x68, 0x6f,
	0x70, 0x70, 0x65, 0x72, 0x12, 0x48, 0x0a, 0x10, 0x64, 0x65, 0x6c, 0x69, 0x76, 0x65, 0x72, 0x79,
	0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d,
	0x2e, 0x6d, 0x69, 0x6b, 0x65, 0x62, 0x77, 0x61, 0x79, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e,
	0x50, 0x6f, 0x73, 0x74, 0x61, 0x6c, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x0f, 0x64,
	0x65, 0x6c, 0x69, 0x76, 0x65, 0x72, 0x79, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x2d,
	0x0a, 0x0a, 0x63, 0x61, 0x72, 0x74, 0x5f, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x07, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x63, 0x61, 0x72, 0x74, 0x2e, 0x43, 0x61, 0x72, 0x74, 0x49, 0x74,
	0x65, 0x6d, 0x52, 0x09, 0x63, 0x61, 0x72, 0x74, 0x49, 0x74, 0x65, 0x6d, 0x73, 0x22, 0xa5, 0x01,
	0x0a, 0x08, 0x43, 0x61, 0x72, 0x74, 0x49, 0x74, 0x65, 0x6d, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x63, 0x61,
	0x72, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x63, 0x61, 0x72,
	0x74, 0x49, 0x64, 0x12, 0x21, 0x0a, 0x0c, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x5f, 0x63,
	0x6f, 0x64, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x70, 0x72, 0x6f, 0x64, 0x75,
	0x63, 0x74, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x71, 0x75, 0x61, 0x6e, 0x74, 0x69,
	0x74, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x71, 0x75, 0x61, 0x6e, 0x74, 0x69,
	0x74, 0x79, 0x12, 0x31, 0x0a, 0x0a, 0x75, 0x6e, 0x69, 0x74, 0x5f, 0x70, 0x72, 0x69, 0x63, 0x65,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x74, 0x79, 0x70, 0x65, 0x2e, 0x4d, 0x6f, 0x6e, 0x65, 0x79, 0x52, 0x09, 0x75, 0x6e, 0x69, 0x74,
	0x50, 0x72, 0x69, 0x63, 0x65, 0x2a, 0x89, 0x01, 0x0a, 0x12, 0x53, 0x68, 0x6f, 0x70, 0x70, 0x69,
	0x6e, 0x67, 0x43, 0x61, 0x72, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x13, 0x0a, 0x0f,
	0x53, 0x43, 0x53, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10,
	0x00, 0x12, 0x0c, 0x0a, 0x08, 0x53, 0x43, 0x53, 0x5f, 0x4f, 0x50, 0x45, 0x4e, 0x10, 0x01, 0x12,
	0x17, 0x0a, 0x13, 0x53, 0x43, 0x53, 0x5f, 0x4f, 0x52, 0x44, 0x45, 0x52, 0x5f, 0x53, 0x55, 0x42,
	0x4d, 0x49, 0x54, 0x54, 0x45, 0x44, 0x10, 0x02, 0x12, 0x19, 0x0a, 0x15, 0x53, 0x43, 0x53, 0x5f,
	0x41, 0x42, 0x41, 0x4e, 0x44, 0x4f, 0x4e, 0x45, 0x44, 0x5f, 0x42, 0x59, 0x5f, 0x55, 0x53, 0x45,
	0x52, 0x10, 0x03, 0x12, 0x1c, 0x0a, 0x18, 0x53, 0x43, 0x53, 0x5f, 0x41, 0x42, 0x41, 0x4e, 0x44,
	0x4f, 0x4e, 0x45, 0x44, 0x5f, 0x42, 0x59, 0x5f, 0x54, 0x49, 0x4d, 0x45, 0x4f, 0x55, 0x54, 0x10,
	0x04, 0x42, 0x2b, 0x5a, 0x29, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x6d, 0x69, 0x6b, 0x65, 0x62, 0x77, 0x61, 0x79, 0x2f, 0x70, 0x6f, 0x63, 0x2d, 0x67, 0x63, 0x70,
	0x2d, 0x65, 0x63, 0x6f, 0x6d, 0x6d, 0x2f, 0x70, 0x62, 0x2f, 0x63, 0x61, 0x72, 0x74, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mikebway_cart_cart_proto_rawDescOnce sync.Once
	file_mikebway_cart_cart_proto_rawDescData = file_mikebway_cart_cart_proto_rawDesc
)

func file_mikebway_cart_cart_proto_rawDescGZIP() []byte {
	file_mikebway_cart_cart_proto_rawDescOnce.Do(func() {
		file_mikebway_cart_cart_proto_rawDescData = protoimpl.X.CompressGZIP(file_mikebway_cart_cart_proto_rawDescData)
	})
	return file_mikebway_cart_cart_proto_rawDescData
}

var file_mikebway_cart_cart_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_mikebway_cart_cart_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_mikebway_cart_cart_proto_goTypes = []interface{}{
	(ShoppingCartStatus)(0),       // 0: cart.ShoppingCartStatus
	(*ShoppingCart)(nil),          // 1: cart.ShoppingCart
	(*CartItem)(nil),              // 2: cart.CartItem
	(*timestamppb.Timestamp)(nil), // 3: google.protobuf.Timestamp
	(*types.Person)(nil),          // 4: mikebway.types.Person
	(*types.PostalAddress)(nil),   // 5: mikebway.types.PostalAddress
	(*money.Money)(nil),           // 6: google.type.Money
}
var file_mikebway_cart_cart_proto_depIdxs = []int32{
	3, // 0: cart.ShoppingCart.creation_time:type_name -> google.protobuf.Timestamp
	3, // 1: cart.ShoppingCart.closed_time:type_name -> google.protobuf.Timestamp
	0, // 2: cart.ShoppingCart.status:type_name -> cart.ShoppingCartStatus
	4, // 3: cart.ShoppingCart.shopper:type_name -> mikebway.types.Person
	5, // 4: cart.ShoppingCart.delivery_address:type_name -> mikebway.types.PostalAddress
	2, // 5: cart.ShoppingCart.cart_items:type_name -> cart.CartItem
	6, // 6: cart.CartItem.unit_price:type_name -> google.type.Money
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_mikebway_cart_cart_proto_init() }
func file_mikebway_cart_cart_proto_init() {
	if File_mikebway_cart_cart_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_mikebway_cart_cart_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ShoppingCart); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mikebway_cart_cart_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CartItem); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_mikebway_cart_cart_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mikebway_cart_cart_proto_goTypes,
		DependencyIndexes: file_mikebway_cart_cart_proto_depIdxs,
		EnumInfos:         file_mikebway_cart_cart_proto_enumTypes,
		MessageInfos:      file_mikebway_cart_cart_proto_msgTypes,
	}.Build()
	File_mikebway_cart_cart_proto = out.File
	file_mikebway_cart_cart_proto_rawDesc = nil
	file_mikebway_cart_cart_proto_goTypes = nil
	file_mikebway_cart_cart_proto_depIdxs = nil
}