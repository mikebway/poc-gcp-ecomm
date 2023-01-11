package orderapi

import (
	"errors"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	cartapi "github.com/mikebway/poc-gcp-ecomm/cart/cartapi"
	"github.com/mikebway/poc-gcp-ecomm/order/schema"
	pborder "github.com/mikebway/poc-gcp-ecomm/pb/order"
	"github.com/mikebway/poc-gcp-ecomm/testutil"
	"github.com/mikebway/poc-gcp-ecomm/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// EnvFirestoreEmulator defines the environment variable name that is used to convey that the Firestore emulator
	// is running, should be used, and how to connect to it
	EnvFirestoreEmulator = "FIRESTORE_EMULATOR_HOST"

	// FirestoreEmulatorHost defines the server name and port (in TCP6 terms) of the Firestore emulator
	FirestoreEmulatorHost = "[::1]:8219"

	// A timestamp string we can use to derive known time values. It will be used as the first submission time in the
	// series of 10 orders that primeFirestore adds to the Firestore emulator
	//
	// Firestore time representation when evaluating query matches is not as accurate as a full 9 digits
	// of nanoseconds, so we zero those out. It has better than second accuracy, but I am not going to
	// bother figuring out exactly how much better for the purposes of the TestSubmissionTimeQuery test.
	timeString = "2022-10-11T10:23:19.000000000-06:00"

	// UnitTestGivenName is used for all given names for the ordering person for all unit test orders written
	// to the Firestore emulator. It is used so that we can find and delete all orders we create in these unit
	// tests (and only those orders) so that we can be sure of starting with clean slate.
	UnitTestGivenName = "Unit~Test"

	// Define the additional person fields that we will for the one mock order where we include more detail
	personId          = "10615145-2010-4c5f-8347-2bb556232c31"
	personFamilyName  = "Grint"
	personMiddleName  = "Alexander Lloyd"
	personDisplayName = "Rupert"

	// Define the postal address fields for the home address of our mock person
	addrLine1      = "55 Yonder St"
	addrLine2      = "Flat B"
	addrLocality   = "Ottery St Catchpole"
	addrPostalCode = "EX11 1HF"
	addrRegionCode = "GB"

	// unitTestErrorMessage is used as the error description for error that are deliberately forced to
	// test error handling.
	unitTestErrorMessage = "unit test of error handling"
)

var (
	// primed is true if primeFirestore() has already populated the Firestore emulator with mock orders
	primed = false

	// mockOrders is a slice of 10 orders that the Firestore emulator is primed with for all our tests.
	mockOrders []*schema.Order
)

// UTQueryExecProxy implements a wrapper function around firestore.Query that will return an iterator over items
// that match the query. For unit test purposes, this version always returns errors when trying to iterate over
// the result set.
type UTQueryExecProxy struct {
	cartapi.QueryExecutionProxy
}

// UTDocIteratorProxy is a unit test implementation of the DocumentIteratorProxy interface that always returns
// errors when trying to iterate over the result set.
type UTDocIteratorProxy struct {
	cartapi.DocumentIteratorProxy
}

// Documents returns a DocumentIteratorProxy wrapping the results of the given query. Unit test implementations
// of this function can be programmed to return an iterator that can insert errors into the flow but this
// production ready implementation returns a transparent passthrough iterator.
func (q *UTQueryExecProxy) Documents(ctx context.Context, query firestore.Query) cartapi.DocumentIteratorProxy {
	return &UTDocIteratorProxy{}
}

// Next would normally return the next document is a result set but this unit test version always
// returns an error.
func (p *UTDocIteratorProxy) Next(target interface{}) error {
	return errors.New(unitTestErrorMessage)
}

// Stop stops the iterator, freeing its resources. Always call Stop when you are done with a DocumentIterator.
// It is not safe to call Stop concurrently with Next.
func (p *UTDocIteratorProxy) Stop() {
	// We have nothing to stop :-)
}

// UTDocRefProxy is a unit test implementation of the DocumentRefProxy interface that allows
// unit tests to have Firestore operations return errors.
type UTDocRefProxy struct {
	cartapi.DocumentRefProxy
}

