package runtime

import (
	"strings"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func ParseAuthorizationToken(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parts := strings.SplitN(value, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func ParseHS256Token(tokenString, signingKey string) (jwtv5.MapClaims, error) {
	claims := jwtv5.MapClaims{}
	_, err := jwtv5.ParseWithClaims(tokenString, claims, func(token *jwtv5.Token) (any, error) {
		return []byte(signingKey), nil
	}, jwtv5.WithValidMethods([]string{jwtv5.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	return claims, nil
}
