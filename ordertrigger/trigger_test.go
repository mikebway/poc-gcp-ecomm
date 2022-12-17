package ordertrigger

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/mikebway/poc-gcp-ecomm/order/schema"
	"github.com/mikebway/poc-gcp-ecomm/order/service"
	"github.com/mikebway/poc-gcp-ecomm/testutil"
	"github.com/mikebway/poc-gcp-ecomm/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"os"
	"testing"
	"time"
)

// TODO: Consolidate repeated unit test support data and functions into a shared module

const (
	// EnvFirestoreEmulator defines the environment variable name that is used to convey that the Firestore emulator
	// is running, should be used, and how to connect to it
	EnvFirestoreEmulator = "FIRESTORE_EMULATOR_HOST"

	// FirestoreEmulatorHost defines the server name and port (in TCP6 terms) of the Firestore emulator
	FirestoreEmulatorHost = "[::1]:8219"

	// A timestamp string we can use to derive known time values. It will be used as the first submission time in the
	// any orders we store.
	timeString = "2022-10-11T10:23:19.000000000-06:00"

	// UnitTestGivenName is used for all given names for the ordering person for all unit test orders written
	// to the Firestore emulator. It is used so that we can find and delete all orders we create in these unit
	// tests (and only those orders) so that we can be sure of starting with clean slate.
	UnitTestGivenName = "Unit~Test"

	// Define the additional person fields that we will for the one mock order where we include more detail
	personId          = "10615145-2010-4c5f-8347-2bb556232c31"
	personFamilyName  = "Grint"
	personMiddleName  = "Alexander Lloyd"
	personDisplayName = "Rupert"

	// Define the postal address fields for the home address of our mock person
	addrLine1      = "55 Yonder St"
	addrLine2      = "Flat B"
	addrLocality   = "Ottery St Catchpole"
	addrPostalCode = "EX11 1HF"
	addrRegionCode = "GB"

	// unitTestErrorMessage is used as the error description for error that are deliberately forced to
	// test error handling.
	unitTestErrorMessage = "unit test of error handling"
)

var (
	// orderService allows unit tests to write populated orders to Firestore
	orderService *service.OrderService

	// ---

	// Shopping order value types that cannot be declared as constants
	shoppingOrderCreationTime time.Time
	shoppingOrderClosedTime   time.Time

	// Firestore value times
	firestoreValueCreateTime time.Time
	firestoreValueUpdateTime time.Time

	// checkedOutOrderId is the ID of a order that we have written to Firestore so that it can be referenced
	// in multiple unit tests rather than creating new carts every time.
	checkedOutOrderId string
)

// TestMain, if defined (it's optional), allows setup code to be run before and after the suite of unit tests
// for this package.
func TestMain(m *testing.M) {

	// Ensure that our Firestore and Pub/Sub requests doe not get routed to the live project by mistake
	service.ProjectId = "demo-" + service.ProjectId
	TopicProjectId = "demo-" + TopicProjectId

	// Configure the environment variable that informs the Firestore client that it should connect to the
	// emulator and how to reach it.
	_ = os.Setenv(EnvFirestoreEmulator, FirestoreEmulatorHost)

	// Do the same for the Pub/Sub emulator
	_ = os.Setenv(EnvPubSubEmulator, PubSubEmulatorHost)
	_ = os.Setenv(EnvPubSubProjectId, TopicProjectId)

	// Instantiate our order service - panic if we cannot obtain one
	var err error
	orderService, err = service.NewOrderService()
	if err != nil {
		zap.L().Panic("unable to instantiate order service / firestore client", zap.Error(err))
	}

	// Create our Pub/Sub topic if it does not already exist
	err = createPubSubTopic()
	if err != nil {
		zap.L().Panic("unable to create pubsub topic", zap.Error(err))
	}

	// Shopping order values
	t, _ := types.TimestampFromRFC3339Nano(earlyTimeString)
	shoppingOrderCreationTime = t.GetTime()
	t, _ = types.TimestampFromRFC3339Nano(lateTimeString)
	shoppingOrderClosedTime = t.GetTime()

	// Firestore value times are closely but not exactly related to the order times
	// We just make them a second apart here but in reality it would be much closer.
	firestoreValueCreateTime = shoppingOrderCreationTime.Add(time.Second)
	firestoreValueUpdateTime = shoppingOrderClosedTime.Add(time.Second)

	// Make sure we have a checked out order in Firestore that we can reference in multiple tests
	checkedOutOrderId = storeMockOrder(true)

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
	event := mockFirestoreEvent(checkedOutOrderId)
	ctx := context.Background()
	var err error
	logged := testutil.CaptureLogging(func() {
		err = UpdateTrigger(ctx, *event)
	})

	// There should have been no errors and some straightforward log output
	req.Nil(err, "no error was expected: %v", err)
	req.Contains(logged, "published checked out order", "did not see happy path log message")
	req.Contains(logged, checkedOutOrderId, "did not see order ID in log message")

	// Repeat a second time (would never happen for the same order in real life) in order
	// to exercise the already loaded paths of the order service and pubsub client lazy loaders.
	logged = testutil.CaptureLogging(func() {
		err = UpdateTrigger(ctx, *event)
	})
	req.Nil(err, "no error was expected on second run: %v", err)
	req.Contains(logged, "published checked out order", "did not see happy path log message on second run")
	req.Contains(logged, checkedOutOrderId, "did not see order ID in log message on second run")
}

