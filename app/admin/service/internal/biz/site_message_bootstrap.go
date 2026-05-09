package biz

import (
	"context"
	"fmt"

	v1 "base-server/api/gen/go/admin/service/v1"
	"base-server/pkg/data/ent"
)

const (
	siteMessageInboxMenuPath        = "/messages"
	siteMessageInboxMenuName        = "SiteMessageInbox"
	siteMessageInboxComponent       = "/_core/messages/inbox"
	siteMessageManageMenuPath       = "/system/site-message"
	siteMessageManageMenuName       = "SiteMessageManage"
	siteMessageManageComponent      = "/_core/messages/manage"
	siteMessageManageParentPath     = "/system"
	siteMessageManageParentName     = "System"
	siteMessageManageResourceGroup  = "site_message_manage"
	siteMessageManageResourceMethod = "(GET|POST|DELETE)"
)

func (uc *AdminUsecase) ensureBuiltinSiteMessageBootstrap(ctx context.Context) error {
	inboxMenu, manageMenu, err := uc.ensureBuiltinSiteMessageMenus(ctx)
	if err != nil {
		return err
	}
	manageResourceID, err := uc.ensureSiteMessageManageResource(ctx)
	if err != nil {
		return err
	}
	roleList, err := uc.repo.ListAllRoles(ctx)
	if err != nil {
		return err
	}
	for _, roleItem := range roleList {
		status := int32(0)
		if roleItem.Status {
			status = 1
		}
		nextMenus := normalizeSiteMessageMenuIDs(
			roleItem.Value,
			roleItem.Menus,
			int32(inboxMenu.ID),
			int32(manageMenu.ID),
		)
		currentResources := resourceIDsFromRole(roleItem)
		nextResources := normalizeSiteMessageResourceIDs(
			roleItem.Value,
			currentResources,
			manageResourceID,
		)
		if equalInt32Slices(nextMenus, roleItem.Menus) && equalStringSets(nextResources, currentResources) {
			continue
		}
		if _, err := uc.repo.UpdateRole(ctx, roleItem.ID, &v1.RoleListItem{
			Id:             fmt.Sprintf("%d", roleItem.ID),
			Name:           roleItem.Name,
			Value:          roleItem.Value,
			Status:         status,
			Remark:         roleItem.Desc,
			Permissions:    nextMenus,
			ApiPermissions: nextResources,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (uc *AdminUsecase) normalizeSiteMessageRoleRequest(ctx context.Context, req *v1.RoleListItem) error {
	if req == nil {
		return nil
	}
	inboxMenu, manageMenu, err := uc.ensureBuiltinSiteMessageMenus(ctx)
	if err != nil {
		return err
	}
	manageResourceID, err := uc.ensureSiteMessageManageResource(ctx)
	if err != nil {
		return err
	}
	req.Permissions = normalizeSiteMessageMenuIDs(
		req.Value,
		req.Permissions,
		int32(inboxMenu.ID),
		int32(manageMenu.ID),
	)
	req.ApiPermissions = normalizeSiteMessageResourceIDs(
		req.Value,
		req.ApiPermissions,
		manageResourceID,
	)
	return nil
}

func (uc *AdminUsecase) ensureBuiltinSiteMessageMenus(ctx context.Context) (*ent.Menu, *ent.Menu, error) {
	menuList, err := uc.repo.GetMenuList(ctx)
	if err != nil {
		return nil, nil, err
	}

	inboxMenu := findMenuByPathOrName(menuList, siteMessageInboxMenuPath, siteMessageInboxMenuName)
	if inboxMenu == nil {
		inboxMenu, err = uc.repo.CreateMenu(ctx, builtinSiteMessageMenu(0, siteMessageInboxMenuPath, siteMessageInboxMenuName, "站内信收件箱", true))
		if err != nil {
			return nil, nil, err
		}
	} else if shouldUpdateSiteMessageMenu(inboxMenu, 0, siteMessageInboxComponent, true, "站内信收件箱") {
		nextMenu := cloneSiteMessageMenu(inboxMenu)
		nextMenu.Pid = 0
		nextMenu.Component = siteMessageInboxComponent
		nextMenu.ActivePath = siteMessageInboxMenuPath
		nextMenu.HideInMenu = true
		nextMenu.Title = "站内信收件箱"
		nextMenu.Icon = "lucide:mail"
		inboxMenu, err = uc.repo.UpdateMenu(ctx, inboxMenu.ID, nextMenu)
		if err != nil {
			return nil, nil, err
		}
	}

	systemMenu := findMenuByPathOrName(menuList, siteMessageManageParentPath, siteMessageManageParentName)
	if systemMenu == nil {
		return nil, nil, fmt.Errorf("site message manage parent menu %s not found", siteMessageManageParentPath)
	}
	manageMenu := findMenuByPathOrName(menuList, siteMessageManageMenuPath, siteMessageManageMenuName)
	if manageMenu == nil {
		manageMenu, err = uc.repo.CreateMenu(ctx, builtinSiteMessageMenu(systemMenu.ID, siteMessageManageMenuPath, siteMessageManageMenuName, "站内信管理", false))
		if err != nil {
			return nil, nil, err
		}
	} else if shouldUpdateSiteMessageMenu(manageMenu, systemMenu.ID, siteMessageManageComponent, false, "站内信管理") {
		nextMenu := cloneSiteMessageMenu(manageMenu)
		nextMenu.Pid = systemMenu.ID
		nextMenu.Component = siteMessageManageComponent
		nextMenu.ActivePath = siteMessageManageMenuPath
		nextMenu.HideInMenu = false
		nextMenu.Title = "站内信管理"
		nextMenu.Icon = "lucide:mail"
		manageMenu, err = uc.repo.UpdateMenu(ctx, manageMenu.ID, nextMenu)
		if err != nil {
			return nil, nil, err
		}
	}

	return inboxMenu, manageMenu, nil
}

func (uc *AdminUsecase) ensureSiteMessageManageResource(ctx context.Context) (string, error) {
	items, _, err := uc.repo.GetResourceList(ctx, &v1.GetResourcePageParams{
		CurrentPage: 1,
		PageSize:    50,
		Type:        "api",
		Value:       siteMessageManageResourceGroup,
	})
	if err != nil {
		return "", err
	}
	if len(items) > 0 {
		return items[0].ID, nil
	}
	item, err := uc.repo.AddResource(ctx, &ent.Resource{
		Name:        "站内信管理接口权限",
		Type:        "api",
		Value:       siteMessageManageResourceGroup,
		Method:      siteMessageManageResourceMethod,
		Description: "站内信管理页面接口权限",
	})
	if err != nil {
		return "", err
	}
	return item.ID, nil
}

func builtinSiteMessageMenu(pid int64, path, name, title string, hideInMenu bool) *ent.Menu {
	component := siteMessageInboxComponent
	order := int32(9999)
	if path == siteMessageManageMenuPath {
		component = siteMessageManageComponent
		order = 1008
	}
	return &ent.Menu{
		Pid:                      pid,
		Type:                     "menu",
		Status:                   true,
		Path:                     path,
		Redirect:                 "",
		Alias:                    "",
		Name:                     name,
		Component:                component,
		Icon:                     "lucide:mail",
		Title:                    title,
		Order:                    order,
		OpenInNewWindow:          false,
		NoBasicLayout:            false,
		MenuVisibleWithForbidden: false,
		Link:                     "",
		IframeSrc:                "",
		ActiveIcon:               "",
		ActivePath:               path,
		MaxNumOfOpenTab:          0,
		Keepalive:                false,
		IgnoreAccess:             false,
		Authority:                "",
		AffixTab:                 false,
		AffixTabOrder:            0,
		HideInMenu:               hideInMenu,
		HideInTab:                false,
		HideInBreadcrumb:         false,
		HideChildrenInMenu:       false,
		FullPathKey:              true,
		Badge:                    "",
		BadgeType:                "normal",
		BadgeVariants:            "success",
	}
}

func cloneSiteMessageMenu(item *ent.Menu) *ent.Menu {
	if item == nil {
		return nil
	}
	next := *item
	return &next
}

func shouldUpdateSiteMessageMenu(item *ent.Menu, pid int64, component string, hideInMenu bool, title string) bool {
	if item == nil {
		return false
	}
	return item.Pid != pid ||
		item.Component != component ||
		item.ActivePath != item.Path ||
		item.HideInMenu != hideInMenu ||
		item.Title != title ||
		item.Icon != "lucide:mail"
}

func findMenuByPathOrName(menuList []*ent.Menu, path, name string) *ent.Menu {
	for _, item := range menuList {
		if item.Path == path || item.Name == name {
			return item
		}
	}
	return nil
}

func normalizeSiteMessageMenuIDs(roleValue string, permissions []int32, inboxMenuID, manageMenuID int32) []int32 {
	if roleValue == "root" {
		return permissions
	}
	next := removeMenuIDs(permissions, inboxMenuID, manageMenuID)
	if roleValue == "default" || roleValue == "admin" {
		next = appendMenuID(next, inboxMenuID)
	}
	if roleValue == "admin" {
		next = appendMenuID(next, manageMenuID)
	}
	return next
}

func normalizeSiteMessageResourceIDs(roleValue string, apiPermissions []string, manageResourceID string) []string {
	if roleValue == "root" {
		return append([]string(nil), apiPermissions...)
	}
	next := removeStringIDs(apiPermissions, manageResourceID)
	if roleValue == "admin" {
		next = appendStringID(next, manageResourceID)
	}
	return next
}

func resourceIDsFromRole(item *ent.Role) []string {
	if item == nil || item.Edges.Resource == nil {
		return []string{}
	}
	items := make([]string, 0, len(item.Edges.Resource))
	for _, resourceItem := range item.Edges.Resource {
		if resourceItem == nil || resourceItem.ID == "" {
			continue
		}
		items = append(items, resourceItem.ID)
	}
	return items
}

func removeMenuIDs(menuIDs []int32, removeIDs ...int32) []int32 {
	removeSet := make(map[int32]struct{}, len(removeIDs))
	for _, removeID := range removeIDs {
		removeSet[removeID] = struct{}{}
	}
	next := make([]int32, 0, len(menuIDs))
	for _, menuID := range menuIDs {
		if _, exists := removeSet[menuID]; exists {
			continue
		}
		next = append(next, menuID)
	}
	return next
}

func appendMenuID(menuIDs []int32, menuID int32) []int32 {
	for _, item := range menuIDs {
		if item == menuID {
			return menuIDs
		}
	}
	next := make([]int32, 0, len(menuIDs)+1)
	next = append(next, menuIDs...)
	next = append(next, menuID)
	return next
}

func removeStringIDs(items []string, removeIDs ...string) []string {
	removeSet := make(map[string]struct{}, len(removeIDs))
	for _, item := range removeIDs {
		if item == "" {
			continue
		}
		removeSet[item] = struct{}{}
	}
	next := make([]string, 0, len(items))
	for _, item := range items {
		if _, exists := removeSet[item]; exists {
			continue
		}
		next = append(next, item)
	}
	return next
}

func appendStringID(items []string, value string) []string {
	if value == "" {
		return items
	}
	for _, item := range items {
		if item == value {
			return items
		}
	}
	next := make([]string, 0, len(items)+1)
	next = append(next, items...)
	next = append(next, value)
	return next
}

func equalInt32Slices(left, right []int32) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func equalStringSets(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	seen := make(map[string]int, len(left))
	for _, item := range left {
		seen[item]++
	}
	for _, item := range right {
		seen[item]--
	}
	for _, count := range seen {
		if count != 0 {
			return false
		}
	}
	return true
}
