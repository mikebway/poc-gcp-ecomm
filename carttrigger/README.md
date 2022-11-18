# The Cart Firestore Trigger Function

The **Cart Firestore Trigger** is a Cloud Function that is invoked in response to Firestore document updates. 
Specifically, updates of root shopping cart documents written by the [gRPC API `cart-servce`](../cart/README.md).

Cart creations are ignored, only updates are of interest. As the [`cart-servce`](../cart/README.md) is currently
implemented updates only happen in two circumstances: when the cart is checked out or when it is abandoned.

Updates to items in the cart or to the delivery address are not monitored and are not of interest to the trigger.

When a cart is tagged as "checked out" and this Cloud Function is triggered, it will gather the complete data
for teh cart, including cart items and delivery address, and submit the resulting package to a Cloud Task for
insertion into the "order archive."
