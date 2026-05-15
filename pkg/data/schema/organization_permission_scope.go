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

// OrganizationPermissionScope records permission entries enabled for an organization.
type OrganizationPermissionScope struct {
	ent.Schema
}

func (OrganizationPermissionScope) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("组织可用权限范围表"),
		entsql.Annotation{
			Table: "sys_organization_permission_scope",
		},
	}
}

func (OrganizationPermissionScope) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			DefaultFunc(uuid.NewString).
			Unique().
			Immutable().
			Comment("范围记录ID"),
		field.String("organization_id").Comment("组织ID"),
		field.String("permission_type").Comment("权限类型: menu/resource"),
		field.String("permission_ref").Comment("权限引用ID"),
		field.String("created_by").Default("").Comment("创建人ID"),
	}
}

func (OrganizationPermissionScope) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("organization_id", "permission_type", "permission_ref").Unique(),
		index.Fields("organization_id", "permission_type"),
	}
}

func (OrganizationPermissionScope) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

func (OrganizationPermissionScope) Edges() []ent.Edge {
	return nil
}
