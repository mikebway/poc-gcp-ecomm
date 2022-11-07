package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"testing"

	"cloud.google.com/go/firestore"
)

// Acknowledgement to Simon Green, AKA "Captain Codeman", for the guide on configuring Go unit tests
// to run against the Google Cloud Firestore emulator.
//
// See https://www.captaincodeman.com/unit-testing-with-firestore-emulator-and-go

// FirestoreEmulatorHost defines the environment variable name that is used to convey that the Firestore emulator is
// running, should be used, and how to connect to it
const FirestoreEmulatorHost = "FIRESTORE_EMULATOR_HOST"

// newFirestoreTestClient instantiates an instance of the Firestore client connected to the Firestore emulator
func newFirestoreTestClient(ctx context.Context) *firestore.Client {
	client, err := firestore.NewClient(ctx, "test")
	if err != nil {
		log.Fatalf("firebase.NewClient err: %v", err)
	}

	return client
}

// TestMain runs as an envelope containing all the unit tests for this package. More specifically, it spins up the
// Firestore emulator for our tests to use.
func TestMain(m *testing.M) {
	// command to start firestore emulator
	cmd := exec.Command("gcloud", "emulators", "firestore", "start", "--host-port=localhost", "--quiet")

	// this makes it killable
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// we need to capture it's output to know when it's started
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	defer func(stderr io.ReadCloser) {
		_ = stderr.Close()
	}(stderr)

	// start her up!
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// ensure the process is killed when we're finished, even if an error occurs
	// (thanks to Brian Moran for suggestion)
	var result int
	defer func() {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		os.Exit(result)
	}()

	// we're going to wait until it's running to start
	var wg sync.WaitGroup
	wg.Add(1)

	// by starting a separate go routine
	go func() {
		// reading it's output
		buf := make([]byte, 256, 256)
		for {
			n, err := stderr.Read(buf[:])
			if err != nil {
				// until it ends
				if err == io.EOF {
					break
				}
				log.Fatalf("reading stderr %v", err)
			}

			if n > 0 {
				d := string(buf[:n])

				// only required if we want to see the emulator output
				log.Printf("%s", d)

				// checking for the message that it's started
				if strings.Contains(d, "Dev App Server is now running") {
					wg.Done()
				}

				// FIXME: I think there is a race condition here where we might start running tests before the
				//        FIRESTORE_EMULATOR_HOST environment variable has been set

				// and capturing the FIRESTORE_EMULATOR_HOST value to set
				pos := strings.Index(d, FirestoreEmulatorHost+"=")
				if pos > 0 {
					host := d[pos+len(FirestoreEmulatorHost)+1 : len(d)-1]
					_ = os.Setenv(FirestoreEmulatorHost, host)
				}
			}
		}
	}()

	// wait until the running message has been received
	wg.Wait()

	// now it's running, we can run our unit tests
	result = m.Run()
}
