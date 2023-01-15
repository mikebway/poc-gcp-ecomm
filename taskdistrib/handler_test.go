package taskdistrib

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/mikebway/poc-gcp-ecomm/fulfillment/schema"
	"github.com/mikebway/poc-gcp-ecomm/testutil"
	"github.com/mikebway/poc-gcp-ecomm/types"
	"github.com/stretchr/testify/require"
	pubsubapi "google.golang.org/api/pubsub/v1"
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

	// funcDomainSuffix defines the tail end of the mock domain name that is common to all Cloud Function URLs
	// served within the same project as the task distributor function we are testing
	funcDomainSuffix = "-unittest--uc.a.run.app"

	// distribFuncHost is specified as the TaskDistributor cloud function domain name, i.e. the domain od the URL
	// of the function we are unit testing, in mock HTTP request invocations we submit in our tests
	distribFuncHost = thisFunctionName + funcDomainSuffix

	// goldYoyoManufactureDomain is the domain name that we expect to see for the fulfillment task function
	// invoked in response to our mock task when the product, task, and status match
	goldYoyoManufactureDomain = "task-gy-man" + funcDomainSuffix

	// salesforceCaseDomain is the domain name that we expect to see for the fulfillment task function
	// invoked in response to our mock task when only the task and status match
	salesforceCaseDomain = "task-sf" + funcDomainSuffix

	// shipDomain is the domain name that we expect to see for the fulfillment task function
	// invoked in response to our mock task when only status matches
	shipDomain = "task-ship" + funcDomainSuffix
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

