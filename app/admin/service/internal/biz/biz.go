package biz

import (
	"context"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	authv1 "base-server/api/gen/go/auth/service/v1"
	v1 "base-server/api/gen/go/admin/service/v1"
	userv1 "base-server/api/gen/go/user/service/v1"
	"base-server/pkg/authx"
	"base-server/pkg/data/ent"
	"base-server/pkg/tools"

	"github.com/google/wire"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewAdminUsecase)

type AdminRepo interface {
	GetMenuList(context.Context) ([]*ent.Menu, error)
	GetUserAuthInfo(context.Context, string) (*userv1.GetUserAuthInfoReply, error)
	GetCurrentUserMenuAuthority(context.Context, string) (*authv1.GetCurrentUserMenuAuthorityReply, error)
	CreateMenu(context.Context, *ent.Menu) (*ent.Menu, error)
	UpdateMenu(context.Context, int64, *ent.Menu) (*ent.Menu, error)
	DeleteMenu(context.Context, int64) error
	GetDeptList(context.Context) ([]*ent.Dept, error)
	AddDept(context.Context, *v1.DeptListItem) (*ent.Dept, error)
	UpdateDept(context.Context, int64, *v1.DeptListItem) (*ent.Dept, error)
	DelDept(context.Context, int64) error
	GetDeptLeafsChildren(context.Context, int64) ([]*ent.Dept, error)
	GetDeptById(context.Context, int64) (*ent.Dept, error)
	CreateSysLog(context.Context, *ent.SysLogRecord) error
	GetSysLogList(context.Context, *v1.GetSysLogListParams) ([]*ent.SysLogRecord, int64, error)
	GetSysLogInfo(context.Context, string) (*ent.SysLogRecord, error)
}

type AdminUsecase struct {
	repo AdminRepo
}

func NewAdminUsecase(repo AdminRepo) *AdminUsecase {
	return &AdminUsecase{repo: repo}
}

func (uc *AdminUsecase) GetDeptList(ctx context.Context) (*v1.GetDeptListReply, error) {
	deptList, err := uc.repo.GetDeptList(ctx)
	if err != nil {
		return nil, err
	}
	forest := make([]*deptNode, 0)
	for _, dept := range deptList {
		if dept.ID == 0 {
			continue
		}
		uc.buildDeptTree(&forest, dept, true)
	}
	return &v1.GetDeptListReply{
		Items: toDeptTree(forest, ""),
		Total: int64(len(deptList)),
	}, nil
}

func (uc *AdminUsecase) AddDept(ctx context.Context, req *v1.DeptListItem) (*v1.DeptListItem, error) {
	dept, err := uc.repo.AddDept(ctx, req)
	if err != nil {
		return nil, err
	}
	return deptToReply(dept, req.Pid), nil
}

func (uc *AdminUsecase) UpdateDept(ctx context.Context, req *v1.DeptListItem) (*v1.DeptListItem, error) {
	deptID, err := tools.DeptStrSplitToInt(req.Id)
	if err != nil {
		return nil, err
	}
	dept, err := uc.repo.UpdateDept(ctx, deptID, req)
	if err != nil {
		return nil, err
	}
	return deptToReply(dept, req.Pid), nil
}

func (uc *AdminUsecase) DelDept(ctx context.Context, deptID string) error {
	id, err := tools.DeptStrSplitToInt(deptID)
	if err != nil {
		return err
	}
	dept, err := uc.repo.GetDeptById(ctx, id)
	if err != nil {
		return err
	}
	for {
		childrenList, err := uc.repo.GetDeptLeafsChildren(ctx, id)
		if err != nil {
			return err
		}
		if len(childrenList) == 0 {
			break
		}
		for _, child := range childrenList {
			if err := uc.repo.DelDept(ctx, child.ID); err != nil {
				return err
			}
		}
	}
	_ = dept
	return uc.repo.DelDept(ctx, id)
}

