// Package carttrigger handles Firestore trigger invocations when shopping cart documents are updated.
//
// The handler is not invoked for the addition of cart items or delivery addresses, nor for creation of carts,
// only for updates to the cart root document. This will almost invariably be due to the cart either being
// submitted or abandoned.
package carttrigger

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/mikebway/poc-gcp-ecomm/cart/schema"
	"github.com/mikebway/poc-gcp-ecomm/cart/service"
	pbcart "github.com/mikebway/poc-gcp-ecomm/pb/cart"
	"go.uber.org/zap"
	"strconv"
	"time"
)

const (
	// TopicId identifies the Pub/Sub topic in the TopicProjectId project that we publish shopping carts to.
	TopicId = "shopping-cart"
)

var (
	// TopicProjectId is a variable so that unit tests can override it to ensures that test requests are not
	// routed to the live project! See https://firebase.google.com/docs/emulator-suite/connect_firestore
	TopicProjectId string

	// cartClient is lazy-loaded and allows us to retrieve complete shopping carts from Firestore. Unit tests
	// can substitute an alternative instance here to force errors.
	cartClient CartServiceClient

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
	CreateTime time.Time     `json:"createTime"`
	Fields     FirestoreCart `json:"fields"`
	Name       string        `json:"name"`
	UpdateTime time.Time     `json:"updateTime"`
}

// FirestoreCart describes the document fields that we need to know about as they will
// be found in the event data (not as we would prefer them, in the structure that we
// submitted to the Firestore API to populate the document in the first place :-(
type FirestoreCart struct {
	Id     StringValue  `json:"id"`
	Status IntegerValue `json:"status"`
}
type StringValue struct {
	StringValue string `json:"stringValue"`
}
type IntegerValue struct {
	IntegerValue string `json:"integerValue"`
}

// CartServiceClient is a wrapper for the cart service that supports lazy loading of the service and unit test
// error generation.
type CartServiceClient interface {
	GetCart(ctx context.Context, cartId string) (*pbcart.ShoppingCart, error)
}

// PubSubClient is a wrapper for our Google client that supports lazy loading of the client and unit test
// error generation.
type PubSubClient interface {
	Publish(ctx context.Context, cart *pbcart.ShoppingCart) error
}

// init is the static initializer used to configure our local and global static variables.
func init() {

	// Default the project ID to be used for live Pub/Sub topic connections
	TopicProjectId = "poc-gcp-ecomm"

	// Initialize our Zap logger
	serviceLogger, _ := zap.NewProduction()
	zap.ReplaceGlobals(serviceLogger)

	// Instantiate our two clients
	cartClient = &CartServiceClientImpl{}
	pubSubClient = &PubSubClientImpl{}
}

// UpdateTrigger receives a document update Firestore trigger event. The function is deployed with a trigger
// configuration (see Makefile) that will notify the handler of all updates to the root document of a Shopping Cart.
func UpdateTrigger(ctx context.Context, e FirestoreEvent) error {

	// Have our big brother sibling do all the real work while we just handle the trigger interfacing and
	// error logging here
	err := doUpdateTrigger(ctx, e)
	if err != nil {

		// Dang - log the error and return it to the caller as well
		zap.L().Error("failed to process cart update trigger", zap.Error(err))
		return err
	}

	// All is well
	return nil
}

// doUpdateTrigger does all the heavy lifting for UpdateTrigger. It is implemented as a separate
// function to isolate the message processing from the trigger interface.
func doUpdateTrigger(ctx context.Context, e FirestoreEvent) error {

	// We need to log multiple times so just get the logger and be done with that
	logger := zap.L()

	// TODO: Google and AWS have this in common: they fail to make their CDC event stream contents
	//       compatible with or easily convertible to their database API models. There is no easy
	//       way to populate a ShoppingCart structure from the FirestoreEvent/FirestoreValue structures!

	// There should be a way to unmarshall this FirestoreEvent data to a ShoppingCart structures but there
	// is not. Fortunately, we need little information from the new FirestoreValue structure to determine
	// how we should respond.

	// Pick up the ID of the cart in question
	newFields := e.Value.Fields
	cartId := newFields.Id.StringValue

	// Is this event for a cart being checked out, i.e. ready to ve submitted as an order?
	if newFields.Status.IntegerValue != strconv.FormatInt(int64(schema.CsCheckedOut), 10) {

		// This cart has not been checked out, we should ignore it
		logger.Info("ignoring cart update", zap.String("cartId", cartId), zap.String("status", newFields.Status.IntegerValue))
		return nil
	}

	// At this point we know that we have a checked out cart that needs to be published
	logger.Info("processing checked out cart", zap.String("cartId", cartId))

	// Retrieve the full cart from Firestore
	cart, err := cartClient.GetCart(ctx, cartId)
	if err != nil {
		return fmt.Errorf("unable to retrieve cart from firestore: %s - %w", cartId, err)
	}

	// Publish the cart to our target topic
	err = pubSubClient.Publish(ctx, cart)
	if err != nil {
		return fmt.Errorf("pubsub publish failed: %s - %w", cartId, err)
	}

	// ... and that is all she wrote!
	logger.Info("published checked out cart", zap.String("cartId", cartId))
	return nil
}

// CartServiceClientImpl is the default implementation of the CartServiceClient interface.
// error generation.
type CartServiceClientImpl struct {
	CartServiceClient

	cartService *service.CartService
}

// GetCart loads a fully populated shopping cart from Firestore.
func (c *CartServiceClientImpl) GetCart(ctx context.Context, cartId string) (*pbcart.ShoppingCart, error) {

	// Lazy-load the underlying Pub/Sub client that we wrap
	err := c.lazyLoad()
	if err != nil {
		return nil, err
	}

	// Attempt to fetch the requested cart
	svcResponse, err := c.cartService.GetShoppingCartByID(ctx, &pbcart.GetShoppingCartByIDRequest{CartId: cartId})
	if err != nil {
		return nil, err
	}

	// It's all good - return the cart
	return svcResponse.Cart, nil
}

// getClient lazy-loads our underlying service.CartService.
func (c *CartServiceClientImpl) lazyLoad() error {

	// In the normal case, we return quickly because the service has been cached before
	if c.cartService != nil {
		return nil
	}

	// Establish our cart service
	var err error
	c.cartService, err = service.NewCartService()

	// Happy or not, we are done
	return err
}

// PubSubClientImpl is the default implementation of the PubSubClient interface.
type PubSubClientImpl struct {
	PubSubClient

	topic *pubsub.Topic
}

// Publish submits a binary message to our configured Pub/Sub topic.
func (c *PubSubClientImpl) Publish(ctx context.Context, cart *pbcart.ShoppingCart) error {

	// Lazy-load the underlying Pub/Sub client that we wrap
	err := c.lazyLoad(ctx)
	if err != nil {
		return err
	}

	// Marshal the protobuf-ready cart structure we just retrieved into base64 encoded binary
	data, err := proto.Marshal(cart)
	if err != nil {
		return fmt.Errorf("unable to marshal cart into protobuf binary: %w", err)
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
