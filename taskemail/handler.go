package taskemail

import (
	"context"
	"fmt"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/golang/protobuf/proto"
	pb "github.com/mikebway/poc-gcp-ecomm/pb/fulfillment"
	"go.uber.org/zap"
)

const (
	// envVarOperation is the name of the environment variable that defines the name of the fulfillment
	// operation performed by this function.
	envVarOperation = "FULFILL_OPERATION"

	// defaultTaskName is used as the operation name if the FULFILL_OPERATION environment variable is not set
	defaultOperation = "UNNAMED OPERATION"
)

// init is the static initializer used to configure our local and global state.
func init() {

	// Inform the Cloud Function framework which Go function to invoke when an event is received via HTTPS POST
	functions.CloudEvent("FulfillTask", fulfillTask)
}

// fulfillTask consumes a CloudEvent message containing a e-commerce fulfillment task description
func fulfillTask(_ context.Context, e cloudevents.Event) error {

	// This code can be deployed to be multiple different mock fulfillment task actions, configured by an environment
	// variable. Real production code would be more efficient and figure this out in an init() function when
	// the function is first instantiated, but then we would not be able to unit test it so ... we do it
	// every time the function is invoked here. I would not do this for a real fulfillment function.
	operation := os.Getenv(envVarOperation)
	if len(operation) == 0 {
		operation = defaultOperation
	}

	// Log that we have been invoked and for what
	zap.L().Info("fulfillTask event", zap.String("operation", operation), zap.String("eventId", e.ID()), zap.String("eventType", e.Type()))

	// Unpack the task data from the event
	task, err := unmarshalTask(e.Data())
	if err != nil {
		return fmt.Errorf("unmarshal task failure: %v", err)
	}

	// For now - just log task information and consider that we have done our work
	zap.L().Info("fulfilled", zap.String("taskId", task.Id), zap.String("product", task.ProductCode),
		zap.String("task", task.TaskCode), zap.Int32("status", int32(task.Status)))

	// TODO: If email address environment variable is set, use SendGrid to send email to that address
	// TODO: If task defines an email address, send email to that address

	// All done, no problems
	return nil
}

// unmarshalTask unpacks the provided binary protobuf message into a task structure.
func unmarshalTask(message []byte) (*pb.Task, error) {

	// TODO: should we validate the task is properly populated etc?

	// Unmarshal the protobuf message bytes if we can
	task := &pb.Task{}
	err := proto.Unmarshal(message, task)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal task protobuf message: %w", err)
	}

	// Alright then, that was easier than I expected
	return task, nil
}