// Create is a pass through to the firestore.DocumentRef Create function  that allows
// unit tests to have Firestore operations return errors.
func (p *UTDocRefProxy) Create(doc *firestore.DocumentRef, ctx context.Context, data interface{}) (*firestore.WriteResult, error) {
	return nil, errors.New(unitTestErrorMessage)
}

// UTDocSnapProxy is a unit test implementation of the DocumentSnapshotProxy interface that allows
// unit tests to have Firestore operations return errors.
type UTDocSnapProxy struct {
	cartapi.DocumentRefProxy
}

// DataTo is a direct pass through to the firestore.DocumentSnapshot DataTo function that allows
// unit tests to have Firestore operations return errors.
func (p *UTDocSnapProxy) DataTo(snap *firestore.DocumentSnapshot, target interface{}) error {
	return errors.New(unitTestErrorMessage)
}

// TestMain, if defined (it's optional), allows setup code to be run before and after the suite of unit tests
// for this package.
func TestMain(m *testing.M) {

	// Ensure that our Firestore requests do not get routed to the live project by mistake
	ProjectId = "demo-" + ProjectId

	// Configure the environment variable that informs the Firestore client that it should connect to the
	// emulator and how to reach it.
	_ = os.Setenv(EnvFirestoreEmulator, FirestoreEmulatorHost)

	// Run all the unit tests
	m.Run()
}

// TestSaveOrderError covers that part of the OrderService.SaveOrder that then 10+ orders we save successfully
// in these test do not get to - the failures.
func TestSaveOrderError(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// Replace the document reference proxy of the service with one that will behave badly at our direction
	service.drProxy = &UTDocRefProxy{}

	// We don't need much of an order to test an error that we know will be thrown
	order := &schema.Order{
		Id: uuid.NewString(),
	}

	// Attempt to save the order, certain this will fail
	err := service.SaveOrder(ctx, order)
	assert.NotNil(err, "should have seen a forced error saving an order")
	assert.Contains(err.Error(), unitTestErrorMessage, "did not see the specific error that we expected")
}

