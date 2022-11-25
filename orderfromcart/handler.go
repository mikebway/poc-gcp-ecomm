// Package orderfromcart implements a Google Cloud Function to receive a shopping cart description via
// a Cloud Task queue. The handler translates the cart into an order and stores it as an order document
// under Firestore.
package orderfromcart

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	orders "github.com/mikebway/poc-gcp-ecomm/order/schema"
	pb "github.com/mikebway/poc-gcp-ecomm/pb/cart"
	"github.com/mikebway/poc-gcp-ecomm/types"
	_ "github.com/mikebway/poc-gcp-ecomm/types"
	"go.uber.org/zap"
	"io"
	"net/http"
)

// init is the static initializer used to configure our local and global static variables.
func init() {
	// Initialize our Zap logger
	serviceLogger, _ := zap.NewProduction()
	zap.ReplaceGlobals(serviceLogger)
}

// OrderFromCart is Cloud Function the entry point. The payload of the HTTP request is a checked out shopping cart
// expressed as a JSON string.
func OrderFromCart(w http.ResponseWriter, r *http.Request) {

	// Have our big brother sibling do all the real work while we just handle the HTTP interfacing here
	err := doOrderFromCart(r.Body)
	if err != nil {

		// Dang - log the error and return it to the caller as well
		zap.L().Error("failed to store order", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// doOrderFromCart does all the heavy lifting for OrderFromCart. It is implemented as a separate
// function to isolate the message processing from the transport interface.
func doOrderFromCart(reader io.Reader) error {

	// Translate the base64 encoded body of the request as a binary byte slice
	pbBytes, err := Base64ReaderToBytes(reader)
	if err != nil {
		return err
	}

	// Unmarshall the protobuf binary message into a shopping cart structure
	cart, err := unmarshalShoppingCart(pbBytes)
	if err != nil {
		return err
	}

	// Convert the shopping cart structure to an order
	order, err := ConvertCartToOrder(cart)
	if err != nil {
		return err
	}

	// TODO: Logging the entire order is temporary until we have implemented order persistence in Firestore
	jsonBytes, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("unable to marshal order as JSON: %w", err)
	}
	zap.L().Info("order content", zap.ByteString("order", jsonBytes))

	// Log the ID of the order we have just ingested
	zap.L().Info("order received", zap.String("id", order.Id))

	// All done, very happy
	return nil
}

// Base64ReaderToBytes reads a base64 encoded byte stream and returns it as a byte slice.
func Base64ReaderToBytes(reader io.Reader) ([]byte, error) {

	// Read all the bytes into memory
	base64Bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("unable to read base64 bytes: %w", err)
	}

	binaryBytes := make([]byte, base64.StdEncoding.DecodedLen(len(base64Bytes)))
	n, err := base64.StdEncoding.Decode(binaryBytes, base64Bytes)
	if err != nil {
		return nil, fmt.Errorf("unable to decode base64 bytes: %w", err)
	}

	// Truncate the slice to the number of bytes in the decoded result and return that
	binaryBytes = binaryBytes[:n]
	return binaryBytes, nil
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
	items := make([]*orders.OrderItem, len(cart.CartItems))
	for i, pbItem := range cart.CartItems {
		items[i] = OrderItemItemFromShoppingCartPB(pbItem)
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
