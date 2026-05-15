package data

import (
	"context"
	"strings"
	"time"

	adminv1 "base-server/api/gen/go/admin/service/v1"
	v1 "base-server/api/gen/go/common/service/v1"
	userv1 "base-server/api/gen/go/user/service/v1"
	"base-server/app/common/service/internal/biz"
	"base-server/pkg/authx"
	"base-server/pkg/data/ent"
	"base-server/pkg/data/ent/sitemessage"
	"base-server/pkg/data/ent/sitemessagereceipt"
	"base-server/pkg/tools"
	sqlx "entgo.io/ent/dialect/sql"
	kratoserrors "github.com/go-kratos/kratos/v2/errors"
)

const (
	defaultOrganizationID       = "9f740c1b-0210-4e3a-858d-d128edea924d"
	siteMessageReceiptBatchSize = 500
)

func (r *commonRepo) GetUserDisplayName(ctx context.Context, userID string) (string, error) {
	ctx = authx.ForwardAuthorizationContext(ctx)
	reply, err := r.data.userClient.GetUserInfo(ctx, &userv1.GetUserInfoRequest{UserId: userID})
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(reply.GetNickname()) != "" {
		return strings.TrimSpace(reply.GetNickname()), nil
	}
	if strings.TrimSpace(reply.GetUsername()) != "" {
		return strings.TrimSpace(reply.GetUsername()), nil
	}
	return userID, nil
}

func (r *commonRepo) GetUserRoleValues(ctx context.Context, userID string, organizationID string) ([]string, error) {
	ctx = authx.ForwardAuthorizationContext(ctx)
	bindingReply, err := r.data.adminClient.GetUserRoleBinding(ctx, &adminv1.GetUserRoleBindingRequest{
		UserId:         userID,
		OrganizationId: strings.TrimSpace(organizationID),
	})
	if err != nil {
		return nil, err
	}
	roleIDs := make([]int64, 0, len(bindingReply.GetRoleIds())+1)
	seen := make(map[int64]struct{}, len(bindingReply.GetRoleIds())+1)
	if bindingReply.GetRoleId() != 0 {
		roleIDs = append(roleIDs, bindingReply.GetRoleId())
		seen[bindingReply.GetRoleId()] = struct{}{}
	}
	for _, roleID := range bindingReply.GetRoleIds() {
		if roleID == 0 {
			continue
		}
		if _, ok := seen[roleID]; ok {
			continue
		}
		seen[roleID] = struct{}{}
		roleIDs = append(roleIDs, roleID)
	}
	if len(roleIDs) == 0 {
		return []string{}, nil
	}
	valueReply, err := r.data.adminClient.ResolveRoleValues(ctx, &adminv1.ResolveRoleValuesRequest{RoleIds: roleIDs})
	if err != nil {
		return nil, err
	}
	values := make([]string, 0, len(valueReply.GetItems()))
	for _, item := range valueReply.GetItems() {
		value := strings.TrimSpace(item.GetValue())
		if value == "" {
			continue
		}
		values = append(values, value)
	}
	return values, nil
}

func (r *commonRepo) CurrentOrganizationID(ctx context.Context) string {
	organizationID := strings.TrimSpace(authx.RequestHeader(ctx, authx.HeaderOrganizationID))
	if organizationID == "" {
		return defaultOrganizationID
	}
	return organizationID
}

func (r *commonRepo) resolveSiteMessageReceivers(ctx context.Context, organizationID string) ([]string, error) {
	ctx = authx.ForwardAuthorizationContext(ctx)
	reply, err := r.data.adminClient.ListOrganizationMemberUserIds(ctx, &adminv1.ListOrganizationMemberUserIdsRequest{
		OrganizationId:  strings.TrimSpace(organizationID),
		ActiveUsersOnly: true,
	})
	if err != nil {
		return nil, err
	}
	receiverIDs := normalizeSiteMessageReceiverIDs(reply.GetUserIds())
	if len(receiverIDs) == 0 {
		return nil, kratoserrors.BadRequest("BAD_REQUEST", "no active receiver found")
	}
	return receiverIDs, nil
}

