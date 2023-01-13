// Package taskdistrib implements a Google Cloud Function to receive fulfillment task descriptions via
// a Pub/Sub topic. The handler maps the task description to the identity of another Cloud Function
// that has the know how to execute the task, then invokes that function.
package taskdistrib

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/mikebway/poc-gcp-ecomm/fulfillment/schema"
	pb "github.com/mikebway/poc-gcp-ecomm/pb/fulfillment"
	"go.uber.org/zap"
	"google.golang.org/api/pubsub/v1"
)

const (
	// thisFunctionName is the name of this function as it appears in HTTPS invocation URLs, i.e.,
	// the "task-distributor" portion of "https://task-distributor-azbqye5oka-uc.a.run.app"
	// where the "azbqye5oka" is unique to the GCP project hosting the function and its siblings.
	thisFunctionName = "task-distributor"

	// urlProtocolPrefix is the HTTPS prefix we expect to see at the front of all cloud function URLs
	urlProtocolPrefix = "https://"
)

var (
	// taskMap relates task description data to the identity of a cloud function that can execute,
	// or at least initiate, fulfillment of that task.
	//
	// The map key is a string formed from a concatenation of the task's productCode (which product is
	// to be fulfilled), the taskCode (fulfillment of a product may entail the completion of multiple tasks),
	// and the task status (e.g., completion of a task may trigger the start of dependent tasks).
	//
	// To a limited degree, the mapping allows for productCode and task_code to be wildcarded. If the
	// full product:task:status key value is not matched, then the distributor will check for just a
	// task:status match. If that fails, then the distributor will try for a match on just the status.
	//
	// This means, for example, that a single map entry for COMPLETED status can trigger the invocation of a
	// common Cloud Function to wake up dependent tasks that have been waiting on completion of other tasks.
	taskMap map[string]*string

	// functionUrlSuffix is the root portion of the Google Cloud Functions that reside in the same GCP project
	// as this function. It is appended as a suffix to the name of a cloud function to form the URL of that
	// function along with its HTTPS protocol prefix.
	functionUrlSuffix string
)

// pushRequest represents the payload of a Pub/Sub push message.
type pushRequest struct {
	Message      pubsub.PubsubMessage `json:"message"`
	Subscription string               `json:"subscription,omitempty"`
}

// init is the static initializer used to configure our local and global static variables.
func init() {
	// Initialize our Zap logger
	serviceLogger, _ := zap.NewProduction()
	zap.ReplaceGlobals(serviceLogger)

	// Cheat and hard code the task to Cloud Function ID map here
	// TODO: load this map from a configuration stored in Firestore
	taskMap = make(map[string]*string)
	addTaskMapping("gold_yoyo", "manufacture", schema.WAITING_SERVICE, "task-gy-man")
	addTaskMapping("plastic_yoyo", "upsell_to_gold", schema.WAITING_CS, "task-py-up")
	addTaskMapping("", "sf_case", schema.WAITING_CS, "task-sf")
	addTaskMapping("", "ship", schema.WAITING_SERVICE, "task-ship")
}

// TaskDistributor is the Cloud Function entry point. The payload of the Pub/Sub push request is a task
// description expressed as a base64 encoded Protocol Buffer message wrapped in a JSON envelop.
//
// See https://cloud.google.com/pubsub/docs/push for documentation of the request body JSON content.
func TaskDistributor(w http.ResponseWriter, r *http.Request) {

	// Use our own function URL to determine the common root domain name of all
	// Cloud functions running in this GCP project
	determineUrlRoot(r)

	// Have our big brother sibling do all the real work while we just handle the HTTP interfacing here
	status, err := doTaskDistributor(r.Context(), r.Body)
	if err != nil {

		// Dang - log the error and return it to the caller as well
		zap.L().Error("failed to execute task function", zap.Error(err))
		http.Error(w, err.Error(), status)
	}

	// Return the successful status code
	w.WriteHeader(status)
}

