package service

import (
	"errors"
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
