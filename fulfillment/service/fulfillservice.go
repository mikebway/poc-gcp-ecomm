// Package service contains the gRPC Fulfillment microservice implementation.
package service

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	cartsvc "github.com/mikebway/poc-gcp-ecomm/cart/service"
	"github.com/mikebway/poc-gcp-ecomm/fulfillment/schema"
	pbfulfillment "github.com/mikebway/poc-gcp-ecomm/pb/fulfillment"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
	"strconv"
	"strings"
	"time"
)

const (
	// defaultPageSize is used if no page size is supplied in a fulfillment.GetTasksRequest, or the value is less than one.
	defaultPageSize = 20

	// maxPageSize is used if the page size is supplied in a fulfillment.GetTasksRequest is greater than 100.
	maxPageSize = 100
)

var (
	// ProjectId is a variable so that unit tests can override it to ensures that test requests are not routed to
	// the live project! See https://firebase.google.com/doos/emulator-suite/connect_firestore
	ProjectId string

	// UnitTestNewFulfillmentServiceError should be returned by NewTaskService if we are running unit tests
	// and unitTestNewTaskServiceError is not nil.
	UnitTestNewFulfillmentServiceError error
)

// init is the static initializer used to configure our local and global static variables.
func init() {

	// Set the project ID to be used for live Firestore etc. connections
	ProjectId = "poc-gcp-ecomm"
}

// FulfillmentService is a structure class with methods that implements the fulfillment.FulfillmentAPIServer gRPC API
// storing the data for managing fulfillment orchestration in a Google Cloud Firestore document collection.
type FulfillmentService struct {
	pbfulfillment.UnimplementedFulfillmentAPIServer

	// FsClient is the GCP Firestore client - it is thread safe and can be reused concurrently
	FsClient *firestore.Client

	// drProxy is used to allow unit tests to intercept firestore.DocumentRef function calls
	// and insert errors etc. into the responses.
	drProxy cartsvc.DocumentRefProxy

	// dsProxy is used to allow unit tests to intercept firestore.DocumentSnapshot function calls
	// and insert errors etc. into the responses.
	dsProxy cartsvc.DocumentSnapshotProxy

	// queryProxy is used to allow unit tests to intercept firestore.Query function calls
	// and insert errors etc. into the responses of the document iterator that the query returns.
	queryProxy cartsvc.QueryExecutionProxy
}

// NewFulfillmentService is a factory method returning an instance of our shopping cart service.
func NewFulfillmentService() (*FulfillmentService, error) {

	// Build our service instance here with our default, direct passthrough, interception proxies
	// for firestore.DocumentRef and firestore.DocumentSnapshot function calls
	svc := &FulfillmentService{
		drProxy:    &cartsvc.DocRefProxy{},
		dsProxy:    &cartsvc.DocSnapProxy{},
		queryProxy: &cartsvc.QueryExecProxy{},
	}

	// Obtain a firestore client and stuff that in the service instance
	ctx := context.Background()
	var err error
	if UnitTestNewFulfillmentServiceError == nil {
		// Set the Firestore client if we are not unit testing an error situation.
		svc.FsClient, err = firestore.NewClient(ctx, ProjectId)

	} else {
		// We are unit testing and required to report an error
		err = UnitTestNewFulfillmentServiceError
	}

	// Check that we obtained a firestore client successfully
	if err != nil {
		return nil, fmt.Errorf("could not obtain firestore client: %w", err)
	}

	// All done - return the populated service instance
	return svc, nil
}

// SaveTask stores the given schema.Task in the Firestore document collection. This is for internal domain use only and
// so does not accept or return protobuf structures.
//
// An error will be returned if the task is already present in Firestore.
func (fs *FulfillmentService) SaveTask(ctx context.Context, task *schema.Task) error {

	// Obtain a shortcut handle on our globally configured logger and log some context
	l := zap.L()
	l.Info("storing task", zap.String("taskId", task.Id), zap.String("orderId", task.OrderId),
		zap.String("itemId", task.OrderItemId), zap.String("product", task.ProductCode), zap.String("task", task.TaskCode))

	// Store the task in firestore
	ref := fs.FsClient.Doc(task.StoreRefPath())
	_, err := fs.drProxy.Create(ref, ctx, task)
	if err != nil {
		err = fmt.Errorf("failed creating task document in Firestore: %w", err)
		l.Error(err.Error(), zap.String("taskId", task.Id))
		return err
	}

	// All good, log our joy and return
	l.Info("task stored successfully", zap.String("taskId", task.Id), zap.String("path", ref.Path))
	return nil
}

