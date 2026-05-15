package data

import (
	"context"

	v1 "base-server/api/gen/go/admin/service/v1"
	"base-server/pkg/authx"
)

type adminProjectionSource struct {
	repo *adminRepo
}

func newAdminProjectionSource(repo *adminRepo) authx.ProjectionSource {
	return &adminProjectionSource{repo: repo}
}

func (s *adminProjectionSource) SourceService() string {
	return "admin"
}

func (s *adminProjectionSource) ListProjectionRoles(ctx context.Context) ([]authx.ProjectionRole, error) {
	roleList, err := s.repo.ListAllRoles(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]authx.ProjectionRole, 0, len(roleList))
	for _, roleItem := range roleList {
		items = append(items, roleToPolicyRole(roleItem))
	}
	return items, nil
}

func (s *adminProjectionSource) ListProjectionAPIs(ctx context.Context) ([]authx.ProjectionAPI, error) {
	apiList, _, err := s.repo.GetApiList(ctx, &v1.GetApiPageParams{})
	if err != nil {
		return nil, err
	}
	items := make([]authx.ProjectionAPI, 0, len(apiList))
	for _, apiItem := range apiList {
		items = append(items, *apiToPolicyAPI(apiItem))
	}
	return items, nil
}

func (s *adminProjectionSource) ListProjectionBindings(ctx context.Context) ([]authx.UserRoleBindingProjection, error) {
	bindingList, err := s.repo.listExplicitUserRoleBindings(ctx)
	if err != nil {
		return nil, err
	}
	roleValues, err := s.repo.resolveBindingRoleValues(ctx, bindingList)
	if err != nil {
		return nil, err
	}
	items := make([]authx.UserRoleBindingProjection, 0, len(bindingList))
	for _, binding := range bindingList {
		items = append(items, userRoleBindingToPolicyBinding(binding, roleValues[binding.RoleID]))
	}
	return items, nil
}
