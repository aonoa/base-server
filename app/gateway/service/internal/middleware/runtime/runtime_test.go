package runtime

import (
	"net/http/httptest"
	"testing"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func TestRouteMatches(t *testing.T) {
	rule := NewRouteRule("/admin-api/v1/*", "GET", "example.com")
	req := httptest.NewRequest("GET", "http://example.com:8000/admin-api/v1/users", nil)
	if !RouteMatches(rule, req) {
		t.Fatal("expected route to match")
	}
}

func TestRouteMatchesMethodMismatch(t *testing.T) {
	rule := NewRouteRule("/admin-api/v1/*", "POST", "")
	req := httptest.NewRequest("GET", "http://example.com/admin-api/v1/users", nil)
	if RouteMatches(rule, req) {
		t.Fatal("expected route not to match")
	}
}

func TestParseAuthorizationToken(t *testing.T) {
	if got := ParseAuthorizationToken("Bearer token-value"); got != "token-value" {
		t.Fatalf("unexpected token: %q", got)
	}
	if got := ParseAuthorizationToken("Basic token-value"); got != "" {
		t.Fatalf("expected empty token, got %q", got)
	}
}

func TestParseHS256Token(t *testing.T) {
	token, err := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, jwtv5.MapClaims{"user_id": "u1"}).SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	claims, err := ParseHS256Token(token, "secret")
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if claims["user_id"] != "u1" {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}
