// Package carttrigger handles Firestore trigger invocations when shopping cart documents are updated.
//
// The handler is not invoked for the addition of cart items or delivery addresses, nor for creation of carts,
// only for updates to the cart root document. This will almost invariably be due to the cart either being
// submitted or abandoned.
package carttrigger

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"time"
)

// FirestoreEvent is the payload of a Firestore event.
type FirestoreEvent struct {
	OldValue   FirestoreValue `json:"oldValue"`
	Value      FirestoreValue `json:"value"`
	UpdateMask struct {
		FieldPaths []string `json:"fieldPaths"`
	} `json:"updateMask"`
}

// FirestoreValue holds Firestore document fields.
type FirestoreValue struct {
	CreateTime time.Time     `json:"createTime"`
	Fields     FirestoreCart `json:"fields"`
	Name       string        `json:"name"`
	UpdateTime time.Time     `json:"updateTime"`
}

// FirestoreCart describes the document fields that we need to know about as they will
// be found in the event data (not as we would prefer them, in the structure that we
// submitted to the Firestore API to populate the document in the first place :-(
type FirestoreCart struct {
	Id     StringValue  `json:"id"`
	Status IntegerValue `json:"status"`
}
type StringValue struct {
	StringValue string `json:"stringValue"`
}
type IntegerValue struct {
	IntegerValue string `json:"integerValue"`
}

// init is the static initializer used to configure our local and global static variables.
func init() {
	serviceLogger, _ := zap.NewProduction()
	zap.ReplaceGlobals(serviceLogger)
}

// UpdateTrigger receives a document update Firestore trigger event. The function is deployed with a trigger
// configuration (see Makefile) that will notify the handler of all updates to the root document of a Shopping Cart.
func UpdateTrigger(ctx context.Context, e FirestoreEvent) error {

	// TODO: Google and AWS have this in common: they fail to make their CDC event stream contents
	//       compatible with or easily convertible to their database API models. There is no easy
	//       way to populate a ShoppingCart structure from the FirestoreEvent/FirestoreValue structures!

	// Fortunately, we need little information from the new FirestoreValue structure since Google
	// makes it harder than it should be to interpret.
	// There should be a way to unmarshall this FirestoreEvent data to a ShoppingCart structures but there
	// is not. Fortunately, we need little information from the new FirestoreValue structure to determine
	// how we should respond.
	jsonBytes, err := json.Marshal(e)
	if err != nil {
		zap.L().Error("failed to marshal firestore event to JSON", zap.Error(err))
		return err
	}

	// log the event etc
	zap.L().Info("firestore event", zap.String("event", string(jsonBytes)))
	zap.L().Info("cart",
		zap.String("id", e.Value.Fields.Id.StringValue),
		zap.String("status", e.Value.Fields.Status.IntegerValue))

	// She did not write much!
	return nil
}
