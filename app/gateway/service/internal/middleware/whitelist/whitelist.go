package whitelist

import (
	"errors"
	"net/http"

	configv1 "github.com/go-kratos/gateway/api/gateway/config/v1"
	gwmiddleware "github.com/go-kratos/gateway/middleware"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	runtimeutil "base-server/app/gateway/service/internal/middleware/runtime"
	whitelistv1 "base-server/app/gateway/service/internal/middleware/whitelist/v1"
)

func init() {
	gwmiddleware.Register("whitelist", Middleware)
}

func Middleware(cfg *configv1.Middleware) (gwmiddleware.Middleware, error) {
	options := &whitelistv1.Whitelist{}
	if cfg.Options != nil {
		if err := anypb.UnmarshalTo(cfg.Options, options, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, err
		}
	}
	if len(options.Rules) == 0 {
		return nil, errors.New("whitelist.rules is required")
	}
	rules := make([]runtimeutil.RouteRule, 0, len(options.Rules))
	for _, item := range options.Rules {
		rules = append(rules, runtimeutil.NewRouteRule(item.Path, item.Method, item.Host))
	}
	statusCode := int(options.StatusCode)
	if statusCode <= 0 {
		statusCode = http.StatusForbidden
	}
	return func(next http.RoundTripper) http.RoundTripper {
		return gwmiddleware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			for _, rule := range rules {
				if runtimeutil.RouteMatches(rule, req) {
					return next.RoundTrip(req)
				}
			}
			return runtimeutil.EmptyResponse(statusCode), nil
		})
	}, nil
}
