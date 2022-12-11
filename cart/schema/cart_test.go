package schema

import (
	pbcart "github.com/mikebway/poc-gcp-ecomm/pb/cart"
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

// TestFirestorePaths evaluates each of the Firestore path methods in turn.
func TestFirestorePaths(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Obtain a fully populated shopping cart
	cart := buildMockCart()

	// Get the firestore path for the shopping cart and confirm that it looks as we expect
	cartPath := cart.StoreRefPath()
	req.Equal("carts/d1cecab3-5bc0-43d4-aef1-99ad69794313", cartPath, "cart path content does not match expected value")

	// Get the firestore path for the delivery address of the shopping cart and confirm that it looks as we expect
	deliveryAddressPath := cart.DeliveryAddressPath()
	req.Equal("carts/d1cecab3-5bc0-43d4-aef1-99ad69794313/addresses/delivery", deliveryAddressPath, "delivery address path content does not match expected value")

	// Get the Firestore collect path that will contain all the cart item documents
	itemCollectionPath := cart.ItemCollectionPath()
	req.Equal("carts/d1cecab3-5bc0-43d4-aef1-99ad69794313/items", itemCollectionPath, "delivery address path content does not match expected value")

	// Get the firestore path for a cart item and confirm that it looks as we expect
	cartItem := buildMockCartItems()[0]
	cartItemPath := cartItem.StoreRefPath()
	req.Equal("carts/d1cecab3-5bc0-43d4-aef1-99ad69794313/items/54f34cb9-fea6-4786-a475-cebd95d93742", cartItemPath, "cart item path content does not match expected value")
}

// TestFullConversion examines what happens when a fully populated cart is converted to its
// protocol buffer form and back again.
func TestFullConversion(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Obtain a fully populated shopping cart
	srcCart := buildMockCart()

	// Convert the cart to protocol buffer form
	pbCart := srcCart.AsPBShoppingCart()

	// Validate that everything made it across OK
	req.NotNil(pbCart, "should have received a PB cart by reference")
	req.Equal(shoppingCartId, pbCart.Id, "cart IDs did not match")
	req.Equal(shoppingCartCreationTime, pbCart.CreationTime.AsTime(), "cart creation times did not match")
	req.Equal(shoppingCartClosedTime, pbCart.ClosedTime.AsTime(), "cart closure times did not match")
	req.Equal(int32(CsCheckedOut), int32(pbCart.Status), "cart status did not match")
	req.NotNil(pbCart.DeliveryAddress, "should have found a delivery address")
	req.Equal(addrPostalCode, pbCart.DeliveryAddress.PostalCode, "delivery address postal code did not match")
	req.Equal(len(srcCart.CartItems), len(pbCart.CartItems), "item count did not match")
	for i, item := range srcCart.CartItems {
		req.Equal(item.Id, pbCart.CartItems[i].Id, "cart item ID did not match: %d", i)
		req.Equal(item.CartId, pbCart.CartItems[i].CartId, "cart item cart ID did not match: %d", i)
		req.NotNil(pbCart.CartItems[i].UnitPrice, "cart item price was not set: %d", i)
		req.Equal(item.UnitPrice.AsPBMoney().String(), pbCart.CartItems[i].UnitPrice.String(), "cart item price did not match: %d", i)
	}

	// Convert the protocol buffer cart back to its local form
	finalCart := ShoppingCartFromPB(pbCart)
	req.NotNil(finalCart, "should have received a local cart by reference")
	req.Equal(srcCart, finalCart, "twice transformed cart did not match")

	// Prove to any skeptic reading this code that the final cart equality check was a full depth comparison.
	// The author was a skeptic so needed to do this for their own peace of mind!
	finalCart.CartItems[1].UnitPrice.Units = 0
	req.NotEqual(srcCart, finalCart, "modified final cart should not have matched")
}

// TestMinimumConversion examines what happens when a minimally populated cart is converted to its
// protocol buffer form and back again.
//
// TestFullConversion achieved 100% line coverage but did not test significant alternative paths
// that would be followed if reference pointers were nil or source values were zero that should
// translate to nil references.
func TestMinimumConversion(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Obtain a minimally populated shopping cart
	srcCart := &ShoppingCart{
		Id: shoppingCartId,
	}

	// Convert the cart to protocol buffer form
	pbCart := srcCart.AsPBShoppingCart()

	// Validate that everything (i.e. almost nothing) made it across OK
	req.NotNil(pbCart, "should have received a PB cart by reference")
	req.Equal(shoppingCartId, pbCart.Id, "cart IDs did not match")
	req.Nil(pbCart.CreationTime, "cart creation time was not nil")
	req.Nil(pbCart.ClosedTime, "cart closure time was not nil")
	req.Equal(pbcart.ShoppingCartStatus_SCS_UNSPECIFIED, pbCart.Status, "cart status should not be specified")
	req.Nil(pbCart.DeliveryAddress, "delivery address was not nil")
	req.Equal(0, len(pbCart.CartItems), "item count should be zero")

	// Convert the protocol buffer cart back to its local form.
	finalCart := ShoppingCartFromPB(pbCart)
	req.NotNil(finalCart, "should have received a local cart by reference")
	req.Equal(shoppingCartId, finalCart.Id, "final cart IDs did not match")
	req.True(finalCart.CreationTime.IsZero(), "final cart creation time was not zero")
	req.True(finalCart.ClosedTime.IsZero(), "final cart closure time was not zero")
	req.Equal(CsUnspecified, finalCart.Status, "final cart status should not be specified")
	req.Nil(finalCart.DeliveryAddress, "final delivery address was not nil")
	req.Equal(0, len(finalCart.CartItems), "final item count should be zero")
}

// buildMockCart returns a ShoppingCart structure populated with a shopper that can be used to
// test storing new shopping carts in our tests.
func buildMockCart() *ShoppingCart {
	return &ShoppingCart{
		Id:              shoppingCartId,
		CreationTime:    shoppingCartCreationTime,
		ClosedTime:      shoppingCartClosedTime,
		Status:          CsCheckedOut,
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
func buildMockCartItems() []*ShoppingCartItem {
	return []*ShoppingCartItem{
		&ShoppingCartItem{
			Id:          itemId1,
			CartId:      shoppingCartId,
			ProductCode: itemProdCode1,
			Quantity:    itemQuantity1,
			UnitPrice:   &itemPrice1,
		},
		&ShoppingCartItem{
			Id:          itemId2,
			CartId:      shoppingCartId,
			ProductCode: itemProdCode2,
			Quantity:    itemQuantity2,
			UnitPrice:   &itemPrice2,
		},
	}
}
