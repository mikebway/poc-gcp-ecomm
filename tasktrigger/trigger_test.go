package tasktrigger

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/mikebway/poc-gcp-ecomm/fulfillment/fulfillapi"
	"github.com/mikebway/poc-gcp-ecomm/fulfillment/schema"
	pbfulfillment "github.com/mikebway/poc-gcp-ecomm/pb/fulfillment"
	"github.com/mikebway/poc-gcp-ecomm/testutil"
	"github.com/mikebway/poc-gcp-ecomm/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TODO: Consolidate repeated unit test support data and functions into a shared module

const (
	// EnvFirestoreEmulator defines the environment variable name that is used to convey that the Firestore emulator
	// is running, should be used, and how to connect to it
	EnvFirestoreEmulator = "FIRESTORE_EMULATOR_HOST"

	// FirestoreEmulatorHost defines the server name and port (in TCP6 terms) of the Firestore emulator
	FirestoreEmulatorHost = "[::1]:8219"

	// EnvPubSubEmulator defines the environment variable name that is used to convey that the Pub/Sub emulator
	// is running, should be used, and how to connect to it
	EnvPubSubEmulator = "PUBSUB_EMULATOR_HOST"

	// PubSubEmulatorHost defines the server name and port (in TCP6 terms) of the Pub/Sub emulator
	PubSubEmulatorHost = "[::1]:8085"

	// EnvPubSubProjectId defines the environment variable name that is used to convey which project the
	// Pub/Sub emulator believes itself to be running under
	EnvPubSubProjectId = "PUBSUB_PROJECT_ID"

	// timeString is time string we can use to derive known time values. It will be used as the first submission time
	// in any tasks we create.
	//
	// Firestore time representation when evaluating query matches is not as accurate as a full 9 digits
	// of nanoseconds, so we zero those out. It has better than second accuracy, but I am not going to
	// bother figuring out exactly how much better for the purposes of the TestSubmissionTimeQuery test.
	timeString = "2022-12-11T10:00:00.000000000-06:00"

	// ItemId is used as the order item UUID  in all our mock task data
	ItemId = "11111111-1111-4111-1111-000000000001"

	// ProductCode is used as the product code in all our mock task data
	ProductCode = "halibut"

	// ReasonCode is used as the reason code in all our mock task data
	ReasonCode = "so-it-goes"

	// OrderID is the UUID ID of all the order with which all the tasks that we create in our unit tests will be
	// associated. It is used so that we can find and delete all tasks we create in these unit
	// tests (and only those tasks) so that we can be sure of starting with clean slate.
	OrderID = "1eef2132-d736-48f4-89cd-a1423148b527"

	// Named parameter values
	paramName1  = "first"
	paramName2  = "second"
	paramValue1 = "Tina"
	paramValue2 = "Mike"
)

var (
	// fulfillmentService allows unit tests to write populated tasks to Firestore
	fulfillmentService *fulfillapi.FulfillmentService

	// The time our mock task was submitted
	taskSubmissionTime time.Time

	// Firestore value times
	firestoreValueCreateTime time.Time

	// storedTaskId is the ID of a task that we have written to Firestore so that it can be referenced
	// in multiple unit tests rather than creating new tasks every time.
	storedTaskId string
)

// TestMain, if defined (it's optional), allows setup code to be run before and after the suite of unit tests
// for this package.
func TestMain(m *testing.M) {

	// Ensure that our Firestore and Pub/Sub requests do not get routed to the live project by mistake
	fulfillapi.ProjectId = "demo-" + fulfillapi.ProjectId
	TopicProjectId = "demo-" + TopicProjectId

	// Configure the environment variable that informs the Firestore client that it should connect to the
	// emulator and how to reach it.
	_ = os.Setenv(EnvFirestoreEmulator, FirestoreEmulatorHost)

	// Do the same for the Pub/Sub emulator
	_ = os.Setenv(EnvPubSubEmulator, PubSubEmulatorHost)
	_ = os.Setenv(EnvPubSubProjectId, TopicProjectId)

	// Instantiate our task service - panic if we cannot obtain one
	var err error
	fulfillmentService, err = fulfillapi.NewFulfillmentService()
	if err != nil {
		zap.L().Panic("unable to instantiate fulfillment service / firestore client", zap.Error(err))
	}

	// Create our Pub/Sub topic if it does not already exist
	err = createPubSubTopic()
	if err != nil {
		zap.L().Panic("unable to create pubsub topic", zap.Error(err))
	}

	// Set our task submission and slightly later Firestore document creation time values
	tempTime, _ := types.TimestampFromRFC3339Nano(timeString)
	taskSubmissionTime = tempTime.GetTime()

	// Build and store a task that we can use as a target in our tests
	storedTaskId = storeMockTask()

	// Run all the unit tests
	m.Run()
}

