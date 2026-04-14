package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// UserRoleBinding holds the schema definition for the current business user-role binding.
type UserRoleBinding struct {
	ent.Schema
}

func (UserRoleBinding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("业务用户角色绑定表"),
		entsql.Annotation{
			Table: "sys_user_role_binding",
		},
	}
}

func (UserRoleBinding) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("user_id").Unique().Comment("用户ID"),
		field.Int64("role_id").Comment("角色ID"),
	}
}

func (UserRoleBinding) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

func (UserRoleBinding) Edges() []ent.Edge {
	return nil
}
