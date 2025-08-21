package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// Resource holds the schema definition for the Resource entity.
type Resource struct {
	ent.Schema
}

// Annotations of the Resource.
func (Resource) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("系统资源表"),
		entsql.Annotation{
			Table: "sys_resources",
		},
	}
}

// Fields of the Resource.
func (Resource) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(uuid.NewString).
			Unique().
			Comment("数据唯一标识"),
		field.String("name").Comment("资源名称"),
		field.String("type").Comment("资源类型"),
		field.String("value").Comment("资源值"),
		field.String("method").Comment("对资源的操作"),
		field.String("description").Comment("资源的描述"),
	}
}

// Mixin .
func (Resource) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Edges of the Resource.
func (Resource) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("roles", Role.Type),
	}
}
