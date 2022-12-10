package types

import (
	"github.com/stretchr/testify/require"
	pbmoney "google.golang.org/genproto/googleapis/type/money"
	"testing"
)

const (
	// Define the money fields that we might use multiple times to define a money
	priceCurrency       = "USD"
	priceUnits    int64 = 1_651
	priceNanos    int32 = 940_000_000
)

// TestPBMoneyConversion tests both Money conversion from and to the protocol buffer equivalent.
func TestPBMoneyConversion(t *testing.T) {

	// Build a protocol buffer postal money as our starting point
	source := buildMockMoney()

	// Convert that to our Money type
	price := MoneyFromPB(source)

	// Validate the conversion
	req := require.New(t)
	req.Equal(priceCurrency, price.CurrencyCode, "wrong currency")
	req.Equal(priceUnits, price.Units, "wrong units")
	req.Equal(priceNanos, price.Nanos, "wrong nanos")

	// Convert it back again and compare with the source
	pbReturn := price.AsPBMoney()
	req.Equal(source, pbReturn, "converted back to protobuf did not match original")
}

// TestMilPBMoneyConversion tests what happens if a nil protobuf money is passed to MoneyFromPB
func TestMilPBMoneyConversion(t *testing.T) {

	// Ask for thePostalMoney equivalent of a nil protobuf money
	money := MoneyFromPB(nil)
	require.Nil(t, money, "expected nil in return for nil")
}

// buildMockMoney returns a pbtypes.Money structure populated with the constant
// attributes defined at the head of this file to be used to create new shopping carts in our tests.
func buildMockMoney() *pbmoney.Money {
	return &pbmoney.Money{
		CurrencyCode: priceCurrency,
		Units:        priceUnits,
		Nanos:        priceNanos,
	}
}
