package types

import (
	"errors"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

// Timestamp defines the interface for a time structure capable of capturing a time as either
// an RFC 3339 UTC string or a google.protobuf.Timestamp protocol buffer value.
type Timestamp interface {

	// GetTime returns the underlying time.Time value.
	GetTime() time.Time

	// GetPBTimestamp returns the time value as a google.protobuf.Timestamp protocol
	// buffer value.
	GetPBTimestamp() *timestamppb.Timestamp

	// String returns the time value as an RFC 3339 UTC string value to nanosecond accuracy.
	String() string
}

// Timestamp is a concrete implementation of the Timestamp interface
type timestamp struct {
	Timestamp

	// time is the standard Go library representation of a time value
	time time.Time
}

// NewTimestamp is a factory method returning a Timestamp value reference for a given time.
func NewTimestamp(t time.Time) Timestamp {
	return &timestamp{
		time: t.In(time.UTC),
	}
}

// TimestampNow is a factory method returning a Timestamp value reference to the time right now.
func TimestampNow() Timestamp {
	return &timestamp{
		time: time.Now().In(time.UTC),
	}
}

// TimestampFromRFC3339Nano is a factory method returning a Timestamp value reference set
// to parsed value of an UTC RFC 3339 string accurate to a nanosecond. If the supplied time
// cannot be parsed, an error is returned.
func TimestampFromRFC3339Nano(timeStr string) (Timestamp, error) {
	t, err := time.Parse(time.RFC3339Nano, timeStr)
	if err != nil {
		return nil, err
	} else {
		return &timestamp{
			time: t.UTC(),
		}, nil
	}
}

// TimestampFromPBTimestamp is a factory method returning a Timestamp value reference set
// to translation of a Google protocol buffer timestamp value. If the value cannot be
// translated (i.e. if the input is not a valid time representation) then an error
// is returned.
func TimestampFromPBTimestamp(timepb *timestamppb.Timestamp) (Timestamp, error) {
	if timepb == nil {
		return nil, errors.New("cannot interpret nil protobuf timestamp")
	}
	return &timestamp{
		time: timepb.AsTime(),
	}, nil
}

// GetTime returns the timestamp as a Go time.Time value.
func (t *timestamp) GetTime() time.Time {
	return t.time
}

// String returns the time value as an RFC 3339 UTC string value to nanosecond accuracy.
func (t *timestamp) String() string {
	return t.time.Format(time.RFC3339Nano)
}
