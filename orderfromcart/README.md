# The Order from Cart Topic Consumer

The **Order from Cart Firestore Consumer** is a Cloud Function that receives Pub/Sub push messages 
from the [Cart Firestore Trigger Function](../carttrigger/README.md) when a shopping cart is 
"checked out."

This function converts completed shopping cart descriptions into order descriptions  and stores them in  
an `orders` Firestore document collection (i.e. a different collection to that used for the carts).