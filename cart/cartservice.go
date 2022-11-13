package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/mikebway/poc-gcp-ecomm/cart/schema"
	pbcart "github.com/mikebway/poc-gcp-ecomm/pb/cart"
	"github.com/mikebway/poc-gcp-ecomm/types"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

const (
	ProjectId = "poc-gcp-ecomm"
)

// CartService is a structure class with methods that implements the cart.CartAPIServer gRPC API
// storing the data for the social graph in a Google Cloud Firestore Kind.
type CartService struct {
	pbcart.UnimplementedCartAPIServer

	// fsClient is the GCP Firestore client - it is thread safe and can be reused concurrently
	fsClient *firestore.Client

	// drProxy is used to allow unit tests to intercept firestore.DocumentRef function calls
	// and insert errors etc. into the responses.
	drProxy DocumentRefProxy

	// dsProxy is used to allow unit tests to intercept firestore.DocumentSnapshot function calls
	// and insert errors etc. into the responses.
	dsProxy DocumentSnapshotProxy

	// itemsGetterProxy is used to obtain an ItemsCollectionProxy for a given cart. Unit tests may
	// substitute an alternative implementation this interface in order to be able to insert errors etc.
	// into the responses of the ItemsCollectionProxy that the ItemCollectionGetterProxy returns.
	itemsGetterProxy ItemCollectionGetterProxy
}

// NewCartService is a factory method returning an instance of our social graph service.
func NewCartService() (*CartService, error) {

	// Build our service instance here with our default, direct passthrough, interception proxies
	// for firestore.DocumentRef and firestore.DocumentSnapshot function calls
	svc := &CartService{
		drProxy: &DocRefProxy{},
		dsProxy: &DocSnapProxy{},
	}

	// Obtain a firestore client and stuff that in the service instance
	ctx := context.Background()
	var err error
	if unitTestNewCartServiceError == nil {
		// Set the Firestore client if we are not unit testing an error situation.
		svc.fsClient, err = firestore.NewClient(ctx, ProjectId)

	} else {
		// We are unit testing and required to report an error
		err = unitTestNewCartServiceError
	}

	// Check that we obtained a firestore client successfully
	if err != nil {
		return nil, fmt.Errorf("could not obtain firestore client: %w", err)
	}

	// Make the firestore client available to the cart item getter proxy
	svc.itemsGetterProxy = &ItemCollGetterProxy{
		fsClient: svc.fsClient,
	}

	// All done - return the populated service instance
	return svc, nil
}

// CreateShoppingCart returns a new shopping cart structure with a unique ID, creation time, and status assigned.
func (cs *CartService) CreateShoppingCart(ctx context.Context, req *pbcart.CreateShoppingCartRequest) (*pbcart.CreateShoppingCartResponse, error) {

	// Obtain a shortcut handle on our globally configured logger
	l := zap.L()

	// TODO: Parameter validation
	// TODO: Access control

	// Create the storable cart structure with a new unique ID
	storableCart := &schema.ShoppingCart{
		Id:           uuid.New().String(),
		CreationTime: time.Now(),
		Status:       schema.CsOpen,
		Shopper:      types.PersonFromPB(req.Shopper),
	}
	l.Info("storing new cart", zap.String("CartId", storableCart.Id))

	// Store the empty new cart in the firestore
	ref := cs.fsClient.Doc(storableCart.StoreRefPath())
	_, err := cs.drProxy.Create(ref, ctx, storableCart)
	if err != nil {
		err = fmt.Errorf("failed creating new cart in Firestore: %w", err)
		l.Error(err.Error(), zap.String("CartId", storableCart.Id))
		return nil, err
	}

	// All good, log our joy and return the protocol buffer transliteration of our shiny new cart
	l.Info("new cart stored successfully", zap.String("CartId", storableCart.Id))
	return &pbcart.CreateShoppingCartResponse{
		Cart: storableCart.AsPBShoppingCart(),
	}, nil
}

