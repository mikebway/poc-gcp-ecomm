// Package service contains the gRPC Order microservice implementation.
package service

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"hash"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	cartsvc "github.com/mikebway/poc-gcp-ecomm/cart/service"
	"github.com/mikebway/poc-gcp-ecomm/order/schema"
	pborder "github.com/mikebway/poc-gcp-ecomm/pb/order"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
)

const (
	// defaultPageSize is used if no page size is supplied in a pborder.GetOrdersRequest, or the value is less than one.
	defaultPageSize = 20

	// maxPageSize is used if the page size is supplied in a pborder.GetOrdersRequest is greater than 100.
	maxPageSize = 100
)

var (
	// ProjectId is a variable so that unit tests can override it to ensures that test requests are not routed to
	// the live project! See https://firebase.google.com/doos/emulator-suite/connect_firestore
	ProjectId string

	// UnitTestNewOrderServiceError should be returned by NewOrderService if we are running unit tests
	// and unitTestNewOrderServiceError is not nil.
	UnitTestNewOrderServiceError error

	// PiiHash is the salt used by the PiiHashString function when rendering a PII string value into
	// a sufficiently unique string to be reliably searched for in logs without exposing the PII itself as
	// plain text.
	PiiHash hash.Hash
)

// init is the static initializer used to configure our local and global static variables.
func init() {

	// Set the project ID to be used for live Firestore etc. connections
	ProjectId = "poc-gcp-ecomm"

	// Initialize the salted hash used to hide PII values in our logs
	PiiHash = sha1.New()
	PiiHash.Write([]byte{0x8a, 0x19, 0x72, 0xa4, 0x00, 0xe8, 0x43, 0xda, 0x94, 0xf0, 0x59, 0x59, 0xc7, 0xb9, 0xaf, 0x7a})
}

// OrderService is a structure class with methods that implements the order.OrderAPIServer gRPC API
// storing the data for orders in a Google Cloud Firestore document collection.
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

	// queryProxy is used to allow unit tests to intercept firestore.Query function calls
	// and insert errors etc. into the responses of the document iterator that the query returns.
	queryProxy cartsvc.QueryExecutionProxy
}

// NewOrderService is a factory method returning an instance of our shopping cart service.
func NewOrderService() (*OrderService, error) {

	// Build our service instance here with our default, direct passthrough, interception proxies
	// for firestore.DocumentRef and firestore.DocumentSnapshot function calls
	svc := &OrderService{
		drProxy:    &cartsvc.DocRefProxy{},
		dsProxy:    &cartsvc.DocSnapProxy{},
		queryProxy: &cartsvc.QueryExecProxy{},
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

	// All done - return the populated service instance
	return svc, nil
}

// SaveOrder stores the given order in the Firestore document collection. This is for internal domain use only and
// so does not accept or return protobuf structures.
//
// An error will be returned if the order is already present in Firestore. Orders are immutable!
func (os *OrderService) SaveOrder(ctx context.Context, order *schema.Order) error {

	// TODO: confirm that Create fails if the order ID already exists in Firestore

	// Obtain a shortcut handle on our globally configured logger and log some context
	l := zap.L()
	l.Info("storing order", zap.String("orderId", order.Id))

	// Store the order in firestore
	ref := os.FsClient.Doc(order.StoreRefPath())
	_, err := os.drProxy.Create(ref, ctx, order)
	if err != nil {
		err = fmt.Errorf("failed creating order document in Firestore: %w", err)
		l.Error(err.Error(), zap.String("orderId", order.Id))
		return err
	}

	// All good, log our joy and return
	l.Info("order stored successfully", zap.String("orderId", order.Id), zap.String("path", ref.Path))
	return nil
}

// GetOrderByID retrieves an order matching the specified UUID ID in the pborder.GetOrderByIDRequest.
func (os *OrderService) GetOrderByID(ctx context.Context, req *pborder.GetOrderByIDRequest) (*pborder.GetOrderByIDResponse, error) {

	// TODO: Access control

	// Obtain a shortcut handle on our globally configured logger and log some context information
	l := zap.L()
	l.Info("retrieving order", zap.String("orderId", req.OrderId))

	// Form an order structure to receive the data from the store
	order := &schema.Order{Id: req.OrderId}

	// Ask the firestore client for the specified order
	ref := os.FsClient.Doc(order.StoreRefPath())
	snap, err := os.drProxy.Get(ref, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve order snapshot with ID %s: %w", req.OrderId, err)
	}

	// Unmarshall the snapshot into our internal structure form
	err = os.dsProxy.DataTo(snap, order)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal order snapshot with ID %s: %w", req.OrderId, err)
	}

	// Convert the internal order structure ro it protobuf form
	pbOrder := order.AsPBOrder()

	// Wrap the order in the response structure and we are done
	return &pborder.GetOrderByIDResponse{
		Order: pbOrder,
	}, nil
}

