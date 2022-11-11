package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"net"
	"os"
	"testing"
)

// resetEnvironment restores environment variables and the like to their default state for testing
// the initializeService function and anything related.
func resetEnvironment() {

	// Clear the gRPC port number environment variable
	_ = os.Setenv(EnvGRPCPort, "")

	// Clear the request for the NewCartService to return a mock error
	unitTestNewCartServiceError = nil
}

// TestMainFailure is the only test we can run against the main() function as we deliberately force a failure
// to initialise the cart service and thereby avoid having main() start the gRPC server and never return.
func TestMainFailure(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Start with a clean slate and leave it that way too
	resetEnvironment()
	defer resetEnvironment()

	// Have the NewCartService call return an error
	const errorMsg = "TestMainFailure mock error"
	unitTestNewCartServiceError = fmt.Errorf(errorMsg)

	// Wrap a call to main() to capture its log output
	logged := CaptureLogging(func() {
		main()
	})

	// Confirm that the error we set to be returned was logged by main()
	req.Contains(logged, "poc-cart-service: starting server", "should have seen server starting message in log")
	req.Contains(logged, "{\"port\": \"8080\"}", "should have seen default port number in log")
	req.Contains(logged, "poc-cart-service: failed to start", "should have seen server failed message in log")
	req.Contains(logged, errorMsg, "should have seen our mock error message in log")
	req.Contains(logged, "poc-cart-service: failed to start\t{\"error\": \"failed to initialize the CartService", "should have seen the cart service error we forced in the final log entry")
}

// TestDefaultInitialization examines the most basic function of the initializeService with default configuration
// and no errors.
func TestDefaultInitialization(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Start with a clean slate and leave it that way too
	resetEnvironment()
	defer resetEnvironment()

	// Initialize the service while capture it's log output
	var service *grpc.Server
	var listener net.Listener
	var err error
	logged := CaptureLogging(func() {
		service, listener, err = initializeService()
	})

	// If a service was returned, stop it immediately
	if service != nil {
		service.Stop()
	}

	// If a listener was returned, stop that too
	if listener != nil {
		_ = listener.Close()
	}

	// Now, see whether we like what happened
	req.Nil(err, "should have successfully initialized the gRPC service but got an error: %v", err)
	req.Contains(logged, "poc-cart-service: starting server", "should have seen server starting message in log")
	req.Contains(logged, "{\"port\": \"8080\"}", "should have seen default port number in log")
}

// TestCustomPortInitialization examines the handling of a custom TCP port configuration
func TestCustomPortInitialization(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Start with a clean slate and leave it that way too
	resetEnvironment()
	defer resetEnvironment()

	// Configure a non-standard TCP port number
	_ = os.Setenv(EnvGRPCPort, "12345")

	// Initialize the service while capture it's log output
	var service *grpc.Server
	var listener net.Listener
	var err error
	logged := CaptureLogging(func() {
		service, listener, err = initializeService()
	})

	// If a service was returned, stop it immediately
	if service != nil {
		service.Stop()
	}

	// If a listener was returned, stop that too
	if listener != nil {
		_ = listener.Close()
	}

	// Now, see whether we like what happened
	req.Nil(err, "should have successfully initialized the gRPC service but got an error: %v", err)
	req.Contains(logged, "poc-cart-service: starting server", "should have seen server starting message in log")
	req.Contains(logged, "{\"port\": \"12345\"}", "should have seen a non-default port number in log")
}

// TestInvalidPortInitialization examines the handling of an invalid custom TCP port configuration that leads
// to a TCP listen failure
func TestInvalidPortInitialization(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Start with a clean slate and leave it that way too
	resetEnvironment()
	defer resetEnvironment()

	// Configure a non-standard TCP port number
	_ = os.Setenv(EnvGRPCPort, "Gandalf")

	// Initialize the service while capture it's log output
	var service *grpc.Server
	var listener net.Listener
	var err error
	logged := CaptureLogging(func() {
		service, listener, err = initializeService()
	})

	// If a service was returned, stop it immediately
	if service != nil {
		service.Stop()
	}

	// If a listener was returned, stop that too
	if listener != nil {
		_ = listener.Close()
	}

	// Now, see whether we like what happened
	req.NotNil(err, "should have failed initialized the gRPC service")
	req.Contains(logged, "poc-cart-service: starting server", "should have seen server starting message in log")
	req.Contains(logged, "{\"port\": \"Gandalf\"}", "should have seen a non-default port number in log")
	req.Contains(logged, "net.Listen error", "should have seen an error reported about net.Listen failing in log")
	req.Nil(listener, "no listener should have been returned")
	req.Nil(service, "no gRPC service should have been returned")
}

// TestNoCartServiceInitialization examines the handling of a failure in the NewCartService call.
func TestNoCartServiceInitialization(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Start with a clean slate and leave it that way too
	resetEnvironment()
	defer resetEnvironment()

	// Have the NewCartService call return an error
	const errorMsg = "TestNoCartServiceInitialization mock error"
	unitTestNewCartServiceError = fmt.Errorf(errorMsg)

	// Initialize the service while capture it's log output
	var service *grpc.Server
	var listener net.Listener
	var err error
	logged := CaptureLogging(func() {
		service, listener, err = initializeService()
	})

	// If a service was returned, stop it immediately
	if service != nil {
		service.Stop()
	}

	// If a listener was returned, stop that too
	if listener != nil {
		_ = listener.Close()
	}

	// Now, see whether we like what happened
	req.NotNil(err, "should have failed initialized the gRPC service")
	req.Contains(logged, "poc-cart-service: starting server", "should have seen server starting message in log")
	req.Contains(logged, "{\"port\": \"8080\"}", "should have seen a default port number in log")
	req.Contains(logged, "NewCartService error", "should have seen an error reported about NewCartService failing in log")
	req.Contains(logged, errorMsg, "should have seen our mock error message in log")
	req.NotNil(listener, "listener should have been returned")
	req.Nil(service, "no gRPC service should have been returned")
}

// CaptureLogging allows unit tests to override the default Zap logger to capture the logging output that results
// from execution of the supplied function. The captured log output is returned as a string after the default
// logger has been restored.
//
// The supplied function parameter would typically be an inline function supplied by a unit test that needs to
// evaluate the log output of some test subject to determine if the test passed or failed.
func CaptureLogging(f func()) string {

	// Configure a Zap logger to record to a buffer
	var loggedBytes bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	writer := bufio.NewWriter(&loggedBytes)
	capLogger := zap.New(zapcore.NewCore(encoder, zapcore.AddSync(writer), zapcore.DebugLevel))

	// Set our capturing logger as the default logger, deferring the returned function to restore
	// the original logger when this function exits
	restoreOriginalLogger := zap.ReplaceGlobals(capLogger)
	defer restoreOriginalLogger()

	// Call the supplied function with our logger recording hat it has to say to teh world
	f()

	// Flush the log then return the buffer contents as string
	err := writer.Flush()
	if err != nil {
		return fmt.Sprintf("FAILED TO FLUSH THE CAPTURED LOG: %v", err)
	}
	return loggedBytes.String()
}
