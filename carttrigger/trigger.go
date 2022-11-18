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

// FirestoreValue holds Firestore fields.
type FirestoreValue struct {
	CreateTime time.Time   `json:"createTime"`
	Fields     interface{} `json:"fields"`
	Name       string      `json:"name"`
	UpdateTime time.Time   `json:"updateTime"`
}

// init is the static initializer used to configure our local and global static variables.
func init() {
	serviceLogger, _ := zap.NewProduction()
	zap.ReplaceGlobals(serviceLogger)
}

// UpdateTrigger receives a document update Firestore trigger event. The function is deployed with a trigger
// configuration (see Makefile) that will notify the handler of all updates to the root document of a Shopping Cart.
func UpdateTrigger(ctx context.Context, e FirestoreEvent) error {

	jsonBytes, err := json.Marshal(e)
	if err != nil {
		zap.L().Error("failed to marshal firestore event to JSON", zap.Error(err))
		return err
	}

	// log the event
	zap.L().Info("firestore event", zap.String("event", string(jsonBytes)))

	//// For now, we just log enough to prove that we got here
	//zap.L().Info("cart document updated",
	//	zap.String("name", e.Value.Name),
	//	zap.String("id", e.Value.Fields.Id),
	//	zap.Int32("status", int32(e.Value.Fields.Status)))

	// She did not write much!
	return nil
}
