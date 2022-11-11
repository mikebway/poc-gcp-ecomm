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
}

// DocumentRefProxy defines the interface for a swappable junction that will allow us to maximize unit test coverage
// by intercepting calls to firestore.DocumentRef methods and having them return mock errors. The default
// production implementation of the interface will add a couple of nanoseconds of delay to normal operation but that
// extra test coverage is worth the price.
//
// See https://pkg.go.dev/cloud.google.com/go/firestore#DocumentRef
type DocumentRefProxy interface {
	Create(doc *firestore.DocumentRef, ctx context.Context, data interface{}) (*firestore.WriteResult, error)
	Get(doc *firestore.DocumentRef, ctx context.Context) (*firestore.DocumentSnapshot, error)
	Set(doc *firestore.DocumentRef, ctx context.Context, data interface{}) (*firestore.WriteResult, error)
}

// DocumentSnapshotProxy defines the interface for a swappable junction that will allow us to maximize unit test
// coverage by intercepting calls to firestore.DocumentSnapshot methods and having them return mock errors. The
// default production implementation of the interface will add a couple of nanoseconds of delay to normal
// operation but that extra test coverage is worth the price.
//
// See https://pkg.go.dev/cloud.google.com/go/firestore#DocumentSnapshot
type DocumentSnapshotProxy interface {
	DataTo(snap *firestore.DocumentSnapshot, target interface{}) error
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
	storedCart := &schema.ShoppingCart{
		Id: cartId,
	}

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
	deliveryAddress, err := cs.getDeliveryAddress(ctx, cartId)
	if err != nil {
		return nil, err
	}
	storedCart.DeliveryAddress = deliveryAddress

	// All good, log our joy and return the protocol buffer transliteration of our retrieved cart
	return storedCart.AsPBShoppingCart(), nil
}

// getDeliveryAddress returns the delivery address for the give cart in its package internal structure form
// or nil if no address was found or an error occurred.
func (cs *CartService) getDeliveryAddress(ctx context.Context, cartId string) (*types.PostalAddress, error) {

	// Ask the firestore client for the delivery address (if there is one)
	ref := cs.fsClient.Doc(schema.DeliveryAddressPath(cartId))
	snap, err := cs.drProxy.Get(ref, ctx)
	if err == nil {

		// Unmarshall the snapshot into our internal structure form
		deliveryAddress := &types.PostalAddress{}
		err = cs.dsProxy.DataTo(snap, deliveryAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal delivery addreess snapshot with cart ID %s: %w", cartId, err)
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
	return nil, fmt.Errorf("failed to retrieve delivery address for cart with ID %s: %w", cartId, err)
}

// SetDeliveryAddress adds (or replaces) the delivery address to be used for physical cart items.
func (cs *CartService) SetDeliveryAddress(ctx context.Context, req *pbcart.SetDeliveryAddressRequest) (*pbcart.SetDeliveryAddressResponse, error) {

	// Obtain a shortcut handle on our globally configured logger
	l := zap.L()
	l.Info("setting delivery address", zap.String("CartId", req.CartId))

	// Store the person that requested the cart be created as a child of the cart in the firestore
	deliveryAddress := types.PostalAddressFromPB(req.DeliveryAddress)
	ref := cs.fsClient.Doc(schema.DeliveryAddressPath(req.CartId))
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

// DocRefProxy is the production (i.e. non- unit test) implementation of the DocumentRefProxy interface.
// NewCartService configures it as the default in new CartService structure constructions.
type DocRefProxy struct {
	DocumentRefProxy
}

// Create is a direct pass through to the firestore.DocumentRef Create function. We use this rather than
// calling the firestore.DocumentRef function directly so that we can replace this implementation with
// one that allows errors to be inserted into the response when executing uni tests.
func (p *DocRefProxy) Create(doc *firestore.DocumentRef, ctx context.Context, data interface{}) (*firestore.WriteResult, error) {
	return doc.Create(ctx, data)
}

// Get is a direct pass through to the firestore.DocumentRef Get function. We use this rather than
// calling the firestore.DocumentRef function directly so that we can replace this implementation with one that
// allows errors to be inserted into the response when executing uni tests.
func (p *DocRefProxy) Get(doc *firestore.DocumentRef, ctx context.Context) (*firestore.DocumentSnapshot, error) {
	return doc.Get(ctx)
}

// Set is a direct pass through to the firestore.DocumentRef Set function. We use this rather than
// calling the firestore.DocumentRef function directly so that we can replace this implementation with one that
// allows errors to be inserted into the response when executing uni tests.
func (p *DocRefProxy) Set(doc *firestore.DocumentRef, ctx context.Context, data interface{}) (*firestore.WriteResult, error) {
	return doc.Set(ctx, data)
}

// DocSnapProxy is the production (i.e. non- unit test) implementation of the DocumentSnapshotProxy interface.
// NewCartService configures it as the default in new CartService structure constructions.
type DocSnapProxy struct {
	DocumentRefProxy
}

// DataTo is a direct pass through to the firestore.DocumentSnapshot DataTo function. We use this rather than
// calling the firestore.DocumentSnapshot function directly so that we can replace this implementation with
// one that allows errors to be inserted into the response when executing uni tests.
func (p *DocSnapProxy) DataTo(snap *firestore.DocumentSnapshot, target interface{}) error {
	return snap.DataTo(target)
}
