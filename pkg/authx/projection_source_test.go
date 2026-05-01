package authx

import (
	"context"
	"testing"
)

type stubProjectionSource struct {
	sourceService string
	roles         []ProjectionRole
	apis          []ProjectionAPI
	bindings      []UserRoleBindingProjection
}

func (s stubProjectionSource) SourceService() string { return s.sourceService }

func (s stubProjectionSource) ListProjectionRoles(context.Context) ([]ProjectionRole, error) {
	return s.roles, nil
}

func (s stubProjectionSource) ListProjectionAPIs(context.Context) ([]ProjectionAPI, error) {
	return s.apis, nil
}

func (s stubProjectionSource) ListProjectionBindings(context.Context) ([]UserRoleBindingProjection, error) {
	return s.bindings, nil
}

func TestBuildSnapshotFromSource(t *testing.T) {
	source := stubProjectionSource{
		sourceService: "crm",
		roles:         []ProjectionRole{{Value: "manager"}},
		apis:          []ProjectionAPI{{Path: "/crm-api/v1/orders"}},
		bindings:      []UserRoleBindingProjection{{UserID: "u1", RoleValue: "manager"}},
	}

	snapshot, err := BuildSnapshotFromSource(context.Background(), source)
	if err != nil {
		t.Fatalf("build snapshot from source: %v", err)
	}
	if snapshot.SourceService != "crm" {
		t.Fatalf("expected source service crm, got %q", snapshot.SourceService)
	}
	if len(snapshot.Roles) != 1 || snapshot.Roles[0].Value != "manager" {
		t.Fatalf("unexpected roles: %#v", snapshot.Roles)
	}
	if len(snapshot.APIs) != 1 || snapshot.APIs[0].Path != "/crm-api/v1/orders" {
		t.Fatalf("unexpected apis: %#v", snapshot.APIs)
	}
	if len(snapshot.Bindings) != 1 || snapshot.Bindings[0].UserID != "u1" {
		t.Fatalf("unexpected bindings: %#v", snapshot.Bindings)
	}
	if snapshot.Revision == 0 {
		t.Fatalf("expected non-zero revision")
	}
}
