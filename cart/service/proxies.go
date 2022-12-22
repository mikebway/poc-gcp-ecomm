package service

import (
	"cloud.google.com/go/firestore"
	"context"
)

// ItemCollectionGetterProxy defines an interface that a Firestore client service can use to obtain an ItemsCollectionProxy
// for a given collection path. Unit tests may substitute an alternative implementation this interface in order to be able
// to insert errors etc. into the responses of the ItemsCollectionProxy that the ItemCollectionGetterProxy returns.
type ItemCollectionGetterProxy interface {
	Items(path string) ItemsCollectionProxy
}

// ItemsCollectionProxy defines the interface for a swappable junction that will allow us to maximize unit test
// coverage by wrapping calls to the firestore.CollectionRef Documents method and so allowing unit test
// implementations to return either mock errors, mock documents, or mock iterator proxies. The default production
// implementation of the interface will add a couple of nanoseconds of delay to normal operation but that extra
// test coverage is worth the price.
//
// See https://pkg.go.dev/cloud.google.com/go/firestore#CollectionRef
type ItemsCollectionProxy interface {
	GetAll(ctx context.Context) DocumentIteratorProxy
}

// QueryExecutionProxy defines an interface that a Firestore client service can use to obtain an DocumentIteratorProxy
// to walk the results of given for a given firestore.Query. Unit tests may substitute an alternative implementation
// this interface in order to be able to insert errors etc. into the responses of the DocumentIteratorProxy that the
// QueryExecutionProxy returns.
type QueryExecutionProxy interface {
	Documents(ctx context.Context, query firestore.Query) DocumentIteratorProxy
}

// DocumentRefProxy defines the interface for a swappable junction that will allow us to maximize unit test coverage
// by intercepting calls to firestore.DocumentRef methods and having them return mock errors. The default
// production implementation of the interface will add a couple of nanoseconds of delay to normal operation but that
// extra test coverage is worth the price.
//
// See https://pkg.go.dev/cloud.google.com/go/firestore#DocumentRef
type DocumentRefProxy interface {
	Create(doc *firestore.DocumentRef, ctx context.Context, data interface{}) (*firestore.WriteResult, error)
	Get(doc *firestore.DocumentRef, ctx context.Context) (*firestore.DocumentSnapshot, error)
	Set(doc *firestore.DocumentRef, ctx context.Context, data interface{}) (*firestore.WriteResult, error)
	Update(doc *firestore.DocumentRef, ctx context.Context, updates []firestore.Update) (*firestore.WriteResult, error)
	Delete(doc *firestore.DocumentRef, ctx context.Context) (*firestore.WriteResult, error)
}

// DocumentSnapshotProxy defines the interface for a swappable junction that will allow us to maximize unit test
// coverage by intercepting calls to firestore.DocumentSnapshot methods and having them return mock errors. The
// default production implementation of the interface will add a couple of nanoseconds of delay to normal
// operation but that extra test coverage is worth the price.
//
// See https://pkg.go.dev/cloud.google.com/go/firestore#DocumentSnapshot
type DocumentSnapshotProxy interface {
	DataTo(snap *firestore.DocumentSnapshot, target interface{}) error
}

// DocumentIteratorProxy defines the interface for a swappable junction that will allow us to maximize unit test
// coverage by intercepting calls to firestore.DocumentIterator methods and having them return either mock errors
// or document and/or iterator proxies. The default production implementation of the interface will add a couple
// of nanoseconds of delay to normal operation but that extra test coverage is worth the price.
//
// See https://pkg.go.dev/cloud.google.com/go/firestore#DocumentIterator
type DocumentIteratorProxy interface {
	Next(target interface{}) error
	Stop()
}

// ItemCollGetterProxy is the production (i.e. non-unit test) implementation of the ItemCollectionGetterProxy interface.
// NewCartService configures it as the default in new CartService structure constructions.
type ItemCollGetterProxy struct {
	ItemCollectionGetterProxy

	// FsClient is the GCP Firestore client - it is thread safe and can be reused concurrently
	FsClient *firestore.Client
}

// Items returns a collection reference proxy that allows the items of the given collection
// to be retrieved from Firestore.
func (p *ItemCollGetterProxy) Items(path string) ItemsCollectionProxy {
	return &ItemsCollProxy{
		ref: p.FsClient.Collection(path),
	}
}

// ItemsCollProxy is the production (i.e. non-unit test) implementation of the ItemsCollectionProxy interface.
//
// See https://pkg.go.dev/cloud.google.com/go/firestore#CollectionRef
type ItemsCollProxy struct {
	ItemsCollectionProxy

	// ref is the items collection references that this proxy is wrapping
	ref *firestore.CollectionRef
}

