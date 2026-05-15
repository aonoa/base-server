package biz

import (
	"context"

	v1 "base-server/api/gen/go/common/service/v1"
	"base-server/pkg/data/ent"
	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewCommonUsecase)

type SiteMessageEnvelope struct {
	Message *ent.SiteMessage
	Receipt *ent.SiteMessageReceipt
}

type CommonRepo interface {
	GetUserDisplayName(context.Context, string) (string, error)
	GetUserRoleValues(context.Context, string, string) ([]string, error)
	CurrentOrganizationID(context.Context) string
	CreateSiteMessage(context.Context, string, string, string, *v1.CreateSiteMessageRequest) (*ent.SiteMessage, error)
	GetMySiteMessageList(context.Context, string, string, *v1.GetMySiteMessageListParams) ([]*SiteMessageEnvelope, int64, error)
	GetMySiteMessageUnreadCount(context.Context, string, string) (int64, error)
	MarkSiteMessageRead(context.Context, string, string, string) error
	MarkSiteMessageUnread(context.Context, string, string, string) error
	MarkAllSiteMessagesRead(context.Context, string, string) (int64, error)
	GetSiteMessageManageList(context.Context, string, *v1.GetSiteMessageManageListParams) ([]*ent.SiteMessage, int64, error)
	RecallSiteMessage(context.Context, string, string) error
	DeletePendingSiteMessage(context.Context, string, string) error
	PromoteDueScheduledSiteMessages(context.Context) error
}

type CommonUsecase struct {
	repo CommonRepo
}

func NewCommonUsecase(repo CommonRepo) *CommonUsecase {
	return &CommonUsecase{repo: repo}
}
