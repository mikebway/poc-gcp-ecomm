# The Fulfillment Task Email Function

The **Fulfillment Task Email Function** is a [CloudEvents](https://cloudevents.io/) Cloud Function that generically 
processes the fulfillment tasks that it receives via direct invocation from the [Task Distribution Function](../taskdistrib/README.md)
and sends an email summarising the task. 

The email is sent using [SendGrid](https://sendgrid.com/solutions/email-api/) to an address configured via environment 
variable.