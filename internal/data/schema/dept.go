package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Dept holds the schema definition for the Dept entity.
type Dept struct {
	ent.Schema
}

// Mixin .
func (Dept) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

func (Dept) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
	}
}

// Fields of the Dept.
func (Dept) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("name").Comment("部门名称"),
		field.Int32("sort").Comment("排序"),
		field.Bool("status").Comment("0-锁定，1-正常"),
		field.String("desc").Comment("备注"),
		field.String("extension").Comment("扩展信息"),
		field.Int64("dom").Comment("域"),
		field.Int64("pid").Comment("父节点id").
			Optional(),
	}
}

// Edges of the Dept.
func (Dept) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("users", User.Type),
		edge.To("roles", Role.Type).Unique(),
		edge.To("children", Dept.Type).
			From("parent").
			Unique().
			Field("pid"),
	}
}
