package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// Menu holds the schema definition for the Menu entity.
type Menu struct {
	ent.Schema
}

func (Menu) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
	}
}

// Fields of the Menu.
func (Menu) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("pid").Comment("上一级ID"),
		field.Int8("type").Comment("菜单类型,0-目录,1-菜单,2-按钮"),
		field.String("name").Comment("组件名"),
		field.String("title").Comment("显示名称"),
		field.Int8("sort").Unique().Comment("排序(越小越前)"),
		field.String("icon").Comment("图标"),
		field.String("path").Comment("路由path"),
		field.String("component").Comment("组件路径"),
		field.String("redirect").Comment("重定向path"),
		field.Bool("status").Comment("状态,0-禁用，1-启用"),
		field.String("permission").Comment("权限标识"),
		field.Bool("isext").Comment("外链,0-否，1-是"),
		field.Bool("isshow").Comment("显示,0-否，1-是"),
		field.Bool("keepalive").Comment("缓存,0-否，1-是"),
		field.Bool("affix").Comment("固钉,0-否，1-是"),
		field.Bool("ignore_auth").Comment("忽略权限,0-否，1-是"),
		field.Bool("breadcrumb").Comment("显示在面包屑,0-否，1-是"),
	}
}

// Edges of the Menu.
func (Menu) Edges() []ent.Edge {
	return nil
}

// Mixin .
func (Menu) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
		// Or, mixin.CreateTime only for create_time
		// and mixin.UpdateTime only for update_time.
	}
}
