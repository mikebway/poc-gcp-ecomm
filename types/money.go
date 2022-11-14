// This is a modified work derived from original source code copyright 2021 Google LLC. See
// https://github.com/googleapis/googleapis/blob/master/google/type/money.proto.
//
// Those portions of this work that are copied from the Google original are subject to the
// following copyright:
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

import (
	pbmoney "google.golang.org/genproto/googleapis/type/money"
)

// Money is a representation of a currency value that can be directly mapped to the Google APIs
// protocol buffer representation of money.
//
// Why not use the Google APIs Money structure directly? Well, it includes a mutex which makes it
// difficult / dangerous to copy and gives rise to compiler warnings. It also lacks the manipulation
// methods that we might want to add at some point (e.g. like those offered by https://github.com/Rhymond/go-money).
//
// And finally, we want control over how the structure might be stored in a Google Cloud Datastore
// and JSON.
type Money struct {

	// The three-letter currency code defined in ISO 4217.
	CurrencyCode string `firestore:"currencyCode,omitempty" json:"currencyCode,omitempty"`

	// The whole units of the amount.
	// For example if `currencyCode` is `"USD"`, then 1 unit is one US dollar.
	Units int64 `firestore:"units,omitempty" json:"units,omitempty"`

	// Number of nano (10^-9) units of the amount.
	// The value must be between -999,999,999 and +999,999,999 inclusive.
	// If `units` is positive, `nanos` must be positive or zero.
	// If `units` is zero, `nanos` can be positive, zero, or negative.
	// If `units` is negative, `nanos` must be negative or zero.
	// For example $-1.75 is represented as `units`=-1 and `nanos`=-750,000,000.
	Nanos int32 `firestore:"nanos,omitempty" json:"nanos,omitempty"`
}

// MoneyFromPB is a factory method that generates a Money structure from its Protocol Buffer equivalent
//
// If nil is passed in then nil will be returned. This saves the caller from having to check an error that
// can only occur if the caller is an idiot. Moral: do not pass in nil :-)
func MoneyFromPB(pbm *pbmoney.Money) *Money {
	if pbm == nil {
		return nil
	}
	return &Money{
		CurrencyCode: pbm.CurrencyCode,
		Units:        pbm.Units,
		Nanos:        pbm.Nanos,
	}
}

// AsPBMoney returns the protocol buffer representation of this Money.
func (m *Money) AsPBMoney() *pbmoney.Money {
	return &pbmoney.Money{
		CurrencyCode: m.CurrencyCode,
		Units:        m.Units,
		Nanos:        m.Nanos,
	}
}
