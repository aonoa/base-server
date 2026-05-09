package data

import (
	"context"
	"fmt"
	"testing"
	"time"

	commonv1 "base-server/api/gen/go/common/service/v1"
	userv1 "base-server/api/gen/go/user/service/v1"
	"base-server/app/common/service/internal/biz"
	"base-server/pkg/data/ent/enttest"
	"base-server/pkg/data/ent/sitemessagereceipt"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type fakeUserServiceClient struct {
	items []*userv1.UserListItem
}

var _ userv1.UserServiceClient = (*fakeUserServiceClient)(nil)

func (f *fakeUserServiceClient) GetUserInfo(context.Context, *userv1.GetUserInfoRequest, ...grpc.CallOption) (*userv1.GetUserInfoReply, error) {
	return &userv1.GetUserInfoReply{
		UserId:   "sender-1",
		Username: "sender-1",
		Nickname: "Sender",
		Status:   1,
	}, nil
}

func (f *fakeUserServiceClient) GetUserList(_ context.Context, in *userv1.GetUserParams, _ ...grpc.CallOption) (*userv1.GetUserListReply, error) {
	activeItems := make([]*userv1.UserListItem, 0, len(f.items))
	for _, item := range f.items {
		if item == nil {
			continue
		}
		if in.GetStatus() != 0 && item.GetStatus() != in.GetStatus() {
			continue
		}
		activeItems = append(activeItems, item)
	}

	currentPage := in.GetCurrentPage()
	pageSize := in.GetPageSize()
	if currentPage <= 0 {
		currentPage = 1
	}
	if pageSize <= 0 {
		pageSize = int64(len(activeItems))
		if pageSize == 0 {
			pageSize = 1
		}
	}

	start := int((currentPage - 1) * pageSize)
	if start >= len(activeItems) {
		return &userv1.GetUserListReply{Items: []*userv1.UserListItem{}, Total: int64(len(activeItems))}, nil
	}
	end := start + int(pageSize)
	if end > len(activeItems) {
		end = len(activeItems)
	}
	return &userv1.GetUserListReply{
		Items: activeItems[start:end],
		Total: int64(len(activeItems)),
	}, nil
}

func (f *fakeUserServiceClient) AddUser(context.Context, *userv1.UserListItem, ...grpc.CallOption) (*userv1.UserListItem, error) {
	return nil, nil
}

func (f *fakeUserServiceClient) UpdateUser(context.Context, *userv1.UserListItem, ...grpc.CallOption) (*userv1.UserListItem, error) {
	return nil, nil
}

func (f *fakeUserServiceClient) DelUser(context.Context, *userv1.DeleteUser, ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (f *fakeUserServiceClient) IsUserExist(context.Context, *userv1.IsUserExistsRequest, ...grpc.CallOption) (*userv1.IsUserExistsReply, error) {
	return nil, nil
}

func (f *fakeUserServiceClient) ChangePassword(context.Context, *userv1.ChangePasswordRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (f *fakeUserServiceClient) ValidateUserAuth(context.Context, *userv1.ValidateUserAuthRequest, ...grpc.CallOption) (*userv1.ValidateUserAuthReply, error) {
	return nil, nil
}

func (f *fakeUserServiceClient) GetUserAuthInfo(context.Context, *userv1.GetUserAuthInfoRequest, ...grpc.CallOption) (*userv1.GetUserAuthInfoReply, error) {
	return nil, nil
}

func (f *fakeUserServiceClient) GetWalkRoute(context.Context, *emptypb.Empty, ...grpc.CallOption) (*userv1.GetWalkRouteReply, error) {
	return nil, nil
}

func newTestCommonRepo(t *testing.T, items []*userv1.UserListItem) *commonRepo {
	t.Helper()

	client := enttest.Open(t, "sqlite3", fmt.Sprintf("file:site-message-common-%d?mode=memory&cache=shared&_fk=1", time.Now().UnixNano()))
	t.Cleanup(func() {
		_ = client.Close()
	})

	return &commonRepo{
		data: &Data{
			db:         client,
			userClient: &fakeUserServiceClient{items: items},
		},
	}
}

func fakeActiveUsers(ids ...string) []*userv1.UserListItem {
	items := make([]*userv1.UserListItem, 0, len(ids))
	for _, id := range ids {
		items = append(items, &userv1.UserListItem{
			Id:       id,
			Username: id,
			Nickname: id,
			Status:   activeUserStatus,
		})
	}
	return items
}

func TestCreateSiteMessageUpdateTransitions(t *testing.T) {
	t.Run("publish draft", func(t *testing.T) {
		repo := newTestCommonRepo(t, fakeActiveUsers("receiver-a", "receiver-b"))
		ctx := context.Background()

		draft, err := repo.CreateSiteMessage(ctx, "sender-1", "Sender", &commonv1.CreateSiteMessageRequest{
			Title:    "draft title",
			Content:  "draft content",
			Category: "system",
			Action:   biz.SiteMessageActionDraft,
		})
		if err != nil {
			t.Fatalf("create draft: %v", err)
		}

		published, err := repo.CreateSiteMessage(ctx, "sender-1", "Sender", &commonv1.CreateSiteMessageRequest{
			Id:       draft.ID,
			Title:    "draft title",
			Content:  "draft content",
			Category: "system",
			Action:   biz.SiteMessageActionPublish,
		})
		if err != nil {
			t.Fatalf("publish draft: %v", err)
		}
		if published.Status != biz.SiteMessageStatusPublished {
			t.Fatalf("unexpected status: %s", published.Status)
		}
		if published.PublishedTime == nil || published.PublishedTime.IsZero() {
			t.Fatal("published time should be set")
		}
		if published.ReceiverCount != 2 {
			t.Fatalf("unexpected receiver count: %d", published.ReceiverCount)
		}

		count, err := repo.data.db.SiteMessageReceipt.Query().Count(ctx)
		if err != nil {
			t.Fatalf("count receipts: %v", err)
		}
		if count != 2 {
			t.Fatalf("unexpected receipt count: %d", count)
		}

		for _, userID := range []string{"receiver-a", "receiver-b"} {
			exists, err := repo.data.db.SiteMessageReceipt.Query().
				Where(
					sitemessagereceipt.UserIDEQ(userID),
					sitemessagereceipt.MessageIDEQ(published.ID),
				).
				Exist(ctx)
			if err != nil {
				t.Fatalf("check receipt for %s: %v", userID, err)
			}
			if !exists {
				t.Fatalf("missing receipt for %s", userID)
			}
		}
	})

	t.Run("schedule draft", func(t *testing.T) {
		repo := newTestCommonRepo(t, fakeActiveUsers("receiver-a", "receiver-b"))
		ctx := context.Background()

		draft, err := repo.CreateSiteMessage(ctx, "sender-1", "Sender", &commonv1.CreateSiteMessageRequest{
			Title:    "draft title",
			Content:  "draft content",
			Category: "system",
			Action:   biz.SiteMessageActionDraft,
		})
		if err != nil {
			t.Fatalf("create draft: %v", err)
		}

		scheduledTime := time.Now().Add(time.Hour).Format(time.DateTime)
		scheduled, err := repo.CreateSiteMessage(ctx, "sender-1", "Sender", &commonv1.CreateSiteMessageRequest{
			Id:                   draft.ID,
			Title:                "draft title",
			Content:              "draft content",
			Category:             "system",
			Action:               biz.SiteMessageActionSchedule,
			ScheduledPublishTime: scheduledTime,
		})
		if err != nil {
			t.Fatalf("schedule draft: %v", err)
		}
		if scheduled.Status != biz.SiteMessageStatusScheduled {
			t.Fatalf("unexpected status: %s", scheduled.Status)
		}
		if scheduled.ScheduledPublishTime == nil || scheduled.ScheduledPublishTime.IsZero() {
			t.Fatal("scheduled publish time should be set")
		}
		if scheduled.PublishedTime != nil {
			t.Fatal("published time should stay empty")
		}
		if scheduled.ReceiverCount != 2 {
			t.Fatalf("unexpected receiver count: %d", scheduled.ReceiverCount)
		}
	})
}

func TestMarkSiteMessageUnread(t *testing.T) {
	repo := newTestCommonRepo(t, fakeActiveUsers("receiver-1"))
	ctx := context.Background()

	published, err := repo.CreateSiteMessage(ctx, "sender-1", "Sender", &commonv1.CreateSiteMessageRequest{
		Title:    "published title",
		Content:  "published content",
		Category: "system",
		Action:   biz.SiteMessageActionPublish,
	})
	if err != nil {
		t.Fatalf("publish message: %v", err)
	}

	if err := repo.MarkSiteMessageRead(ctx, "receiver-1", published.ID); err != nil {
		t.Fatalf("mark read: %v", err)
	}

	unreadAfterRead, err := repo.GetMySiteMessageUnreadCount(ctx, "receiver-1")
	if err != nil {
		t.Fatalf("count unread after read: %v", err)
	}
	if unreadAfterRead != 0 {
		t.Fatalf("unexpected unread count after read: %d", unreadAfterRead)
	}

	if err := repo.MarkSiteMessageUnread(ctx, "receiver-1", published.ID); err != nil {
		t.Fatalf("mark unread: %v", err)
	}

	unreadCount, err := repo.GetMySiteMessageUnreadCount(ctx, "receiver-1")
	if err != nil {
		t.Fatalf("count unread after unread: %v", err)
	}
	if unreadCount != 1 {
		t.Fatalf("unexpected unread count after unread: %d", unreadCount)
	}

	receipt, err := repo.data.db.SiteMessageReceipt.Query().
		Where(
			sitemessagereceipt.UserIDEQ("receiver-1"),
			sitemessagereceipt.MessageIDEQ(published.ID),
		).
		Only(ctx)
	if err != nil {
		t.Fatalf("load receipt: %v", err)
	}
	if receipt.IsRead {
		t.Fatal("receipt should be unread")
	}
	if !receipt.ReadTime.IsZero() {
		t.Fatal("read time should be reset")
	}
}
