// Package schema defines order document structures as they might be stored in a Google Firestore
// or represented in JSON
package schema

import (
	"github.com/mikebway/poc-gcp-ecomm/types"
	"time"
)

// Order represents the permanent record of what a customer has purchased. An order is derived from a shopping cart
// upon checkout.
type Order struct {

	// TODO: Perhaps include omitempty on all fields - they would not be valid orders
	//       but if an order is presented from the shopping cart we should perhaps record it
	//       in Firestore and return it in API requests as a record of whatever came out of
	//       an apparently submitted cart

	// Id is a UUID ID in hexadecimal string form - a unique ID for this order.
	// This will be set by the cart when the order is submitted.
	Id string `firestore:"id" json:"id"`

	// SubmissionTime is the time at which cart checkout was completed and the order was submitted for payment
	SubmissionTime time.Time `firestore:"submissionTime" json:"submissionTime"`

	// OrderedBy identifies the person who submitted the order
	OrderedBy *types.Person `firestore:"orderedBy" json:"orderedBy"`

	// DeliveryAddress is the delivery address for the order
	DeliveryAddress *types.PostalAddress `firestore:"deliveryAddress,omitempty" json:"deliveryAddress,omitempty"`

	// OrderItems is the list of one to many items that make up the potential order
	OrderItems []*OrderItem `firestore:"cartItems" json:"cartItems"`
}

// OrderItem represents a single entry in an order. An order will contain one
// to many order items.
type OrderItem struct {

	// TODO: Perhaps include omitempty on all fields - they would npt be valid orders
	//       but if an order is present from teh shopping cart we should perhaps record it
	//       in Firestore and return it in API requests as a record of whatever came out of
	//       an apparently submitted cart

	// Id is a UUID ID in hexadecimal string form - a unique ID for this order item.
	Id string `firestore:"id" json:"id"`

	// ProductCode is the equivalent of a SKU code identifying the type of
	// product or service ordered.
	ProductCode string `firestore:"productCode" json:"productCode"`

	// Quantity is the number of this item type that was ordered.
	Quantity int32 `firestore:"quantity" json:"quantity"`

	// UnitPrice is the price that the customer was shown for a single item
	// when they selected the item for their cart.
	UnitPrice *types.Money `firestore:"unitPrice" json:"unitPrice"`
}