func (uc *AdminUsecase) GetCurrentUserMenus(ctx context.Context, userID string) (*v1.GetCurrentUserMenusReply, error) {
	_, err := uc.repo.GetUserAuthInfo(ctx, userID)
	if err != nil {
		return nil, err
	}
	authority, err := uc.repo.GetCurrentUserMenuAuthority(ctx, userID)
	if err != nil {
		return nil, err
	}
	menuList, err := uc.repo.GetMenuList(ctx)
	if err != nil {
		return nil, err
	}
	visibleMenus := filterVisibleMenus(menuList, authority)
	return &v1.GetCurrentUserMenusReply{Items: createCurrentUserMenuTree(visibleMenus)}, nil
}

func (uc *AdminUsecase) GetSysMenuList(ctx context.Context) (*v1.GetSysMenuListReply, error) {
	menuList, err := uc.repo.GetMenuList(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetSysMenuListReply{Items: uc.createMenuTree(menuList)}, nil
}

func (uc *AdminUsecase) ListMenus(ctx context.Context) (*v1.ListMenusReply, error) {
	menuList, err := uc.repo.GetMenuList(ctx)
	if err != nil {
		return nil, err
	}
	res := &v1.ListMenusReply{Items: make([]*v1.MenuRecord, 0, len(menuList))}
	for _, item := range menuList {
		res.Items = append(res.Items, entMenuToRecord(item))
	}
	return res, nil
}

func (uc *AdminUsecase) IsMenuNameExists(ctx context.Context, req *v1.IsMenuNameExistsRequest) (bool, error) {
	menuList, err := uc.repo.GetMenuList(ctx)
	if err != nil {
		return false, err
	}
	for _, menu := range menuList {
		if req.Name == menu.Name && menu.ID != req.Id {
			return true, nil
		}
	}
	return false, nil
}

func (uc *AdminUsecase) IsMenuPathExists(ctx context.Context, req *v1.IsMenuPathExistsRequest) (bool, error) {
	menuList, err := uc.repo.GetMenuList(ctx)
	if err != nil {
		return false, err
	}
	for _, menu := range menuList {
		if req.Path == menu.Path && menu.ID != req.Id {
			return true, nil
		}
	}
	return false, nil
}

func (uc *AdminUsecase) CreateMenu(ctx context.Context, req *v1.SysMenuListItem) error {
	_, err := uc.repo.CreateMenu(ctx, menuToEntMenu(req))
	return err
}

func (uc *AdminUsecase) UpdateMenu(ctx context.Context, req *v1.SysMenuListItem) error {
	_, err := uc.repo.UpdateMenu(ctx, int64(req.Id), menuToEntMenu(req))
	return err
}

func (uc *AdminUsecase) DeleteMenu(ctx context.Context, req *v1.DeleteMenuRequest) error {
	return uc.repo.DeleteMenu(ctx, req.Id)
}

func (uc *AdminUsecase) CreateSysLog(ctx context.Context, req *v1.CreateSysLogRequest) error {
	requestTime, err := time.Parse(time.DateTime, req.RequestTime)
	if err != nil {
		return err
	}
	return uc.repo.CreateSysLog(ctx, &ent.SysLogRecord{
		UserID:      req.UserId,
		UserName:    req.UserName,
		IsLogin:     req.IsLogin,
		SessionID:   req.SessionId,
		Method:      req.Method,
		Path:        req.Path,
		RequestTime: requestTime,
		IPAddress:   req.IpAddress,
		IPLocation:  req.IpLocation,
		Latency:     req.Latency,
		Os:          req.Os,
		Browser:     req.Browser,
		UserAgent:   req.UserAgent,
		Header:      req.Header,
		GetParams:   req.GetParams,
		PostData:    req.PostData,
		ResCode:     req.ResCode,
		Reason:      req.Reason,
		ResStatus:   req.ResStatus,
		Stack:       req.Stack,
	})
}

func (uc *AdminUsecase) GetSysLogList(ctx context.Context, req *v1.GetSysLogListParams) (*v1.GetSysLogListReply, error) {
	list, count, err := uc.repo.GetSysLogList(ctx, req)
	if err != nil {
		return nil, err
	}
	res := &v1.GetSysLogListReply{Items: make([]*v1.SysLogItem, 0, len(list)), Total: count}
	for _, item := range list {
		res.Items = append(res.Items, &v1.SysLogItem{
			Id:          item.ID,
			UserId:      item.UserID,
			UserName:    item.UserName,
			Method:      item.Method,
			Path:        item.Path,
			RequestTime: item.RequestTime.Format(time.DateTime),
			IpAddress:   item.IPAddress,
			IpLocation:  item.IPLocation,
			Latency:     item.Latency,
			Os:          item.Os,
			Browser:     item.Browser,
			ResCode:     item.ResCode,
			ResStatus:   item.ResStatus,
		})
	}
	return res, nil
}

func (uc *AdminUsecase) GetSysLogInfo(ctx context.Context, id string) (*v1.GetSysLogInfoReply, error) {
	info, err := uc.repo.GetSysLogInfo(ctx, id)
	if err != nil {
		return nil, err
	}
	return &v1.GetSysLogInfoReply{
		Id:          info.ID,
		UserId:      info.UserID,
		UserName:    info.UserName,
		IsLogin:     info.IsLogin,
		SessionId:   info.SessionID,
		Method:      info.Method,
		Path:        info.Path,
		RequestTime: info.RequestTime.Format(time.DateTime),
		IpAddress:   info.IPAddress,
		IpLocation:  info.IPLocation,
		Latency:     info.Latency,
		Os:          info.Os,
		Browser:     info.Browser,
		UserAgent:   info.UserAgent,
		Header:      info.Header,
		GetParams:   info.GetParams,
		PostData:    info.PostData,
		ResCode:     info.ResCode,
		Reason:      info.Reason,
		ResStatus:   info.ResStatus,
		Stack:       info.Stack,
		CreateTime:  info.CreateTime.Format(time.DateTime),
	}, nil
}

type deptNode struct {
	Id       int64
	Pid      int64
	Value    *ent.Dept
	Children []*deptNode
}

func deptToNode(value *ent.Dept) *deptNode {
	return &deptNode{Id: value.ID, Pid: value.Pid, Value: value}
}

func (uc *AdminUsecase) buildDeptTree(forest *[]*deptNode, dept *ent.Dept, top bool) bool {
	if dept.Pid == 0 {
		*forest = append(*forest, deptToNode(dept))
		return true
	}
	for _, item := range *forest {
		if item.Id == dept.Pid {
			item.Children = append(item.Children, deptToNode(dept))
			return true
		}
		if len(item.Children) > 0 && uc.buildDeptTree(&item.Children, dept, false) {
			return true
		}
	}
	if top {
		*forest = append(*forest, deptToNode(dept))
		return true
	}
	return false
}

func toDeptTree(forest []*deptNode, strPid string) []*v1.DeptListItem {
	items := make([]*v1.DeptListItem, 0, len(forest))
	for _, item := range forest {
		id := strconv.FormatInt(item.Id, 10)
		status := int32(0)
		if item.Value.Status {
			status = 1
		}
		items = append(items, &v1.DeptListItem{
			Id:         id,
			Pid:        strPid,
			Name:       item.Value.Name,
			OrderNo:    item.Value.Sort,
			Remark:     item.Value.Desc,
			Status:     status,
			CreateTime: item.Value.CreateTime.Format(time.DateTime),
			Dom:        item.Value.Dom,
			Children:   toDeptTree(item.Children, id),
		})
	}
	return items
}

func deptToReply(dept *ent.Dept, pid string) *v1.DeptListItem {
	status := int32(0)
	if dept.Status {
		status = 1
	}
	return &v1.DeptListItem{
		Id:         strconv.FormatInt(dept.ID, 10),
		Pid:        pid,
		Name:       dept.Name,
		OrderNo:    dept.Sort,
		Remark:     dept.Desc,
		Status:     status,
		CreateTime: dept.CreateTime.Format(time.DateTime),
		Dom:        dept.Dom,
	}
}

func (uc *AdminUsecase) createMenuTree(menuList []*ent.Menu) []*v1.SysMenuListItem {
	items := make([]*v1.SysMenuListItem, 0)
	for _, menu := range menuList {
		if menu.Pid == 0 {
			items = append(items, entMenuToMenu(menu))
			continue
		}
		if !buildMenuTree(&items, menu) {
			items = append(items, entMenuToMenu(menu))
		}
	}
	return items
}

func buildMenuTree(menuList *[]*v1.SysMenuListItem, menu *ent.Menu) bool {
	for _, item := range *menuList {
		if len(item.Children) > 0 && buildMenuTree(&item.Children, menu) {
			return true
		}
		if int64(item.Id) == menu.Pid {
			item.Children = append(item.Children, entMenuToMenu(menu))
			return true
		}
	}
	return false
}

func filterVisibleMenus(menuList []*ent.Menu, authority *authv1.GetCurrentUserMenuAuthorityReply) []*ent.Menu {
	if authority.GetIsRoot() {
		items := make([]*ent.Menu, 0, len(menuList))
		for _, item := range menuList {
			if item.Status {
				items = append(items, item)
			}
		}
		sortMenus(items)
		return items
	}
	allowed := make(map[int64]struct{}, len(authority.MenuIds))
	for _, id := range authority.MenuIds {
		allowed[int64(id)] = struct{}{}
	}
	menuByID := make(map[int64]*ent.Menu, len(menuList))
	for _, item := range menuList {
		menuByID[item.ID] = item
	}
	for _, id := range authority.MenuIds {
		item, ok := menuByID[int64(id)]
		if !ok {
			continue
		}
		pid := item.Pid
		for pid > 0 {
			parent, ok := menuByID[pid]
			if !ok {
				break
			}
			allowed[parent.ID] = struct{}{}
			pid = parent.Pid
		}
	}
	items := make([]*ent.Menu, 0, len(menuList))
	for _, item := range menuList {
		if !item.Status {
			continue
		}
		if _, ok := allowed[item.ID]; !ok {
			continue
		}
		items = append(items, item)
	}
	sortMenus(items)
	return items
}

func sortMenus(items []*ent.Menu) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Pid != items[j].Pid {
			return items[i].Pid < items[j].Pid
		}
		if items[i].Order != items[j].Order {
			return items[i].Order < items[j].Order
		}
		return items[i].ID < items[j].ID
	})
}