// UpdateTaskStatus allows the caller to modify just the status of the task and the associate reason code, i.e.
// giving a description of why the status was changed. The reason code is optional.
func (fs *FulfillmentService) UpdateTaskStatus(ctx context.Context, req *pbfulfillment.UpdateTaskStatusRequest) (*pbfulfillment.UpdateTaskStatusResponse, error) {

	// Obtain a shortcut handle on our globally configured logger and log some context
	l := zap.L()
	l.Info("updating task status", zap.String("taskId", req.TaskId), zap.Int("status", int(req.Status)), zap.String("reason", req.ReasonCode))

	// TODO: validate the request - ID and status etc must be provided

	// Define a minimal task structure so that we can ask for its reference path
	skeleton := &schema.Task{Id: req.TaskId}

	// Build a slice containing the field level updates we are to apply
	updates := []firestore.Update{
		firestore.Update{Path: "status", Value: req.Status},
		firestore.Update{Path: "reasonCode", Value: req.ReasonCode},
	}

	// If we are marking the task as completed, add the completing time field as well
	if req.Status == schema.COMPLETED {
		updates = append(updates, firestore.Update{Path: "completionTime", Value: time.Now()})
	}

	// Update the small number of fields in the task in firestore
	ref := fs.FsClient.Doc(skeleton.StoreRefPath())
	_, err := fs.drProxy.Update(ref, ctx, updates)
	if err != nil {
		err = fmt.Errorf("failed updating task document in Firestore: %w", err)
		l.Error(err.Error(), zap.String("taskId", req.TaskId))
		return nil, err
	}

	// Assemble and return our response
	l.Info("task updated successfully", zap.String("taskId", req.TaskId), zap.String("path", ref.Path))
	return &pbfulfillment.UpdateTaskStatusResponse{}, nil
}

// GetTaskByID retrieves a task matching the specified UUID ID in the fulfillment.GetTaskByIDRequest.
func (fs *FulfillmentService) GetTaskByID(ctx context.Context, req *pbfulfillment.GetTaskByIDRequest) (*pbfulfillment.GetTaskByIDResponse, error) {

	// TODO: Access control

	// Obtain a shortcut handle on our globally configured logger and log some context information
	l := zap.L()
	l.Info("retrieving task", zap.String("taskId", req.TaskId))

	// Form an task structure to receive the data from the store
	task := &schema.Task{Id: req.TaskId}

	// Ask the firestore client for the specified task
	ref := fs.FsClient.Doc(task.StoreRefPath())
	snap, err := fs.drProxy.Get(ref, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve task snapshot with ID %s: %w", req.TaskId, err)
	}

	// Unmarshall the snapshot into our internal structure form
	err = fs.dsProxy.DataTo(snap, task)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal task snapshot with ID %s: %w", req.TaskId, err)
	}

	// Convert the internal task structure ro it protobuf form
	pbTask := task.AsPBTask()

	// Wrap the task in the response structure and we are done
	l.Info("task retrieved successfully", zap.String("taskId", req.TaskId), zap.String("path", ref.Path))
	return &pbfulfillment.GetTaskByIDResponse{
		Task: pbTask,
	}, nil
}

// GetTasks retrieves a page of tasks matching the search criteria specified UUID ID in the fulfillment.GetTasksRequest.
func (fs *FulfillmentService) GetTasks(ctx context.Context, req *pbfulfillment.GetTasksRequest) (*pbfulfillment.GetTasksResponse, error) {

	// Log what we have been asked to do as context for any later logging on this thread. PII values will be masked.
	fs.logQuery(req)

	// Adjust the page size if it is unreasonable
	if req.PageSize < 1 {
		zap.L().Warn("negative/zero page size adjusted to default", zap.Int32("requested", req.PageSize), zap.Int("default", defaultPageSize))
		req.PageSize = defaultPageSize
	} else if req.PageSize > maxPageSize {
		zap.L().Warn("excessive page size adjusted to maximum", zap.Int32("requested", req.PageSize), zap.Int("max", maxPageSize))
		req.PageSize = maxPageSize
	}

	// Start with the whole collection and build up the query from there
	query, err := fs.buildTaskQuery(fs.FsClient.Collection(schema.TaskCollection).Query, req)
	if err != nil {
		return nil, err
	}

	// Run the query to the set of matching tasks. Casting the page size is safe - we just checked that it lies between 1 and 100
	tasks, nextPageToken, err := fs.executeQuery(ctx, query, int(req.PageSize))
	if err != nil {
		return nil, err
	}

	// Loop, converting the slice of internal format tasks to their protobuf equivalents
	pbTasks := make([]*pbfulfillment.Task, len(tasks))
	for i, task := range tasks {
		pbTasks[i] = task.AsPBTask()
	}

	// And we are all done
	zap.L().Info("tasks retrieved successfully", zap.Int("count", len(tasks)), zap.String("nextPageToken", nextPageToken))
	return &pbfulfillment.GetTasksResponse{
		Tasks:         pbTasks,
		NextPageToken: nextPageToken,
	}, nil
}

