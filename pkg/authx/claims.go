package authx

import (
	"context"
	"strings"

	kratosjwt "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/transport"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

const (
	ClaimUserID    = "user_id"
	ClaimAudience  = "aud"
	ClaimSessionID = "session_id"

	AudienceLogin   = "login"
	AudienceRefresh = "refresh"

	HeaderAuthorization  = "Authorization"
	HeaderAction         = "x-action"
	HeaderOrganizationID = "x-organization-id"
	ActionRefreshToken   = "refreshToken"
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

func Action(ctx context.Context) string {
	return strings.TrimSpace(RequestHeader(ctx, HeaderAction))
}

func ForwardAuthorizationContext(ctx context.Context) context.Context {
	authorization := strings.TrimSpace(AuthorizationFromContext(ctx))
	organizationID := strings.TrimSpace(RequestHeader(ctx, HeaderOrganizationID))
	if authorization != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", authorization)
	}
	if organizationID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, HeaderOrganizationID, organizationID)
	}
	return ctx
}

func AuthorizationFromContext(ctx context.Context) string {
	return RequestHeader(ctx, HeaderAuthorization)
}

func RequestHeader(ctx context.Context, key string) string {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return ""
	}
	return tr.RequestHeader().Get(key)
}

func claimString(ctx context.Context, key string) string {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return ""
	}
	value, _ := claims[key].(string)
	return value
}
