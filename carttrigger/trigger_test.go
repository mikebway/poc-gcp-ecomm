package carttrigger

import (
	"cloud.google.com/go/pubsub"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/mikebway/poc-gcp-ecomm/cart/schema"
	"github.com/mikebway/poc-gcp-ecomm/cart/service"
	pbcart "github.com/mikebway/poc-gcp-ecomm/pb/cart"
	pbtypes "github.com/mikebway/poc-gcp-ecomm/pb/types"
	"github.com/mikebway/poc-gcp-ecomm/types"
	"github.com/mikebway/poc-gcp-ecomm/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	pbmoney "google.golang.org/genproto/googleapis/type/money"
	"os"
	"strconv"
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

	// EnvPubSubEmulator defines the environment variable name that is used to convey that the Pub/Sub emulator
	// is running, should be used, and how to connect to it
	EnvPubSubEmulator = "PUBSUB_EMULATOR_HOST"

	// PubSubEmulatorHost defines the server name and port (in TCP6 terms) of the Pub/Sub emulator
	PubSubEmulatorHost = "[::1]:8085"

	// EnvPubSubProjectId defines the environment variable name that is used to convey which project the
	// Pub/Sub emulator believes itself to be running under
	EnvPubSubProjectId = "PUBSUB_PROJECT_ID"

	// A couple of timestamp strings we can use to derive known time values
	earlyTimeString = "2022-10-29T16:23:19.123456789-06:00"
	lateTimeString  = "2022-10-30T09:28:42.987654321-06:00"

	// Define the person fields that we will use multiple times to define a shopper
	shopperId          = "10615145-2010-4c5f-8347-2bb556232c31"
	shopperFamilyName  = "Grint"
	shopperGivenName   = "Rupert"
	shopperMiddleName  = "Alexander Lloyd"
	shopperDisplayName = "Rupert"

	// Define the postal address fields for the home address of our mock shopper
	addrLine1      = "55 Yonder St"
	addrLine2      = "Flat B"
	addrLocality   = "Ottery St Catchpole"
	addrPostalCode = "EX11 1HF"
	addrRegionCode = "GB"

	// Define the cart item fields for our mock cart item
	cartItemPriceCurrency       = "USD"
	cartItemProductCode1        = "gold_yoyo"
	cartItemQuantity1     int32 = 3
	cartItemPriceUnits1   int64 = 1_651
	cartItemPriceNanos1   int32 = 940_000_000
	cartItemProductCode2        = "plastic_yoyo"
	cartItemQuantity2     int32 = 13
	cartItemPriceUnits2   int64 = 1
	cartItemPriceNanos2   int32 = 990_000_000
)

var (
	// Shopping cart value types that cannot be declared as constants
	shoppingCartCreationTime time.Time
	shoppingCartClosedTime   time.Time

	// Firestore value times
	firestoreValueCreateTime time.Time
	firestoreValueUpdateTime time.Time

	// A UUID string value retrieved from the cart we register as the foundation of amy
	// test that has to load data from Firestore (i.e. a lot of them)
	// shoppingCartId string
)

var (
	// cartService allows unit tests to write populated shopping carts to Firestore
	cartService *service.CartService

	// mockError is used as an error result when we wish to have our mock Firestore client pretend to fail
	mockError error

	// checkedOutCartId is the ID of a cart that we have written to Firestore so that it can be referenced
	// in multiple unit tests rather than creating new carts every time.
	checkedOutCartId string
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

	// Instantiate our cart service - panic if we cannot obtain one
	var err error
	cartService, err = service.NewCartService()
	if err != nil {
		zap.L().Panic("unable to instantiate cart service / firestore client", zap.Error(err))
	}

	// Create our Pub/Sub topic if it does not already exist
	err = createPubSubTopic()
	if err != nil {
		zap.L().Panic("unable to create pubsub topic", zap.Error(err))
	}

	// Populate our mock error
	mockError = errors.New("this is a mock error")

	// Shopping cart values
	t, _ := types.TimestampFromRFC3339Nano(earlyTimeString)
	shoppingCartCreationTime = t.GetTime()
	t, _ = types.TimestampFromRFC3339Nano(lateTimeString)
	shoppingCartClosedTime = t.GetTime()

	// Firestore value times are closely but not exactly related to the shopping cart times
	// We just make them a second apart here but in reality it would be much closer.
	firestoreValueCreateTime = shoppingCartCreationTime.Add(time.Second)
	firestoreValueUpdateTime = shoppingCartClosedTime.Add(time.Second)

	// Make sure we have a checked out cart in Firestore that we can reference in multiple tests
	checkedOutCartId = storeMockCart(true)

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
	event := mockFirestoreEvent(checkedOutCartId)
	ctx := context.Background()
	var err error
	logged := util.CaptureLogging(func() {
		err = UpdateTrigger(ctx, *event)
	})

	// There should have been no errors and some straightforward log output
	req.Nil(err, "no error was expected: %v", err)
	req.Contains(logged, "published checked out cart", "did not see happy path log message")
	req.Contains(logged, checkedOutCartId, "did not see cart ID in log message")

	// Repeat a second time (would never happen for the same cart in real life) in order
	// to exercise the already loaded paths of the cart service and pubsub client lazy loaders.
	logged = util.CaptureLogging(func() {
		err = UpdateTrigger(ctx, *event)
	})
	req.Nil(err, "no error was expected on second run: %v", err)
	req.Contains(logged, "published checked out cart", "did not see happy path log message on second run")
	req.Contains(logged, checkedOutCartId, "did not see cart ID in log message on second run")
}

