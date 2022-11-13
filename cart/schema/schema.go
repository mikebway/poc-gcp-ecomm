// Package schema defines generic type document structures as they might be stored in a Google Firestore
// or represented in JSON
package schema

import (
	pbcart "github.com/mikebway/poc-gcp-ecomm/pb/cart"
	pbtypes "github.com/mikebway/poc-gcp-ecomm/pb/types"
	"github.com/mikebway/poc-gcp-ecomm/types"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

const (
	// CartCollection names the firestore collection under which all of our documents are stored
	CartCollection = "carts/"

	// ItemCollection names the sub-collection of an individual cart in which shopping cart item documents are stored
	ItemCollection = "/items"

	// AddressCollection names the sub-collection of an individual cart in which postal address documents are stored
	AddressCollection = "/addresses"

	// DeliverAddressDoc names the single delivery address document that may bee associated with the AddressCollection
	// of a single shopping cart.
	DeliverAddressDoc = "/delivery"
)

// ShoppingCart collects the cart items that a shopper is considering purchasing
// or has purchased. A cart should be considered immutable once purchase has been
// processed.
//
// It is persisted in the cart firestore collection.
type ShoppingCart struct {
	// A UUID ID in hexadecimal string form - a unique ID for this cart.
	// This will be set by the cart service when the cart is first created.
	Id string `firestore:"id" json:"id,omitempty"`

	// CreationTime is the time at which the cart was created
	CreationTime time.Time `firestore:"creationTime" json:"creationTime"`

	// ClosedTime (Optional) is the time at which shopping cart was closed, either
	// as abandoned or submitted / checked out.
	ClosedTime time.Time `firestore:"closedTime,omitempty" json:"closedTime,omitempty"`

	// Status describes the state of the shopping cart (duh!).
	Status CartStatus `firestore:"status" json:"status"`

	// Shopper identifies the person who submitted the order
	Shopper *types.Person `firestore:"shopper" json:"shopper"`

	// DeliveryAddress is the postal address to which any physical items in the order
	// are to be delivered.
	//
	// NOTE: delivery address is stored as a separate sub-document in the Google Firestore with
	// the cart reference as their ancestor.
	DeliveryAddress *types.PostalAddress `firestore:"-" json:"deliveryAddress"`

	// CartItems  is the list of one to many items that make up the potential order.
	//
	// NOTE: cart items are stored as separate sub-documents in the Google Firestore with
	// the cart reference as their ancestor.
	CartItems []*ShoppingCartItem `firestore:"-" json:"cartItems"`
}

// CartStatus is an enumeration type defining the overall status of a shopping cart
type CartStatus int32

const (
	CsUnspecified        CartStatus = 0
	CsOpen               CartStatus = 1
	CsCheckedOut         CartStatus = 2
	CsAbandonedByUser    CartStatus = 3
	CsAbandonedByTimeout CartStatus = 4
)

// StoreRefPath returns the string representation of the document reference path for this ShoppingCart.
func (c *ShoppingCart) StoreRefPath() string {
	return CartCollection + c.Id
}

// DeliveryAddressPath returns the string representation of the document reference path for the one and only
// delivery address that can be associated with ths ShoppingCart.
func (c *ShoppingCart) DeliveryAddressPath() string {
	return CartCollection + c.Id + AddressCollection + DeliverAddressDoc
}

// ItemCollectionPath returns the string representation of the collection reference path under which
// cart items associated with ths ShoppingCart may be stored
func (c *ShoppingCart) ItemCollectionPath() string {
	return CartCollection + c.Id + ItemCollection
}

// AsPBShoppingCart returns the protocol buffer representation of this cart.
func (c *ShoppingCart) AsPBShoppingCart() *pbcart.ShoppingCart {

	// Creation time should be set, but we will play it safe just the same
	var creationTimePB *timestamppb.Timestamp
	if !c.CreationTime.IsZero() {
		creationTimePB = timestamppb.New(c.CreationTime)
	}

	// ClosedTime will frequently not be set
	var closedTimePB *timestamppb.Timestamp
	if !c.ClosedTime.IsZero() {
		closedTimePB = timestamppb.New(c.ClosedTime)
	}

	// Only convert the shopper if they have been defined in the cart
	// That should always be the case but we will play defensively.
	var pbShopper *pbtypes.Person
	if c.Shopper != nil {
		pbShopper = c.Shopper.AsPBPerson()
	}

	// Only convert the delivery address if that has been defined in the cart
	var pbAddress *pbtypes.PostalAddress
	if c.DeliveryAddress != nil {
		pbAddress = c.DeliveryAddress.AsPBPostalAddress()
	}

	// Finally, add the cart items (if any)
	pbItems := make([]*pbcart.CartItem, len(c.CartItems))
	for i, item := range c.CartItems {
		pbItems[i] = item.AsPBCartItem()
	}

	// Return a populated protocol buffer version of the cart
	return &pbcart.ShoppingCart{
		Id:              c.Id,
		CreationTime:    creationTimePB,
		ClosedTime:      closedTimePB,
		Status:          pbcart.ShoppingCartStatus(c.Status),
		Shopper:         pbShopper,
		DeliveryAddress: pbAddress,
		CartItems:       pbItems,
	}
}

// ShoppingCartFromPB is a factory method that populates a ShoppingCart structure from its
// protocol buffer equivalent.
func ShoppingCartFromPB(pbc *pbcart.ShoppingCart) *ShoppingCart {

	// Creation time should be set, but we will play it safe just the same
	var creationTime time.Time
	if pbc.CreationTime != nil {
		creationTime = pbc.CreationTime.AsTime()
	}

	// ClosedTime will frequently not be set
	var closedTime time.Time
	if pbc.ClosedTime != nil {
		closedTime = pbc.ClosedTime.AsTime()
	}

	// Only convert the shopper if they have been defined in the cart
	// That should always be the case but we will play defensively.
	var shopper *types.Person
	if pbc.Shopper != nil {
		shopper = types.PersonFromPB(pbc.Shopper)
	}

	// Only convert the delivery address if that has been defined in the cart
	var address *types.PostalAddress
	if pbc.DeliveryAddress != nil {
		address = types.PostalAddressFromPB(pbc.DeliveryAddress)
	}

	// Finally, add the cart items (if any)
	items := make([]*ShoppingCartItem, len(pbc.CartItems))
	for i, pbItem := range pbc.CartItems {
		items[i] = ShoppingCartItemFromPB(pbItem)
	}

	// Return a populated protocol buffer version of the cart
	return &ShoppingCart{
		Id:              pbc.Id,
		CreationTime:    creationTime,
		ClosedTime:      closedTime,
		Status:          CartStatus(pbc.Status),
		Shopper:         shopper,
		DeliveryAddress: address,
		CartItems:       items,
	}
}

// ShoppingCartItem represents a single entry in an order. An order will contain one
// to many order items.
type ShoppingCartItem struct {
	// Id is a UUID ID in hexadecimal string form - a unique ID for this cart item.
	Id string `firestore:"id" json:"id,omitempty"`

	// CartId is a UUID ID in hexadecimal string form - a unique ID for  this item's parent cart
	CartId string `firestore:"cartId" json:"cartId,omitempty"`

	// ProductCode is the equivalent of a SKU code identifying the type of
	// product or service being ordered.
	ProductCode string `firestore:"productCode" json:"productCode"`

	// Quantity is the number of this item type that is being ordered.
	Quantity int32 `firestore:"quantity" json:"quantity"`

	// UnitPrice is the price that the customer was shown for a single item
	// when they selected the item for their cart
	UnitPrice *types.Money `firestore:"unitPrice" json:"unitPrice"`
}

// StoreRefPath returns the string representation of the document reference path for this ShoppingCartItem.
func (item *ShoppingCartItem) StoreRefPath() string {
	return CartCollection + item.CartId + ItemCollection + "/" + item.Id
}

// AsPBCartItem returns the protocol buffer representation of this cart item.
func (item *ShoppingCartItem) AsPBCartItem() *pbcart.CartItem {
	return &pbcart.CartItem{
		Id:          item.Id,
		CartId:      item.CartId,
		ProductCode: item.ProductCode,
		Quantity:    item.Quantity,
		UnitPrice:   item.UnitPrice.AsPBMoney(),
	}
}

// ShoppingCartItemFromPB is a factory method that returns a ShoppingCartItem representation
// derived from its protocol buffer equivalent.
func ShoppingCartItemFromPB(pbItem *pbcart.CartItem) *ShoppingCartItem {
	return &ShoppingCartItem{
		Id:          pbItem.Id,
		CartId:      pbItem.CartId,
		ProductCode: pbItem.ProductCode,
		Quantity:    pbItem.Quantity,
		UnitPrice:   types.MoneyFromPB(pbItem.UnitPrice),
	}
}