// GetShoppingCartByID retrieves a cart by its UUID ID.
func (cs *CartService) GetShoppingCartByID(ctx context.Context, req *pbcart.GetShoppingCartByIDRequest) (*pbcart.GetShoppingCartByIDResponse, error) {

	// Obtain a shortcut handle on our globally configured logger
	l := zap.L()
	l.Info("retrieving cart", zap.String("CartId", req.CartId))

	// Have our internal sibling do all the hard work
	pbCart, err := cs.getShoppingCart(ctx, req.CartId)
	if err != nil {
		return nil, err
	}

	// All good, log our joy and return the protocol buffer transliteration of our retrieved cart
	l.Info("cart retrieved successfully", zap.String("CartId", req.CartId))
	return &pbcart.GetShoppingCartByIDResponse{
		Cart: pbCart,
	}, nil
}

// getShoppingCart is a shared internal function that retrieves a shopping cart and returns it in protocol
// buffer form. It is used by all the public set and get methods that include a copy of the cart in their
// response.
func (cs *CartService) getShoppingCart(ctx context.Context, cartId string) (*pbcart.ShoppingCart, error) {

	// Obtain a shortcut handle on our globally configured logger
	l := zap.L()
	l.Info("retrieving cart", zap.String("CartId", cartId))

	// Form a cart structure to receive the data from the store
	storedCart := &schema.ShoppingCart{Id: cartId}

	// TODO: wrap cart retrieval in a transaction
	// TODO: use a query to get everything is a single round trip

	// Ask the firestore client for the specified cart
	ref := cs.fsClient.Doc(storedCart.StoreRefPath())
	snap, err := cs.drProxy.Get(ref, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cart snapshot with ID %s: %w", cartId, err)
	}

	// Unmarshall the snapshot into our internal structure form
	err = cs.dsProxy.DataTo(snap, storedCart)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cart snapshot with ID %s: %w", cartId, err)
	}

	// Get the delivery address if one has been set
	storedCart.DeliveryAddress, err = cs.getDeliveryAddress(ctx, storedCart)
	if err != nil {
		return nil, err
	}

	// Get the cart items
	storedCart.CartItems, err = cs.getCartItems(ctx, storedCart)
	if err != nil {
		return nil, err
	}

	// All good, log our joy and return the protocol buffer transliteration of our retrieved cart
	return storedCart.AsPBShoppingCart(), nil
}

// getDeliveryAddress returns the delivery address for the given cart in its package internal structure form
// or nil if no address was found or an error occurred.
func (cs *CartService) getDeliveryAddress(ctx context.Context, cart *schema.ShoppingCart) (*types.PostalAddress, error) {

	// Ask the firestore client for the delivery address (if there is one)
	ref := cs.fsClient.Doc(cart.DeliveryAddressPath())
	snap, err := cs.drProxy.Get(ref, ctx)
	if err == nil {

		// Unmarshall the snapshot into our internal structure form
		deliveryAddress := &types.PostalAddress{}
		err = cs.dsProxy.DataTo(snap, deliveryAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal delivery addreess snapshot with cart ID %s: %w", cart.Id, err)
		}

		// Stuff the delivery address into the parent cart structure
		return deliveryAddress, nil
	}

	// We got an error but it might just be that the address was not found, i.e. not really an error
	if status.Code(err) == codes.NotFound {

		// Return absolutely nothing at all
		return nil, nil
	}

	// We experienced a significant error, report that back to the caller
	return nil, fmt.Errorf("failed to retrieve delivery address for cart with ID %s: %w", cart.Id, err)
}

// getCartItems returns the collection of cart items for the given cart in their package internal structure form.
// The returned slice may be empty if the cart does not currently contain any selected items.
func (cs *CartService) getCartItems(ctx context.Context, cart *schema.ShoppingCart) ([]*schema.ShoppingCartItem, error) {

	// Build our result set here
	var items []*schema.ShoppingCartItem

	// Obtain an iterator that can walk the cart's item collection's documents. Under the hood, a "GetAll"
	// operation retrieves all the item documents in a single round trip
	docs := cs.itemsGetterProxy.Items(cart).GetAll(ctx)

	// Close the iterator when we are done with it regardless of whether we are successful or not
	defer docs.Stop()

	// Do the iteration: gather a slice containing all the internal package representations of the cart items
	for {
		// Get the next document, if there is one
		item := &schema.ShoppingCartItem{}
		err := docs.Next(item)
		if err != nil {

			// We are either out of documents or have a real error
			if err != iterator.Done {
				return nil, fmt.Errorf("failed to retrieve cart item for cart with ID %s: %w", cart.Id, err)
			}

			//  For better or worse, we are done with this collection
			break
		}

		// Add this one to our result set and loop around for the next
		items = append(items, item)
	}

	// Against all odds, we made it through to a happy result
	return items, nil
}