// TestGetOrderByID retrieves one of the mock orders that primeFirestore has stores in the Firestore emulator.
func TestGetOrderByID(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// Get our second mock order - the one that is more fully populated than all the others
	response, err := service.GetOrderByID(ctx, &pborder.GetOrderByIDRequest{OrderId: mockOrders[1].Id})
	assert.Nil(err, "should not have failed retrieving order ID %s: %v", mockOrders[1].Id, err)
	order := response.Order
	assert.NotNil(order, "did not get an order in the response")

	// Confirm all the values are present as expected
	assert.Equal(mockOrders[1].Id, order.Id, "order ID did not match")
	assert.Equal(mockOrders[1].SubmissionTime.Unix(), order.SubmissionTime.AsTime().Unix(), "submission time does not match")
	assert.NotNil(order.OrderedBy, "ordered by person missing")
	assert.Equal(personId, order.OrderedBy.Id, "ordered by person ID does not match")
	assert.Equal(personFamilyName, order.OrderedBy.FamilyName, "ordered by person family name does not match")
	assert.Equal(UnitTestGivenName, order.OrderedBy.GivenName, "ordered by person given name does not match")
	assert.Equal(personMiddleName, order.OrderedBy.MiddleName, "ordered by person middle does not match")
	assert.Equal(personDisplayName, order.OrderedBy.DisplayName, "ordered by person display name does not match")

	assert.NotNil(order.DeliveryAddress, "delivery address missing")
	assert.Equal(2, len(order.DeliveryAddress.AddressLines), "delivery address line count does not match")
	assert.Equal(addrLine1, order.DeliveryAddress.AddressLines[0], "delivery address line 1 does not match")
	assert.Equal(addrLine2, order.DeliveryAddress.AddressLines[1], "delivery address line 2 does not match")
	assert.Equal(addrLocality, order.DeliveryAddress.Locality, "delivery address locality does not match")
	assert.Equal(addrPostalCode, order.DeliveryAddress.PostalCode, "delivery address postal code does not match")
	assert.Equal(addrRegionCode, order.DeliveryAddress.RegionCode, "delivery address region code does not match")

	assert.Equal(2, len(order.OrderItems), "order item count does not match")

	assert.NotNil(order.OrderItems[0], "order item 1 missing")
	assert.Equal(mockOrders[1].OrderItems[0].Id, order.OrderItems[0].Id, "order item 1 ID does not match")
	assert.Equal(mockOrders[1].OrderItems[0].ProductCode, order.OrderItems[0].ProductCode, "order item 1 product code does not match")
	assert.Equal(mockOrders[1].OrderItems[0].Quantity, order.OrderItems[0].Quantity, "order item 1 quantity does not match")
	assert.NotNil(order.OrderItems[0].UnitPrice, "order item 1 unit price missing")
	assert.Equal(mockOrders[1].OrderItems[0].UnitPrice.CurrencyCode, order.OrderItems[0].UnitPrice.CurrencyCode, "order item 1 price currency does not match")
	assert.Equal(mockOrders[1].OrderItems[0].UnitPrice.Units, order.OrderItems[0].UnitPrice.Units, "order item 1 price units does not match")
	assert.Equal(mockOrders[1].OrderItems[0].UnitPrice.Nanos, order.OrderItems[0].UnitPrice.Nanos, "order item 1 price nanos does not match")

	assert.NotNil(order.OrderItems[1], "order item 2 missing")
	assert.Equal(mockOrders[1].OrderItems[1].Id, order.OrderItems[1].Id, "order item 2 ID does not match")
	assert.Equal(mockOrders[1].OrderItems[1].ProductCode, order.OrderItems[1].ProductCode, "order item 2 product code does not match")
	assert.Equal(mockOrders[1].OrderItems[1].Quantity, order.OrderItems[1].Quantity, "order item 2 quantity does not match")
	assert.NotNil(order.OrderItems[1].UnitPrice, "order item 2 unit price missing")
	assert.Equal(mockOrders[1].OrderItems[1].UnitPrice.CurrencyCode, order.OrderItems[1].UnitPrice.CurrencyCode, "order item 2 price currency does not match")
	assert.Equal(mockOrders[1].OrderItems[1].UnitPrice.Units, order.OrderItems[1].UnitPrice.Units, "order item 2 price units does not match")
	assert.Equal(mockOrders[1].OrderItems[1].UnitPrice.Nanos, order.OrderItems[1].UnitPrice.Nanos, "order item 2 price nanos does not match")
}

// TestGetOrderByIDNotFound tries to retrieve a document that does not exist.
func TestGetOrderByIDNotFound(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// Ask for the non-existent order
	response, err := service.GetOrderByID(ctx, &pborder.GetOrderByIDRequest{OrderId: "no-way-this-exists"})
	assert.NotNil(err, "should have failed retrieving a non-existent order")
	assert.Contains(err.Error(), "failed to retrieve order snapshot with ID no-way-this-exists", "did not see the error we expected")
	assert.Nil(response, "should not have received a response")
}

// TestGetOrderByIDCorrupt tries to retrieve a document that cannot be unmarshalled from its snapshot
func TestGetOrderByIDCorrupt(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// Replace the document reference proxy of the service with one that will behave badly at our direction
	service.dsProxy = &UTDocSnapProxy{}

	// Ask for our second mock order - the one that is more fully populated than all the others but
	// the overriden proxy will reject
	response, err := service.GetOrderByID(ctx, &pborder.GetOrderByIDRequest{OrderId: mockOrders[1].Id})
	assert.NotNil(err, "should have failed retrieving a non-existent order")
	assert.Contains(err.Error(), "failed to unmarshal order snapshot", "did not see the error we expected")
	assert.Nil(response, "should not have received a response")
}

