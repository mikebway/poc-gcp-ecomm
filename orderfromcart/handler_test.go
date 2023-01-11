package orderfromcart

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	carts "github.com/mikebway/poc-gcp-ecomm/cart/schema"
	"github.com/mikebway/poc-gcp-ecomm/order/orderapi"
	"github.com/mikebway/poc-gcp-ecomm/order/schema"
	pborder "github.com/mikebway/poc-gcp-ecomm/pb/order"
	"github.com/mikebway/poc-gcp-ecomm/testutil"
	"github.com/mikebway/poc-gcp-ecomm/types"
	"github.com/stretchr/testify/require"
	pubsubapi "google.golang.org/api/pubsub/v1"
)

const (
	// EnvFirestoreEmulator defines the environment variable name that is used to convey that the Firestore emulator
	// is running, should be used, and how to connect to it
	EnvFirestoreEmulator = "FIRESTORE_EMULATOR_HOST"

	// FirestoreEmulatorHost defines the server name and port (in TCP6 terms) of the Firestore emulator
	FirestoreEmulatorHost = "[::1]:8219"

	// A couple of timestamp strings we can use to derive known time values
	earlyTimeString = "2022-10-29T16:23:19.123456789-06:00"
	lateTimeString  = "2022-10-30T09:28:42.987654321-06:00"

	// A UUID string value that we can use as a shopping cart ID in our tests
	shoppingCartId = "d1cecab3-5bc0-43d4-aef1-99ad69794313"

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

	// Define shopping cart item field values for two shopping cart items
	// As of October 2022, the product codes are for an Apple Mac Studio and Studio Display :-)
	itemId1               = "54f34cb9-fea6-4786-a475-cebd95d93742"
	itemId2               = "b719efe9-4189-453c-96b0-7e229520d316"
	itemProdCode1         = "B09V3G22BH"
	itemProdCode2         = "B09V3GZD32"
	itemQuantity1         = 1
	itemQuantity2         = 2
	itemPriceCurrencyCode = "USD"
	itemPriceUnits1       = 1899
	itemPriceNanos1       = 550_000_000
	itemPriceUnits2       = 1598
	itemPriceNanos2       = 960_000_000
)

var (
	// Shopping cart value types that cannot be declared as constants
	shoppingCartCreationTime time.Time
	shoppingCartClosedTime   time.Time

	// Shopping cart item value types that cannot be declared as constants. Since the
	itemPrice1 = types.Money{CurrencyCode: itemPriceCurrencyCode, Units: itemPriceUnits1, Nanos: itemPriceNanos1}
	itemPrice2 = types.Money{CurrencyCode: itemPriceCurrencyCode, Units: itemPriceUnits2, Nanos: itemPriceNanos2}
)

// TestMain, if defined (it's optional), allows setup code to be run before and after the suite of unit tests
// for this package.
func TestMain(m *testing.M) {

	// Ensure that our Firestore requests do not get routed to the live project by mistake
	orderapi.ProjectId = "demo-" + orderapi.ProjectId

	// Configure the environment variable that informs the Firestore client that it should connect to the
	// emulator and how to reach it.
	_ = os.Setenv(EnvFirestoreEmulator, FirestoreEmulatorHost)

	// Shopping cart values
	t, _ := types.TimestampFromRFC3339Nano(earlyTimeString)
	shoppingCartCreationTime = t.GetTime()
	t, _ = types.TimestampFromRFC3339Nano(lateTimeString)
	shoppingCartClosedTime = t.GetTime()

	// Run all the unit tests
	m.Run()
}

