package types

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestNewTimestamp evaluates the NewTimestamp factory method when supplied with a known non-UTC time value
func TestNewTimestamp(t *testing.T) {

	// Avoid having to pass t in to every assertion
	req := require.New(t)

	// Construct a time known to be for the US central time zone
	centralTime, err := time.Parse(time.RFC3339, "2021-01-01T16:23:19-06:00")
	req.Nil(err, "should have been able to parse our seed time of 2021-01-01T16:23:19-06:00")

	// Just to prove to Mike that he is not crazy bothering with this, render that central time value as a string and confirm
	// that it does not come back as UTC.
	req.Equal("2021-01-01T16:23:19-06:00", centralTime.Format(time.RFC3339), "did not get the expected central time string value")

	// Use the target factory method to obtain a Timestamp object, then confirm that it renders the expect UTC string value
	utc := NewTimestamp(centralTime)
	req.Equal("2021-01-01T22:23:19Z", utc.String(), "did not get the expected UTC string value")

	// While we are here, confirm that the GetTime class method works as we expect - our two time values should be the same in UNIX form
	req.Equal(centralTime.Unix(), utc.GetTime().Unix(), "GetTime did not return the expected value")
}

// TestTimestampNow evaluates the TimestampNow factory method
func TestTimestampNow(t *testing.T) {

	// The raw time now ...
	startTime := time.Now()

	// ... should be less than or equal to the Timestamp now ...
	utc := TimestampNow()

	// ... which should be less than or equal to the raw time now
	endTime := time.Now()

	// Confirm our beliefs
	req := require.New(t)
	req.LessOrEqualf(startTime.Unix(), utc.GetTime().Unix(), "start time should have been less than or equal to our UTC time")
	req.LessOrEqualf(utc.GetTime().Unix(), endTime.Unix(), "UTC time should have been less than or equal to our end time")
}

// TestGetPBTimestamp evaluates the GetPBTimestamp conversion method
func TestGetPBTimestamp(t *testing.T) {

	// The raw time now ...
	startTime := time.Now()

	// .Form a timestamp from that
	timestamp := NewTimestamp(startTime)

	// Ask for that in Protocol Buffer form
	pbTimestamp := timestamp.GetPBTimestamp()

	// Confirm that this has the same core value as our original time
	require.Equal(t, startTime.Unix(), pbTimestamp.AsTime().Unix(), "protocol buffer time value does not match the original seed")
}

// TestTimestampFromRFC3339 evaluates the TimestampFromRFC3339 factory method when supplied with a known non-UTC time value.
func TestTimestampFromRFC3339Nano(t *testing.T) {

	// Use the target factory method to obtain a Timestamp object, then confirm that it renders the expect UTC string value
	utc, err := TimestampFromRFC3339Nano("2021-01-01T16:23:19.123456789-06:00")
	req := require.New(t)
	req.Nil(err, "should have been able to parse our seed time of 2021-01-01T16:23:19.123456789-06:00")
	req.Equal("2021-01-01T22:23:19.123456789Z", utc.String(), "did not get the expected UTC string value")
}

// TestBadTimestampFromRFC3339 evaluates the TimestampFromRFC3339 factory method when supplied with an invalid time string.
func TestBadTimestampFromRFC3339Nano(t *testing.T) {

	// Use the target factory method to obtain a Timestamp objcet, then confirm that it renders the expect UTC string value
	utc, err := TimestampFromRFC3339Nano("there is no time like the present")
	req := require.New(t)
	req.NotNil(err, "should not have been able to parse our invalid seed time string")
	req.Nil(utc, "should not have received a Timestamp object")
}

// TestTimestampFromPBTimestamp evaluates the TimestampFromPBTimestamp factory method when supplied with a known non-UTC time value.
func TestTimestampFromPBTimestamp(t *testing.T) {

	// Construct a protocol buffer timestamp with a known value
	tme, _ := time.Parse(time.RFC3339Nano, "2021-01-01T16:23:19.123456789-06:00")
	timepb := timestamppb.New(tme)

	// Use the target factory method to obtain a Timestamp object, then confirm that it renders the expect UTC string value
	utc, err := TimestampFromPBTimestamp(timepb)
	req := require.New(t)
	req.Nil(err, "should have been able to translate our seed time of 2021-01-01T16:23:19.123456789-06:00")
	req.Equal("2021-01-01T22:23:19.123456789Z", utc.String(), "did not get the expected UTC string value")
}

// func TestTimestampFromNilPBTimestamp(t *testing.T) { evaluates the TimestampFromPBTimestamp factory method when
// supplied with a nil PB time value
func TestTimestampFromNilPBTimestamp(t *testing.T) {

	// Hit the target factory method with a nil input value
	utc, err := TimestampFromPBTimestamp(nil)
	req := require.New(t)
	req.NotNil(err, "should not have been able to translate a nil value")
	req.Contains(err.Error(), "cannot interpret nil protobuf timestamp", "did not get the expected error")
	req.Nil(utc, "should not have had a timestamp value returned")
}
