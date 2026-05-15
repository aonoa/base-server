package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// UserOrganization holds the schema definition for user organization membership.
type UserOrganization struct {
	ent.Schema
}

func (UserOrganization) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("用户组织成员表"),
		entsql.Annotation{
			Table: "sys_user_organization",
		},
	}
}

func (UserOrganization) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("user_id").Comment("用户ID"),
		field.String("organization_id").Comment("组织ID"),
		field.Bool("is_primary").Default(false).Comment("是否主组织"),
		field.Bool("status").Default(true).Comment("0-禁用，1-启用"),
	}
}

func (UserOrganization) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "organization_id").Unique(),
		index.Fields("organization_id"),
		index.Fields("user_id"),
	}
}

func (UserOrganization) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

func (UserOrganization) Edges() []ent.Edge {
	return nil
}
