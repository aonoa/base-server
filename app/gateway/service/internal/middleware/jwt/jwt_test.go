package jwt

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/go-kratos/gateway/api/gateway/config/v1"
	gwmiddleware "github.com/go-kratos/gateway/middleware"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"google.golang.org/protobuf/types/known/anypb"

	jwtv1 "base-server/app/gateway/service/internal/middleware/jwt/v1"
)

func TestMiddlewareRejectsMissingToken(t *testing.T) {
	mw := newTestMiddleware(t, &jwtv1.JWT{SigningKey: "secret"})
	resp, err := mw(gwmiddleware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody, Header: make(http.Header)}, nil
	})).RoundTrip(httptest.NewRequest(http.MethodGet, "http://example.com/user-api/v1/profile", nil))
	if err != nil {
		t.Fatalf("round trip: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestMiddlewareAllowsValidToken(t *testing.T) {
	mw := newTestMiddleware(t, &jwtv1.JWT{SigningKey: "secret"})
	token := signedToken(t, "secret", jwtv5.MapClaims{"user_id": "u1", "aud": "login"})
	req := httptest.NewRequest(http.MethodGet, "http://example.com/user-api/v1/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := mw(gwmiddleware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody, Header: make(http.Header)}, nil
	})).RoundTrip(req)
	if err != nil {
		t.Fatalf("round trip: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected next round tripper to run, got %d", resp.StatusCode)
	}
}

func TestMiddlewareRejectsRefreshTokenOnNonRefreshPath(t *testing.T) {
	mw := newTestMiddleware(t, &jwtv1.JWT{SigningKey: "secret"})
	token := signedToken(t, "secret", jwtv5.MapClaims{"user_id": "u1", "aud": "refresh"})
	req := httptest.NewRequest(http.MethodGet, "http://example.com/user-api/v1/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := mw(gwmiddleware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody, Header: make(http.Header)}, nil
	})).RoundTrip(req)
	if err != nil {
		t.Fatalf("round trip: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestMiddlewareRejectsExpiredToken(t *testing.T) {
	mw := newTestMiddleware(t, &jwtv1.JWT{SigningKey: "secret"})
	token := signedToken(t, "secret", jwtv5.MapClaims{
		"user_id": "u1",
		"exp":     time.Now().Add(-time.Minute).Unix(),
	})
	req := httptest.NewRequest(http.MethodGet, "http://example.com/user-api/v1/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := mw(gwmiddleware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody, Header: make(http.Header)}, nil
	})).RoundTrip(req)
	if err != nil {
		t.Fatalf("round trip: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestMiddlewareAllowsWhitelistBypass(t *testing.T) {
	mw := newTestMiddleware(t, &jwtv1.JWT{
		SigningKey: "secret",
		Whitelist:  []*jwtv1.WhitelistRule{{Path: "/auth-api/v1/login", Method: http.MethodPost}},
	})
	req := httptest.NewRequest(http.MethodPost, "http://example.com/auth-api/v1/login", nil)
	resp, err := mw(gwmiddleware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusCreated, Body: http.NoBody, Header: make(http.Header)}, nil
	})).RoundTrip(req)
	if err != nil {
		t.Fatalf("round trip: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected whitelist bypass, got %d", resp.StatusCode)
	}
}

func newTestMiddleware(t *testing.T, options *jwtv1.JWT) gwmiddleware.Middleware {
	t.Helper()
	anyOptions, err := anypb.New(options)
	if err != nil {
		t.Fatalf("pack options: %v", err)
	}
	mw, err := Middleware(&configv1.Middleware{Options: anyOptions})
	if err != nil {
		t.Fatalf("new middleware: %v", err)
	}
	return mw
}

func signedToken(t *testing.T, signingKey string, claims jwtv5.MapClaims) string {
	t.Helper()
	token, err := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims).SignedString([]byte(signingKey))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}
