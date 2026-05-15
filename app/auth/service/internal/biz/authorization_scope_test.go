package biz

import (
	"context"
	"io"
	"testing"

	v1 "base-server/api/gen/go/auth/service/v1"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/go-kratos/kratos/v2/log"
)

func TestCheckAuthorizationUsesOrganizationDomain(t *testing.T) {
	m, err := model.NewModelFromString(textModel)
	if err != nil {
		t.Fatalf("new model: %v", err)
	}
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		t.Fatalf("new enforcer: %v", err)
	}

	uc := &AuthUsecase{
		e:   e,
		log: log.NewHelper(log.NewStdLogger(io.Discard)),
	}

	req := &v1.RegisterPermissionSnapshotRequest{
		Revision: 1,
		Roles: []*v1.PolicyRole{
			{
				Name:           "Admin",
				Value:          "manager",
				Status:         true,
				OrganizationId: "dept-a",
				Resources: []*v1.PolicyRoleResource{
					{
						Type:   "api",
						Value:  "system",
						Method: "GET",
					},
				},
			},
		},
		Apis: []*v1.PolicyApi{
			{
				Path:           "/admin-api/v1/resources",
				Method:         "GET",
				ResourcesGroup: "system",
			},
		},
		Bindings: []*v1.PolicyUserRoleBinding{
			{
				UserId:         "u1",
				RoleValue:      "manager",
				OrganizationId: "dept-a",
			},
		},
	}
	uc.applyPermissionSnapshot(req)

	allowed, err := uc.CheckAuthorization(context.Background(), &v1.CheckAuthorizationRequest{
		UserId:         "u1",
		Path:           "/admin-api/v1/resources",
		Method:         "GET",
		OrganizationId: "dept-a",
	})
	if err != nil {
		t.Fatalf("check authorization: %v", err)
	}
	if !allowed.Allowed {
		t.Fatalf("expected organization authorization to be allowed")
	}

	denied, err := uc.CheckAuthorization(context.Background(), &v1.CheckAuthorizationRequest{
		UserId:         "u1",
		Path:           "/admin-api/v1/resources",
		Method:         "GET",
		OrganizationId: "dept-b",
	})
	if err != nil {
		t.Fatalf("check authorization with wrong organization: %v", err)
	}
	if denied.Allowed {
		t.Fatalf("expected authorization to be denied for a different organization")
	}
}

func TestCheckAuthorizationUsesGlobalAPIGroups(t *testing.T) {
	uc := newTestAuthUsecase(t)
	err := uc.RegisterPermissionSnapshot(context.Background(), &v1.RegisterPermissionSnapshotRequest{
		Roles: []*v1.PolicyRole{
			{
				Value:          "admin",
				Status:         true,
				OrganizationId: "org-a",
				Resources: []*v1.PolicyRoleResource{
					{Type: "api", Value: "api_catalog", Method: "GET"},
				},
			},
		},
		Apis: []*v1.PolicyApi{
			{Path: "/admin-api/v1/apis", Method: "GET", ResourcesGroup: "api_catalog"},
		},
		Bindings: []*v1.PolicyUserRoleBinding{
			{UserId: "u1", RoleValue: "admin", OrganizationId: "org-a"},
		},
	})
	if err != nil {
		t.Fatalf("RegisterPermissionSnapshot() error = %v", err)
	}

	allowed, err := uc.CheckAuthorization(context.Background(), &v1.CheckAuthorizationRequest{
		UserId:         "u1",
		Path:           "/admin-api/v1/apis",
		Method:         "GET",
		OrganizationId: "org-a",
	})
	if err != nil {
		t.Fatalf("CheckAuthorization() error = %v", err)
	}
	if !allowed.Allowed {
		t.Fatalf("expected authorization to use global API groups")
	}
}