func createCurrentUserMenuTree(menuList []*ent.Menu) []*v1.CurrentUserMenuItem {
	items := make([]*v1.CurrentUserMenuItem, 0)
	for _, menu := range menuList {
		if menu.Pid == 0 {
			items = append(items, entMenuToCurrentUserMenu(menu))
			continue
		}
		if !buildCurrentUserMenuTree(&items, menu) {
			items = append(items, entMenuToCurrentUserMenu(menu))
		}
	}
	return items
}

func buildCurrentUserMenuTree(menuList *[]*v1.CurrentUserMenuItem, menu *ent.Menu) bool {
	for _, item := range *menuList {
		if len(item.Children) > 0 && buildCurrentUserMenuTree(&item.Children, menu) {
			return true
		}
		if int64(item.Id) == menu.Pid {
			item.Children = append(item.Children, entMenuToCurrentUserMenu(menu))
			return true
		}
	}
	return false
}

func entMenuToMenu(menu *ent.Menu) *v1.SysMenuListItem {
	status := int32(0)
	if menu.Status {
		status = 1
	}
	redirect := menu.Redirect
	activePath := menu.ActivePath
	link := menu.Link
	iframeSrc := menu.IframeSrc
	keepAlive := menu.Keepalive
	maxNumOfOpenTab := int64(menu.MaxNumOfOpenTab)
	hideInMenu := menu.HideInMenu
	hideInTab := menu.HideInTab
	hideInBreadcrumb := menu.HideInBreadcrumb
	hideChildrenInMenu := menu.HideChildrenInMenu
	return &v1.SysMenuListItem{
		Id:         int32(menu.ID),
		Component:  menu.Component,
		Status:     &status,
		AuthCode:   "",
		Name:       menu.Name,
		Path:       menu.Path,
		Pid:        menu.Pid,
		Redirect:   &redirect,
		Type:       menu.Type,
		CreateTime: menu.CreateTime.Format(time.DateTime),
		Meta: &v1.Meta{
			Order:              int64(menu.Order),
			Icon:               menu.Icon,
			Title:              menu.Title,
			ActivePath:         &activePath,
			IframeSrc:          &iframeSrc,
			Link:               &link,
			KeepAlive:          &keepAlive,
			MaxNumOfOpenTab:    &maxNumOfOpenTab,
			IgnoreAccess:       menu.IgnoreAccess,
			HideInMenu:         &hideInMenu,
			HideInTab:          &hideInTab,
			HideInBreadcrumb:   &hideInBreadcrumb,
			HideChildrenInMenu: &hideChildrenInMenu,
			Authority:          splitAuthority(menu.Authority),
		},
	}
}

