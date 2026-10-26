package server

import (
	"bytes"
	"errors"
	"testing"
)

type failingSnapshotCloser struct {
	bytes.Buffer
	closeErr error
	closed   bool
}

func (w *failingSnapshotCloser) Close() error {
	w.closed = true
	return w.closeErr
}

// SnapshotResult.errCh must include errors raised while closing the archive,
// compressor, or transport. A successful Flush alone is not the final result.
func TestStreamSnapshotV2ReportsFinalTransportCloseFailure(t *testing.T) {
	failure := errors.New("simulated snapshot transport close failure")
	w := &failingSnapshotCloser{closeErr: failure}
	ch := make(chan error, 1)
	js := &jetStream{}
	js.streamSnapshotV2(nil, &StreamState{}, w, false, nil, ch)
	if !w.closed {
		t.Fatal("snapshot transport was not closed")
	}
	got, ok := <-ch
	if !ok {
		t.Fatal("snapshot reported success, ignoring the final transport error")
	}
	if !errors.Is(got, failure) {
		t.Fatalf("snapshot error = %v, want %v", got, failure)
	}
	if _, open := <-ch; open {
		t.Fatal("snapshot error channel was not closed")
	}
}

func TestStreamSnapshotV2SuccessfulCloseHasNoError(t *testing.T) {
	w := &failingSnapshotCloser{}
	ch := make(chan error, 1)
	js := &jetStream{}
	js.streamSnapshotV2(nil, &StreamState{}, w, false, nil, ch)
	if !w.closed {
		t.Fatal("snapshot transport was not closed")
	}
	if got, open := <-ch; open {
		t.Fatalf("successful snapshot close reported unexpected error %v", got)
	}
}