// executeQuery uses the supplied query to obtain an task document iterator, then build a slice of
// results from that.
func (fs *FulfillmentService) executeQuery(ctx context.Context, query firestore.Query, pageSize int) ([]*schema.Task, string, error) {

	// Run the query to obtain an iterator over the matching documents
	docs := fs.queryProxy.Documents(ctx, query)

	// Close the iterator when we are done with it regardless of whether we are successful or not
	defer docs.Stop()

	// Do the iteration: gather a slice containing all the internal package task representations
	var tasks []*schema.Task
	for {
		// Get the next document, if there is one
		item := &schema.Task{}
		err := docs.Next(item)
		if err != nil {

			// We are either out of documents or have a real error
			if err != iterator.Done {
				return nil, "", fmt.Errorf("failed to retrieve task matching query: %w", err)
			}

			//  For better or worse, we are done with this collection
			break
		}

		// Add this one to our result set and loop around for the next
		tasks = append(tasks, item)
	}

	// Could there be another page? The tasks slice should never be longer than pageSize, but we test for >= defensively
	nextPageToken := ""
	orderCount := len(tasks)
	if orderCount >= pageSize {

		// There could be more to load, use the last processed task's ID as our position marker
		lastTask := tasks[orderCount-1]
		nextPageToken = fmt.Sprintf("%x,%s", lastTask.SubmissionTime.UnixNano(), tasks[orderCount-1].Id)
	}

	// All is well if we reach this point
	return tasks, nextPageToken, nil
}

// buildTaskQuery translates the fulfillment.GetTasksRequest parameters into filters on the given firestore.Query.
// Always check the error return value - a query is returned whether an error occurred or not.
func (fs *FulfillmentService) buildTaskQuery(query firestore.Query, req *pbfulfillment.GetTasksRequest) (firestore.Query, error) {

	// Ignore documents that fall outside the time window first then narrow the result set down from there
	// if the caller specified that mach in their request ...
	if req.StartTime != nil {
		query = query.Where("submissionTime", ">=", req.GetStartTime().AsTime())
	}
	if req.EndTime != nil {
		query = query.Where("submissionTime", "<", req.GetEndTime().AsTime())
	}
	if len(req.OrderId) > 0 {
		query = query.Where("orderId", "==", req.OrderId)
	}
	if len(req.OrderItemId) > 0 {
		query = query.Where("orderItemId", "==", req.OrderItemId)
	}
	if len(req.ProductCode) > 0 {
		query = query.Where("productCode", "==", req.ProductCode)
	}

	// Order the results by submission time first, then by task ID (necessary for us to have a unique cursor position for paging)
	query = query.OrderBy("submissionTime", firestore.Asc).OrderBy("id", firestore.Asc)

	// If a page token was specified, use that as the marker for the last document that has
	// already been returned, i.e. start after that one.
	if len(req.PageToken) > 0 {

		// Split the page token into its two parts - there should be two parts, submission time and ID
		afterTime, afterTaskId, err := splitPageToken(req.PageToken)
		if err != nil {
			return query, err
		}

		// Add the start after factors to our query
		query = query.StartAfter(afterTime, afterTaskId)
	}

	// Limit the size of the result set to the page size
	query = query.Limit(int(req.PageSize))

	// All done
	return query, nil
}

// splitPageToken breaks a page token string into its time and task ID components.
func splitPageToken(token string) (time.Time, string, error) {

	// Split the page token into its two parts - there should be two parts
	parts := strings.Split(token, ",")
	if len(parts) == 2 {

		// Convert the first part from a Unix time value to a time.Time
		unixNanoTime, err := strconv.ParseInt(parts[0], 16, 64)
		if err == nil {

			// All looks good (as shallowly as we are bothering to look), return what we have
			return time.Unix(0, unixNanoTime), parts[1], nil
		}
	}

	// Only arrive here if the page token is invalid
	return time.Time{}, "", fmt.Errorf("invalid page token: %s", token)
}

// logQuery writes a log entry documenting the attributes that make up a FulfillmentService.GetTasks Firestore query.
// PII fields, i.e. family name, will be one way encrypted so that they can be safely logged (hashed with
// a salt, encoded as base65, and truncated to 12 characters).
func (fs *FulfillmentService) logQuery(req *pbfulfillment.GetTasksRequest) {

	// Build a slice of zap.Fields with whatever query parameters are found in the request
	var fields []zap.Field
	if req.StartTime != nil {
		fields = append(fields, zap.Time("startTime", req.StartTime.AsTime()))
	}
	if req.EndTime != nil {
		fields = append(fields, zap.Time("endTime", req.EndTime.AsTime()))
	}
	if len(req.OrderId) > 0 {
		fields = append(fields, zap.String("orderId", req.OrderId))
	}
	if len(req.OrderItemId) > 0 {
		fields = append(fields, zap.String("orderItemId", req.OrderItemId))
	}
	if len(req.ProductCode) > 0 {
		fields = append(fields, zap.String("productCode", req.ProductCode))
	}
	if len(req.PageToken) > 0 {
		fields = append(fields, zap.String("pageToken", req.PageToken))
	}

	// Log that slice and we are done
	zap.L().Info("get task", fields...)
}