// SetDeliveryAddress adds (or replaces) the delivery address to be used for physical cart items.
func (cs *CartService) SetDeliveryAddress(ctx context.Context, req *pbcart.SetDeliveryAddressRequest) (*pbcart.SetDeliveryAddressResponse, error) {

	// Obtain a shortcut handle on our globally configured logger
	l := zap.L()
	l.Info("setting delivery address", zap.String("CartId", req.CartId))

	// Form a skeleton cart representation that we can query for the delivery address path
	cart := &schema.ShoppingCart{Id: req.CartId}

	// Store the delivery address as a child of the cart in the firestore
	deliveryAddress := types.PostalAddressFromPB(req.DeliveryAddress)
	ref := cs.fsClient.Doc(cart.DeliveryAddressPath())
	_, err := cs.drProxy.Set(ref, ctx, deliveryAddress)
	if err != nil {
		err = fmt.Errorf("failed setting delivery address to firestore for cart: %w", err)
		l.Error(err.Error(), zap.String("CartId", req.CartId))
		return nil, err
	}

	// All good, log our joy and return the protocol buffer transliteration of our retrieved cart
	l.Info("delivery address set successfully", zap.String("CartId", req.CartId))

	// Have our internal sibling do all the remaining work
	pbCart, err := cs.getShoppingCart(ctx, req.CartId)
	if err != nil {
		return nil, err
	}
	return &pbcart.SetDeliveryAddressResponse{
		Cart: pbCart,
	}, nil
}

// AddItemToShoppingCart adds an item to a cart.
//
// TODO: Confirm that the item type is not already in the cart. Combine multiples / increasing quantity??
// TODO: Mock handling of requirements / dependencies / rejecting invalid combinations
// TODO: Access control? - Cannot change if cart already closed. Cannot change if user/shopper does not match.
func (cs *CartService) AddItemToShoppingCart(ctx context.Context, req *pbcart.AddItemToShoppingCartRequest) (*pbcart.AddItemToShoppingCartResponse, error) {

	// Obtain a shortcut handle on our globally configured logger
	l := zap.L()

	// Form a unique ID for the new cart item and include that in our entry log statement
	itemId := uuid.New().String()
	l.Info("adding cart item", zap.String("CartId", req.CartId), zap.String("ItemID", itemId))

	// Configure the ID values for this new item in the cart then transform it to our stored struct type
	req.Item.Id = itemId
	req.Item.CartId = req.CartId
	item := schema.ShoppingCartItemFromPB(req.Item)

	// Store the item as a child of the cart
	ref := cs.fsClient.Doc(item.StoreRefPath())
	_, err := cs.drProxy.Set(ref, ctx, item)
	if err != nil {
		err = fmt.Errorf("failed setting cart item to firestore for cart: %w", err)
		l.Error(err.Error(), zap.String("CartId", item.CartId), zap.String("ItemID", item.Id))
		return nil, err
	}

	// All good, log our joy before returning the protocol buffer transliteration of our retrieved cart
	l.Info("cart item added successfully", zap.String("CartId", item.CartId), zap.String("ItemID", item.Id))

	// Have our internal sibling do all the remaining work to return the complete cart as it now stands
	pbCart, err := cs.getShoppingCart(ctx, req.CartId)
	if err != nil {
		return nil, err
	}
	return &pbcart.AddItemToShoppingCartResponse{Cart: pbCart}, nil
}

