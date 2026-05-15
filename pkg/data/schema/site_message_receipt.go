package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
	"time"
)

type SiteMessageReceipt struct {
	ent.Schema
}

func (SiteMessageReceipt) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("站内信收件记录表"),
		entsql.Annotation{
			Table: "sys_site_message_receipt",
		},
	}
}

func (SiteMessageReceipt) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(uuid.NewString).
			Unique().
			Immutable().
			Comment("数据唯一标识"),
		field.String("message_id").Comment("站内信ID"),
		field.String("user_id").Comment("收件用户ID"),
		field.String("organization_id").
			Default("9f740c1b-0210-4e3a-858d-d128edea924d").
			NotEmpty().
			Comment("组织ID"),
		field.Bool("is_read").Default(false).Comment("是否已读"),
		field.Time("read_time").Default(time.Time{}).Comment("已读时间"),
	}
}

func (SiteMessageReceipt) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

func (SiteMessageReceipt) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("organization_id", "user_id", "is_read"),
		index.Fields("message_id", "user_id").Unique(),
		index.Fields("create_time"),
	}
}
