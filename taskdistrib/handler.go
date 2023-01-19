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
	nethttp "net/http"
	"strconv"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/mikebway/poc-gcp-ecomm/fulfillment/schema"
	pb "github.com/mikebway/poc-gcp-ecomm/pb/fulfillment"
	"go.uber.org/zap"
	"google.golang.org/api/idtoken"
	"google.golang.org/api/pubsub/v1"
)

const (
	// thisFunctionName is the name of this function as it appears in HTTPS invocation URLs, i.e.,
	// the "task-distributor" portion of "https://task-distributor-azbqye5oka-uc.a.run.app"
	// where the "azbqye5oka" is unique to the GCP project hosting the function and its siblings.
	thisFunctionName = "task-distributor"

	// cloudEventType is the type that we set in the CloudEvent envelope that we send to fulfillment
	// operation Cloud Functions.
	cloudEventType = "fulfillment-op"

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
	taskMap map[string]string

	// unitTestOverrideUrl will be set to the URL of a mock HTTP server if we are running unit tests.
	// This URL should be returned by functionURL() if it is not nil and the task being evaluated does
	// match an entry in the taskMap.
	unitTestOverrideUrl string
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
	taskMap = make(map[string]string)
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

	// Flush the logs before exiting each invocation of this Cloud Function
	//goland:noinspection GoUnhandledErrorResult
	defer zap.L().Sync()

	// Use our own function URL to determine the common root domain name of all
	// Cloud functions running in this GCP project
	urlRoot := determineUrlRoot(r)

	// Have our big brother sibling do all the real work while we just handle the HTTP interfacing here
	status, err := doTaskDistributor(r.Context(), r.Body, urlRoot)
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
func doTaskDistributor(ctx context.Context, reader io.Reader, urlRoot string) (int, error) {

	// Obtain our logger once for multiple uses in this function
	logger := zap.L()

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
	logger.Info("matching task to handler", zap.String("taskId", task.Id),
		zap.String("product", task.ProductCode),
		zap.String("task", task.TaskCode), zap.String("status", pb.TaskStatus_name[int32(task.Status)]))
	fulfillOpHandlerUrl := functionURL(urlRoot, task)
	if len(fulfillOpHandlerUrl) != 0 {

		// We have a fulfillment operation match - pass the task to the designated Cloud Function as a CloudEvent
		err = dispatchCloudEvent(ctx, fulfillOpHandlerUrl, pbBytes)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("fulfillment function failed: %w", err)
		}

	} else {
		logger.Info("no task handler match")
	}

	// Everything is good
	logger.Info("handled", zap.String("taskId", task.Id))
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
// same GCP project as this one and return it to serve as the basis for forming the URLs of the
// functions we are going to call.
func determineUrlRoot(r *http.Request) string {

	// From experimentation, it seems that r.Host does (as you might expect) contains the DNS name for
	// this Cloud Function. We can strip out function name from that to leave us with the common root
	// for all Cloud Function DNS names within the same GCP project.
	return strings.TrimPrefix(r.Host, thisFunctionName)
}

// dispatchCloudEvent sends the given data bytes to the target URL as a CloudEvent.
//
// See https://github.com/cloudevents/sdk-go and https://cloudevents.io/
func dispatchCloudEvent(ctx context.Context, targetUrl string, data []byte) error {

	// Create a CloudEvents structure and populate that with our data
	//
	// NOTE: the CloudEvents client will set a unique ID and timestamp for us
	eventId := uuid.NewString()
	event := cloudevents.NewEvent()
	event.SetSource(thisFunctionName)
	event.SetType(cloudEventType)
	_ = event.SetData(cloudevents.Base64, data)

	// Establish the target URL context
	ctx = cloudevents.ContextWithTarget(ctx, targetUrl)

	// Obtain a CloudEvents client that is authorized to access the target URL
	ceClient, err := getAuthorizedCEClient(ctx, targetUrl)
	if err != nil {
		return fmt.Errorf("failed to obtain authorized CloudEvents client: %w", err)
	}

	// Send the event
	zap.L().Info("dispatching fulfillment operation event", zap.String("id", eventId))
	result := ceClient.Send(ctx, event)

	// Most of the time the result is an HTTP result, so we cast it to that and look to see if
	// the status code indicates any problems on the Clod Function side.
	httpResult, ok := result.(*cehttp.Result)
	if ok {
		// If the status code is anything that indicates the Cloud Function service call failed
		// return the result as an error
		if httpResult.StatusCode >= 300 {
			return error(result)
		}
	} else {
		// Handle the corner case where cloudevents.Client.Send did not return an HTTP result.
		// This happens if DNS lookup fails or the connection is refused etc., i.e. when the
		// error occurs at a lower layer of the TCP stack before HTTPS actually gets involved.
		return error(result)
	}

	// All is well
	return nil
}

// getAuthorizedCEClient returns a CloudEvents client configured with oauth.TokenSource authentication / authorization.
//
// See https://cloud.google.com/run/docs/authenticating/service-to-service#run-service-to-service-example-go
func getAuthorizedCEClient(ctx context.Context, audience string) (cloudevents.Client, error) {

	// TODO: Make this more efficient by caching and reusing the CloudEvent clients generated.
	// The problem is that the tokens associated with the clients may expire if the clients are not used
	// frequently enough. Without direct access to the token structure embedded in the httpClient
	// we can have no idea whether this might be the case.
	//
	// We could watch for 403 errors and recreate the client if one is seen. Or, if we knew the TTL of the
	// access token we could cache a tuple of the client and the timeout and create a new client if the
	// timeout was outside a safe margin of the current time.
	//
	// We could estimate the expiration time but obtaining another token immediately before creating the
	// httpClient, as follows:
	//
	// source, err := idtoken.NewTokenSource(ctx, audience)
	// token, err := source.Token()
	// expires := token.Expiry.Add(-time.Second * 30) // Subtract 30 second safety margin

	// Establish a regular http.Client that automatically adds an "Authorization" header
	// to any requests made.
	httpClient, err := idtoken.NewClient(ctx, audience)
	if err != nil {
		return nil, fmt.Errorf("idtoken.NewClient failed: %v", err)
	}

	// Establish our CloudEvents submission client, wrapping our authorized http.Client
	cloudEventsClient, err := cloudevents.NewClientHTTP(WithHttpClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("cloudevents.NewClientHTTP failed: %v", err)
	}
	zap.L().Info("established CloudEvents client")

	// We are all good and happy
	return cloudEventsClient, nil
}

// WithHttpClient defines an http.Option that sets the given http.Client into a CloudEvents HTTP protocol handler
// allowing the caller to provide a client with OAuth token authorization configure.
func WithHttpClient(httpClient *nethttp.Client) cehttp.Option {
	return func(p *cehttp.Protocol) error {
		p.Client = httpClient
		return nil
	}
}

// functionURL looks up a task in the taskMap and return a URL for a corresponding function if there is one,
// otherwise nil.
func functionURL(urlRoot string, task *pb.Task) string {

	// Form a key from all the relevant task values and lookup a function URL for that key
	funcName := taskMap[fullTaskKey(task)]
	if len(funcName) == 0 {

		// We found nothing, try matching on just the task type and status
		funcName = taskMap[allProductsTaskKey(task)]
		if len(funcName) == 0 {

			// We still found nothing, try matching on just the task status
			funcName = taskMap[statusOnlyTaskKey(task)]
			if len(funcName) == 0 {

				// There is nothing to find, give up
				return ""
			}
		}
	}

	// Combine the function name with the protocol prefix and shared domain root to
	// form the full URL of the function, then return that
	funcUrl := urlProtocolPrefix + funcName + urlRoot

	// Log the URL match that we came up with then see if there is a unit test override
	zap.L().Info("matched task to handler", zap.String("url", funcUrl))
	if len(unitTestOverrideUrl) != 0 {
		funcUrl = unitTestOverrideUrl
	}

	// We got what we got - return it
	return funcUrl
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

// addTaskMapping adds a single task to Cloud Function mapping to the taskMap.
func addTaskMapping(productCode, taskCode string, status pb.TaskStatus, taskFuncName string) {
	taskMap[taskKey(productCode, taskCode, status)] = taskFuncName
}

// taskKey assembles the given values into a taskMao key.
func taskKey(productCode, taskCode string, status pb.TaskStatus) string {
	return productCode + ":" + taskCode + ":" + strconv.Itoa(int(status))
}
