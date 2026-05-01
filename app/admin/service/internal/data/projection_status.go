package data

import (
	"context"

	v1 "base-server/api/gen/go/admin/service/v1"
	"base-server/pkg/authx"
)

type adminProjectionStatusReporter struct {
	repo *adminRepo
}

func newAdminProjectionStatusReporter(repo *adminRepo) authx.ProjectionStatusReporter {
	return &adminProjectionStatusReporter{repo: repo}
}

func (r *adminProjectionStatusReporter) ReportProjectionStatus(ctx context.Context, status authx.ProjectionStatus) error {
	if r == nil || r.repo == nil {
		return nil
	}
	_, err := r.repo.ReportProjectionSourceStatus(ctx, &v1.ProjectionSourceStatusItem{
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
