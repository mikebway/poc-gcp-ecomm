package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/mikebway/poc-gcp-ecomm/cart/fsproxies"
	pbcart "github.com/mikebway/poc-gcp-ecomm/pb/cart"
	pbtypes "github.com/mikebway/poc-gcp-ecomm/pb/types"
	"github.com/stretchr/testify/require"
	pbmoney "google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"testing"
)

const (
	// EnvFirestoreEmulator defines the environment variable name that is used to convey that the Firestore emulator
	// is running, should be used, and how to connect to it
	EnvFirestoreEmulator = "FIRESTORE_EMULATOR_HOST"

	// FirestoreEmulatorHost defines the server name and port (in TCP6 terms) of the Firestore emulator
	FirestoreEmulatorHost = "[::1]:8219"

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
	// mockError is used as an error result when we wish to have our mock Firestore client pretend to fail
	mockError error
)

// init performs static initialization of our constants that cannot actually be literal constants
func init() {
	mockError = errors.New("this is a mock error")
}

// TestMain, if defined (it's optional), allows setup code to be run before and after the suite of unit tests
// for this package.
func TestMain(m *testing.M) {

	// Ensure that our Firestore requests do not get routed to the live project by mistake
	ProjectId = "demo-" + ProjectId

	// Configure the environment variable that informs the Firestore client that it should connect to the
	// emulator and how to reach it.
	_ = os.Setenv(EnvFirestoreEmulator, FirestoreEmulatorHost)

	// Run all the unit tests
	m.Run()
}

// TestCreateAndGetCart examines whether a new cart can be properly added to the database
func TestCreateAndGetCart(t *testing.T) {

	// The cart creation time will be greater than or equal to the current time now
	startTime := timestamppb.Now()

	// Do the basic foundation stuff that most of this package's tests require
	req, ctx, service, cart := commonTestSetup(t)

	// Evaluate the configuration of the just created shopping cart
	req.NotEmpty(cart.Id, "created cart should an ID")
	req.Equal(36, len(cart.Id), "created cart ID does not look like a UUID")
	req.NotNil(cart.CreationTime, "cart creation time should have been set for Rupert's new cart")
	req.GreaterOrEqual(cart.CreationTime.AsTime(), startTime.AsTime(), "cart creation time should be at or after the time that this test started")
	req.Equal(pbcart.ShoppingCartStatus_SCS_OPEN, cart.Status, "cart status should be 'open'")
	req.Equal(shopperId, cart.Shopper.Id, "created shopper ID did not match")
	req.Equal(shopperFamilyName, cart.Shopper.FamilyName, "created shopper family name did not match")
	req.Equal(shopperGivenName, cart.Shopper.GivenName, "created shopper given name did not match")
	req.Equal(shopperMiddleName, cart.Shopper.MiddleName, "created shopper middle name(s) did not match")
	req.Equal(shopperDisplayName, cart.Shopper.DisplayName, "created shopper display name did not match")

	// Should be able to retrieve the new record
	getResp, err := service.GetShoppingCartByID(ctx, &pbcart.GetShoppingCartByIDRequest{CartId: cart.Id})
	req.Nil(err, "should not have seen an error asking for Rupert's new cart: %v", err)
	req.NotNil(getResp, "should have obtained a response asking for Rupert's new cart")
	req.NotEmpty(getResp.GetCart().Id, "retrieved new cart should an ID")
	req.Equal(36, len(getResp.GetCart().Id), "retrieved new cart ID does not look like a UUID")
	req.Equal(shopperId, getResp.GetCart().Shopper.Id, "retrieved new shopper ID did not match")
	req.Equal(shopperFamilyName, getResp.GetCart().Shopper.FamilyName, "retrieved new shopper family name did not match")
	req.Equal(shopperGivenName, getResp.GetCart().Shopper.GivenName, "retrieved new shopper given name did not match")
	req.Equal(shopperMiddleName, getResp.GetCart().Shopper.MiddleName, "retrieved new shopper middle name(s) did not match")
	req.Equal(shopperDisplayName, getResp.GetCart().Shopper.DisplayName, "retrieved new shopper display name did not match")
}

