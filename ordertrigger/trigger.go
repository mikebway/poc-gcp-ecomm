// Package ordertrigger handles Firestore trigger invocations when order documents are updated.
package ordertrigger

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/mikebway/poc-gcp-ecomm/order/orderapi"
	pborder "github.com/mikebway/poc-gcp-ecomm/pb/order"
	"go.uber.org/zap"
)

var (
	// TopicProjectId is a variable so that unit tests can override it to ensures that test requests are not
	// routed to the live project! See https://firebase.google.com/docs/emulator-suite/connect_firestore
	TopicProjectId string

	// TopicId a variable so that unit tests can override it to force errors to occur
	TopicId string

	// orderClient is lazy-loaded and allows us to retrieve complete orders from Firestore. Unit tests
	// can substitute an alternative instance here to force errors.
	orderClient OrderServiceClient

	// pubSubClient is an instance of a wrapper interface for our Pub/Sub client that allows us to inject errors
	// when unit testing
	pubSubClient PubSubClient
)

// FirestoreEvent is the payload of a Firestore event.
type FirestoreEvent struct {
	OldValue   FirestoreValue `json:"oldValue"`
	Value      FirestoreValue `json:"value"`
	UpdateMask struct {
		FieldPaths []string `json:"fieldPaths"`
	} `json:"updateMask"`
}

// FirestoreValue holds Firestore document fields.
type FirestoreValue struct {
	CreateTime time.Time      `json:"createTime"`
	Fields     FirestoreOrder `json:"fields"`
	Name       string         `json:"name"`
	UpdateTime time.Time      `json:"updateTime"`
}

// FirestoreOrder describes the document fields that we need to know about as they will
// be found in the event data (not as we would prefer them, in the structure that we
// submitted to the Firestore API to populate the document in the first place :-(
type FirestoreOrder struct {
	Id     StringValue  `json:"id"`
	Status IntegerValue `json:"status"`
}
type StringValue struct {
	StringValue string `json:"stringValue"`
}
type IntegerValue struct {
	IntegerValue string `json:"integerValue"`
}

// OrderServiceClient is a wrapper for the order service that supports lazy loading of the service and unit test
// error generation.
type OrderServiceClient interface {
	GetOrder(ctx context.Context, orderId string) (*pborder.Order, error)
}

// PubSubClient is a wrapper for our Google client that supports lazy loading of the client and unit test
// error generation.
type PubSubClient interface {
	Publish(ctx context.Context, order *pborder.Order) error
}

// init is the static initializer used to configure our local and global static variables.
func init() {

	// Default the project ID and topic ID to be used for live Pub/Sub topic connections
	TopicProjectId = "poc-gcp-ecomm"
	TopicId = "ecomm-order"

	// Initialize our Zap logger
	serviceLogger, _ := zap.NewProduction()
	zap.ReplaceGlobals(serviceLogger)

	// Instantiate our two clients
	orderClient = &OrderServiceClientImpl{}
	pubSubClient = &PubSubClientImpl{}
}

// OrderTrigger receives a document update Firestore trigger event. The function is deployed with a trigger
// configuration (see Makefile) that will notify the handler of all updates to the root document of an Order.
func OrderTrigger(ctx context.Context, e FirestoreEvent) error {

	// Have our big brother sibling do all the real work while we just handle the trigger interfacing and
	// error logging here
	err := doCartTrigger(ctx, e)
	if err != nil {

		// Dang - log the error and return it to the caller as well
		zap.L().Error("failed to process order update trigger", zap.Error(err))
		return err
	}

	// All is well
	return nil
}

// doCartTrigger does all the heavy lifting for CartTrigger. It is implemented as a separate
// function to isolate the message processing from the trigger interface.
func doCartTrigger(ctx context.Context, e FirestoreEvent) error {

	// We need to log multiple times so just get the logger and be done with that
	logger := zap.L()

	// TODO: Google and AWS have this in common: they fail to make their CDC event stream contents
	//       compatible with or easily convertible to their database API models. There is no easy
	//       way to populate an Order structure from the FirestoreEvent/FirestoreValue structures!

	// There should be a way to unmarshall this FirestoreEvent data to an Order structures but there
	// is not. Fortunately, we need little information from the new FirestoreValue structure to determine
	// how we should respond.

	// Pick up the ID of the order in question
	newFields := e.Value.Fields
	orderId := newFields.Id.StringValue

	// At this point we know that we have a order that needs to be published
	logger.Info("processing order", zap.String("orderId", orderId))

	// Retrieve the full order from Firestore
	order, err := orderClient.GetOrder(ctx, orderId)
	if err != nil {
		return fmt.Errorf("unable to retrieve order from firestore: %s - %w", orderId, err)
	}

	// Publish the order to our target topic
	err = pubSubClient.Publish(ctx, order)
	if err != nil {
		return fmt.Errorf("pubsub publish failed: %s - %w", orderId, err)
	}

	// ... and that is all she wrote!
	logger.Info("published order", zap.String("orderId", orderId))
	return nil
}

// OrderServiceClientImpl is the default implementation of the OrderServiceClient interface.
// error generation.
type OrderServiceClientImpl struct {
	OrderServiceClient

	orderService *orderapi.OrderService
}

// GetOrder loads a fully populated order from Firestore.
func (c *OrderServiceClientImpl) GetOrder(ctx context.Context, orderId string) (*pborder.Order, error) {

	// Lazy-load the underlying Pub/Sub client that we wrap
	err := c.lazyLoad()
	if err != nil {
		return nil, err
	}

	// Attempt to fetch the requested order
	svcResponse, err := c.orderService.GetOrderByID(ctx, &pborder.GetOrderByIDRequest{OrderId: orderId})
	if err != nil {
		return nil, err
	}

	// It's all good - return the order
	return svcResponse.Order, nil
}

// getClient lazy-loads our underlying orderapi.OrderService.
func (c *OrderServiceClientImpl) lazyLoad() error {

	// In the normal case, we return quickly because the service has been cached before
	if c.orderService != nil {
		return nil
	}

	// Establish our order service
	var err error
	c.orderService, err = orderapi.NewOrderService()

	// Happy or not, we are done
	return err
}

// PubSubClientImpl is the default implementation of the PubSubClient interface.
type PubSubClientImpl struct {
	PubSubClient

	topic *pubsub.Topic
}

// Publish submits a binary message to our configured Pub/Sub topic.
func (c *PubSubClientImpl) Publish(ctx context.Context, order *pborder.Order) error {

	// Lazy-load the underlying Pub/Sub client that we wrap
	err := c.lazyLoad(ctx)
	if err != nil {
		return err
	}

	// Marshal the protobuf-ready order structure we just retrieved into base64 encoded binary
	data, err := proto.Marshal(order)
	if err != nil {
		return fmt.Errorf("unable to marshal order into protobuf binary: %w", err)
	}

	// Publish the data to our target topic
	result := c.topic.Publish(ctx, &pubsub.Message{
		Data: data,
	})
	_, err = result.Get(ctx)
	return err
}

// getClient lazy-loads our underlying Pub/Sub client.
func (c *PubSubClientImpl) lazyLoad(ctx context.Context) error {

	// In the normal case, we return quickly because the client has been cached before
	if c.topic != nil {
		return nil
	}

	// Get a new Pub/Sub client
	client, err := pubsub.NewClient(ctx, TopicProjectId)
	if err != nil {
		return err
	}

	// Instantiate a topic with that client and configure it to send immediately (max batch size = 1)
	c.topic = client.Topic(TopicId)
	c.topic.PublishSettings.CountThreshold = 1

	// And we are all happy and done
	return nil
}
