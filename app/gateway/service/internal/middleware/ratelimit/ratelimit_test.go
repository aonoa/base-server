package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/go-kratos/gateway/api/gateway/config/v1"
	gwmiddleware "github.com/go-kratos/gateway/middleware"
	"google.golang.org/protobuf/types/known/anypb"

	ratelimitv1 "base-server/app/gateway/service/internal/middleware/ratelimit/v1"
)

func TestMiddlewareLimitsGlobalRequests(t *testing.T) {
	mw := newTestMiddleware(t, &ratelimitv1.RateLimit{
		RequestsPerSecond: 1,
		Burst:             1,
		Scope:             ratelimitv1.Scope_SCOPE_GLOBAL,
	})
	next := gwmiddleware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody, Header: make(http.Header)}, nil
	})
	firstResp, err := mw.Process(next).RoundTrip(httptest.NewRequest(http.MethodGet, "http://example.com/admin-api/v1/users", nil))
	if err != nil {
		t.Fatalf("first round trip: %v", err)
	}
	if firstResp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected first request to pass, got %d", firstResp.StatusCode)
	}
	secondResp, err := mw.Process(next).RoundTrip(httptest.NewRequest(http.MethodGet, "http://example.com/admin-api/v1/users", nil))
	if err != nil {
		t.Fatalf("second round trip: %v", err)
	}
	if secondResp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected second request to be limited, got %d", secondResp.StatusCode)
	}
}

func TestMiddlewareLimitsPerIPIndependently(t *testing.T) {
	mw := newTestMiddleware(t, &ratelimitv1.RateLimit{
		RequestsPerSecond: 1,
		Burst:             1,
		Scope:             ratelimitv1.Scope_SCOPE_IP,
	})
	next := gwmiddleware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody, Header: make(http.Header)}, nil
	})
	req1 := httptest.NewRequest(http.MethodGet, "http://example.com/admin-api/v1/users", nil)
	req1.RemoteAddr = "10.0.0.1:1234"
	req2 := httptest.NewRequest(http.MethodGet, "http://example.com/admin-api/v1/users", nil)
	req2.RemoteAddr = "10.0.0.2:1234"
	resp1, err := mw.Process(next).RoundTrip(req1)
	if err != nil {
		t.Fatalf("first ip round trip: %v", err)
	}
	resp2, err := mw.Process(next).RoundTrip(req2)
	if err != nil {
		t.Fatalf("second ip round trip: %v", err)
	}
	if resp1.StatusCode != http.StatusNoContent || resp2.StatusCode != http.StatusNoContent {
		t.Fatalf("expected both IPs to pass once, got %d and %d", resp1.StatusCode, resp2.StatusCode)
	}
}

func newTestMiddleware(t *testing.T, options *ratelimitv1.RateLimit) gwmiddleware.MiddlewareV2 {
	t.Helper()
	anyOptions, err := anypb.New(options)
	if err != nil {
		t.Fatalf("pack options: %v", err)
	}
	mw, err := New(&configv1.Middleware{Options: anyOptions})
	if err != nil {
		t.Fatalf("new middleware: %v", err)
	}
	return mw
}
