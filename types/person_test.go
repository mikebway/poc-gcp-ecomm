package types

import (
	pbtypes "github.com/mikebway/poc-gcp-ecomm/pb/types"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	// Define the person fields that we might use multiple times to define a person
	personId          = "10615145-2010-4c5f-8347-2bb556232c31"
	personFamilyName  = "Grint"
	personGivenName   = "Rupert"
	personMiddleName  = "Alexander Lloyd"
	personDisplayName = "Rupert"
)

// TestPBPersonConversion tests both Person conversion from and to the protocol buffer equivalent.
func TestPBPersonConversion(t *testing.T) {

	// Build a protocol buffer postal person as our starting point
	source := buildMockPerson()

	// Convert that to our Person type
	person := PersonFromPB(source)

	// Validate the conversion
	req := require.New(t)
	req.Equal(personId, person.Id, "wrong ID")
	req.Equal(personFamilyName, person.FamilyName, "wrong family name")
	req.Equal(personGivenName, person.GivenName, "wrong given name")
	req.Equal(personMiddleName, person.MiddleName, "wrong middle name")
	req.Equal(personDisplayName, person.DisplayName, "wrong display name")

	// Convert it back again and compare with the source
	pbReturn := person.AsPBPerson()
	req.Equal(source, pbReturn, "converted back to protobuf did not match original")
}

// TestMilPBPersonConversion tests what happens if a nil protobuf person is passed to PersonFromPB
func TestMilPBPersonConversion(t *testing.T) {

	// Ask for thePostalPerson equivalent of a nil protobuf person
	person := PersonFromPB(nil)
	require.Nil(t, person, "expected nil in return for nil")
}

// buildMockPerson returns a pbtypes.Person structure populated with the constant
// attributes defined at the head of this file to be used to create new shopping carts in our tests.
func buildMockPerson() *pbtypes.Person {
	return &pbtypes.Person{
		Id:          personId,
		FamilyName:  personFamilyName,
		GivenName:   personGivenName,
		MiddleName:  personMiddleName,
		DisplayName: personDisplayName,
	}
}
