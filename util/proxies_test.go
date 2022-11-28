package util

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/mikebway/poc-gcp-ecomm/cart/schema"
)

// UTItemCollGetterProxy is a unit test implementation of the ItemCollectionGetterProxy interface that allows
// unit tests to have Firestore operations return errors.
type UTItemCollGetterProxy struct {
	ItemCollectionGetterProxy

	// FsClient is the GCP Firestore client - it is thread safe and can be reused concurrently
	FsClient *firestore.Client

	// err is the error to be returned if it is not nil
	err error

	// allowCount, if greater than zero, is the number of calls to allow before returning the error
	allowCount int
}

// Items returns a collection reference proxy that allows the shopping cart items of the given cart
// to be retrieved from Firestore.
func (p *UTItemCollGetterProxy) Items(cart *schema.ShoppingCart) ItemsCollectionProxy {
	return &UTItemsCollProxy{
		ref:        p.FsClient.Collection(cart.ItemCollectionPath()),
		err:        p.err,
		allowCount: p.allowCount,
	}
}

// UTItemsCollProxy is a unit test implementation of the ItemsCollectionProxy interface that allows
// unit tests to have Firestore operations return errors.
type UTItemsCollProxy struct {
	ItemsCollectionProxy

	// ref is the items collection references that this proxy is wrapping
	ref *firestore.CollectionRef

	// err is the error to be returned if it is not nil
	err error

	// allowCount, if greater than zero, is the number of calls to allow before returning the error
	allowCount int
}

// GetAll is a wrapper around the firestore.CollectionRef Documents function (actually the firestore.Query Documents
// function) that will return an iterator over shopping cart items from a Firestore cart representation.
func (p *UTItemsCollProxy) GetAll(ctx context.Context) DocumentIteratorProxy {
	return &UTDocIteratorProxy{
		docsIterator: p.ref.Documents(ctx),
		err:          p.err,
		allowCount:   p.allowCount,
		dsProxy:      &DocSnapProxy{},
	}
}

// UTDocRefProxy is a unit test implementation of the DocumentRefProxy interface that allows
// unit tests to have Firestore operations return errors.
type UTDocRefProxy struct {
	DocumentRefProxy

	// err is the error to be returned if it is not nil
	err error

	// allowCount, if greater than zero, is the number of calls to allow before returning the error
	allowCount int
}

// Create is a pass through to the firestore.DocumentRef Create function  that allows
// unit tests to have Firestore operations return errors.
func (p *UTDocRefProxy) Create(doc *firestore.DocumentRef, ctx context.Context, data interface{}) (*firestore.WriteResult, error) {

	// Are we to return an error and if so, do we return it now or after some later call?
	if p.err != nil && p.allowCount <= 0 {
		return nil, p.err
	}

	// We are to allow the call through this time, but maybe not next time
	p.allowCount--
	return doc.Create(ctx, data)
}

// Get is a pass through to the firestore.DocumentRef Get function that allows
// unit tests to have Firestore operations return errors.
func (p *UTDocRefProxy) Get(doc *firestore.DocumentRef, ctx context.Context) (*firestore.DocumentSnapshot, error) {

	// Are we to return an error and if so, do we return it now or after some later call?
	if p.err != nil && p.allowCount <= 0 {
		return nil, p.err
	}

	// We are to allow the call through this time, but maybe not next time
	p.allowCount--
	return doc.Get(ctx)
}

// Set is a pass through to the firestore.DocumentRef Set function that allows
// unit tests to have Firestore operations return errors.
func (p *UTDocRefProxy) Set(doc *firestore.DocumentRef, ctx context.Context, data interface{}) (*firestore.WriteResult, error) {

	// Are we to return an error and if so, do we return it now or after some later call?
	if p.err != nil && p.allowCount <= 0 {
		return nil, p.err
	}

	// We are to allow the call through this time, but maybe not next time
	p.allowCount--
	return doc.Set(ctx, data)
}

// Delete is a pass through to the firestore.DocumentRef Delete function that allows
// unit tests to have Firestore operations return errors.
func (p *UTDocRefProxy) Delete(doc *firestore.DocumentRef, ctx context.Context) (*firestore.WriteResult, error) {

	// Are we to return an error and if so, do we return it now or after some later call?
	if p.err != nil && p.allowCount <= 0 {
		return nil, p.err
	}

	// We are to allow the call through this time, but maybe not next time
	p.allowCount--
	return doc.Delete(ctx)
}

// UTDocSnapProxy is a unit test implementation of the DocumentSnapshotProxy interface that allows
// unit tests to have Firestore operations return errors.
type UTDocSnapProxy struct {
	DocumentRefProxy

	// err is the error to be returned if it is not nil
	err error

	// allowCount, if greater than zero, is the number of calls to allow before returning the error
	allowCount int
}

// DataTo is a direct pass through to the firestore.DocumentSnapshot DataTo function that allows
// unit tests to have Firestore operations return errors.
func (p *UTDocSnapProxy) DataTo(snap *firestore.DocumentSnapshot, target interface{}) error {

	// Are we to return an error and if so, do we return it now or after some later call?
	if p.err != nil && p.allowCount <= 0 {
		return p.err
	}

	// We are to allow the call through this time, but maybe not next time
	p.allowCount--
	return snap.DataTo(target)
}

// UTDocIteratorProxy is a unit test implementation of the DocumentIteratorProxy interface that allows
// unit tests to have Firestore operations return errors.
type UTDocIteratorProxy struct {
	DocumentIteratorProxy

	docsIterator *firestore.DocumentIterator
	dsProxy      DocumentSnapshotProxy

	// err is the error to be returned if it is not nil
	err error

	// allowCount, if greater than zero, is the number of calls to allow before returning the error
	allowCount int
}

// Next returns the next result. Its second return value is iterator.Done if there are no more results.
// Once Next returns Done, all subsequent calls will return Done.
func (p *UTDocIteratorProxy) Next(target *schema.ShoppingCartItem) error {

	// Are we to return an error and if so, do we return it now or after some later call?
	if p.err != nil && p.allowCount <= 0 {
		return p.err
	}

	// We are to allow the call through this time, but maybe not next time
	p.allowCount--

	// Ask the real iterator for the next document ref
	snap, err := p.docsIterator.Next()
	if err == nil {

		// We have a snapshot, unmarshal it
		return p.dsProxy.DataTo(snap, target)
	}

	// We either have a problem or just reached the end of the collection
	return err
}

// Stop stops the iterator, freeing its resources. Always call Stop when you are done with a DocumentIterator.
// It is not safe to call Stop concurrently with Next.
func (p *UTDocIteratorProxy) Stop() {
	p.docsIterator.Stop()
}
