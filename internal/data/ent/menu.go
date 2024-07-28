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
	// 菜单类型,0-目录,1-菜单,2-按钮
	Type int8 `json:"type,omitempty"`
	// 状态,0-禁用，1-启用
	Status bool `json:"status,omitempty"`
	// 组件名
	Name string `json:"name,omitempty"`
	// 显示名称
	Title string `json:"title,omitempty"`
	// 图标
	Icon string `json:"icon,omitempty"`
	// 排序(越小越前)
	Order int32 `json:"order,omitempty"`
	// 路由path
	Path string `json:"path,omitempty"`
	// 组件路径
	Component string `json:"component,omitempty"`
	// 重定向path
	Redirect string `json:"redirect,omitempty"`
	// 外链-跳转路径
	Link string `json:"link,omitempty"`
	// iframe地址
	IframeSrc string `json:"iframeSrc,omitempty"`
	// 忽略权限,0-否，1-是
	IgnoreAuth bool `json:"ignore_auth,omitempty"`
	// 缓存,0-否，1-是
	Keepalive bool `json:"keepalive,omitempty"`
	// 权限标识
	Permission string `json:"permission,omitempty"`
	// 固钉,0-否，1-是
	AffixTab bool `json:"affix_tab,omitempty"`
	// 显示在面包屑,0-否，1-是
	HideInBreadcrumb bool `json:"hideInBreadcrumb,omitempty"`
	selectValues     sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Menu) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case menu.FieldStatus, menu.FieldIgnoreAuth, menu.FieldKeepalive, menu.FieldAffixTab, menu.FieldHideInBreadcrumb:
			values[i] = new(sql.NullBool)
		case menu.FieldID, menu.FieldPid, menu.FieldType, menu.FieldOrder:
			values[i] = new(sql.NullInt64)
		case menu.FieldName, menu.FieldTitle, menu.FieldIcon, menu.FieldPath, menu.FieldComponent, menu.FieldRedirect, menu.FieldLink, menu.FieldIframeSrc, menu.FieldPermission:
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
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field type", values[i])
			} else if value.Valid {
				m.Type = int8(value.Int64)
			}
		case menu.FieldStatus:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field status", values[i])
			} else if value.Valid {
				m.Status = value.Bool
			}
		case menu.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				m.Name = value.String
			}
		case menu.FieldTitle:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field title", values[i])
			} else if value.Valid {
				m.Title = value.String
			}
		case menu.FieldIcon:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field icon", values[i])
			} else if value.Valid {
				m.Icon = value.String
			}
		case menu.FieldOrder:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field order", values[i])
			} else if value.Valid {
				m.Order = int32(value.Int64)
			}
		case menu.FieldPath:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field path", values[i])
			} else if value.Valid {
				m.Path = value.String
			}
		case menu.FieldComponent:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field component", values[i])
			} else if value.Valid {
				m.Component = value.String
			}
		case menu.FieldRedirect:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field redirect", values[i])
			} else if value.Valid {
				m.Redirect = value.String
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
		case menu.FieldIgnoreAuth:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field ignore_auth", values[i])
			} else if value.Valid {
				m.IgnoreAuth = value.Bool
			}
		case menu.FieldKeepalive:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field keepalive", values[i])
			} else if value.Valid {
				m.Keepalive = value.Bool
			}
		case menu.FieldPermission:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field permission", values[i])
			} else if value.Valid {
				m.Permission = value.String
			}
		case menu.FieldAffixTab:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field affix_tab", values[i])
			} else if value.Valid {
				m.AffixTab = value.Bool
			}
		case menu.FieldHideInBreadcrumb:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field hideInBreadcrumb", values[i])
			} else if value.Valid {
				m.HideInBreadcrumb = value.Bool
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
	builder.WriteString(fmt.Sprintf("%v", m.Type))
	builder.WriteString(", ")
	builder.WriteString("status=")
	builder.WriteString(fmt.Sprintf("%v", m.Status))
	builder.WriteString(", ")
	builder.WriteString("name=")
	builder.WriteString(m.Name)
	builder.WriteString(", ")
	builder.WriteString("title=")
	builder.WriteString(m.Title)
	builder.WriteString(", ")
	builder.WriteString("icon=")
	builder.WriteString(m.Icon)
	builder.WriteString(", ")
	builder.WriteString("order=")
	builder.WriteString(fmt.Sprintf("%v", m.Order))
	builder.WriteString(", ")
	builder.WriteString("path=")
	builder.WriteString(m.Path)
	builder.WriteString(", ")
	builder.WriteString("component=")
	builder.WriteString(m.Component)
	builder.WriteString(", ")
	builder.WriteString("redirect=")
	builder.WriteString(m.Redirect)
	builder.WriteString(", ")
	builder.WriteString("link=")
	builder.WriteString(m.Link)
	builder.WriteString(", ")
	builder.WriteString("iframeSrc=")
	builder.WriteString(m.IframeSrc)
	builder.WriteString(", ")
	builder.WriteString("ignore_auth=")
	builder.WriteString(fmt.Sprintf("%v", m.IgnoreAuth))
	builder.WriteString(", ")
	builder.WriteString("keepalive=")
	builder.WriteString(fmt.Sprintf("%v", m.Keepalive))
	builder.WriteString(", ")
	builder.WriteString("permission=")
	builder.WriteString(m.Permission)
	builder.WriteString(", ")
	builder.WriteString("affix_tab=")
	builder.WriteString(fmt.Sprintf("%v", m.AffixTab))
	builder.WriteString(", ")
	builder.WriteString("hideInBreadcrumb=")
	builder.WriteString(fmt.Sprintf("%v", m.HideInBreadcrumb))
	builder.WriteByte(')')
	return builder.String()
}

// Menus is a parsable slice of Menu.
type Menus []*Menu
