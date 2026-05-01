package biz

import (
	"context"
	"io"
	"testing"

	v1 "base-server/api/gen/go/auth/service/v1"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/go-kratos/kratos/v2/log"
)

type noopAdapter struct{}

func (noopAdapter) LoadPolicy(model model.Model) error { return nil }
func (noopAdapter) SavePolicy(model model.Model) error { return nil }
func (noopAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	return nil
}
func (noopAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return nil
}
func (noopAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return nil
}

var _ persist.Adapter = noopAdapter{}

func newTestAuthUsecase(t *testing.T) *AuthUsecase {
	t.Helper()
	m, err := model.NewModelFromString(textModel)
	if err != nil {
		t.Fatalf("new model: %v", err)
	}
	e, err := casbin.NewEnforcer(m)
	if err != nil {
		t.Fatalf("new enforcer: %v", err)
	}
	e.SetAdapter(noopAdapter{})
	return &AuthUsecase{
		e:   e,
		log: log.NewHelper(log.NewStdLogger(io.Discard)),
	}
}

func TestRegisterPermissionSnapshotMergesDifferentSources(t *testing.T) {
	uc := newTestAuthUsecase(t)

	err := uc.RegisterPermissionSnapshot(context.Background(), &v1.RegisterPermissionSnapshotRequest{
		SourceService: "crm",
		Revision:      1,
		Roles: []*v1.PolicyRole{
			{
				Value:   "manager",
				Status:  true,
				Service: "crm",
				Resources: []*v1.PolicyRoleResource{
					{Type: "api", Value: "orders", Method: "GET"},
				},
			},
		},
		Apis: []*v1.PolicyApi{
			{Path: "/crm-api/v1/orders", Method: "GET", ResourcesGroup: "orders", Service: "crm"},
		},
		Bindings: []*v1.PolicyUserRoleBinding{
			{UserId: "u1", RoleValue: "manager", Service: "crm", ScopeId: "tenant-a"},
		},
	})
	if err != nil {
		t.Fatalf("register crm snapshot: %v", err)
	}

	err = uc.RegisterPermissionSnapshot(context.Background(), &v1.RegisterPermissionSnapshotRequest{
		SourceService: "oms",
		Revision:      1,
		Roles: []*v1.PolicyRole{
			{
				Value:   "operator",
				Status:  true,
				Service: "oms",
				Resources: []*v1.PolicyRoleResource{
					{Type: "api", Value: "shipments", Method: "GET"},
				},
			},
		},
		Apis: []*v1.PolicyApi{
			{Path: "/oms-api/v1/shipments", Method: "GET", ResourcesGroup: "shipments", Service: "oms"},
		},
		Bindings: []*v1.PolicyUserRoleBinding{
			{UserId: "u2", RoleValue: "operator", Service: "oms", ScopeId: "warehouse-a"},
		},
	})
	if err != nil {
		t.Fatalf("register oms snapshot: %v", err)
	}

	res, err := uc.CheckAuthorization(context.Background(), &v1.CheckAuthorizationRequest{
		UserId:  "u1",
		Path:    "/crm-api/v1/orders",
		Method:  "GET",
		Service: "crm",
		ScopeId: "tenant-a",
	})
	if err != nil {
		t.Fatalf("crm check authorization: %v", err)
	}
	if !res.Allowed {
		t.Fatalf("expected crm projection to remain after oms snapshot")
	}

	res, err = uc.CheckAuthorization(context.Background(), &v1.CheckAuthorizationRequest{
		UserId:  "u2",
		Path:    "/oms-api/v1/shipments",
		Method:  "GET",
		Service: "oms",
		ScopeId: "warehouse-a",
	})
	if err != nil {
		t.Fatalf("oms check authorization: %v", err)
	}
	if !res.Allowed {
		t.Fatalf("expected oms projection to be allowed")
	}
}

func TestRegisterPermissionSnapshotReplacesSingleSourceOnly(t *testing.T) {
	uc := newTestAuthUsecase(t)

	err := uc.RegisterPermissionSnapshot(context.Background(), &v1.RegisterPermissionSnapshotRequest{
		SourceService: "crm",
		Revision:      1,
		Roles: []*v1.PolicyRole{
			{
				Value:   "manager",
				Status:  true,
				Service: "crm",
				Resources: []*v1.PolicyRoleResource{
					{Type: "api", Value: "orders", Method: "GET"},
				},
			},
		},
		Apis: []*v1.PolicyApi{
			{Path: "/crm-api/v1/orders", Method: "GET", ResourcesGroup: "orders", Service: "crm"},
		},
		Bindings: []*v1.PolicyUserRoleBinding{
			{UserId: "u1", RoleValue: "manager", Service: "crm", ScopeId: "tenant-a"},
		},
	})
	if err != nil {
		t.Fatalf("register initial crm snapshot: %v", err)
	}

	err = uc.RegisterPermissionSnapshot(context.Background(), &v1.RegisterPermissionSnapshotRequest{
		SourceService: "crm",
		Revision:      2,
		Roles: []*v1.PolicyRole{
			{
				Value:   "auditor",
				Status:  true,
				Service: "crm",
				Resources: []*v1.PolicyRoleResource{
					{Type: "api", Value: "reports", Method: "GET"},
				},
			},
		},
		Apis: []*v1.PolicyApi{
			{Path: "/crm-api/v1/reports", Method: "GET", ResourcesGroup: "reports", Service: "crm"},
		},
		Bindings: []*v1.PolicyUserRoleBinding{
			{UserId: "u1", RoleValue: "auditor", Service: "crm", ScopeId: "tenant-a"},
		},
	})
	if err != nil {
		t.Fatalf("register replacement crm snapshot: %v", err)
	}

	oldRes, err := uc.CheckAuthorization(context.Background(), &v1.CheckAuthorizationRequest{
		UserId:  "u1",
		Path:    "/crm-api/v1/orders",
		Method:  "GET",
		Service: "crm",
		ScopeId: "tenant-a",
	})
	if err != nil {
		t.Fatalf("old crm check authorization: %v", err)
	}
	if oldRes.Allowed {
		t.Fatalf("expected old crm source projection to be replaced")
	}

	newRes, err := uc.CheckAuthorization(context.Background(), &v1.CheckAuthorizationRequest{
		UserId:  "u1",
		Path:    "/crm-api/v1/reports",
		Method:  "GET",
		Service: "crm",
		ScopeId: "tenant-a",
	})
	if err != nil {
		t.Fatalf("new crm check authorization: %v", err)
	}
	if !newRes.Allowed {
		t.Fatalf("expected replacement crm projection to be allowed")
	}
}
