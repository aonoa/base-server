package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

type SiteMessage struct {
	ent.Schema
}

func (SiteMessage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("站内信发布记录表"),
		entsql.Annotation{
			Table: "sys_site_message",
		},
	}
}

func (SiteMessage) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(uuid.NewString).
			Unique().
			Immutable().
			Comment("数据唯一标识"),
		field.String("title").NotEmpty().Comment("消息标题"),
		field.String("content").NotEmpty().Comment("消息正文"),
		field.String("category").Default("system").Comment("消息分类"),
		field.String("status").Default("published").Comment("消息状态 draft|scheduled|published|recalled"),
		field.String("receiver_type").Default("all").Comment("接收范围 all|user"),
		field.JSON("receiver_ids", []string{}).Optional().Comment("指定接收用户ID列表"),
		field.Int64("receiver_count").Default(0).Comment("接收人数"),
		field.String("link").Default("").Comment("消息跳转链接"),
		field.String("sender_id").Default("").Comment("发送人ID"),
		field.String("sender_name").Default("").Comment("发送人名称"),
		field.Time("scheduled_publish_time").Optional().Nillable().Comment("定时发布时间"),
		field.Time("published_time").Optional().Nillable().Comment("实际发布时间"),
		field.Time("recalled_time").Optional().Nillable().Comment("撤回时间"),
	}
}

func (SiteMessage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

func (SiteMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("create_time"),
		index.Fields("receiver_type"),
		index.Fields("category"),
		index.Fields("status"),
		index.Fields("scheduled_publish_time"),
	}
}
