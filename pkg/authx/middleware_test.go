package authx

import (
	"context"
	"testing"
)

func TestNewWhitelistMatcherMatchesOperationVariants(t *testing.T) {
	matcher := NewWhitelistMatcher([]string{
		" /api.auth.service.v1.AuthService/RegisterPermissionSnapshot ",
	})

	cases := []string{
		"/api.auth.service.v1.AuthService/RegisterPermissionSnapshot",
		"api.auth.service.v1.AuthService/RegisterPermissionSnapshot",
		"RegisterPermissionSnapshot",
	}
	for _, operation := range cases {
		t.Run(operation, func(t *testing.T) {
			if matcher(context.Background(), operation) {
				t.Fatalf("expected whitelisted operation %q to skip middleware", operation)
			}
		})
	}
}

func TestNewWhitelistMatcherRequiresMiddlewareForUnknownOperation(t *testing.T) {
	matcher := NewWhitelistMatcher([]string{
		"/api.auth.service.v1.AuthService/RegisterPermissionSnapshot",
	})

	if !matcher(context.Background(), "/api.auth.service.v1.AuthService/CheckAuthorization") {
		t.Fatal("expected unknown operation to require middleware")
	}
}
