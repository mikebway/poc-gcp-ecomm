// Package schema defines fulfillment task document structures as they might be stored in a Google Firestore
// or represented in JSON
package schema

import (
	pbfulfillment "github.com/mikebway/poc-gcp-ecomm/pb/fulfillment"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

const (
	// TaskCollection names the firestore collection under which all of our documents are stored
	TaskCollection = "tasks"
)

// Task defines a fulfilment task that is being tracked by the Fulfillment Orchestration Service. Fulfillment tasks
// map to order items in a many-to-one relationship, i.e. a single item in a customer order may map to multiple
// fulfillment tasks.
//
// For example, an order might include two items such as a custom sofa and two end tables (one order line item with a
// quantity of two). Fulfillment tasks for the sofa could include: order cover material, manufacture, ship, and install.
type Task struct {
	// Id is a UUID ID in hexadecimal string form - a unique ID for this task.
	// This will be set when the task is created in response to a new order being
	// received by the fulfilment service.
	Id string `firestore:"id" json:"id"`

	// SubmissionTime is the time at which task was submitted to the Fulfillment Orchestration Service.
	SubmissionTime time.Time `firestore:"submissionTime" json:"submissionTime"`

	// CompletionTime is the time at which task was marked as completed
	CompletionTime time.Time `firestore:"completionTime,omitempty" json:"completionTime"`

	// OrderId relates this task to the order containing the item that requires the task to be performed.
	// OrderId is a UUID ID in hexadecimal string form - a unique ID for the order.
	OrderId string `firestore:"orderId" json:"orderId"`

	// OrderItemId relates this task to the order item that requires the task to be performed.
	// OrderItemId is a UUID ID in hexadecimal string form - a unique ID for the order item.
	OrderItemId string `firestore:"orderItemId" json:"orderItemId"`

	// ProductCode is the equivalent of a SKU code identifying the type of product or service that the OrderItemId
	// is for.
	ProductCode string `firestore:"productCode" json:"productCode"`

	// TaskCode, in combination with the ProductCode, identifies the type of activity to be performed and the data,
	// if any, to be collected.
	TaskCode string `firestore:"taskCode" json:"taskCode"`

	// Status identifies the status of the task, i.e. whether we are waiting for customer input, waiting for a
	// response from a third party service, or that the task has bee completed or failed.
	Status TaskStatus `firestore:"status" json:"status"`

	// ReasonCode expands on the status, providing a key that can be used to look up a localized explanation for why
	// the status is WAITING_CUSTOMER, PAUSED, CANCELED, etc. Hopefully, people will choose reason codes that convey
	// some meaning by themselves saving engineers with only the raw data from having to translate the value into
	// something more intelligible.
	//
	// It is possible that the ReasonCode might need to be interpreted in context with the combination of the task_code
	// and product_code.
	ReasonCode string `firestore:"reasonCode" json:"reasonCode"`

	// Parameters is a map of zero to many named value parameters that might be required to complete the task. For example,
	// extending the custom sofa analogy, the parameters might define the model and cover fabric for the sofa.
	Parameters []*Parameter `firestore:"parameters" json:"parameters"`
}

// StoreRefPath returns the string representation of the document reference path for this Task.
func (t *Task) StoreRefPath() string {
	return TaskCollection + "/" + t.Id
}

// TaskStatus is an integer enumeration of the possible task states
type TaskStatus int32

const (
	// UNDEFINED_STATUS should not be seen - indicates that the status has not be set
	UNDEFINED_STATUS TaskStatus = 0

	// WAITING_TASK signals that the task is waiting on another task to complete.
	WAITING_TASK = 1

	// WAITING_CUSTOMER signals that the task is waiting for customer input.
	WAITING_CUSTOMER = 2

	// WAITING_PAYMENT signals that the task is waiting for a customer payment to be confirmed.
	WAITING_PAYMENT = 3

	// WAITING_CS signals that the task is waiting on customer service.
	WAITING_CS = 4

	// WAITING_SERVICE signals that the task is waiting on an internal customer service.
	WAITING_SERVICE = 5

	// WAITING_THIRD_PARTY signals that the task is waiting on a third party service.
	WAITING_THIRD_PARTY = 6

	// PAUSED signals that the task has been paused; see ReasonCode
	PAUSED = 98

	// CANCELED signals that the task has been canceled; see ReasonCode
	CANCELED = 99

	// COMPLETED signals that the task has been completed
	COMPLETED = 100
)

// Parameter represents a single named string value. allowing the parameter to have different types
// just proved two painful in the conversions between internal versions and, most especially, the Firestore
// clients protocol buffer API over which we have no control.
type Parameter struct {

	// Name is the parameter name
	Name string `firestore:"name" json:"name"`

	// Value is one of a limited number of value type
	Value string `firestore:"value" json:"value"`
}

// AsPBTask returns the protocol buffer representation of this task.
func (t *Task) AsPBTask() *pbfulfillment.Task {

	// Submission time should be set, but we will play it safe just the same
	var pbSubmissionTime *timestamppb.Timestamp
	if !t.SubmissionTime.IsZero() {
		pbSubmissionTime = timestamppb.New(t.SubmissionTime)
	}

	// CompletionTime time may not be set, so we must play it safe
	var pbCompletionTime *timestamppb.Timestamp
	if !t.CompletionTime.IsZero() {
		pbCompletionTime = timestamppb.New(t.CompletionTime)
	}

	// Return a populated protocol buffer version of the cart
	return &pbfulfillment.Task{
		Id:             t.Id,
		SubmissionTime: pbSubmissionTime,
		CompletionTime: pbCompletionTime,
		OrderId:        t.OrderId,
		OrderItemId:    t.OrderItemId,
		ProductCode:    t.ProductCode,
		TaskCode:       t.TaskCode,
		Status:         pbfulfillment.TaskStatus(t.Status),
		ReasonCode:     t.ReasonCode,
		Parameters:     t.asPBParameters(),
	}
}

// asPBParameters converts our internal parameter set representation to the protocol buffer equivalent.
func (t *Task) asPBParameters() []*pbfulfillment.Parameter {

	// If we have no parameters, return nil for lowest cost
	if len(t.Parameters) == 0 {
		return nil
	}

	// Build our response here
	pbParams := make([]*pbfulfillment.Parameter, len(t.Parameters))

	// Loop through all our internal parameters adding them to the result
	for i, param := range t.Parameters {

		// Set the protocol buffer version of the parameter into the protocol buffer slice that we will return
		pbParams[i] = &pbfulfillment.Parameter{Name: param.Name, Value: param.Value}
	}

	// Return what we built, perhaps noting at all
	return pbParams
}
