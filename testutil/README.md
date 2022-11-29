# Unit Test Utility Functions

This module is used solely to support the unit testing of other modules. As such, this module lacks unit testing
of its own - the he code here is exercised by the unit tests of other modules.

## Functions

### `CaptureLogging(f func())`

**`CaptureLogging`** allows unit tests to override the default Zap logger to capture the logging output that results
from execution of the supplied function. The captured log output is returned as a string after the default
logger has been restored.

The supplied function parameter would typically be an inline function supplied by a unit test that needs to
evaluate the log output of some test subject to determine if the test passed or failed.