// doTaskDistributor does all the heavy lifting for TaskDistributor. It is implemented as a separate
// function to isolate the message processing from the transport interface.
//
// An HTTP status code is always returned, this should be set in the response regardless of whether
// an error is also returned.
//
// See https://cloud.google.com/pubsub/docs/push for documentation of the reader JSON content.
func doTaskDistributor(_ context.Context, reader io.Reader) (int, error) {

	// Unpack the JSON push request message from the request body
	var pushReq pushRequest
	if err := json.NewDecoder(reader).Decode(&pushReq); err != nil {
		return http.StatusBadRequest, fmt.Errorf("could not decode push request json body: %v", err)
	}

	// Translate the base64 encoded body of the request as a binary byte slice
	pbBytes, err := base64.StdEncoding.DecodeString(pushReq.Message.Data)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("unable to decode base64 data: %w", err)
	}

	// Unmarshall the protobuf binary message into a task structure
	task, err := unmarshalTask(pbBytes)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Figure out which cloud function matches this task
	zap.L().Info("matching task to handler", zap.String("taskId", task.Id),
		zap.String("product", task.ProductCode),
		zap.String("task", task.TaskCode), zap.String("status", pb.TaskStatus_name[int32(task.Status)]))
	taskHandlerUrl := functionURL(task)
	if taskHandlerUrl != nil {
		zap.L().Info("invoking task handler", zap.String("url", *taskHandlerUrl))
	} else {
		zap.L().Info("no task handler match")
	}

	// Everything is good
	zap.L().Info("handled", zap.String("taskId", task.Id))
	return http.StatusOK, nil
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

// determineUrlRoot will extract the common domain name suffix for all Cloud Functions in the
// same GCP project as this one and set that aside as the basis for forming the URLs of the
// functions we are going to call.
func determineUrlRoot(r *http.Request) {

	// From experimentation, it seems that r.Host does (as you might expect) contain the DNS name for
	// this Cloud Function. We can strip out function name from that to leave us with the common root
	// for all Cloud Function DNS names within the same GCP project.
	functionUrlSuffix = strings.TrimPrefix(r.Host, thisFunctionName)
}

// addTaskMapping adds a single task to Cloud Function mapping to the taskMap.
func addTaskMapping(productCode, taskCode string, status pb.TaskStatus, taskFuncName string) {
	taskMap[taskKey(productCode, taskCode, status)] = &taskFuncName
}

// functionURL looks uop a task in the taskMap and return a URL for a corresponding function if there is one,
// otherwise nil.
func functionURL(task *pb.Task) *string {

	// Form a key from all the relevant task values and lookup a function URL for that key
	funcName := taskMap[fullTaskKey(task)]
	if funcName == nil {

		// We found nothing, try matching on just the task type and status
		funcName = taskMap[allProductsTaskKey(task)]
		if funcName == nil {

			// We still found nothing, try matching on just the task status
			funcName = taskMap[statusOnlyTaskKey(task)]
			if funcName == nil {

				// There is nothing to find, give up
				return nil
			}
		}
	}

	// Combine the function name with the protocol prefix and shared domain root to
	// form the full URL of the function, then return that
	funcUrl := urlProtocolPrefix + *funcName + functionUrlSuffix
	return &funcUrl
}

// fullTaskKey forms a taskMap key from all the possibly relevant values of the given task.
func fullTaskKey(task *pb.Task) string {
	return taskKey(task.ProductCode, task.TaskCode, task.Status)
}

// allProductsTaskKey forms a taskMap key from just the taskCode and status of the given task, i.e.
// matching regardless of product type if tasks share a task type and status.
func allProductsTaskKey(task *pb.Task) string {
	return taskKey("", task.TaskCode, task.Status)
}

// statusOnlyTaskKey forms a taskMap key from just the status of the given task, i.e.
// matching regardless of product type and task type.
func statusOnlyTaskKey(task *pb.Task) string {
	return taskKey("", "", task.Status)
}

// taskKey assembles the given values into a taskMao key.
func taskKey(productCode, taskCode string, status pb.TaskStatus) string {
	return productCode + ":" + taskCode + ":" + strconv.Itoa(int(status))
}