// TestCartCreateFailure examines what happens if the Firestore Create fails creating a new cart.
func TestCartCreateFailure(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Initialize our target cart service
	ctx := context.Background()
	service, err := NewCartService()
	req.Nil(err, "failed to obtain cart service: %v", err)

	// Modify the cart service to return an error on any and every DocumentRef operation
	service.drProxy = &fsproxies.UTDocRefProxy{Err: mockError, AllowCount: 0}

	// Attempt to create a new cart
	createResp, err := service.CreateShoppingCart(ctx, &pbcart.CreateShoppingCartRequest{Shopper: buildMockShopper()})
	req.NotNil(err, "should have seen an error creating a new cart")
	req.Contains(err.Error(), mockError.Error(), "should have seen the configured error creating a new cart")
	req.Nil(createResp, "should not have obtained a response after failing to create a new cart")
}

// TestCartNotFound looks at what happens when trying to retrieve a shopping cart that does not exist.
func TestCartNotFound(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Initialize our target cart service
	ctx := context.Background()
	service, err := NewCartService()
	req.Nil(err, "failed to obtain cart service: %v", err)

	// Ask for the data for a bogus person ID
	const bogusId = "not-even-a-uuid"
	getResp, err := service.GetShoppingCartByID(ctx, &pbcart.GetShoppingCartByIDRequest{CartId: bogusId})
	req.NotNil(err, "should have seen an error asking for a cart with a bogus ID")
	req.Contains(err.Error(), "rpc error: code = NotFound", "did not get the expected error message")
	req.Contains(err.Error(), bogusId, "error message did not include the bogus ID")
	req.Nil(getResp, "should not have obtained a response asking for a cart with a bogus ID")
}

// func TestCartCorrupt looks at what happens when trying to retrieve a shopping cart that exists but for
// which the data is somehow corrupt and cannot be unmarshalled from the retrieved snapshot.
func TestCartCorrupt(t *testing.T) {

	// Do the basic foundation stuff that most of this package's tests require
	req, ctx, service, cart := commonTestSetup(t)

	// Modify the cart service to return a snapshot unmarshalling error every time we try to unmarshall a snapshot
	service.dsProxy = &fsproxies.UTDocSnapProxy{Err: mockError, AllowCount: 0}

	// Ask for the cart again
	getResp, err := service.GetShoppingCartByID(ctx, &pbcart.GetShoppingCartByIDRequest{CartId: cart.Id})
	req.NotNil(err, "should have seen an error asking for a corrupt cart")
	req.Contains(err.Error(), "failed to unmarshal cart snapshot with ID "+cart.Id, "did not get the expected error message")
	req.Nil(getResp, "should not have obtained a response asking for a corrupt cart")
}

// TestCartItemCorrupt looks at what happens when the cart itself can be retrieved but one of the items
// in the cart is corrupt and cannot be unmarshalled.
func TestCartItemCorrupt(t *testing.T) {

	// Do the basic foundation stuff that most of this package's tests require
	req, ctx, service, cart := commonTestSetup(t)

	// Modify the cart service to return a snapshot unmarshalling error every time we try to unmarshall an item snapshot
	service.itemsGetterProxy = &fsproxies.UTItemCollGetterProxy{FsClient: service.FsClient, Err: mockError, AllowCount: 0}

	// Ask for the cart again
	getResp, err := service.GetShoppingCartByID(ctx, &pbcart.GetShoppingCartByIDRequest{CartId: cart.Id})
	req.NotNil(err, "should have seen an error asking for a corrupt cart")
	req.Contains(err.Error(), "failed to retrieve cart item for cart with ID "+cart.Id, "did not get the expected error message")
	req.Nil(getResp, "should not have obtained a response asking for a cart with a corrupt item")
}

