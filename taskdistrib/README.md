# The Fulfillment Task Distribution Function

The **Fulfillment Task Distribution Function** is a Cloud Function that receives Pub/Sub push messages
from the [Fulfillment Task Firestore Trigger Function](../tasktrigger/README.md) when fulfillment task
is recorded or updated.

This function maps the task description to a task (and status) specific Cloud Function and invokes that
as function as a [CloudEvents](https://cloudevents.io/) event handler. In other words, fulfillment task
task distribution is a source of CloudEvents; if and when Google gets around to properly implementing
EventArc as a true event bus for custom events that does not require two PubSub topics for every event 
type, we would be able to ditch the distributor and simply configure EventArc rules much as we might with
AWS EventBridge.

In a real world implementation, the mapping of task type and status to task execution function 
would be soft configured and loaded when **Fulfillment Task Distribution Function** was instantiated, perhaps
checked every 15 minutes, but for this POC the mapping is hard coded in the [`handler.go` source file](./handler.go).

## Unit Testing

Since the Task Distribution function  includes code to obtain a OAuth token, and CloudEvents requires that this token 
must be either for a service account or a service account impersonator, a couple of hoops have to be jumped in order 
for unit tests to be run successfully.

See [Allow Cloud Functions to Invoke Cloud Functions](../docs/CF_INVOKE_CF.md) for instructions on how to set this up.

