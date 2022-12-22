package ordertofulfill

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/mikebway/poc-gcp-ecomm/fulfillment/schema"
	pbfulfillment "github.com/mikebway/poc-gcp-ecomm/pb/fulfillment"
	"github.com/mikebway/poc-gcp-ecomm/testutil"
	"go.uber.org/zap"
	pubsubapi "google.golang.org/api/pubsub/v1"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/mikebway/poc-gcp-ecomm/fulfillment/service"
	ord "github.com/mikebway/poc-gcp-ecomm/order/schema"

	"github.com/mikebway/poc-gcp-ecomm/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	// EnvFirestoreEmulator defines the environment variable name that is used to convey that the Firestore emulator
	// is running, should be used, and how to connect to it
	EnvFirestoreEmulator = "FIRESTORE_EMULATOR_HOST"

	// FirestoreEmulatorHost defines the server name and port (in TCP6 terms) of the Firestore emulator
	FirestoreEmulatorHost = "[::1]:8219"

	// A timestamp string we can use to derive known time values
	timeString = "2022-10-29T16:23:19.123456789-06:00"

	// A UUID string value that we can use as order OD in our tests
	orderId = "d1cecab3-5bc0-43d4-aef1-99ad69794313"

	// Define the person fields that we will use multiple times to define the person that ordered the items
	orderedById          = "10615145-2010-4c5f-8347-2bb556232c31"
	orderedByFamilyName  = "Grint"
	orderedByGivenName   = "Rupert"
	orderedByMiddleName  = "Alexander Lloyd"
	orderedByDisplayName = "Rupert"

	// Define the postal address fields for the home address of our mock orderedBy person
	addrLine1      = "55 Yonder St"
	addrLine2      = "Flat B"
	addrLocality   = "Ottery St Catchpole"
	addrPostalCode = "EX11 1HF"
	addrRegionCode = "GB"

	// Define shopping order item field values for two shopping order items
	// As of October 2022, the product codes are for an Apple Mac Studio and Studio Display :-)
	itemId1               = "54f34cb9-fea6-4786-a475-cebd95d93742"
	itemId2               = "b719efe9-4189-453c-96b0-7e229520d316"
	itemProdCode1         = "gold_yoyo"
	itemProdCode2         = "plastic_yoyo"
	itemQuantity1         = 1
	itemQuantity2         = 2
	itemPriceCurrencyCode = "USD"
	itemPriceUnits1       = 1899
	itemPriceNanos1       = 550_000_000
	itemPriceUnits2       = 2
	itemPriceNanos2       = 990_000_000
)

var (
	// Order value types that cannot be declared as constants
	orderSubmissionTime time.Time

	// Shopping order item value types that cannot be declared as constants. Since the
	itemPrice1 = types.Money{CurrencyCode: itemPriceCurrencyCode, Units: itemPriceUnits1, Nanos: itemPriceNanos1}
	itemPrice2 = types.Money{CurrencyCode: itemPriceCurrencyCode, Units: itemPriceUnits2, Nanos: itemPriceNanos2}
)

// TestMain, if defined (it's optional), allows setup code to be run before and after the suite of unit tests
// for this package.
func TestMain(m *testing.M) {

	// Ensure that our Firestore requests do not get routed to the live project by mistake
	service.ProjectId = "demo-" + service.ProjectId

	// Configure the environment variable that informs the Firestore client that it should connect to the
	// emulator and how to reach it.
	_ = os.Setenv(EnvFirestoreEmulator, FirestoreEmulatorHost)

	// Shopping order values
	t, _ := types.TimestampFromRFC3339Nano(timeString)
	orderSubmissionTime = t.GetTime()

	// Run all the unit tests
	m.Run()
}

