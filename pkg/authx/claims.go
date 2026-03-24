package authx

import (
	"context"

	kratosjwt "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

const (
	ClaimUserID    = "user_id"
	ClaimAudience  = "aud"
	ClaimSessionID = "session_id"

	AudienceLogin   = "login"
	AudienceRefresh = "refresh"
)

func ClaimsFromContext(ctx context.Context) (jwtv5.MapClaims, bool) {
	claims, ok := kratosjwt.FromContext(ctx)
	if !ok {
		return nil, false
	}

	switch c := claims.(type) {
	case *jwtv5.MapClaims:
		if c == nil {
			return nil, false
		}
		return *c, true
	case jwtv5.MapClaims:
		return c, true
	default:
		return nil, false
	}
}

func UserID(ctx context.Context) string {
	return claimString(ctx, ClaimUserID)
}

func Audience(ctx context.Context) string {
	return claimString(ctx, ClaimAudience)
}

func SessionID(ctx context.Context) string {
	return claimString(ctx, ClaimSessionID)
}

func claimString(ctx context.Context, key string) string {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return ""
	}
	value, _ := claims[key].(string)
	return value
}
