package schema

import (
	"github.com/mikebway/poc-gcp-ecomm/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	// A couple of timestamp strings we can use to derive known time values
	earlyTimeString = "2022-10-29T16:23:19.123456789-06:00"
	lateTimeString  = "2022-10-30T09:28:42.987654321-06:00"

	// A UUID string value that we can use as a task ID in our tests
	taskId = "41168fa7-ff28-42db-af6b-5542cb235a55"

	// A UUID string value that we can use as an order ID in our tests
	orderId = "d1cecab3-5bc0-43d4-aef1-99ad69794313"

	// A UUID string value that we can use as an order item ID in our tests
	itemId = "54f34cb9-fea6-4786-a475-cebd95d93742"

	// Product, task, and reason codes that we use in our tests
	productCode = "B09V3G22BH"
	taskCode    = "assemble"
	reasonCode1 = "dye_shop"
	reasonCode2 = "waiting_for_parts"
	reasonCode3 = "customer_cancelled"

	// Named parameter values
	paramName1   = "string_1"
	paramName2   = "string_2"
	valueString1 = "first string"
	valueString2 = "second fiddle"
)

var (
	// Shopping cart value types that cannot be declared as constants
	taskSubmissionTime time.Time
	taskCompletionTime time.Time
)

// init is used to initialize the task "constant" values that cannot be declared as literal constants.
func init() {

	// Shopping cart values
	t, _ := types.TimestampFromRFC3339Nano(earlyTimeString)
	taskSubmissionTime = t.GetTime()
	t, _ = types.TimestampFromRFC3339Nano(lateTimeString)
	taskCompletionTime = t.GetTime()
}

// TestAsPBTask exercises the ability to render our internal Task form, as written to Firestore, into
// its protocol buffer form.
func TestAsPBTask(t *testing.T) {

	// Build a task populated with all the bells and whistles
	task := buildMockTask()

	// Ask the task for its protocol buffer doppelganger
	pbTask := task.AsPBTask()

	// Check that everything came back as we expected in our freshly minted twin
	req := require.New(t)
	req.NotNil(pbTask, "should have received a result from Task.AsPBTask()")
	req.Equal(taskId, pbTask.Id, "expect task IDs to match")
	req.Equal(taskSubmissionTime, pbTask.SubmissionTime.AsTime(), "expect task submission times to match")
	req.Equal(taskCompletionTime, pbTask.CompletionTime.AsTime(), "expect task completion times to match")
	req.Equal(orderId, pbTask.OrderId, "expect task order IDs to match")
	req.Equal(itemId, pbTask.OrderItemId, "expect order item IDs to match")
	req.Equal(productCode, pbTask.ProductCode, "expect task IDs to match")
	req.Equal(taskCode, pbTask.TaskCode, "expect task task codes to match")
	req.Equal(int32(WAITING_SERVICE), int32(pbTask.Status), "expect task stays to match")
	req.Equal(reasonCode1, pbTask.ReasonCode, "expect task reason codes to match")

	// Confirm that one of the expected parameter count made it across
	req.Equal(2, len(pbTask.Parameters), "unexpected count of converted task parameters")
	req.Equal(paramName1, pbTask.Parameters[0].Name, "parameter name 0 does not match")
	req.Equal(valueString1, pbTask.Parameters[0].Value, "parameter value 0 does not match")
	req.Equal(paramName2, pbTask.Parameters[1].Name, "parameter name 1 does not match")
	req.Equal(valueString2, pbTask.Parameters[1].Value, "parameter value 1 does not match")
}

// buildMockTask assembles a task structure with everything populated with known values.
func buildMockTask() *Task {
	return &Task{
		Id:             taskId,
		SubmissionTime: taskSubmissionTime,
		CompletionTime: taskCompletionTime,
		OrderId:        orderId,
		OrderItemId:    itemId,
		ProductCode:    productCode,
		TaskCode:       taskCode,
		Status:         WAITING_SERVICE,
		ReasonCode:     reasonCode1,
		Parameters: []*Parameter{
			&Parameter{Name: paramName1, Value: valueString1},
			&Parameter{Name: paramName2, Value: valueString2},
		},
	}
}
