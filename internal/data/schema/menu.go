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
		field.String("type").Comment("菜单类型,catalog-目录，menu-菜单，embedded-内嵌，link-外链，button-按钮"),
		field.Bool("status").Default(false).Comment("状态,0-禁用，1-启用"),

		field.String("path").Comment("路由path"),
		field.String("redirect").Default("").Comment("重定向path"),
		field.String("alias").Default("").Comment("记录的别名，例如/users/:id和/u/:id。所有alias和path值必须共享相同的参数。"), //+
		field.String("name").Comment("组件名"),
		field.String("component").Comment("组件路径"),

		field.String("icon").Default("").Comment("图标"),
		field.String("title").Comment("显示名称"),
		//field.String("query").Comment("菜单所携带的参数"), //+
		field.Int32("order").Default(0).Comment("排序(越小越前)"),

		field.Bool("openInNewWindow").Default(false).Comment("在新窗口打开"),        //+
		field.Bool("noBasicLayout").Default(false).Comment("不使用基础布局（仅在顶级生效）"), //+

		field.Bool("menuVisibleWithForbidden").Default(false).Comment("菜单可以看到，但是访问会被重定向到403"), //+
		//field.Bool("loaded").Comment("路由是否已经加载过"),                                             //+

		field.String("link").Default("").Comment("外链-跳转路径"),
		field.String("iframeSrc").Default("").Comment("iframe地址"),

		field.String("activeIcon").Default("").Comment("激活图标（菜单）"),
		field.String("activePath").Default("").Comment("当前激活的菜单，有时候不想激活现有菜单，需要激活父级菜单时使用"),

		field.Int16("maxNumOfOpenTab").Default(-1).Comment("标签页最大打开数量"),

		field.Bool("keepalive").Default(false).Comment("缓存,0-否，1-是"),
		field.Bool("ignoreAccess").Default(false).Comment("忽略权限，直接可以访问"),

		field.String("authority").Default("").Comment("需要特定的角色标识才可以访问(数组，','分割)"),
		field.Bool("affixTab").Default(false).Comment("是否固定标签页,0-否，1-是"),
		field.Int16("affixTabOrder").Default(0).Comment("固定标签页的顺序"),
		field.Bool("hideInMenu").Default(false).Comment("当前路由在菜单中不展现,0-否，1-是"),
		field.Bool("hideInTab").Default(false).Comment("当前路由在标签页不展现,0-否，1-是"),
		field.Bool("hideInBreadcrumb").Default(false).Comment("当前路由在面包屑中不展现,0-否，1-是"),
		field.Bool("hideChildrenInMenu").Default(false).Comment("当前路由的子级在菜单中不展现,0-否，1-是"),
		field.Bool("fullPathKey").Default(true).Comment("路由的完整路径作为key（默认true）"),

		field.String("badge").Default("").Comment("用于配置页面的徽标，会在菜单显示"),
		field.String("badgeType").Default("normal").Comment("用于配置页面的徽标类型，dot 为小红点，normal 为文本"),
		field.String("badgeVariants").Default("success").Comment("用于配置页面的徽标颜色。类型：'default' | 'destructive' | 'primary' | 'success' | 'warning' | string"),

		//field.String("name").Comment("组件名"),
		//field.String("title").Comment("显示名称"),
		//field.String("icon").Comment("图标"),
		//field.Int32("order").Comment("排序(越小越前)"),
		//field.String("path").Comment("路由path"),
		//field.String("component").Comment("组件路径"),
		//field.String("redirect").Comment("重定向path"),

		//field.String("link").Comment("外链-跳转路径"),
		//field.String("iframeSrc").Comment("iframe地址"),

		//field.String("activeIcon").Comment("激活图标（菜单）"),
		//field.String("activePath").Comment("当前激活的菜单，有时候不想激活现有菜单，需要激活父级菜单时使用"),

		//field.Int16("maxNumOfOpenTab").Comment("标签页最大打开数量"),

		//field.Bool("ignoreAuth").Comment("忽略权限,0-否，1-是"),
		//field.Bool("keepalive").Comment("缓存,0-否，1-是"),

		//field.String("permission").Comment("权限标识"),
		//field.Bool("affixTab").Comment("固钉,0-否，1-是"),
		//field.Int64("affixTabOrder").Comment("固定标签页的顺序"),
		//field.Bool("hideInMenu").Default(false).Comment("隐藏在菜单,0-否，1-是"),
		//field.Bool("hideInTab").Default(false).Comment("隐藏在标签页,0-否，1-是"),
		//field.Bool("hideInBreadcrumb").Default(false).Comment("隐藏在面包屑,0-否，1-是"),
		//field.Bool("hideChildrenInMenu").Default(false).Comment("子页面隐藏在菜单中,0-否，1-是"),
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
