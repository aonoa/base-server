package data

import (
	"context"
	"fmt"
	"testing"
	"time"

	pb "base-server/api/gen/go/base_api/v1"
	"base-server/internal/biz"
	"base-server/internal/data/ent"
	"base-server/internal/data/ent/enttest"
	"base-server/internal/data/ent/sitemessagereceipt"
	_ "github.com/mattn/go-sqlite3"
)

func newTestBaseRepo(t *testing.T) *baseRepo {
	t.Helper()
	client := enttest.Open(t, "sqlite3", fmt.Sprintf("file:site-message-%d?mode=memory&cache=shared&_fk=1", time.Now().UnixNano()))
	t.Cleanup(func() {
		_ = client.Close()
	})
	return &baseRepo{
		data: &Data{db: client},
	}
}

func createActiveSiteMessageUser(t *testing.T, repo *baseRepo, ctx context.Context, username string) *ent.User {
	t.Helper()

	item, err := repo.data.db.User.Create().
		SetUsername(username).
		SetPassword("secret").
		SetNickname(username).
		SetStatus(1).
		SetAvatar("").
		SetDesc("").
		SetExtension("{}").
		Save(ctx)
	if err != nil {
		t.Fatalf("create active user %s: %v", username, err)
	}
	return item
}

func TestCreateSiteMessage_UpdateDraftTransitions(t *testing.T) {
	t.Run("publish", func(t *testing.T) {
		repo := newTestBaseRepo(t)
		ctx := context.Background()
		userA := createActiveSiteMessageUser(t, repo, ctx, "receiver-a")
		userB := createActiveSiteMessageUser(t, repo, ctx, "receiver-b")

		draft, err := repo.CreateSiteMessage(ctx, "sender-1", "Sender", &pb.CreateSiteMessageRequest{
			Title:    "draft title",
			Content:  "draft content",
			Category: "system",
			Action:   biz.SiteMessageActionDraft,
		})
		if err != nil {
			t.Fatalf("create draft: %v", err)
		}

		published, err := repo.CreateSiteMessage(ctx, "sender-1", "Sender", &pb.CreateSiteMessageRequest{
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
		if published.PublishedTime == nil {
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
		for _, userID := range []string{userA.ID.String(), userB.ID.String()} {
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

	t.Run("schedule", func(t *testing.T) {
		repo := newTestBaseRepo(t)
		ctx := context.Background()
		createActiveSiteMessageUser(t, repo, ctx, "receiver-a")
		createActiveSiteMessageUser(t, repo, ctx, "receiver-b")

		draft, err := repo.CreateSiteMessage(ctx, "sender-1", "Sender", &pb.CreateSiteMessageRequest{
			Title:    "draft title",
			Content:  "draft content",
			Category: "system",
			Action:   biz.SiteMessageActionDraft,
		})
		if err != nil {
			t.Fatalf("create draft: %v", err)
		}

		scheduledTime := time.Now().Add(time.Hour).Format(time.DateTime)
		scheduled, err := repo.CreateSiteMessage(ctx, "sender-1", "Sender", &pb.CreateSiteMessageRequest{
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
		if scheduled.ScheduledPublishTime == nil {
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
	repo := newTestBaseRepo(t)
	ctx := context.Background()
	receiver := createActiveSiteMessageUser(t, repo, ctx, "receiver-1")

	published, err := repo.CreateSiteMessage(ctx, "sender-1", "Sender", &pb.CreateSiteMessageRequest{
		Title:    "published title",
		Content:  "published content",
		Category: "system",
		Action:   biz.SiteMessageActionPublish,
	})
	if err != nil {
		t.Fatalf("publish message: %v", err)
	}

	if err := repo.MarkSiteMessageRead(ctx, receiver.ID.String(), published.ID); err != nil {
		t.Fatalf("mark read: %v", err)
	}

	updatedCount, err := repo.GetMySiteMessageUnreadCount(ctx, receiver.ID.String())
	if err != nil {
		t.Fatalf("count unread after read: %v", err)
	}
	if updatedCount != 0 {
		t.Fatalf("unexpected unread count after read: %d", updatedCount)
	}

	if err := repo.MarkSiteMessageUnread(ctx, receiver.ID.String(), published.ID); err != nil {
		t.Fatalf("mark unread: %v", err)
	}

	unreadCount, err := repo.GetMySiteMessageUnreadCount(ctx, receiver.ID.String())
	if err != nil {
		t.Fatalf("count unread after unread: %v", err)
	}
	if unreadCount != 1 {
		t.Fatalf("unexpected unread count after unread: %d", unreadCount)
	}

	receipt, err := repo.data.db.SiteMessageReceipt.Query().
		Where(
			sitemessagereceipt.UserIDEQ(receiver.ID.String()),
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
