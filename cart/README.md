# The gRPC Cart Microservice

The **Cart Service** is a crude implementation of an e-commerce shopping cart. It's purpose is not to serve in an
actual online business but as a test bed to explore:

* Building and deploying a GCP Cloud Run service
* Implementing a gRPC API CRUD microservice on Cloud Run
* Utilizing Cloud Firestore as the backend database

## Planned Enhancements

The current implementation is very basic, far less than MVP level. Still to explore are:

* Figure out a better way to get high unit test coverage on Firestore access code
* Authorization and access control
* Transactions
* Specifying preconditions
* Use of queries to retrieve multiple Firestore documents in a single roundtrip
* CDC triggering of EventArc events, Cloud Tasks, or other GCP capabilities 
* gRPC field masking to select fields to be returned
* gRPC request validation: required fields are present, field values are valid, etc
* Adding a search API utilizing the Firestore indexes

## How to Exercise the Cart API

The Cart Service is the subject of the [Use BloomRPC to invoke gRPC Cloud Run services](docs/BLOOMRPC.md) guide
so we won't repeat that documentation here. Instead, we will provide an overview of how the API might be used
in a typical cart workflow along with some sample data to clarify what the inputs might actually look like (the 
BloomRPC template displays verbosely offer all of the possible fields without any explanation of what each
value might need to be or whether the fields are required or optional).

### Creating a New Cart: `CreateShoppingCart`

* Any ID field in the request will be ignored; it can be removed before the request is submitted. The **Cart Service**
  will always generate a new ID for a new cart.

Request example:

```json
{
  "shopper": {
    "family_name": "Smith",
    "given_name": "John",
    "middle_name": "James Henry",
    "display_name": "JJ"
  }
}
```

**IMPORTANT:** Make a note of the UUID cart ID returned in the response. At the time of writing there is no search
API to allow you to find it again other than through the GCP Firestore Web console.

### Adding a Delivery Address: `SetDeliveryAddress`

Set the `cart_id` value to one obtained in the `CreateShoppingCart` response.

The address message structure, originated by Google, is designed to support homes and businesses anywhere in the 
world, as a result it is not intuitively obvious how each of the fields offered map to a US address, even if you
read the comments in the Protocol Buffer source. 

The `region_code` identifies the country and should be expressed as a CLDR value (Unicode Common Locale Data Repository).
The documentation for the [https://cldr.unicode.org/index](https://cldr.unicode.org/index) is extensive, comprehensive,
and utterly opaque but you can find the `region_code` values defined as "territories" in the 
[Territory Information Index](https://unicode-org.github.io/cldr-staging/charts/42/supplemental/territory_information.html)
on GitHub.

Request example for a US address:

```json
{
  "cart_id": "1455b26a-7c6a-4608-af2b-1037d6fa7047",
  "delivery_address": {
    "region_code": "US",
    "postal_code": "42952",
    "administrative_area": "Kentucky",
    "locality": "Big Town",
    "address_lines": [
      "7243 Hampton Parkway"
    ]
  }
}
```

**NOTE:** The ZIP code prefix of 429 is not in use. The above address cannot exist in the real world.

### Adding an Item to the Cart: `AddItemToShoppingCart`

Set the `cart_id` value to one obtained in the `CreateShoppingCart` response.

The template offered by **BloomRPC** will include generated UUID values for the item `id` and `cart_id`; you can
strip these out as they will be overwritten in by the API code and returned in the response as read-only values.

The `unit_price` is expressed as structure modeled as a near exact clone defined by 
[Google API's](https://github.com/googleapis/googleapis) [money type](https://github.com/googleapis/googleapis/blob/master/google/type/money.proto).
The only reason that we did not copy the Apache licensed code exactly is that it includes a mutex that made 
it impossible for code to copying the structure contents between instances. The things to know are: 

* `units` should be the the whole units of the amount. For example if `currency_code` is `USD`, then 1 unit is one US dollar.
* `nanos` is the number of nano units in the price so, `990000000` is 99 cents for a `currency_code` of `USD`.
   * If `units` is positive, `nanos` must be positive or zero. 
   * If `units` is zero, `nanos` can be positive, zero, or negative. 
   * If `units` is negative, `nanos` must be negative or zero. For example $-1.75 is represented as `units`=-1 and `nanos`=-750,000,000.

```json
{
  "cart_id": "1455b26a-7c6a-4608-af2b-1037d6fa7047",
  "item": {
    "product_code": "gold_yoyo",
    "quantity": 10,
    "unit_price": {
      "currency_code": "USD",
      "units": 21499,
      "nanos": 990000000
    }
  }
}
```

### Checking Out or Abandoning the Cart: `CheckoutShoppingCart` or `AbandonShoppingCart`

Both the check out and abandon operations take the same minimal inout of just the cart ID.

Once the status the cart has been so modified, it cannot be changed again. In theory, you should **not**
be able to add or remove cart items after this point but, at the time of writing, there are no checks
in place to prevent that.

```json
{
  "cart_id": "1455b26a-7c6a-4608-af2b-1037d6fa7047"
}
```
