package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Role holds the schema definition for the Role entity.
type Role struct {
	ent.Schema
}

func (Role) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("角色表"),
		entsql.Annotation{
			Table: "sys_role",
		},
	}
}

// Fields of the Role.
func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("name").Comment("角色名称"),
		field.String("value").Comment("角色值"),
		field.String("organization_id").Comment("组织ID"),
		field.Bool("status").Comment("0-禁用，1-启用"),
		field.String("desc").Comment("简介"),
		field.JSON("menus", []int32{}).Comment("权限菜单ID列表"),
		field.String("data_scope").
			Default("self").
			Comment("数据范围: all/self_dept/self_dept_and_child/self/custom_depts"),
		field.JSON("data_scope_dept_ids", []int64{}).
			Optional().
			Default([]int64{}).
			Comment("自定义数据范围部门ID列表"),
	}
}

// Mixin .
func (Role) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Edges of the Role.
func (Role) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("api", ApiResources.Type).
			Ref("roles"),
		edge.From("resource", Resource.Type).
			Ref("roles"),
	}
}
