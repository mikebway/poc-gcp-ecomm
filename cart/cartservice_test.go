package main

import (
	"context"
	"errors"
	pbcart "github.com/mikebway/poc-gcp-ecomm/pb/cart"
	pbtypes "github.com/mikebway/poc-gcp-ecomm/pb/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
)

const (
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
)

var (
	// mockError is used as an error result when we wish to have our mock Firestore client pretend to fail
	mockError error
)

// init performs static initialization of our constants that cannot actually be literal constants
func init() {
	mockError = errors.New("this is a mock error")
}

// TestCreateAndGetCart examines whether a new cart can be properly added to the database
func TestCreateAndGetCart(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// The cart creation time will be greater than or equal to the current time now
	startTime := timestamppb.Now()

	// Initialize our target cart service and establish a cart
	ctx := context.Background()
	service, shoppingCart := storeMockCart(ctx, req)

	// Evaluate the configuration of the just created shopping cart
	req.NotEmpty(shoppingCart.Id, "created cart should an ID")
	req.Equal(36, len(shoppingCart.Id), "created cart ID does not look like a UUID")
	req.NotNil(shoppingCart.CreationTime, "cart creation time should have been set for Rupert's new cart")
	req.GreaterOrEqual(shoppingCart.CreationTime.AsTime(), startTime.AsTime(), "cart creation time should be at or after the time that this test started")
	req.Equal(pbcart.ShoppingCartStatus_SCS_OPEN, shoppingCart.Status, "cart status should be 'open'")
	req.Equal(shopperId, shoppingCart.Shopper.Id, "created shopper ID did not match")
	req.Equal(shopperFamilyName, shoppingCart.Shopper.FamilyName, "created shopper family name did not match")
	req.Equal(shopperGivenName, shoppingCart.Shopper.GivenName, "created shopper given name did not match")
	req.Equal(shopperMiddleName, shoppingCart.Shopper.MiddleName, "created shopper middle name(s) did not match")
	req.Equal(shopperDisplayName, shoppingCart.Shopper.DisplayName, "created shopper display name did not match")

	// Should be able to retrieve the new record
	getResp, err := service.GetShoppingCartByID(ctx, &pbcart.GetShoppingCartByIDRequest{CartId: shoppingCart.Id})
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
//func TestCartCreateFailure(t *testing.T) {
//
//	// Avoid having to pass t in to every assertion
//	req := require.New(t)
//
//	// Obtain a cart service with a mock firestore client and tell the client
//	// to return an error the first time a Put operation is invoked
//	ctx := context.Background()
//	service, cart := storeMockCart(ctx, req)
//
//	// Attempt to create a new cart
//	ctx := context.Background()
//	createResp, err := service.CreateShoppingCart(ctx, &pbcart.CreateShoppingCartRequest{Shopper: buildMockShopper()})
//	req.NotNil(err, "should have seen an error creating a new cart")
//	req.Contains(err.Error(), mockError.Error(), "should have seen the configured error creating a new cart")
//	req.Nil(createResp, "should not have obtained a response after failing to create a new cart")
//}

// TestCartNotFound looks at what happens when trying to retrieve a shopping cart that does not exist.
//func TestCartNotFound(t *testing.T) {
//
//	// Avoid having to pass t in to every assertion
//	req := require.New(t)
//
//	// Initialize our target cart service
//	service, err := storeMockCart()
//	req.Nil(err, "failed to obtain service with mocked firestore client")
//
//	// Ask for the data for a bogus person ID
//	ctx := context.Background()
//	const bogusId = "not-even-a-uuid"
//	getResp, err := service.GetShoppingCartByID(ctx, &pbcart.GetShoppingCartByIDRequest{CartId: bogusId})
//	req.NotNil(err, "should have seen an error asking for a cart with a bogus ID")
//	req.Contains(err.Error(), "firestore: no such entity", "did not get the expected error message")
//	req.Contains(err.Error(), bogusId, "error message did not include the bogus ID")
//	req.Nil(getResp, "should not have obtained a response asking for a cart with a bogus ID")
//}

// TestSetDeliveryAddress confirms that an address can be added to the cart, that the amended cart gets returned
// with the response, and that the delivery address is present in a follow-up get of the cart.
func TestSetDeliveryAddress(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Initialize our target cart service and establish a cart
	ctx := context.Background()
	service, cart := storeMockCart(ctx, req)

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
//func TestDeliveryAddressSetFailure(t *testing.T) {
//	// Avoid having to pass t in to every assertion
//	req := require.New(t)
//
//	// Initialize our target cart service and establish a cart
//	ctx := context.Background()
//	service, cart := storeMockCart(ctx, req)
//
//	// Configure our mock firestore client to return an error when the cart service attempts to
//	// Put the delivery address to the firestore
//	unitTestNewCartServiceError = mockError
//	defer func() { unitTestNewCartServiceError = nil }()
//
//	// Add a delivery address to Harry Potter's cart
//	deliveryAddress := buildMockDeliveryAddress()
//	setResponse, err := service.SetDeliveryAddress(ctx, &pbcart.SetDeliveryAddressRequest{
//		CartId:          cart.Id,
//		DeliveryAddress: deliveryAddress,
//	})
//
//	// Check that everything came back as expected
//	req.NotNil(err, "should have seen an error adding an address to Rupert's cart")
//	req.Contains(err.Error(), mockError.Error(), "should have seen the configured error adding an address to Rupert's cart")
//	req.Nil(setResponse, "should have not obtained a response adding an address to Rupert's cart")
//}

// TestDeliveryAddressSetResponseFailure examines what happens if the Firestore Get fails when trying to gather the
// whole shopping cart from the firestore after setting the delivery address into it.
//func TestDeliveryAddressSetResponseFailure(t *testing.T) {
//	// Avoid having to pass t in to every assertion
//	req := require.New(t)
//
//	// Initialize our target cart service and establish a cart
//	ctx := context.Background()
//	service, cart := storeMockCart(ctx, req)
//
//	// Configure our mock firestore client to return an error when the cart service attempts to
//	// Get the shopping cart out of the firestore after the address has been successfully Put.
//	service.dsClient.(*mockFirestoreClient).GetError = mockError
//
//	// Add a delivery address to Harry Potter's cart
//	deliveryAddress := buildMockDeliveryAddress()
//	setResponse, err := service.SetDeliveryAddress(ctx, &pbcart.SetDeliveryAddressRequest{
//		CartId:          cart.Id,
//		DeliveryAddress: deliveryAddress,
//	})
//
//	// Check that everything came back as expected
//	req.NotNil(err, "should have seen an error adding an address to Rupert's cart")
//	req.Contains(err.Error(), mockError.Error(), "should have seen the configured error adding an address to Rupert's cart")
//	req.Nil(setResponse, "should have not obtained a response adding an address to Rupert's cart")
//}

// TestDeliveryAddressGetFailure examines what happens if the Firestore Get fails reading the delivery address
// associated with a cart
//func TestDeliveryAddressGetFailure(t *testing.T) {
//
//	// Avoid having to pass t in to every assertion
//	req := require.New(t)
//
//	// Initialize our target cart service and establish a cart
//	ctx := context.Background()
//	service, shoppingCart := storeMockCart(ctx, req)
//
//	// Evaluate the configuration of the just created shopping cart
//	req.NotEmpty(shoppingCart.Id, "created cart should an ID")
//
//	// Instruct the mock firestore client to return an error on the second Get request (for the delivery address)
//	service.dsClient.(*mockFirestoreClient).GetError = mockError
//	service.dsClient.(*mockFirestoreClient).GetErrorAfterNCalls = 1
//
//	// Attempt to get the cart from the firestore
//	getResp, err := service.GetShoppingCartByID(ctx, &pbcart.GetShoppingCartByIDRequest{CartId: shoppingCart.Id})
//	req.NotNil(err, "should have seen an error getting a stored cart")
//	req.Contains(err.Error(), mockError.Error(), "should have seen the configured error getting a stored cart")
//	req.Nil(getResp, "should not have obtained a response after failing to get a stored cart")
//}

// storeMockCart establishes and caches a cart service the first time it is called and and uses that to create
// and empty shopping cart, returning both to the calling unit test. The supplied require.Assertions will be
// used to report any issues occur with either step, aborting the unit test before it really gets started.
func storeMockCart(ctx context.Context, req *require.Assertions) (*CartService, *pbcart.ShoppingCart) {

	// Initialize our target cart service
	service, err := getTestTarget()
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

// testTarget is a cached instance of the CartService used as a target for the majority of the tests in
// this package. It is initialized the first time that getTestTarget is called.
var testTarget *CartService

// getTestTarget returns a singleton instance of the CartService for use in the majority of unit tests.
func getTestTarget() (*CartService, error) {
	if testTarget == nil {
		svc, err := NewCartService()
		if err != nil {
			return nil, err
		}
		testTarget = svc
	}
	return testTarget, nil
}
