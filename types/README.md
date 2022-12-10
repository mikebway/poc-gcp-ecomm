# Common Types

A number of basic type structures are shared across the `poc-gcp-ecomm` project; things like `Money` and `Person`, etc.
All of those are defined here.

This module contains very little code, typically only that required to convert to and from these internal structures
and their Protocol Buffer equivalents.

Why do we need two representations of essentially the same things? The Protocol Buffer structures are generated while
these internal structures are annotated for saving and loading into and out of Firestore.
  