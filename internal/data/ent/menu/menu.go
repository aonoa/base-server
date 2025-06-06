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
	// FieldPath holds the string denoting the path field in the database.
	FieldPath = "path"
	// FieldRedirect holds the string denoting the redirect field in the database.
	FieldRedirect = "redirect"
	// FieldAlias holds the string denoting the alias field in the database.
	FieldAlias = "alias"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldComponent holds the string denoting the component field in the database.
	FieldComponent = "component"
	// FieldIcon holds the string denoting the icon field in the database.
	FieldIcon = "icon"
	// FieldTitle holds the string denoting the title field in the database.
	FieldTitle = "title"
	// FieldOrder holds the string denoting the order field in the database.
	FieldOrder = "order"
	// FieldOpenInNewWindow holds the string denoting the openinnewwindow field in the database.
	FieldOpenInNewWindow = "open_in_new_window"
	// FieldNoBasicLayout holds the string denoting the nobasiclayout field in the database.
	FieldNoBasicLayout = "no_basic_layout"
	// FieldMenuVisibleWithForbidden holds the string denoting the menuvisiblewithforbidden field in the database.
	FieldMenuVisibleWithForbidden = "menu_visible_with_forbidden"
	// FieldLink holds the string denoting the link field in the database.
	FieldLink = "link"
	// FieldIframeSrc holds the string denoting the iframesrc field in the database.
	FieldIframeSrc = "iframe_src"
	// FieldActiveIcon holds the string denoting the activeicon field in the database.
	FieldActiveIcon = "active_icon"
	// FieldActivePath holds the string denoting the activepath field in the database.
	FieldActivePath = "active_path"
	// FieldMaxNumOfOpenTab holds the string denoting the maxnumofopentab field in the database.
	FieldMaxNumOfOpenTab = "max_num_of_open_tab"
	// FieldKeepalive holds the string denoting the keepalive field in the database.
	FieldKeepalive = "keepalive"
	// FieldIgnoreAccess holds the string denoting the ignoreaccess field in the database.
	FieldIgnoreAccess = "ignore_access"
	// FieldAuthority holds the string denoting the authority field in the database.
	FieldAuthority = "authority"
	// FieldAffixTab holds the string denoting the affixtab field in the database.
	FieldAffixTab = "affix_tab"
	// FieldAffixTabOrder holds the string denoting the affixtaborder field in the database.
	FieldAffixTabOrder = "affix_tab_order"
	// FieldHideInMenu holds the string denoting the hideinmenu field in the database.
	FieldHideInMenu = "hide_in_menu"
	// FieldHideInTab holds the string denoting the hideintab field in the database.
	FieldHideInTab = "hide_in_tab"
	// FieldHideInBreadcrumb holds the string denoting the hideinbreadcrumb field in the database.
	FieldHideInBreadcrumb = "hide_in_breadcrumb"
	// FieldHideChildrenInMenu holds the string denoting the hidechildreninmenu field in the database.
	FieldHideChildrenInMenu = "hide_children_in_menu"
	// FieldFullPathKey holds the string denoting the fullpathkey field in the database.
	FieldFullPathKey = "full_path_key"
	// FieldBadge holds the string denoting the badge field in the database.
	FieldBadge = "badge"
	// FieldBadgeType holds the string denoting the badgetype field in the database.
	FieldBadgeType = "badge_type"
	// FieldBadgeVariants holds the string denoting the badgevariants field in the database.
	FieldBadgeVariants = "badge_variants"
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
	FieldPath,
	FieldRedirect,
	FieldAlias,
	FieldName,
	FieldComponent,
	FieldIcon,
	FieldTitle,
	FieldOrder,
	FieldOpenInNewWindow,
	FieldNoBasicLayout,
	FieldMenuVisibleWithForbidden,
	FieldLink,
	FieldIframeSrc,
	FieldActiveIcon,
	FieldActivePath,
	FieldMaxNumOfOpenTab,
	FieldKeepalive,
	FieldIgnoreAccess,
	FieldAuthority,
	FieldAffixTab,
	FieldAffixTabOrder,
	FieldHideInMenu,
	FieldHideInTab,
	FieldHideInBreadcrumb,
	FieldHideChildrenInMenu,
	FieldFullPathKey,
	FieldBadge,
	FieldBadgeType,
	FieldBadgeVariants,
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
	// DefaultStatus holds the default value on creation for the "status" field.
	DefaultStatus bool
	// DefaultRedirect holds the default value on creation for the "redirect" field.
	DefaultRedirect string
	// DefaultAlias holds the default value on creation for the "alias" field.
	DefaultAlias string
	// DefaultIcon holds the default value on creation for the "icon" field.
	DefaultIcon string
	// DefaultOrder holds the default value on creation for the "order" field.
	DefaultOrder int32
	// DefaultOpenInNewWindow holds the default value on creation for the "openInNewWindow" field.
	DefaultOpenInNewWindow bool
	// DefaultNoBasicLayout holds the default value on creation for the "noBasicLayout" field.
	DefaultNoBasicLayout bool
	// DefaultMenuVisibleWithForbidden holds the default value on creation for the "menuVisibleWithForbidden" field.
	DefaultMenuVisibleWithForbidden bool
	// DefaultLink holds the default value on creation for the "link" field.
	DefaultLink string
	// DefaultIframeSrc holds the default value on creation for the "iframeSrc" field.
	DefaultIframeSrc string
	// DefaultActiveIcon holds the default value on creation for the "activeIcon" field.
	DefaultActiveIcon string
	// DefaultActivePath holds the default value on creation for the "activePath" field.
	DefaultActivePath string
	// DefaultMaxNumOfOpenTab holds the default value on creation for the "maxNumOfOpenTab" field.
	DefaultMaxNumOfOpenTab int16
	// DefaultKeepalive holds the default value on creation for the "keepalive" field.
	DefaultKeepalive bool
	// DefaultIgnoreAccess holds the default value on creation for the "ignoreAccess" field.
	DefaultIgnoreAccess bool
	// DefaultAuthority holds the default value on creation for the "authority" field.
	DefaultAuthority string
	// DefaultAffixTab holds the default value on creation for the "affixTab" field.
	DefaultAffixTab bool
	// DefaultAffixTabOrder holds the default value on creation for the "affixTabOrder" field.
	DefaultAffixTabOrder int16
	// DefaultHideInMenu holds the default value on creation for the "hideInMenu" field.
	DefaultHideInMenu bool
	// DefaultHideInTab holds the default value on creation for the "hideInTab" field.
	DefaultHideInTab bool
	// DefaultHideInBreadcrumb holds the default value on creation for the "hideInBreadcrumb" field.
	DefaultHideInBreadcrumb bool
	// DefaultHideChildrenInMenu holds the default value on creation for the "hideChildrenInMenu" field.
	DefaultHideChildrenInMenu bool
	// DefaultFullPathKey holds the default value on creation for the "fullPathKey" field.
	DefaultFullPathKey bool
	// DefaultBadge holds the default value on creation for the "badge" field.
	DefaultBadge string
	// DefaultBadgeType holds the default value on creation for the "badgeType" field.
	DefaultBadgeType string
	// DefaultBadgeVariants holds the default value on creation for the "badgeVariants" field.
	DefaultBadgeVariants string
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

