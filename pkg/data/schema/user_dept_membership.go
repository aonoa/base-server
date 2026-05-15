package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// UserDeptMembership holds the schema definition for organization scoped user-department membership.
type UserDeptMembership struct {
	ent.Schema
}

func (UserDeptMembership) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("用户部门绑定表"),
		entsql.Annotation{
			Table: "sys_user_dept_membership",
		},
	}
}

func (UserDeptMembership) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("user_id").Comment("用户ID"),
		field.Int64("dept_id").Comment("部门ID"),
		field.String("organization_id").Comment("组织ID"),
	}
}

func (UserDeptMembership) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "organization_id").Unique(),
		index.Fields("organization_id"),
		index.Fields("dept_id"),
	}
}

func (UserDeptMembership) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

func (UserDeptMembership) Edges() []ent.Edge {
	return nil
}
