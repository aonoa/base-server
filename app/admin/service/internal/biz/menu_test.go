package biz

import (
	"reflect"
	"testing"
	"time"

	v1 "base-server/api/gen/go/admin/service/v1"
	"base-server/pkg/data/ent"
)

func TestMenuToEntMenuMapsPersistedMetaFields(t *testing.T) {
	status := int32(1)
	redirect := "/dashboard"
	activeIcon := "carbon:home"
	activePath := "/system"
	affixTab := true
	affixTabOrder := int64(8)
	badge := "9"
	badgeType := "normal"
	badgeVariants := "warning"
	hideChildrenInMenu := true
	hideInBreadcrumb := true
	hideInMenu := true
	hideInTab := true
	iframeSrc := "https://example.com/frame"
	link := "https://example.com"
	keepAlive := true
	maxNumOfOpenTab := int64(3)
	noBasicLayout := true
	openInNewWindow := true
	fullPathKey := false
	menuVisibleWithForbidden := true

	item := menuToEntMenu(&v1.SysMenuListItem{
		Id:        10,
		Component: "views/system/menu/index",
		Status:    &status,
		AuthCode:  "legacy-auth",
		Name:      "SystemMenu",
		Path:      "/system/menu",
		Pid:       1,
		Redirect:  &redirect,
		Type:      "menu",
		Meta: &v1.Meta{
			Order:                    7,
			Icon:                     "carbon:menu",
			Title:                    "system.menu.title",
			ActiveIcon:               &activeIcon,
			ActivePath:               &activePath,
			AffixTab:                 &affixTab,
			AffixTabOrder:            &affixTabOrder,
			Badge:                    &badge,
			BadgeType:                &badgeType,
			BadgeVariants:            &badgeVariants,
			HideChildrenInMenu:       &hideChildrenInMenu,
			HideInBreadcrumb:         &hideInBreadcrumb,
			HideInMenu:               &hideInMenu,
			HideInTab:                &hideInTab,
			IframeSrc:                &iframeSrc,
			Link:                     &link,
			KeepAlive:                &keepAlive,
			MaxNumOfOpenTab:          &maxNumOfOpenTab,
			NoBasicLayout:            &noBasicLayout,
			OpenInNewWindow:          &openInNewWindow,
			IgnoreAccess:             true,
			Authority:                []string{"admin", " admin ", "", "operator"},
			FullPathKey:              &fullPathKey,
			MenuVisibleWithForbidden: &menuVisibleWithForbidden,
		},
	})

	if item.ID != 10 || item.Pid != 1 || item.Type != "menu" || !item.Status {
		t.Fatalf("base fields not mapped: %+v", item)
	}
	if item.Order != 7 || item.ActiveIcon != activeIcon || item.ActivePath != activePath {
		t.Fatalf("route display fields not mapped: %+v", item)
	}
	if item.AffixTab != affixTab || item.AffixTabOrder != int16(affixTabOrder) || item.MaxNumOfOpenTab != int16(maxNumOfOpenTab) {
		t.Fatalf("tab fields not mapped: %+v", item)
	}
	if item.Badge != badge || item.BadgeType != badgeType || item.BadgeVariants != badgeVariants {
		t.Fatalf("badge fields not mapped: %+v", item)
	}
	if !item.OpenInNewWindow || !item.NoBasicLayout || item.FullPathKey || !item.MenuVisibleWithForbidden || !item.IgnoreAccess {
		t.Fatalf("access/layout fields not mapped: %+v", item)
	}
	if !item.HideInMenu || !item.HideInTab || !item.HideInBreadcrumb || !item.HideChildrenInMenu {
		t.Fatalf("visibility fields not mapped: %+v", item)
	}
	if item.Authority != "admin,operator" {
		t.Fatalf("authority normalized incorrectly: %q", item.Authority)
	}
}

func TestMenuToEntMenuUsesDefaultsAndLegacyAuthCode(t *testing.T) {
	item := menuToEntMenu(&v1.SysMenuListItem{
		AuthCode: "legacy-a, legacy-b",
		Meta:     &v1.Meta{},
	})

	if item.MaxNumOfOpenTab != -1 {
		t.Fatalf("expected default max open tab -1, got %d", item.MaxNumOfOpenTab)
	}
	if !item.FullPathKey {
		t.Fatal("expected default fullPathKey true")
	}
	if item.BadgeType != "normal" || item.BadgeVariants != "success" {
		t.Fatalf("badge defaults not preserved: type=%q variants=%q", item.BadgeType, item.BadgeVariants)
	}
	if item.Authority != "legacy-a,legacy-b" {
		t.Fatalf("legacy authCode not mapped: %q", item.Authority)
	}
}

func TestEntMenuToMenuReturnsPersistedMetaFields(t *testing.T) {
	item := entMenuToMenu(&ent.Menu{
		ID:                       10,
		Pid:                      1,
		Type:                     "menu",
		Status:                   true,
		Path:                     "/system/menu",
		Redirect:                 "/dashboard",
		Name:                     "SystemMenu",
		Component:                "views/system/menu/index",
		Icon:                     "carbon:menu",
		Title:                    "system.menu.title",
		Order:                    7,
		OpenInNewWindow:          true,
		NoBasicLayout:            true,
		MenuVisibleWithForbidden: true,
		Link:                     "https://example.com",
		IframeSrc:                "https://example.com/frame",
		ActiveIcon:               "carbon:home",
		ActivePath:               "/system",
		MaxNumOfOpenTab:          3,
		Keepalive:                true,
		IgnoreAccess:             true,
		Authority:                "admin,operator",
		AffixTab:                 true,
		AffixTabOrder:            8,
		HideInMenu:               true,
		HideInTab:                true,
		HideInBreadcrumb:         true,
		HideChildrenInMenu:       true,
		FullPathKey:              false,
		Badge:                    "9",
		BadgeType:                "normal",
		BadgeVariants:            "warning",
		CreateTime:               time.Date(2026, 5, 15, 1, 2, 3, 0, time.Local),
	})

	if item.AuthCode != "admin,operator" {
		t.Fatalf("authCode compatibility field not mapped: %q", item.AuthCode)
	}
	if item.Status == nil || *item.Status != 1 {
		t.Fatalf("status not mapped: %+v", item.Status)
	}
	meta := item.Meta
	if meta == nil {
		t.Fatal("meta is nil")
	}
	if meta.Order != 7 || meta.ActiveIcon == nil || *meta.ActiveIcon != "carbon:home" {
		t.Fatalf("route display fields not returned: %+v", meta)
	}
	if meta.AffixTab == nil || !*meta.AffixTab || meta.AffixTabOrder == nil || *meta.AffixTabOrder != 8 {
		t.Fatalf("tab fields not returned: %+v", meta)
	}
	if meta.MaxNumOfOpenTab == nil || *meta.MaxNumOfOpenTab != 3 {
		t.Fatalf("maxNumOfOpenTab not returned: %+v", meta)
	}
	if meta.FullPathKey == nil || *meta.FullPathKey {
		t.Fatalf("fullPathKey not returned as false: %+v", meta)
	}
	if meta.MenuVisibleWithForbidden == nil || !*meta.MenuVisibleWithForbidden {
		t.Fatalf("menuVisibleWithForbidden not returned: %+v", meta)
	}
	if !reflect.DeepEqual(meta.Authority, []string{"admin", "operator"}) {
		t.Fatalf("authority not split: %#v", meta.Authority)
	}
}