func TestCheckAuthorizationFallsBackToDefaultRoleForUnboundUser(t *testing.T) {
	uc := newTestAuthUsecase(t)
	err := uc.RegisterPermissionSnapshot(context.Background(), &v1.RegisterPermissionSnapshotRequest{
		Roles: []*v1.PolicyRole{
			{
				Value:          "default",
				Status:         true,
				OrganizationId: DefaultOrganizationID,
				Resources: []*v1.PolicyRoleResource{
					{Type: "api", Value: "profile", Method: "GET"},
				},
			},
		},
		Apis: []*v1.PolicyApi{
			{Path: "/user-api/v1/userinfo", Method: "GET", ResourcesGroup: "profile"},
		},
	})
	if err != nil {
		t.Fatalf("RegisterPermissionSnapshot() error = %v", err)
	}

	allowed, err := uc.CheckAuthorization(context.Background(), &v1.CheckAuthorizationRequest{
		UserId: "new-user",
		Path:   "/user-api/v1/userinfo",
		Method: "GET",
	})
	if err != nil {
		t.Fatalf("CheckAuthorization() error = %v", err)
	}
	if !allowed.Allowed {
		t.Fatalf("expected unbound user to use default role fallback")
	}
}

func TestCheckAuthorizationFallsBackToAdminProjectionDefaultRole(t *testing.T) {
	uc := newTestAuthUsecase(t)
	err := uc.RegisterPermissionSnapshot(context.Background(), &v1.RegisterPermissionSnapshotRequest{
		Roles: []*v1.PolicyRole{
			{
				Value:          "default",
				Status:         true,
				OrganizationId: DefaultOrganizationID,
				Resources: []*v1.PolicyRoleResource{
					{Type: "api", Value: "profile", Method: "GET"},
				},
			},
		},
		Apis: []*v1.PolicyApi{
			{Path: "/user-api/v1/users/{user_id}", Method: "GET", ResourcesGroup: "profile"},
		},
	})
	if err != nil {
		t.Fatalf("RegisterPermissionSnapshot() error = %v", err)
	}

	allowed, err := uc.CheckAuthorization(context.Background(), &v1.CheckAuthorizationRequest{
		UserId: "new-user",
		Path:   "/user-api/v1/users/new-user",
		Method: "GET",
	})
	if err != nil {
		t.Fatalf("CheckAuthorization() error = %v", err)
	}
	if !allowed.Allowed {
		t.Fatalf("expected user-service request to use admin projection default role fallback")
	}
}

func TestCheckAuthorizationDoesNotOverrideExplicitBinding(t *testing.T) {
	uc := newTestAuthUsecase(t)
	err := uc.RegisterPermissionSnapshot(context.Background(), &v1.RegisterPermissionSnapshotRequest{
		Roles: []*v1.PolicyRole{
			{
				Value:          "default",
				Status:         true,
				OrganizationId: DefaultOrganizationID,
				Resources: []*v1.PolicyRoleResource{
					{Type: "api", Value: "profile", Method: "GET"},
				},
			},
			{
				Value:          "guest",
				Status:         true,
				OrganizationId: DefaultOrganizationID,
				Resources: []*v1.PolicyRoleResource{
					{Type: "api", Value: "guest_only", Method: "GET"},
				},
			},
		},
		Apis: []*v1.PolicyApi{
			{Path: "/user-api/v1/userinfo", Method: "GET", ResourcesGroup: "profile"},
		},
		Bindings: []*v1.PolicyUserRoleBinding{
			{UserId: "bound-user", RoleValue: "guest"},
		},
	})
	if err != nil {
		t.Fatalf("RegisterPermissionSnapshot() error = %v", err)
	}

	allowed, err := uc.CheckAuthorization(context.Background(), &v1.CheckAuthorizationRequest{
		UserId: "bound-user",
		Path:   "/user-api/v1/userinfo",
		Method: "GET",
	})
	if err != nil {
		t.Fatalf("CheckAuthorization() error = %v", err)
	}
	if allowed.Allowed {
		t.Fatalf("expected explicit non-default binding to prevent default fallback")
	}
}

func TestCheckAuthorizationGlobalRootIgnoresDomain(t *testing.T) {
	uc := newTestAuthUsecase(t)
	err := uc.RegisterPermissionSnapshot(context.Background(), &v1.RegisterPermissionSnapshotRequest{
		Roles: []*v1.PolicyRole{
			{
				Value:          "limited",
				Status:         true,
				OrganizationId: "org-a",
				Resources: []*v1.PolicyRoleResource{
					{Type: "api", Value: "limited", Method: "GET"},
				},
			},
		},
		Apis: []*v1.PolicyApi{
			{Path: "/admin-api/v1/roles", Method: "DELETE", ResourcesGroup: "role"},
		},
		Bindings: []*v1.PolicyUserRoleBinding{
			{UserId: "root-user", RoleValue: "root", OrganizationId: "org-a"},
		},
	})
	if err != nil {
		t.Fatalf("RegisterPermissionSnapshot() error = %v", err)
	}

	allowed, err := uc.CheckAuthorization(context.Background(), &v1.CheckAuthorizationRequest{
		UserId:         "root-user",
		Path:           "/admin-api/v1/roles",
		Method:         "DELETE",
		OrganizationId: "org-b",
	})
	if err != nil {
		t.Fatalf("CheckAuthorization() error = %v", err)
	}
	if !allowed.Allowed {
		t.Fatalf("expected global root to be allowed in any domain")
	}
}

