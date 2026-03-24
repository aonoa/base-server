package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewCommonUsecase)

type CommonRepo interface{}

type CommonUsecase struct {
	repo CommonRepo
}

func NewCommonUsecase(repo CommonRepo) *CommonUsecase {
	return &CommonUsecase{repo: repo}
}
