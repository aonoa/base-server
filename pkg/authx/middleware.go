package authx

import (
	"context"
	"strings"

	kratoserrors "github.com/go-kratos/kratos/v2/errors"
	kratosmiddleware "github.com/go-kratos/kratos/v2/middleware"
	kratosjwt "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func NewWhitelistMatcher(whitelist []string) selector.MatchFunc {
	allowed := make(map[string]struct{}, len(whitelist))
	allowedMethods := make(map[string]struct{}, len(whitelist))
	for _, operation := range whitelist {
		operation = strings.TrimSpace(operation)
		if operation == "" {
			continue
		}
		allowed[operation] = struct{}{}
		allowed[strings.TrimPrefix(operation, "/")] = struct{}{}
		if method := operationMethod(operation); method != "" {
			allowedMethods[method] = struct{}{}
		}
	}
	return func(ctx context.Context, operation string) bool {
		operation = strings.TrimSpace(operation)
		_, ok := allowed[operation]
		if !ok {
			_, ok = allowed[strings.TrimPrefix(operation, "/")]
		}
		if !ok {
			_, ok = allowedMethods[operationMethod(operation)]
		}
		return !ok
	}
}

func operationMethod(operation string) string {
	operation = strings.Trim(strings.TrimSpace(operation), "/")
	if operation == "" {
		return ""
	}
	parts := strings.Split(operation, "/")
	return strings.TrimSpace(parts[len(parts)-1])
}

func NewServerMiddleware(signingKey string, whitelist []string) kratosmiddleware.Middleware {
	jwtMiddleware := kratosjwt.Server(
		func(token *jwtv5.Token) (interface{}, error) {
			return []byte(signingKey), nil
		},
		kratosjwt.WithSigningMethod(jwtv5.SigningMethodHS256),
		kratosjwt.WithClaims(func() jwtv5.Claims {
			return &jwtv5.MapClaims{}
		}),
	)

	return selector.Server(
		jwtMiddleware,
		validateTokenAudience(),
	).Match(NewWhitelistMatcher(whitelist)).Build()
}

func validateTokenAudience() kratosmiddleware.Middleware {
	return func(handler kratosmiddleware.Handler) kratosmiddleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			aud := Audience(ctx)
			action := strings.TrimSpace(Action(ctx))
			operation := ""
			if tr, ok := transport.FromServerContext(ctx); ok {
				operation = tr.Operation()
			}

			if aud == AudienceRefresh {
				if action != ActionRefreshToken || operation != "/api.auth.service.v1.AuthService/RefreshToken" {
					return nil, kratoserrors.Unauthorized("UNAUTHORIZED", "refresh token can only be used for refresh")
				}
			}

			if operation == "/api.auth.service.v1.AuthService/RefreshToken" {
				if action != ActionRefreshToken {
					return nil, kratoserrors.Unauthorized("UNAUTHORIZED", "missing refresh token action")
				}
				if aud != AudienceRefresh {
					return nil, kratoserrors.Unauthorized("UNAUTHORIZED", "invalid refresh token")
				}
			}

			return handler(ctx, req)
		}
	}
}
