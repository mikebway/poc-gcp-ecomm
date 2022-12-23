# e-Commerce Prototype Overview

![POC Component Relationships](docs/poc-gcp-ecomm-scenario.png)

## Data Flow

### Building the Shopping Cart

The process starts, naturally, with the [Shopping Cart gRPC API](../cart/README.md). The cart API is used to create
a new cart, add items to it, set the delivery address, then checkout. The cart is written to a Firestore "carts" 
document collection, with items and the deliverey address as child documents under the cart.

In the real world, checkout would entail collecting payment etc but this POC skips over all that.

Every update to a shopping cart (not to the items it contains) results in the [Cart Firestore Trigger](../carttrigger/README.md)
Cloud Function being invoked. If this trigger function sees that the cart has been completed / checked out, then it 
retrieves the full cart content, including items and delivery address and publishes that as protocol buffer message 
to an "ecomm-cart" Pub/Sub topic.

### Recording the Order

The cart Pub/Sub topic is configured to push cart messages to the [Order from Cart Topic Consumer](../orderfromcart/README.md)
Cloud Function. This function simply writes the order to a Firestore "carts" document collection as single documents,
i.e. with the items as a list of maps field and the delivery address as a map field.

The [Order History gRPC API](../order/README.md) is purely read only. Orders are immutable after the shopping cart 
has been checked out. 

As soon as the order has been recorded for posterity in Firestore, the [Order Firestore Trigger](../ordertrigger/README.md)
Cloud Function is invoked. This reads the document from Firestore and publishes it to an "ecomm-order" Pub/Sub topic.

**NOTE:** While it is true that the complete order data is present in the Cloud Function invocation parameters, the
structure of those parameters is radically different from that of the native order document. While it would be possible
to extract the order data from those parameters this is both painful to do and would result in there being two places 
to maintain in the code, i.e. it would introduce the chance of inconsistent implementation and bugs. Hence, the 
trigger function uses the Order Service code to read the order from Firestore after getting just the order ID from
the trigger data.

### Fulfilling the Order

The order Pub/Sub topic is configured to push order messages to the [Order To Fulfillment Topic Consumer](../ordertofulfill/README.md)
Cloud Function. This function examines the order and generates zero to many fulfillment tasks for each order item,
writing task description documents to a Firestore "tasks" document collection.

Following the now predictable pattern, a [Task Firestore Trigger](../tasktrigger/README.md) is invoked for each of
the tasks written to Firestore. This reads the tasks from Firestore and publishes them to an "ecomm-task" Pub/Sub topic.

Receiving pushes from this task topic is a [Task Distributor](..taskdistrib/README.md) Cloud Function. This function 
knows nothing about how to execute the task but is configured with a map that informs it of additional Cloud Functions
that it can synchronously invoke to handle task initiation (e.g. send an email etc).