// TestSetDeliveryAddress confirms that an address can be added to the cart, that the amended cart gets returned
// with the response, and that the delivery address is present in a follow-up get of the cart.
func TestSetDeliveryAddress(t *testing.T) {

	// Do the basic foundation stuff that most of this package's tests require
	req, ctx, service, cart := commonTestSetup(t)

	// Add a delivery address to Harry Potter's cart
	deliveryAddress := buildMockDeliveryAddress()
	setResponse, err := service.SetDeliveryAddress(ctx, &pbcart.SetDeliveryAddressRequest{
		CartId:          cart.Id,
		DeliveryAddress: deliveryAddress,
	})

	// Check that everything came back as expected
	req.Nil(err, "should not have seen an error adding an address to Rupert's cart: %v", err)
	req.NotNil(setResponse, "should have obtained a response adding an address to Rupert's cart")
	setResponseCart := setResponse.Cart
	req.NotNil(setResponseCart, "response adding an address to Rupert's cart should have included a cart")
	req.Equal(cart.Id, setResponseCart.Id, "response adding an address to Rupert's cart should have had the correct cart ID")
	setResponseAddr := setResponseCart.DeliveryAddress
	req.NotNil(setResponseAddr, "response adding an address to Rupert's cart should have included an address")
	req.Equal(addrRegionCode, setResponseAddr.RegionCode, "set delivery address response did not have correct region code")
	req.Equal(addrPostalCode, setResponseAddr.PostalCode, "set delivery address response did not have correct region code")
	req.Equal(addrLocality, setResponseAddr.Locality, "set delivery address response did not have correct locality")
	req.Equal(addrRegionCode, setResponseAddr.RegionCode, "set delivery address response did not have correct region code")
	req.Equal(2, len(setResponseAddr.AddressLines), "set delivery address response did not have correct address line count")
	req.Equal(addrLine1, setResponseAddr.AddressLines[0], "set delivery address response did not have correct address line 1")
	req.Equal(addrLine2, setResponseAddr.AddressLines[1], "set delivery address response did not have correct address line 2")

	// Npw get the cart and confirm that the address is in place in the store
	getResponse, err := service.GetShoppingCartByID(ctx, &pbcart.GetShoppingCartByIDRequest{CartId: cart.Id})
	req.Nil(err, "should not have seen an error getting  an address to Rupert's cart: %v", err)
	req.NotNil(getResponse, "should have obtained a response getting Rupert's cart")
	getResponseCart := getResponse.Cart
	req.NotNil(getResponseCart, "response getting Rupert's cart should have included a cart")
	req.Equal(cart.Id, getResponseCart.Id, "response getting Rupert's cart should have had the correct cart ID")
	getResponseAddr := getResponseCart.DeliveryAddress
	req.NotNil(getResponseAddr, "response getting Rupert's cart should have included an address")
	req.Equal(addrRegionCode, getResponseAddr.RegionCode, "getting Rupert's cart response did not have correct region code")
	req.Equal(addrPostalCode, getResponseAddr.PostalCode, "getting Rupert's cart response did not have correct region code")
	req.Equal(addrLocality, getResponseAddr.Locality, "getting Rupert's cart response did not have correct locality")
	req.Equal(addrRegionCode, getResponseAddr.RegionCode, "getting Rupert's cart response did not have correct region code")
	req.Equal(2, len(getResponseAddr.AddressLines), "getting Rupert's cart response did not have correct address line count")
	req.Equal(addrLine1, getResponseAddr.AddressLines[0], "getting Rupert's cart response did not have correct address line 1")
	req.Equal(addrLine2, getResponseAddr.AddressLines[1], "getting Rupert's cart response did not have correct address line 2")
}

// TestDeliveryAddressSetFailure examines what happens if the Firestore Get fails setting the delivery address
// associated with a cart
func TestDeliveryAddressSetFailure(t *testing.T) {

	// Do the basic foundation stuff that most of this package's tests require
	req, ctx, service, cart := commonTestSetup(t)

	// Modify the cart service to return an error on any and every DocumentRef operation
	service.drProxy = &fsproxies.UTDocRefProxy{Err: mockError, AllowCount: 0}

	// Add a delivery address to Harry Potter's cart
	deliveryAddress := buildMockDeliveryAddress()
	setResponse, err := service.SetDeliveryAddress(ctx, &pbcart.SetDeliveryAddressRequest{
		CartId:          cart.Id,
		DeliveryAddress: deliveryAddress,
	})

	// Check that everything came back as expected
	req.NotNil(err, "should have seen an error adding an address to Rupert's cart")
	req.Contains(err.Error(), mockError.Error(), "should have seen the configured error adding an address to Rupert's cart")
	req.Nil(setResponse, "should have not obtained a response adding an address to Rupert's cart")
}

