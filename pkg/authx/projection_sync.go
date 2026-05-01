package authx

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
)

// ProjectionStartupSyncServer retries snapshot registration on service startup
// until auth accepts the current projection snapshot. It is reusable by any
// permission-source service, not only admin.
type ProjectionStartupSyncServer struct {
	registerSnapshot func(context.Context) error
	log              *log.Helper
	successMessage   string
	failureMessage   string
}

func NewProjectionStartupSyncServer(registerSnapshot func(context.Context) error, logger log.Logger, sourceService string) transport.Server {
	helper := log.NewHelper(logger)
	return &ProjectionStartupSyncServer{
		registerSnapshot: registerSnapshot,
		log:              helper,
		successMessage:   fmt.Sprintf("%s registered permission snapshot to auth", sourceService),
		failureMessage:   fmt.Sprintf("%s register permission snapshot to auth failed", sourceService),
	}
}

func (s *ProjectionStartupSyncServer) Start(ctx context.Context) error {
	backoff := time.Second
	for {
		err := s.registerSnapshot(ctx)
		if err == nil {
			s.log.Info(s.successMessage)
			<-ctx.Done()
			return ctx.Err()
		}
		s.log.Errorf("%s: %v", s.failureMessage, err)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
		}
	}
}

func (s *ProjectionStartupSyncServer) Stop(ctx context.Context) error {
	_ = ctx
	return nil
}

// SyncProjectionDelta applies a projection delta and falls back to a snapshot
// rebuild when the delta cannot be applied. This is the shared pattern that
// every permission-source service should use when publishing changes into auth.
func SyncProjectionDelta(ctx context.Context, helper *log.Helper, label string, apply func() error, registerSnapshot func(context.Context) error) error {
	if err := apply(); err != nil {
		helper.Warnf("auth projection delta failed, fallback to snapshot: %s err=%v", label, err)
		if snapshotErr := registerSnapshot(ctx); snapshotErr != nil {
			return fmt.Errorf("apply auth projection update: %w; snapshot fallback failed: %v", err, snapshotErr)
		}
		helper.Infof("auth projection snapshot fallback succeeded: %s", label)
		return nil
	}
	helper.Infof("auth projection delta applied: %s", label)
	return nil
}
