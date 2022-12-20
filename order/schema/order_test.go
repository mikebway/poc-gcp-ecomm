package schema

import (
	"github.com/mikebway/poc-gcp-ecomm/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	// A timestamp string we can use to derive known time values
	timeString = "2022-10-29T16:23:19.123456789-06:00"

	// A UUID string value that we can use as an order ID in our tests
	orderId = "d1cecab3-5bc0-43d4-aef1-99ad69794313"

	// Define the person fields that we will use multiple times to define a person
	personId          = "10615145-2010-4c5f-8347-2bb556232c31"
	personFamilyName  = "Grint"
	personGivenName   = "Rupert"
	personMiddleName  = "Alexander Lloyd"
	personDisplayName = "Rupert"

	// Define the postal address fields for the home address of our mock person
	addrLine1      = "55 Yonder St"
	addrLine2      = "Flat B"
	addrLocality   = "Ottery St Catchpole"
	addrPostalCode = "EX11 1HF"
	addrRegionCode = "GB"

	// Define shopping cart item field values for two shopping cart items
	// As of October 2022, the product codes are for an Apple Mac Studio and Studio Display :-)
	itemId1                     = "54f34cb9-fea6-4786-a475-cebd95d93742"
	itemId2                     = "b719efe9-4189-453c-96b0-7e229520d316"
	itemProdCode1               = "B09V3G22BH"
	itemProdCode2               = "B09V3GZD32"
	itemQuantity1         int32 = 1
	itemQuantity2         int32 = 2
	itemPriceCurrencyCode       = "USD"
	itemPriceUnits1       int64 = 1899
	itemPriceNanos1       int32 = 550_000_000
	itemPriceUnits2       int64 = 1598
	itemPriceNanos2       int32 = 960_000_000
)

var (
	// Test values cannot be declared as constants
	submissionTime time.Time
	itemPrice1     = types.Money{CurrencyCode: itemPriceCurrencyCode, Units: itemPriceUnits1, Nanos: itemPriceNanos1}
	itemPrice2     = types.Money{CurrencyCode: itemPriceCurrencyCode, Units: itemPriceUnits2, Nanos: itemPriceNanos2}
)

// init is used to initialize the order item "constant" values that cannot be declared
// as literal constants.
func init() {

	// Shopping cart values
	t, _ := types.TimestampFromRFC3339Nano(timeString)
	submissionTime = t.GetTime()
}

// TestStoreRefPath checks out how well Order.StoreRefPath does its job.
func TestStoreRefPath(t *testing.T) {

	// Create an order that we can ask the store path of
	order := buildMockOrder()

	// Get the Firestore path for the order and confirm that it is what we would expect
	orderPath := order.StoreRefPath()
	require.Equal(t, "orders/"+orderId, orderPath, "order store reference path incorrect")
}

