package types

import (
	pbtypes "github.com/mikebway/poc-gcp-ecomm/pb/types"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	// Define the postal address fields for the home address of our mock shopper
	addrRecipient    = "care of Molly Weasley"
	addrOrganization = "The Order of the Phoenix"
	addrMailbox      = "PO 7"
	addrLine1        = "55 Yonder St"
	addrLine2        = "Flat B"
	addrSublocality  = "St Catchpole"
	addrLocality     = "Ottery St Catchpole"
	addrAdminArea    = "Exeter"
	addrPostalCode   = "EX11 1HF"
	addrSortingCode  = "63/63" // This is a reserved, "all dots," test code in the UK (see https://www.gbps.org.uk/tools/mechanisation/codes-overview.php)
	addrRegionCode   = "GB"
	addrLanguageCode = "en-GB"
)

// TestPBAddressConversion tests both PostalAddress conversion from and to the protocol buffer equivalent.
func TestPBAddressConversion(t *testing.T) {

	// Build a protocol buffer postal address as our starting point
	source := buildMockDeliveryAddress()

	// Convert that to our postal address type
	address := PostalAddressFromPB(source)

	// Validate the conversion
	req := require.New(t)
	req.Equal(1, len(address.Recipients), "wrong recipient count")
	req.Equal(addrRecipient, address.Recipients[0], "wrong recipient")
	req.Equal(addrOrganization, address.Organization, "wrong organization")
	req.Equal(addrMailbox, address.MailboxId, "wrong organization")
	req.Equal(2, len(address.AddressLines), "wrong address line count")
	req.Equal(addrLine1, address.AddressLines[0], "wrong address line 1")
	req.Equal(addrLine2, address.AddressLines[1], "wrong address line 1")
	req.Equal(addrSublocality, address.Sublocality, "wrong organization")
	req.Equal(addrLocality, address.Locality, "wrong organization")
	req.Equal(addrAdminArea, address.AdministrativeArea, "wrong organization")
	req.Equal(addrPostalCode, address.PostalCode, "wrong organization")
	req.Equal(addrSortingCode, address.SortingCode, "wrong organization")
	req.Equal(addrRegionCode, address.RegionCode, "wrong organization")
	req.Equal(addrLanguageCode, address.LanguageCode, "wrong organization")

	// Convert it back again and compare with the source
	pbReturn := address.AsPBPostalAddress()
	req.Equal(source, pbReturn, "converted back to protobuf did not match original")
}

// TestMilPBAddressConversion tests what happens if a nil protobuf address is passed to PostalAddressFromPB
func TestMilPBAddressConversion(t *testing.T) {

	// Ask for thePostalAddress equivalent of a nil protobuf address
	address := PostalAddressFromPB(nil)
	require.Nil(t, address, "expected nil in return for nil")
}

// buildMockDeliveryAddress returns a pbtypes.PostalAddress structure populated with the constant
// attributes defined at the head of this file to be used to create new shopping carts in our tests.
func buildMockDeliveryAddress() *pbtypes.PostalAddress {
	return &pbtypes.PostalAddress{
		Recipients:         []string{addrRecipient},
		Organization:       addrOrganization,
		MailboxId:          addrMailbox,
		AddressLines:       []string{addrLine1, addrLine2},
		Sublocality:        addrSublocality,
		Locality:           addrLocality,
		AdministrativeArea: addrAdminArea,
		PostalCode:         addrPostalCode,
		SortingCode:        addrSortingCode,
		RegionCode:         addrRegionCode,
		LanguageCode:       addrLanguageCode,
	}
}