// TestSubmissionTimeQuery tries out finding multiple orders that fall within a given time span and paging to boot.
func TestSubmissionTimeQuery(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// Request what we know should be the first 5 of 8 orders starting with the second one and excluding the last
	// of our mock set
	request := &pborder.GetOrdersRequest{
		StartTime: timestamppb.New(mockOrders[1].SubmissionTime), // order[1] should be in the result set
		EndTime:   timestamppb.New(mockOrders[9].SubmissionTime), // order[1] should NOT be in the result set
		PageSize:  5,
	}
	response, err := service.GetOrders(ctx, request)
	assert.Nil(err, "did not expect an error calling GetOrders for the first page: %v", err)

	// The first page returned should be index entries 1 through 5 of the complete 0 to 9 set
	orders := response.Orders
	assert.Equal(5, len(orders), "expect 5 orders in the first full page")
	assert.Equal(mockOrders[1].Id, orders[0].Id, "order ID 1 of 8 did not match")
	assert.Equal(mockOrders[2].Id, orders[1].Id, "order ID 2 of 8 did not match")
	assert.Equal(mockOrders[3].Id, orders[2].Id, "order ID 3 of 8 did not match")
	assert.Equal(mockOrders[4].Id, orders[3].Id, "order ID 4 of 8 did not match")
	assert.Equal(mockOrders[5].Id, orders[4].Id, "order ID 5 of 8 did not match")
	assert.NotEmpty(response.NextPageToken, "next page token should have been set after retrieving first page")

	// Update the request to retrieve the second page
	request.PageToken = response.NextPageToken
	response, err = service.GetOrders(ctx, request)
	assert.Nil(err, "did not expect an error calling GetOrders for the second page: %v", err)
	orders = response.Orders
	assert.Equal(3, len(orders), "expect 3 orders in the second partial page")
	assert.Equal(mockOrders[6].Id, orders[0].Id, "order ID 6 of 8 did not match")
	assert.Equal(mockOrders[7].Id, orders[1].Id, "order ID 7 of 8 did not match")
	assert.Equal(mockOrders[8].Id, orders[2].Id, "order ID 8 of 8 did not match")
	assert.Empty(response.NextPageToken, "next page token should NOT have been set after retrieving second page")
}

// TestFamilyNameQuery tries out finding multiple orders that were for the same family name.
func TestFamilyNameQuery(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// Request what we know should be the first 5 of 8 orders starting with the second one and excluding the last
	// of our mock set
	request := &pborder.GetOrdersRequest{
		FamilyName: "RepeatedName",
		PageSize:   5,
	}
	response, err := service.GetOrders(ctx, request)
	assert.Nil(err, "did not expect an error calling GetOrders for the first and only page: %v", err)

	// The first page returned should have two entries. There should not be another page.
	orders := response.Orders
	assert.Equal(2, len(orders), "expect 2 orders in the first and only page")
	assert.Equal("4th", orders[0].DeliveryAddress.Locality, "locality 1 of 2 did not match")
	assert.Equal("6th", orders[1].DeliveryAddress.Locality, "locality 2 of 2 did not match")
	assert.Empty(response.NextPageToken, "next page token should NOT have been set after retrieving first page")
}

