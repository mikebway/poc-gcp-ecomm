package carttrigger

import (
	"context"
	"github.com/mikebway/poc-gcp-ecomm/cart/schema"
	"github.com/mikebway/poc-gcp-ecomm/types"
	"github.com/mikebway/poc-gcp-ecomm/util"
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
)

var (
	// Shopping cart value types that cannot be declared as constants
	shoppingCartCreationTime time.Time
	shoppingCartClosedTime   time.Time

	// Firestore value times
	firestoreValueCreateTime time.Time
	firestoreValueUpdateTime time.Time
)

// init is used to initialize the shopping cart item "constant" values that cannot be declared
// as literal constants.
func init() {

	// Shopping cart values
	t, _ := types.TimestampFromRFC3339Nano(earlyTimeString)
	shoppingCartCreationTime = t.GetTime()
	t, _ = types.TimestampFromRFC3339Nano(lateTimeString)
	shoppingCartClosedTime = t.GetTime()

	// Firestore value times are closely but not exactly related to the shopping cart times
	// We just make them a second apart here but in reality it would be much closer.
	firestoreValueCreateTime = shoppingCartCreationTime.Add(time.Second)
	firestoreValueUpdateTime = shoppingCartClosedTime.Add(time.Second)
}

func TestHandlerHappyPath(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Submit a known FirestoreEvent to the handler while capturing its log output
	event := mockFirestoreEvent()
	ctx := context.Background()
	var err error
	logged := util.CaptureLogging(func() {
		err = UpdateTrigger(ctx, *event)
	})

	// There should have been no errors and some straightforward log output
	req.Nil(err, "no error was expected: %v", err)
	req.Contains(logged, "firestore event", "did not see happy path log message")
	req.Contains(logged, event.Value.Name, "did not see cart document path in log message")
}

// mockFirestoreEvent constructs a FirestoreEvent populated with known values that we can check in out unit tests.
func mockFirestoreEvent() *FirestoreEvent {
	return &FirestoreEvent{
		OldValue: *mockOldValue(),
		Value:    *mockNewValue(),
	}
}

// mockOldValue returns a FirestoreValue populated with an open shopping cart structure as its value and
// with the creation and update times being the same, set a little after the cart creation time.
func mockOldValue() *FirestoreValue {

	// Get the "open" shopping cart
	cart := buildMockOpenCart()

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
func mockNewValue() *FirestoreValue {

	// Get the "open" shopping cart and close it
	cart := buildMockOpenCart()
	cart.Status = schema.CsCheckedOut

	// Build and return our result - update time will be after the creation time
	return &FirestoreValue{
		CreateTime: firestoreValueCreateTime,
		Fields:     *cart,
		Name:       cart.StoreRefPath(),
		UpdateTime: firestoreValueUpdateTime,
	}
}

// buildMockOpenCart returns a types.ShoppingCart structure populated with a shopper that can be used to
// test storing new shopping carts in our tests.
func buildMockOpenCart() *schema.ShoppingCart {
	return &schema.ShoppingCart{
		Id:           shoppingCartId,
		CreationTime: shoppingCartCreationTime,
		Status:       schema.CsOpen,
		Shopper:      buildMockShopper(),
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
