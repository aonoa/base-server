package authx

import (
	"context"

	kratosmiddleware "github.com/go-kratos/kratos/v2/middleware"
	kratosjwt "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func NewWhitelistMatcher(whitelist []string) selector.MatchFunc {
	allowed := make(map[string]struct{}, len(whitelist))
	for _, operation := range whitelist {
		allowed[operation] = struct{}{}
	}
	return func(ctx context.Context, operation string) bool {
		_, ok := allowed[operation]
		return !ok
	}
}

func NewServerMiddleware(signingKey string, whitelist []string) kratosmiddleware.Middleware {
	return selector.Server(
		kratosjwt.Server(
			func(token *jwtv5.Token) (interface{}, error) {
				return []byte(signingKey), nil
			},
			kratosjwt.WithSigningMethod(jwtv5.SigningMethodHS256),
			kratosjwt.WithClaims(func() jwtv5.Claims {
				return &jwtv5.MapClaims{}
			}),
		),
	).Match(NewWhitelistMatcher(whitelist)).Build()
}