// GetOrders retrieves a page of orders matching the search criteria specified UUID ID in the pborder.GetOrdersRequest.
func (os *OrderService) GetOrders(ctx context.Context, req *pborder.GetOrdersRequest) (*pborder.GetOrdersResponse, error) {

	// Log what we have been asked to do as context for any later logging on this thread. PII values will be masked.
	os.logQuery(req)

	// Adjust the page size if it is unreasonable
	if req.PageSize < 1 {
		zap.L().Warn("negative/zero page size adjusted to default", zap.Int32("requested", req.PageSize), zap.Int("default", defaultPageSize))
		req.PageSize = defaultPageSize
	} else if req.PageSize > maxPageSize {
		zap.L().Warn("excessive page size adjusted to maximum", zap.Int32("requested", req.PageSize), zap.Int("max", maxPageSize))
		req.PageSize = maxPageSize
	}

	// Start with the whole collection and build up the query from there
	query, err := os.buildOrderQuery(os.FsClient.Collection(schema.OrderCollection).Query, req)
	if err != nil {
		return nil, err
	}

	// Run the query to the set of matching orders. Casting the page size is safe - we just checked that it lies between 1 and 100
	orders, nextPageToken, err := os.executeQuery(ctx, query, int(req.PageSize))
	if err != nil {
		return nil, err
	}

	// Loop, converting the slice of internal format orders to their protobuf equivalents
	pbOrders := make([]*pborder.Order, len(orders))
	for i, order := range orders {
		pbOrders[i] = order.AsPBOrder()
	}

	// And we are all done
	return &pborder.GetOrdersResponse{
		Orders:        pbOrders,
		NextPageToken: nextPageToken,
	}, nil
}

// executeQuery uses the supplied query to obtain an order document iterator, then build a slice of
// results from that.
func (os *OrderService) executeQuery(ctx context.Context, query firestore.Query, pageSize int) ([]*schema.Order, string, error) {

	// Run the query to obtain an iterator over the matching documents
	docs := os.queryProxy.Documents(ctx, query)

	// Close the iterator when we are done with it regardless of whether we are successful or not
	defer docs.Stop()

	// Do the iteration: gather a slice containing all the internal package order representations
	var orders []*schema.Order
	for {
		// Get the next document, if there is one
		item := &schema.Order{}
		err := docs.Next(item)
		if err != nil {

			// We are either out of documents or have a real error
			if err != iterator.Done {
				return nil, "", fmt.Errorf("failed to retrieve order matching query: %w", err)
			}

			//  For better or worse, we are done with this collection
			break
		}

		// Add this one to our result set and loop around for the next
		orders = append(orders, item)
	}

	// Could there be another page? The orders slice should never be longer than pageSize, but we test for >= defensively
	nextPageToken := ""
	orderCount := len(orders)
	if orderCount >= pageSize {

		// There could be more to load, use the last processed order's ID as our position marker
		lastOrder := orders[orderCount-1]
		nextPageToken = fmt.Sprintf("%x,%s", lastOrder.SubmissionTime.UnixNano(), orders[orderCount-1].Id)
	}

	// All is well if we reach this point
	return orders, nextPageToken, nil
}