// ByPath orders the results by the path field.
func ByPath(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPath, opts...).ToFunc()
}

// ByRedirect orders the results by the redirect field.
func ByRedirect(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldRedirect, opts...).ToFunc()
}

// ByAlias orders the results by the alias field.
func ByAlias(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAlias, opts...).ToFunc()
}

// ByName orders the results by the name field.
func ByName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldName, opts...).ToFunc()
}

// ByComponent orders the results by the component field.
func ByComponent(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldComponent, opts...).ToFunc()
}

// ByIcon orders the results by the icon field.
func ByIcon(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIcon, opts...).ToFunc()
}

// ByTitle orders the results by the title field.
func ByTitle(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldTitle, opts...).ToFunc()
}

// ByOrder orders the results by the order field.
func ByOrder(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldOrder, opts...).ToFunc()
}

// ByOpenInNewWindow orders the results by the openInNewWindow field.
func ByOpenInNewWindow(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldOpenInNewWindow, opts...).ToFunc()
}

// ByNoBasicLayout orders the results by the noBasicLayout field.
func ByNoBasicLayout(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldNoBasicLayout, opts...).ToFunc()
}

// ByMenuVisibleWithForbidden orders the results by the menuVisibleWithForbidden field.
func ByMenuVisibleWithForbidden(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldMenuVisibleWithForbidden, opts...).ToFunc()
}

// ByLink orders the results by the link field.
func ByLink(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldLink, opts...).ToFunc()
}

// ByIframeSrc orders the results by the iframeSrc field.
func ByIframeSrc(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIframeSrc, opts...).ToFunc()
}

// ByActiveIcon orders the results by the activeIcon field.
func ByActiveIcon(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldActiveIcon, opts...).ToFunc()
}

// ByActivePath orders the results by the activePath field.
func ByActivePath(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldActivePath, opts...).ToFunc()
}

// ByMaxNumOfOpenTab orders the results by the maxNumOfOpenTab field.
func ByMaxNumOfOpenTab(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldMaxNumOfOpenTab, opts...).ToFunc()
}

// ByKeepalive orders the results by the keepalive field.
func ByKeepalive(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldKeepalive, opts...).ToFunc()
}

// ByIgnoreAccess orders the results by the ignoreAccess field.
func ByIgnoreAccess(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIgnoreAccess, opts...).ToFunc()
}

// ByAuthority orders the results by the authority field.
func ByAuthority(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAuthority, opts...).ToFunc()
}

// ByAffixTab orders the results by the affixTab field.
func ByAffixTab(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAffixTab, opts...).ToFunc()
}

// ByAffixTabOrder orders the results by the affixTabOrder field.
func ByAffixTabOrder(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAffixTabOrder, opts...).ToFunc()
}

// ByHideInMenu orders the results by the hideInMenu field.
func ByHideInMenu(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldHideInMenu, opts...).ToFunc()
}

// ByHideInTab orders the results by the hideInTab field.
func ByHideInTab(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldHideInTab, opts...).ToFunc()
}

// ByHideInBreadcrumb orders the results by the hideInBreadcrumb field.
func ByHideInBreadcrumb(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldHideInBreadcrumb, opts...).ToFunc()
}

// ByHideChildrenInMenu orders the results by the hideChildrenInMenu field.
func ByHideChildrenInMenu(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldHideChildrenInMenu, opts...).ToFunc()
}

// ByFullPathKey orders the results by the fullPathKey field.
func ByFullPathKey(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldFullPathKey, opts...).ToFunc()
}

// ByBadge orders the results by the badge field.
func ByBadge(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldBadge, opts...).ToFunc()
}

// ByBadgeType orders the results by the badgeType field.
func ByBadgeType(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldBadgeType, opts...).ToFunc()
}

// ByBadgeVariants orders the results by the badgeVariants field.
func ByBadgeVariants(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldBadgeVariants, opts...).ToFunc()
}