// GetAll is a wrapper around the firestore.CollectionRef Documents function (actually the firestore.Query Documents
// function) that will return an iterator over the items from a Firestore collection or query representation.
func (p *ItemsCollProxy) GetAll(ctx context.Context) DocumentIteratorProxy {
	return &DocIteratorProxy{
		docsIterator: p.ref.Documents(ctx),
		dsProxy:      &DocSnapProxy{},
	}
}

// QueryExecProxy implements a wrapper function around firestore.Query that will return an iterator over items
// that match the query.
type QueryExecProxy struct {
	QueryExecutionProxy
}

// Documents returns a DocumentIteratorProxy wrapping the results of the given query. Unit test implementations
// of this function can be programmed to return an iterator that can insert errors into the flow but this
// production ready implementation returns a transparent passthrough iterator.
func (q *QueryExecProxy) Documents(ctx context.Context, query firestore.Query) DocumentIteratorProxy {
	return &DocIteratorProxy{
		docsIterator: query.Documents(ctx),
		dsProxy:      &DocSnapProxy{},
	}
}

// DocRefProxy is the production (i.e. non-unit test) implementation of the DocumentRefProxy interface.
// NewCartService configures it as the default in new CartService structure constructions.
type DocRefProxy struct {
	DocumentRefProxy
}

// Create is a direct pass through to the firestore.DocumentRef Create function. We use this rather than
// calling the firestore.DocumentRef function directly so that we can replace this implementation with
// one that allows errors to be inserted into the response when executing uni tests.
func (p *DocRefProxy) Create(doc *firestore.DocumentRef, ctx context.Context, data interface{}) (*firestore.WriteResult, error) {
	return doc.Create(ctx, data)
}

// Get is a direct pass through to the firestore.DocumentRef Get function. We use this rather than
// calling the firestore.DocumentRef function directly so that we can replace this implementation with one that
// allows errors to be inserted into the response when executing uni tests.
func (p *DocRefProxy) Get(doc *firestore.DocumentRef, ctx context.Context) (*firestore.DocumentSnapshot, error) {
	return doc.Get(ctx)
}

// Set is a direct pass through to the firestore.DocumentRef Set function. We use this rather than
// calling the firestore.DocumentRef function directly so that we can replace this implementation with one that
// allows errors to be inserted into the response when executing uni tests.
func (p *DocRefProxy) Set(doc *firestore.DocumentRef, ctx context.Context, data interface{}) (*firestore.WriteResult, error) {
	return doc.Set(ctx, data)
}

func (p *DocRefProxy) Update(doc *firestore.DocumentRef, ctx context.Context, updates []firestore.Update) (*firestore.WriteResult, error) {
	return doc.Update(ctx, updates)
}

// Delete is a direct pass through to the firestore.DocumentRef Delete function. We use this rather than
// calling the firestore.DocumentRef function directly so that we can replace this implementation with one that
// allows errors to be inserted into the response when executing uni tests.
func (p *DocRefProxy) Delete(doc *firestore.DocumentRef, ctx context.Context) (*firestore.WriteResult, error) {
	return doc.Delete(ctx)
}

// DocSnapProxy is the production (i.e. non-unit test) implementation of the DocumentSnapshotProxy interface.
// NewCartService configures it as the default in new CartService structure constructions.
type DocSnapProxy struct {
	DocumentRefProxy
}

// DataTo is a direct pass through to the firestore.DocumentSnapshot DataTo function. We use this rather than
// calling the firestore.DocumentSnapshot function directly so that we can replace this implementation with
// one that allows errors to be inserted into the response when executing uni tests.
func (p *DocSnapProxy) DataTo(snap *firestore.DocumentSnapshot, target interface{}) error {
	return snap.DataTo(target)
}

// DocIteratorProxy is the production (i.e. non-unit test) implementation of the DocumentIteratorProxy interface.
//
// It is returned by calls to the CollRefProxy Documents function, itself a proxy implementation of
// firestore.CollectionRef.
type DocIteratorProxy struct {
	DocumentIteratorProxy

	docsIterator *firestore.DocumentIterator
	dsProxy      DocumentSnapshotProxy
}

// Next returns the next result. Its second return value is iterator.Done if there are no more results.
// Once Next returns Done, all subsequent calls will return Done.
func (p *DocIteratorProxy) Next(target interface{}) error {

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
func (p *DocIteratorProxy) Stop() {
	p.docsIterator.Stop()
}