// TestDeliveryAddressGetFailure examines what happens if the Firestore Get fails getting the delivery address
// associated with a cart
func TestDeliveryAddressGetFailure(t *testing.T) {

	// Do the basic foundation stuff that most of this package's tests require
	req, ctx, service, cart := commonTestSetup(t)

	// Add a delivery address to Harry Potter's cart
	deliveryAddress := buildMockDeliveryAddress()
	_, err := service.SetDeliveryAddress(ctx, &pbcart.SetDeliveryAddressRequest{
		CartId:          cart.Id,
		DeliveryAddress: deliveryAddress,
	})

	// Check that everything came back as expected from setting the address
	req.Nil(err, "should not have seen an error adding an address to Rupert's cart")

	// Now, modify the cart service to return an error on the second get from Firestore, i.e. the address
	service.drProxy = &fsproxies.UTDocRefProxy{Err: mockError, AllowCount: 1}

	// .. and try to get the cart
	getResponse, err := service.GetShoppingCartByID(ctx, &pbcart.GetShoppingCartByIDRequest{CartId: cart.Id})
	req.NotNil(err, "should have seen an error getting the address from Rupert's cart")
	req.Contains(err.Error(), "failed to retrieve delivery address for cart with ID "+cart.Id, "should have seen the expected delivery address retrieval error but got: %v", err)
	req.Nil(getResponse, "should not have got a response getting Rupert's cart")

}

// TestDeliveryAddressSetResponseFailure examines what happens if the Firestore Get fails when trying to gather the
// whole shopping cart from the firestore after setting the delivery address into it.
func TestDeliveryAddressSetResponseFailure(t *testing.T) {

	// Do the basic foundation stuff that most of this package's tests require
	req, ctx, service, cart := commonTestSetup(t)

	// Modify the cart service to return an error only after the address has been set into the cart successfully
	service.drProxy = &fsproxies.UTDocRefProxy{Err: mockError, AllowCount: 1}

	// Add a delivery address to Rupert Potter's cart
	deliveryAddress := buildMockDeliveryAddress()
	setResponse, err := service.SetDeliveryAddress(ctx, &pbcart.SetDeliveryAddressRequest{
		CartId:          cart.Id,
		DeliveryAddress: deliveryAddress,
	})

	// Check that everything came back as expected
	req.NotNil(err, "should have seen an error adding an address to Rupert's cart")
	req.Contains(err.Error(), mockError.Error(), "should have seen the configured error adding an address to Rupert's cart")
	req.Nil(setResponse, "should have not obtained a response adding an address to Rupert's cart")
}

// TestDeliveryAddressCorrupt examines what happens if the Firestore Get fails reading the delivery address
// associated with a cart because the delivery address snapshot cannot be unmarshalled.
func TestDeliveryAddressCorrupt(t *testing.T) {

	// Do the basic foundation stuff that most of this package's tests require
	req, ctx, service, cart := commonTestSetup(t)

	// Evaluate the configuration of the just created shopping cart
	req.NotEmpty(cart.Id, "created cart should an ID")

	// Modify the cart service to return an error only after the address has been set into the cart successfully
	service.dsProxy = &fsproxies.UTDocSnapProxy{Err: mockError, AllowCount: 1}

	// Add a delivery address to Rupert Potter's cart
	deliveryAddress := buildMockDeliveryAddress()
	setResponse, err := service.SetDeliveryAddress(ctx, &pbcart.SetDeliveryAddressRequest{
		CartId:          cart.Id,
		DeliveryAddress: deliveryAddress,
	})

	// Attempt to get the cart from the firestore
	req.NotNil(err, "should have seen an error setting an address into a stored cart")
	req.Contains(err.Error(), mockError.Error(), "should have seen the configured error getting a stored cart")
	req.Nil(setResponse, "should not have obtained a response after failing to set an address into a stored cart")
}

