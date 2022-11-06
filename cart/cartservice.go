package main

import (
	"cloud.google.com/go/datastore"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/mikebway/poc-gcp-ecomm/cart/schema"
	pbcart "github.com/mikebway/poc-gcp-ecomm/pb/cart"
	"github.com/mikebway/poc-gcp-ecomm/types"
	"go.uber.org/zap"
	"time"
)

// CartService is a structure class with methods that implements the cart.CartAPIServer gRPC API
// storing the data for the social graph in a Google Cloud Datastore Kind.
type CartService struct {
	pbcart.UnimplementedCartAPIServer

	// dsClient is the GCP Datastore client - it is thread safe and can be reused concurrently
	dsClient dsinterface.DatastoreClient
}

// NewCartService is a factory method returning an instance of our social graph service.
func NewCartService() (*CartService, error) {

	// Build our service instance here
	svc := &CartService{}

	// Obtain a datastore client and stuff that in the service instance
	ctx := context.Background()
	var dsClient dsinterface.DatastoreClient
	var err error
	if !isUnitTesting {
		// Only set the datastore client if we are not unit testing. Unit test will add their own mock
		// datastore client later if need be. datastore.NewClient will fail if we are not running in
		// either a real GCP environment or an emulation of one so can't be invoked in a vanilla unit
		// test context.
		dsClient, err = datastore.NewClient(ctx, datastore.DetectProjectID)

	} else {
		// We are unit testing - should we report and error to see how our caller handles it?
		if unitTestNewCartServiceError != nil {
			return nil, unitTestNewCartServiceError
		}
	}

	// Check that we obtained a datastore client successfully
	if err != nil {
		return nil, fmt.Errorf("could not obtain datastore client: %w", err)
	}
	svc.dsClient = dsClient

	// All done - return the populated service instance
	return svc, nil
}

// CreateShoppingCart returns a new shopping cart structure with a unique ID, creation time, and status assigned.
func (cs *CartService) CreateShoppingCart(ctx context.Context, req *pbcart.CreateShoppingCartRequest) (*pbcart.CreateShoppingCartResponse, error) {

	// Obtain a shortcut handle on our globally configured logger
	l := zap.L()

	// Create the storable cart structure with a new unique ID
	storableCart := schema.ShoppingCart{
		Id:           uuid.New().String(),
		CreationTime: time.Now(),
		Status:       schema.CsOpen,
		Shopper:      types.PersonFromPB(req.Shopper),
	}
	l.Info("storing new cart", zap.String("CartId", storableCart.Id))

	// TODO: wrap cart creation in a transaction
	// TODO: use MultiPut

	// Store the so far empty cart in the datastore
	cartKey := storableCart.DatastoreKey()
	_, err := cs.dsClient.Put(ctx, cartKey, &storableCart)
	if err != nil {
		err = fmt.Errorf("failed putting new cart to datastore: %w", err)
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
	storedCart := schema.ShoppingCart{
		Id: cartId,
	}

	// TODO: wrap cart retrieval in a transaction
	// TODO: use MultiGet
	// TODO: Convert cart in one step after populating with shopper and address
	// TODO: Store shopper directly in cart not as a child entity

	// Ask the datastore client for the specified cart
	err := cs.dsClient.Get(ctx, storedCart.DatastoreKey(), &storedCart)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cart with ID %s: %w", cartId, err)
	}

	// Convert the stored cart data to the protocol buffer equivalent
	pbCart := storedCart.AsPBShoppingCart()

	// Ask the datastore client for the delivery address (if there is one)
	deliveryAddress := types.PostalAddress{}
	err = cs.dsClient.Get(ctx, schema.DeliveryAddressKey(cartId), &deliveryAddress)
	if err == nil {

		// We retrieved a delivery address - set that into the cart
		pbCart.DeliveryAddress = deliveryAddress.AsPBPostalAddress()

	} else if err != datastore.ErrNoSuchEntity {

		// We experienced a significant error
		return nil, fmt.Errorf("failed to retrieve delivery address for cart with ID %s: %w", cartId, err)
	}

	// All good, log our joy and return the protocol buffer transliteration of our retrieved cart
	return pbCart, nil
}

// SetDeliveryAddress adds (or replaces) the delivery address to be used for physical cart items.
func (cs *CartService) SetDeliveryAddress(ctx context.Context, req *pbcart.SetDeliveryAddressRequest) (*pbcart.SetDeliveryAddressResponse, error) {

	// Obtain a shortcut handle on our globally configured logger
	l := zap.L()
	l.Info("setting delivery address", zap.String("CartId", req.CartId))

	// TODO: Access control?

	// Store the person that requested the cart be created as a child of the cart in the datastore
	deliveryAddrKey := schema.DeliveryAddressKey(req.CartId)
	deliveryAddress := types.PostalAddressFromPB(req.DeliveryAddress)
	_, err := cs.dsClient.Put(ctx, deliveryAddrKey, deliveryAddress)
	if err != nil {
		err = fmt.Errorf("failed putting cart delivery address to datastore: %w", err)
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
