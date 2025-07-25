// Code generated by ent, DO NOT EDIT.

package ent

import (
	"base-server/internal/data/ent/menu"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
)

// Menu is the model entity for the Menu schema.
type Menu struct {
	config `json:"-"`
	// ID of the ent.
	ID int64 `json:"id,omitempty"`
	// CreateTime holds the value of the "create_time" field.
	CreateTime time.Time `json:"create_time,omitempty"`
	// UpdateTime holds the value of the "update_time" field.
	UpdateTime time.Time `json:"update_time,omitempty"`
	// 上一级ID
	Pid int64 `json:"pid,omitempty"`
	// 菜单类型,catalog-目录，menu-菜单，embedded-内嵌，link-外链，button-按钮
	Type string `json:"type,omitempty"`
	// 状态,0-禁用，1-启用
	Status bool `json:"status,omitempty"`
	// 路由path
	Path string `json:"path,omitempty"`
	// 重定向path
	Redirect string `json:"redirect,omitempty"`
	// 记录的别名，例如/users/:id和/u/:id。所有alias和path值必须共享相同的参数。
	Alias string `json:"alias,omitempty"`
	// 组件名
	Name string `json:"name,omitempty"`
	// 组件路径
	Component string `json:"component,omitempty"`
	// 图标
	Icon string `json:"icon,omitempty"`
	// 显示名称
	Title string `json:"title,omitempty"`
	// 排序(越小越前)
	Order int32 `json:"order,omitempty"`
	// 在新窗口打开
	OpenInNewWindow bool `json:"openInNewWindow,omitempty"`
	// 不使用基础布局（仅在顶级生效）
	NoBasicLayout bool `json:"noBasicLayout,omitempty"`
	// 菜单可以看到，但是访问会被重定向到403
	MenuVisibleWithForbidden bool `json:"menuVisibleWithForbidden,omitempty"`
	// 外链-跳转路径
	Link string `json:"link,omitempty"`
	// iframe地址
	IframeSrc string `json:"iframeSrc,omitempty"`
	// 激活图标（菜单）
	ActiveIcon string `json:"activeIcon,omitempty"`
	// 当前激活的菜单，有时候不想激活现有菜单，需要激活父级菜单时使用
	ActivePath string `json:"activePath,omitempty"`
	// 标签页最大打开数量
	MaxNumOfOpenTab int16 `json:"maxNumOfOpenTab,omitempty"`
	// 缓存,0-否，1-是
	Keepalive bool `json:"keepalive,omitempty"`
	// 忽略权限，直接可以访问
	IgnoreAccess bool `json:"ignoreAccess,omitempty"`
	// 需要特定的角色标识才可以访问(数组，','分割)
	Authority string `json:"authority,omitempty"`
	// 是否固定标签页,0-否，1-是
	AffixTab bool `json:"affixTab,omitempty"`
	// 固定标签页的顺序
	AffixTabOrder int16 `json:"affixTabOrder,omitempty"`
	// 当前路由在菜单中不展现,0-否，1-是
	HideInMenu bool `json:"hideInMenu,omitempty"`
	// 当前路由在标签页不展现,0-否，1-是
	HideInTab bool `json:"hideInTab,omitempty"`
	// 当前路由在面包屑中不展现,0-否，1-是
	HideInBreadcrumb bool `json:"hideInBreadcrumb,omitempty"`
	// 当前路由的子级在菜单中不展现,0-否，1-是
	HideChildrenInMenu bool `json:"hideChildrenInMenu,omitempty"`
	// 路由的完整路径作为key（默认true）
	FullPathKey bool `json:"fullPathKey,omitempty"`
	// 用于配置页面的徽标，会在菜单显示
	Badge string `json:"badge,omitempty"`
	// 用于配置页面的徽标类型，dot 为小红点，normal 为文本
	BadgeType string `json:"badgeType,omitempty"`
	// 用于配置页面的徽标颜色。类型：'default' | 'destructive' | 'primary' | 'success' | 'warning' | string
	BadgeVariants string `json:"badgeVariants,omitempty"`
	selectValues  sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Menu) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case menu.FieldStatus, menu.FieldOpenInNewWindow, menu.FieldNoBasicLayout, menu.FieldMenuVisibleWithForbidden, menu.FieldKeepalive, menu.FieldIgnoreAccess, menu.FieldAffixTab, menu.FieldHideInMenu, menu.FieldHideInTab, menu.FieldHideInBreadcrumb, menu.FieldHideChildrenInMenu, menu.FieldFullPathKey:
			values[i] = new(sql.NullBool)
		case menu.FieldID, menu.FieldPid, menu.FieldOrder, menu.FieldMaxNumOfOpenTab, menu.FieldAffixTabOrder:
			values[i] = new(sql.NullInt64)
		case menu.FieldType, menu.FieldPath, menu.FieldRedirect, menu.FieldAlias, menu.FieldName, menu.FieldComponent, menu.FieldIcon, menu.FieldTitle, menu.FieldLink, menu.FieldIframeSrc, menu.FieldActiveIcon, menu.FieldActivePath, menu.FieldAuthority, menu.FieldBadge, menu.FieldBadgeType, menu.FieldBadgeVariants:
			values[i] = new(sql.NullString)
		case menu.FieldCreateTime, menu.FieldUpdateTime:
			values[i] = new(sql.NullTime)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Menu fields.
