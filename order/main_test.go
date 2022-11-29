package main

import (
	"fmt"
	"github.com/mikebway/poc-gcp-ecomm/order/service"
	"github.com/mikebway/poc-gcp-ecomm/testutil"
	"github.com/stretchr/testify/require"
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

	// Clear the request for the NewOrderService to return a mock error
	service.UnitTestNewOrderServiceError = nil
}

// TestMainFailure is the only test we can run against the main() function as we deliberately force a failure
// to initialise the order service and thereby avoid having main() start the gRPC server and never return.
func TestMainFailure(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Start with a clean slate and leave it that way too
	resetEnvironment()
	defer resetEnvironment()

	// Have the NewOrderService call return an error
	const errorMsg = "TestMainFailure mock error"
	service.UnitTestNewOrderServiceError = fmt.Errorf(errorMsg)

	// Wrap a call to main() to capture its log output
	logged := testutil.CaptureLogging(func() {
		main()
	})

	// Confirm that the error we set to be returned was logged by main()
	req.Contains(logged, "poc-order-service: starting server", "should have seen server starting message in log")
	req.Contains(logged, "{\"port\": \"8080\"}", "should have seen default port number in log")
	req.Contains(logged, "poc-order-service: failed to start", "should have seen server failed message in log")
	req.Contains(logged, errorMsg, "should have seen our mock error message in log")
	req.Contains(logged, "poc-order-service: failed to start\t{\"error\": \"failed to initialize the OrderService", "should have seen the order service error we forced in the final log entry")
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
	var svc *grpc.Server
	var listener net.Listener
	var err error
	logged := testutil.CaptureLogging(func() {
		svc, listener, err = initializeService()
	})

	// If a service was returned, stop it immediately
	if svc != nil {
		svc.Stop()
	}

	// If a listener was returned, stop that too
	if listener != nil {
		_ = listener.Close()
	}

	// Now, see whether we like what happened
	req.Nil(err, "should have successfully initialized the gRPC service but got an error: %v", err)
	req.Contains(logged, "poc-order-service: starting server", "should have seen server starting message in log")
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
	var svc *grpc.Server
	var listener net.Listener
	var err error
	logged := testutil.CaptureLogging(func() {
		svc, listener, err = initializeService()
	})

	// If a service was returned, stop it immediately
	if svc != nil {
		svc.Stop()
	}

	// If a listener was returned, stop that too
	if listener != nil {
		_ = listener.Close()
	}

	// Now, see whether we like what happened
	req.Nil(err, "should have successfully initialized the gRPC service but got an error: %v", err)
	req.Contains(logged, "poc-order-service: starting server", "should have seen server starting message in log")
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

	// Initialize the service while capturing its log output
	var svc *grpc.Server
	var listener net.Listener
	var err error
	logged := testutil.CaptureLogging(func() {
		svc, listener, err = initializeService()
	})

	// If a service was returned, stop it immediately
	if svc != nil {
		svc.Stop()
	}

	// If a listener was returned, stop that too
	if listener != nil {
		_ = listener.Close()
	}

	// Now, see whether we like what happened
	req.NotNil(err, "should have failed initialized the gRPC service")
	req.Contains(logged, "poc-order-service: starting server", "should have seen server starting message in log")
	req.Contains(logged, "{\"port\": \"Gandalf\"}", "should have seen a non-default port number in log")
	req.Contains(logged, "net.Listen error", "should have seen an error reported about net.Listen failing in log")
	req.Nil(listener, "no listener should have been returned")
	req.Nil(svc, "no gRPC service should have been returned")
}

// TestNoOrderServiceInitialization examines the handling of a failure in the NewOrderService call.
func TestNoOrderServiceInitialization(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Start with a clean slate and leave it that way too
	resetEnvironment()
	defer resetEnvironment()

	// Have the NewOrderService call return an error
	const errorMsg = "TestNoOrderServiceInitialization mock error"
	service.UnitTestNewOrderServiceError = fmt.Errorf(errorMsg)

	// Initialize the service while capture it's log output
	var svc *grpc.Server
	var listener net.Listener
	var err error
	logged := testutil.CaptureLogging(func() {
		svc, listener, err = initializeService()
	})

	// If a service was returned, stop it immediately
	if svc != nil {
		svc.Stop()
	}

	// If a listener was returned, stop that too
	if listener != nil {
		_ = listener.Close()
	}

	// Now, see whether we like what happened
	req.NotNil(err, "should have failed initialized the gRPC service")
	req.Contains(logged, "poc-order-service: starting server", "should have seen server starting message in log")
	req.Contains(logged, "{\"port\": \"8080\"}", "should have seen a default port number in log")
	req.Contains(logged, "NewOrderService error", "should have seen an error reported about NewOrderService failing in log")
	req.Contains(logged, errorMsg, "should have seen our mock error message in log")
	req.NotNil(listener, "listener should have been returned")
	req.Nil(svc, "no gRPC service should have been returned")
}
