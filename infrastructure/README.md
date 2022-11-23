# Infrastructure Setup

So, that's a lie. The [Makefile](Makefile) in this subdirectory will not create all of the GCP infrastructure
that you need to run this POC: you have to create the project and you have to make the access control changes
to allow Cloud Build etc to do its stuff. All the [Makefile](Makefile) here helps with is the wiring between 
components:

* Sets up the Pub/Sub topics
* Sets up the Pub/Sub schemas
* ... more to come

Running the `make` without specifying a target recipe will display the following usage help:

```text
help                           List of available commands
setup                          Setup the Google Cloud infrastructure
teardown                       Tear down the Google Cloud infrastructure
#
# Running setup or teardown multiple times will not fail. If you add something
# to setup, running it will report errors for the components without aborting.
```