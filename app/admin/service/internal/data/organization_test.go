package data

import (
	"context"
	"reflect"
	"testing"

	userv1 "base-server/api/gen/go/user/service/v1"
	"base-server/pkg/authx"
	"base-server/pkg/data/ent"
	"base-server/pkg/data/ent/organizationpermissionscope"
	"base-server/pkg/data/ent/userdeptmembership"
	"base-server/pkg/data/ent/userorganization"

	"github.com/go-kratos/kratos/v2/transport"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestNormalizeUserIDs(t *testing.T) {
	got := normalizeUserIDs([]string{" user-a ", "", "user-b", "user-a", "  ", "user-c"})
	want := []string{"user-a", "user-b", "user-c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeUserIDs() = %#v, want %#v", got, want)
	}
}

func TestGetUserListForwardsAuthorization(t *testing.T) {
	const authorization = "Bearer test-token"
	client := &captureUserServiceClient{}
	repo := &adminRepo{data: &Data{userClient: client}}
	ctx := transport.NewServerContext(context.Background(), testTransport{
		headers: testHeader{authx.HeaderAuthorization: authorization},
	})

	_, err := repo.getUserList(ctx, &userv1.GetUserParams{CurrentPage: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("getUserList() error = %v", err)
	}
	values := client.outgoingMetadata.Get("authorization")
	if len(values) != 1 || values[0] != authorization {
		t.Fatalf("forwarded authorization = %#v, want %#v", values, []string{authorization})
	}
}

func TestSwitchCurrentOrganizationUpdatesSingleCurrentMembership(t *testing.T) {
	ctx := context.Background()
	repo, client := newOrganizationTestRepo(t)
	defer client.Close()

	const userID = "user-1"
	const secondOrganizationID = "5d76c984-7e30-4b87-8433-ff271f902489"
	createOrganization(t, client, defaultOrganizationID, "默认组织", "default")
	createOrganization(t, client, secondOrganizationID, "第二组织", "second")

	if err := repo.EnsureUserInDefaultOrganization(ctx, userID); err != nil {
		t.Fatalf("EnsureUserInDefaultOrganization() error = %v", err)
	}
	if _, err := client.UserOrganization.Create().
		SetUserID(userID).
		SetOrganizationID(secondOrganizationID).
		SetStatus(true).
		SetIsPrimary(false).
		Save(ctx); err != nil {
		t.Fatalf("create second membership: %v", err)
	}

	if err := repo.SwitchCurrentOrganization(ctx, userID, secondOrganizationID); err != nil {
		t.Fatalf("SwitchCurrentOrganization() error = %v", err)
	}
	currentID, err := repo.GetCurrentOrganizationID(ctx, userID)
	if err != nil {
		t.Fatalf("GetCurrentOrganizationID() error = %v", err)
	}
	if currentID != secondOrganizationID {
		t.Fatalf("current organization = %s, want %s", currentID, secondOrganizationID)
	}
	defaultMembership, err := client.UserOrganization.Query().
		Where(
			userorganization.UserIDEQ(userID),
			userorganization.OrganizationIDEQ(defaultOrganizationID),
		).
		Only(ctx)
	if err != nil {
		t.Fatalf("query default membership: %v", err)
	}
	if defaultMembership.IsPrimary {
		t.Fatal("default membership should not remain current after switch")
	}
}

func TestSwitchCurrentOrganizationRejectsNonMember(t *testing.T) {
	ctx := context.Background()
	repo, client := newOrganizationTestRepo(t)
	defer client.Close()

	const secondOrganizationID = "5d76c984-7e30-4b87-8433-ff271f902489"
	createOrganization(t, client, defaultOrganizationID, "默认组织", "default")
	createOrganization(t, client, secondOrganizationID, "第二组织", "second")

	if err := repo.SwitchCurrentOrganization(ctx, "user-2", secondOrganizationID); err == nil {
		t.Fatal("expected non-member switch to fail")
	}
}

func TestEnsureUserInDefaultOrganizationPreservesExistingCurrentOrganization(t *testing.T) {
	ctx := context.Background()
	repo, client := newOrganizationTestRepo(t)
	defer client.Close()

	const userID = "user-3"
	const secondOrganizationID = "5d76c984-7e30-4b87-8433-ff271f902489"
	createOrganization(t, client, defaultOrganizationID, "默认组织", "default")
	createOrganization(t, client, secondOrganizationID, "第二组织", "second")
	if _, err := client.UserOrganization.Create().
		SetUserID(userID).
		SetOrganizationID(secondOrganizationID).
		SetStatus(true).
		SetIsPrimary(true).
		Save(ctx); err != nil {
		t.Fatalf("create current membership: %v", err)
	}

	if err := repo.EnsureUserInDefaultOrganization(ctx, userID); err != nil {
		t.Fatalf("EnsureUserInDefaultOrganization() error = %v", err)
	}
	currentID, err := repo.GetCurrentOrganizationID(ctx, userID)
	if err != nil {
		t.Fatalf("GetCurrentOrganizationID() error = %v", err)
	}
	if currentID != secondOrganizationID {
		t.Fatalf("current organization = %s, want %s", currentID, secondOrganizationID)
	}
	defaultMembership, err := client.UserOrganization.Query().
		Where(
			userorganization.UserIDEQ(userID),
			userorganization.OrganizationIDEQ(defaultOrganizationID),
		).
		Only(ctx)
	if err != nil {
		t.Fatalf("query default membership: %v", err)
	}
	if defaultMembership.IsPrimary {
		t.Fatal("default membership should not replace existing current organization")
	}
}

func TestGetUserRoleBindingsFallsBackToDefaultRole(t *testing.T) {
	ctx := context.Background()
	repo, client := newOrganizationTestRepo(t)
	defer client.Close()

	createOrganization(t, client, defaultOrganizationID, "默认组织", "default")
	createDefaultRole(t, client, 7)

	items, err := repo.GetUserRoleBindings(ctx, "user-default", defaultOrganizationID)
	if err != nil {
		t.Fatalf("GetUserRoleBindings() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("GetUserRoleBindings() len = %d, want 1", len(items))
	}
	if items[0].UserID != "user-default" || items[0].RoleID != 7 {
		t.Fatalf("fallback binding = %#v", items[0])
	}
	if items[0].ID != 0 {
		t.Fatalf("fallback binding should be synthetic, got id=%d", items[0].ID)
	}
}

func TestListUserRoleBindingsIncludesDefaultRoleForUnboundUsers(t *testing.T) {
	ctx := context.Background()
	client, err := ent.Open("sqlite3", "file:organization-role-list?mode=memory&cache=shared&_fk=1")
	if err != nil {
		t.Fatalf("open sqlite ent client: %v", err)
	}
	defer client.Close()
	if err := client.Schema.Create(ctx); err != nil {
		t.Fatalf("create ent schema: %v", err)
	}
	userClient := &captureUserServiceClient{
		getUserListReply: &userv1.GetUserListReply{
			Total: 2,
			Items: []*userv1.UserListItem{
				{Id: "user-a"},
				{Id: "user-b"},
			},
		},
	}
	repo := &adminRepo{data: &Data{db: client, userClient: userClient}}
	createOrganization(t, client, defaultOrganizationID, "默认组织", "default")
	createDefaultRole(t, client, 8)
	if _, err := client.UserRoleBinding.Create().
		SetUserID("user-a").
		SetRoleID(99).
		SetOrganizationID(defaultOrganizationID).
		Save(ctx); err != nil {
		t.Fatalf("create explicit user-role binding: %v", err)
	}

	items, err := repo.ListUserRoleBindings(ctx)
	if err != nil {
		t.Fatalf("ListUserRoleBindings() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("ListUserRoleBindings() len = %d, want 2", len(items))
	}
	if items[0].UserID != "user-a" || items[0].RoleID != 99 {
		t.Fatalf("first binding = %#v", items[0])
	}
	if items[1].UserID != "user-b" || items[1].RoleID != 8 {
		t.Fatalf("second binding = %#v", items[1])
	}
}

func TestProjectionSourceListsOnlyExplicitUserRoleBindings(t *testing.T) {
	ctx := context.Background()
	client, err := ent.Open("sqlite3", "file:projection-bindings?mode=memory&cache=shared&_fk=1")
	if err != nil {
		t.Fatalf("open sqlite ent client: %v", err)
	}
	defer client.Close()
	if err := client.Schema.Create(ctx); err != nil {
		t.Fatalf("create ent schema: %v", err)
	}
	repo := &adminRepo{data: &Data{
		db: client,
		userClient: &captureUserServiceClient{
			getUserListReply: &userv1.GetUserListReply{
				Total: 2,
				Items: []*userv1.UserListItem{
					{Id: "user-a"},
					{Id: "user-b"},
				},
			},
		},
	}}
	createOrganization(t, client, defaultOrganizationID, "默认组织", "default")
	createDefaultRole(t, client, 8)
	if _, err := client.Role.Create().
		SetID(99).
		SetName("管理员").
		SetValue("admin").
		SetOrganizationID(defaultOrganizationID).
		SetStatus(true).
		SetDesc("").
		SetMenus([]int32{}).
		Save(ctx); err != nil {
		t.Fatalf("create admin role: %v", err)
	}
	if _, err := client.UserRoleBinding.Create().
		SetUserID("user-a").
		SetRoleID(99).
		SetOrganizationID(defaultOrganizationID).
		Save(ctx); err != nil {
		t.Fatalf("create explicit user-role binding: %v", err)
	}

	items, err := newAdminProjectionSource(repo).ListProjectionBindings(ctx)
	if err != nil {
		t.Fatalf("ListProjectionBindings() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ListProjectionBindings() len = %d, want 1", len(items))
	}
	if items[0].UserID != "user-a" || items[0].RoleID != 99 || items[0].RoleValue != "admin" {
		t.Fatalf("projected binding = %#v", items[0])
	}
}

func TestRemoveOrganizationDeptBindingDeletesScopedMembership(t *testing.T) {
	ctx := context.Background()
	repo, client := newOrganizationTestRepo(t)
	defer client.Close()

	if _, err := client.UserDeptMembership.Create().
		SetUserID("user-1").
		SetOrganizationID("org-a").
		SetDeptID(101).
		Save(ctx); err != nil {
		t.Fatalf("create user dept membership: %v", err)
	}
	if _, err := client.UserDeptMembership.Create().
		SetUserID("user-1").
		SetOrganizationID("org-b").
		SetDeptID(102).
		Save(ctx); err != nil {
		t.Fatalf("create second user dept membership: %v", err)
	}

	if err := repo.RemoveOrganizationDeptBinding(ctx, "user-1", "org-a"); err != nil {
		t.Fatalf("RemoveOrganizationDeptBinding() error = %v", err)
	}

	exists, err := client.UserDeptMembership.Query().
		Where(
			userdeptmembership.UserIDEQ("user-1"),
			userdeptmembership.OrganizationIDEQ("org-a"),
		).
		Exist(ctx)
	if err != nil {
		t.Fatalf("query removed membership: %v", err)
	}
	if exists {
		t.Fatal("expected scoped dept membership to be removed")
	}
	stillExists, err := client.UserDeptMembership.Query().
		Where(
			userdeptmembership.UserIDEQ("user-1"),
			userdeptmembership.OrganizationIDEQ("org-b"),
		).
		Exist(ctx)
	if err != nil {
		t.Fatalf("query remaining membership: %v", err)
	}
	if !stillExists {
		t.Fatal("expected other organization membership to remain")
	}
}

func TestRemoveDeptBindingsDeletesAllBindingsForDept(t *testing.T) {
	ctx := context.Background()
	repo, client := newOrganizationTestRepo(t)
	defer client.Close()

	if _, err := client.UserDeptMembership.Create().
		SetUserID("user-1").
		SetOrganizationID("org-a").
		SetDeptID(201).
		Save(ctx); err != nil {
		t.Fatalf("create first user dept membership: %v", err)
	}
	if _, err := client.UserDeptMembership.Create().
		SetUserID("user-2").
		SetOrganizationID("org-b").
		SetDeptID(201).
		Save(ctx); err != nil {
		t.Fatalf("create second user dept membership: %v", err)
	}
	if _, err := client.UserDeptMembership.Create().
		SetUserID("user-3").
		SetOrganizationID("org-a").
		SetDeptID(202).
		Save(ctx); err != nil {
		t.Fatalf("create unaffected user dept membership: %v", err)
	}

	if err := repo.RemoveDeptBindings(ctx, 201); err != nil {
		t.Fatalf("RemoveDeptBindings() error = %v", err)
	}

	count, err := client.UserDeptMembership.Query().
		Where(userdeptmembership.DeptIDEQ(201)).
		Count(ctx)
	if err != nil {
		t.Fatalf("query removed dept memberships: %v", err)
	}
	if count != 0 {
		t.Fatalf("remaining bindings for dept 201 = %d, want 0", count)
	}
	otherCount, err := client.UserDeptMembership.Query().
		Where(userdeptmembership.DeptIDEQ(202)).
		Count(ctx)
	if err != nil {
		t.Fatalf("query unaffected dept memberships: %v", err)
	}
	if otherCount != 1 {
		t.Fatalf("remaining bindings for dept 202 = %d, want 1", otherCount)
	}
}

func TestSaveOrganizationPermissionScopesReplacesScopedEntries(t *testing.T) {
	ctx := context.Background()
	repo, client := newOrganizationTestRepo(t)
	defer client.Close()

	createOrganization(t, client, "org-a", "组织 A", "org-a")
	if _, err := client.OrganizationPermissionScope.Create().
		SetOrganizationID("org-a").
		SetPermissionType("menu").
		SetPermissionRef("1").
		Save(ctx); err != nil {
		t.Fatalf("create stale scope: %v", err)
	}
	if _, err := client.OrganizationPermissionScope.Create().
		SetOrganizationID("org-b").
		SetPermissionType("menu").
		SetPermissionRef("9").
		Save(ctx); err != nil {
		t.Fatalf("create other org scope: %v", err)
	}

	items, err := repo.SaveOrganizationPermissionScopes(
		ctx,
		"org-a",
		[]int32{2, 3},
		[]string{"resource-a"},
		"actor-1",
	)
	if err != nil {
		t.Fatalf("SaveOrganizationPermissionScopes() error = %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("scope count = %d, want 3", len(items))
	}
	staleExists, err := client.OrganizationPermissionScope.Query().
		Where(
			organizationpermissionscope.OrganizationIDEQ("org-a"),
			organizationpermissionscope.PermissionRefEQ("1"),
		).
		Exist(ctx)
	if err != nil {
		t.Fatalf("query stale scope: %v", err)
	}
	if staleExists {
		t.Fatal("expected stale organization scope to be replaced")
	}
	otherExists, err := client.OrganizationPermissionScope.Query().
		Where(organizationpermissionscope.OrganizationIDEQ("org-b")).
		Exist(ctx)
	if err != nil {
		t.Fatalf("query other org scope: %v", err)
	}
	if !otherExists {
		t.Fatal("expected other organization scope to remain")
	}
}

func newOrganizationTestRepo(t *testing.T) (*adminRepo, *ent.Client) {
	t.Helper()
	client, err := ent.Open("sqlite3", "file:organization?mode=memory&cache=shared&_fk=1")
	if err != nil {
		t.Fatalf("open sqlite ent client: %v", err)
	}
	if err := client.Schema.Create(context.Background()); err != nil {
		_ = client.Close()
		t.Fatalf("create ent schema: %v", err)
	}
	return &adminRepo{data: &Data{db: client}}, client
}

func createOrganization(t *testing.T, client *ent.Client, id string, name string, code string) {
	t.Helper()
	if _, err := client.Organization.Create().
		SetID(id).
		SetName(name).
		SetCode(code).
		SetSort(0).
		SetStatus(true).
		SetDesc("").
		SetExtension("").
		Save(context.Background()); err != nil {
		t.Fatalf("create organization %s: %v", id, err)
	}
}

type captureUserServiceClient struct {
	outgoingMetadata metadata.MD
	getUserListReply *userv1.GetUserListReply
}

func (c *captureUserServiceClient) GetUserList(ctx context.Context, _ *userv1.GetUserParams, _ ...grpc.CallOption) (*userv1.GetUserListReply, error) {
	c.outgoingMetadata, _ = metadata.FromOutgoingContext(ctx)
	if c.getUserListReply != nil {
		return c.getUserListReply, nil
	}
	return &userv1.GetUserListReply{}, nil
}

func (c *captureUserServiceClient) GetUserInfo(context.Context, *userv1.GetUserInfoRequest, ...grpc.CallOption) (*userv1.GetUserInfoReply, error) {
	return nil, nil
}

func (c *captureUserServiceClient) AddUser(context.Context, *userv1.UserListItem, ...grpc.CallOption) (*userv1.UserListItem, error) {
	return nil, nil
}

func (c *captureUserServiceClient) UpdateUser(context.Context, *userv1.UserListItem, ...grpc.CallOption) (*userv1.UserListItem, error) {
	return nil, nil
}

func (c *captureUserServiceClient) DelUser(context.Context, *userv1.DeleteUser, ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (c *captureUserServiceClient) IsUserExist(context.Context, *userv1.IsUserExistsRequest, ...grpc.CallOption) (*userv1.IsUserExistsReply, error) {
	return nil, nil
}

func (c *captureUserServiceClient) ChangePassword(context.Context, *userv1.ChangePasswordRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, nil
}

func (c *captureUserServiceClient) ValidateUserAuth(context.Context, *userv1.ValidateUserAuthRequest, ...grpc.CallOption) (*userv1.ValidateUserAuthReply, error) {
	return nil, nil
}

func (c *captureUserServiceClient) GetUserAuthInfo(context.Context, *userv1.GetUserAuthInfoRequest, ...grpc.CallOption) (*userv1.GetUserAuthInfoReply, error) {
	return nil, nil
}

func (c *captureUserServiceClient) GetWalkRoute(context.Context, *emptypb.Empty, ...grpc.CallOption) (*userv1.GetWalkRouteReply, error) {
	return nil, nil
}

type testTransport struct {
	headers testHeader
}

func (t testTransport) Kind() transport.Kind { return transport.KindHTTP }
func (t testTransport) Endpoint() string     { return "" }
func (t testTransport) Operation() string    { return "" }
func (t testTransport) RequestHeader() transport.Header {
	return t.headers
}
func (t testTransport) ReplyHeader() transport.Header {
	return testHeader{}
}

type testHeader map[string]string

func (h testHeader) Get(key string) string {
	return h[key]
}

func createDefaultRole(t *testing.T, client *ent.Client, id int64) {
	t.Helper()
	if _, err := client.Role.Create().
		SetID(id).
		SetName("默认角色").
		SetValue("default").
		SetOrganizationID(defaultOrganizationID).
		SetStatus(true).
		SetDesc("").
		SetMenus([]int32{}).
		Save(context.Background()); err != nil {
		t.Fatalf("create default role: %v", err)
	}
}

func (h testHeader) Set(key string, value string) {
	h[key] = value
}

func (h testHeader) Add(key string, value string) {
	h[key] = value
}

func (h testHeader) Keys() []string {
	keys := make([]string, 0, len(h))
	for key := range h {
		keys = append(keys, key)
	}
	return keys
}

func (h testHeader) Values(key string) []string {
	value := h.Get(key)
	if value == "" {
		return nil
	}
	return []string{value}
}

var _ userv1.UserServiceClient = (*captureUserServiceClient)(nil)
var _ transport.Transporter = (*testTransport)(nil)
var _ transport.Header = (testHeader)(nil)
