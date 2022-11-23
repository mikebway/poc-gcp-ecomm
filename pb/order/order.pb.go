// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.9
// source: mikebway/order/order.proto

package order

import (
	types "github.com/mikebway/poc-gcp-ecomm/pb/types"
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

// An order is generated by the shopping cart service on checkout. Posted to the event bus by the cart service,
// it will trigger the payment phase.
type Order struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A UUID ID in hexadecimal string form - a unique ID for this order.
	// This will be set by the cart when the order is submitted.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// The time at which cart checkout was completed and the order was submitted for payment
	SubmissionTime *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=submission_time,json=submissionTime,proto3" json:"submission_time,omitempty"`
	// The person who submitted the order
	OrderedBy *types.Person `protobuf:"bytes,3,opt,name=ordered_by,json=orderedBy,proto3" json:"ordered_by,omitempty"`
	// The delivery address for the order
	DeliveryAddress *types.PostalAddress `protobuf:"bytes,4,opt,name=delivery_address,json=deliveryAddress,proto3" json:"delivery_address,omitempty"`
	// Order items is the list of one to many items that make up the potential order
	OrderItems []*OrderItem `protobuf:"bytes,5,rep,name=order_items,json=orderItems,proto3" json:"order_items,omitempty"`
}

func (x *Order) Reset() {
	*x = Order{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mikebway_order_order_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Order) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Order) ProtoMessage() {}

func (x *Order) ProtoReflect() protoreflect.Message {
	mi := &file_mikebway_order_order_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Order.ProtoReflect.Descriptor instead.
func (*Order) Descriptor() ([]byte, []int) {
	return file_mikebway_order_order_proto_rawDescGZIP(), []int{0}
}

func (x *Order) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Order) GetSubmissionTime() *timestamppb.Timestamp {
	if x != nil {
		return x.SubmissionTime
	}
	return nil
}

func (x *Order) GetOrderedBy() *types.Person {
	if x != nil {
		return x.OrderedBy
	}
	return nil
}

func (x *Order) GetDeliveryAddress() *types.PostalAddress {
	if x != nil {
		return x.DeliveryAddress
	}
	return nil
}

func (x *Order) GetOrderItems() []*OrderItem {
	if x != nil {
		return x.OrderItems
	}
	return nil
}

var File_mikebway_order_order_proto protoreflect.FileDescriptor

var file_mikebway_order_order_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x6d, 0x69, 0x6b, 0x65, 0x62, 0x77, 0x61, 0x79, 0x2f, 0x6f, 0x72, 0x64, 0x65, 0x72,
	0x2f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x6d, 0x69,
	0x6b, 0x65, 0x62, 0x77, 0x61, 0x79, 0x2e, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x1a, 0x1b, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x6d, 0x69, 0x6b, 0x65, 0x62,
	0x77, 0x61, 0x79, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x6d, 0x69, 0x6b, 0x65, 0x62, 0x77, 0x61,
	0x79, 0x2f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x2f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x69, 0x74, 0x65,
	0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x6d, 0x69, 0x6b, 0x65, 0x62, 0x77, 0x61,
	0x79, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x70, 0x65, 0x72, 0x73, 0x6f, 0x6e, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x99, 0x02, 0x0a, 0x05, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x43,
	0x0a, 0x0f, 0x73, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x69, 0x6d,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x52, 0x0e, 0x73, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x54,
	0x69, 0x6d, 0x65, 0x12, 0x35, 0x0a, 0x0a, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x64, 0x5f, 0x62,
	0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x6d, 0x69, 0x6b, 0x65, 0x62, 0x77,
	0x61, 0x79, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x50, 0x65, 0x72, 0x73, 0x6f, 0x6e, 0x52,
	0x09, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x65, 0x64, 0x42, 0x79, 0x12, 0x48, 0x0a, 0x10, 0x64, 0x65,
	0x6c, 0x69, 0x76, 0x65, 0x72, 0x79, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x6d, 0x69, 0x6b, 0x65, 0x62, 0x77, 0x61, 0x79, 0x2e,
	0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x61, 0x6c, 0x41, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x52, 0x0f, 0x64, 0x65, 0x6c, 0x69, 0x76, 0x65, 0x72, 0x79, 0x41, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x12, 0x3a, 0x0a, 0x0b, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x5f, 0x69, 0x74,
	0x65, 0x6d, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x6d, 0x69, 0x6b, 0x65,
	0x62, 0x77, 0x61, 0x79, 0x2e, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x2e, 0x4f, 0x72, 0x64, 0x65, 0x72,
	0x49, 0x74, 0x65, 0x6d, 0x52, 0x0a, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x49, 0x74, 0x65, 0x6d, 0x73,
	0x42, 0x2c, 0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d,
	0x69, 0x6b, 0x65, 0x62, 0x77, 0x61, 0x79, 0x2f, 0x70, 0x6f, 0x63, 0x2d, 0x67, 0x63, 0x70, 0x2d,
	0x65, 0x63, 0x6f, 0x6d, 0x6d, 0x2f, 0x70, 0x62, 0x2f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mikebway_order_order_proto_rawDescOnce sync.Once
	file_mikebway_order_order_proto_rawDescData = file_mikebway_order_order_proto_rawDesc
)

func file_mikebway_order_order_proto_rawDescGZIP() []byte {
	file_mikebway_order_order_proto_rawDescOnce.Do(func() {
		file_mikebway_order_order_proto_rawDescData = protoimpl.X.CompressGZIP(file_mikebway_order_order_proto_rawDescData)
	})
	return file_mikebway_order_order_proto_rawDescData
}

var file_mikebway_order_order_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_mikebway_order_order_proto_goTypes = []interface{}{
	(*Order)(nil),                 // 0: mikebway.order.Order
	(*timestamppb.Timestamp)(nil), // 1: google.protobuf.Timestamp
	(*types.Person)(nil),          // 2: mikebway.types.Person
	(*types.PostalAddress)(nil),   // 3: mikebway.types.PostalAddress
	(*OrderItem)(nil),             // 4: mikebway.order.OrderItem
}
var file_mikebway_order_order_proto_depIdxs = []int32{
	1, // 0: mikebway.order.Order.submission_time:type_name -> google.protobuf.Timestamp
	2, // 1: mikebway.order.Order.ordered_by:type_name -> mikebway.types.Person
	3, // 2: mikebway.order.Order.delivery_address:type_name -> mikebway.types.PostalAddress
	4, // 3: mikebway.order.Order.order_items:type_name -> mikebway.order.OrderItem
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_mikebway_order_order_proto_init() }
func file_mikebway_order_order_proto_init() {
	if File_mikebway_order_order_proto != nil {
		return
	}
	file_mikebway_order_orderitem_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_mikebway_order_order_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Order); i {
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
			RawDescriptor: file_mikebway_order_order_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mikebway_order_order_proto_goTypes,
		DependencyIndexes: file_mikebway_order_order_proto_depIdxs,
		MessageInfos:      file_mikebway_order_order_proto_msgTypes,
	}.Build()
	File_mikebway_order_order_proto = out.File
	file_mikebway_order_order_proto_rawDesc = nil
	file_mikebway_order_order_proto_goTypes = nil
	file_mikebway_order_order_proto_depIdxs = nil
}
