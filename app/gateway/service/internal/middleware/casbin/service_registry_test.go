package casbin

import (
	"context"
	"errors"
	"testing"
	"time"

	adminv1 "base-server/api/gen/go/admin/service/v1"
)

func TestSelectServiceCodeUsesLongestPrefix(t *testing.T) {
	entries := buildServiceRegistryEntries([]*adminv1.ServiceRegistryItem{
		{ServiceCode: "admin", HttpPrefix: "/admin-api", Status: 1},
		{ServiceCode: "admin-v2", HttpPrefix: "/admin-api/v2", Status: 1},
		{ServiceCode: "disabled", HttpPrefix: "/disabled", Status: 0},
	})

	if got := selectServiceCode(entries, "/admin-api/v2/users"); got != "admin-v2" {
		t.Fatalf("expected admin-v2, got %q", got)
	}
	if got := selectServiceCode(entries, "/admin-api/v1/users"); got != "admin" {
		t.Fatalf("expected admin, got %q", got)
	}
	if got := selectServiceCode(entries, "/disabled/ping"); got != "" {
		t.Fatalf("expected empty result for disabled service, got %q", got)
	}
}

func TestServiceRegistryResolverFallsBackToCachedEntries(t *testing.T) {
	calls := 0
	resolver := newServiceRegistryResolver(func(context.Context) ([]*adminv1.ServiceRegistryItem, error) {
		calls++
		if calls == 1 {
			return []*adminv1.ServiceRegistryItem{
				{ServiceCode: "user", HttpPrefix: "/user-api", Status: 1},
			}, nil
		}
		return nil, errors.New("admin unavailable")
	})
	now := time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	resolver.now = func() time.Time { return now }
	resolver.ttl = time.Second

	code, err := resolver.Resolve(context.Background(), "/user-api/v1/profile")
	if err != nil {
		t.Fatalf("unexpected first resolve error: %v", err)
	}
	if code != "user" {
		t.Fatalf("expected user, got %q", code)
	}

	now = now.Add(2 * time.Second)
	code, err = resolver.Resolve(context.Background(), "/user-api/v1/profile")
	if err != nil {
		t.Fatalf("expected cached fallback, got error: %v", err)
	}
	if code != "user" {
		t.Fatalf("expected user from cache, got %q", code)
	}
}

func TestNormalizeHTTPPrefix(t *testing.T) {
	tests := map[string]string{
		"admin-api/":  "/admin-api",
		"/admin-api/": "/admin-api",
		"/":           "/",
		"":            "",
	}
	for in, want := range tests {
		if got := normalizeHTTPPrefix(in); got != want {
			t.Fatalf("normalizeHTTPPrefix(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSelectAPIOwnershipMatchesPathMethodAndDomain(t *testing.T) {
	entries := buildAPICatalogEntries([]*adminv1.ApiListItem{
		{Path: "/admin-api/v1/apis/{id}", Method: "GET", ServiceCode: "admin", DomainCode: "platform", ResourcesGroup: "api_catalog"},
		{Path: "/admin-api/v1/apis/{id}", Method: "DELETE", ServiceCode: "admin", DomainCode: "platform", ResourcesGroup: "api_catalog_delete"},
	})

	ownership := selectAPIOwnership(entries, "/admin-api/v1/apis/123", "GET")
	if ownership.ServiceCode != "admin" || ownership.DomainCode != "platform" || ownership.ResourceGroup != "api_catalog" {
		t.Fatalf("unexpected ownership: %+v", ownership)
	}

	ownership = selectAPIOwnership(entries, "/admin-api/v1/apis/123", "POST")
	if ownership != (apiOwnership{}) {
		t.Fatalf("expected empty ownership for method mismatch, got %+v", ownership)
	}
}
