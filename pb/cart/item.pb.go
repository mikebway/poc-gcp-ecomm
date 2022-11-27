// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.8
// source: mikebway/cart/item.proto

package cart

import (
	money "google.golang.org/genproto/googleapis/type/money"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

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
		mi := &file_mikebway_cart_item_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CartItem) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CartItem) ProtoMessage() {}

func (x *CartItem) ProtoReflect() protoreflect.Message {
	mi := &file_mikebway_cart_item_proto_msgTypes[0]
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
	return file_mikebway_cart_item_proto_rawDescGZIP(), []int{0}
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

var File_mikebway_cart_item_proto protoreflect.FileDescriptor

var file_mikebway_cart_item_proto_rawDesc = []byte{
	0x0a, 0x18, 0x6d, 0x69, 0x6b, 0x65, 0x62, 0x77, 0x61, 0x79, 0x2f, 0x63, 0x61, 0x72, 0x74, 0x2f,
	0x69, 0x74, 0x65, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x63, 0x61, 0x72, 0x74,
	0x1a, 0x17, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x2f, 0x6d, 0x6f,
	0x6e, 0x65, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa5, 0x01, 0x0a, 0x08, 0x43, 0x61,
	0x72, 0x74, 0x49, 0x74, 0x65, 0x6d, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x63, 0x61, 0x72, 0x74, 0x5f, 0x69,
	0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x63, 0x61, 0x72, 0x74, 0x49, 0x64, 0x12,
	0x21, 0x0a, 0x0c, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x43, 0x6f,
	0x64, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x71, 0x75, 0x61, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x71, 0x75, 0x61, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x31,
	0x0a, 0x0a, 0x75, 0x6e, 0x69, 0x74, 0x5f, 0x70, 0x72, 0x69, 0x63, 0x65, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x12, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x74, 0x79, 0x70, 0x65,
	0x2e, 0x4d, 0x6f, 0x6e, 0x65, 0x79, 0x52, 0x09, 0x75, 0x6e, 0x69, 0x74, 0x50, 0x72, 0x69, 0x63,
	0x65, 0x42, 0x2b, 0x5a, 0x29, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x6d, 0x69, 0x6b, 0x65, 0x62, 0x77, 0x61, 0x79, 0x2f, 0x70, 0x6f, 0x63, 0x2d, 0x67, 0x63, 0x70,
	0x2d, 0x65, 0x63, 0x6f, 0x6d, 0x6d, 0x2f, 0x70, 0x62, 0x2f, 0x63, 0x61, 0x72, 0x74, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mikebway_cart_item_proto_rawDescOnce sync.Once
	file_mikebway_cart_item_proto_rawDescData = file_mikebway_cart_item_proto_rawDesc
)

func file_mikebway_cart_item_proto_rawDescGZIP() []byte {
	file_mikebway_cart_item_proto_rawDescOnce.Do(func() {
		file_mikebway_cart_item_proto_rawDescData = protoimpl.X.CompressGZIP(file_mikebway_cart_item_proto_rawDescData)
	})
	return file_mikebway_cart_item_proto_rawDescData
}

var file_mikebway_cart_item_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_mikebway_cart_item_proto_goTypes = []interface{}{
	(*CartItem)(nil),    // 0: cart.CartItem
	(*money.Money)(nil), // 1: google.type.Money
}
var file_mikebway_cart_item_proto_depIdxs = []int32{
	1, // 0: cart.CartItem.unit_price:type_name -> google.type.Money
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_mikebway_cart_item_proto_init() }
func file_mikebway_cart_item_proto_init() {
	if File_mikebway_cart_item_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_mikebway_cart_item_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
			RawDescriptor: file_mikebway_cart_item_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mikebway_cart_item_proto_goTypes,
		DependencyIndexes: file_mikebway_cart_item_proto_depIdxs,
		MessageInfos:      file_mikebway_cart_item_proto_msgTypes,
	}.Build()
	File_mikebway_cart_item_proto = out.File
	file_mikebway_cart_item_proto_rawDesc = nil
	file_mikebway_cart_item_proto_goTypes = nil
	file_mikebway_cart_item_proto_depIdxs = nil
}