// buildOrderQuery translates the pborder.GetOrdersRequest parameters into filters on the given firestore.Query.
// Always check the error return value - a query is returned whether an error occurred or not.
func (os *OrderService) buildOrderQuery(query firestore.Query, req *pborder.GetOrdersRequest) (firestore.Query, error) {

	// Ignore documents that fall outside the time window first then narrow the result set down from there
	// if the caller specified that mach in their request ...
	if req.StartTime != nil {
		query = query.Where("submissionTime", ">=", req.GetStartTime().AsTime())
	}
	if req.EndTime != nil {
		query = query.Where("submissionTime", "<", req.GetEndTime().AsTime())
	}
	if len(req.FamilyName) > 0 {
		query = query.Where("orderedBy.familyName", "==", req.FamilyName)
	}
	if len(req.GivenName) > 0 {
		query = query.Where("orderedBy.givenName", "==", req.GivenName)
	}

	// Order the results by submission time first, then by order ID (necessary for us to have a unique cursor position for paging)
	query = query.OrderBy("submissionTime", firestore.Asc).OrderBy("id", firestore.Asc)

	// If a page token was specified, use that as the marker for the last document that has
	// already been returned, i.e. start after that one.
	if len(req.PageToken) > 0 {

		// Split the page token into its two parts - there should be two parts, submission time and ID
		afterTime, afterOrderId, err := splitPageToken(req.PageToken)
		if err != nil {
			return query, err
		}

		// Add the start after factors to our query
		query = query.StartAfter(afterTime, afterOrderId)
	}

	// Limit the size of the result set to the page size
	query = query.Limit(int(req.PageSize))

	// All done
	return query, nil
}

// splitPageToken breaks a page token string into its time and order ID components.
func splitPageToken(token string) (time.Time, string, error) {

	// Split the page token into its two parts - there should be two parts
	parts := strings.Split(token, ",")
	if len(parts) == 2 {

		// Convert the first part from a Unix time value to a time.Time
		unixNanoTime, err := strconv.ParseInt(parts[0], 16, 64)
		if err == nil {

			// All looks good (as shallowly as we are bothering to look), return what we have
			return time.Unix(0, unixNanoTime), parts[1], nil
		}
	}

	// Only arrive here if the page token is invalid
	return time.Time{}, "", fmt.Errorf("invalid page token: %s", token)
}

// logQuery writes a log entry documenting the attributes that make up a OrderService.GetOrders Firestore query.
// PII fields, i.e. family name, will be one way encrypted so that they can be safely logged (hashed with
// a salt, encoded as base65, and truncated to 12 characters).
func (os *OrderService) logQuery(req *pborder.GetOrdersRequest) {

	// Build a slice of zap.Fields with whatever query parameters are found in the request
	var fields []zap.Field
	if req.StartTime != nil {
		fields = append(fields, zap.Time("startTime", req.StartTime.AsTime()))
	}
	if req.EndTime != nil {
		fields = append(fields, zap.Time("endTime", req.EndTime.AsTime()))
	}
	if len(req.FamilyName) > 0 {
		fields = append(fields, zap.String("familyName", PiiHashString(req.FamilyName)))
	}
	if len(req.GivenName) > 0 {
		fields = append(fields, zap.String("givenName", PiiHashString(req.GivenName)))
	}
	if len(req.PageToken) > 0 {
		fields = append(fields, zap.String("pageToken", req.PageToken))
	}

	// Log that slice and we are done
	zap.L().Info("get order", fields...)
}

// PiiHashString renders a string PII value as a truncated base64 salted hash that can be safely rendered
// in logs to support searching without exposing PII values such as name and address as plain text.
func PiiHashString(value string) string {

	// Obtain the salted SH1 hash of the string value
	hashedValue := PiiHash.Sum([]byte(value))

	// Express that as base64 string
	b64HashedValue := base64.StdEncoding.EncodeToString(hashedValue)

	// Return just the first 12 characters of that
	return b64HashedValue[:12]
}
