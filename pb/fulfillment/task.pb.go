// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.9
// source: mikebway/fulfillment/task.proto

package fulfillment

import (
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

// An enumeration of the possible task states
type TaskStatus int32

const (
	TaskStatus_UNDEFINED           TaskStatus = 0   // Should not be seen - indicates that the status has not be set
	TaskStatus_WAITING_TASK        TaskStatus = 1   // Waiting on another task to complete.
	TaskStatus_WAITING_CUSTOMER    TaskStatus = 2   // Waiting for customer input.
	TaskStatus_WAITING_PAYMENT     TaskStatus = 3   // Waiting for a customer payment to be confirmed.
	TaskStatus_WAITING_CS          TaskStatus = 4   // Waiting on customer service.
	TaskStatus_WAITING_SERVICE     TaskStatus = 5   // Waiting on an internal customer service.
	TaskStatus_WAITING_THIRD_PARTY TaskStatus = 6   // Waiting on a third party service.
	TaskStatus_PAUSED              TaskStatus = 98  // Paused; see reason_code.
	TaskStatus_CANCELED            TaskStatus = 99  // The task has been canceled; see reason_code.
	TaskStatus_COMPLETED           TaskStatus = 100 // The task has been completed
)

// Enum value maps for TaskStatus.
var (
	TaskStatus_name = map[int32]string{
		0:   "UNDEFINED",
		1:   "WAITING_TASK",
		2:   "WAITING_CUSTOMER",
		3:   "WAITING_PAYMENT",
		4:   "WAITING_CS",
		5:   "WAITING_SERVICE",
		6:   "WAITING_THIRD_PARTY",
		98:  "PAUSED",
		99:  "CANCELED",
		100: "COMPLETED",
	}
	TaskStatus_value = map[string]int32{
		"UNDEFINED":           0,
		"WAITING_TASK":        1,
		"WAITING_CUSTOMER":    2,
		"WAITING_PAYMENT":     3,
		"WAITING_CS":          4,
		"WAITING_SERVICE":     5,
		"WAITING_THIRD_PARTY": 6,
		"PAUSED":              98,
		"CANCELED":            99,
		"COMPLETED":           100,
	}
)

func (x TaskStatus) Enum() *TaskStatus {
	p := new(TaskStatus)
	*p = x
	return p
}

func (x TaskStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TaskStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_mikebway_fulfillment_task_proto_enumTypes[0].Descriptor()
}

func (TaskStatus) Type() protoreflect.EnumType {
	return &file_mikebway_fulfillment_task_proto_enumTypes[0]
}

func (x TaskStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TaskStatus.Descriptor instead.
func (TaskStatus) EnumDescriptor() ([]byte, []int) {
	return file_mikebway_fulfillment_task_proto_rawDescGZIP(), []int{0}
}

// Task defines a fulfilment activity that is being tracked by the Fulfillment Orchestration Service. Fulfillment tasks
// map to order items in a many-to-one relationship, i.e. a single item in a customer order may map to multiple
// fulfillment tasks.
//
// For example, an order might include two items such as a custom sofa and two end tables (one order line item with a
// quantity of two). Fulfillment tasks for the sofa could include: order cover material, manufacture, ship, and install.
type Task struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A UUID ID in hexadecimal string form - a unique ID for this task.
	// This will be set when the task is created in response to a new order being
	// received by the fulfilment service.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// submission_time is the time at which task was submitted to the Fulfillment Orchestration Service.
	SubmissionTime *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=submission_time,json=submissionTime,proto3" json:"submission_time,omitempty"`
	// completion_time is the time at which task was marked as completed
	CompletionTime *timestamppb.Timestamp `protobuf:"bytes,3,opt,name=completion_time,json=completionTime,proto3" json:"completion_time,omitempty"`
	// order_id relates this task to the order containing the item that requires the task to be performed.
	// order_id is a UUID ID in hexadecimal string form - a unique ID for the order.
	OrderId string `protobuf:"bytes,4,opt,name=order_id,json=orderId,proto3" json:"order_id,omitempty"`
	// order_item_id relates this task to the order item that requires the task to be performed.
	// order_item_id is a UUID ID in hexadecimal string form - a unique ID for the order item.
	OrderItemId string `protobuf:"bytes,5,opt,name=order_item_id,json=orderItemId,proto3" json:"order_item_id,omitempty"`
	// product_code is the equivalent of a SKU code identifying the type of product or service that the OrderItemId
	// is for.
	ProductCode string `protobuf:"bytes,6,opt,name=product_code,json=productCode,proto3" json:"product_code,omitempty"`
	// task_code, in combination with the product_code, identifies the type of activity to be performed and the data,
	// if any, to be collected.
	TaskCode string `protobuf:"bytes,7,opt,name=task_code,json=taskCode,proto3" json:"task_code,omitempty"`
	// Status identifies the status of the task, i.e. whether we are waiting for customer input, waiting for a
	// response from a third party service, or that the task has bee completed or failed.
	Status TaskStatus `protobuf:"varint,8,opt,name=status,proto3,enum=mikebway.fulfillment.TaskStatus" json:"status,omitempty"`
	// reason_code provides a key that can be used to look up a localized explanation for why the status
	// is WAITING_CUSTOMER, PAUSED, CANCELED, etc. Hopefully, people will choose reason codes that convey some meaning
	// by themselves saving engineers with only the raw data from having to translate the value into something more
	// intelligible.
	//
	// It is possible that the ReasonCode might need to be interpreted in context with the combination of the task_code
	// and product_code.
	ReasonCode string `protobuf:"bytes,9,opt,name=reason_code,json=reasonCode,proto3" json:"reason_code,omitempty"`
	// Parameters is a list of zero to many named string parameters that might be required to complete the task.
	Parameters []*Parameter `protobuf:"bytes,10,rep,name=parameters,proto3" json:"parameters,omitempty"`
}

func (x *Task) Reset() {
	*x = Task{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mikebway_fulfillment_task_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Task) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Task) ProtoMessage() {}

func (x *Task) ProtoReflect() protoreflect.Message {
	mi := &file_mikebway_fulfillment_task_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Task.ProtoReflect.Descriptor instead.
func (*Task) Descriptor() ([]byte, []int) {
	return file_mikebway_fulfillment_task_proto_rawDescGZIP(), []int{0}
}

func (x *Task) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Task) GetSubmissionTime() *timestamppb.Timestamp {
	if x != nil {
		return x.SubmissionTime
	}
	return nil
}

func (x *Task) GetCompletionTime() *timestamppb.Timestamp {
	if x != nil {
		return x.CompletionTime
	}
	return nil
}

func (x *Task) GetOrderId() string {
	if x != nil {
		return x.OrderId
	}
	return ""
}

func (x *Task) GetOrderItemId() string {
	if x != nil {
		return x.OrderItemId
	}
	return ""
}

func (x *Task) GetProductCode() string {
	if x != nil {
		return x.ProductCode
	}
	return ""
}

func (x *Task) GetTaskCode() string {
	if x != nil {
		return x.TaskCode
	}
	return ""
}

func (x *Task) GetStatus() TaskStatus {
	if x != nil {
		return x.Status
	}
	return TaskStatus_UNDEFINED
}

func (x *Task) GetReasonCode() string {
	if x != nil {
		return x.ReasonCode
	}
	return ""
}

func (x *Task) GetParameters() []*Parameter {
	if x != nil {
		return x.Parameters
	}
	return nil
}

// A named parameter value
type Parameter struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The name of this parameter
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// The value of this parameter
	Value string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *Parameter) Reset() {
	*x = Parameter{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mikebway_fulfillment_task_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Parameter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Parameter) ProtoMessage() {}

func (x *Parameter) ProtoReflect() protoreflect.Message {
	mi := &file_mikebway_fulfillment_task_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Parameter.ProtoReflect.Descriptor instead.
func (*Parameter) Descriptor() ([]byte, []int) {
	return file_mikebway_fulfillment_task_proto_rawDescGZIP(), []int{1}
}

func (x *Parameter) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Parameter) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

var File_mikebway_fulfillment_task_proto protoreflect.FileDescriptor

var file_mikebway_fulfillment_task_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x6d, 0x69, 0x6b, 0x65, 0x62, 0x77, 0x61, 0x79, 0x2f, 0x66, 0x75, 0x6c, 0x66, 0x69,
	0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x2f, 0x74, 0x61, 0x73, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x14, 0x6d, 0x69, 0x6b, 0x65, 0x62, 0x77, 0x61, 0x79, 0x2e, 0x66, 0x75, 0x6c, 0x66,
	0x69, 0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x74, 0x79, 0x70, 0x65, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0xbb, 0x03, 0x0a, 0x04, 0x54, 0x61, 0x73, 0x6b, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x43, 0x0a,
	0x0f, 0x73, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x69, 0x6d, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x52, 0x0e, 0x73, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x54, 0x69,
	0x6d, 0x65, 0x12, 0x43, 0x0a, 0x0f, 0x63, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e,
	0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0e, 0x63, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74,
	0x69, 0x6f, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x6f, 0x72, 0x64, 0x65, 0x72,
	0x5f, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6f, 0x72, 0x64, 0x65, 0x72,
	0x49, 0x64, 0x12, 0x22, 0x0a, 0x0d, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x5f, 0x69, 0x74, 0x65, 0x6d,
	0x5f, 0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x6f, 0x72, 0x64, 0x65, 0x72,
	0x49, 0x74, 0x65, 0x6d, 0x49, 0x64, 0x12, 0x21, 0x0a, 0x0c, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63,
	0x74, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x70, 0x72,
	0x6f, 0x64, 0x75, 0x63, 0x74, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x74, 0x61, 0x73,
	0x6b, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x74, 0x61,
	0x73, 0x6b, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x38, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x18, 0x08, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x20, 0x2e, 0x6d, 0x69, 0x6b, 0x65, 0x62, 0x77, 0x61,
	0x79, 0x2e, 0x66, 0x75, 0x6c, 0x66, 0x69, 0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x54, 0x61,
	0x73, 0x6b, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x12, 0x1f, 0x0a, 0x0b, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18,
	0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x72, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x43, 0x6f, 0x64,
	0x65, 0x12, 0x3f, 0x0a, 0x0a, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x18,
	0x0a, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x6d, 0x69, 0x6b, 0x65, 0x62, 0x77, 0x61, 0x79,
	0x2e, 0x66, 0x75, 0x6c, 0x66, 0x69, 0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x50, 0x61, 0x72,
	0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x52, 0x0a, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65,
	0x72, 0x73, 0x22, 0x35, 0x0a, 0x09, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x2a, 0xbf, 0x01, 0x0a, 0x0a, 0x54, 0x61,
	0x73, 0x6b, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x0d, 0x0a, 0x09, 0x55, 0x4e, 0x44, 0x45,
	0x46, 0x49, 0x4e, 0x45, 0x44, 0x10, 0x00, 0x12, 0x10, 0x0a, 0x0c, 0x57, 0x41, 0x49, 0x54, 0x49,
	0x4e, 0x47, 0x5f, 0x54, 0x41, 0x53, 0x4b, 0x10, 0x01, 0x12, 0x14, 0x0a, 0x10, 0x57, 0x41, 0x49,
	0x54, 0x49, 0x4e, 0x47, 0x5f, 0x43, 0x55, 0x53, 0x54, 0x4f, 0x4d, 0x45, 0x52, 0x10, 0x02, 0x12,
	0x13, 0x0a, 0x0f, 0x57, 0x41, 0x49, 0x54, 0x49, 0x4e, 0x47, 0x5f, 0x50, 0x41, 0x59, 0x4d, 0x45,
	0x4e, 0x54, 0x10, 0x03, 0x12, 0x0e, 0x0a, 0x0a, 0x57, 0x41, 0x49, 0x54, 0x49, 0x4e, 0x47, 0x5f,
	0x43, 0x53, 0x10, 0x04, 0x12, 0x13, 0x0a, 0x0f, 0x57, 0x41, 0x49, 0x54, 0x49, 0x4e, 0x47, 0x5f,
	0x53, 0x45, 0x52, 0x56, 0x49, 0x43, 0x45, 0x10, 0x05, 0x12, 0x17, 0x0a, 0x13, 0x57, 0x41, 0x49,
	0x54, 0x49, 0x4e, 0x47, 0x5f, 0x54, 0x48, 0x49, 0x52, 0x44, 0x5f, 0x50, 0x41, 0x52, 0x54, 0x59,
	0x10, 0x06, 0x12, 0x0a, 0x0a, 0x06, 0x50, 0x41, 0x55, 0x53, 0x45, 0x44, 0x10, 0x62, 0x12, 0x0c,
	0x0a, 0x08, 0x43, 0x41, 0x4e, 0x43, 0x45, 0x4c, 0x45, 0x44, 0x10, 0x63, 0x12, 0x0d, 0x0a, 0x09,
	0x43, 0x4f, 0x4d, 0x50, 0x4c, 0x45, 0x54, 0x45, 0x44, 0x10, 0x64, 0x42, 0x32, 0x5a, 0x30, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x69, 0x6b, 0x65, 0x62, 0x77,
	0x61, 0x79, 0x2f, 0x70, 0x6f, 0x63, 0x2d, 0x67, 0x63, 0x70, 0x2d, 0x65, 0x63, 0x6f, 0x6d, 0x6d,
	0x2f, 0x70, 0x62, 0x2f, 0x66, 0x75, 0x6c, 0x66, 0x69, 0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mikebway_fulfillment_task_proto_rawDescOnce sync.Once
	file_mikebway_fulfillment_task_proto_rawDescData = file_mikebway_fulfillment_task_proto_rawDesc
)