// TestAddItemSuccess examines what happens when adding a cart time has no issues writing to the data store.
func TestAddItemSuccess(t *testing.T) {

	// Add the first item to the cart
	req, ctx, service, cart, responseItem1 := addFirstItemToCart(t)

	// Define the second cart item, add that to the cart and confirm the response
	itemTwo := buildMockCartItem(cartItemProductCode2)
	response, err := service.AddItemToShoppingCart(ctx, &pbcart.AddItemToShoppingCartRequest{CartId: cart.Id, Item: itemTwo})
	req.Nil(err, "should not have seen an error adding %s to Rupert's cart: %v", cartItemProductCode2, err)
	req.NotNil(response, "should have obtained a response after adding %s Rupert's cart", cartItemProductCode2)
	req.NotNil(response.GetCart(), "response should have contained a cart after adding %s", cartItemProductCode2)
	responseItems := response.GetCart().GetCartItems()
	req.Equal(2, len(responseItems), "returned cart should have contained two items")

	// The order of the two added items in the response is non-deterministic :-( a=so fix that
	var responseItem2 *pbcart.CartItem
	if responseItems[0].ProductCode == cartItemProductCode1 {
		responseItem1 = responseItems[0]
		responseItem2 = responseItems[1]
	} else {
		responseItem1 = responseItems[1]
		responseItem2 = responseItems[0]
	}

	// Verify both items in the second response
	validateCartItem1(req, responseItem1, "second")
	req.Equal(cartItemProductCode2, responseItem2.ProductCode, "returned cart should have contained %s", cartItemProductCode2)
	req.Equal(cartItemQuantity2, responseItem2.Quantity, "returned cart had the wrong quantity of %s", cartItemProductCode2)
	req.NotNil(responseItem2.UnitPrice, "returned cart had no price for %s", cartItemProductCode2)
	req.Equal(cartItemPriceCurrency, responseItem2.UnitPrice.CurrencyCode, "returned cart the wrong currency code for %s", cartItemProductCode2)
	req.Equal(cartItemPriceUnits2, responseItem2.UnitPrice.Units, "returned cart had the wrong unit dollars for %s", cartItemProductCode2)
	req.Equal(cartItemPriceNanos2, responseItem2.UnitPrice.Nanos, "returned cart had the wrong unit cents for %s", cartItemProductCode2)
}

// addFirstItemToCart does the common work for a series of tests around adding and removing cart items, i.e.
// it adds the first item to the cart and confirms that it has been stored successfully.
func addFirstItemToCart(t *testing.T) (*require.Assertions, context.Context, *CartService, *pbcart.ShoppingCart, *pbcart.CartItem) {

	// Do the basic foundation stuff that most of this package's tests require
	req, ctx, service, cart := commonTestSetup(t)

	// Define two cart items to add to the - we want to test the looping aspects of returning a cart
	item := buildMockCartItem(cartItemProductCode1)

	// Add both items to the cart in turn
	response, err := service.AddItemToShoppingCart(ctx, &pbcart.AddItemToShoppingCartRequest{CartId: cart.Id, Item: item})
	req.Nil(err, "should not have seen an error adding %s to Rupert's cart: %v", cartItemProductCode1, err)
	req.NotNil(response, "should have obtained a response after adding %s Rupert's cart", cartItemProductCode1)
	responseCart := response.GetCart()
	req.NotNil(responseCart, "response should have contained a cart after adding %s", cartItemProductCode1)
	responseItems := responseCart.GetCartItems()
	req.Equal(1, len(responseItems), "returned cart should have contained one item")
	responseItem1 := responseItems[0]
	validateCartItem1(req, responseItem1, "first")

	// Pass everything back that the caller needs to performa additional testing
	return req, ctx, service, responseCart, responseItem1
}

// validateCartItem1 is used by TestAddItemSuccess to check the first addition of cartItemProductCode1 is present and
// correct after both the first and second item are added to a cart.
func validateCartItem1(req *require.Assertions, item *pbcart.CartItem, validationContext string) {
	req.Equal(cartItemProductCode1, item.ProductCode, "returned cart after %s add should have contained %s", validationContext, cartItemProductCode1)
	req.Equal(cartItemQuantity1, item.Quantity, "returned cart after %s add had the wrong quantity of %s", validationContext, cartItemProductCode1)
	req.NotNil(item.UnitPrice, "returned cart after %s add had no price for %s", validationContext, cartItemProductCode1)
	req.Equal(cartItemPriceCurrency, item.UnitPrice.CurrencyCode, "returned cart after %s add had the wrong currency code for %s", validationContext, cartItemProductCode1)
	req.Equal(cartItemPriceUnits1, item.UnitPrice.Units, "returned cart after %s add had the wrong unit dollars for %s", validationContext, cartItemProductCode1)
	req.Equal(cartItemPriceNanos1, item.UnitPrice.Nanos, "returned cart after %s add had the wrong unit cents for %s", validationContext, cartItemProductCode1)
}

