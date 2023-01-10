package taskdistrib

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		TaskDistributor(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(http.StatusOK, responseRecorder.Code, "should have a 200 OK response code")
	req.Contains(logged, "matching task to handler", "should have seen the expected information message in the logs")
	req.Contains(logged, "\"taskId\": \""+taskId+"\"", "should have seen the expected task ID in the logs")
	req.Contains(logged, "invoking task handler", "should have seen the expected invocation message in the logs")
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

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		TaskDistributor(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(http.StatusOK, responseRecorder.Code, "should have a 200 OK response code")
	req.Contains(logged, "matching task to handler", "should have seen the expected information message in the logs")
	req.Contains(logged, "\"taskId\": \""+taskId+"\"", "should have seen the expected task ID in the logs")
	req.Contains(logged, "invoking task handler", "should have seen the expected invocation message in the logs")
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
	req.Contains(logged, "invoking task handler", "should have seen the expected invocation message in the logs")
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

// buildHttpRequest assembles a mock http.Request POSTing the given task to the TaskDistributor function.
func buildHttpRequest(task *schema.Task) *http.Request {

	// Render the task to its protocol buffer structure form
	pbTask := task.AsPBTask()

	// Marshal that to its binary message state and return that
	pbBytes, _ := proto.Marshal(pbTask)

	httpRequest := httptest.NewRequest("POST", "/", buildPushRequestBody(pbBytes))
	httpRequest.URL = &url.URL{Host: distribFuncHost}
	httpRequest.RequestURI = urlProtocolPrefix + distribFuncHost
	httpRequest.Header.Add("X-Forwarded-Host", distribFuncHost)

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
