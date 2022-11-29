package testutil

import (
	"bufio"
	"bytes"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// CaptureLogging allows unit tests to override the default Zap logger to capture the logging output that results
// from execution of the supplied function. The captured log output is returned as a string after the default
// logger has been restored.
//
// The supplied function parameter would typically be an inline function supplied by a unit test that needs to
// evaluate the log output of some test subject to determine if the test passed or failed.
func CaptureLogging(f func()) string {

	// Configure a Zap logger to record to a buffer
	var loggedBytes bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	writer := bufio.NewWriter(&loggedBytes)
	capLogger := zap.New(zapcore.NewCore(encoder, zapcore.AddSync(writer), zapcore.DebugLevel))

	// Set our capturing logger as the default logger, deferring the returned function to restore
	// the original logger when this function exits
	restoreOriginalLogger := zap.ReplaceGlobals(capLogger)
	defer restoreOriginalLogger()

	// Call the supplied function with our logger recording hat it has to say to teh world
	f()

	// Flush the log then return the buffer contents as string
	err := writer.Flush()
	if err != nil {
		return fmt.Sprintf("FAILED TO FLUSH THE CAPTURED LOG: %v", err)
	}
	return loggedBytes.String()
}
