// Package ordertofulfill implements a Google Cloud Function to receive an order description via
// a Pub/Sub topic. The handler translates the order into a a series of fulfillment tasks descriptions
// stored it as as Task documents under Firestore.
package ordertofulfill

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/mikebway/poc-gcp-ecomm/fulfillment/schema"
	"github.com/mikebway/poc-gcp-ecomm/fulfillment/service"
	pb "github.com/mikebway/poc-gcp-ecomm/pb/order"
	_ "github.com/mikebway/poc-gcp-ecomm/types"
	"go.uber.org/zap"
	"google.golang.org/api/pubsub/v1"
	"io"
	"net/http"
)

var (
	// productTasks is a map of product codes to potentially one or more skeleton fulfillment tasks
	//
	// In a production implementation, this mapp would be loaded from Firestore when the function is
	// first instantiated. We will cheat and hard code the map in the init() function.
	productTasks map[string][]*schema.Task

	// lazyFulfillmentService is the lazy-loaded fulfillment service implementation that we use to save tasks to Firestore
	lazyFulfillmentService *service.FulfillmentService
)

// init is the static initializer used to configure our local and global static variables.
func init() {
	// Initialize our Zap logger
	serviceLogger, _ := zap.NewProduction()
	zap.ReplaceGlobals(serviceLogger)

	// Cheat and hard code the product to task map here
	// TODO: load this map from a configuration stored in Firestore
	productTasks = make(map[string][]*schema.Task)
	addProductTaskMapping(&schema.Task{ProductCode: "gold_yoyo", TaskCode: "manufacture", Status: schema.WAITING_SERVICE})
	addProductTaskMapping(&schema.Task{ProductCode: "gold_yoyo", TaskCode: "ship", Status: schema.WAITING_TASK, ReasonCode: "wait_for_manufacture"})
	addProductTaskMapping(&schema.Task{ProductCode: "plastic_yoyo", TaskCode: "upsell_to_gold", Status: schema.WAITING_CS, ReasonCode: "no_stock"})
}

// pushRequest represents the payload of a Pub/Sub push message.
type pushRequest struct {
	Message      pubsub.PubsubMessage `json:"message"`
	Subscription string               `json:"subscription,omitempty"`
}

// OrderToFulfill is Cloud Function entry point. The payload of the HTTP request is an order
// expressed as a base64 encoded Protocol Buffer message wrapped in a JSON envelop.
//
// See https://cloud.google.com/pubsub/docs/push for documentation of the request body JSON content.
func OrderToFulfill(w http.ResponseWriter, r *http.Request) {

	// Have our big brother sibling do all the real work while we just handle the HTTP interfacing here
	status, err := doOrderToFulfill(r.Context(), r.Body)
	if err != nil {

		// Dang - log the error and return it to the caller as well
		zap.L().Error("failed to store tasks", zap.Error(err))
		http.Error(w, err.Error(), status)
	}

	// Return the successful status code
	w.WriteHeader(status)
}

// doOrderToFulfill does all the heavy lifting for OrderToFulfill. It is implemented as a separate
// function to isolate the message processing from the transport interface.
//
// An HTTP status code is always returned, this should be set in the response regardless of whether
// an error is also returned.
//
// See https://cloud.google.com/pubsub/docs/push for documentation of the reader JSON content.
func doOrderToFulfill(ctx context.Context, reader io.Reader) (int, error) {

	// Lazy load the fulfillment service that we wil use to write tasks to Firestore
	svc, err := getFulfillemntService()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Unpack the JSON push request message from the request body
	var pushReq pushRequest
	if err := json.NewDecoder(reader).Decode(&pushReq); err != nil {
		return http.StatusBadRequest, fmt.Errorf("could not decode push request json body: %v", err)
	}

	// Translate the base64 encoded body of the request as a binary byte slice
	pbBytes, err := base64.StdEncoding.DecodeString(pushReq.Message.Data)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("unable to decode base64 data: %w", err)
	}

	// Unmarshall the protobuf binary message into an order structure
	order, err := unmarshalOrder(pbBytes)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Convert the order structure to a set of fulfillment tasks
	tasks := convertOrderToTasks(order)

	// Save all the tasks in a single transaction - all or nothing service.
	err = svc.SaveTasks(ctx, tasks)
	if err != nil {
		zap.L().Error("failed to save tasks for order", zap.String("orderId", order.Id), zap.Error(err))
		return http.StatusInternalServerError, err
	}

	// Everything is good
	zap.L().Error("established tasks for order", zap.String("orderId", order.Id))
	return http.StatusCreated, nil
}

// getFulfillemntService lazy loads the fulfillment service that we use to write tasks to Firestore
func getFulfillemntService() (*service.FulfillmentService, error) {

	// if we already have the service in hand, return it fast
	if lazyFulfillmentService != nil {
		return lazyFulfillmentService, nil
	}

	// Try to load the service and cache it for posterity
	var err error
	lazyFulfillmentService, err = service.NewFulfillmentService()
	return lazyFulfillmentService, err
}

// unmarshalOrder unpacks the provided binary protobuf message into a order structure.
func unmarshalOrder(message []byte) (*pb.Order, error) {

	// TODO: we should validate the order is properly populated etc

	// Unmarshal the protobuf message bytes if we can
	order := &pb.Order{}
	err := proto.Unmarshal(message, order)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal order protobuf message: %w", err)
	}

	// Alright then, that was easier than I feared
	return order, nil
}

// convertOrderToTasks translates the order items into fulfillment task structures/
func convertOrderToTasks(pbOrder *pb.Order) []*schema.Task {

	// TODO: Validate the order before converting it???

	// Build our result here
	var tasks []*schema.Task

	// Walk the set of order items, translating them into one to several fulfillment tasks each
	for _, pbItem := range pbOrder.OrderItems {

		// Lookup the fulfillment tasks that match the item product type
		itemTasks := productTasks[pbItem.ProductCode]

		// If we found any tasks for this product, add a copy of the template tasks to our task set
		for _, template := range itemTasks {

			// Make a copy of the template task
			task := *template

			// Add some context detail to it from the current order
			task.Id = uuid.NewString()
			task.OrderId = pbOrder.Id
			task.OrderItemId = pbItem.Id

			// Add the task to our total set
			tasks = append(tasks, &task)
		}
	}

	// All done, return the fruit of our labor
	return tasks
}

// addProductTaskMapping adds a single product to task mapping to our global map
func addProductTaskMapping(task *schema.Task) {

	// Fetch any existing lists of tasks we already have for the product associated with this skeleton task?
	taskList := productTasks[task.ProductCode]

	// If there is not already a task set for this product, create one now
	if taskList == nil {
		taskList = make([]*schema.Task, 0)
	}

	// Append this new task to that set and it back in the map
	productTasks[task.ProductCode] = append(taskList, task)
}
