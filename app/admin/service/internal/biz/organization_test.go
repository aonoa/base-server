package biz

import (
	"context"
	"io"
	"reflect"
	"sort"
	"testing"

	v1 "base-server/api/gen/go/admin/service/v1"
	"base-server/pkg/authx"
	"base-server/pkg/data/ent"

	"github.com/go-kratos/kratos/v2/log"
	kratosjwt "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

const rootActorID = "root-actor"

func TestValidateDeptParentOrganization(t *testing.T) {
	tests := []struct {
		name           string
		parent         *ent.Dept
		organizationID string
		wantErr        bool
	}{
		{
			name:           "same organization",
			parent:         &ent.Dept{OrganizationID: "org-a"},
			organizationID: "org-a",
		},
		{
			name:           "root department",
			organizationID: "org-a",
		},
		{
			name:           "different organization",
			parent:         &ent.Dept{OrganizationID: "org-b"},
			organizationID: "org-a",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDeptParentOrganization(tt.parent, tt.organizationID)
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateRolePermissionScopeRejectsOutsideConfiguredScope(t *testing.T) {
	uc := NewAdminUsecase(&organizationPermissionScopeRepo{
		scopes: []*ent.OrganizationPermissionScope{
			{OrganizationID: "org-a", PermissionType: "menu", PermissionRef: "10"},
			{OrganizationID: "org-a", PermissionType: "resource", PermissionRef: "resource-a"},
		},
	}, log.NewStdLogger(io.Discard))

	if err := uc.validateRolePermissionScope(
		context.Background(),
		"org-a",
		[]int32{10},
		[]string{"resource-a"},
	); err != nil {
		t.Fatalf("validateRolePermissionScope() allowed scope error = %v", err)
	}
	if err := uc.validateRolePermissionScope(
		context.Background(),
		"org-a",
		[]int32{11},
		[]string{"resource-a"},
	); err == nil {
		t.Fatal("expected outside menu to be rejected")
	}
	if err := uc.validateRolePermissionScope(
		context.Background(),
		"org-a",
		[]int32{10},
		[]string{"resource-b"},
	); err == nil {
		t.Fatal("expected outside resource to be rejected")
	}
}

func TestValidateRolePermissionScopeRejectsEmptyScope(t *testing.T) {
	uc := NewAdminUsecase(&organizationPermissionScopeRepo{}, log.NewStdLogger(io.Discard))

	if err := uc.validateRolePermissionScope(
		context.Background(),
		"org-a",
		[]int32{999},
		[]string{"resource-any"},
	); err == nil {
		t.Fatal("expected empty organization scope to reject requested permissions")
	}
}

func TestValidateOrganizationPermissionScopeShrinkRejectsUsedPermission(t *testing.T) {
	uc := NewAdminUsecase(&organizationPermissionScopeRepo{
		roles: []*ent.Role{
			{
				Name:  "管理员",
				Menus: []int32{10, 11},
				Edges: ent.RoleEdges{
					Resource: []*ent.Resource{{ID: "resource-a"}, {ID: "resource-b"}},
				},
			},
		},
	}, log.NewStdLogger(io.Discard))

	if err := uc.validateOrganizationPermissionScopeShrink(
		context.Background(),
		"org-a",
		[]int32{10},
		[]string{"resource-a", "resource-b"},
	); err == nil {
		t.Fatal("expected used menu shrink to be rejected")
	}
	if err := uc.validateOrganizationPermissionScopeShrink(
		context.Background(),
		"org-a",
		[]int32{10, 11},
		[]string{"resource-a"},
	); err == nil {
		t.Fatal("expected used resource shrink to be rejected")
	}
}

func TestEnsureUserOrganizationMemberRejectsNonMember(t *testing.T) {
	uc := NewAdminUsecase(&organizationPermissionScopeRepo{
		userBelongs: map[string]bool{"user-a|org-a": true},
	}, log.NewStdLogger(io.Discard))

	if err := uc.ensureUserOrganizationMember(context.Background(), "user-a", "org-a"); err != nil {
		t.Fatalf("ensureUserOrganizationMember() member error = %v", err)
	}
	if err := uc.ensureUserOrganizationMember(context.Background(), "user-b", "org-a"); err == nil {
		t.Fatal("expected non-member to be rejected")
	}
}

func TestSaveOrganizationPermissionScopeRequiresPlatformRoot(t *testing.T) {
	repo := &organizationPermissionScopeRepo{
		roles:                 []*ent.Role{},
		currentOrganizationID: "default-org",
		defaultOrganizationID: "default-org",
		roleBindings: []*ent.UserRoleBinding{
			{UserID: rootActorID, RoleID: 1, OrganizationID: "default-org"},
		},
		roleValues: map[int64]string{1: "root"},
	}
	uc := NewAdminUsecase(repo, log.NewStdLogger(io.Discard))

	if _, err := uc.SaveOrganizationPermissionScope(
		testUserContext("actor-a"),
		&v1.SaveOrganizationPermissionScopeRequest{OrganizationId: "org-a"},
	); err == nil {
		t.Fatal("expected non-root actor to be rejected")
	}

	if _, err := uc.SaveOrganizationPermissionScope(
		testUserContext(rootActorID),
		&v1.SaveOrganizationPermissionScopeRequest{OrganizationId: "org-b"},
	); err != nil {
		t.Fatalf("platform root should save organization scope, got %v", err)
	}
	if repo.savedOrganizationID != "org-b" {
		t.Fatalf("saved organization = %s, want org-b", repo.savedOrganizationID)
	}
}

func TestSaveOrganizationPermissionScopeRejectsRootOutsideDefaultOrganization(t *testing.T) {
	repo := &organizationPermissionScopeRepo{
		roles:                 []*ent.Role{},
		currentOrganizationID: "org-a",
		defaultOrganizationID: "default-org",
		roleBindings: []*ent.UserRoleBinding{
			{UserID: rootActorID, RoleID: 1, OrganizationID: "default-org"},
		},
		roleValues: map[int64]string{1: "root"},
	}
	uc := NewAdminUsecase(repo, log.NewStdLogger(io.Discard))

	if _, err := uc.SaveOrganizationPermissionScope(
		testUserContext(rootActorID),
		&v1.SaveOrganizationPermissionScopeRequest{OrganizationId: "org-b"},
	); err == nil {
		t.Fatal("expected root outside default organization to be rejected")
	}
}

func TestOrganizationMutationRequiresPlatformRoot(t *testing.T) {
	repo := &organizationPermissionScopeRepo{
		currentOrganizationID: "default-org",
		defaultOrganizationID: "default-org",
		roleBindings: []*ent.UserRoleBinding{
			{UserID: rootActorID, RoleID: 1, OrganizationID: "default-org"},
		},
		roleValues: map[int64]string{1: "root"},
	}
	uc := NewAdminUsecase(repo, log.NewStdLogger(io.Discard))

	if _, err := uc.AddOrganization(
		testUserContext("actor-a"),
		&v1.OrganizationItem{Id: "org-a", Name: "A"},
	); err == nil {
		t.Fatal("expected non-root organization create to be rejected")
	}
	if _, err := uc.UpdateOrganization(
		testUserContext("actor-a"),
		&v1.OrganizationItem{Id: "org-a", Name: "A"},
	); err == nil {
		t.Fatal("expected non-root organization update to be rejected")
	}
	if err := uc.DelOrganization(testUserContext("actor-a"), "org-a"); err == nil {
		t.Fatal("expected non-root organization delete to be rejected")
	}
	if _, err := uc.AddOrganization(
		testUserContext(rootActorID),
		&v1.OrganizationItem{Id: "org-a", Name: "A"},
	); err != nil {
		t.Fatalf("platform root organization create error = %v", err)
	}
	if repo.addOrganizationCalled != 1 {
		t.Fatalf("add organization calls = %d, want 1", repo.addOrganizationCalled)
	}
}

func TestGetOrganizationListScopesNonRootToCurrent(t *testing.T) {
	uc := NewAdminUsecase(&organizationPermissionScopeRepo{
		currentOrganizationID: "org-a",
		defaultOrganizationID: "default-org",
		organizations: map[string]*ent.Organization{
			"default-org": {ID: "default-org", Name: "默认组织", Code: "default", Sort: 0, Status: true},
			"org-a":       {ID: "org-a", Name: "A 组织", Code: "a", Sort: 1, Status: true},
			"org-b":       {ID: "org-b", Name: "B 组织", Code: "b", Sort: 2, Status: true},
		},
	}, log.NewStdLogger(io.Discard))

	reply, err := uc.GetOrganizationList(testUserContext("actor-a"), &v1.GetOrganizationListParams{PageSize: 200})
	if err != nil {
		t.Fatalf("GetOrganizationList() error = %v", err)
	}
	got := make([]string, 0, len(reply.Items))
	for _, item := range reply.Items {
		got = append(got, item.Id)
	}
	want := []string{"org-a"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("organization ids = %#v, want %#v", got, want)
	}
	if reply.CanManageOrganizations {
		t.Fatal("non-root organization list should not expose organization management capability")
	}
}

func TestGetOrganizationListAllowsRootRoleToManageOrganizations(t *testing.T) {
	uc := NewAdminUsecase(&organizationPermissionScopeRepo{
		currentOrganizationID: "default-org",
		defaultOrganizationID: "default-org",
		roleBindings: []*ent.UserRoleBinding{
			{UserID: rootActorID, RoleID: 1, OrganizationID: "default-org"},
		},
		roleValues: map[int64]string{1: "root"},
		organizations: map[string]*ent.Organization{
			"default-org": {ID: "default-org", Name: "默认组织", Code: "default", Sort: 0, Status: true},
			"org-a":       {ID: "org-a", Name: "A 组织", Code: "a", Sort: 1, Status: true},
			"org-b":       {ID: "org-b", Name: "B 组织", Code: "b", Sort: 2, Status: true},
		},
	}, log.NewStdLogger(io.Discard))

	reply, err := uc.GetOrganizationList(testUserContext(rootActorID), &v1.GetOrganizationListParams{PageSize: 200})
	if err != nil {
		t.Fatalf("GetOrganizationList() error = %v", err)
	}
	got := make([]string, 0, len(reply.Items))
	for _, item := range reply.Items {
		got = append(got, item.Id)
	}
	want := []string{"default-org", "org-a", "org-b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("organization ids = %#v, want %#v", got, want)
	}
	if !reply.CanManageOrganizations {
		t.Fatal("root role organization list should expose organization management capability")
	}
}

func TestGetOrganizationListScopesRootOutsideDefaultOrganizationToCurrent(t *testing.T) {
	uc := NewAdminUsecase(&organizationPermissionScopeRepo{
		currentOrganizationID: "org-a",
		defaultOrganizationID: "default-org",
		roleBindings: []*ent.UserRoleBinding{
			{UserID: rootActorID, RoleID: 1, OrganizationID: "default-org"},
		},
		roleValues: map[int64]string{1: "root"},
		organizations: map[string]*ent.Organization{
			"default-org": {ID: "default-org", Name: "默认组织", Code: "default", Sort: 0, Status: true},
			"org-a":       {ID: "org-a", Name: "A 组织", Code: "a", Sort: 1, Status: true},
			"org-b":       {ID: "org-b", Name: "B 组织", Code: "b", Sort: 2, Status: true},
		},
	}, log.NewStdLogger(io.Discard))

	reply, err := uc.GetOrganizationList(testUserContext(rootActorID), &v1.GetOrganizationListParams{PageSize: 200})
	if err != nil {
		t.Fatalf("GetOrganizationList() error = %v", err)
	}
	got := make([]string, 0, len(reply.Items))
	for _, item := range reply.Items {
		got = append(got, item.Id)
	}
	want := []string{"org-a"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("organization ids = %#v, want %#v", got, want)
	}
	if reply.CanManageOrganizations {
		t.Fatal("root outside default organization should not expose organization management capability")
	}
}

func TestResolveOrganizationIDRejectsCrossOrganizationForNonRoot(t *testing.T) {
	nonRootUC := NewAdminUsecase(&organizationPermissionScopeRepo{
		currentOrganizationID: "org-a",
		defaultOrganizationID: "default-org",
	}, log.NewStdLogger(io.Discard))

	if _, err := nonRootUC.resolveOrganizationID(testUserContext("actor-a"), "org-b"); err == nil {
		t.Fatal("expected cross-organization id to be rejected for non-root actor")
	}
	if got, err := nonRootUC.resolveOrganizationID(testUserContext("actor-a"), "org-a"); err != nil || got != "org-a" {
		t.Fatalf("current organization resolve = %s, err = %v", got, err)
	}

	rootUC := NewAdminUsecase(&organizationPermissionScopeRepo{
		currentOrganizationID: "default-org",
		defaultOrganizationID: "default-org",
		roleBindings: []*ent.UserRoleBinding{
			{UserID: rootActorID, RoleID: 1, OrganizationID: "default-org"},
		},
		roleValues: map[int64]string{1: "root"},
	}, log.NewStdLogger(io.Discard))
	if got, err := rootUC.resolveOrganizationID(testUserContext(rootActorID), "org-b"); err != nil || got != "org-b" {
		t.Fatalf("platform root cross-organization resolve = %s, err = %v", got, err)
	}
}

func TestGetOrganizationMembersAllowsDefaultAndCurrentForNonRoot(t *testing.T) {
	uc := NewAdminUsecase(&organizationPermissionScopeRepo{
		currentOrganizationID: "org-a",
		defaultOrganizationID: "default-org",
	}, log.NewStdLogger(io.Discard))
	ctx := testUserContext("actor-a")

	if _, err := uc.GetOrganizationMembers(ctx, &v1.GetOrganizationMembersRequest{OrganizationId: "default-org"}); err != nil {
		t.Fatalf("default organization members should be readable, got %v", err)
	}
	if _, err := uc.GetOrganizationMembers(ctx, &v1.GetOrganizationMembersRequest{OrganizationId: "org-a"}); err != nil {
		t.Fatalf("current organization members should be readable, got %v", err)
	}
	if _, err := uc.GetOrganizationMembers(ctx, &v1.GetOrganizationMembersRequest{OrganizationId: "org-b"}); err == nil {
		t.Fatal("expected other organization members to be rejected")
	}
}

func TestSaveOrganizationMembersRejectsCrossOrganizationForNonRoot(t *testing.T) {
	uc := NewAdminUsecase(&organizationPermissionScopeRepo{
		currentOrganizationID: "org-a",
		defaultOrganizationID: "default-org",
	}, log.NewStdLogger(io.Discard))

	if _, err := uc.SaveOrganizationMembers(
		testUserContext("actor-a"),
		&v1.SaveOrganizationMembersRequest{OrganizationId: "org-b"},
	); err == nil {
		t.Fatal("expected non-root actor to be rejected when saving another organization")
	}
}

func TestGetRoleListRejectsCrossOrganizationForNonRoot(t *testing.T) {
	uc := NewAdminUsecase(&organizationPermissionScopeRepo{
		currentOrganizationID: "org-a",
	}, log.NewStdLogger(io.Discard))

	if _, err := uc.GetRoleList(
		testUserContext("actor-a"),
		&v1.RolePageParams{OrganizationId: "org-b"},
	); err == nil {
		t.Fatal("expected cross-organization role list to be rejected")
	}
}

func TestNormalizePermissionScopeIDs(t *testing.T) {
	if got, want := normalizeMenuIDs([]int32{3, 1, 3, 0, -1, 2}), []int32{1, 2, 3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeMenuIDs() = %#v, want %#v", got, want)
	}
	if got, want := normalizeStringIDs([]string{" b ", "", "a", "b"}), []string{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeStringIDs() = %#v, want %#v", got, want)
	}
}

func TestValidateRoleGrantableRejectsOutsideActorGrant(t *testing.T) {
	uc := NewAdminUsecase(&organizationPermissionScopeRepo{
		userRoles: []*ent.Role{
			{
				DataScope: "all",
				Menus:     []int32{10},
				Edges: ent.RoleEdges{
					Resource: []*ent.Resource{{ID: "resource-a", Type: "api"}},
				},
			},
		},
	}, log.NewStdLogger(io.Discard))
	ctx := testUserContext("actor-a")

	if err := uc.validateRoleGrantable(
		ctx,
		"org-a",
		[]int32{10},
		[]string{"resource-a"},
	); err != nil {
		t.Fatalf("validateRoleGrantable() allowed actor grant error = %v", err)
	}
	if err := uc.validateRoleGrantable(
		ctx,
		"org-a",
		[]int32{11},
		[]string{"resource-a"},
	); err == nil {
		t.Fatal("expected menu outside actor grant to be rejected")
	}
	if err := uc.validateRoleGrantable(
		ctx,
		"org-a",
		[]int32{10},
		[]string{"resource-b"},
	); err == nil {
		t.Fatal("expected resource outside actor grant to be rejected")
	}
}

func TestValidateRoleGrantableAllowsRootRole(t *testing.T) {
	uc := NewAdminUsecase(&organizationPermissionScopeRepo{
		currentOrganizationID: "default-org",
		defaultOrganizationID: "default-org",
		roleBindings: []*ent.UserRoleBinding{
			{UserID: rootActorID, RoleID: 1, OrganizationID: "default-org"},
		},
		roleValues: map[int64]string{1: "root"},
	}, log.NewStdLogger(io.Discard))

	if err := uc.validateRoleGrantable(
		testUserContext(rootActorID),
		"org-a",
		[]int32{999},
		[]string{"resource-any"},
	); err != nil {
		t.Fatalf("root role should bypass grantable scope, got %v", err)
	}
}

func TestValidateUserRoleBindingGrantableRejectsRoleOutsideActorGrant(t *testing.T) {
	uc := NewAdminUsecase(&organizationPermissionScopeRepo{
		rolesByID: map[int64]*ent.Role{
			1: {
				ID:    1,
				Menus: []int32{10, 11},
				Edges: ent.RoleEdges{
					Resource: []*ent.Resource{{ID: "resource-a", Type: "api"}},
				},
			},
		},
		userRoles: []*ent.Role{
			{
				DataScope: "all",
				Menus:     []int32{10},
				Edges: ent.RoleEdges{
					Resource: []*ent.Resource{{ID: "resource-a", Type: "api"}},
				},
			},
		},
	}, log.NewStdLogger(io.Discard))

	if err := uc.validateUserRoleBindingGrantable(
		testUserContext("actor-a"),
		"org-a",
		[]int64{1},
	); err == nil {
		t.Fatal("expected role with menu outside actor grant to be rejected")
	}
}

func TestValidateRoleDataScopeGrantableRejectsEscalation(t *testing.T) {
	uc := NewAdminUsecase(&organizationPermissionScopeRepo{
		userRoles: []*ent.Role{
			{DataScope: "self_dept"},
		},
	}, log.NewStdLogger(io.Discard))
	ctx := testUserContext("actor-a")

	if err := uc.validateRoleDataScopeGrantable(ctx, "org-a", "member", "self", nil); err != nil {
		t.Fatalf("self should be assignable by self_dept actor, got %v", err)
	}
	if err := uc.validateRoleDataScopeGrantable(ctx, "org-a", "member", "all", nil); err == nil {
		t.Fatal("expected all data scope to be rejected")
	}
	if err := uc.validateRoleDataScopeGrantable(ctx, "org-a", "member", "self_dept_and_child", nil); err == nil {
		t.Fatal("expected child department data scope to be rejected")
	}
}

func TestValidateRoleDataScopeGrantableAllowsCustomDeptSubset(t *testing.T) {
	uc := NewAdminUsecase(&organizationPermissionScopeRepo{
		userRoles: []*ent.Role{
			{DataScope: "custom_depts", DataScopeDeptIds: []int64{1, 2}},
		},
	}, log.NewStdLogger(io.Discard))
	ctx := testUserContext("actor-a")

	if err := uc.validateRoleDataScopeGrantable(ctx, "org-a", "member", "custom_depts", []int64{2}); err != nil {
		t.Fatalf("custom department subset should be assignable, got %v", err)
	}
	if err := uc.validateRoleDataScopeGrantable(ctx, "org-a", "member", "custom_depts", []int64{3}); err == nil {
		t.Fatal("expected custom department outside actor grant to be rejected")
	}
}

type organizationPermissionScopeRepo struct {
	AdminRepo
	scopes                []*ent.OrganizationPermissionScope
	roles                 []*ent.Role
	rolesByID             map[int64]*ent.Role
	roleBindings          []*ent.UserRoleBinding
	roleValues            map[int64]string
	userRoles             []*ent.Role
	userBelongs           map[string]bool
	currentOrganizationID string
	defaultOrganizationID string
	savedOrganizationID   string
	addOrganizationCalled int
	organizations         map[string]*ent.Organization
}

func (r *organizationPermissionScopeRepo) ListOrganizationPermissionScopes(context.Context, string) ([]*ent.OrganizationPermissionScope, error) {
	return r.scopes, nil
}

func (r *organizationPermissionScopeRepo) SaveOrganizationPermissionScopes(_ context.Context, organizationID string, _ []int32, _ []string, _ string) ([]*ent.OrganizationPermissionScope, error) {
	r.savedOrganizationID = organizationID
	return r.scopes, nil
}

func (r *organizationPermissionScopeRepo) GetOrganization(_ context.Context, organizationID string) (*ent.Organization, error) {
	if r.organizations != nil {
		if item := r.organizations[organizationID]; item != nil {
			return item, nil
		}
	}
	return &ent.Organization{ID: organizationID}, nil
}

func (r *organizationPermissionScopeRepo) ListOrganizations(context.Context, *v1.GetOrganizationListParams) ([]*ent.Organization, int64, error) {
	items := make([]*ent.Organization, 0, len(r.organizations))
	for _, item := range r.organizations {
		items = append(items, item)
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Sort != items[j].Sort {
			return items[i].Sort < items[j].Sort
		}
		return items[i].ID < items[j].ID
	})
	return items, int64(len(items)), nil
}

func (r *organizationPermissionScopeRepo) AddOrganization(_ context.Context, req *v1.OrganizationItem) (*ent.Organization, error) {
	r.addOrganizationCalled++
	return &ent.Organization{ID: req.GetId(), Name: req.GetName(), Status: true}, nil
}

func (r *organizationPermissionScopeRepo) CountOrganizationMembers(context.Context, string) (int64, error) {
	return 0, nil
}

func (r *organizationPermissionScopeRepo) CountOrganizationDepts(context.Context, string) (int64, error) {
	return 0, nil
}

func (r *organizationPermissionScopeRepo) ListRoles(context.Context, *v1.RolePageParams) ([]*ent.Role, error) {
	return r.roles, nil
}

func (r *organizationPermissionScopeRepo) GetDefaultOrganizationID(context.Context) (string, error) {
	if r.defaultOrganizationID == "" {
		return "default-org", nil
	}
	return r.defaultOrganizationID, nil
}

func (r *organizationPermissionScopeRepo) ListOrganizationMembers(context.Context, string) ([]*v1.OrganizationMemberItem, error) {
	return []*v1.OrganizationMemberItem{}, nil
}

func (r *organizationPermissionScopeRepo) SaveOrganizationMembers(context.Context, string, []string) ([]*v1.OrganizationMemberItem, error) {
	return []*v1.OrganizationMemberItem{}, nil
}

func (r *organizationPermissionScopeRepo) EnsureAllUsersInDefaultOrganization(context.Context) error {
	return nil
}

func (r *organizationPermissionScopeRepo) ListUserOrganizationRoles(context.Context, string, string) ([]*ent.Role, error) {
	return r.userRoles, nil
}

func (r *organizationPermissionScopeRepo) GetCurrentOrganizationID(context.Context, string) (string, error) {
	if r.currentOrganizationID == "" {
		return "org-a", nil
	}
	return r.currentOrganizationID, nil
}

func (r *organizationPermissionScopeRepo) GetRole(_ context.Context, roleID int64) (*ent.Role, error) {
	if r.rolesByID != nil {
		if item := r.rolesByID[roleID]; item != nil {
			return item, nil
		}
	}
	return &ent.Role{ID: roleID}, nil
}

func (r *organizationPermissionScopeRepo) ListUserRoleBindings(context.Context) ([]*ent.UserRoleBinding, error) {
	return r.roleBindings, nil
}

func (r *organizationPermissionScopeRepo) ResolveRoleValues(_ context.Context, roleIDs []int64) (map[int64]string, error) {
	values := make(map[int64]string, len(roleIDs))
	for _, roleID := range roleIDs {
		if value := r.roleValues[roleID]; value != "" {
			values[roleID] = value
			continue
		}
		if item := r.rolesByID[roleID]; item != nil {
			values[roleID] = item.Value
		}
	}
	return values, nil
}

func (r *organizationPermissionScopeRepo) UserBelongsToOrganization(_ context.Context, userID string, organizationID string) (bool, error) {
	return r.userBelongs[userID+"|"+organizationID], nil
}

func testUserContext(userID string) context.Context {
	return kratosjwt.NewContext(
		context.Background(),
		jwtv5.MapClaims{authx.ClaimUserID: userID},
	)
}