// TestGivenNameQuery tries out finding multiple orders that were for the same given name.
func TestGivenNameQuery(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// Request what we know should be the first 5 of 10 orders starting with the second one and excluding the last
	// of our mock set
	request := &pborder.GetOrdersRequest{
		GivenName: UnitTestGivenName, // All 10 or our mock orders have the same given name
		PageSize:  5,
	}
	response, err := service.GetOrders(ctx, request)
	assert.Nil(err, "did not expect an error calling GetOrders for the first full page: %v", err)

	// The first page returned should have two entries. There should not be another page.
	orders := response.Orders
	assert.Equal(5, len(orders), "expect 5 orders in the first full page")
	assert.Equal(mockOrders[0].Id, orders[0].Id, "order ID 1 of 10 did not match")
	assert.Equal(mockOrders[1].Id, orders[1].Id, "order ID 2 of 10 did not match")
	assert.Equal(mockOrders[2].Id, orders[2].Id, "order ID 3 of 10 did not match")
	assert.Equal(mockOrders[3].Id, orders[3].Id, "order ID 4 of 10 did not match")
	assert.Equal(mockOrders[4].Id, orders[4].Id, "order ID 5 of 10 did not match")
	assert.NotEmpty(response.NextPageToken, "next page token should have been set after retrieving first page")

	// Update the request to retrieve the second page
	request.PageToken = response.NextPageToken
	response, err = service.GetOrders(ctx, request)
	assert.Nil(err, "did not expect an error calling GetOrders for the second page: %v", err)
	orders = response.Orders
	assert.Equal(5, len(orders), "expect 5 orders in the second full page")
	assert.Equal(mockOrders[5].Id, orders[0].Id, "order ID 6 of 10 did not match")
	assert.Equal(mockOrders[6].Id, orders[1].Id, "order ID 7 of 10 did not match")
	assert.Equal(mockOrders[7].Id, orders[2].Id, "order ID 8 of 10 did not match")
	assert.Equal(mockOrders[8].Id, orders[3].Id, "order ID 9 of 10 did not match")
	assert.Equal(mockOrders[9].Id, orders[4].Id, "order ID 10 of 10 did not match")
	assert.NotEmpty(response.NextPageToken, "next page token should have been set after retrieving second page")

	// Update the request to retrieve the third page
	request.PageToken = response.NextPageToken
	response, err = service.GetOrders(ctx, request)
	assert.Nil(err, "did not expect an error calling GetOrders for the third page: %v", err)
	assert.Equal(0, len(response.Orders), "expect zero orders in the third page")
	assert.Empty(response.NextPageToken, "next page token should NOT have been set after retrieving third page")
}

// TestLoggingAndBadPageToken kills two birds with one stone, verifying that queries are logged for diagnostic
// purposes and looking at how the code handles an invalid "net page token."
func TestLoggingAndBadPageToken(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// For a request with an invalid page token
	request := &pborder.GetOrdersRequest{
		GivenName: UnitTestGivenName, // All 10 or our mock orders have the same given name
		PageSize:  5,
		PageToken: "not_a_number,not_a_uuid",
	}

	// Wrap the query execution to capture its log output
	var response *pborder.GetOrdersResponse
	var err error
	logged := testutil.CaptureLogging(func() {
		response, err = service.GetOrders(ctx, request)
	})

	// We should have see a error complaining about the page token
	assert.NotNil(err, "expected an error calling GetOrders")
	assert.Contains(err.Error(), "invalid page token", "did not get the expected error")
	assert.Nil(response, "should not have received a response")

	// Confirm that the log output described the query
	assert.Contains(logged, "get order", "did not find the query in the log output")
	assert.Contains(logged, "givenName", "log output should name the givenName field")
	assert.NotContains(logged, UnitTestGivenName, "log output should not contain the given name in plain text")
	assert.Contains(logged, PiiHashString(UnitTestGivenName), "log output should contain the given name in encrypted form")
	assert.Contains(logged, "pageToken", "log output should name the pageToken field")
	assert.Contains(logged, request.PageToken, "log output should contain the duff page token")
}

// TestQueryError looks at how the code handles an error while executing a Firestore query
func TestQueryError(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// Replace the query proxy of the service with one that will behave badly at our direction
	service.queryProxy = &UTQueryExecProxy{}

	// Request what we know would normally return the two orders with the same family name
	request := &pborder.GetOrdersRequest{
		FamilyName: "RepeatedName",
		PageSize:   5,
	}
	response, err := service.GetOrders(ctx, request)
	assert.NotNil(err, "expected an error calling GetOrders")
	assert.Contains(err.Error(), unitTestErrorMessage, "did not see the expected error text calling GetOrders")
	assert.Nil(response, "did not expect a response to a failed GetOrders call")
}