// RemoveItemFromShoppingCart removes an item from the cart.
func (cs *CartService) RemoveItemFromShoppingCart(ctx context.Context, req *pbcart.RemoveItemFromShoppingCartRequest) (*pbcart.RemoveItemFromShoppingCartResponse, error) {

	// Obtain a shortcut handle on our globally configured logger then log what we are about to do
	l := zap.L()
	l.Info("removing cart item", zap.String("CartId", req.CartId), zap.String("ItemID", req.ItemId))

	// Use a partial item structure to form the key of the target item to be deleted
	target := schema.ShoppingCartItem{
		Id:     req.ItemId,
		CartId: req.CartId,
	}

	// Instruct Firestore to remove the item with extreme prejudice
	ref := cs.fsClient.Doc(target.StoreRefPath())
	_, err := cs.drProxy.Delete(ref, ctx)
	if err != nil {
		err = fmt.Errorf("failed deleting cart item from firestore: %w", err)
		l.Error(err.Error(), zap.String("CartId", target.CartId), zap.String("ItemID", target.Id))
		return nil, err
	}

	// Have our internal sibling do all the remaining work to return the complete cart as it now stands
	pbCart, err := cs.getShoppingCart(ctx, req.CartId)
	if err != nil {
		return nil, err
	}
	return &pbcart.RemoveItemFromShoppingCartResponse{Cart: pbCart}, nil
}

// CheckoutShoppingCart submits the order / checkout the shopping cart
//
// TODO: consider more informative structured error reporting in the response rather than as an HTTP / protocol error?
func (cs *CartService) CheckoutShoppingCart(ctx context.Context, req *pbcart.CheckoutShoppingCartRequest) (*pbcart.CheckoutShoppingCartResponse, error) {

	// Obtain a shortcut handle on our globally configured logger then log what we are about to do
	l := zap.L()
	l.Info("checking out cart", zap.String("CartId", req.CartId))

	// Check out and abandon are almost the same data operation except for the status value and log messaging
	pbCart, err := cs.closeCart(ctx, req.CartId, schema.CsCheckedOut)
	if err != nil {
		return nil, err
	}

	// All done, all happy
	l.Info("cart checked out successfully", zap.String("CartId", req.CartId))
	return &pbcart.CheckoutShoppingCartResponse{Cart: pbCart}, nil
}

// AbandonShoppingCart explicitly abandons a shopping cart in response to a user request (as against the
// system cancelling a cart that has gone unused for some period of time).
func (cs *CartService) AbandonShoppingCart(ctx context.Context, req *pbcart.AbandonShoppingCartRequest) (*pbcart.AbandonShoppingCartResponse, error) {

	// Obtain a shortcut handle on our globally configured logger then log what we are about to do
	l := zap.L()
	l.Info("abandoning out cart", zap.String("CartId", req.CartId))

	// Check out and abandon are almost the same data operation except for the status value and log messaging
	pbCart, err := cs.closeCart(ctx, req.CartId, schema.CsAbandonedByUser)
	if err != nil {
		return nil, err
	}

	// All done, all happy
	l.Info("cart abandoned successfully", zap.String("CartId", req.CartId))
	return &pbcart.AbandonShoppingCartResponse{Cart: pbCart}, nil
}

// closeCart updates the status of an open cart to one of the closed status options.
func (cs *CartService) closeCart(ctx context.Context, cartId string, closedState schema.CartStatus) (*pbcart.ShoppingCart, error) {

	// Ask the firestore client for the specified cart
	storedCart := schema.ShoppingCart{Id: cartId}
	ref := cs.fsClient.Doc(storedCart.StoreRefPath())
	snap, err := cs.drProxy.Get(ref, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cart snapshot with ID %s: %w", cartId, err)
	}

	// Unmarshall the snapshot into our internal structure form
	err = cs.dsProxy.DataTo(snap, storedCart)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cart snapshot with ID %s: %w", cartId, err)
	}

	// If the status is not currently open, we can't abandon it!
	if storedCart.Status != schema.CsOpen {

		// Watch out in case the cart status is one that we don't know about
		state := pbcart.ShoppingCartStatus_name[int32(storedCart.Status)]
		if state == "" {
			state = "unrecognized"
		}
		return nil, fmt.Errorf("cannot change status of cart that is not open: cart ID=%s, status=%s", storedCart.Id, state)
	}

	// Change the status and write the cart back to teh store
	storedCart.Status = closedState
	_, err = cs.drProxy.Set(ref, ctx, storedCart)
	if err != nil {
		err = fmt.Errorf("failed putting updated cart status to datastore: %w", err)
		zap.L().Error(err.Error(), zap.String("CartId", storedCart.Id))
		return nil, err
	}

	// All good, return the full updated cart or an error we get trying to retrieve it
	return cs.getShoppingCart(ctx, storedCart.Id)
}