// TestOrderFromCartHappyPath exercises the main handler function with good data that should be processed
// without error.
func TestOrderFromCartHappyPath(t *testing.T) {

	// do the common setup that we share with some other tests, this includes deleting all tasks
	// written to Firestore by others
	req, ctx, svc := commonTestSetup(t)

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := httptest.NewRequest("POST", "/", buildPushRequest(mockShoppingCartPB()))
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		OrderFromCart(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(responseRecorder.Code, http.StatusOK, "should have a 200 OK response code")
	req.Contains(logged, "order received", "should have seen the expected completion message in the logs")
	req.Contains(logged, "\"id\": \"d1cecab3-5bc0-43d4-aef1-99ad69794313\"", "should have seen the expected order ID in the logs")

	// Load the order we just wrote to confirm it all arrived correctly in Firestore
	response, err := svc.GetOrderByID(ctx, &pborder.GetOrderByIDRequest{OrderId: shoppingCartId})
	req.Nil(err, "did not expect an error calling GetOrderByID: %v", err)
	order := response.Order
	req.NotNil(order, "expected GetOrderByID response to contain an order")

	// Confirm all the order values are present as expected
	req.Equal(shoppingCartId, order.Id, "order ID did not match")
	req.Equal(shoppingCartClosedTime.Unix(), order.SubmissionTime.AsTime().Unix(), "submission time does not match")
	req.NotNil(order.OrderedBy, "ordered by person missing")
	req.Equal(shopperId, order.OrderedBy.Id, "ordered by person ID does not match")
	req.Equal(shopperFamilyName, order.OrderedBy.FamilyName, "ordered by person family name does not match")
	req.Equal(shopperGivenName, order.OrderedBy.GivenName, "ordered by person given name does not match")
	req.Equal(shopperMiddleName, order.OrderedBy.MiddleName, "ordered by person middle does not match")
	req.Equal(shopperDisplayName, order.OrderedBy.DisplayName, "ordered by person display name does not match")

	req.NotNil(order.DeliveryAddress, "delivery address missing")
	req.Equal(2, len(order.DeliveryAddress.AddressLines), "delivery address line count does not match")
	req.Equal(addrLine1, order.DeliveryAddress.AddressLines[0], "delivery address line 1 does not match")
	req.Equal(addrLine2, order.DeliveryAddress.AddressLines[1], "delivery address line 2 does not match")
	req.Equal(addrLocality, order.DeliveryAddress.Locality, "delivery address locality does not match")
	req.Equal(addrPostalCode, order.DeliveryAddress.PostalCode, "delivery address postal code does not match")
	req.Equal(addrRegionCode, order.DeliveryAddress.RegionCode, "delivery address region code does not match")

	req.Equal(2, len(order.OrderItems), "order item count does not match")

	req.NotNil(order.OrderItems[0], "order item 1 missing")
	req.Equal(itemId1, order.OrderItems[0].Id, "order item 1 ID does not match")
	req.Equal(itemProdCode1, order.OrderItems[0].ProductCode, "order item 1 product code does not match")
	req.Equal(itemQuantity1, int(order.OrderItems[0].Quantity), "order item 1 quantity does not match")
	req.NotNil(order.OrderItems[0].UnitPrice, "order item 1 unit price missing")
	req.Equal(itemPrice1.CurrencyCode, order.OrderItems[0].UnitPrice.CurrencyCode, "order item 1 price currency does not match")
	req.Equal(itemPrice1.Units, order.OrderItems[0].UnitPrice.Units, "order item 1 price units does not match")
	req.Equal(itemPrice1.Nanos, order.OrderItems[0].UnitPrice.Nanos, "order item 1 price nanos does not match")

	req.NotNil(order.OrderItems[1], "order item 2 missing")
	req.Equal(itemId2, order.OrderItems[1].Id, "order item 2 ID does not match")
	req.Equal(itemProdCode2, order.OrderItems[1].ProductCode, "order item 2 product code does not match")
	req.Equal(itemQuantity2, int(order.OrderItems[1].Quantity), "order item 2 quantity does not match")
	req.NotNil(order.OrderItems[1].UnitPrice, "order item 2 unit price missing")
	req.Equal(itemPrice2.CurrencyCode, order.OrderItems[1].UnitPrice.CurrencyCode, "order item 2 price currency does not match")
	req.Equal(itemPrice2.Units, order.OrderItems[1].UnitPrice.Units, "order item 2 price units does not match")
	req.Equal(itemPrice2.Nanos, order.OrderItems[1].UnitPrice.Nanos, "order item 2 price nanos does not match")
}

// TestInvalidPushRequest exercises the main handler function with an invalid request that does not
// match a Pub/Sub push.
func TestInvalidPushRequest(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := httptest.NewRequest("POST", "/", strings.NewReader("this is not a valid push request"))
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		OrderFromCart(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(responseRecorder.Code, http.StatusBadRequest, "should have a 400 Bad Request response code")
	req.Contains(responseRecorder.Body.String(), "could not decode push request json body", "should have seen the expected invalid push request message in the response")
	req.Contains(logged, "could not decode push request json body", "should have seen the expected invalid base64 encoding message in the logs")
}

// TestInvalidBase64 exercises the main handler function with bad data that should result in an error response.
func TestInvalidBase64(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := httptest.NewRequest("POST", "/", buildPushRequestFromString("this is not a valid base64 payload"))
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		OrderFromCart(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(responseRecorder.Code, http.StatusBadRequest, "should have a 400 Bad Request response code")
	req.Contains(responseRecorder.Body.String(), "unable to decode base64 data", "should have seen the expected invalid push request message in the response")
	req.Contains(logged, "unable to decode base64 data", "should have seen the expected invalid base64 encoding message in the logs")
}

// TestWrongBinary looks at what happens when a valid base64 string is passed to the order loader but
// is not a shopping cart message.
func TestWrongBinary(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := httptest.NewRequest("POST", "/", buildPushRequest([]byte("this is not a valid protobuf shopping cart")))
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		OrderFromCart(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(responseRecorder.Code, http.StatusBadRequest, "should have a 400 Bad Request response code")
	req.Contains(responseRecorder.Body.String(), "failed to unmarshal shopping cart protobuf message", "should have seen the expected protobuf unmarshal error in the response")
	req.Contains(logged, "failed to unmarshal shopping cart protobuf message", "should have seen the expected protobuf unmarshal error in the logs")
}

// TestNoCartID looks at what happens when an otherwise valid protobuf cart message defines a cart with no ID.
func TestNoCartID(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Obtain a shopping cart in our internal form and clear its ID
	cart := buildMockCart()
	cart.Id = ""

	// Convert that to the binary protobuf form
	pbCart := cart.AsPBShoppingCart()
	pbBytes, _ := proto.Marshal(pbCart)

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := httptest.NewRequest("POST", "/", buildPushRequest(pbBytes))
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		OrderFromCart(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(responseRecorder.Code, http.StatusBadRequest, "should have a 400 Bad Request response code")
	req.Contains(responseRecorder.Body.String(), "cart does not have an ID", "should have seen the expected missing ID error in the response")
	req.Contains(logged, "cart does not have an ID", "should have seen the expected missing ID error in the logs")
}

// TestBodyReaderError looks at what happens in the unlikely event that the request.Body cannot be read. Perhaps
// this could happen if TCP error (connection lost?) happened in the middle of reading the body of the POST.
func TestBodyReaderError(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// We don't bother generating a protocol buffer shopping cart, we use a bad reader instead
	badReader := &BadReader{}

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := httptest.NewRequest("POST", "/", badReader)
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		OrderFromCart(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(responseRecorder.Code, http.StatusBadRequest, "should have a 400 Bad Request response code")
	req.Contains(responseRecorder.Body.String(), "could not decode push request json body: i am a bad reader", "should have seen the expected read failure error in the response")
	req.Contains(logged, "could not decode push request json body: i am a bad reader", "should have seen the expected read failure error in the logs")
}

// TestServiceLoadFailure looks at how the code handles being unable to establish a orderapi.OrderService.
func TestServiceLoadFailure(t *testing.T) {

	// Force NewFulfillmentService to fail - be sure to clear that after we are done
	lazyOrderService = nil
	orderapi.UnitTestNewOrderServiceError = errors.New("unit test forced error")
	defer func() { orderapi.UnitTestNewOrderServiceError = nil }()

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := httptest.NewRequest("POST", "/", buildPushRequest(mockShoppingCartPB()))
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		OrderFromCart(responseRecorder, httpRequest)
	})

	// Confirm the result was the sad one that we expected
	req := require.New(t)
	req.Equal(http.StatusInternalServerError, responseRecorder.Code, "should have a 500 internal server error code")
	req.Contains(logged, "could not obtain firestore client", "should have seen the expected firestore client failure message in the logs")
	req.Contains(logged, orderapi.UnitTestNewOrderServiceError.Error(), "should have seen the expected error message in the logs")
}

// TestSaveOrderFailure tricks the handler orderapi.OrderService SaveOrder function into failing
// not using the emulator and so not finding the GCP project referenced by the order service.
func TestSaveOrderFailure(t *testing.T) {

	// Put everything back when it should be when we leave this test
	defer func() {
		lazyOrderService = nil
		_ = os.Setenv(EnvFirestoreEmulator, FirestoreEmulatorHost)
	}()

	// Force the handler to obtain a new fulfillment service ...
	lazyOrderService = nil

	// ... but without using the emulator and targeting a non-existent project
	_ = os.Setenv(EnvFirestoreEmulator, "")

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := httptest.NewRequest("POST", "/", buildPushRequest(mockShoppingCartPB()))
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := testutil.CaptureLogging(func() {
		OrderFromCart(responseRecorder, httpRequest)
	})

	// Confirm the result was the sad one that we expected
	req := require.New(t)
	req.Equal(http.StatusInternalServerError, responseRecorder.Code, "should have a 500 internal server error code")
	req.Contains(logged, "failed creating order document in Firestore", "should have seen the expected SaveOrder failure message in the logs")
}

// commonTestSetup helps us to be a little DRY (Don't Repeat Yourself) in this file, doing the steps that
// several of the unit test functions in here need to do before going on to anything else.
func commonTestSetup(t *testing.T) (*require.Assertions, context.Context, *orderapi.OrderService) {

	// Avoid having to pass t in to every assertion
	assert := require.New(t)

	// Clear any manipulations that might have been made to the order service used by the handler
	lazyOrderService = nil

	// Make sure that Firestore has been depopulated of any orders left over from prior test runs
	ctx := context.Background()
	svc := deleteAllMockOrders(ctx, assert)

	// return everything the caller needs to perform their tests
	return assert, ctx, svc
}

// deleteAllMockOrders removes our mock orders from the Firestore emulator. We use this to ensure that our tests are
// run against a clean slate. Returns the order service used to delete the orders for the caller to use for
// additional tinkering.
func deleteAllMockOrders(ctx context.Context, assert *require.Assertions) *orderapi.OrderService {

	// Obtain a clean instance of the order service, i.e. one that we know has not been monkeyed with
	// to return errors for testing purposes
	svc, err := orderapi.NewOrderService()
	assert.Nil(err, "did not expect an error obtaining a new OrderService: %v", err)

	// At present, as this test suite stands, there is only one order that we need to take care of
	// so no query is required to get its ID - it will have the same ID as the mock shopping cart it
	// was derived from.

	// Establish a document reference for that task path and ask for the document to be deleted
	order := schema.Order{Id: shoppingCartId}
	ref := svc.FsClient.Doc(order.StoreRefPath())
	_, err = ref.Delete(ctx)
	assert.Nil(err, "failed deleting order ID %s: %v", order.Id, err)

	// Return the order service for additional use by our caller
	return svc
}

// BadReader implements the io.Reader interface but deliberately fails every time anyone tries to read from it.
type BadReader struct {
	io.Reader
}

// Read is BadReader being bad at reading but good at helping to test how read failures are handled.
func (r BadReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("i am a bad reader")
}

// buildPushRequest wraps the provided data bytes as the data payload of a push request, returning that as byte reader.
func buildPushRequest(data []byte) io.Reader {

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

// mockShoppingCartPB returns a protobuf binary bytes slice populated with a checked out shopping cart structure as
// its value and with the update time being later than its creation time.
func mockShoppingCartPB() []byte {

	// Get a "checked out" shopping cart in its internal memory form
	cart := buildMockCart()

	// Render that to its protocol buffer structure form
	pbCart := cart.AsPBShoppingCart()

	// Marshal that to its binary message state and return that
	pbBytes, _ := proto.Marshal(pbCart)
	return pbBytes
}

// buildMockCart returns a types.ShoppingCart structure populated with a shopper that can be used to
// test storing new shopping carts in our tests.
func buildMockCart() *carts.ShoppingCart {
	return &carts.ShoppingCart{
		Id:              shoppingCartId,
		CreationTime:    shoppingCartCreationTime,
		ClosedTime:      shoppingCartClosedTime,
		Status:          carts.CsCheckedOut,
		Shopper:         buildMockShopper(),
		DeliveryAddress: buildMockDeliveryAddress(),
		CartItems:       buildMockCartItems(),
	}
}

// buildMockShopper returns a types.Person structure populated with the shopper constant attributes
// defined at the head of this file to be used to create new shopping carts in our tests.
func buildMockShopper() *types.Person {
	return &types.Person{
		Id:          shopperId,
		FamilyName:  shopperFamilyName,
		GivenName:   shopperGivenName,
		MiddleName:  shopperMiddleName,
		DisplayName: shopperDisplayName,
	}
}

// buildMockDeliveryAddress returns a types.PostalAddress structure populated with the constant
// attributes defined at the head of this file to be used to create new shopping carts in our tests.
func buildMockDeliveryAddress() *types.PostalAddress {
	return &types.PostalAddress{
		RegionCode:   addrRegionCode,
		PostalCode:   addrPostalCode,
		Locality:     addrLocality,
		AddressLines: []string{addrLine1, addrLine2},
	}
}

// buildMockCartItems returns an array containing two shopping cart items populated  with the constant
// attributes defined at the head of this file to be used to create new shopping carts in our tests.
func buildMockCartItems() []*carts.ShoppingCartItem {
	return []*carts.ShoppingCartItem{
		&carts.ShoppingCartItem{
			Id:          itemId1,
			CartId:      shoppingCartId,
			ProductCode: itemProdCode1,
			Quantity:    itemQuantity1,
			UnitPrice:   &itemPrice1,
		},
		&carts.ShoppingCartItem{
			Id:          itemId2,
			CartId:      shoppingCartId,
			ProductCode: itemProdCode2,
			Quantity:    itemQuantity2,
			UnitPrice:   &itemPrice2,
		},
	}
}