// TestAddItemFailure examines what happens when adding a cart time has issues writing to the data store.
func TestAddItemFailure(t *testing.T) {

	// Do the basic foundation stuff that most of this package's tests require
	req, ctx, service, cart := commonTestSetup(t)

	// Modify the cart service to return an error on any and every DocumentRef operation
	service.drProxy = &fsproxies.UTDocRefProxy{Err: mockError, AllowCount: 0}

	// Define one cart items to add to the cart
	item := buildMockCartItem(cartItemProductCode1)

	// Attempt to Add the item to the cart and confirm that it fails as expected
	response, err := service.AddItemToShoppingCart(ctx, &pbcart.AddItemToShoppingCartRequest{CartId: cart.Id, Item: item})
	req.NotNil(err, "should have seen an error adding an item to Rupert's cart")
	req.Contains(err.Error(), mockError.Error(), "should have seen the configured error adding an item to Rupert's cart")
	req.Nil(response, "should have not obtained a response adding an item to Rupert's cart")
}

// TestAddItemResponseFailure examines what happens if a Firestore Get fails when trying to gather the
// whole shopping cart from the Firestore after adding an item to it.
func TestAddItemResponseFailure(t *testing.T) {

	// Do the basic foundation stuff that most of this package's tests require
	req, ctx, service, cart := commonTestSetup(t)

	// Modify the cart service to return an error only after the item has been set into the cart successfully
	service.drProxy = &fsproxies.UTDocRefProxy{Err: mockError, AllowCount: 1}

	// Define one cart items to add to the cart
	item := buildMockCartItem(cartItemProductCode1)

	// Attempt to Add the item to the cart and confirm that it fails as expected
	response, err := service.AddItemToShoppingCart(ctx, &pbcart.AddItemToShoppingCartRequest{CartId: cart.Id, Item: item})
	req.NotNil(err, "should have seen an error adding an item to Rupert's cart")
	req.Contains(err.Error(), mockError.Error(), "should have seen the configured error adding an item to Rupert's cart")
	req.Nil(response, "should have not obtained a response adding an item to Rupert's cart")
}

// TestRemoveItemSuccess examines what happens when removing a cart time has no issues with the data store.
func TestRemoveItemSuccess(t *testing.T) {

	// Add the first item to the cart
	req, ctx, service, cart, responseItem1 := addFirstItemToCart(t)

	// Now remove that item
	removeResp, err := service.RemoveItemFromShoppingCart(ctx, &pbcart.RemoveItemFromShoppingCartRequest{CartId: cart.Id, ItemId: responseItem1.Id})
	req.Nil(err, "should not have seen an error removing %s to Rupert's cart: %v", cartItemProductCode1, err)
	req.NotNil(removeResp, "should have obtained a response after removing %s Rupert's cart", cartItemProductCode1)
	req.NotNil(removeResp.GetCart(), "response should have contained a cart after removing %s", cartItemProductCode1)
	responseItems := removeResp.GetCart().GetCartItems()
	req.Equal(0, len(responseItems), "returned cart should have contained zero items")
}

// TestRemoveItemFailure examines what happens when removing a cart time has some issues with the data store.
func TestRemoveItemFailure(t *testing.T) {

	// Add the first item to the cart
	req, ctx, service, cart, responseItem1 := addFirstItemToCart(t)

	// Modify the cart service to return an error on any and every DocumentRef operation
	service.drProxy = &fsproxies.UTDocRefProxy{Err: mockError, AllowCount: 0}

	// Now remove that item
	removeResp, err := service.RemoveItemFromShoppingCart(ctx, &pbcart.RemoveItemFromShoppingCartRequest{CartId: cart.Id, ItemId: responseItem1.Id})
	req.NotNil(err, "should have seen an error removing an item to Rupert's cart")
	req.Contains(err.Error(), mockError.Error(), "should have seen the configured error removing an item to Rupert's cart")
	req.Nil(removeResp, "should have not obtained a response removing an item to Rupert's cart")
}

