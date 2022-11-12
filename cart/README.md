# The gRPC Cart Microservice

The **Cart Service** is a crude implementation of an e-commerce shopping cart. It's purpose is not to serve in an
actual online business but as a test bed to explore:

* Building and deploying a GCP Cloud Run service
* Implementing a gRPC API CRUD microservice on Cloud Run
* Utilizing Cloud Firestore as the backend database

## Planned Enhancements

The current implementation is very basic, far less than MVP level. Still to explore are:

* Authorization and access control
* Transactions
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