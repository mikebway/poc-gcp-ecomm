// This is a modified work derived from original source code copyright 2016 Google LLC. See
// https://github.com/googleapis/googleapis/blob/master/google/type/postal_address.proto.
//
// Those portions of this work that are copied from the Google original are subject to the
// following copyright:
//
// Copyright 2016 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import pb "github.com/mikebway/poc-gcp-ecomm/pb/types"

const (
	// KeyPrefixPostalAddress may be combined with a postal address's UUID ID to form the datastore key name for a postal address entity
	KeyPrefixPostalAddress = "postaladdress:"
)

// PostalAddress represents a postal address, e.g. for postal delivery or payments
// addresses. Given a postal address, a postal service can deliver items to a
// premise, P.O. Box or similar. It is not intended to model geographical locations
// (roads, towns, mountains).
//
// In typical usage an address would be created via user input or from importing
// existing data, depending on the type of process.
//
// Advice on address input / editing:
//   - Use an i18n-ready address widget such as
//     https://github.com/googlei18n/libaddressinput)
//   - Users should not be presented with UI elements for input or editing of
//     fields outside countries where that field is used.
//
// For more guidance on how to use this schema, please see:
// https://support.google.com/business/answer/6397478
type PostalAddress struct {
	// RegionCode (Required) id the  CLDR region code of the country/region of the
	// address. This is never inferred and it is up to the user to ensure the value
	// is correct. See http://cldr.unicode.org/ and
	// http://www.unicode.org/cldr/charts/30/supplemental/territory_information.html
	// for details. Example: "CH" for Switzerland.
	RegionCode string `datastore:"regionCode,omitempty" json:"regionCode,omitempty"`

	// LanguageCode (Optional) is BCP-47 language code of the contents of this address
	// (if known). This is often the UI language of the input form or is expected
	// to match one of the languages used in the address' country/region, or their
	// transliterated equivalents.
	// This can affect formatting in certain countries, but is not critical
	// to the correctness of the data and will never affect any validation or
	// other non-formatting related operations.
	//
	// If this value is not known, it should be omitted (rather than specifying a
	// possibly incorrect default).
	//
	// Examples: "zh-Hant", "ja", "ja-Latn", "en".
	LanguageCode string `datastore:"languageCode,omitempty" json:"languageCode,omitempty"`

	// PostalCode (Optional) is the postal code of the address. Not all countries use or
	// require postal codes to be present, but where they are used, they may trigger
	// additional validation with other parts of the address (e.g. state/zip
	// validation in the U.S.A.).
	PostalCode string `datastore:"postalCode,omitempty" json:"postalCode,omitempty"`

	// SortingCode (Optional) is additional, country-specific, sorting code. This is not
	// used in most regions. Where it is used, the value is either a string like
	// "CEDEX", optionally followed by a number (e.g. "CEDEX 7"), or just a number
	// alone, representing the "sector code" (Jamaica), "delivery area indicator"
	// (Malawi) or "post office indicator" (e.g. Côte d'Ivoire).
	SortingCode string `datastore:"sortingCode,omitempty" json:"sortingCode,omitempty"`

	// AdministrativeArea (Optional) is the highest administrative subdivision which
	//  is used for postal addresses of a country or region.
	// For example, this can be a state, a province, an oblast, or a prefecture.
	// Specifically, for Spain this is the province and not the autonomous
	// community (e.g. "Barcelona" and not "Catalonia").
	// Many countries don't use an administrative area in postal addresses. E.g.
	// in Switzerland this should be left unpopulated.
	AdministrativeArea string `datastore:"administrativeArea,omitempty" json:"administrativeArea,omitempty"`

	// Locality (Optional) generally refers to the city/town portion of the address.
	// Examples: US city, IT comune, UK post town.
	// In regions of the world where localities are not well defined or do not fit
	// into this structure well, leave locality empty and use address_lines.
	Locality string `datastore:"locality,omitempty" json:"locality,omitempty"`

	// Sublocality (Optional) is a subdivision of the locality of the address.
	// For example, this can be neighborhoods, boroughs, districts.
	Sublocality string `datastore:"sublocality,omitempty" json:"sublocality,omitempty"`

	// AddressLines (Optional) are unstructured address lines describing the lower
	// levels of an address.
	//
	// Because values in address_lines do not have type information and may
	// sometimes contain multiple values in a single field (e.g.
	// "Austin, TX"), it is important that the line order is clear. The order of
	// address lines should be "envelope order" for the country/region of the
	// address. In places where this can vary (e.g. Japan), address_language is
	// used to make it explicit (e.g. "ja" for large-to-small ordering and
	// "ja-Latn" or "en" for small-to-large). This way, the most specific line of
	// an address can be selected based on the language.
	//
	// The minimum permitted structural representation of an address consists
	// of a region_code with all remaining information placed in the
	// address_lines. It would be possible to format such an address very
	// approximately without geocoding, but no semantic reasoning could be
	// made about any of the address components until it was at least
	// partially resolved.
	//
	// Creating an address only containing a region_code and address_lines, and
	// then geocoding is the recommended way to handle completely unstructured
	// addresses (as opposed to guessing which parts of the address should be
	// localities or administrative areas).
	AddressLines []string `datastore:"addressLines,omitempty" json:"addressLines,omitempty"`

	// Recipients (Optional) are the recipients at the address.
	// This field may, under certain circumstances, contain multiline information.
	// For example, it might contain "care of" information.
	Recipients []string `datastore:"recipients,omitempty" json:"recipients,omitempty"`

	// Organization (Optional) is the name of the organization (e.g. business) at
	// the address.
	Organization string `datastore:"organization,omitempty" json:"organization,omitempty"`

	// MailboxId (Optional) is the identifier of a mailbox at this address. Used in the
	// case where multiple entities may be registered at the same address, divided
	// by per entity delivery boxes.
	//
	// Some address values might specify the mailbox as an address line but this field
	// allows for a mailbox to be explicitly labeled in the data.
	MailboxId string `datastore:"mailboxId,omitempty" json:"mailboxId,omitempty"`
}

// PostalAddressFromPB is a factory method that generates a PostalAddress structure from its Protocol Buffer equivalent
//
// If nil is passed in then nil will be returned. This saves the caller from having to check an error that
// can only occur if the caller is an idiot. Moral: do not pass in nil :-)
func PostalAddressFromPB(p *pb.PostalAddress) *PostalAddress {
	if p == nil {
		return nil
	}
	return &PostalAddress{
		RegionCode:         p.RegionCode,
		LanguageCode:       p.LanguageCode,
		PostalCode:         p.PostalCode,
		SortingCode:        p.SortingCode,
		AdministrativeArea: p.AdministrativeArea,
		Locality:           p.Locality,
		Sublocality:        p.Sublocality,
		AddressLines:       p.AddressLines,
		Recipients:         p.Recipients,
		Organization:       p.Organization,
		MailboxId:          p.MailboxId,
	}
}

// AsPBPostalAddress returns the protocol buffer representation of this Person.
func (p *PostalAddress) AsPBPostalAddress() *pb.PostalAddress {
	return &pb.PostalAddress{
		RegionCode:         p.RegionCode,
		LanguageCode:       p.LanguageCode,
		PostalCode:         p.PostalCode,
		SortingCode:        p.SortingCode,
		AdministrativeArea: p.AdministrativeArea,
		Locality:           p.Locality,
		Sublocality:        p.Sublocality,
		AddressLines:       p.AddressLines,
		Recipients:         p.Recipients,
		Organization:       p.Organization,
		MailboxId:          p.MailboxId,
	}
}
