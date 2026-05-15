package biz

import (
	"context"
	"strings"
	"time"

	v1 "base-server/api/gen/go/common/service/v1"
	"base-server/pkg/authx"
	kratoserrors "github.com/go-kratos/kratos/v2/errors"
)

const (
	SiteMessageActionDraft    = "draft"
	SiteMessageActionSchedule = "schedule"
	SiteMessageActionPublish  = "publish"

	SiteMessageStatusDraft     = "draft"
	SiteMessageStatusScheduled = "scheduled"
	SiteMessageStatusPublished = "published"
	SiteMessageStatusRecalled  = "recalled"
)

func normalizeSiteMessageAction(action string) string {
	switch action {
	case SiteMessageActionDraft:
		return SiteMessageActionDraft
	case SiteMessageActionSchedule:
		return SiteMessageActionSchedule
	default:
		return SiteMessageActionPublish
	}
}

func normalizeSiteMessageStatus(status string) string {
	switch status {
	case SiteMessageStatusDraft:
		return SiteMessageStatusDraft
	case SiteMessageStatusScheduled:
		return SiteMessageStatusScheduled
	case SiteMessageStatusPublished:
		return SiteMessageStatusPublished
	case SiteMessageStatusRecalled:
		return SiteMessageStatusRecalled
	default:
		return ""
	}
}

func formatSiteMessageTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format(time.DateTime)
}

func siteMessageInboxCreatedTime(item SiteMessageEnvelope) string {
	if item.Message != nil && item.Message.PublishedTime != nil && !item.Message.PublishedTime.IsZero() {
		return item.Message.PublishedTime.Format(time.DateTime)
	}
	if item.Message == nil {
		return ""
	}
	return item.Message.CreateTime.Format(time.DateTime)
}

func (uc *CommonUsecase) currentUserID(ctx context.Context) (string, error) {
	userID := strings.TrimSpace(authx.UserID(ctx))
	if userID == "" {
		return "", kratoserrors.Unauthorized("UNAUTHORIZED", "missing user id")
	}
	return userID, nil
}

func (uc *CommonUsecase) ensureManageAccess(ctx context.Context, organizationID string) (string, error) {
	userID, err := uc.currentUserID(ctx)
	if err != nil {
		return "", err
	}
	values, err := uc.repo.GetUserRoleValues(ctx, userID, organizationID)
	if err != nil {
		return "", err
	}
	for _, value := range values {
		if value == "admin" || value == "root" {
			return userID, nil
		}
	}
	return "", kratoserrors.Forbidden("FORBIDDEN", "site message manage access denied")
}

func (uc *CommonUsecase) promoteDueScheduledSiteMessages(ctx context.Context) error {
	return uc.repo.PromoteDueScheduledSiteMessages(ctx)
}