// TestNotCheckedOut looks at what happens when a cart update triggers the handler but the cart in
// question has not yet been checked out (hint: we should not publish that cart).
func TestNotCheckedOut(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Configure a Firestore event where the "new value" status is not "checked out"
	cartId := uuid.New().String()
	event := mockFirestoreEvent(cartId)
	event.Value.Fields.Status.IntegerValue = strconv.FormatInt(int64(schema.CsAbandonedByUser), 10)

	// Submit the FirestoreEvent to the handler while capturing its log output
	ctx := context.Background()
	var err error
	logged := util.CaptureLogging(func() {
		err = UpdateTrigger(ctx, *event)
	})

	// There should have been no errors and some straightforward log output
	req.Nil(err, "no error was expected: %v", err)
	req.Contains(logged, "ignoring cart update", "did not see ignored cart log message")
	req.Contains(logged, cartId, "did not see cart ID in log message")
}

// TestCartNotExist looks at what happens when a cart update triggers the handler but the cart in
// question does not exists - can't see how that could happen but it has the side benefit of testing
// onm of the error paths trying to load a cart from Firestore without having to mock an error.
func TestCartNotExist(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Configure a Firestore event where the cart ID won't be found when the trigger function
	// tries to load the full cart.
	cartId := uuid.New().String()
	event := mockFirestoreEvent(cartId)
	ctx := context.Background()
	var err error
	logged := util.CaptureLogging(func() {
		err = UpdateTrigger(ctx, *event)
	})

	// There should have been no errors and some straightforward log output
	req.NotNil(err, "and error was expected")
	req.Contains(logged, "unable to retrieve cart from firestore", "did not see cart retrieval failure log message")
	req.Contains(logged, cartId, "did not see cart ID in log message")
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

	// Submit a checked out cart FirestoreEvent to the handler while capturing its log output
	event := mockFirestoreEvent(checkedOutCartId)
	ctx := context.Background()
	var err error
	logged := util.CaptureLogging(func() {
		err = UpdateTrigger(ctx, *event)
	})

	// There should have been no errors and some straightforward log output
	req.NotNil(err, "and error was expected")
	req.Contains(logged, "pubsub publish failed", "did not see publish failure log message")
	req.Contains(logged, checkedOutCartId, "did not see cart ID in log message")
}

// mockFirestoreEvent constructs a FirestoreEvent populated with known values that we can check in out unit tests.
func mockFirestoreEvent(cartId string) *FirestoreEvent {
	return &FirestoreEvent{
		OldValue: *mockOldValue(cartId),
		Value:    *mockNewValue(cartId),
	}
}

// mockOldValue returns a FirestoreValue populated with an open shopping cart structure as its value and
// with the creation and update times being the same, set a little after the cart creation time.
func mockOldValue(cartId string) *FirestoreValue {

	// Get the "open" shopping cart
	cart := buildMockOpenCart(cartId)

	// Build and return our result - update and creation times will be the sane value
	return &FirestoreValue{
		CreateTime: firestoreValueCreateTime,
		Fields:     *cart,
		Name:       cart.StoreRefPath(),
		UpdateTime: firestoreValueCreateTime,
	}
}

// mockNewValue returns a FirestoreValue populated with a checked out shopping cart structure as its value and
// with the update time being later than its creation time.
func mockNewValue(cartId string) *FirestoreValue {

	// Get the "open" shopping cart and close it
	cart := buildMockOpenCart(cartId)
	cart.Status.IntegerValue = strconv.FormatInt(int64(schema.CsCheckedOut), 10)

	// Build and return our result - update time will be after the creation time
	return &FirestoreValue{
		CreateTime: firestoreValueCreateTime,
		Fields:     *cart,
		Name:       cart.StoreRefPath(),
		UpdateTime: firestoreValueUpdateTime,
	}
}

// StoreRefPath returns the string representation of the document reference path for this ShoppingCart.
func (c *FirestoreCart) StoreRefPath() string {
	return schema.CartCollection + c.Id.StringValue
}