// createPubSubTopic ensures that the Pub/Sub topic that the trigger handle publishes to exists within
// the emulator. Calling this function more than once will do no harm.
func createPubSubTopic() error {

	// Obtain a Pub/Sub client
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, TopicProjectId)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}

	// Ensure that the client gets closed regardless
	defer func(client *pubsub.Client) {
		_ = client.Close()
	}(client)

	// Try to access the topic to see if it already exists
	topic := client.Topic(TopicId)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return fmt.Errorf("topic.Exists: %v", err)
	}

	// If the topic does not already exist, create it now
	if !exists {
		_, err = client.CreateTopic(ctx, TopicId)
	}
	return err
}

// TestHandlerHappyPath evaluates normal operation of the Firestore trigger handler function when all goes well.
func TestHandlerHappyPath(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Submit a known FirestoreEvent to the handler while capturing its log output
	event := mockFirestoreEvent(storedTaskId)
	ctx := context.Background()
	var err error
	logged := testutil.CaptureLogging(func() {
		err = CartTrigger(ctx, *event)
	})

	// There should have been no errors and some straightforward log output
	req.Nil(err, "no error was expected: %v", err)
	req.Contains(logged, "published task", "did not see happy path log message")
	req.Contains(logged, storedTaskId, "did not see task ID in log message")

	// Repeat a second time (would never happen for the same task in real life) in task
	// to exercise the already loaded paths of the task service and pubsub client lazy loaders.
	logged = testutil.CaptureLogging(func() {
		err = CartTrigger(ctx, *event)
	})
	req.Nil(err, "no error was expected on second run: %v", err)
	req.Contains(logged, "published task", "did not see happy path log message on second run")
	req.Contains(logged, storedTaskId, "did not see task ID in log message on second run")
}

// TestTaskNotExist looks at what happens when a task update triggers the handler but the task in
// question does not exists - can't see how that could happen but it has the side benefit of testing
// onm of the error paths trying to load a task from Firestore without having to mock an error.
func TestTaskNotExist(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Configure a Firestore event where the task ID won't be found when the trigger function
	// tries to load the full task.
	taskId := uuid.NewString()
	event := mockFirestoreEvent(taskId)
	ctx := context.Background()
	var err error
	logged := testutil.CaptureLogging(func() {
		err = CartTrigger(ctx, *event)
	})

	// There should have been no errors and some straightforward log output
	req.NotNil(err, "an error was expected")
	req.Contains(logged, "unable to retrieve task from firestore", "did not see task retrieval failure log message")
	req.Contains(logged, taskId, "did not see task ID in log message")
}

// TestPublishError forces publishing to fail by screwing with the topic ID, setting it to a topic that does not
// exist and forcing a fresh lazy load.
func TestPublishError(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Reset the publishing client after we are done so that other tests won't be tripped up
	originalTopicId := TopicId
	defer func() {
		TopicId = originalTopicId
		pubSubClient.(*PubSubClientImpl).topic = nil
	}()

	// Force the pubsub client to lazy load a second time but with the ID of a topic that does not exist
	pubSubClient.(*PubSubClientImpl).topic = nil
	TopicId = "no-way-this-topic-id-matches-anything"

	// Submit a checked out task FirestoreEvent to the handler while capturing its log output
	event := mockFirestoreEvent(storedTaskId)
	ctx := context.Background()
	var err error
	logged := testutil.CaptureLogging(func() {
		err = CartTrigger(ctx, *event)
	})

	// There should have been no errors and some straightforward log output
	req.NotNil(err, "an error was expected")
	req.Contains(logged, "pubsub publish failed", "did not see publish failure log message")
	req.Contains(logged, storedTaskId, "did not see task ID in log message")
}

