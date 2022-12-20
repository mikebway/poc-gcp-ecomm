// Package schema defines order document structures as they might be stored in a Google Firestore
// or represented in JSON
package schema

import (
	pborder "github.com/mikebway/poc-gcp-ecomm/pb/order"
	pbtypes "github.com/mikebway/poc-gcp-ecomm/pb/types"
	"github.com/mikebway/poc-gcp-ecomm/types"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

const (
	// OrderCollection names the firestore collection under which all of our documents are stored
	OrderCollection = "orders"
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
	OrderItems []*OrderItem `firestore:"orderItems" json:"orderItems"`
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

// StoreRefPath returns the string representation of the document reference path for this ShoppingCart.
func (o *Order) StoreRefPath() string {
	return OrderCollection + "/" + o.Id
}

// AsPBOrder returns the protocol buffer representation of this order.
func (o *Order) AsPBOrder() *pborder.Order {

	// Submission time should be set, but we will play it safe just the same
	var pbSubmissionTime *timestamppb.Timestamp
	if !o.SubmissionTime.IsZero() {
		pbSubmissionTime = timestamppb.New(o.SubmissionTime)
	}

	// Only convert the person that put in the order  if they have been defined in the order.
	// That should always be the case but we will play defensively.
	var pbOrderedBy *pbtypes.Person
	if o.OrderedBy != nil {
		pbOrderedBy = o.OrderedBy.AsPBPerson()
	}

	// Only convert the delivery address if that has been defined in the orded
	var pbAddress *pbtypes.PostalAddress
	if o.DeliveryAddress != nil {
		pbAddress = o.DeliveryAddress.AsPBPostalAddress()
	}

	// Finally, add the order items (if any)
	pbItems := make([]*pborder.OrderItem, len(o.OrderItems))
	for i, item := range o.OrderItems {
		pbItems[i] = item.AsPBOrderItem()
	}

	// Return a populated protocol buffer version of the order
	return &pborder.Order{
		Id:              o.Id,
		SubmissionTime:  pbSubmissionTime,
		OrderedBy:       pbOrderedBy,
		DeliveryAddress: pbAddress,
		OrderItems:      pbItems,
	}
}

// AsPBOrderItem returns the protocol buffer representation of this cart item.
func (item *OrderItem) AsPBOrderItem() *pborder.OrderItem {
	return &pborder.OrderItem{
		Id:          item.Id,
		ProductCode: item.ProductCode,
		Quantity:    item.Quantity,
		UnitPrice:   item.UnitPrice.AsPBMoney(),
	}
}