// TestRemoveItemResponseFailure examines what happens when removing a cart time is successful but retrieving the
// resulting cart afterwards fails.
func TestRemoveItemResponseFailure(t *testing.T) {

	// Add the first item to the cart
	req, ctx, service, cart, responseItem1 := addFirstItemToCart(t)

	// Modify the cart service to return an error after the item deletion has been performed successfully
	service.drProxy = &fsproxies.UTDocRefProxy{Err: mockError, AllowCount: 1}

	// Now remove that item
	removeResp, err := service.RemoveItemFromShoppingCart(ctx, &pbcart.RemoveItemFromShoppingCartRequest{CartId: cart.Id, ItemId: responseItem1.Id})
	req.NotNil(err, "should have seen an error removing an item to Rupert's cart")
	req.Contains(err.Error(), mockError.Error(), "should have seen the configured error removing an item to Rupert's cart")
	req.Nil(removeResp, "should have not obtained a response removing an item to Rupert's cart")
}

// TestCheckoutSuccess examines both the successful checking out of a cart and the special case of trying
// to check the same cart out a second time (which should fail).
func TestCheckoutSuccess(t *testing.T) {

	// Register a cart with a single item in it
	req, ctx, service, cart, _ := addFirstItemToCart(t)

	// Check the cart out and confirm all went well
	response, err := service.CheckoutShoppingCart(ctx, &pbcart.CheckoutShoppingCartRequest{CartId: cart.Id})
	req.Nil(err, "failed to check cart out: %v", err)
	req.NotNil(response, "should have obtained a response after checking cart out")
	responseCart := response.GetCart()
	req.NotNil(response.GetCart(), "response should have contained the final cart state")
	req.Equal(cart.Id, responseCart.Id, "response cart ID did not match the cart we checked out")
	req.Equal(pbcart.ShoppingCartStatus_SCS_CHECKED_OUT, responseCart.Status, "response cart does not have the checked out status")

	// Try to check the cart out a second time
	response, err = service.CheckoutShoppingCart(ctx, &pbcart.CheckoutShoppingCartRequest{CartId: cart.Id})
	req.NotNil(err, "should have seen an error checking out Rupert's cart a second time")
	req.Contains(err.Error(), "cannot change status of cart that is not open: cart ID="+cart.Id, "should have seen cannot check out a closed cart error")
	req.Nil(response, "should have not obtained a response after failing to change cart status")
}

// TestCheckoutCartMissing examines what happens if the caller tries to check out a cart that does not exist.
func TestCheckoutCartMissing(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Initialize our target cart service
	service, err := NewCartService()
	req.Nil(err, "failed to obtain cart service: %v", err)

	// Form the ID of a cart that we know cannot exist
	cartId := uuid.New().String()

	// Check the cart out and confirm all went well
	ctx := context.Background()
	response, err := service.CheckoutShoppingCart(ctx, &pbcart.CheckoutShoppingCartRequest{CartId: cartId})
	req.NotNil(err, "should have seen an error checking out a non-existent cart")
	req.Contains(err.Error(), "failed to retrieve cart snapshot with ID "+cartId, "should have seen failed to load cart error")
	req.Nil(response, "should have not obtained a response after failing to change non-existent cart status")
}

// TestCheckoutOriginalCartUnmarshalFailure looks at what happens if the target cart exists but cannot be loaded to
// see what its current state is.
func TestCheckoutOriginalCartUnmarshalFailure(t *testing.T) {

	// Register a cart with a single item in it
	req, ctx, service, cart, _ := addFirstItemToCart(t)

	// Modify the cart service to return an error the first time the code attempts to unmarshal a document snapshot
	service.dsProxy = &fsproxies.UTDocSnapProxy{Err: mockError, AllowCount: 0}

	// Check the cart out and confirm all went well
	response, err := service.CheckoutShoppingCart(ctx, &pbcart.CheckoutShoppingCartRequest{CartId: cart.Id})
	req.NotNil(err, "should have seen an error checking out a corrupt cart")
	req.Contains(err.Error(), "failed to unmarshal cart snapshot with ID "+cart.Id, "should have seen an unmarshalling error")
	req.Nil(response, "should have not obtained a response after failing to unmarshall the original cart")
}