// TestSillyPageSizes looks at how the code handles negative and very large page sizes
func TestSillyPageSizes(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// Request a huge count of orders - we only have 10 so won't get back that many anyway but the logs will tell us
	// that our request was modified
	request := &pborder.GetOrdersRequest{
		GivenName: UnitTestGivenName,
		PageSize:  5000,
	}

	// Wrap the query execution to capture its log output
	var response *pborder.GetOrdersResponse
	var err error
	logged := testutil.CaptureLogging(func() {
		response, err = service.GetOrders(ctx, request)
	})

	// The first page returned should be index entries 1 through 10 of the complete 0 to 9 set
	assert.Nil(err, "did not expect an error calling GetOrders with a large page size (1 of 3)", err)
	assert.Equal(10, len(response.Orders), "expect all 10 orders to be returned (1 of 3)")
	assert.Contains(logged, "excessive page size adjusted to maximum", "did not see logging to say that page size had been reduced")

	// Repeat but with a zero page size
	request.PageSize = 0
	logged = testutil.CaptureLogging(func() {
		response, err = service.GetOrders(ctx, request)
	})

	// The first page returned should be index entries 1 through 10 of the complete 0 to 9 set
	assert.Nil(err, "did not expect an error calling GetOrders with a zero page size (2 of 3)", err)
	assert.Equal(10, len(response.Orders), "expect all 10 orders to be returned (2 of 3)")
	assert.Contains(logged, "negative/zero page size adjusted to default", "did not see logging to say that page size had been increased (from zero)")

	// Repeat but with a negative page size
	request.PageSize = -1
	logged = testutil.CaptureLogging(func() {
		response, err = service.GetOrders(ctx, request)
	})

	// The first page returned should be index entries 1 through 10 of the complete 0 to 9 set
	assert.Nil(err, "did not expect an error calling GetOrders with a negative page size (3 of 3)", err)
	assert.Equal(10, len(response.Orders), "expect all 10 orders to be returned (3 of 3)")
	assert.Contains(logged, "negative/zero page size adjusted to default", "did not see logging to say that page size had been increased (from negative)")
}

// commonTestSetup helps us to be a little DRY (Don't Repeat Yourself) in this file, doing the steps that more
// than half the unit test functions in here need to do before going on to anything else.
func commonTestSetup(t *testing.T) (*require.Assertions, context.Context, *OrderService) {

	// Avoid having to pass t in to every assertion
	assert := require.New(t)

	// Make sure that Firestore has been populated with a known set of orders
	ctx := context.Background()
	primeFirestore(ctx, assert)

	// Obtain a clean instance of the order service
	service, err := NewOrderService()
	assert.Nil(err, "should not have failed asking for an instance of the OrderService: %v", err)

	// return everything the caller needs to perform their tests
	return assert, ctx, service
}

// primeFirestore loads the Firestore emulator with 10 orders the first time it is invoked, thereafter it's a no-op.
func primeFirestore(ctx context.Context, assert *require.Assertions) {

	// Have we been called already? We have nothing to do if so ...
	if primed {
		return
	}
	primed = true

	// Obtain a clean instance of the order service, i.e. one that we know has not been monkeyed with
	// to return errors for testing purposes
	service, err := NewOrderService()
	assert.Nil(err, "did not expect an error obtaining a new OderService: %v", err)

	// Clean out any debris left in the Firestore emulator from previous runs
	deleteAllMockOrders(ctx, service, assert)

	// Populate Firestore with our mock orders for this run
	for _, order := range mockOrders {
		err := service.SaveOrder(ctx, order)
		assert.Nil(err, "did not expect an error storing a mock order: %v", err)
	}
}

// deleteAllMockOrders removes our mock orders from the Firestore emulator. We use this to ensure that our tests are
// run against a clean slate.
func deleteAllMockOrders(ctx context.Context, service *OrderService, assert *require.Assertions) {

	// Eat our own dog food to retrieve all orders that already exist for the UnitTestGivenName and delete them!
	// Using OrderService.GetOrders is not the most efficient way to do walk the orders to be deleted but we
	// might as well exercise our production code rather than have custom test code just for this.

	// Start by forming a pborder.GetOrdersRequest to retrieve the first page of results ...
	request := &pborder.GetOrdersRequest{
		GivenName: UnitTestGivenName,
		PageSize:  10,
	}

	// Loop until we get no more results
	for {
		// Ask for the next page of results
		response, err := service.GetOrders(ctx, request)
		assert.Nil(err, "did not expect an error calling GetOrders: %v", err)
		if len(response.Orders) == 0 {
			return
		}

		// We have some orders to remove
		for _, pbOrder := range response.Orders {

			// Convert the one field we care about to obtain the reference path to our internal Order form
			order := schema.Order{Id: pbOrder.Id}

			// Establish a document reference for that order path and ask for the document to be deleted
			ref := service.FsClient.Doc(order.StoreRefPath())
			_, err = ref.Delete(ctx)
			assert.Nil(err, "failed deleting order ID %s: %v", order.Id, err)

			// Let the impatient engineer watching our progress know who things are going
			zap.L().Info("delete debris order", zap.String("orderId", order.Id))
		}

		// Adjust the request to include tye next page token
		request.PageToken = response.NextPageToken
	}
}