// TestTaskDistributorHappyPath exercises the main handler function with good data that should be processed
// without error.
func TestTaskDistributorHappyPath(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := buildHttpRequest(buildMockTask())
	responseRecorder := httptest.NewRecorder()

	// Obtain a mock fulfillment operation server that will return a successful response and set
	// its URL as the unit test override for functionURL()
	svr := newCloudEventServer(nil)
	unitTestOverrideUrl = svr.URL

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		TaskDistributor(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(http.StatusOK, responseRecorder.Code, "should have a 200 OK response code")
	req.Contains(logged, "matching task to handler", "should have seen the expected information message in the logs")
	req.Contains(logged, "\"taskId\": \""+taskId+"\"", "should have seen the expected task ID in the logs")
	req.Contains(logged, "matched task to handler", "should have seen the expected url match message in the logs")
	req.Contains(logged, "\"url\": \""+urlProtocolPrefix+goldYoyoManufactureDomain+"\"", "should have seen the fulfillment task URL in the logs")
	req.Contains(logged, "handled", "should have seen the expected completion message in the logs")
}

// TestAllProductsMatch confirms that a task and status key will be properly formed and matched if a full match of
// product, task, and status is not matched first.
//
// The full key match case is covered by TestTaskDistributorHappyPath.
func TestAllProductsMatch(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Obtain a mock task then change its product and task to something we know will match with an
	// all products case in the TaskDistributor task map
	task := buildMockTask()
	task.ProductCode = "this-will-not-match"
	task.TaskCode = "sf_case"
	task.Status = schema.WAITING_CS

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := buildHttpRequest(task)
	responseRecorder := httptest.NewRecorder()

	// Obtain a mock fulfillment operation server that will return a successful response and set
	// its URL as the unit test override for functionURL()
	svr := newCloudEventServer(nil)
	unitTestOverrideUrl = svr.URL

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		TaskDistributor(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(http.StatusOK, responseRecorder.Code, "should have a 200 OK response code")
	req.Contains(logged, "matching task to handler", "should have seen the expected information message in the logs")
	req.Contains(logged, "\"taskId\": \""+taskId+"\"", "should have seen the expected task ID in the logs")
	req.Contains(logged, "matched task to handler", "should have seen the expected url match message in the logs")
	req.Contains(logged, "\"url\": \""+urlProtocolPrefix+salesforceCaseDomain+"\"", "should have seen the fulfillment task URL in the logs")
	req.Contains(logged, "handled", "should have seen the expected completion message in the logs")
}

// TestStatusOnlyMatch confirms that a status only key will be properly formed and matched if neither a full match of
// product, task, and status nor of task and status is not matched first.
//
// The full key match case is covered by TestTaskDistributorHappyPath.
func TestStatusOnlyMatch(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Obtain a mock task then change its product and task to something we know will match with an
	// a status only case in the TaskDistributor task map
	task := buildMockTask()
	task.ProductCode = "this-will-not-match"
	task.TaskCode = "ship"
	task.Status = schema.WAITING_SERVICE

	// Obtain a mock fulfillment operation server that will return a successful response and set
	// its URL as the unit test override for functionURL()
	svr := newCloudEventServer(nil)
	unitTestOverrideUrl = svr.URL

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := buildHttpRequest(task)
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		TaskDistributor(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(http.StatusOK, responseRecorder.Code, "should have a 200 OK response code")
	req.Contains(logged, "matching task to handler", "should have seen the expected information message in the logs")
	req.Contains(logged, "\"taskId\": \""+taskId+"\"", "should have seen the expected task ID in the logs")
	req.Contains(logged, "matched task to handler", "should have seen the expected url match message in the logs")
	req.Contains(logged, "\"url\": \""+urlProtocolPrefix+shipDomain+"\"", "should have seen the fulfillment task URL in the logs")
	req.Contains(logged, "handled", "should have seen the expected completion message in the logs")
}

// TestNoMatch confirms that a failure to match a task to a fulfillment function by any criteria will result
// in a NO-OP success.
func TestNoMatch(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Obtain a mock task then change its product and task to something we know there will be no fulfillment
	// cloud function for
	task := buildMockTask()
	task.Status = schema.WAITING_CUSTOMER

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := buildHttpRequest(task)
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		TaskDistributor(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(http.StatusOK, responseRecorder.Code, "should have a 200 OK response code")
	req.Contains(logged, "matching task to handler", "should have seen the expected information message in the logs")
	req.Contains(logged, "\"taskId\": \""+taskId+"\"", "should have seen the expected task ID in the logs")
	req.Contains(logged, "no task handler match", "should have seen the expected nothing to do message in the logs")
	req.Contains(logged, "handled", "should have seen the expected completion message in the logs")
}

// TestCorruptTask exercises the main handler function with duff task data that should fail to be
// unmarshalled.
func TestCorruptTask(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Assemble a mock HTTP request with bad data and a means to record the response
	httpRequest := httptest.NewRequest("POST", "/", buildPushRequestBody([]byte("this is not a valid task")))
	httpRequest.Host = distribFuncHost
	responseRecorder := httptest.NewRecorder()

	// Obtain a mock fulfillment operation server that will return a successful response and set
	// its URL as the unit test override for functionURL()
	svr := newCloudEventServer(nil)
	unitTestOverrideUrl = svr.URL

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		TaskDistributor(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(http.StatusBadRequest, responseRecorder.Code, "should have a 400 Bad Request response code")
	req.Contains(logged, "failed to execute task function", "should have seen the expected failure message in the logs")
	req.Contains(logged, "failed to unmarshal task protobuf message", "should have seen the expected protobuf error message in the logs")
}

// TestCorruptPush exercises the main handler function with a duff Pub/Sub push JSON envelope
func TestCorruptPush(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Assemble a mock HTTP request with bad data and a means to record the response
	httpRequest := httptest.NewRequest("POST", "/", bytes.NewReader([]byte("this is not JSON")))
	httpRequest.Host = distribFuncHost
	responseRecorder := httptest.NewRecorder()

	// Obtain a mock fulfillment operation server that will return a successful response and set
	// its URL as the unit test override for functionURL()
	svr := newCloudEventServer(nil)
	unitTestOverrideUrl = svr.URL

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		TaskDistributor(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(http.StatusBadRequest, responseRecorder.Code, "should have a 400 Bad Request response code")
	req.Contains(logged, "failed to execute task function", "should have seen the expected failure message in the logs")
	req.Contains(logged, "could not decode push request json body", "should have seen the expected JSON error message in the logs")
}

// TestInvalidBase64 data exercises the main handler function with invalid base64 encoding of a task.
func TestInvalidBase64(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Assemble a mock HTTP request with bad data and a means to record the response
	pushReq := &pushRequest{
		Message: pubsubapi.PubsubMessage{
			Data: "-- not base64 data --",
		},
	}
	jsonBytes, _ := json.Marshal(pushReq)
	httpRequest := httptest.NewRequest("POST", "/", bytes.NewReader(jsonBytes))
	httpRequest.Host = distribFuncHost
	responseRecorder := httptest.NewRecorder()

	// Obtain a mock fulfillment operation server that will return a successful response and set
	// its URL as the unit test override for functionURL()
	svr := newCloudEventServer(nil)
	unitTestOverrideUrl = svr.URL

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		TaskDistributor(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(http.StatusBadRequest, responseRecorder.Code, "should have a 400 Bad Request response code")
	req.Contains(logged, "failed to execute task function", "should have seen the expected failure message in the logs")
	req.Contains(logged, "unable to decode base64 data", "should have seen the expected base64 error message in the logs")
}

// TestTaskFunctionFailure exercises the main handler function with good data that should be processed
// without error but where the invoked Cloud Function fails.
func TestTaskFunctionFailure(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := buildHttpRequest(buildMockTask())
	responseRecorder := httptest.NewRecorder()

	// Obtain a mock fulfillment operation server that will return a failure response and set
	// its URL as the unit test override for functionURL()
	const functionErrMsg = "sorry, we are out of tea"
	svr := newCloudEventServer(errors.New(functionErrMsg))
	unitTestOverrideUrl = svr.URL

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		TaskDistributor(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(http.StatusInternalServerError, responseRecorder.Code, "should have a 200 OK response code")
	req.Contains(logged, "matching task to handler", "should have seen the expected information message in the logs")
	req.Contains(logged, "\"taskId\": \""+taskId+"\"", "should have seen the expected task ID in the logs")
	req.Contains(logged, "matched task to handler", "should have seen the expected url match message in the logs")
	req.Contains(logged, "\"url\": \""+urlProtocolPrefix+goldYoyoManufactureDomain+"\"", "should have seen the fulfillment task URL in the logs")
	req.Contains(logged, "failed to execute task function", "should have seen the expected execution failure message in the logs")
	req.Contains(logged, "fulfillment function failed", "should have seen the expected failed function message in the logs")
	req.Contains(logged, functionErrMsg, "should have seen the expected function error response message in the logs")
	req.NotContains(logged, "handled", "should not have seen the successful completion message in the logs")
}

// TestTaskUndeliveredFailure exercises the main handler function with good data that should be processed
// without error but where the invoked Cloud Function cannot be invoked / does not exist.
func TestTaskUndeliveredFailure(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := buildHttpRequest(buildMockTask())
	responseRecorder := httptest.NewRecorder()

	// Have the task distributor try to invoke a Cloud Function that does not respond at all
	unitTestOverrideUrl = "https://this-is-not-a-valid-domain.nosuch.domain"

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		TaskDistributor(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(http.StatusInternalServerError, responseRecorder.Code, "should have a 200 OK response code")
	req.Contains(logged, "matching task to handler", "should have seen the expected information message in the logs")
	req.Contains(logged, "\"taskId\": \""+taskId+"\"", "should have seen the expected task ID in the logs")
	req.Contains(logged, "matched task to handler", "should have seen the expected url match message in the logs")
	req.Contains(logged, "\"url\": \""+urlProtocolPrefix+goldYoyoManufactureDomain+"\"", "should have seen the fulfillment task URL in the logs")
	req.Contains(logged, "failed to execute task function", "should have seen the expected execution failure message in the logs")
	req.Contains(logged, "fulfillment function failed", "should have seen the expected failed function message in the logs")
	req.Contains(logged, "no such host", "should have seen the expected undeliverable 'no such host' message in the logs")
	req.NotContains(logged, "handled", "should not have seen the successful completion message in the logs")
}

// newCloudEventServer returns an httptest.Server configured to respond to requests with the given error or nil
// where nil would return a 200 OK response.
func newCloudEventServer(err error) *httptest.Server {

	// Return a mock HTTP service that will do what we ask
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err != nil {
			http.Error(w, err.Error(), http.StatusTeapot)
		}
	}))
}

// buildHttpRequest assembles a mock http.Request POSTing the given task to the TaskDistributor function.
func buildHttpRequest(task *schema.Task) *http.Request {

	// Render the task to its protocol buffer structure form
	pbTask := task.AsPBTask()

	// Marshal that to its binary message state and return that
	pbBytes, _ := proto.Marshal(pbTask)

	httpRequest := httptest.NewRequest("POST", "/", buildPushRequestBody(pbBytes))
	httpRequest.Host = distribFuncHost

	// And we are done
	return httpRequest
}

// buildPushRequest wraps the provided data bytes as the data payload of a push request, returning that as byte reader.
func buildPushRequestBody(data []byte) io.Reader {

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

// buildMockTask returns a populated schema.Task structure that can be used to test
// the distribution of fulfillment tasks to subordinate Cloud Functions in our tests.
func buildMockTask() *schema.Task {
	return &schema.Task{
		Id:             taskId,
		SubmissionTime: taskSubmissionTime,
		OrderId:        orderId,
		OrderItemId:    orderItemId,
		ProductCode:    productCode,
		TaskCode:       taskCode,
		Status:         schema.WAITING_SERVICE,
	}
}