// mockFirestoreEvent constructs a FirestoreEvent populated with known values that we can check in out unit tests.
func mockFirestoreEvent(taskId string) *FirestoreEvent {
	return &FirestoreEvent{
		Value: *mockNewValue(taskId),
	}
}

// mockNewValue returns a FirestoreValue populated with a checked out task structure as its value and
// with the update time being later than its creation time.
func mockNewValue(taskId string) *FirestoreValue {

	// Get a task in Firestore trigger image form (minimally populated)
	task := buildMockTriggerTask(taskId)

	// Build and return our result - update time will be after the creation time
	return &FirestoreValue{
		CreateTime: firestoreValueCreateTime,
		Fields:     *task,
		Name:       task.StoreRefPath(),
		UpdateTime: firestoreValueCreateTime,
	}
}

// StoreRefPath returns the string representation of the document reference path for this Task.
func (c *FirestoreTask) StoreRefPath() string {
	return schema.TaskCollection + c.Id.StringValue
}

// buildMockTriggerTask returns a minimally populated FirestoreTask structure.
func buildMockTriggerTask(taskId string) *FirestoreTask {
	return &FirestoreTask{
		Id: StringValue{
			taskId,
		},
	}
}

// storeMockTask stores a task in the Firestore emulator so that it can be retrieved when
// we invoke the trigger (i.e. the heart of this Cloud Function) in our tests.
//
// This will panic if the task cannot be saved - our unit tests cannot run without it so why let them
// run at all if we can't save this corner stone.
func storeMockTask() string {

	// Clan out any debris left by preior test runs
	deleteAllMockTasks()

	// Build a mock task that we can write to Firestore
	task := buildMockTask()

	// Write that to the Firestore database
	ctx := context.Background()
	err := fulfillmentService.SaveTasks(ctx, []*schema.Task{task})
	if err != nil {
		zap.L().Panic("failed saving mock task to firestore", zap.Error(err))
	}

	// And we are done!
	return task.Id
}

// deleteAllMockTasks removes our mock tasks from the Firestore emulator. We use this to ensure that our tests are
// run against a clean slate.
func deleteAllMockTasks() {

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
		ctx := context.Background()
		response, err := fulfillmentService.GetTasks(ctx, request)
		if err != nil {
			zap.L().Panic("did not expect an error calling GetTasks", zap.Error(err))
		}
		if len(response.Tasks) == 0 {
			return
		}

		// We have some tasks to remove
		for _, pbTask := range response.Tasks {

			// Convert the one field we care about to obtain the reference path to our internal Order form
			task := schema.Task{Id: pbTask.Id}

			// Establish a document reference for that task path and ask for the document to be deleted
			ref := fulfillmentService.FsClient.Doc(task.StoreRefPath())
			_, err = ref.Delete(ctx)
			if err != nil {
				zap.L().Panic("failed deleting task ID", zap.Error(err))
			}

			// Let the impatient engineer watching our progress know who things are going
			zap.L().Info("deleted debris task", zap.String("taskId", task.Id))
		}

		// Adjust the request to include tye next page token
		request.PageToken = response.NextPageToken
	}
}

// buildMockTask returns a Task structure populated with known data
func buildMockTask() *schema.Task {
	return &schema.Task{
		Id:             uuid.NewString(),
		SubmissionTime: taskSubmissionTime,
		OrderId:        OrderID,
		OrderItemId:    ItemId,
		ProductCode:    ProductCode,
		Status:         schema.WAITING_CUSTOMER,
		ReasonCode:     ReasonCode,
		Parameters: []*schema.Parameter{
			{Name: paramName1, Value: paramValue1},
			{Name: paramName2, Value: paramValue2},
		},
	}
}
