// Package service contains the gRPC Shopping Cart microservice implementation.
package service

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	cartsvc "github.com/mikebway/poc-gcp-ecomm/cart/fsproxies"
	pborder "github.com/mikebway/poc-gcp-ecomm/pb/order"
)

var (
	// ProjectId is a variable so that unit tests can override it to ensures that test requests are not routed to
	// the live project! See https://firebase.google.com/docs/emulator-suite/connect_firestore
	ProjectId string

	// UnitTestNewOrderServiceError should be returned by NewOrderService if we are running unit tests
	// and unitTestNewOrderServiceError is not nil.
	UnitTestNewOrderServiceError error
)

// init is the static initializer used to configure our local and global static variables.
func init() {

	// Set the project ID to be used for live Firestore etc. connections
	ProjectId = "poc-gcp-ecomm"
}

// OrderService is a structure class with methods that implements the order.OrderAPIServer gRPC API
// storing the data for shopping carts in a Google Cloud Firestore document collection.
type OrderService struct {
	pborder.UnimplementedOrderAPIServer

	// FsClient is the GCP Firestore client - it is thread safe and can be reused concurrently
	FsClient *firestore.Client

	// drProxy is used to allow unit tests to intercept firestore.DocumentRef function calls
	// and insert errors etc. into the responses.
	drProxy cartsvc.DocumentRefProxy

	// dsProxy is used to allow unit tests to intercept firestore.DocumentSnapshot function calls
	// and insert errors etc. into the responses.
	dsProxy cartsvc.DocumentSnapshotProxy

	// itemsGetterProxy is used to obtain an ItemsCollectionProxy for a given cart. Unit tests may
	// substitute an alternative implementation this interface in order to be able to insert errors etc.
	// into the responses of the ItemsCollectionProxy that the ItemCollectionGetterProxy returns.
	itemsGetterProxy cartsvc.ItemCollectionGetterProxy
}

// NewOrderService is a factory method returning an instance of our shopping cart service.
func NewOrderService() (*OrderService, error) {

	// Build our service instance here with our default, direct passthrough, interception proxies
	// for firestore.DocumentRef and firestore.DocumentSnapshot function calls
	svc := &OrderService{
		drProxy: &cartsvc.DocRefProxy{},
		dsProxy: &cartsvc.DocSnapProxy{},
	}

	// Obtain a firestore client and stuff that in the service instance
	ctx := context.Background()
	var err error
	if UnitTestNewOrderServiceError == nil {
		// Set the Firestore client if we are not unit testing an error situation.
		svc.FsClient, err = firestore.NewClient(ctx, ProjectId)

	} else {
		// We are unit testing and required to report an error
		err = UnitTestNewOrderServiceError
	}

	// Check that we obtained a firestore client successfully
	if err != nil {
		return nil, fmt.Errorf("could not obtain firestore client: %w", err)
	}

	// Make the firestore client available to the cart item getter proxy
	svc.itemsGetterProxy = &cartsvc.ItemCollGetterProxy{
		// FsClient: svc.FsClient,
	}

	// All done - return the populated service instance
	return svc, nil
}