func (r *commonRepo) countActiveSiteMessageReceivers(ctx context.Context, organizationID string) (int64, error) {
	ctx = authx.ForwardAuthorizationContext(ctx)
	reply, err := r.data.adminClient.ListOrganizationMemberUserIds(ctx, &adminv1.ListOrganizationMemberUserIdsRequest{
		OrganizationId:  strings.TrimSpace(organizationID),
		ActiveUsersOnly: true,
	})
	if err != nil {
		return 0, err
	}
	return int64(len(normalizeSiteMessageReceiverIDs(reply.GetUserIds()))), nil
}

func normalizeSiteMessageReceiverIDs(userIDs []string) []string {
	receiverIDs := make([]string, 0, len(userIDs))
	seen := make(map[string]struct{}, len(userIDs))
	for _, userID := range userIDs {
		userID = strings.TrimSpace(userID)
		if userID == "" {
			continue
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		receiverIDs = append(receiverIDs, userID)
	}
	return receiverIDs
}

func parseSiteMessageSchedule(value string) (time.Time, error) {
	return time.ParseInLocation(time.DateTime, value, time.Local)
}

func (r *commonRepo) createSiteMessageReceipts(ctx context.Context, tx *ent.Tx, messageID string, organizationID string, receiverIDs []string) error {
	if len(receiverIDs) == 0 {
		return nil
	}
	for start := 0; start < len(receiverIDs); start += siteMessageReceiptBatchSize {
		end := start + siteMessageReceiptBatchSize
		if end > len(receiverIDs) {
			end = len(receiverIDs)
		}
		creates := make([]*ent.SiteMessageReceiptCreate, 0, end-start)
		for _, receiverID := range receiverIDs[start:end] {
			creates = append(creates, tx.SiteMessageReceipt.Create().
				SetMessageID(messageID).
				SetOrganizationID(organizationID).
				SetUserID(receiverID))
		}
		if err := tx.SiteMessageReceipt.CreateBulk(creates...).Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *commonRepo) CreateSiteMessage(ctx context.Context, senderID, senderName string, organizationID string, req *v1.CreateSiteMessageRequest) (*ent.SiteMessage, error) {
	organizationID = strings.TrimSpace(organizationID)
	if organizationID == "" {
		organizationID = defaultOrganizationID
	}
	receiverCount, err := r.countActiveSiteMessageReceivers(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	var scheduledPublishTime time.Time
	if req.GetAction() == biz.SiteMessageActionSchedule {
		scheduledPublishTime, err = parseSiteMessageSchedule(req.GetScheduledPublishTime())
		if err != nil {
			return nil, kratoserrors.BadRequest("BAD_REQUEST", "invalid scheduled_publish_time")
		}
	}

	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var message *ent.SiteMessage
	switch {
	case strings.TrimSpace(req.GetId()) == "":
		createBuilder := tx.SiteMessage.Create().
			SetTitle(req.GetTitle()).
			SetContent(req.GetContent()).
			SetCategory(req.GetCategory()).
			SetLink(req.GetLink()).
			SetOrganizationID(organizationID).
			SetSenderID(senderID).
			SetSenderName(senderName)
		switch req.GetAction() {
		case biz.SiteMessageActionDraft:
			createBuilder.
				SetStatus(biz.SiteMessageStatusDraft).
				SetReceiverCount(receiverCount)
		case biz.SiteMessageActionSchedule:
			createBuilder.
				SetStatus(biz.SiteMessageStatusScheduled).
				SetReceiverCount(receiverCount).
				SetScheduledPublishTime(scheduledPublishTime)
		default:
			actualReceiverIDs, err := r.resolveSiteMessageReceivers(ctx, organizationID)
			if err != nil {
				return nil, err
			}
			createBuilder.
				SetStatus(biz.SiteMessageStatusPublished).
				SetReceiverCount(int64(len(actualReceiverIDs))).
				SetPublishedTime(time.Now())
			message, err = createBuilder.Save(ctx)
			if err != nil {
				return nil, err
			}
			if err := r.createSiteMessageReceipts(ctx, tx, message.ID, organizationID, actualReceiverIDs); err != nil {
				return nil, err
			}
		}
		if message == nil {
			message, err = createBuilder.Save(ctx)
			if err != nil {
				return nil, err
			}
		}
	default:
		existing, err := tx.SiteMessage.Query().
			Where(
				sitemessage.IDEQ(req.GetId()),
				sitemessage.OrganizationIDEQ(organizationID),
			).
			Only(ctx)
		if err != nil {
			return nil, err
		}
		if existing.Status == biz.SiteMessageStatusPublished || existing.Status == biz.SiteMessageStatusRecalled {
			return nil, kratoserrors.BadRequest("BAD_REQUEST", "only draft or scheduled messages can be updated")
		}
		updateBuilder := tx.SiteMessage.UpdateOneID(existing.ID).
			SetTitle(req.GetTitle()).
			SetContent(req.GetContent()).
			SetCategory(req.GetCategory()).
			SetLink(req.GetLink()).
			SetOrganizationID(organizationID).
			SetSenderID(senderID).
			SetSenderName(senderName)
		switch req.GetAction() {
		case biz.SiteMessageActionDraft:
			updateBuilder.
				ClearScheduledPublishTime().
				ClearPublishedTime().
				ClearRecalledTime().
				SetStatus(biz.SiteMessageStatusDraft).
				SetReceiverCount(receiverCount)
		case biz.SiteMessageActionSchedule:
			updateBuilder.
				ClearPublishedTime().
				ClearRecalledTime().
				SetStatus(biz.SiteMessageStatusScheduled).
				SetReceiverCount(receiverCount).
				SetScheduledPublishTime(scheduledPublishTime)
		default:
			actualReceiverIDs, err := r.resolveSiteMessageReceivers(ctx, organizationID)
			if err != nil {
				return nil, err
			}
			updateBuilder.
				ClearScheduledPublishTime().
				ClearRecalledTime().
				SetStatus(biz.SiteMessageStatusPublished).
				SetReceiverCount(int64(len(actualReceiverIDs))).
				SetPublishedTime(time.Now())
			message, err = updateBuilder.Save(ctx)
			if err != nil {
				return nil, err
			}
			if err := r.createSiteMessageReceipts(ctx, tx, message.ID, organizationID, actualReceiverIDs); err != nil {
				return nil, err
			}
		}
		if message == nil {
			message, err = updateBuilder.Save(ctx)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true
	return message, nil
}

func getMySiteMessageListQuery(userID string, organizationID string, params *v1.GetMySiteMessageListParams, isPage bool) func(s *sqlx.Selector) {
	return func(s *sqlx.Selector) {
		s.Where(sqlx.EQ(sitemessagereceipt.FieldUserID, userID))
		s.Where(sqlx.EQ(sitemessagereceipt.FieldOrganizationID, organizationID))
		switch params.GetReadStatus() {
		case 1:
			s.Where(sqlx.EQ(sitemessagereceipt.FieldIsRead, true))
		case 2:
			s.Where(sqlx.EQ(sitemessagereceipt.FieldIsRead, false))
		}
		if isPage {
			s.OrderBy(sqlx.Desc(sitemessagereceipt.FieldCreateTime))
			if params.GetPageSize() != 0 {
				s.Limit(int(params.GetPageSize()))
			}
			if params.GetCurrentPage() != 0 {
				s.Offset(int(tools.GetPageOffset(params.GetCurrentPage(), params.GetPageSize())))
			}
		}
	}
}

func (r *commonRepo) GetMySiteMessageList(ctx context.Context, userID string, organizationID string, req *v1.GetMySiteMessageListParams) ([]*biz.SiteMessageEnvelope, int64, error) {
	receipts, err := r.data.db.SiteMessageReceipt.Query().Modify(getMySiteMessageListQuery(userID, organizationID, req, true)).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	count, err := r.data.db.SiteMessageReceipt.Query().Modify(getMySiteMessageListQuery(userID, organizationID, req, false)).Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if len(receipts) == 0 {
		return []*biz.SiteMessageEnvelope{}, int64(count), nil
	}
	messageIDs := make([]string, 0, len(receipts))
	for _, receipt := range receipts {
		messageIDs = append(messageIDs, receipt.MessageID)
	}
	messages, err := r.data.db.SiteMessage.Query().Where(sitemessage.IDIn(messageIDs...)).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	messageMap := make(map[string]*ent.SiteMessage, len(messages))
	for _, item := range messages {
		messageMap[item.ID] = item
	}
	items := make([]*biz.SiteMessageEnvelope, 0, len(receipts))
	for _, receipt := range receipts {
		message := messageMap[receipt.MessageID]
		if message == nil {
			continue
		}
		items = append(items, &biz.SiteMessageEnvelope{
			Message: message,
			Receipt: receipt,
		})
	}
	return items, int64(count), nil
}

func (r *commonRepo) GetMySiteMessageUnreadCount(ctx context.Context, userID string, organizationID string) (int64, error) {
	count, err := r.data.db.SiteMessageReceipt.Query().
		Where(
			sitemessagereceipt.UserIDEQ(userID),
			sitemessagereceipt.OrganizationIDEQ(organizationID),
			sitemessagereceipt.IsReadEQ(false),
		).
		Count(ctx)
	return int64(count), err
}

func (r *commonRepo) MarkSiteMessageRead(ctx context.Context, userID, organizationID, messageID string) error {
	_, err := r.data.db.SiteMessageReceipt.Update().
		Where(
			sitemessagereceipt.UserIDEQ(userID),
			sitemessagereceipt.OrganizationIDEQ(organizationID),
			sitemessagereceipt.MessageIDEQ(messageID),
		).
		SetIsRead(true).
		SetReadTime(time.Now()).
		Save(ctx)
	return err
}

func (r *commonRepo) MarkSiteMessageUnread(ctx context.Context, userID, organizationID, messageID string) error {
	_, err := r.data.db.SiteMessageReceipt.Update().
		Where(
			sitemessagereceipt.UserIDEQ(userID),
			sitemessagereceipt.OrganizationIDEQ(organizationID),
			sitemessagereceipt.MessageIDEQ(messageID),
		).
		SetIsRead(false).
		SetReadTime(time.Time{}).
		Save(ctx)
	return err
}

func (r *commonRepo) MarkAllSiteMessagesRead(ctx context.Context, userID string, organizationID string) (int64, error) {
	updated, err := r.data.db.SiteMessageReceipt.Update().
		Where(
			sitemessagereceipt.UserIDEQ(userID),
			sitemessagereceipt.OrganizationIDEQ(organizationID),
			sitemessagereceipt.IsReadEQ(false),
		).
		SetIsRead(true).
		SetReadTime(time.Now()).
		Save(ctx)
	return int64(updated), err
}

func getSiteMessageManageListQuery(organizationID string, params *v1.GetSiteMessageManageListParams, isPage bool) func(s *sqlx.Selector) {
	return func(s *sqlx.Selector) {
		s.Where(sqlx.EQ(sitemessage.FieldOrganizationID, organizationID))
		if params.GetStatus() != "" {
			s.Where(sqlx.EQ(sitemessage.FieldStatus, params.GetStatus()))
		}
		if isPage {
			s.OrderBy(sqlx.Desc(sitemessage.FieldUpdateTime))
			if params.GetPageSize() != 0 {
				s.Limit(int(params.GetPageSize()))
			}
			if params.GetCurrentPage() != 0 {
				s.Offset(int(tools.GetPageOffset(params.GetCurrentPage(), params.GetPageSize())))
			}
		}
	}
}

func (r *commonRepo) GetSiteMessageManageList(ctx context.Context, organizationID string, req *v1.GetSiteMessageManageListParams) ([]*ent.SiteMessage, int64, error) {
	items, err := r.data.db.SiteMessage.Query().Modify(getSiteMessageManageListQuery(organizationID, req, true)).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	count, err := r.data.db.SiteMessage.Query().Modify(getSiteMessageManageListQuery(organizationID, req, false)).Count(ctx)
	return items, int64(count), err
}

func (r *commonRepo) RecallSiteMessage(ctx context.Context, organizationID string, messageID string) error {
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	messageItem, err := tx.SiteMessage.Query().
		Where(
			sitemessage.IDEQ(messageID),
			sitemessage.OrganizationIDEQ(organizationID),
		).
		Only(ctx)
	if err != nil {
		return err
	}
	if messageItem.Status != biz.SiteMessageStatusPublished {
		return kratoserrors.BadRequest("BAD_REQUEST", "only published messages can be recalled")
	}
	if _, err := tx.SiteMessage.UpdateOneID(messageID).
		SetStatus(biz.SiteMessageStatusRecalled).
		SetRecalledTime(time.Now()).
		Save(ctx); err != nil {
		return err
	}
	if _, err := tx.SiteMessageReceipt.Delete().
		Where(
			sitemessagereceipt.MessageIDEQ(messageID),
			sitemessagereceipt.OrganizationIDEQ(organizationID),
		).
		Exec(ctx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func (r *commonRepo) DeletePendingSiteMessage(ctx context.Context, organizationID string, messageID string) error {
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	messageItem, err := tx.SiteMessage.Query().
		Where(
			sitemessage.IDEQ(messageID),
			sitemessage.OrganizationIDEQ(organizationID),
		).
		Only(ctx)
	if err != nil {
		return err
	}
	if messageItem.Status != biz.SiteMessageStatusDraft && messageItem.Status != biz.SiteMessageStatusScheduled {
		return kratoserrors.BadRequest("BAD_REQUEST", "only draft or scheduled messages can be deleted")
	}
	if _, err := tx.SiteMessageReceipt.Delete().
		Where(
			sitemessagereceipt.MessageIDEQ(messageID),
			sitemessagereceipt.OrganizationIDEQ(organizationID),
		).
		Exec(ctx); err != nil {
		return err
	}
	if err := tx.SiteMessage.DeleteOneID(messageID).Exec(ctx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func (r *commonRepo) promoteScheduledSiteMessage(ctx context.Context, messageItem *ent.SiteMessage) error {
	tx, err := r.data.db.Tx(ctx)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	receiverIDs, err := r.resolveSiteMessageReceivers(ctx, messageItem.OrganizationID)
	if err != nil {
		return err
	}
	updated, err := tx.SiteMessage.Update().
		Where(
			sitemessage.IDEQ(messageItem.ID),
			sitemessage.StatusEQ(biz.SiteMessageStatusScheduled),
		).
		SetStatus(biz.SiteMessageStatusPublished).
		SetReceiverCount(int64(len(receiverIDs))).
		SetPublishedTime(time.Now()).
		ClearScheduledPublishTime().
		Save(ctx)
	if err != nil {
		return err
	}
	if updated == 0 {
		return nil
	}
	if err := r.createSiteMessageReceipts(ctx, tx, messageItem.ID, messageItem.OrganizationID, receiverIDs); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func (r *commonRepo) PromoteDueScheduledSiteMessages(ctx context.Context) error {
	now := time.Now()
	items, err := r.data.db.SiteMessage.Query().
		Where(
			sitemessage.StatusEQ(biz.SiteMessageStatusScheduled),
			sitemessage.ScheduledPublishTimeLTE(now),
		).
		All(ctx)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := r.promoteScheduledSiteMessage(ctx, item); err != nil {
			return err
		}
	}
	return nil
}
