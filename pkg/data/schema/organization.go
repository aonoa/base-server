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

// Organization holds the schema definition for organization master data.
type Organization struct {
	ent.Schema
}

func (Organization) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("组织表"),
		entsql.Annotation{
			Table: "sys_organization",
		},
	}
}

func (Organization) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(uuid.NewString).
			Unique().
			Immutable().
			Comment("组织ID"),
		field.String("name").Comment("组织名称"),
		field.String("code").Comment("组织编码"),
		field.Int32("sort").Default(0).Comment("排序"),
		field.Bool("status").Default(true).Comment("0-禁用，1-启用"),
		field.String("desc").Default("").Comment("备注"),
		field.String("extension").Default("").Comment("扩展信息"),
	}
}

func (Organization) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
	}
}

func (Organization) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

func (Organization) Edges() []ent.Edge {
	return nil
}
