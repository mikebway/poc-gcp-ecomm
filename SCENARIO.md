# Proof of Concept Scenario

![POC Component Relationships](docs/poc-gcp-ecomm-scenario.png)

The scenario is of a shopping cart and order fulfillment system. The modeling is crude, lacking a great deal that
a real-world implementation would necessarily include - not least, there is no payment step in the checkout 
process!

The purpose of the project is to explore and demonstrate the use of several capabilities of the Goggle Cloud 
Platform including:

* Deploying gRPC services with Cloud Run
  * [The gRPC Cart Microservice](cart/README.md)
  * [The gRPC Order Microservice](order/README.md)
  * [The gRPC Fulfillment Orchestration Microservice](fulfillment/README.md)

* Persisting hierarchical document tree data structures with Firestore
  * [The gRPC Cart Microservice](cart/README.md)

* Having Firestore Change Data Capture (CDC) events trigger Cloud Functions
  * [The Cart Firestore Trigger Function](carttrigger/README.md)
  * [The Order Firestore Trigger Function](ordertrigger/README.md)
  * [The Fulfillment Task Firestore Trigger Function](tasktrigger/README.md)

* Asynchronous messaging via Pub/Sub to transform checked out shopping carts to immutable orders
  * [The Cart Firestore Trigger Function](carttrigger/README.md)
  * [The Order from Cart Pub/Sub Consumer](orderfromcart/README.md)
  * [The Fulfillment Task Distribution Function](taskdistrib/README.md)

* Implementation of a fulfillment service as a state machine events triggering EventArc rule driven business 
  logic Cloud Functions 
  * [The Fulfillment Task Distribution Function](taskdistrib/README.md)
  * [The Fulfillment Task Email Function](taskdistrib/README.md)

