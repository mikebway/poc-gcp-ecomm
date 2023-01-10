package main

import (
	"fmt"
	"net"
	"os"

	"github.com/mikebway/poc-gcp-ecomm/cart/cartapi"

	"google.golang.org/grpc"

	pb "github.com/mikebway/poc-gcp-ecomm/pb/cart"
	"go.uber.org/zap"
)

const (
	// EnvGRPCPort names the environment variable that may be set to override the default port number
	// used by the gRPC service to listen for TCP connection requests.
	EnvGRPCPort = "PORT"

	// DefaultGRPCPort defines the default gRPC TCP port number as a string
	DefaultGRPCPort = "8080"
)

// init is the static initializer used to configure our local and global static variables.
func init() {
	serviceLogger, _ := zap.NewProduction()
	zap.ReplaceGlobals(serviceLogger)
}

// main is the entry point to start the Social Graph gRPC service
func main() {

	// Have our unit testable sibling do most of the work
	grpcServer, listener, err := initializeService()
	if err == nil {

		// Start the service
		err = grpcServer.Serve(listener)
	}

	// If there was an error, log it and let the server die when we return.
	//
	// There is no need to use Fatal as we will exit the program anyway and by avoiding Fatal
	// we can run some unit testing on this function.
	if err != nil {
		zap.L().Error("poc-cart-service: failed to start", zap.String("error", err.Error()))
	}
}

// initializeService has been extracted from the main function so that it can be unit tested
// without worrying about fatal errors crashing the test run and without starting the gRPC
// service such that the tests never get to finish.
func initializeService() (*grpc.Server, net.Listener, error) {

	port := os.Getenv(EnvGRPCPort)
	if port == "" {
		port = DefaultGRPCPort
	}
	zap.L().Info("poc-cart-service: starting server", zap.String("port", port))

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		zap.L().Error("net.Listen error", zap.String("error", err.Error()))
		return nil, nil, fmt.Errorf("net.Listen: %v", err)
	}

	// Initialize our shopping cart service
	service, err := cartapi.NewCartService()
	if err != nil {
		// Log our discomfort
		zap.L().Error("NewCartService error", zap.String("error", err.Error()))

		// Make sure the listener gets shut down so that if we are running unit tests they don't get
		// caught out by the port still being in use
		_ = listener.Close()

		// And tell the caller how we have let them down
		return nil, listener, fmt.Errorf("failed to initialize the CartService: %v", err)
	}

	// Initialize the gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterCartAPIServer(grpcServer, service)

	// All went well
	return grpcServer, listener, nil
}