func (m *Menu) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case menu.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			m.ID = int64(value.Int64)
		case menu.FieldCreateTime:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field create_time", values[i])
			} else if value.Valid {
				m.CreateTime = value.Time
			}
		case menu.FieldUpdateTime:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field update_time", values[i])
			} else if value.Valid {
				m.UpdateTime = value.Time
			}
		case menu.FieldPid:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field pid", values[i])
			} else if value.Valid {
				m.Pid = value.Int64
			}
		case menu.FieldType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field type", values[i])
			} else if value.Valid {
				m.Type = value.String
			}
		case menu.FieldStatus:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field status", values[i])
			} else if value.Valid {
				m.Status = value.Bool
			}
		case menu.FieldPath:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field path", values[i])
			} else if value.Valid {
				m.Path = value.String
			}
		case menu.FieldRedirect:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field redirect", values[i])
			} else if value.Valid {
				m.Redirect = value.String
			}
		case menu.FieldAlias:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field alias", values[i])
			} else if value.Valid {
				m.Alias = value.String
			}
		case menu.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				m.Name = value.String
			}
		case menu.FieldComponent:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field component", values[i])
			} else if value.Valid {
				m.Component = value.String
			}
		case menu.FieldIcon:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field icon", values[i])
			} else if value.Valid {
				m.Icon = value.String
			}
		case menu.FieldTitle:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field title", values[i])
			} else if value.Valid {
				m.Title = value.String
			}
		case menu.FieldOrder:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field order", values[i])
			} else if value.Valid {
				m.Order = int32(value.Int64)
			}
		case menu.FieldOpenInNewWindow:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field openInNewWindow", values[i])
			} else if value.Valid {
				m.OpenInNewWindow = value.Bool
			}
		case menu.FieldNoBasicLayout:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field noBasicLayout", values[i])
			} else if value.Valid {
				m.NoBasicLayout = value.Bool
			}
		case menu.FieldMenuVisibleWithForbidden:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field menuVisibleWithForbidden", values[i])
			} else if value.Valid {
				m.MenuVisibleWithForbidden = value.Bool
			}
		case menu.FieldLink:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field link", values[i])
			} else if value.Valid {
				m.Link = value.String
			}
		case menu.FieldIframeSrc:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field iframeSrc", values[i])
			} else if value.Valid {
				m.IframeSrc = value.String
			}
		case menu.FieldActiveIcon:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field activeIcon", values[i])
			} else if value.Valid {
				m.ActiveIcon = value.String
			}
		case menu.FieldActivePath:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field activePath", values[i])
			} else if value.Valid {
				m.ActivePath = value.String
			}
		case menu.FieldMaxNumOfOpenTab:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field maxNumOfOpenTab", values[i])
			} else if value.Valid {
				m.MaxNumOfOpenTab = int16(value.Int64)
			}
		case menu.FieldKeepalive:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field keepalive", values[i])
			} else if value.Valid {
				m.Keepalive = value.Bool
			}
		case menu.FieldIgnoreAccess:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field ignoreAccess", values[i])
			} else if value.Valid {
				m.IgnoreAccess = value.Bool
			}
		case menu.FieldAuthority:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field authority", values[i])
			} else if value.Valid {
				m.Authority = value.String
			}
		case menu.FieldAffixTab:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field affixTab", values[i])
			} else if value.Valid {
				m.AffixTab = value.Bool
			}
		case menu.FieldAffixTabOrder:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field affixTabOrder", values[i])
			} else if value.Valid {
				m.AffixTabOrder = int16(value.Int64)
			}
		case menu.FieldHideInMenu:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field hideInMenu", values[i])
			} else if value.Valid {
				m.HideInMenu = value.Bool
			}
		case menu.FieldHideInTab:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field hideInTab", values[i])
			} else if value.Valid {
				m.HideInTab = value.Bool
			}
		case menu.FieldHideInBreadcrumb:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field hideInBreadcrumb", values[i])
			} else if value.Valid {
				m.HideInBreadcrumb = value.Bool
			}
		case menu.FieldHideChildrenInMenu:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field hideChildrenInMenu", values[i])
			} else if value.Valid {
				m.HideChildrenInMenu = value.Bool
			}
		case menu.FieldFullPathKey:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field fullPathKey", values[i])
			} else if value.Valid {
				m.FullPathKey = value.Bool
			}
		case menu.FieldBadge:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field badge", values[i])
			} else if value.Valid {
				m.Badge = value.String
			}
		case menu.FieldBadgeType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field badgeType", values[i])
			} else if value.Valid {
				m.BadgeType = value.String
			}
		case menu.FieldBadgeVariants:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field badgeVariants", values[i])
			} else if value.Valid {
				m.BadgeVariants = value.String
			}
		default:
			m.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Menu.
