package data

import (
	"context"

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
	return r.repo.SaveProjectionSourceStatus(ctx, status)
}
