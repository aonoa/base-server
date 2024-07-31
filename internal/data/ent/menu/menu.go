// Code generated by ent, DO NOT EDIT.

package menu

import (
	"time"

	"entgo.io/ent/dialect/sql"
)

const (
	// Label holds the string label denoting the menu type in the database.
	Label = "menu"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldCreateTime holds the string denoting the create_time field in the database.
	FieldCreateTime = "create_time"
	// FieldUpdateTime holds the string denoting the update_time field in the database.
	FieldUpdateTime = "update_time"
	// FieldPid holds the string denoting the pid field in the database.
	FieldPid = "pid"
	// FieldType holds the string denoting the type field in the database.
	FieldType = "type"
	// FieldStatus holds the string denoting the status field in the database.
	FieldStatus = "status"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldTitle holds the string denoting the title field in the database.
	FieldTitle = "title"
	// FieldIcon holds the string denoting the icon field in the database.
	FieldIcon = "icon"
	// FieldOrder holds the string denoting the order field in the database.
	FieldOrder = "order"
	// FieldPath holds the string denoting the path field in the database.
	FieldPath = "path"
	// FieldComponent holds the string denoting the component field in the database.
	FieldComponent = "component"
	// FieldRedirect holds the string denoting the redirect field in the database.
	FieldRedirect = "redirect"
	// FieldLink holds the string denoting the link field in the database.
	FieldLink = "link"
	// FieldIframeSrc holds the string denoting the iframesrc field in the database.
	FieldIframeSrc = "iframe_src"
	// FieldIgnoreAuth holds the string denoting the ignore_auth field in the database.
	FieldIgnoreAuth = "ignore_auth"
	// FieldKeepalive holds the string denoting the keepalive field in the database.
	FieldKeepalive = "keepalive"
	// FieldPermission holds the string denoting the permission field in the database.
	FieldPermission = "permission"
	// FieldAffixTab holds the string denoting the affix_tab field in the database.
	FieldAffixTab = "affix_tab"
	// FieldHideInBreadcrumb holds the string denoting the hideinbreadcrumb field in the database.
	FieldHideInBreadcrumb = "hide_in_breadcrumb"
	// Table holds the table name of the menu in the database.
	Table = "menus"
)

// Columns holds all SQL columns for menu fields.
var Columns = []string{
	FieldID,
	FieldCreateTime,
	FieldUpdateTime,
	FieldPid,
	FieldType,
	FieldStatus,
	FieldName,
	FieldTitle,
	FieldIcon,
	FieldOrder,
	FieldPath,
	FieldComponent,
	FieldRedirect,
	FieldLink,
	FieldIframeSrc,
	FieldIgnoreAuth,
	FieldKeepalive,
	FieldPermission,
	FieldAffixTab,
	FieldHideInBreadcrumb,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

var (
	// DefaultCreateTime holds the default value on creation for the "create_time" field.
	DefaultCreateTime func() time.Time
	// DefaultUpdateTime holds the default value on creation for the "update_time" field.
	DefaultUpdateTime func() time.Time
	// UpdateDefaultUpdateTime holds the default value on update for the "update_time" field.
	UpdateDefaultUpdateTime func() time.Time
)

// OrderOption defines the ordering options for the Menu queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByCreateTime orders the results by the create_time field.
func ByCreateTime(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreateTime, opts...).ToFunc()
}

// ByUpdateTime orders the results by the update_time field.
func ByUpdateTime(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdateTime, opts...).ToFunc()
}

// ByPid orders the results by the pid field.
func ByPid(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPid, opts...).ToFunc()
}

// ByType orders the results by the type field.
func ByType(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldType, opts...).ToFunc()
}

// ByStatus orders the results by the status field.
func ByStatus(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStatus, opts...).ToFunc()
}

// ByName orders the results by the name field.
func ByName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldName, opts...).ToFunc()
}

// ByTitle orders the results by the title field.
func ByTitle(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldTitle, opts...).ToFunc()
}

// ByIcon orders the results by the icon field.
func ByIcon(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIcon, opts...).ToFunc()
}

// ByOrder orders the results by the order field.
func ByOrder(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldOrder, opts...).ToFunc()
}

// ByPath orders the results by the path field.
func ByPath(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPath, opts...).ToFunc()
}

// ByComponent orders the results by the component field.
func ByComponent(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldComponent, opts...).ToFunc()
}

// ByRedirect orders the results by the redirect field.
func ByRedirect(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldRedirect, opts...).ToFunc()
}

// ByLink orders the results by the link field.
func ByLink(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldLink, opts...).ToFunc()
}

// ByIframeSrc orders the results by the iframeSrc field.
func ByIframeSrc(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIframeSrc, opts...).ToFunc()
}

// ByIgnoreAuth orders the results by the ignore_auth field.
func ByIgnoreAuth(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIgnoreAuth, opts...).ToFunc()
}

// ByKeepalive orders the results by the keepalive field.
func ByKeepalive(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldKeepalive, opts...).ToFunc()
}

// ByPermission orders the results by the permission field.
func ByPermission(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPermission, opts...).ToFunc()
}

// ByAffixTab orders the results by the affix_tab field.
func ByAffixTab(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAffixTab, opts...).ToFunc()
}

// ByHideInBreadcrumb orders the results by the hideInBreadcrumb field.
func ByHideInBreadcrumb(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldHideInBreadcrumb, opts...).ToFunc()
}