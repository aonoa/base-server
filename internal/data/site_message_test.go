package data

import (
	"context"
	"fmt"
	"testing"
	"time"

	pb "base-server/api/gen/go/base_api/v1"
	"base-server/internal/biz"
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

func TestCreateSiteMessage_UpdateDraftTransitions(t *testing.T) {
	t.Run("publish", func(t *testing.T) {
		repo := newTestBaseRepo(t)
		ctx := context.Background()

		draft, err := repo.CreateSiteMessage(ctx, "sender-1", "Sender", &pb.CreateSiteMessageRequest{
			Title:        "draft title",
			Content:      "draft content",
			Category:     "system",
			ReceiverType: "user",
			ReceiverIds:  []string{"receiver-1"},
			Action:       biz.SiteMessageActionDraft,
		})
		if err != nil {
			t.Fatalf("create draft: %v", err)
		}

		published, err := repo.CreateSiteMessage(ctx, "sender-1", "Sender", &pb.CreateSiteMessageRequest{
			Id:           draft.ID,
			Title:        "draft title",
			Content:      "draft content",
			Category:     "system",
			ReceiverType: "user",
			ReceiverIds:  []string{"receiver-1"},
			Action:       biz.SiteMessageActionPublish,
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
		count, err := repo.data.db.SiteMessageReceipt.Query().Count(ctx)
		if err != nil {
			t.Fatalf("count receipts: %v", err)
		}
		if count != 1 {
			t.Fatalf("unexpected receipt count: %d", count)
		}
	})

	t.Run("schedule", func(t *testing.T) {
		repo := newTestBaseRepo(t)
		ctx := context.Background()

		draft, err := repo.CreateSiteMessage(ctx, "sender-1", "Sender", &pb.CreateSiteMessageRequest{
			Title:        "draft title",
			Content:      "draft content",
			Category:     "system",
			ReceiverType: "user",
			ReceiverIds:  []string{"receiver-1"},
			Action:       biz.SiteMessageActionDraft,
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
			ReceiverType:         "user",
			ReceiverIds:          []string{"receiver-1"},
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
	})
}

func TestMarkSiteMessageUnread(t *testing.T) {
	repo := newTestBaseRepo(t)
	ctx := context.Background()

	published, err := repo.CreateSiteMessage(ctx, "sender-1", "Sender", &pb.CreateSiteMessageRequest{
		Title:        "published title",
		Content:      "published content",
		Category:     "system",
		ReceiverType: "user",
		ReceiverIds:  []string{"receiver-1"},
		Action:       biz.SiteMessageActionPublish,
	})
	if err != nil {
		t.Fatalf("publish message: %v", err)
	}

	if err := repo.MarkSiteMessageRead(ctx, "receiver-1", published.ID); err != nil {
		t.Fatalf("mark read: %v", err)
	}

	updatedCount, err := repo.GetMySiteMessageUnreadCount(ctx, "receiver-1")
	if err != nil {
		t.Fatalf("count unread after read: %v", err)
	}
	if updatedCount != 0 {
		t.Fatalf("unexpected unread count after read: %d", updatedCount)
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