// TestAsPBOrder examines the behavior of Order.AsPBOrder for fully populated order.
func TestAsPBOrder(t *testing.T) {

	// Create a fully populated order that we can convert
	order := buildMockOrder()

	// Convert that to the Protocol Buffer form
	pbOrder := order.AsPBOrder()

	// Validate the Protocol Buffer order content
	req := require.New(t)
	req.Equal(orderId, pbOrder.Id, "order ID does not match")
	req.NotNil(pbOrder.SubmissionTime, "submission time missing")
	subTimestamp, _ := types.TimestampFromPBTimestamp(pbOrder.SubmissionTime)
	req.Equal(submissionTime, subTimestamp.GetTime(), "submission time does not match")
	req.NotNil(pbOrder.OrderedBy, "ordered by person missing")
	req.Equal(personId, pbOrder.OrderedBy.Id, "ordered by person ID does not match")
	req.Equal(personFamilyName, pbOrder.OrderedBy.FamilyName, "ordered by person family name does not match")
	req.Equal(personGivenName, pbOrder.OrderedBy.GivenName, "ordered by person given name does not match")
	req.Equal(personMiddleName, pbOrder.OrderedBy.MiddleName, "ordered by person middle does not match")
	req.Equal(personDisplayName, pbOrder.OrderedBy.DisplayName, "ordered by person display name does not match")

	req.NotNil(pbOrder.DeliveryAddress, "delivery address missing")
	req.Equal(2, len(pbOrder.DeliveryAddress.AddressLines), "delivery address line count does not match")
	req.Equal(addrLine1, pbOrder.DeliveryAddress.AddressLines[0], "delivery address line 1 does not match")
	req.Equal(addrLine2, pbOrder.DeliveryAddress.AddressLines[1], "delivery address line 2 does not match")
	req.Equal(addrLocality, pbOrder.DeliveryAddress.Locality, "delivery address locality does not match")
	req.Equal(addrPostalCode, pbOrder.DeliveryAddress.PostalCode, "delivery address postal code does not match")
	req.Equal(addrRegionCode, pbOrder.DeliveryAddress.RegionCode, "delivery address region code does not match")

	req.Equal(2, len(pbOrder.OrderItems), "order item count does not match")

	req.NotNil(pbOrder.OrderItems[0], "order item 1 missing")
	req.Equal(itemId1, pbOrder.OrderItems[0].Id, "order item 1 ID does not match")
	req.Equal(itemProdCode1, pbOrder.OrderItems[0].ProductCode, "order item 1 product code does not match")
	req.Equal(itemQuantity1, pbOrder.OrderItems[0].Quantity, "order item 1 quantity does not match")
	req.NotNil(pbOrder.OrderItems[0].UnitPrice, "order item 1 unit price missing")
	req.Equal(itemPriceCurrencyCode, pbOrder.OrderItems[0].UnitPrice.CurrencyCode, "order item 1 price currency does not match")
	req.Equal(itemPriceUnits1, pbOrder.OrderItems[0].UnitPrice.Units, "order item 1 price units does not match")
	req.Equal(itemPriceNanos1, pbOrder.OrderItems[0].UnitPrice.Nanos, "order item 1 price nanos does not match")

	req.NotNil(pbOrder.OrderItems[1], "order item 2 missing")
	req.Equal(itemId2, pbOrder.OrderItems[1].Id, "order item 2 ID does not match")
	req.Equal(itemProdCode2, pbOrder.OrderItems[1].ProductCode, "order item 2 product code does not match")
	req.Equal(itemQuantity2, pbOrder.OrderItems[1].Quantity, "order item 2 quantity does not match")
	req.NotNil(pbOrder.OrderItems[1].UnitPrice, "order item 2 unit price missing")
	req.Equal(itemPriceCurrencyCode, pbOrder.OrderItems[1].UnitPrice.CurrencyCode, "order item 2 price currency does not match")
	req.Equal(itemPriceUnits2, pbOrder.OrderItems[1].UnitPrice.Units, "order item 2 price units does not match")
	req.Equal(itemPriceNanos2, pbOrder.OrderItems[1].UnitPrice.Nanos, "order item 2 price nanos does not match")
}

// TestEmptyOrderAsPBOrder examines the behavior of Order.AsPBOrder for a completely unpopulated order. This is
// a corner case that will never occur in the wild but it ensures that all the individual condition checks that
// might fire are exercised.
func TestEmptyOrderAsPBOrder(t *testing.T) {

	// Create an unpopulated order that we can convert
	order := &Order{}

	// Convert that to the Protocol Buffer form
	pbOrder := order.AsPBOrder()

	// Validate the Protocol Buffer order content
	req := require.New(t)
	req.Empty(pbOrder.Id, "order ID is defined and should not be")
	req.Nil(pbOrder.SubmissionTime, "submission time is defined and should not be")
	req.Nil(pbOrder.OrderedBy, "ordered by person is defined and should not be")
	req.Nil(pbOrder.DeliveryAddress, "delivery address is defined and should not be")
	req.Equal(0, len(pbOrder.OrderItems), "order item count is non-zero is defined and should not be")
}

// buildMockOrder returns a Order structure populated with a person that can be used to
// test storing new shopping carts in our tests.
func buildMockOrder() *Order {
	return &Order{
		Id:              orderId,
		SubmissionTime:  submissionTime,
		OrderedBy:       buildMockPerson(),
		DeliveryAddress: buildMockDeliveryAddress(),
		OrderItems:      buildMockOrderItems(),
	}
}

// buildMockPerson returns a types.Person structure populated with the person constant attributes
// defined at the head of this file to be used to create new shopping carts in our tests.
func buildMockPerson() *types.Person {
	return &types.Person{
		Id:          personId,
		FamilyName:  personFamilyName,
		GivenName:   personGivenName,
		MiddleName:  personMiddleName,
		DisplayName: personDisplayName,
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

// buildMockOrderItems returns an array containing two shopping cart items populated  with the constant
// attributes defined at the head of this file to be used to create new shopping carts in our tests.
func buildMockOrderItems() []*OrderItem {
	return []*OrderItem{
		&OrderItem{
			Id:          itemId1,
			ProductCode: itemProdCode1,
			Quantity:    itemQuantity1,
			UnitPrice:   &itemPrice1,
		},
		&OrderItem{
			Id:          itemId2,
			ProductCode: itemProdCode2,
			Quantity:    itemQuantity2,
			UnitPrice:   &itemPrice2,
		},
	}
}
