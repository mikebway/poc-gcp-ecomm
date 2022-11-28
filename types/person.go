// Package types defines generic type entity structures as they might be stored in a Google Datastore
// or represented in JSON
package types

import (
	pb "github.com/mikebway/poc-gcp-ecomm/pb/types"
)

const (
	// KeyPrefixPerson may be combined with a person's UUID ID to form the datastore key name for a person entity
	KeyPrefixPerson = "person:"
)

// Person describes a human individual
//
// See https://developers.google.com/people/api/rest/v1/people#Person.Name as field naming reference.
type Person struct {
	Id               string `firestore:"id,omitempty" json:"id,omitempty"`
	FamilyName       string `firestore:"familyName,omitempty" json:"familyName,omitempty"`
	GivenName        string `firestore:"givenName,omitempty" json:"givenName,omitempty"`
	MiddleName       string `firestore:"middleName,omitempty" json:"middleName,omitempty"`
	DisplayName      string `firestore:"displayName,omitempty" json:"displayName,omitempty"`
	DisplayLastFirst string `firestore:"displayNameLastFirst,omitempty" json:"displayNameLastFirst,omitempty"`
}

// PersonFromPB is a factory method that generates a Person structure from its Protocol Buffer equivalent
//
// If nil is passed in then nil will be returned. This saves the caller from having to check an error that
// can only occur if the caller is an idiot. Moral: do not pass in nil :-)
func PersonFromPB(pbPerson *pb.Person) *Person {
	if pbPerson == nil {
		return nil
	}
	return &Person{
		Id:          pbPerson.Id,
		FamilyName:  pbPerson.FamilyName,
		GivenName:   pbPerson.GivenName,
		MiddleName:  pbPerson.MiddleName,
		DisplayName: pbPerson.DisplayName,
	}
}

// AsPBPerson returns the protocol buffer representation of this Person.
func (p *Person) AsPBPerson() *pb.Person {
	return &pb.Person{
		Id:          p.Id,
		FamilyName:  p.FamilyName,
		GivenName:   p.GivenName,
		MiddleName:  p.MiddleName,
		DisplayName: p.DisplayName,
	}
}
