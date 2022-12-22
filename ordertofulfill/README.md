# The Order to Fulfillment Topic Consumer

The **Order to Fulfillment Firestore Consumer** is a Cloud Function that receives Pub/Sub push messages 
from the [Order Firestore Trigger Function](../ordertrigger/README.md) when an order is recorded.

This function translates each of the order items into one to many fulfillment tasks, storing the tasks in  
an `tasks` Firestore document collection (i.e. a different collection to that used for the carts and orders).