# The Fulfillment Task Firestore Trigger Function

The **Fulfillment Task Firestore Trigger** is a Cloud Function that is invoked in response to Firestore document updates. 
Specifically, creates/updates of tasks documents written by the [gRPC fulfillment-service](../fulfillment/README.md).

The function publishes the tasks to a Pub/Sub topic, which in turn pushes to a [Task Distributor](..taskdistrib/README.md)
Cloud Function. 