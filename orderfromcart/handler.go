// Package orderfromcart implements a Google Cloud Function to receive a shopping cart description via
// a Pub/Sub topic. The handler translates the cart into an order and stores it as an order document
// under Firestore.
package orderfromcart

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/golang/protobuf/proto"
	orders "github.com/mikebway/poc-gcp-ecomm/order/schema"
	"github.com/mikebway/poc-gcp-ecomm/order/service"
	pb "github.com/mikebway/poc-gcp-ecomm/pb/cart"
	"github.com/mikebway/poc-gcp-ecomm/types"
	_ "github.com/mikebway/poc-gcp-ecomm/types"
	"go.uber.org/zap"
	"google.golang.org/api/pubsub/v1"
)

var (
	// lazyOrderService is the lazy-loaded order service implementation that we use to save orders to Firestore
	lazyOrderService *service.OrderService
)

// init is the static initializer used to configure our local and global static variables.
func init() {
	// Initialize our Zap logger
	serviceLogger, _ := zap.NewProduction()
	zap.ReplaceGlobals(serviceLogger)
}

// pushRequest represents the payload of a Pub/Sub push message.
type pushRequest struct {
	Message      pubsub.PubsubMessage `json:"message"`
	Subscription string               `json:"subscription,omitempty"`
}

// OrderFromCart is Cloud Function entry point. The payload of the HTTP request is a checked out shopping cart
// expressed as a base64 encoded Protocol Buffer message wrapped in a JSON envelop.
//
// See https://cloud.google.com/pubsub/docs/push for documentation of the request body JSON content.
func OrderFromCart(w http.ResponseWriter, r *http.Request) {

	// Have our big brother sibling do all the real work while we just handle the HTTP interfacing here
	status, err := doOrderFromCart(r.Context(), r.Body)
	if err != nil {

		// Dang - log the error and return it to the caller as well
		zap.L().Error("failed to store order", zap.Error(err))
		http.Error(w, err.Error(), status)
	}

	// Return the successful status code
	w.WriteHeader(status)
}

// doOrderFromCart does all the heavy lifting for OrderFromCart. It is implemented as a separate
// function to isolate the message processing from the transport interface.
//
// An HTTP status code is always returned, this should be set in the response regardless of whether
// an error is also returned.
//
// See https://cloud.google.com/pubsub/docs/push for documentation of the reader JSON content.
func doOrderFromCart(ctx context.Context, reader io.Reader) (int, error) {

	// Lazy load the order service that we wil use to write the order to Firestore
	svc, err := getOrderService()
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

	// Unmarshall the protobuf binary message into a shopping cart structure
	cart, err := unmarshalShoppingCart(pbBytes)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Convert the shopping cart structure to an order
	order, err := ConvertCartToOrder(cart)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Save the order to Firestore
	err = svc.SaveOrder(ctx, order)
	if err != nil {
		zap.L().Error("failed to save cart as order", zap.String("cartId", cart.Id), zap.Error(err))
		return http.StatusInternalServerError, err
	}

	// Log the ID of the order we have just ingested
	zap.L().Info("order received", zap.String("id", order.Id))

	// All done, very happy
	return http.StatusOK, nil
}

// getOrderService lazy loads the order service that we use to write orders to Firestore
func getOrderService() (*service.OrderService, error) {

	// if we already have the service in hand, return it fast
	if lazyOrderService != nil {
		return lazyOrderService, nil
	}

	// Try to load the service and cache it for posterity
	var err error
	lazyOrderService, err = service.NewOrderService()
	return lazyOrderService, err
}

// unmarshalShoppingCart unpacks the provided binary protobuf message into a shopping cart structure.
func unmarshalShoppingCart(message []byte) (*pb.ShoppingCart, error) {

	// Unmarshal the protobuf message bytes if we can
	cart := &pb.ShoppingCart{}
	err := proto.Unmarshal(message, cart)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal shopping cart protobuf message: %w", err)
	}

	// Alright then, that was easier than I feared
	return cart, nil
}

// ConvertCartToOrder clones the data of a shopping cart into an order structure
func ConvertCartToOrder(cart *pb.ShoppingCart) (*orders.Order, error) {

	// TODO: Validate the order before recording it??? At least that it has ID values so that it can be found,
	//       If an invalid order is presented from the shopping cart we should perhaps record it
	//       in Firestore and return it in API requests regardless as a record of whatever came out of
	//       an apparently submitted cart. We could at least log it as an error and maybe flag it
	//       in Firestore and not fulfill it???

	// At this stage, the only field that we absolutely must have is the cart ID because we are going
	// to use that as the order ID - the order and the cart are that closely aligned that they share
	// the same ID.
	if len(cart.Id) == 0 {
		return nil, errors.New("cart does not have an ID")
	}

	// Build the order here
	order := &orders.Order{
		Id:             cart.Id,
		SubmissionTime: cart.ClosedTime.AsTime(),
	}

	// Only convert the delivery address if that has been defined in the cart
	var address *types.PostalAddress
	if cart.DeliveryAddress != nil {
		address = types.PostalAddressFromPB(cart.DeliveryAddress)
	}
	order.DeliveryAddress = address

	// Only convert the shopper if they have been defined in the cart
	// That should always be the case but we will play defensively.
	var shopper *types.Person
	if cart.Shopper != nil {
		shopper = types.PersonFromPB(cart.Shopper)
	}
	order.OrderedBy = shopper

	// Finally, add the order items (if any - there really should be)
	order.OrderItems = make([]*orders.OrderItem, len(cart.CartItems))
	for i, pbItem := range cart.CartItems {
		order.OrderItems[i] = OrderItemItemFromShoppingCartPB(pbItem)
	}

	// All done, return the fruit of our labor
	return order, nil
}

// OrderItemItemFromShoppingCartPB is a factory method that returns a ShoppingCartItem representation
// derived from its protocol buffer equivalent.
func OrderItemItemFromShoppingCartPB(pbItem *pb.CartItem) *orders.OrderItem {
	return &orders.OrderItem{
		Id:          pbItem.Id,
		ProductCode: pbItem.ProductCode,
		Quantity:    pbItem.Quantity,
		UnitPrice:   types.MoneyFromPB(pbItem.UnitPrice),
	}
}
