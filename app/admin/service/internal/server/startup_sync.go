package server

import (
	"base-server/app/admin/service/internal/biz"
	"base-server/pkg/authx"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
)

func NewStartupSyncServer(uc *biz.AdminUsecase, logger log.Logger) transport.Server {
	return authx.NewProjectionStartupSyncServer(uc.RegisterPermissionSnapshot, logger, "admin")
}
