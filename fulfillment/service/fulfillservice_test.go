package service

import (
	"cloud.google.com/go/firestore"
	"errors"
	"fmt"
	"github.com/google/uuid"
	cartsvc "github.com/mikebway/poc-gcp-ecomm/cart/service"
	"github.com/mikebway/poc-gcp-ecomm/fulfillment/schema"
	pbfulfillment "github.com/mikebway/poc-gcp-ecomm/pb/fulfillment"
	"github.com/mikebway/poc-gcp-ecomm/testutil"
	"github.com/mikebway/poc-gcp-ecomm/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"strconv"
	"testing"
	"time"
)

const (
	// EnvFirestoreEmulator defines the environment variable name that is used to convey that the Firestore emulator
	// is running, should be used, and how to connect to it
	EnvFirestoreEmulator = "FIRESTORE_EMULATOR_HOST"

	// FirestoreEmulatorHost defines the server name and port (in TCP6 terms) of the Firestore emulator
	FirestoreEmulatorHost = "[::1]:8219"

	// A timestamp string we can use to derive known time values. It will be used as the first submission time in the
	// series of 12 tasks that primeFirestore adds to the Firestore emulator
	//
	// Firestore time representation when evaluating query matches is not as accurate as a full 9 digits
	// of nanoseconds, so we zero those out. It has better than second accuracy, but I am not going to
	// bother figuring out exactly how much better for the purposes of the TestSubmissionTimeQuery test.
	timeString = "2022-12-11T10:00:00.000000000-06:00"

	// ItemUUIDPrefix is used as the first 35 of 36 characters in order item UUIDS that our mock tasks will reference.
	// for all order IDs for all unit test tasks written.
	ItemUUIDPrefix = "11111111-1111-4111-1111-00000000000"

	// ProductCodePrefix is used as the root of a numbered series of product codes used in our mock task data
	ProductCodePrefix = "product_code_"

	// ReasonCodePrefix is used as the root of a numbered series of status reason codes used in our mock task data
	ReasonCodePrefix = "reason_code_"

	// OrderID is the UUID ID of all the order with which all the tasks that we create in our unit tests will be
	// associated. It is used so that we can find and delete all tasks we create in these unit
	// tests (and only those tasks) so that we can be sure of starting with clean slate.
	OrderID = "1eef2132-d736-48f4-89cd-a1423148b527"

	// Named parameter values
	paramName1 = "sequenceNumber"
	paramName2 = "productNumber"

	// unitTestErrorMessage is used as the error description for error that are deliberately forced to
	// test error handling.
	unitTestErrorMessage = "unit test of error handling"
)

var (
	// primed is true if primeFirestore() has already populated the Firestore emulator with mock tasks
	primed = false

	// mockTasks is a slice of 12 tasks that the Firestore emulator is primed with for all our tests.
	mockTasks []*schema.Task
)

// UTQueryExecProxy implements a wrapper function around firestore.Query that will return an iterator over items
// that match the query. For unit test purposes, this version always returns errors when trying to iterate over
// the result set.
type UTQueryExecProxy struct {
	cartsvc.QueryExecutionProxy
}

// UTDocIteratorProxy is a unit test implementation of the DocumentIteratorProxy interface that always returns
// errors when trying to iterate over the result set.
type UTDocIteratorProxy struct {
	cartsvc.DocumentIteratorProxy
}

// Documents returns a DocumentIteratorProxy wrapping the results of the given query. Unit test implementations
// of this function can be programmed to return an iterator that can insert errors into the flow but this
// production ready implementation returns a transparent passthrough iterator.
func (q *UTQueryExecProxy) Documents(ctx context.Context, query firestore.Query) cartsvc.DocumentIteratorProxy {
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
	cartsvc.DocumentRefProxy
}

