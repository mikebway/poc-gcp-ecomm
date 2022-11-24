package orderfromcart

import (
	"encoding/base64"
	"errors"
	"github.com/mikebway/poc-gcp-ecomm/util"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/golang/protobuf/proto"
	carts "github.com/mikebway/poc-gcp-ecomm/cart/schema"
	"github.com/mikebway/poc-gcp-ecomm/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
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

// init is used to initialize the shopping cart item "constant" values that cannot be declared
// as literal constants.
func init() {

	// Shopping cart values
	t, _ := types.TimestampFromRFC3339Nano(earlyTimeString)
	shoppingCartCreationTime = t.GetTime()
	t, _ = types.TimestampFromRFC3339Nano(lateTimeString)
	shoppingCartClosedTime = t.GetTime()
}

// TestOrderFromCartHappyPath exercises the main handler function with good data that should be processed
// without error.
func TestOrderFromCartHappyPath(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := httptest.NewRequest("POST", "/", strings.NewReader(mockShoppingCartBase64()))
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := util.CaptureLogging(func() {
		OrderFromCart(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(responseRecorder.Code, http.StatusOK, "should have a 200 OK response code")
	req.Contains(logged, "order received", "should have seen the expected completion message in the logs")
	req.Contains(logged, "\"id\": \"d1cecab3-5bc0-43d4-aef1-99ad69794313\"", "should have seen the expected order ID in the logs")
}

// TestOrderFromCartSadPath exercises the main handler function with bad data that should result in an error response.
func TestOrderFromCartSadPath(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Assemble a mock HTTP request and a means to record the response
	httpRequest := httptest.NewRequest("POST", "/", strings.NewReader("this is not a valid base64 protobuf message"))
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := util.CaptureLogging(func() {
		OrderFromCart(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(responseRecorder.Code, http.StatusBadRequest, "should have a 400 Bad Request response code")
	req.Contains(responseRecorder.Body.String(), "unable to decode base64 bytes", "should have seen the expected invalid base64 encoding message in the logs")
	req.Contains(logged, "unable to decode base64 bytes", "should have seen the expected invalid base64 encoding message in the logs")
}

// TestWrongBinary looks at what happens when a valid base64 string is passed to the order loader but
// is not a shopping cart message.
func TestWrongBinary(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Assemble a mock HTTP request and a means to record the response
	base64Body := base64.StdEncoding.EncodeToString([]byte("this is not a valid protobuf shopping cart"))
	httpRequest := httptest.NewRequest("POST", "/", strings.NewReader(base64Body))
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := util.CaptureLogging(func() {
		OrderFromCart(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(responseRecorder.Code, http.StatusBadRequest, "should have a 400 Bad Request response code")
	req.Contains(responseRecorder.Body.String(), "failed to unmarshal shopping cart protobuf message", "should have seen the expected protobuf unmarshal error in the logs")
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
	base64Body := base64.StdEncoding.EncodeToString(pbBytes)
	httpRequest := httptest.NewRequest("POST", "/", strings.NewReader(base64Body))
	responseRecorder := httptest.NewRecorder()

	// Wrap a call to the target function so that we can capture its log output
	logged := util.CaptureLogging(func() {
		OrderFromCart(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(responseRecorder.Code, http.StatusBadRequest, "should have a 400 Bad Request response code")
	req.Contains(responseRecorder.Body.String(), "cart does not have an ID", "should have seen the expected missing ID error in the logs")
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
	logged := util.CaptureLogging(func() {
		OrderFromCart(responseRecorder, httpRequest)
	})

	// Confirm the result was a happy one
	req.Equal(responseRecorder.Code, http.StatusBadRequest, "should have a 400 Bad Request response code")
	req.Contains(responseRecorder.Body.String(), "unable to read base64 bytes", "should have seen the expected read failure error in the logs")
	req.Contains(logged, "unable to read base64 bytes", "should have seen the expected read failure error in the logs")
}

// BadReader implements the io.Reader interface but deliberately fails every time anyone tries to read from it.
type BadReader struct {
	io.Reader
}

// Read is BadReader being bad at reading but good at helping to test how read failures are handled.
func (r BadReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("i am a bad reader")
}

// mockShoppingCartBase64 returns a protobuf binary message encoded as a base64 string for a checked out shopping
// cart structure as its value and with the update time being later than its creation time.
func mockShoppingCartBase64() string {

	// Get the binary bytes of a shopping cart as a protobuf message
	bytes := mockShoppingCartPB()

	// Return that as a base64 encoded string
	return base64.StdEncoding.EncodeToString(bytes)
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