// init performs static initialization of our "constants" that cannot actually be literal constants
func init() {

	// firstOrderTime is used as the submission time of our first mock order and the foundation for all the following
	// order times with an incremental adjustment for each one
	timestamp, _ := types.TimestampFromRFC3339Nano(timeString)
	firstSubmissionTime := timestamp.GetTime()
	timeIncrement := 25 * time.Hour // One day + one hour

	// Define our 10 mock orders
	mockOrders = []*schema.Order{
		&schema.Order{
			Id:              uuid.NewString(),
			SubmissionTime:  firstSubmissionTime,
			OrderedBy:       &types.Person{Id: uuid.NewString(), FamilyName: "One", GivenName: UnitTestGivenName},
			DeliveryAddress: &types.PostalAddress{Locality: "1st"},
			OrderItems: []*schema.OrderItem{
				&schema.OrderItem{Id: uuid.NewString(), ProductCode: "first_1", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 1, Nanos: 100_000_000}},
				&schema.OrderItem{Id: uuid.NewString(), ProductCode: "first_2", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 1, Nanos: 200_000_000}},
			},
		},
		&schema.Order{ // This one entry is fully populated to confirm that all fields are written and retrieved
			Id:             uuid.NewString(),
			SubmissionTime: firstSubmissionTime.Add(timeIncrement * 1),
			OrderedBy: &types.Person{
				Id:          personId,
				FamilyName:  personFamilyName,
				GivenName:   UnitTestGivenName,
				MiddleName:  personMiddleName,
				DisplayName: personDisplayName,
			},
			DeliveryAddress: &types.PostalAddress{
				RegionCode:   addrRegionCode,
				PostalCode:   addrPostalCode,
				Locality:     addrLocality,
				AddressLines: []string{addrLine1, addrLine2},
			},
			OrderItems: []*schema.OrderItem{
				&schema.OrderItem{Id: uuid.NewString(), ProductCode: "second_1", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 2, Nanos: 100_000_000}},
				&schema.OrderItem{Id: uuid.NewString(), ProductCode: "second_2", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 2, Nanos: 200_000_000}},
			},
		},
		&schema.Order{
			Id:              uuid.NewString(),
			SubmissionTime:  firstSubmissionTime.Add(timeIncrement * 2),
			OrderedBy:       &types.Person{Id: uuid.NewString(), FamilyName: "Three", GivenName: UnitTestGivenName},
			DeliveryAddress: &types.PostalAddress{Locality: "3rd"},
			OrderItems: []*schema.OrderItem{
				&schema.OrderItem{Id: uuid.NewString(), ProductCode: "third_1", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 3, Nanos: 100_000_000}},
				&schema.OrderItem{Id: uuid.NewString(), ProductCode: "third_2", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 3, Nanos: 200_000_000}},
			},
		},
		&schema.Order{
			Id:              uuid.NewString(),
			SubmissionTime:  firstSubmissionTime.Add(timeIncrement * 3),
			OrderedBy:       &types.Person{Id: uuid.NewString(), FamilyName: "RepeatedName", GivenName: UnitTestGivenName},
			DeliveryAddress: &types.PostalAddress{Locality: "4th"},
			OrderItems: []*schema.OrderItem{
				&schema.OrderItem{Id: uuid.NewString(), ProductCode: "fourth_1", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 4, Nanos: 100_000_000}},
				&schema.OrderItem{Id: uuid.NewString(), ProductCode: "fourth_2", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 4, Nanos: 200_000_000}},
			},
		},
		&schema.Order{
			Id:              uuid.NewString(),
			SubmissionTime:  firstSubmissionTime.Add(timeIncrement * 4),
			OrderedBy:       &types.Person{Id: uuid.NewString(), FamilyName: "Five", GivenName: UnitTestGivenName},
			DeliveryAddress: &types.PostalAddress{Locality: "5th"},
			OrderItems: []*schema.OrderItem{
				&schema.OrderItem{Id: uuid.NewString(), ProductCode: "fifth_1", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 5, Nanos: 100_000_000}},
				&schema.OrderItem{Id: uuid.NewString(), ProductCode: "fifth_2", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 5, Nanos: 200_000_000}},
			},
		},
		&schema.Order{
			Id:              uuid.NewString(),
			SubmissionTime:  firstSubmissionTime.Add(timeIncrement * 5),
			OrderedBy:       &types.Person{Id: uuid.NewString(), FamilyName: "RepeatedName", GivenName: UnitTestGivenName},
			DeliveryAddress: &types.PostalAddress{Locality: "6th"},
			OrderItems: []*schema.OrderItem{
				&schema.OrderItem{Id: uuid.NewString(), ProductCode: "sixth_1", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 6, Nanos: 100_000_000}},
				&schema.OrderItem{Id: uuid.NewString(), ProductCode: "sixth_2", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 6, Nanos: 200_000_000}},
			},
		},
		&schema.Order{
			Id:              uuid.NewString(),
			SubmissionTime:  firstSubmissionTime.Add(timeIncrement * 6),
			OrderedBy:       &types.Person{Id: uuid.NewString(), FamilyName: "Seven", GivenName: UnitTestGivenName},
			DeliveryAddress: &types.PostalAddress{Locality: "7th"},
			OrderItems: []*schema.OrderItem{
				&schema.OrderItem{Id: uuid.NewString(), ProductCode: "seventh_1", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 7, Nanos: 100_000_000}},
				&schema.OrderItem{Id: uuid.NewString(), ProductCode: "seventh_2", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 7, Nanos: 200_000_000}},
			},
		},
		&schema.Order{
			Id:              uuid.NewString(),
			SubmissionTime:  firstSubmissionTime.Add(timeIncrement * 7),
			OrderedBy:       &types.Person{Id: uuid.NewString(), FamilyName: "Eight", GivenName: UnitTestGivenName},
			DeliveryAddress: &types.PostalAddress{Locality: "8th"},
			OrderItems: []*schema.OrderItem{
				&schema.OrderItem{Id: uuid.NewString(), ProductCode: "eighth_1", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 8, Nanos: 100_000_000}},
				&schema.OrderItem{Id: uuid.NewString(), ProductCode: "eighth_2", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 8, Nanos: 200_000_000}},
			},
		},
		&schema.Order{
			Id:              uuid.NewString(),
			SubmissionTime:  firstSubmissionTime.Add(timeIncrement * 8),
			OrderedBy:       &types.Person{Id: uuid.NewString(), FamilyName: "Nine", GivenName: UnitTestGivenName},
			DeliveryAddress: &types.PostalAddress{Locality: "9th"},
			OrderItems: []*schema.OrderItem{
				&schema.OrderItem{Id: uuid.NewString(), ProductCode: "ninth_1", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 9, Nanos: 100_000_000}},
				&schema.OrderItem{Id: uuid.NewString(), ProductCode: "ninth_2", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 9, Nanos: 200_000_000}},
			},
		},
		&schema.Order{
			Id:              uuid.NewString(),
			SubmissionTime:  firstSubmissionTime.Add(timeIncrement * 9),
			OrderedBy:       &types.Person{Id: uuid.NewString(), FamilyName: "Ten", GivenName: UnitTestGivenName},
			DeliveryAddress: &types.PostalAddress{Locality: "10th"},
			OrderItems: []*schema.OrderItem{
				&schema.OrderItem{Id: uuid.NewString(), ProductCode: "tenth_1", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 10, Nanos: 100_000_000}},
				&schema.OrderItem{Id: uuid.NewString(), ProductCode: "tenth_2", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 10, Nanos: 200_000_000}},
			},
		},
	}
}
