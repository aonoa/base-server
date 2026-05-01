package authx

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
)

func TestSyncProjectionDeltaFallsBackToSnapshot(t *testing.T) {
	applyCalls := 0
	snapshotCalls := 0

	err := SyncProjectionDelta(
		context.Background(),
		log.NewHelper(log.NewStdLogger(io.Discard)),
		"role update",
		func() error {
			applyCalls++
			return errors.New("delta failed")
		},
		func(context.Context) error {
			snapshotCalls++
			return nil
		},
	)
	if err != nil {
		t.Fatalf("sync projection delta: %v", err)
	}
	if applyCalls != 1 {
		t.Fatalf("expected apply to be called once, got %d", applyCalls)
	}
	if snapshotCalls != 1 {
		t.Fatalf("expected snapshot fallback to be called once, got %d", snapshotCalls)
	}
}

func TestSyncProjectionDeltaReturnsSnapshotFailure(t *testing.T) {
	expected := "snapshot failed"
	err := SyncProjectionDelta(
		context.Background(),
		log.NewHelper(log.NewStdLogger(io.Discard)),
		"role update",
		func() error {
			return errors.New("delta failed")
		},
		func(context.Context) error {
			return errors.New(expected)
		},
	)
	if err == nil {
		t.Fatalf("expected error")
	}
	if got := err.Error(); got == "" || !strings.Contains(got, expected) {
		t.Fatalf("expected error to contain %q, got %q", expected, got)
	}
}