func entMenuToCurrentUserMenu(menu *ent.Menu) *v1.CurrentUserMenuItem {
	item := entMenuToMenu(menu)
	return &v1.CurrentUserMenuItem{
		Id:         item.Id,
		Component:  item.Component,
		Status:     item.Status,
		AuthCode:   item.AuthCode,
		Name:       item.Name,
		Path:       item.Path,
		Pid:        item.Pid,
		Redirect:   item.Redirect,
		Type:       item.Type,
		Meta:       item.Meta,
		CreateTime: item.CreateTime,
	}
}

func entMenuToRecord(menu *ent.Menu) *v1.MenuRecord {
	item := entMenuToMenu(menu)
	return &v1.MenuRecord{
		Id:         item.Id,
		Component:  item.Component,
		Status:     item.Status,
		AuthCode:   item.AuthCode,
		Name:       item.Name,
		Path:       item.Path,
		Pid:        item.Pid,
		Redirect:   item.Redirect,
		Type:       item.Type,
		Meta:       item.Meta,
		CreateTime: item.CreateTime,
	}
}

func menuToEntMenu(menu *v1.SysMenuListItem) *ent.Menu {
	meta := menu.Meta
	if meta == nil {
		meta = &v1.Meta{}
	}
	status := false
	if menu.Status != nil && *menu.Status == 1 {
		status = true
	}
	return &ent.Menu{
		ID:                 int64(menu.Id),
		Pid:                menu.Pid,
		Type:               menu.Type,
		Status:             status,
		Path:               menu.Path,
		Redirect:           ptrToString(menu.Redirect),
		Name:               menu.Name,
		Component:          menu.Component,
		Icon:               meta.Icon,
		Title:              meta.Title,
		Order:              int32(meta.Order),
		Link:               ptrToString(meta.Link),
		IframeSrc:          ptrToString(meta.IframeSrc),
		ActivePath:         ptrToString(meta.ActivePath),
		MaxNumOfOpenTab:    int16(ptrToInt64(meta.MaxNumOfOpenTab)),
		Keepalive:          ptrToBool(meta.KeepAlive),
		IgnoreAccess:       meta.IgnoreAccess,
		Authority:          strings.Join(meta.Authority, ","),
		HideInMenu:         ptrToBool(meta.HideInMenu),
		HideInTab:          ptrToBool(meta.HideInTab),
		HideInBreadcrumb:   ptrToBool(meta.HideInBreadcrumb),
		HideChildrenInMenu: ptrToBool(meta.HideChildrenInMenu),
	}
}

func ptrToString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func ptrToBool(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

func ptrToInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func splitAuthority(v string) []string {
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	res := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			res = append(res, part)
		}
	}
	return res
}

func replaceBracesIfExists(str string) (bool, string) {
	hasBraces := strings.Contains(str, "{") || strings.Contains(str, "}")
	if !hasBraces {
		return false, str
	}
	re := regexp.MustCompile(`{[^}]*}`)
	return true, re.ReplaceAllString(str, "%")
}

var _ = emptypb.Empty{}
var _ = authx.UserID
