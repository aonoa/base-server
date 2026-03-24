package data

import (
	"base-server/app/common/service/internal/biz"
	"base-server/app/common/service/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewCommonRepo)

// Data .
type Data struct{}

// NewData .
func NewData(c *conf.Data) (*Data, func(), error) {
	cleanup := func() {
		log.Info("closing the data resources")
	}
	_ = c
	return &Data{}, cleanup, nil
}

type commonRepo struct {
	data *Data
	log  *log.Helper
}

func NewCommonRepo(data *Data, logger log.Logger) biz.CommonRepo {
	return &commonRepo{data: data, log: log.NewHelper(logger)}
}
