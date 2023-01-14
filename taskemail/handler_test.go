package taskemail

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2/event"
	"github.com/golang/protobuf/proto"
	"github.com/mikebway/poc-gcp-ecomm/fulfillment/schema"
	pb "github.com/mikebway/poc-gcp-ecomm/pb/fulfillment"
	"github.com/mikebway/poc-gcp-ecomm/testutil"
	"github.com/mikebway/poc-gcp-ecomm/types"
	"github.com/stretchr/testify/require"
)

const (
	// timeString is a timestamp string we can use to derive known time values
	timeString = "2022-10-29T16:23:19.123456789-06:00"

	// taskId is a UUID string value that we use as task ID in our tests
	taskId = "aeadfb36-eb58-464e-b12b-d2d23b731a49"

	// orderId is a UUID string value that we use as order ID in our tests
	orderId = "d1cecab3-5bc0-43d4-aef1-99ad69794313"

	// orderItemId is a UUID string value that we use as itme ID in our tests
	orderItemId = "d1cecab3-5bc0-43d4-aef1-99ad69794313"

	// productCode identifies the product our mock fulfilment task is for
	productCode = "gold_yoyo"

	// taskCode defines the specific responsibility of our mock fulfilment task
	taskCode = "manufacture"

	// eventId defines the ID of our mock event
	eventId = "a4750be2-8f42-486f-8e77-ad7713a9abc1"

	// eventSource defines the source of our mock event as being from a unit test
	eventSource = "https://unitest.com/taskemail"

	// eventType defines the source of our mock event as being from a unit test
	eventType = "test"

	// fulfillmentOperation defines the fulfillment operation that we will ask the target function
	// to pretend to be performing - it is set via n environment variable
	fulfillmentOperation = "UNIT_TEST_OPERATION"
)

var (
	// Task value types that cannot be declared as constants
	taskSubmissionTime time.Time
)

// TestMain, if defined (it's optional), allows setup code to be run before and after the suite of unit tests
// for this package.
func TestMain(m *testing.M) {

	// Mock task values that cannot be defined as constants
	t, _ := types.TimestampFromRFC3339Nano(timeString)
	taskSubmissionTime = t.GetTime()

	// Run all the unit tests
	m.Run()
}

// TestFulfillTaskHappyPath exercises the main event handler function with good data that should be processed
// without error.
func TestFulfillTaskHappyPath(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Set the environment variable for the fulfillment operation the function is to claim it is performing.
	// we will check that this is reported in the logs.
	_ = os.Setenv(envVarOperation, fulfillmentOperation)

	// Assemble a mock event
	event := buildEvent()

	// Wrap a call to the target function so that we can capture its log output
	ctx := context.Background()
	var err error
	logged := testutil.CaptureLogging(func() {
		err = fulfillTask(ctx, *event)
	})

	// Confirm the result was a happy one
	req.Nil(err, "should have been successful")
	req.Contains(logged, "fulfillTask event", "should have seen the expected function entry message in the logs")
	req.Contains(logged, "\"operation\": \""+fulfillmentOperation+"\"", "should have seen the expected fulfillment operation in the logs")
	req.Contains(logged, "\"eventId\": \""+event.ID()+"\"", "should have seen the expected event ID in the logs")
	req.Contains(logged, "\"eventType\": \""+event.Type()+"\"", "should have seen the expected event type in the logs")
	req.Contains(logged, "fulfilled", "should have seen the expected completion message in the logs")
	req.Contains(logged, "\"taskId\": \""+taskId+"\"", "should have seen the fulfillment task ID in the logs")
	req.Contains(logged, "\"product\": \""+productCode+"\"", "should have seen the fulfillment product in the logs")
	req.Contains(logged, "\"task\": \""+taskCode+"\"", "should have seen the fulfillment task code in the logs")
	req.Contains(logged, "\"status\": "+strconv.Itoa(int(schema.WAITING_SERVICE)), "should have seen the fulfillment status ID in the logs")
}

// TestFulfillTaskSadPath exercises the main event handler function with bad data that should result in an error.
func TestFulfillTaskSadPath(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Reset the environment variable for the fulfillment operation the function is to claim it is performing
	// so that the default name will be reported in the logs
	_ = os.Setenv(envVarOperation, "")

	// Assemble a mock event then mess with the data so that it won't unmarshal as a task
	event := buildEvent()
	_ = event.SetData(cloudevents.Base64, []byte("Ceci n’est pas une pipe"))

	// Wrap a call to the target function so that we can capture its log output
	ctx := context.Background()
	var err error
	logged := testutil.CaptureLogging(func() {
		err = fulfillTask(ctx, *event)
	})

	// Confirm the result was a happy one
	req.Contains(logged, "fulfillTask event", "should have seen the expected function entry message in the logs")
	req.Contains(logged, "\"operation\": \""+defaultOperation+"\"", "should have seen the default fulfillment operation in the logs")
	req.Contains(logged, "\"eventId\": \""+event.ID()+"\"", "should have seen the expected event ID in the logs")
	req.Contains(logged, "\"eventType\": \""+event.Type()+"\"", "should have seen the expected event type in the logs")
	req.NotNil(err, "should have seen an error")
	req.Contains(err.Error(), "unmarshal task failure:", "should have seen the expected error message in the logs")
	req.Contains(err.Error(), "failed to unmarshal task protobuf message:", "should have seen the expected error cause in the logs")
}

// buildEvent returns a populated CloudEvent containing our mock task description encoded as base64.
func buildEvent() *cloudevents.Event {

	// Get the task that we are going to embed in the event
	task := buildMockTask()

	// Marshal that to its binary message state
	pbBytes, _ := proto.Marshal(task)

	// Create a CloudEvents structure and populate that with our data
	e := cloudevents.New()
	e.SetID(eventId)
	e.SetSource(eventSource)
	e.SetType(eventType)
	_ = e.SetData(cloudevents.Base64, pbBytes)

	// All done
	return &e
}

// buildMockTask returns a populated schema.Task structure that can be used to test
// the invocation of our fulfillment tasks event handler.
func buildMockTask() *pb.Task {
	return (&schema.Task{
		Id:             taskId,
		SubmissionTime: taskSubmissionTime,
		OrderId:        orderId,
		OrderItemId:    orderItemId,
		ProductCode:    productCode,
		TaskCode:       taskCode,
		Status:         schema.WAITING_SERVICE,
	}).AsPBTask()
}
