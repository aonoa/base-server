package jwt

import (
	"errors"
	"net/http"

	configv1 "github.com/go-kratos/gateway/api/gateway/config/v1"
	gwmiddleware "github.com/go-kratos/gateway/middleware"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	jwtv1 "base-server/app/gateway/service/internal/middleware/jwt/v1"
	runtimeutil "base-server/app/gateway/service/internal/middleware/runtime"
)

const refreshPath = "/auth-api/v1/refresh"

func init() {
	gwmiddleware.Register("jwt", Middleware)
}

func Middleware(cfg *configv1.Middleware) (gwmiddleware.Middleware, error) {
	options := &jwtv1.JWT{}
	if cfg.Options != nil {
		if err := anypb.UnmarshalTo(cfg.Options, options, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, err
		}
	}
	if options.SigningKey == "" {
		return nil, errors.New("jwt.signing_key is required")
	}
	whitelist := make([]runtimeutil.RouteRule, 0, len(options.Whitelist))
	for _, item := range options.Whitelist {
		whitelist = append(whitelist, runtimeutil.NewRouteRule(item.Path, item.Method, item.Host))
	}
	return func(next http.RoundTripper) http.RoundTripper {
		return gwmiddleware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			for _, rule := range whitelist {
				if runtimeutil.RouteMatches(rule, req) {
					return next.RoundTrip(req)
				}
			}
			tokenString := runtimeutil.ParseAuthorizationToken(req.Header.Get("Authorization"))
			if tokenString == "" {
				resp := runtimeutil.EmptyResponse(http.StatusUnauthorized)
				resp.Header.Set("WWW-Authenticate", "Bearer")
				return resp, nil
			}
			claims, err := runtimeutil.ParseHS256Token(tokenString, options.SigningKey)
			if err != nil {
				resp := runtimeutil.EmptyResponse(http.StatusUnauthorized)
				resp.Header.Set("WWW-Authenticate", "Bearer error=\"invalid_token\"")
				return resp, nil
			}
			aud, _ := claims["aud"].(string)
			if aud == "refresh" && req.URL.Path != refreshPath {
				return runtimeutil.EmptyResponse(http.StatusUnauthorized), nil
			}
			return next.RoundTrip(req)
		})
	}, nil
}
