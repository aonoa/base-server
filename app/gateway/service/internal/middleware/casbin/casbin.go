package casbin

import (
	"context"
	"errors"
	"net/http"
	"strings"

	adminv1 "base-server/api/gen/go/admin/service/v1"
	authv1 "base-server/api/gen/go/auth/service/v1"
	"base-server/pkg/authx"
	configv1 "github.com/go-kratos/gateway/api/gateway/config/v1"
	gwmiddleware "github.com/go-kratos/gateway/middleware"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"

	jwtv1 "base-server/app/gateway/service/internal/middleware/casbin/v1"
	runtimeutil "base-server/app/gateway/service/internal/middleware/runtime"
)

var authClientProvider func() authv1.AuthServiceClient
var adminClientProvider func() adminv1.AdminServiceClient
var registryResolver = newServiceRegistryResolver(func(ctx context.Context) ([]*adminv1.ServiceRegistryItem, error) {
	if adminClientProvider == nil || adminClientProvider() == nil {
		return nil, errors.New("admin client is not configured")
	}
	res, err := adminClientProvider().GetServiceRegistryList(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	return res.Items, nil
})
var catalogResolver = newAPICatalogResolver(func(ctx context.Context) ([]*adminv1.ApiListItem, error) {
	if adminClientProvider == nil || adminClientProvider() == nil {
		return nil, errors.New("admin client is not configured")
	}
	res, err := adminClientProvider().GetApiList(ctx, &adminv1.GetApiPageParams{})
	if err != nil {
		return nil, err
	}
	return res.Items, nil
})

func init() {
	gwmiddleware.Register("casbin", Middleware)
}

func SetAuthClient(provider func() authv1.AuthServiceClient) {
	authClientProvider = provider
}

func SetAdminClient(provider func() adminv1.AdminServiceClient) {
	adminClientProvider = provider
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
			ownership := resolveAPIOwnership(req.URL.Path, req.Method)
			res, err := authClientProvider().CheckAuthorization(ctx, &authv1.CheckAuthorizationRequest{
				UserId:        userID,
				Path:          req.URL.Path,
				Method:        req.Method,
				Service:       ownership.ServiceCode,
				ScopeId:       strings.TrimSpace(req.Header.Get(authx.HeaderScopeID)),
				ResourceGroup: ownership.ResourceGroup,
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

func resolveAPIOwnership(path, method string) apiOwnership {
	serviceCode := resolveServiceFromPath(path)
	if ownership, err := catalogResolver.Resolve(context.Background(), path, method); err == nil {
		if ownership.ResourceGroup != "" {
			ownership.ServiceCode = serviceCode
			return ownership
		}
	}
	return apiOwnership{ServiceCode: serviceCode}
}

func resolveServiceFromPath(path string) string {
	if code, err := registryResolver.Resolve(context.Background(), path); err == nil && code != "" {
		return code
	}
	return serviceFromPathFallback(path)
}

func serviceFromPathFallback(path string) string {
	switch {
	case strings.HasPrefix(path, "/auth-api/"):
		return "auth"
	case strings.HasPrefix(path, "/user-api/"):
		return "user"
	case strings.HasPrefix(path, "/admin-api/"):
		return "admin"
	case strings.HasPrefix(path, "/common-api/"):
		return "common"
	default:
		return ""
	}
}