// This includes values selected through modifiers, order, etc.
func (m *Menu) Value(name string) (ent.Value, error) {
	return m.selectValues.Get(name)
}

// Update returns a builder for updating this Menu.
// Note that you need to call Menu.Unwrap() before calling this method if this Menu
// was returned from a transaction, and the transaction was committed or rolled back.
func (m *Menu) Update() *MenuUpdateOne {
	return NewMenuClient(m.config).UpdateOne(m)
}

// Unwrap unwraps the Menu entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (m *Menu) Unwrap() *Menu {
	_tx, ok := m.config.driver.(*txDriver)
	if !ok {
		panic("ent: Menu is not a transactional entity")
	}
	m.config.driver = _tx.drv
	return m
}

// String implements the fmt.Stringer.
func (m *Menu) String() string {
	var builder strings.Builder
	builder.WriteString("Menu(")
	builder.WriteString(fmt.Sprintf("id=%v, ", m.ID))
	builder.WriteString("create_time=")
	builder.WriteString(m.CreateTime.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("update_time=")
	builder.WriteString(m.UpdateTime.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("pid=")
	builder.WriteString(fmt.Sprintf("%v", m.Pid))
	builder.WriteString(", ")
	builder.WriteString("type=")
	builder.WriteString(m.Type)
	builder.WriteString(", ")
	builder.WriteString("status=")
	builder.WriteString(fmt.Sprintf("%v", m.Status))
	builder.WriteString(", ")
	builder.WriteString("path=")
	builder.WriteString(m.Path)
	builder.WriteString(", ")
	builder.WriteString("redirect=")
	builder.WriteString(m.Redirect)
	builder.WriteString(", ")
	builder.WriteString("alias=")
	builder.WriteString(m.Alias)
	builder.WriteString(", ")
	builder.WriteString("name=")
	builder.WriteString(m.Name)
	builder.WriteString(", ")
	builder.WriteString("component=")
	builder.WriteString(m.Component)
	builder.WriteString(", ")
	builder.WriteString("icon=")
	builder.WriteString(m.Icon)
	builder.WriteString(", ")
	builder.WriteString("title=")
	builder.WriteString(m.Title)
	builder.WriteString(", ")
	builder.WriteString("order=")
	builder.WriteString(fmt.Sprintf("%v", m.Order))
	builder.WriteString(", ")
	builder.WriteString("openInNewWindow=")
	builder.WriteString(fmt.Sprintf("%v", m.OpenInNewWindow))
	builder.WriteString(", ")
	builder.WriteString("noBasicLayout=")
	builder.WriteString(fmt.Sprintf("%v", m.NoBasicLayout))
	builder.WriteString(", ")
	builder.WriteString("menuVisibleWithForbidden=")
	builder.WriteString(fmt.Sprintf("%v", m.MenuVisibleWithForbidden))
	builder.WriteString(", ")
	builder.WriteString("link=")
	builder.WriteString(m.Link)
	builder.WriteString(", ")
	builder.WriteString("iframeSrc=")
	builder.WriteString(m.IframeSrc)
	builder.WriteString(", ")
	builder.WriteString("activeIcon=")
	builder.WriteString(m.ActiveIcon)
	builder.WriteString(", ")
	builder.WriteString("activePath=")
	builder.WriteString(m.ActivePath)
	builder.WriteString(", ")
	builder.WriteString("maxNumOfOpenTab=")
	builder.WriteString(fmt.Sprintf("%v", m.MaxNumOfOpenTab))
	builder.WriteString(", ")
	builder.WriteString("keepalive=")
	builder.WriteString(fmt.Sprintf("%v", m.Keepalive))
	builder.WriteString(", ")
	builder.WriteString("ignoreAccess=")
	builder.WriteString(fmt.Sprintf("%v", m.IgnoreAccess))
	builder.WriteString(", ")
	builder.WriteString("authority=")
	builder.WriteString(m.Authority)
	builder.WriteString(", ")
	builder.WriteString("affixTab=")
	builder.WriteString(fmt.Sprintf("%v", m.AffixTab))
	builder.WriteString(", ")
	builder.WriteString("affixTabOrder=")
	builder.WriteString(fmt.Sprintf("%v", m.AffixTabOrder))
	builder.WriteString(", ")
	builder.WriteString("hideInMenu=")
	builder.WriteString(fmt.Sprintf("%v", m.HideInMenu))
	builder.WriteString(", ")
	builder.WriteString("hideInTab=")
	builder.WriteString(fmt.Sprintf("%v", m.HideInTab))
	builder.WriteString(", ")
	builder.WriteString("hideInBreadcrumb=")
	builder.WriteString(fmt.Sprintf("%v", m.HideInBreadcrumb))
	builder.WriteString(", ")
	builder.WriteString("hideChildrenInMenu=")
	builder.WriteString(fmt.Sprintf("%v", m.HideChildrenInMenu))
	builder.WriteString(", ")
	builder.WriteString("fullPathKey=")
	builder.WriteString(fmt.Sprintf("%v", m.FullPathKey))
	builder.WriteString(", ")
	builder.WriteString("badge=")
	builder.WriteString(m.Badge)
	builder.WriteString(", ")
	builder.WriteString("badgeType=")
	builder.WriteString(m.BadgeType)
	builder.WriteString(", ")
	builder.WriteString("badgeVariants=")
	builder.WriteString(m.BadgeVariants)
	builder.WriteByte(')')
	return builder.String()
}

// Menus is a parsable slice of Menu.
type Menus []*Menu