// TestOrderToFulfillHappyPath exercises the main handler function with good data that should be processed
// without error.
func TestOrderToFulfillHappyPath(t *testing.T) {

	// Record the time at which we are starting this test to compare with the submission times of the tasks we generate
	testStartTime := time.Now()

	// do the common setup that we share with some other tests, this includes deleting all tasks
	// written to Firestore by others
	req, ctx, svc := commonTestSetup(t)

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := httptest.NewRequest("POST", "/", buildPushRequest(mockOrderPB()))
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		OrderToFulfill(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(responseRecorder.Code, http.StatusCreated, "should have a 200 OK response code")
	req.Contains(logged, "established tasks for order", "should have seen the expected completion message in the logs")
	req.Contains(logged, "\"orderId\": \"d1cecab3-5bc0-43d4-aef1-99ad69794313\"", "should have seen the expected order ID in the logs")

	// Confirm that two tasks were written to Firestore
	request := &pbfulfillment.GetTasksRequest{
		OrderId:  orderId,
		PageSize: 10,
	}
	response, err := svc.GetTasks(ctx, request)
	req.Nil(err, "did not expect an error calling GetTasks: %v", err)

	// We expect three tasks to have been created
	tasks := response.Tasks
	req.Equal(3, len(tasks), "expected number of tasks was not created")

	// Confirming that the tasks are the ones we expect is a little tricky since we can't be certain of their sort order
	validateTask(req, findTask(req, tasks, "manufacture"), testStartTime, orderId, itemId1, itemProdCode1, pbfulfillment.TaskStatus_WAITING_SERVICE, "")
	validateTask(req, findTask(req, tasks, "ship"), testStartTime, orderId, itemId1, itemProdCode1, pbfulfillment.TaskStatus_WAITING_TASK, "wait_for_manufacture")
	validateTask(req, findTask(req, tasks, "upsell_to_gold"), testStartTime, orderId, itemId2, itemProdCode2, pbfulfillment.TaskStatus_WAITING_CS, "no_stock")
}

// validateTask confirms that the supplied task matches the field values supplied, failing the test if it does not.
func validateTask(req *require.Assertions, task *pbfulfillment.Task, earliestTime time.Time, orderId string, itemId string, product string, status pbfulfillment.TaskStatus, reason string) {
	req.True(earliestTime.Before(task.SubmissionTime.AsTime()) || earliestTime.Equal(task.SubmissionTime.AsTime()), "%s task submission time is wrong", status.String())
	req.Nil(task.CompletionTime, "%s task completion time should not have been set", status.String())
	req.Equal(orderId, task.OrderId, "%s task order ID is wrong", status.String())
	req.Equal(itemId, task.OrderItemId, "%s task order item ID is wrong", status.String())
	req.Equal(product, task.ProductCode, "%s task product code is wrong", status.String())
	req.Equal(status, task.Status, "%s task status is wrong", status.String())
	req.Equal(reason, task.ReasonCode, "%s task reason code is wrong", status.String())
	req.Nil(task.Parameters, "%s task should not have had any parameters", status.String())
}

// findTask looks through the supplied slice of tasks for one that matches the given taskCode and returns that.
// If no match is found, the test will be failed.
func findTask(req *require.Assertions, tasks []*pbfulfillment.Task, taskCode string) *pbfulfillment.Task {
	for _, task := range tasks {
		if task.TaskCode == taskCode {
			return task
		}
	}

	// The expected task was not found - fail the test
	req.Fail("task not found: %s", taskCode)
	return nil
}

// TestInvalidPushRequest exercises the main handler function with an invalid request that does not
// match a Pub/Sub push.
func TestInvalidPushRequest(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := httptest.NewRequest("POST", "/", strings.NewReader("this is not a valid push request"))
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		OrderToFulfill(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(responseRecorder.Code, http.StatusBadRequest, "should have a 400 Bad Request response code")
	req.Contains(responseRecorder.Body.String(), "could not decode push request json body", "should have seen the expected invalid push request message in the response")
	req.Contains(logged, "could not decode push request json body", "should have seen the expected invalid base64 encoding message in the logs")
}

// TestInvalidBase64 exercises the main handler function with bad data that should result in an error response.
func TestInvalidBase64(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := httptest.NewRequest("POST", "/", buildPushRequestFromString("this is not a valid base64 payload"))
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		OrderToFulfill(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(responseRecorder.Code, http.StatusBadRequest, "should have a 400 Bad Request response code")
	req.Contains(responseRecorder.Body.String(), "unable to decode base64 data", "should have seen the expected invalid push request message in the response")
	req.Contains(logged, "unable to decode base64 data", "should have seen the expected invalid base64 encoding message in the logs")
}

// TestWrongBinary looks at what happens when a valid base64 string is passed to the order loader but
// is not a shopping order message.
func TestWrongBinary(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := httptest.NewRequest("POST", "/", buildPushRequest([]byte("this is not a valid protobuf shopping order")))
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		OrderToFulfill(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(responseRecorder.Code, http.StatusBadRequest, "should have a 400 Bad Request response code")
	req.Contains(responseRecorder.Body.String(), "failed to unmarshal order protobuf message", "should have seen the expected protobuf unmarshal error in the response")
	req.Contains(logged, "failed to unmarshal order protobuf message", "should have seen the expected protobuf unmarshal error in the logs")
}

// TestBodyReaderError looks at what happens in the unlikely event that the request.Body cannot be read. Perhaps
// this could happen if TCP error (connection lost?) happened in the middle of reading the body of the POST.
func TestBodyReaderError(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// We don't bother generating a protocol buffer shopping order, we use a bad reader instead
	badReader := &BadReader{}

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := httptest.NewRequest("POST", "/", badReader)
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		OrderToFulfill(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(responseRecorder.Code, http.StatusBadRequest, "should have a 400 Bad Request response code")
	req.Contains(responseRecorder.Body.String(), "could not decode push request json body: i am a bad reader", "should have seen the expected read failure error in the response")
	req.Contains(logged, "could not decode push request json body: i am a bad reader", "should have seen the expected read failure error in the logs")
}

// TestServiceLoadFailure looks at how the code handles being unable to establish a service.FulfillmentService.
func TestServiceLoadFailure(t *testing.T) {

	// Force NewFulfillmentService to fail - be sure to clear that after we are done
	lazyFulfillmentService = nil
	service.UnitTestNewFulfillmentServiceError = errors.New("unit test forced error")
	defer func() { service.UnitTestNewFulfillmentServiceError = nil }()

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := httptest.NewRequest("POST", "/", buildPushRequest(mockOrderPB()))
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		OrderToFulfill(responseRecorder, httpRequest)
	})

	// Confirm the result was the sad one that we expected
	req := require.New(t)
	req.Equal(http.StatusInternalServerError, responseRecorder.Code, "should have a 500 internal server error code")
	req.Contains(logged, "could not obtain firestore client", "should have seen the expected firestore client failure message in the logs")
	req.Contains(logged, service.UnitTestNewFulfillmentServiceError.Error(), "should have seen the expected error message in the logs")
}

// TestSaveTasksFailure tricks the handler service.FulfillmentService SaveTasks function into failing
// by messing with the service port number.
func TestSaveTasksFailure(t *testing.T) {

	// Put everything back when it should be when we leave this test
	defer func() {
		lazyFulfillmentService = nil
		_ = os.Setenv(EnvFirestoreEmulator, FirestoreEmulatorHost)
	}()

	// Force the handler to obtain a new fulfillment service ...
	lazyFulfillmentService = nil

	// ... but without using the emulator and targeting a non-existent project
	_ = os.Setenv(EnvFirestoreEmulator, "")

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := httptest.NewRequest("POST", "/", buildPushRequest(mockOrderPB()))
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		OrderToFulfill(responseRecorder, httpRequest)
	})

	// Confirm the result was the sad one that we expected
	req := require.New(t)
	req.Equal(http.StatusInternalServerError, responseRecorder.Code, "should have a 500 internal server error code")
	req.Contains(logged, "task save transaction failed", "should have seen the expected SaveTasks failure message in the logs")
}

// commonTestSetup helps us to be a little DRY (Don't Repeat Yourself) in this file, doing the steps that more
// several of the unit test functions in here need to do before going on to anything else.
func commonTestSetup(t *testing.T) (*require.Assertions, context.Context, *service.FulfillmentService) {

	// Avoid having to pass t in to every assertion
	assert := require.New(t)

	// Clear any manipulations that might have been made to the fulfillment service used by the handled
	lazyFulfillmentService = nil

	// Make sure that Firestore has been populated with a known set of tasks
	ctx := context.Background()
	svc := deleteAllMockTasks(ctx, assert)

	// return everything the caller needs to perform their tests
	return assert, ctx, svc
}

// deleteAllMockTasks removes our mock tasks from the Firestore emulator. We use this to ensure that our tests are
// run against a clean slate. Returns the fulfillment service used to delete the tasks for the caller to use for
// additional tinkering.
func deleteAllMockTasks(ctx context.Context, assert *require.Assertions) *service.FulfillmentService {

	// Obtain a clean instance of the fulfillment service, i.e. one that we know has not been monkeyed with
	// to return errors for testing purposes
	svc, err := service.NewFulfillmentService()
	assert.Nil(err, "did not expect an error obtaining a new FulfillmentService: %v", err)

	// Start by forming a pbfulfillment.GetTasksRequest to retrieve the first page of results ...
	request := &pbfulfillment.GetTasksRequest{
		OrderId:  orderId,
		PageSize: 10,
	}

	// Loop until we get no more results
	for {
		// Ask for the next page of results
		response, err := svc.GetTasks(ctx, request)
		assert.Nil(err, "did not expect an error calling GetTasks: %v", err)
		if len(response.Tasks) == 0 {
			// Np more pages - we are done
			break
		}

		// We have some tasks to remove
		for _, pbTask := range response.Tasks {

			// Convert the one field we care about to obtain the reference path to our internal Order form
			task := schema.Task{Id: pbTask.Id}

			// Establish a document reference for that task path and ask for the document to be deleted
			ref := svc.FsClient.Doc(task.StoreRefPath())
			_, err = ref.Delete(ctx)
			assert.Nil(err, "failed deleting task ID %s: %v", task.Id, err)

			// Let the impatient engineer watching our progress know who things are going
			zap.L().Info("delete debris task", zap.String("taskId", task.Id))
		}

		// Adjust the request to include tye next page token
		request.PageToken = response.NextPageToken
	}

	// Return the fulfillment service for additional use by our caller
	return svc
}

// BadReader implements the io.Reader interface but deliberately fails every time anyone tries to read from it.
type BadReader struct {
	io.Reader
}

// Read is BadReader being bad at reading but good at helping to test how read failures are handled.
func (r BadReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("i am a bad reader")
}

// buildPushRequest wraps the provided data bytes as the data payload of a push request, returning that as byte reader.
func buildPushRequest(data []byte) io.Reader {

	// Encode the data as base64
	b64 := base64.StdEncoding.EncodeToString(data)

	// Have our sibling to the rest
	return buildPushRequestFromString(b64)
}

// buildPushRequestFromString wraps the provided data string as the data payload of a push request, returning that as
// a byte reader. A valid string would be base64 encoded data but this function can be used to test what happens
// if the data is not base64 too :-)
func buildPushRequestFromString(data string) io.Reader {

	// Wrap that in a push request message structure
	pushReq := &pushRequest{
		Message: pubsubapi.PubsubMessage{
			Data: data,
		},
	}

	// Marshal the push request as JSON string
	jsonBytes, _ := json.Marshal(pushReq)

	// Return a readr om those bytes
	return bytes.NewReader(jsonBytes)
}

// mockOrderPB returns a protobuf binary bytes slice populated with a checked out shopping order structure as
// its value and with the update time being later than its creation time.
func mockOrderPB() []byte {

	// Get a "checked out" shopping order in its internal memory form
	order := buildMockOrder()

	// Render that to its protocol buffer structure form
	pbOrder := order.AsPBOrder()

	// Marshal that to its binary message state and return that
	pbBytes, _ := proto.Marshal(pbOrder)
	return pbBytes
}

// buildMockOrder returns a types.Order structure populated with a orderedBy that can be used to
// test storing new orders in our tests.
func buildMockOrder() *ord.Order {
	return &ord.Order{
		Id:              orderId,
		SubmissionTime:  orderSubmissionTime,
		OrderedBy:       buildMockOrderedBy(),
		DeliveryAddress: buildMockDeliveryAddress(),
		OrderItems:      buildMockOrderItems(),
	}
}

// buildMockOrderedBy returns a types.Person structure populated with the orderedBy constant attributes
// defined at the head of this file to be used to create new orders in our tests.
func buildMockOrderedBy() *types.Person {
	return &types.Person{
		Id:          orderedById,
		FamilyName:  orderedByFamilyName,
		GivenName:   orderedByGivenName,
		MiddleName:  orderedByMiddleName,
		DisplayName: orderedByDisplayName,
	}
}

// buildMockDeliveryAddress returns a types.PostalAddress structure populated with the constant
// attributes defined at the head of this file to be used to create new orders in our tests.
func buildMockDeliveryAddress() *types.PostalAddress {
	return &types.PostalAddress{
		RegionCode:   addrRegionCode,
		PostalCode:   addrPostalCode,
		Locality:     addrLocality,
		AddressLines: []string{addrLine1, addrLine2},
	}
}

// buildMockOrderItems returns an array containing two shopping order items populated  with the constant
// attributes defined at the head of this file to be used to create new orders in our tests.
func buildMockOrderItems() []*ord.OrderItem {
	return []*ord.OrderItem{
		&ord.OrderItem{
			Id:          itemId1,
			ProductCode: itemProdCode1,
			Quantity:    itemQuantity1,
			UnitPrice:   &itemPrice1,
		},
		&ord.OrderItem{
			Id:          itemId2,
			ProductCode: itemProdCode2,
			Quantity:    itemQuantity2,
			UnitPrice:   &itemPrice2,
		},
	}
}
