package schema

import (
	pbfulfillment "github.com/mikebway/poc-gcp-ecomm/pb/fulfillment"
	"github.com/mikebway/poc-gcp-ecomm/testutil"
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
	nameInt1           = "int_1"
	nameInt2           = "int_2"
	nameString1        = "string_1"
	nameString2        = "string_2"
	nameBool1          = "bool_1"
	nameBool2          = "bool_2"
	nameDuff           = "duff"
	valueInt1    int32 = 1111
	valueInt2    int32 = 2222
	valueString1       = "first string"
	valueString2       = "second fiddle"
	valueBool1         = true
	valueBool2         = false
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

	// Ask the task for its protocol buffer doppelganger, but capture the logging while we do it
	// since one of the parameter values in our task won't be translatable
	var pbTask *pbfulfillment.Task
	logged := testutil.CaptureLogging(func() {
		pbTask = task.AsPBTask()
	})

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

	// Confirm that one of the original parameters did not make it across
	req.Equal(7, len(task.Parameters), "unexpected count of source task parameters")
	req.Equal(6, len(pbTask.Parameters), "unexpected count of converted task parameters")
	req.Contains(logged, "ignoring unrecognized parameter type in task", "did not see the expected error log entry")

	// ... and that all the expected ones made it across correctly
	req.Equal(nameInt1, pbTask.Parameters[0].Name, "parameter name 0 does not match")
	req.Equal(nameInt2, pbTask.Parameters[1].Name, "parameter name 1 does not match")
	req.Equal(nameString1, pbTask.Parameters[2].Name, "parameter name 2 does not match")
	req.Equal(nameString2, pbTask.Parameters[3].Name, "parameter name 3 does not match")
	req.Equal(nameBool1, pbTask.Parameters[4].Name, "parameter name 4 does not match")
	req.Equal(nameBool2, pbTask.Parameters[5].Name, "parameter name 5 does not match")

	number, ok := pbTask.Parameters[0].Value.(*pbfulfillment.Parameter_Number)
	req.True(ok, "parameter value 0 is not the expected type")
	req.Equal(valueInt1, number.Number, "parameter value 0 does not match")
	number, ok = pbTask.Parameters[1].Value.(*pbfulfillment.Parameter_Number)
	req.True(ok, "parameter value 1 is not the expected type")
	req.Equal(valueInt2, number.Number, "parameter value 1 does not match")

	txt, ok := pbTask.Parameters[2].Value.(*pbfulfillment.Parameter_Text)
	req.True(ok, "parameter value 2 is not the expected type")
	req.Equal(valueString1, txt.Text, "parameter value 2 does not match")
	txt, ok = pbTask.Parameters[3].Value.(*pbfulfillment.Parameter_Text)
	req.True(ok, "parameter value 3 is not the expected type")
	req.Equal(valueString2, txt.Text, "parameter value 3 does not match")

	tf, ok := pbTask.Parameters[4].Value.(*pbfulfillment.Parameter_TrueFalse)
	req.True(ok, "parameter value 4 is not the expected type")
	req.Equal(valueBool1, tf.TrueFalse, "parameter value 2 does not match")
	tf, ok = pbTask.Parameters[5].Value.(*pbfulfillment.Parameter_TrueFalse)
	req.True(ok, "parameter value 5 is not the expected type")
	req.Equal(valueBool2, tf.TrueFalse, "parameter value 3 does not match")
}

// TestAsParameterValue goes for maximum coverage by confirming tha all the Parameter.AsParameterValue interface
// marker functions return the string value.
func TestAsParameterValue(t *testing.T) {

	// Try each ParameterValue type in turn
	req := require.New(t)
	number := NumberValue{Number: valueInt1}
	req.Equal("1111", number.AsParameterValue(), "number AsParameterValue did not match")
	txt := TextValue{Text: valueString1}
	req.Equal(valueString1, txt.AsParameterValue(), "text AsParameterValue did not match")
	tf := BooleanValue{Boolean: true}
	req.Equal("true", tf.AsParameterValue(), "boolean true AsParameterValue did not match")
	tf = BooleanValue{Boolean: false}
	req.Equal("false", tf.AsParameterValue(), "boolean false AsParameterValue did not match")
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
			&Parameter{Name: nameInt1, Value: &NumberValue{Number: valueInt1}},
			&Parameter{Name: nameInt2, Value: &NumberValue{Number: valueInt2}},
			&Parameter{Name: nameString1, Value: &TextValue{Text: valueString1}},
			&Parameter{Name: nameString2, Value: &TextValue{Text: valueString2}},
			&Parameter{Name: nameBool1, Value: &BooleanValue{Boolean: valueBool1}},
			&Parameter{Name: nameBool2, Value: &BooleanValue{Boolean: valueBool2}},
			&Parameter{Name: nameDuff, Value: &DuffValue{}},
		},
	}
}

// DuffValue implements the ParameterValue interface to represent a type that Task.asPBParameters cannot
// recognize in order to exercise that corner case circumstance
type DuffValue struct {
	ParameterValue

	Nil interface{}
}

// IsParameterValue confirms by its existence that DuffValue is a ParameterValue implementation type.
func (nv *DuffValue) IsParameterValue() {}
