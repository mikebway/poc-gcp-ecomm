package util

import (
	"cloud.google.com/go/firestore"
	"context"
)

// UTItemCollGetterProxy is a unit test implementation of the ItemCollectionGetterProxy interface that allows
// unit tests to have Firestore operations return errors.
type UTItemCollGetterProxy struct {
	ItemCollectionGetterProxy

	// FsClient is the GCP Firestore client - it is thread safe and can be reused concurrently
	FsClient *firestore.Client

	// Err is the error to be returned if it is not nil
	Err error

	// AllowCount, if greater than zero, is the number of calls to allow before returning the error
	AllowCount int
}

// Items returns a collection reference proxy that allows the shopping cart items of the given cart
// to be retrieved from Firestore.
func (p *UTItemCollGetterProxy) Items(path string) ItemsCollectionProxy {
	return &UTItemsCollProxy{
		Ref:        p.FsClient.Collection(path),
		Err:        p.Err,
		AllowCount: p.AllowCount,
	}
}

// UTItemsCollProxy is a unit test implementation of the ItemsCollectionProxy interface that allows
// unit tests to have Firestore operations return errors.
type UTItemsCollProxy struct {
	ItemsCollectionProxy

	// Ref is the items collection references that this proxy is wrapping
	Ref *firestore.CollectionRef

	// Err is the error to be returned if it is not nil
	Err error

	// AllowCount, if greater than zero, is the number of calls to allow before returning the error
	AllowCount int
}

// GetAll is a wrapper around the firestore.CollectionRef Documents function (actually the firestore.Query Documents
// function) that will return an iterator over shopping cart items from a Firestore cart representation.
func (p *UTItemsCollProxy) GetAll(ctx context.Context) DocumentIteratorProxy {
	return &UTDocIteratorProxy{
		DocsIterator: p.Ref.Documents(ctx),
		Err:          p.Err,
		AllowCount:   p.AllowCount,
		DsProxy:      &DocSnapProxy{},
	}
}

// UTDocRefProxy is a unit test implementation of the DocumentRefProxy interface that allows
// unit tests to have Firestore operations return errors.
type UTDocRefProxy struct {
	DocumentRefProxy

	// Err is the error to be returned if it is not nil
	Err error

	// AllowCount, if greater than zero, is the number of calls to allow before returning the error
	AllowCount int
}

// Create is a pass through to the firestore.DocumentRef Create function  that allows
// unit tests to have Firestore operations return errors.
func (p *UTDocRefProxy) Create(doc *firestore.DocumentRef, ctx context.Context, data interface{}) (*firestore.WriteResult, error) {

	// Are we to return an error and if so, do we return it now or after some later call?
	if p.Err != nil && p.AllowCount <= 0 {
		return nil, p.Err
	}

	// We are to allow the call through this time, but maybe not next time
	p.AllowCount--
	return doc.Create(ctx, data)
}

// Get is a pass through to the firestore.DocumentRef Get function that allows
// unit tests to have Firestore operations return errors.
func (p *UTDocRefProxy) Get(doc *firestore.DocumentRef, ctx context.Context) (*firestore.DocumentSnapshot, error) {

	// Are we to return an error and if so, do we return it now or after some later call?
	if p.Err != nil && p.AllowCount <= 0 {
		return nil, p.Err
	}

	// We are to allow the call through this time, but maybe not next time
	p.AllowCount--
	return doc.Get(ctx)
}

// Set is a pass through to the firestore.DocumentRef Set function that allows
// unit tests to have Firestore operations return errors.
func (p *UTDocRefProxy) Set(doc *firestore.DocumentRef, ctx context.Context, data interface{}) (*firestore.WriteResult, error) {

	// Are we to return an error and if so, do we return it now or after some later call?
	if p.Err != nil && p.AllowCount <= 0 {
		return nil, p.Err
	}

	// We are to allow the call through this time, but maybe not next time
	p.AllowCount--
	return doc.Set(ctx, data)
}

// Delete is a pass through to the firestore.DocumentRef Delete function that allows
// unit tests to have Firestore operations return errors.
func (p *UTDocRefProxy) Delete(doc *firestore.DocumentRef, ctx context.Context) (*firestore.WriteResult, error) {

	// Are we to return an error and if so, do we return it now or after some later call?
	if p.Err != nil && p.AllowCount <= 0 {
		return nil, p.Err
	}

	// We are to allow the call through this time, but maybe not next time
	p.AllowCount--
	return doc.Delete(ctx)
}

// UTDocSnapProxy is a unit test implementation of the DocumentSnapshotProxy interface that allows
// unit tests to have Firestore operations return errors.
type UTDocSnapProxy struct {
	DocumentRefProxy

	// err is the error to be returned if it is not nil
	Err error

	// AllowCount, if greater than zero, is the number of calls to allow before returning the error
	AllowCount int
}

// DataTo is a direct pass through to the firestore.DocumentSnapshot DataTo function that allows
// unit tests to have Firestore operations return errors.
func (p *UTDocSnapProxy) DataTo(snap *firestore.DocumentSnapshot, target interface{}) error {

	// Are we to return an error and if so, do we return it now or after some later call?
	if p.Err != nil && p.AllowCount <= 0 {
		return p.Err
	}

	// We are to allow the call through this time, but maybe not next time
	p.AllowCount--
	return snap.DataTo(target)
}

// UTDocIteratorProxy is a unit test implementation of the DocumentIteratorProxy interface that allows
// unit tests to have Firestore operations return errors.
type UTDocIteratorProxy struct {
	DocumentIteratorProxy

	DocsIterator *firestore.DocumentIterator
	DsProxy      DocumentSnapshotProxy

	// Err is the error to be returned if it is not nil
	Err error

	// AllowCount, if greater than zero, is the number of calls to allow before returning the error
	AllowCount int
}

// Next returns the next result. Its second return value is iterator.Done if there are no more results.
// Once Next returns Done, all subsequent calls will return Done.
func (p *UTDocIteratorProxy) Next(target interface{}) error {

	// Are we to return an error and if so, do we return it now or after some later call?
	if p.Err != nil && p.AllowCount <= 0 {
		return p.Err
	}

	// We are to allow the call through this time, but maybe not next time
	p.AllowCount--

	// Ask the real iterator for the next document ref
	snap, err := p.DocsIterator.Next()
	if err == nil {

		// We have a snapshot, unmarshal it
		return p.DsProxy.DataTo(snap, target)
	}

	// We either have a problem or just reached the end of the collection
	return err
}

// Stop stops the iterator, freeing its resources. Always call Stop when you are done with a DocumentIterator.
// It is not safe to call Stop concurrently with Next.
func (p *UTDocIteratorProxy) Stop() {
	p.DocsIterator.Stop()
}