// TestOrderNotExist looks at what happens when a order update triggers the handler but the order in
// question does not exists - can't see how that could happen but it has the side benefit of testing
// onm of the error paths trying to load a order from Firestore without having to mock an error.
func TestOrderNotExist(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Configure a Firestore event where the order ID won't be found when the trigger function
	// tries to load the full order.
	orderId := uuid.NewString()
	event := mockFirestoreEvent(orderId)
	ctx := context.Background()
	var err error
	logged := testutil.CaptureLogging(func() {
		err = UpdateTrigger(ctx, *event)
	})

	// There should have been no errors and some straightforward log output
	req.NotNil(err, "an error was expected")
	req.Contains(logged, "unable to retrieve order from firestore", "did not see order retrieval failure log message")
	req.Contains(logged, orderId, "did not see order ID in log message")
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

	// Submit a checked out order FirestoreEvent to the handler while capturing its log output
	event := mockFirestoreEvent(checkedOutOrderId)
	ctx := context.Background()
	var err error
	logged := testutil.CaptureLogging(func() {
		err = UpdateTrigger(ctx, *event)
	})

	// There should have been no errors and some straightforward log output
	req.NotNil(err, "an error was expected")
	req.Contains(logged, "pubsub publish failed", "did not see publish failure log message")
	req.Contains(logged, checkedOutOrderId, "did not see order ID in log message")
}

// mockFirestoreEvent constructs a FirestoreEvent populated with known values that we can check in out unit tests.
func mockFirestoreEvent(orderId string) *FirestoreEvent {
	return &FirestoreEvent{
		Value: *mockNewValue(orderId),
	}
}

// mockNewValue returns a FirestoreValue populated with a checked out order structure as its value and
// with the update time being later than its creation time.
func mockNewValue(orderId string) *FirestoreValue {

	// Get an order in Firestore trigger image form (minimally populated)
	order := buildMockTriggerOrder(orderId)

	// Build and return our result - update time will be after the creation time
	return &FirestoreValue{
		CreateTime: firestoreValueCreateTime,
		Fields:     *order,
		Name:       order.StoreRefPath(),
		UpdateTime: firestoreValueUpdateTime,
	}
}

// StoreRefPath returns the string representation of the document reference path for this Order.
func (c *FirestoreOrder) StoreRefPath() string {
	return schema.OrderCollection + c.Id.StringValue
}

// buildMockTriggerOrder returns a minimally populated FirestoreOrder structure.
func buildMockTriggerOrder(orderId string) *FirestoreOrder {
	return &FirestoreOrder{
		Id: StringValue{
			orderId,
		},
	}
}

// storeMockOrder stores an order in the Firestore emulator so that it can be retrieved when
// we invoke the trigger (i.e. the heart of this Cloud Function) in our tests.
//
// This will panic if the order cannot be saved - our unit tests cannot run without it so why let them
// run at all if we can't save this corner stone.
func storeMockOrder() string {

	// Build a mock order that we can write to Firestore
	order := buildMockOrder()

	// Write that to the Firestore database
	ctx := context.Background()
	err := orderService.SaveOrder(ctx, order)
	if err != nil {
		zap.L().Panic("failed saving mock order to firestore", zap.Error(err))
	}

	// And we are done!
	return order.Id
}

// buildMockOrder returns a Order structure populated with a person that can be used to
// test storing new shopping carts in our tests.
func buildMockOrder() *schema.Order {
	timestamp, _ := types.TimestampFromRFC3339Nano(timeString)
	return &schema.Order{
		Id:             uuid.NewString(),
		SubmissionTime: timestamp.GetTime(),
		OrderedBy: &types.Person{
			Id:          personId,
			FamilyName:  personFamilyName,
			GivenName:   UnitTestGivenName,
			MiddleName:  personMiddleName,
			DisplayName: personDisplayName,
		},
		DeliveryAddress: &types.PostalAddress{
			RegionCode:   addrRegionCode,
			PostalCode:   addrPostalCode,
			Locality:     addrLocality,
			AddressLines: []string{addrLine1, addrLine2},
		},
		OrderItems: []*schema.OrderItem{
			{Id: uuid.NewString(), ProductCode: "second_1", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 2, Nanos: 100_000_000}},
			{Id: uuid.NewString(), ProductCode: "second_2", Quantity: 1, UnitPrice: &types.Money{CurrencyCode: "USD", Units: 2, Nanos: 200_000_000}},
		},
	}
}
