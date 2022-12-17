# The Order Firestore Trigger Function

The **Order Firestore Trigger** is a Cloud Function that is invoked in response to Firestore document updates. 
Specifically, creates/updates of order documents written by the [gRPC API `order-servce`](../order/README.md).

Essentially, since orders are only written to Firestore once, and the `order-service` API is read only, orders
will be published as soon as they are added to the `order-service` and not again thereafter unless a republish is 
forced.