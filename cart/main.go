package main

import (
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"

	pb "github.com/mikebway/poc-gcp-ecomm/pb/cart"
	"go.uber.org/zap"
)

// isUnitTesting is set to true when running unit tests. This will override the initialization of the service
// in small ways, e.g. to use a mock datastore client instead of the real thing.
var isUnitTesting bool

// unitTestNewCartServiceError should be returned by NewCartService if we are running unit tests
// and unitTestNewCartServiceError is not nil.
var unitTestNewCartServiceError error

// main is the entry point to start the Social Graph gRPC service
func main() {

	// Have our unit testable sibling do all the work
	_, err := initializeService()
	if err != nil {
		zap.L().Fatal(err.Error())
	}
}

// initializeService has been extracted from the main function so that it can be unit tested
// without worrying about fatal errors crashing the test run and with the ability to stop
// the gRPC service so that the tests can actually finish!
func initializeService() (*grpc.Server, error) {

	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)
	lg := zap.L()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	lg.Info("poc-cart-service: starting server", zap.String("port", port))

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, fmt.Errorf("net.Listen: %v", err)
	}

	// Initialize our shopping cart service
	service, err := NewCartService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the CartService: %v", err)
	}

	// Initialize the gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterCartAPIServer(grpcServer, service)
	if err = grpcServer.Serve(listener); err != nil {
		return nil, fmt.Errorf("failed to start the gRPC service: %v", err)
	}

	// All went well
	return grpcServer, nil
}