// Create is a pass through to the firestore.DocumentRef Create function that allows
// unit tests to have Firestore operations return errors.
func (p *UTDocRefProxy) Create(doc *firestore.DocumentRef, ctx context.Context, data interface{}) (*firestore.WriteResult, error) {
	return nil, errors.New(unitTestErrorMessage)
}

// TransactionalCreate is a pass through to the firestore.Transaction Create function that allows
// unit tests to have Firestore operations return errors.
func (p *UTDocRefProxy) TransactionalCreate(doc *firestore.DocumentRef, tx *firestore.Transaction, data interface{}) error {
	return errors.New(unitTestErrorMessage)
}

// Update is a pass through to the firestore.DocumentRef Update function  that allows
// unit tests to have Firestore operations return errors.
func (p *UTDocRefProxy) Update(doc *firestore.DocumentRef, ctx context.Context, updates []firestore.Update) (*firestore.WriteResult, error) {
	return nil, errors.New(unitTestErrorMessage)
}

// UTDocSnapProxy is a unit test implementation of the DocumentSnapshotProxy interface that allows
// unit tests to have Firestore operations return errors.
type UTDocSnapProxy struct {
	cartsvc.DocumentRefProxy
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

// TestSaveTaskError covers that part of the FulfillmentService.SaveTask that the 12+ tasks we save successfully
// in these test do not get to: the failures.
func TestSaveTaskError(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// Replace the document reference proxy of the service with one that will behave badly at our direction
	service.drProxy = &UTDocRefProxy{}

	// We don't need much of a task to test an error that we know will be thrown
	task := &schema.Task{
		Id: uuid.NewString(),
	}

	// Attempt to save the task, certain this will fail
	err := service.SaveTasks(ctx, []*schema.Task{task})
	assert.NotNil(err, "should have seen a forced error saving a task")
	assert.Contains(err.Error(), unitTestErrorMessage, "did not see the specific error that we expected")
}

// TestGetTaskByID retrieves one of the mock tasks that primeFirestore has stores in the Firestore emulator.
func TestGetTaskByID(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// Get our last mock task - the one that is more fully populated than all the others since it includes a completion time
	response, err := service.GetTaskByID(ctx, &pbfulfillment.GetTaskByIDRequest{TaskId: mockTasks[11].Id})
	assert.Nil(err, "should not have failed retrieving task ID %s: %v", mockTasks[11].Id, err)
	task := response.Task
	assert.NotNil(task, "did not get an task in the response")

	// Confirm all the values are present as expected
	assert.Equal(mockTasks[11].Id, task.Id, "task ID did not match")
	assert.Equal(mockTasks[11].SubmissionTime.Unix(), task.SubmissionTime.AsTime().Unix(), "submission time does not match")
	assert.Equal(mockTasks[11].CompletionTime.Unix(), task.CompletionTime.AsTime().Unix(), "completion time does not match")
	assert.Equal(mockTasks[11].OrderId, task.OrderId, "order ID does not match")
	assert.Equal(mockTasks[11].OrderItemId, task.OrderItemId, "order item ID does not match")
	assert.Equal(mockTasks[11].ProductCode, task.ProductCode, "product code does not match")
	assert.Equal(schema.COMPLETED, int(task.Status), "products code does not match")
	assert.Equal(mockTasks[11].ReasonCode, task.ReasonCode, "reason code does not match")
	assert.Equal(2, len(task.Parameters), "did not find the expected two named value parameters")
	assert.Equal(paramName1, task.Parameters[0].Name, "parameter name 0 does not match")
	assert.Equal("12", task.Parameters[0].Value, "parameter value 0 does not match")
	assert.Equal(paramName2, task.Parameters[1].Name, "parameter name 1 does not match")
	assert.Equal("4", task.Parameters[1].Value, "parameter value 1 does not match")
}

// TestGetTaskByIDNotFound tries to retrieve a document that does not exist.
func TestGetTaskByIDNotFound(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// Ask for the non-existent task
	response, err := service.GetTaskByID(ctx, &pbfulfillment.GetTaskByIDRequest{TaskId: "no-way-this-exists"})
	assert.NotNil(err, "should have failed retrieving a non-existent task")
	assert.Contains(err.Error(), "failed to retrieve task snapshot with ID no-way-this-exists", "did not see the error we expected")
	assert.Nil(response, "should not have received a response")
}

// TestGetTaskByIDCorrupt tries to retrieve a document that cannot be unmarshalled from its snapshot
func TestGetTaskByIDCorrupt(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// Replace the document reference proxy of the service with one that will behave badly at our direction
	service.dsProxy = &UTDocSnapProxy{}

	// Ask for our second mock task - the one that is more fully populated than all the others but
	// the overriden proxy will reject
	response, err := service.GetTaskByID(ctx, &pbfulfillment.GetTaskByIDRequest{TaskId: mockTasks[1].Id})
	assert.NotNil(err, "should have failed retrieving a non-existent task")
	assert.Contains(err.Error(), "failed to unmarshal task snapshot", "did not see the error we expected")
	assert.Nil(response, "should not have received a response")
}

// TestSubmissionTimeQuery tries out finding multiple tasks that fall within a given time span and paging to boot.
func TestSubmissionTimeQuery(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// Request what we know should be the first 5 of 8 tasks starting with the second one and excluding the tenth
	// of our mock set
	request := &pbfulfillment.GetTasksRequest{
		StartTime: timestamppb.New(mockTasks[1].SubmissionTime), // task[1] should be in the result set
		EndTime:   timestamppb.New(mockTasks[9].SubmissionTime), // task[9] should NOT be in the result set
		PageSize:  5,
	}
	response, err := service.GetTasks(ctx, request)
	assert.Nil(err, "did not expect an error calling GetTasks for the first page: %v", err)

	// The first page returned should be index entries 1 through 5 of the complete0 to 11 set
	tasks := response.Tasks
	assert.Equal(5, len(tasks), "expect 5 tasks in the first full page")
	assert.Equal(mockTasks[1].Id, tasks[0].Id, "task ID 1 of 8 did not match")
	assert.Equal(mockTasks[2].Id, tasks[1].Id, "task ID 2 of 8 did not match")
	assert.Equal(mockTasks[3].Id, tasks[2].Id, "task ID 3 of 8 did not match")
	assert.Equal(mockTasks[4].Id, tasks[3].Id, "task ID 4 of 8 did not match")
	assert.Equal(mockTasks[5].Id, tasks[4].Id, "task ID 5 of 8 did not match")
	assert.NotEmpty(response.NextPageToken, "next page token should have been set after retrieving first page")

	// Update the request to retrieve the second page
	request.PageToken = response.NextPageToken
	response, err = service.GetTasks(ctx, request)
	assert.Nil(err, "did not expect an error calling GetTasks for the second page: %v", err)
	tasks = response.Tasks
	assert.Equal(3, len(tasks), "expect 3 tasks in the second partial page")
	assert.Equal(mockTasks[6].Id, tasks[0].Id, "task ID 6 of 8 did not match")
	assert.Equal(mockTasks[7].Id, tasks[1].Id, "task ID 7 of 8 did not match")
	assert.Equal(mockTasks[8].Id, tasks[2].Id, "task ID 8 of 8 did not match")
	assert.Empty(response.NextPageToken, "next page token should NOT have been set after retrieving second page")
}

// TestOrderItemIdQuery tries out finding multiple tasks that were for the same order item.
func TestOrderItemIdQuery(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// Request what we know should be 3 tasks with the same order item ID of our mock set
	itemId := mockTasks[3].OrderItemId
	request := &pbfulfillment.GetTasksRequest{
		OrderItemId: itemId,
		PageSize:    5,
	}
	response, err := service.GetTasks(ctx, request)
	assert.Nil(err, "did not expect an error calling GetTasks for the first and only page: %v", err)

	// The first page returned should have three entries. There should not be another page.
	tasks := response.Tasks
	assert.Equal(3, len(tasks), "expect 3 tasks in the first and only page")
	assert.Equal(int(schema.WAITING_CS), int(tasks[0].Status), "status 1 of 3 did not match")
	assert.Equal(int(schema.WAITING_SERVICE), int(tasks[1].Status), "status 2 of 3 did not match")
	assert.Equal(int(schema.WAITING_THIRD_PARTY), int(tasks[2].Status), "status 3 of 3 did not match")
	assert.Empty(response.NextPageToken, "next page token should NOT have been set after retrieving first page")
}

// TestOrderIDQuery tries out finding multiple tasks that were for the same order ID.
func TestOrderIDQuery(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// Request what we know should be the first 5 of 12 tasks starting with the second one and excluding the last
	// of our mock set
	request := &pbfulfillment.GetTasksRequest{
		OrderId:  OrderID, // All 12 or our mock tasks have the same order ID
		PageSize: 6,
	}
	response, err := service.GetTasks(ctx, request)
	assert.Nil(err, "did not expect an error calling GetTasks for the first full page: %v", err)

	// The first page returned should have two entries. There should not be another page.
	tasks := response.Tasks
	assert.Equal(6, len(tasks), "expect 5 tasks in the first full page")
	assert.Equal(mockTasks[0].Id, tasks[0].Id, "task ID 1 of 12 did not match")
	assert.Equal(mockTasks[1].Id, tasks[1].Id, "task ID 2 of 12 did not match")
	assert.Equal(mockTasks[2].Id, tasks[2].Id, "task ID 3 of 12 did not match")
	assert.Equal(mockTasks[3].Id, tasks[3].Id, "task ID 4 of 12 did not match")
	assert.Equal(mockTasks[4].Id, tasks[4].Id, "task ID 5 of 12 did not match")
	assert.Equal(mockTasks[5].Id, tasks[5].Id, "task ID 6 of 12 did not match")
	assert.NotEmpty(response.NextPageToken, "next page token should have been set after retrieving first page")

	// Update the request to retrieve the second page
	request.PageToken = response.NextPageToken
	response, err = service.GetTasks(ctx, request)
	assert.Nil(err, "did not expect an error calling GetTasks for the second page: %v", err)
	tasks = response.Tasks
	assert.Equal(6, len(tasks), "expect 5 tasks in the second full page")
	assert.Equal(mockTasks[6].Id, tasks[0].Id, "task ID 7 of 12 did not match")
	assert.Equal(mockTasks[7].Id, tasks[1].Id, "task ID 8 of 12 did not match")
	assert.Equal(mockTasks[8].Id, tasks[2].Id, "task ID 9 of 12 did not match")
	assert.Equal(mockTasks[9].Id, tasks[3].Id, "task ID 10 of 12 did not match")
	assert.Equal(mockTasks[10].Id, tasks[4].Id, "task ID 11 of 12 did not match")
	assert.Equal(mockTasks[11].Id, tasks[5].Id, "task ID 12 of 12 did not match")
	assert.NotEmpty(response.NextPageToken, "next page token should have been set after retrieving second page")

	// Update the request to retrieve the third page
	request.PageToken = response.NextPageToken
	response, err = service.GetTasks(ctx, request)
	assert.Nil(err, "did not expect an error calling GetTasks for the third page: %v", err)
	assert.Equal(0, len(response.Tasks), "expect zero tasks in the third page")
	assert.Empty(response.NextPageToken, "next page token should NOT have been set after retrieving third page")
}

// TestProductCodeQuery tries out finding multiple tasks that were for the same product code
func TestProductCodeQuery(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// Request what we know should be 3 tasks with the samep product code of our mock set
	productCode := mockTasks[10].ProductCode
	request := &pbfulfillment.GetTasksRequest{
		ProductCode: productCode,
		PageSize:    5,
	}
	response, err := service.GetTasks(ctx, request)
	assert.Nil(err, "did not expect an error calling GetTasks for the first and only page: %v", err)

	// The first page returned should have three entries. There should not be another page.
	tasks := response.Tasks
	assert.Equal(3, len(tasks), "expect 3 tasks in the first and only page")
	assert.Equal(int(schema.CANCELED), int(tasks[0].Status), "status 1 of 3 did not match")
	assert.Equal(int(schema.CANCELED), int(tasks[1].Status), "status 2 of 3 did not match")
	assert.Equal(int(schema.COMPLETED), int(tasks[2].Status), "status 3 of 3 did not match")
	assert.Empty(response.NextPageToken, "next page token should NOT have been set after retrieving first page")
}

// TestLoggingAndBadPageToken kills two birds with one stone, verifying that queries are logged for diagnostic
// purposes and looking at how the code handles an invalid "net page token."
func TestLoggingAndBadPageToken(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// For a request with an invalid page token
	request := &pbfulfillment.GetTasksRequest{
		OrderId:     OrderID, // All 12 or our mock tasks have the same order ID
		OrderItemId: "this is not an order item ID but it will do",
		ProductCode: "fake_product",
		PageSize:    5,
		PageToken:   "not_a_number,not_a_uuid",
	}

	// Wrap the query execution to capture its log output
	var response *pbfulfillment.GetTasksResponse
	var err error
	logged := testutil.CaptureLogging(func() {
		response, err = service.GetTasks(ctx, request)
	})

	// We should have see a error complaining about the page token
	assert.NotNil(err, "expected an error calling GetTasks")
	assert.Contains(err.Error(), "invalid page token", "did not get the expected error")
	assert.Nil(response, "should not have received a response")

	// Confirm that the log output described the query
	assert.Contains(logged, "get task", "did not find the query in the log output")
	assert.Contains(logged, "orderId", "log output should name the orderId field")
	assert.Contains(logged, OrderID, "log output should name the orderId value")
	assert.Contains(logged, "orderItemId", "log output should name the orderItemId field")
	assert.Contains(logged, request.OrderItemId, "log output should name the orderItemId value")
	assert.Contains(logged, "productCode", "log output should name the productCode field")
	assert.Contains(logged, request.ProductCode, "log output should name the productCode value")
	assert.Contains(logged, "pageToken", "log output should name the pageToken field")
	assert.Contains(logged, request.PageToken, "log output should contain the duff page token")
}

// TestQueryError looks at how the code handles an error while executing a Firestore query
func TestQueryError(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// Replace the query proxy of the service with one that will behave badly at our direction
	service.queryProxy = &UTQueryExecProxy{}

	// Request what we know should be the first 5 of 12 tasks starting with the second one and excluding the last
	// of our mock set
	request := &pbfulfillment.GetTasksRequest{
		OrderId:  OrderID,
		PageSize: 5,
	}
	response, err := service.GetTasks(ctx, request)
	assert.NotNil(err, "expected an error calling GetTasks")
	assert.Contains(err.Error(), unitTestErrorMessage, "did not see the expected error text calling GetTasks")
	assert.Nil(response, "did not expect a response to a failed GetTasks call")
}

// TestSillyPageSizes looks at how the code handles negative and very large page sizes
func TestSillyPageSizes(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// Request a huge count of tasks - we only have 12 so won't get back that many anyway but the logs will tell us
	// that our request was modified
	request := &pbfulfillment.GetTasksRequest{
		OrderId:  OrderID,
		PageSize: 5000,
	}

	// Wrap the query execution to capture its log output
	var response *pbfulfillment.GetTasksResponse
	var err error
	logged := testutil.CaptureLogging(func() {
		response, err = service.GetTasks(ctx, request)
	})

	// The first page returned should be index entries 1 through 12 of the complete 0 to 11 set
	assert.Nil(err, "did not expect an error calling GetTasks with a large page size (1 of 3)", err)
	assert.Equal(12, len(response.Tasks), "expect all 12 tasks to be returned (1 of 3)")
	assert.Contains(logged, "excessive page size adjusted to maximum", "did not see logging to say that page size had been reduced")

	// Repeat but with a zero page size
	request.PageSize = 0
	logged = testutil.CaptureLogging(func() {
		response, err = service.GetTasks(ctx, request)
	})

	// The first page returned should be index entries 1 through 5 of the complete 0 to 11 set
	assert.Nil(err, "did not expect an error calling GetTasks with a zero page size (2 of 3)", err)
	assert.Equal(12, len(response.Tasks), "expect all 12 tasks to be returned (2 of 3)")
	assert.Contains(logged, "negative/zero page size adjusted to default", "did not see logging to say that page size had been increased (from zero)")

	// Repeat but with a negative page size
	request.PageSize = -1
	logged = testutil.CaptureLogging(func() {
		response, err = service.GetTasks(ctx, request)
	})

	// The first page returned should be index entries 1 through 5 of the complete0 to 11 set
	assert.Nil(err, "did not expect an error calling GetTasks with a negative page size (3 of 3)", err)
	assert.Equal(12, len(response.Tasks), "expect all 12 tasks to be returned (3 of 3)")
	assert.Contains(logged, "negative/zero page size adjusted to default", "did not see logging to say that page size had been increased (from negative)")
}

// TestUpdateTaskStatus looks at the happy path for changing the status of a task
func TestUpdateTaskStatus(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// First, create a task in Firestore that is well away from our other test targets and so won't
	// brake any other test
	targetTask := generateMockTask(1, 1, time.Now(), schema.WAITING_TASK)
	err := service.SaveTasks(ctx, []*schema.Task{targetTask})
	assert.Nil(err, "failed to save update target task")

	// Check that the saved task has the expected status
	getResponse, err := service.GetTaskByID(ctx, &pbfulfillment.GetTaskByIDRequest{TaskId: targetTask.Id})
	assert.Nil(err, "should not have failed retrieving task ID %s: %v", targetTask.Id, err)
	first := getResponse.Task
	assert.NotNil(first, "did not get an task in the first get response")
	assert.Equal(int(schema.WAITING_TASK), int(first.Status), "first response did not have the original status")
	assert.Equal(ReasonCodePrefix+"1", first.ReasonCode, "first response did not have the original reason")

	// Update the status and reason code
	updateRequest := &pbfulfillment.UpdateTaskStatusRequest{
		TaskId:     targetTask.Id,
		Status:     pbfulfillment.TaskStatus_WAITING_CUSTOMER,
		ReasonCode: "they keep changing their mind",
	}
	updateResponse, err := service.UpdateTaskStatus(ctx, updateRequest)
	assert.Nil(err, "should not have failed updating task ID %s: %v", targetTask.Id, err)
	assert.NotNil(updateResponse, "should have received a response updating the task")

	// Fetch the task again to confirm that it has the changes
	getResponse, err = service.GetTaskByID(ctx, &pbfulfillment.GetTaskByIDRequest{TaskId: targetTask.Id})
	assert.Nil(err, "should not have failed retrieving updated task ID %s: %v", targetTask.Id, err)
	second := getResponse.Task
	assert.NotNil(second, "did not get an task in the second get response")
	assert.Equal(int(schema.WAITING_CUSTOMER), int(second.Status), "second response did not have the updated status")
	assert.Equal(updateRequest.ReasonCode, second.ReasonCode, "second response did not have the updated reason")

	// Update the status and reason code a second time - to completed
	updateRequest = &pbfulfillment.UpdateTaskStatusRequest{
		TaskId:     targetTask.Id,
		Status:     pbfulfillment.TaskStatus_COMPLETED,
		ReasonCode: "they made up their mind at last",
	}
	updateResponse, err = service.UpdateTaskStatus(ctx, updateRequest)
	assert.Nil(err, "should not have failed completing task ID %s: %v", targetTask.Id, err)
	assert.NotNil(updateResponse, "should have received a response completing the task")

	// Fetch the task again to confirm that it has the changes
	getResponse, err = service.GetTaskByID(ctx, &pbfulfillment.GetTaskByIDRequest{TaskId: targetTask.Id})
	assert.Nil(err, "should not have failed retrieving completing task ID %s: %v", targetTask.Id, err)
	third := getResponse.Task
	assert.NotNil(third, "did not get an task in the third get response")
	assert.Equal(int(schema.COMPLETED), int(third.Status), "third response did not have the completing status")
	assert.Equal(updateRequest.ReasonCode, third.ReasonCode, "third response did not have the completing reason")
}

// TestUpdateTaskStatusFailure looks at the sad path for changing the status of a task
func TestUpdateTaskStatusFailure(t *testing.T) {

	// Do the common setup that most of our tests require
	assert, ctx, service := commonTestSetup(t)

	// We force a failure by asking to update a task that we know does not exist
	updateRequest := &pbfulfillment.UpdateTaskStatusRequest{
		TaskId:     uuid.NewString(),
		Status:     pbfulfillment.TaskStatus_WAITING_CUSTOMER,
		ReasonCode: "they have not entered an order yet",
	}
	updateResponse, err := service.UpdateTaskStatus(ctx, updateRequest)
	assert.NotNil(err, "should have failed updating task that does not exist")
	assert.Nil(updateResponse, "should not have received a response updating a nonexistent task")
}

// commonTestSetup helps us to be a little DRY (Don't Repeat Yourself) in this file, doing the steps that more
// than half the unit test functions in here need to do before going on to anything else.
func commonTestSetup(t *testing.T) (*require.Assertions, context.Context, *FulfillmentService) {

	// Avoid having to pass t in to every assertion
	assert := require.New(t)

	// Make sure that Firestore has been populated with a known set of tasks
	ctx := context.Background()
	primeFirestore(ctx, assert)

	// Obtain a clean instance of the fulfillment service
	service, err := NewFulfillmentService()
	assert.Nil(err, "should not have failed asking for an instance of the FulfillmentService: %v", err)

	// return everything the caller needs to perform their tests
	return assert, ctx, service
}

// primeFirestore loads the Firestore emulator with 12 tasks the first time it is invoked, thereafter it's a no-op.
func primeFirestore(ctx context.Context, assert *require.Assertions) {

	// Have we been called already? We have nothing to do if so ...
	if primed {
		return
	}
	primed = true

	// Obtain a clean instance of the fulfillment service, i.e. one that we know has not been monkeyed with
	// to return errors for testing purposes
	service, err := NewFulfillmentService()
	assert.Nil(err, "did not expect an error obtaining a new FulfillmentService: %v", err)

	// Clean out any debris left in the Firestore emulator from previous runs
	deleteAllMockTasks(ctx, service, assert)

	// Populate Firestore with our mock tasks for this run
	err = service.SaveTasks(ctx, mockTasks)
	assert.Nil(err, "did not expect an error storing a mock tasks: %v", err)
}

// deleteAllMockTasks removes our mock tasks from the Firestore emulator. We use this to ensure that our tests are
// run against a clean slate.
func deleteAllMockTasks(ctx context.Context, service *FulfillmentService, assert *require.Assertions) {

	// Eat our own dog food to retrieve all tasks that already exist for the UnitTestGivenName and delete them!
	// Using FulfillmentService.GetTasks is not the most efficient way to do walk the tasks to be deleted but we
	// might as well exercise our production code rather than have custom test code just for this.

	// Start by forming a pbfulfillment.GetTasksRequest to retrieve the first page of results ...
	request := &pbfulfillment.GetTasksRequest{
		OrderId:  OrderID,
		PageSize: 12,
	}

	// Loop until we get no more results
	for {
		// Ask for the next page of results
		response, err := service.GetTasks(ctx, request)
		assert.Nil(err, "did not expect an error calling GetTasks: %v", err)
		if len(response.Tasks) == 0 {
			return
		}

		// We have some tasks to remove
		for _, pbTask := range response.Tasks {

			// Convert the one field we care about to obtain the reference path to our internal Order form
			task := schema.Task{Id: pbTask.Id}

			// Establish a document reference for that task path and ask for the document to be deleted
			ref := service.FsClient.Doc(task.StoreRefPath())
			_, err = ref.Delete(ctx)
			assert.Nil(err, "failed deleting task ID %s: %v", task.Id, err)

			// Let the impatient engineer watching our progress know who things are going
			zap.L().Info("delete debris task", zap.String("taskId", task.Id))
		}

		// Adjust the request to include tye next page token
		request.PageToken = response.NextPageToken
	}
}

// init performs static initialization of our "constants" that cannot actually be literal constants
func init() {

	// timeString is used as the submission time of our first mock task and the foundation for all the following
	// task times with an incremental adjustment for each one
	timestamp, _ := types.TimestampFromRFC3339Nano(timeString)
	firstSubmissionTime := timestamp.GetTime()

	// Define our 12 mock tasks
	mockTasks = []*schema.Task{
		generateMockTask(1, 1, firstSubmissionTime, schema.WAITING_TASK),
		generateMockTask(2, 1, firstSubmissionTime, schema.WAITING_CUSTOMER),
		generateMockTask(3, 1, firstSubmissionTime, schema.WAITING_PAYMENT),

		generateMockTask(4, 2, firstSubmissionTime, schema.WAITING_CS),
		generateMockTask(5, 2, firstSubmissionTime, schema.WAITING_SERVICE),
		generateMockTask(6, 2, firstSubmissionTime, schema.WAITING_THIRD_PARTY),

		generateMockTask(7, 3, firstSubmissionTime, schema.PAUSED),
		generateMockTask(8, 3, firstSubmissionTime, schema.PAUSED),
		generateMockTask(9, 3, firstSubmissionTime, schema.PAUSED),

		generateMockTask(10, 4, firstSubmissionTime, schema.CANCELED),
		generateMockTask(11, 4, firstSubmissionTime, schema.CANCELED),
		generateMockTask(12, 4, firstSubmissionTime, schema.COMPLETED),
	}

	// Set the completion time on the last task in our set
	mockTasks[11].CompletionTime = mockTasks[11].SubmissionTime.Add(time.Hour * 24)
}

func generateMockTask(sequenceNumber int, productNumber int, firstSubmissionTime time.Time, status schema.TaskStatus) *schema.Task {

	// We use the UUID ID of he task in two places so generate its value here
	taskId := uuid.NewString()

	// Convert the numbers to strings that we can append to our prefixes
	itemSuffix := fmt.Sprintf("%x", productNumber)
	productSuffix := strconv.Itoa(productNumber)

	// Form a unique submission time for this task
	submissionTime := firstSubmissionTime.Add(time.Duration(int64(sequenceNumber-1) * int64(time.Second)))

	return &schema.Task{
		Id:             taskId,
		SubmissionTime: submissionTime,
		OrderId:        OrderID,
		OrderItemId:    ItemUUIDPrefix + itemSuffix,
		ProductCode:    ProductCodePrefix + productSuffix,
		Status:         status,
		ReasonCode:     ReasonCodePrefix + itemSuffix,
		Parameters: []*schema.Parameter{
			{Name: paramName1, Value: strconv.Itoa(sequenceNumber)},
			{Name: paramName2, Value: strconv.Itoa(productNumber)},
		},
	}
}
