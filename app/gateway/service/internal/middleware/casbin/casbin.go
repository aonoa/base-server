package casbin

import (
	"context"
	"errors"
	"net/http"

	authv1 "base-server/api/gen/go/auth/service/v1"
	configv1 "github.com/go-kratos/gateway/api/gateway/config/v1"
	gwmiddleware "github.com/go-kratos/gateway/middleware"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	jwtv1 "base-server/app/gateway/service/internal/middleware/casbin/v1"
	runtimeutil "base-server/app/gateway/service/internal/middleware/runtime"
)

var authClientProvider func() authv1.AuthServiceClient

func init() {
	gwmiddleware.Register("casbin", Middleware)
}

func SetAuthClient(provider func() authv1.AuthServiceClient) {
	authClientProvider = provider
}

func Middleware(cfg *configv1.Middleware) (gwmiddleware.Middleware, error) {
	options := &jwtv1.Casbin{}
	if cfg.Options != nil {
		if err := anypb.UnmarshalTo(cfg.Options, options, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, err
		}
	}
	if options.SigningKey == "" {
		return nil, errors.New("casbin.signing_key is required")
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
			userID, _ := claims["user_id"].(string)
			aud, _ := claims["aud"].(string)
			if aud == "refresh" && req.URL.Path != "/auth-api/v1/refresh" {
				return runtimeutil.EmptyResponse(http.StatusUnauthorized), nil
			}
			if userID == "" {
				return runtimeutil.EmptyResponse(http.StatusUnauthorized), nil
			}
			if authClientProvider == nil || authClientProvider() == nil {
				return nil, errors.New("auth client is not configured")
			}
			ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", req.Header.Get("Authorization"))
			res, err := authClientProvider().CheckAuthorization(ctx, &authv1.CheckAuthorizationRequest{
				UserId: userID,
				Path:   req.URL.Path,
				Method: req.Method,
			})
			if err != nil {
				return nil, err
			}
			if !res.Allowed {
				return runtimeutil.EmptyResponse(http.StatusForbidden), nil
			}
			return next.RoundTrip(req)
		})
	}, nil
}
