package authx

import (
	"context"
	"time"
)

const (
	ProjectionSyncModeSnapshot = "snapshot"
	ProjectionSyncModeDelta    = "delta"

	ProjectionStateSynced = "synced"
	ProjectionStateError  = "error"
)

// ProjectionStatus describes the control-plane status emitted by a permission
// source after projection snapshot or delta publication.
type ProjectionStatus struct {
	SourceService        string
	SyncMode             string
	State                string
	LastSnapshotRevision uint64
	LastSyncTime         string
	LastError            string
	Description          string
}

type ProjectionStatusReporter interface {
	ReportProjectionStatus(context.Context, ProjectionStatus) error
}

func projectionStatusNow() string {
	return time.Now().Format(time.DateTime)
}
