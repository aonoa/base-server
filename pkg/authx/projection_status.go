package authx

import (
	"context"
	"time"

	adminv1 "base-server/api/gen/go/admin/service/v1"
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
	DomainCode           string
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

type AdminProjectionStatusReporter struct {
	client adminv1.AdminServiceClient
}

func NewAdminProjectionStatusReporter(client adminv1.AdminServiceClient) *AdminProjectionStatusReporter {
	return &AdminProjectionStatusReporter{client: client}
}

func (r *AdminProjectionStatusReporter) ReportProjectionStatus(ctx context.Context, status ProjectionStatus) error {
	if r == nil || r.client == nil {
		return nil
	}
	ctx = ForwardAuthorizationContext(ctx)
	_, err := r.client.ReportProjectionSourceStatus(ctx, &adminv1.ProjectionSourceStatusItem{
		SourceService:        status.SourceService,
		DomainCode:           status.DomainCode,
		SyncMode:             status.SyncMode,
		State:                status.State,
		LastSnapshotRevision: status.LastSnapshotRevision,
		LastSyncTime:         status.LastSyncTime,
		LastError:            status.LastError,
		Description:          status.Description,
	})
	return err
}

func projectionStatusNow() string {
	return time.Now().Format(time.DateTime)
}
