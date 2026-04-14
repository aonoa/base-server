package server

import (
	"context"
	"time"

	"base-server/app/admin/service/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
)

type startupSyncServer struct {
	uc  *biz.AdminUsecase
	log *log.Helper
}

func NewStartupSyncServer(uc *biz.AdminUsecase, logger log.Logger) transport.Server {
	return &startupSyncServer{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

func (s *startupSyncServer) Start(ctx context.Context) error {
	backoff := time.Second
	for {
		err := s.uc.RegisterPermissionSnapshot(ctx)
		if err == nil {
			s.log.Info("registered permission snapshot to auth")
			<-ctx.Done()
			return ctx.Err()
		}
		s.log.Errorf("register permission snapshot to auth failed: %v", err)
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

func (s *startupSyncServer) Stop(ctx context.Context) error {
	_ = ctx
	return nil
}
