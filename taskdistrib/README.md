# The Fulfillment Task Distribution Function

The **Fulfillment Task Distribution Function** is a Cloud Function that receives Pub/Sub push messages
from the [Fulfillment Task Firestore Trigger Function](../tasktrigger/README.md) when fulfillment task
is recorded or updated.

This function maps the task description to a task (and status) specific Cloud Function and invokes that
function. 

In a real world implementation, the mapping of task type ans status to task execution function 
would be soft configured and loaded when **Fulfillment Task Distribution Function** was instantiated, perhaps
checked every 15 minutes, but for this POC the mapping is hard coded in the [`handler.go` source file](./handler.go).
