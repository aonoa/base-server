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

type ApiResources struct {
	ent.Schema
}

func (ApiResources) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment(""),
	}
}

func (ApiResources) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(uuid.NewString).
			Unique().
			Comment("数据唯一标识"),
		field.String("description").Comment("描述"),
		field.String("path").Comment("路径"),
		field.String("method").Comment("方法"),
		field.String("module").Comment("模块"),
		field.String("module_description").Comment("模块描述"),
		field.String("resources_group").Comment("资源组"),
	}
}

func (ApiResources) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

func (ApiResources) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("roles", Role.Type),
	}
}