func file_mikebway_fulfillment_task_proto_rawDescGZIP() []byte {
	file_mikebway_fulfillment_task_proto_rawDescOnce.Do(func() {
		file_mikebway_fulfillment_task_proto_rawDescData = protoimpl.X.CompressGZIP(file_mikebway_fulfillment_task_proto_rawDescData)
	})
	return file_mikebway_fulfillment_task_proto_rawDescData
}

var file_mikebway_fulfillment_task_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_mikebway_fulfillment_task_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_mikebway_fulfillment_task_proto_goTypes = []interface{}{
	(TaskStatus)(0),               // 0: mikebway.fulfillment.TaskStatus
	(*Task)(nil),                  // 1: mikebway.fulfillment.Task
	(*Parameter)(nil),             // 2: mikebway.fulfillment.Parameter
	(*timestamppb.Timestamp)(nil), // 3: google.protobuf.Timestamp
}
var file_mikebway_fulfillment_task_proto_depIdxs = []int32{
	3, // 0: mikebway.fulfillment.Task.submission_time:type_name -> google.protobuf.Timestamp
	3, // 1: mikebway.fulfillment.Task.completion_time:type_name -> google.protobuf.Timestamp
	0, // 2: mikebway.fulfillment.Task.status:type_name -> mikebway.fulfillment.TaskStatus
	2, // 3: mikebway.fulfillment.Task.parameters:type_name -> mikebway.fulfillment.Parameter
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_mikebway_fulfillment_task_proto_init() }
func file_mikebway_fulfillment_task_proto_init() {
	if File_mikebway_fulfillment_task_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_mikebway_fulfillment_task_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Task); i {
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
		file_mikebway_fulfillment_task_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Parameter); i {
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
			RawDescriptor: file_mikebway_fulfillment_task_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mikebway_fulfillment_task_proto_goTypes,
		DependencyIndexes: file_mikebway_fulfillment_task_proto_depIdxs,
		EnumInfos:         file_mikebway_fulfillment_task_proto_enumTypes,
		MessageInfos:      file_mikebway_fulfillment_task_proto_msgTypes,
	}.Build()
	File_mikebway_fulfillment_task_proto = out.File
	file_mikebway_fulfillment_task_proto_rawDesc = nil
	file_mikebway_fulfillment_task_proto_goTypes = nil
	file_mikebway_fulfillment_task_proto_depIdxs = nil
}