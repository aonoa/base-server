package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// ProjectionSourceStatus holds the schema definition for projection source status records.
type ProjectionSourceStatus struct {
	ent.Schema
}

func (ProjectionSourceStatus) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("权限投影源状态表"),
		entsql.Annotation{
			Table: "sys_projection_source_status",
		},
	}
}

func (ProjectionSourceStatus) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(uuid.NewString).
			Unique().
			Comment("数据唯一标识"),
		field.String("source_service").
			Unique().
			Comment("投影源服务编码"),
		field.String("sync_mode").Default("").Comment("同步模式"),
		field.String("state").Default("").Comment("当前状态"),
		field.Uint64("last_snapshot_revision").Default(0).Comment("最后一次快照版本"),
		field.String("last_sync_time").Default("").Comment("最后同步时间"),
		field.String("last_error").Default("").Comment("最后错误信息"),
		field.String("description").Default("").Comment("描述"),
	}
}

func (ProjectionSourceStatus) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

func (ProjectionSourceStatus) Edges() []ent.Edge {
	return nil
}