func (uc *CommonUsecase) GetMySiteMessageList(ctx context.Context, req *v1.GetMySiteMessageListParams) (*v1.GetMySiteMessageListReply, error) {
	userID, err := uc.currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := uc.promoteDueScheduledSiteMessages(ctx); err != nil {
		return nil, err
	}
	organizationID := uc.repo.CurrentOrganizationID(ctx)
	items, total, err := uc.repo.GetMySiteMessageList(ctx, userID, organizationID, req)
	if err != nil {
		return nil, err
	}
	reply := &v1.GetMySiteMessageListReply{
		Items: make([]*v1.SiteMessageItem, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		if item == nil || item.Message == nil || item.Receipt == nil {
			continue
		}
		reply.Items = append(reply.Items, &v1.SiteMessageItem{
			Id:             item.Message.ID,
			Title:          item.Message.Title,
			Content:        item.Message.Content,
			IsRead:         item.Receipt.IsRead,
			Link:           item.Message.Link,
			SenderId:       item.Message.SenderID,
			SenderName:     item.Message.SenderName,
			CreatedTime:    siteMessageInboxCreatedTime(*item),
			ReadTime:       formatSiteMessageTime(&item.Receipt.ReadTime),
			OrganizationId: item.Message.OrganizationID,
		})
	}
	return reply, nil
}

func (uc *CommonUsecase) GetMySiteMessageUnreadCount(ctx context.Context) (*v1.GetMySiteMessageUnreadCountReply, error) {
	userID, err := uc.currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	if err := uc.promoteDueScheduledSiteMessages(ctx); err != nil {
		return nil, err
	}
	organizationID := uc.repo.CurrentOrganizationID(ctx)
	count, err := uc.repo.GetMySiteMessageUnreadCount(ctx, userID, organizationID)
	if err != nil {
		return nil, err
	}
	return &v1.GetMySiteMessageUnreadCountReply{UnreadCount: count}, nil
}

func (uc *CommonUsecase) MarkSiteMessageRead(ctx context.Context, req *v1.MarkSiteMessageReadRequest) error {
	userID, err := uc.currentUserID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(req.GetMessageId()) == "" {
		return kratoserrors.BadRequest("BAD_REQUEST", "message_id is required")
	}
	return uc.repo.MarkSiteMessageRead(ctx, userID, uc.repo.CurrentOrganizationID(ctx), req.MessageId)
}

func (uc *CommonUsecase) MarkSiteMessageUnread(ctx context.Context, req *v1.MarkSiteMessageReadRequest) error {
	userID, err := uc.currentUserID(ctx)
	if err != nil {
		return err
	}
	if strings.TrimSpace(req.GetMessageId()) == "" {
		return kratoserrors.BadRequest("BAD_REQUEST", "message_id is required")
	}
	return uc.repo.MarkSiteMessageUnread(ctx, userID, uc.repo.CurrentOrganizationID(ctx), req.MessageId)
}

func (uc *CommonUsecase) MarkAllSiteMessagesRead(ctx context.Context) (*v1.MarkAllSiteMessagesReadReply, error) {
	userID, err := uc.currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	updatedCount, err := uc.repo.MarkAllSiteMessagesRead(ctx, userID, uc.repo.CurrentOrganizationID(ctx))
	if err != nil {
		return nil, err
	}
	return &v1.MarkAllSiteMessagesReadReply{UpdatedCount: updatedCount}, nil
}

func (uc *CommonUsecase) GetSiteMessageManageList(ctx context.Context, req *v1.GetSiteMessageManageListParams) (*v1.GetSiteMessageManageListReply, error) {
	organizationID := uc.repo.CurrentOrganizationID(ctx)
	if _, err := uc.ensureManageAccess(ctx, organizationID); err != nil {
		return nil, err
	}
	req.Status = normalizeSiteMessageStatus(req.Status)
	if err := uc.promoteDueScheduledSiteMessages(ctx); err != nil {
		return nil, err
	}
	items, total, err := uc.repo.GetSiteMessageManageList(ctx, organizationID, req)
	if err != nil {
		return nil, err
	}
	reply := &v1.GetSiteMessageManageListReply{
		Items: make([]*v1.SiteMessageManageItem, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		reply.Items = append(reply.Items, &v1.SiteMessageManageItem{
			Id:                   item.ID,
			Title:                item.Title,
			Content:              item.Content,
			Status:               item.Status,
			ReceiverCount:        item.ReceiverCount,
			Link:                 item.Link,
			SenderId:             item.SenderID,
			SenderName:           item.SenderName,
			CreatedTime:          item.CreateTime.Format(time.DateTime),
			UpdatedTime:          item.UpdateTime.Format(time.DateTime),
			ScheduledPublishTime: formatSiteMessageTime(item.ScheduledPublishTime),
			PublishedTime:        formatSiteMessageTime(item.PublishedTime),
			RecalledTime:         formatSiteMessageTime(item.RecalledTime),
			OrganizationId:       item.OrganizationID,
		})
	}
	return reply, nil
}

func (uc *CommonUsecase) CreateSiteMessage(ctx context.Context, req *v1.CreateSiteMessageRequest) (*v1.CreateSiteMessageReply, error) {
	organizationID := uc.repo.CurrentOrganizationID(ctx)
	userID, err := uc.ensureManageAccess(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.GetTitle()) == "" {
		return nil, kratoserrors.BadRequest("BAD_REQUEST", "title is required")
	}
	if strings.TrimSpace(req.GetContent()) == "" {
		return nil, kratoserrors.BadRequest("BAD_REQUEST", "content is required")
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)
	req.Link = strings.TrimSpace(req.Link)
	req.Category = strings.TrimSpace(req.Category)
	if req.Category == "" {
		req.Category = "system"
	}
	req.Action = normalizeSiteMessageAction(req.Action)
	if req.Action == SiteMessageActionSchedule {
		if strings.TrimSpace(req.GetScheduledPublishTime()) == "" {
			return nil, kratoserrors.BadRequest("BAD_REQUEST", "scheduled_publish_time is required")
		}
		parsed, err := time.ParseInLocation(time.DateTime, req.ScheduledPublishTime, time.Local)
		if err != nil {
			return nil, kratoserrors.BadRequest("BAD_REQUEST", "scheduled_publish_time must use YYYY-MM-DD HH:mm:ss")
		}
		if !parsed.After(time.Now()) {
			return nil, kratoserrors.BadRequest("BAD_REQUEST", "scheduled_publish_time must be in the future")
		}
	} else {
		req.ScheduledPublishTime = ""
	}

	senderName, err := uc.repo.GetUserDisplayName(ctx, userID)
	if err != nil {
		return nil, err
	}
	message, err := uc.repo.CreateSiteMessage(ctx, userID, senderName, organizationID, req)
	if err != nil {
		return nil, err
	}
	return &v1.CreateSiteMessageReply{
		Id:                   message.ID,
		ReceiverCount:        message.ReceiverCount,
		Status:               message.Status,
		ScheduledPublishTime: formatSiteMessageTime(message.ScheduledPublishTime),
		PublishedTime:        formatSiteMessageTime(message.PublishedTime),
		OrganizationId:       message.OrganizationID,
	}, nil
}

func (uc *CommonUsecase) RecallSiteMessage(ctx context.Context, req *v1.RecallSiteMessageRequest) error {
	organizationID := uc.repo.CurrentOrganizationID(ctx)
	if _, err := uc.ensureManageAccess(ctx, organizationID); err != nil {
		return err
	}
	if strings.TrimSpace(req.GetId()) == "" {
		return kratoserrors.BadRequest("BAD_REQUEST", "id is required")
	}
	return uc.repo.RecallSiteMessage(ctx, organizationID, req.Id)
}

func (uc *CommonUsecase) DeletePendingSiteMessage(ctx context.Context, req *v1.DeletePendingSiteMessageRequest) error {
	organizationID := uc.repo.CurrentOrganizationID(ctx)
	if _, err := uc.ensureManageAccess(ctx, organizationID); err != nil {
		return err
	}
	if strings.TrimSpace(req.GetId()) == "" {
		return kratoserrors.BadRequest("BAD_REQUEST", "id is required")
	}
	return uc.repo.DeletePendingSiteMessage(ctx, organizationID, req.Id)
}
