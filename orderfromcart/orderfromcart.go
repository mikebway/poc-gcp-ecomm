// Package orderfromcart implements a Google Cloud Function to receive a shopping cart description via
// a Cloud Task queue. The handler translates the cart into an order and stores it as an order document
// under Firestore.
package orderfromcart

import (
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
)

// init is the static initializer used to configure our local and global static variables.
func init() {
	serviceLogger, _ := zap.NewProduction()
	zap.ReplaceGlobals(serviceLogger)
}

// OrderFromCart is Cloud Function the entry point. The payload of the HTTP request is a checked out shopping cart
// expressed as a JSON string.
func OrderFromCart(w http.ResponseWriter, r *http.Request) {

	// Get the body of the request as a byte array
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("unable to read request body: %w", err)
		zap.L().Error("unable to read request body", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// At this point we are just proving that our Cloud Task queue configuration and task submission worked.
	// We do that by crudely logging the payload.
	zap.L().Info("received shopping cart", zap.ByteString("body", bodyBytes))
}