// buildMockOpenCart returns a types.ShoppingCart structure populated with a shopper that can be used to
// test storing new shopping carts in our tests.
func buildMockOpenCart(cartId string) *FirestoreCart {
	return &FirestoreCart{
		Id: StringValue{
			cartId,
		},
		Status: IntegerValue{
			IntegerValue: strconv.FormatInt(int64(schema.CsOpen), 10),
		},
	}
}

// storeMockCart stores a shopping cart in the Firestore emulator so that it can be retrieved when
// we invoke the trigger (i.e. the heart of this Cloud Function) in our tests. The checkedOut
// parameter determines whether the cart should be marked as having been checked out.
//
// This will panic if the cart cannot be saved - our unit tests cannot run without it so why let them
// run at all if we can't save this corner stone.
func storeMockCart(checkedOut bool) string {

	// Create an empty cart
	ctx := context.Background()
	createResp, err := cartService.CreateShoppingCart(ctx, &pbcart.CreateShoppingCartRequest{Shopper: buildMockShopper()})
	if err != nil {
		zap.L().Panic("failed creating mock cart in firestore", zap.Error(err))
	}

	// Pick up the cart ID that we need for the rest of our operations
	cartId := createResp.Cart.Id

	// Add the delivery address
	_, err = cartService.SetDeliveryAddress(ctx, &pbcart.SetDeliveryAddressRequest{
		CartId: cartId, DeliveryAddress: buildMockDeliveryAddress(),
	})
	if err != nil {
		zap.L().Panic("failed setting delivery address in mock firestore cart", zap.Error(err))
	}

	// Add the first cart item
	_, err = cartService.AddItemToShoppingCart(ctx, &pbcart.AddItemToShoppingCartRequest{
		CartId: cartId, Item: buildMockCartItem(cartItemProductCode1),
	})
	if err != nil {
		zap.L().Panic("failed adding first item to mock firestore cart", zap.Error(err))
	}

	// Add the second cart item
	_, err = cartService.AddItemToShoppingCart(ctx, &pbcart.AddItemToShoppingCartRequest{
		CartId: cartId, Item: buildMockCartItem(cartItemProductCode2),
	})
	if err != nil {
		zap.L().Panic("failed adding second item to mock firestore cart", zap.Error(err))
	}

	// If required, flag the cart as checked out
	if checkedOut {
		_, err := cartService.CheckoutShoppingCart(ctx, &pbcart.CheckoutShoppingCartRequest{CartId: cartId})
		if err != nil {
			zap.L().Panic("failed checking out mock firestore cart", zap.Error(err))
		}
	}

	// And we are done!
	return cartId
}

// buildMockShopper returns a pbtypes.Person structure populated with the shopper constant attributes
// defined at the head of this file to be used to create new shopping carts in our tests.
func buildMockShopper() *pbtypes.Person {
	return &pbtypes.Person{
		Id:          shopperId,
		FamilyName:  shopperFamilyName,
		GivenName:   shopperGivenName,
		MiddleName:  shopperMiddleName,
		DisplayName: shopperDisplayName,
	}
}

// buildMockDeliveryAddress returns a pbtypes.PostalAddress structure populated with the constant
// attributes defined at the head of this file to be used to create new shopping carts in our tests.
func buildMockDeliveryAddress() *pbtypes.PostalAddress {
	return &pbtypes.PostalAddress{
		RegionCode:   addrRegionCode,
		PostalCode:   addrPostalCode,
		Locality:     addrLocality,
		AddressLines: []string{addrLine1, addrLine2},
	}
}

// buildMockCartItem returns a pbcart.CartItem structure populated with the constant attributes
// defined at the head of this file to be used to create new shopping carts in our tests.
//
// Which quantity and price you get depends on which product code you ask for. Implementation
// is crude: ask for a product we didn't code for and you will get cartItemProductCode1 with
// no warning or complaint.
func buildMockCartItem(productCode string) *pbcart.CartItem {

	// Which product are we building for?
	switch productCode {

	case cartItemProductCode2:
		return &pbcart.CartItem{
			ProductCode: cartItemProductCode2,
			Quantity:    cartItemQuantity2,
			UnitPrice: &pbmoney.Money{
				CurrencyCode: cartItemPriceCurrency,
				Units:        cartItemPriceUnits2,
				Nanos:        cartItemPriceNanos2,
			},
		}

	case cartItemProductCode1:
		fallthrough
	default:
		// If we don't recognize the product type we were asked for we will return cartItemProductCode1 (gold yoyo)
		return &pbcart.CartItem{
			ProductCode: cartItemProductCode1,
			Quantity:    cartItemQuantity1,
			UnitPrice: &pbmoney.Money{
				CurrencyCode: cartItemPriceCurrency,
				Units:        cartItemPriceUnits1,
				Nanos:        cartItemPriceNanos1,
			},
		}
	}
}