func TestApplyUserRoleBindingDeltaRemovesRootGlobalDomain(t *testing.T) {
	uc := newTestAuthUsecase(t)
	err := uc.RegisterPermissionSnapshot(context.Background(), &v1.RegisterPermissionSnapshotRequest{
		Roles: []*v1.PolicyRole{
			{
				Value:          "limited",
				Status:         true,
				OrganizationId: "org-a",
				Resources: []*v1.PolicyRoleResource{
					{Type: "api", Value: "limited", Method: "GET"},
				},
			},
		},
		Apis: []*v1.PolicyApi{
			{Path: "/admin-api/v1/roles", Method: "DELETE", ResourcesGroup: "role"},
		},
		Bindings: []*v1.PolicyUserRoleBinding{
			{UserId: "root-user", RoleValue: "root", OrganizationId: "org-a"},
		},
	})
	if err != nil {
		t.Fatalf("RegisterPermissionSnapshot() error = %v", err)
	}

	err = uc.ApplyUserRoleBindingDelta(context.Background(), &v1.ApplyUserRoleBindingDeltaRequest{
		Before: &v1.PolicyUserRoleBinding{UserId: "root-user", RoleValue: "root", OrganizationId: "org-a"},
	})
	if err != nil {
		t.Fatalf("ApplyUserRoleBindingDelta() error = %v", err)
	}

	allowed, err := uc.CheckAuthorization(context.Background(), &v1.CheckAuthorizationRequest{
		UserId:         "root-user",
		Path:           "/admin-api/v1/roles",
		Method:         "DELETE",
		OrganizationId: "org-b",
	})
	if err != nil {
		t.Fatalf("CheckAuthorization() error = %v", err)
	}
	if allowed.Allowed {
		t.Fatalf("expected deleted root binding to lose global authorization")
	}
}

func TestApplyUserRoleBindingDeltaPreservesOtherRolesInSameDomain(t *testing.T) {
	uc := newTestAuthUsecase(t)
	err := uc.RegisterPermissionSnapshot(context.Background(), &v1.RegisterPermissionSnapshotRequest{
		Roles: []*v1.PolicyRole{
			{
				Value:          "reader",
				Status:         true,
				OrganizationId: "org-a",
				Resources: []*v1.PolicyRoleResource{
					{Type: "api", Value: "profile", Method: "GET"},
				},
			},
			{
				Value:          "writer",
				Status:         true,
				OrganizationId: "org-a",
				Resources: []*v1.PolicyRoleResource{
					{Type: "api", Value: "profile", Method: "POST"},
				},
			},
		},
		Apis: []*v1.PolicyApi{
			{Path: "/user-api/v1/userinfo", Method: "GET", ResourcesGroup: "profile"},
		},
		Bindings: []*v1.PolicyUserRoleBinding{
			{UserId: "u1", RoleValue: "reader", OrganizationId: "org-a"},
		},
	})
	if err != nil {
		t.Fatalf("RegisterPermissionSnapshot() error = %v", err)
	}

	err = uc.ApplyUserRoleBindingDelta(context.Background(), &v1.ApplyUserRoleBindingDeltaRequest{
		After: &v1.PolicyUserRoleBinding{UserId: "u1", RoleValue: "writer", OrganizationId: "org-a"},
	})
	if err != nil {
		t.Fatalf("ApplyUserRoleBindingDelta() error = %v", err)
	}

	allowed, err := uc.CheckAuthorization(context.Background(), &v1.CheckAuthorizationRequest{
		UserId:         "u1",
		Path:           "/user-api/v1/userinfo",
		Method:         "GET",
		OrganizationId: "org-a",
	})
	if err != nil {
		t.Fatalf("CheckAuthorization() error = %v", err)
	}
	if !allowed.Allowed {
		t.Fatalf("expected existing reader role to be preserved when writer role is added")
	}
}