// TestCheckoutFailure looks at what happens if Firestore is unable to set the checked out status on the target cart.
func TestCheckoutFailure(t *testing.T) {

	// Register a cart with a single item in it
	req, ctx, service, cart, _ := addFirstItemToCart(t)

	// Modify the cart service to return an error the second time we try to use a document reference proxy
	service.drProxy = &fsproxies.UTDocRefProxy{Err: mockError, AllowCount: 1}

	// Check the cart out and confirm all went well
	response, err := service.CheckoutShoppingCart(ctx, &pbcart.CheckoutShoppingCartRequest{CartId: cart.Id})
	req.NotNil(err, "should have seen an error checking out a corrupt cart")
	req.Contains(err.Error(), "failed putting updated cart status to datastore with ID "+cart.Id, "should have seen an document set error")
	req.Nil(response, "should have not obtained a response after failing to set the checked out status")
}

// commonTestSetup helps us to be a little DRY (Don't Repeat Yourself) in this file, doing the steps that more
// than have the unit test functions in here need to do before going on to anything else.
func commonTestSetup(t *testing.T) (*require.Assertions, context.Context, *CartService, *pbcart.ShoppingCart) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Initialize our target cart service and establish a cart
	ctx := context.Background()
	service, cart := storeMockCart(ctx, req)

	// return everything the caller needs to perform their tests
	return req, ctx, service, cart
}

// TestAbandonSuccess examines both the successful abandoning of a cart and the special case of trying
// to abandon the same cart out a second time (which should fail).
func TestAbandonSuccess(t *testing.T) {

	// Register a cart with a single item in it
	req, ctx, service, cart, _ := addFirstItemToCart(t)

	// Check the cart out and confirm all went well
	response, err := service.AbandonShoppingCart(ctx, &pbcart.AbandonShoppingCartRequest{CartId: cart.Id})
	req.Nil(err, "failed to abandon cart: %v", err)
	req.NotNil(response, "should have obtained a response after abandoning the cart")
	responseCart := response.GetCart()
	req.NotNil(response.GetCart(), "response should have contained the final cart state")
	req.Equal(cart.Id, responseCart.Id, "response cart ID did not match the cart we abandoned")
	req.Equal(pbcart.ShoppingCartStatus_SCS_ABANDONED_BY_USER, responseCart.Status, "response cart does not have the abandoned status")

	// Try to check the cart out a second time
	response, err = service.AbandonShoppingCart(ctx, &pbcart.AbandonShoppingCartRequest{CartId: cart.Id})
	req.NotNil(err, "should have seen an error abandoning out Rupert's cart a second time")
	req.Contains(err.Error(), "cannot change status of cart that is not open: cart ID="+cart.Id, "should have seen cannot check out a closed cart error")
	req.Nil(response, "should have not obtained a response after failing to change cart status")
}

// storeMockCart establishes and caches a cart service the first time it is called and and uses that to create
// and empty shopping cart, returning both to the calling unit test. The supplied require.Assertions will be
// used to report any issues occur with either step, aborting the unit test before it really gets started.
func storeMockCart(ctx context.Context, req *require.Assertions) (*CartService, *pbcart.ShoppingCart) {

	// Initialize our target cart service
	service, err := NewCartService()
	req.Nil(err, "failed to obtain cart service: %v", err)

	// Establish a cart for our mock shopper
	createResp, err := service.CreateShoppingCart(ctx, &pbcart.CreateShoppingCartRequest{Shopper: buildMockShopper()})
	req.Nil(err, "should not have seen an error creating a new cart: %v", err)
	req.NotNil(createResp, "should have obtained a response after creating a new cart")
	req.NotNil(createResp.GetCart(), "response should have contained a new cart")

	// All done here - let the caller decide what to do from here on
	return service, createResp.GetCart()
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
