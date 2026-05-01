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

func TestCheckAuthorizationSupportsServiceAndScope(t *testing.T) {
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
		SourceService: "admin",
		Revision:      1,
		Roles: []*v1.PolicyRole{
			{
				Name:    "Admin",
				Value:   "manager",
				Status:  true,
				Service: "admin",
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
				Service:        "admin",
			},
		},
		Bindings: []*v1.PolicyUserRoleBinding{
			{
				UserId:    "u1",
				RoleValue: "manager",
				Service:   "admin",
				ScopeId:   "dept-a",
			},
		},
	}
	uc.applyPermissionSnapshot(req)

	allowed, err := uc.CheckAuthorization(context.Background(), &v1.CheckAuthorizationRequest{
		UserId:  "u1",
		Path:    "/admin-api/v1/resources",
		Method:  "GET",
		Service: "admin",
		ScopeId: "dept-a",
	})
	if err != nil {
		t.Fatalf("check authorization: %v", err)
	}
	if !allowed.Allowed {
		t.Fatalf("expected scoped authorization to be allowed")
	}

	denied, err := uc.CheckAuthorization(context.Background(), &v1.CheckAuthorizationRequest{
		UserId:  "u1",
		Path:    "/admin-api/v1/resources",
		Method:  "GET",
		Service: "admin",
		ScopeId: "dept-b",
	})
	if err != nil {
		t.Fatalf("check authorization with wrong scope: %v", err)
	}
	if denied.Allowed {
		t.Fatalf("expected authorization to be denied for a different scope")
	}
}

func TestCheckAuthorizationPrefersDomainCode(t *testing.T) {
	uc := newTestAuthUsecase(t)
	err := uc.RegisterPermissionSnapshot(context.Background(), &v1.RegisterPermissionSnapshotRequest{
		SourceService: "admin",
		Roles: []*v1.PolicyRole{
			{
				Value:      "admin",
				Status:     true,
				Service:    "admin",
				DomainCode: "platform",
				Resources: []*v1.PolicyRoleResource{
					{Type: "api", Value: "api_catalog", Method: "GET"},
				},
			},
		},
		Apis: []*v1.PolicyApi{
			{Path: "/admin-api/v1/apis", Method: "GET", ResourcesGroup: "api_catalog", Service: "admin", DomainCode: "platform"},
		},
		Bindings: []*v1.PolicyUserRoleBinding{
			{UserId: "u1", RoleValue: "admin", Service: "admin", DomainCode: "platform", ScopeId: "global"},
		},
	})
	if err != nil {
		t.Fatalf("RegisterPermissionSnapshot() error = %v", err)
	}

	allowed, err := uc.CheckAuthorization(context.Background(), &v1.CheckAuthorizationRequest{
		UserId:     "u1",
		Path:       "/admin-api/v1/apis",
		Method:     "GET",
		Service:    "admin",
		DomainCode: "platform",
		ScopeId:    "global",
	})
	if err != nil {
		t.Fatalf("CheckAuthorization() error = %v", err)
	}
	if !allowed.Allowed {
		t.Fatalf("expected authorization to use domain_code namespace")
	}
}
