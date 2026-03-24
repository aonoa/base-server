package whitelist

import (
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/go-kratos/gateway/api/gateway/config/v1"
	gwmiddleware "github.com/go-kratos/gateway/middleware"
	"google.golang.org/protobuf/types/known/anypb"

	whitelistv1 "base-server/app/gateway/service/internal/middleware/whitelist/v1"
)

func TestMiddlewareAllowsConfiguredRoute(t *testing.T) {
	mw := newTestMiddleware(t, &whitelistv1.Whitelist{
		Rules: []*whitelistv1.Rule{{Path: "/admin-api/v1/*", Method: http.MethodGet}},
	})
	resp, err := mw(gwmiddleware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusAccepted, Body: http.NoBody, Header: make(http.Header)}, nil
	})).RoundTrip(httptest.NewRequest(http.MethodGet, "http://example.com/admin-api/v1/users", nil))
	if err != nil {
		t.Fatalf("round trip: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected allowed response, got %d", resp.StatusCode)
	}
}

func TestMiddlewareRejectsUnconfiguredRoute(t *testing.T) {
	mw := newTestMiddleware(t, &whitelistv1.Whitelist{
		Rules: []*whitelistv1.Rule{{Path: "/admin-api/v1/*", Method: http.MethodGet}},
	})
	resp, err := mw(gwmiddleware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusAccepted, Body: http.NoBody, Header: make(http.Header)}, nil
	})).RoundTrip(httptest.NewRequest(http.MethodPost, "http://example.com/admin-api/v1/users", nil))
	if err != nil {
		t.Fatalf("round trip: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

func newTestMiddleware(t *testing.T, options *whitelistv1.Whitelist) gwmiddleware.Middleware {
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
