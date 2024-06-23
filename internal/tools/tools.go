package tools

import (
	"context"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"strconv"
	"strings"
)

func GetUserId(ctx context.Context) string {
	uid := ""
	if claims, ok := jwt.FromContext(ctx); ok {
		uid = (*claims.(*jwtv4.MapClaims))["user_id"].(string)
	}
	return uid
}

func DeptStrSplitToInt(dept string) (int64, error) {
	deptList := strings.Split(dept, "-")
	deptId := deptList[len(deptList)-1]
	return strconv.ParseInt(deptId, 10, 64)
}

func GetPageOffset(pageNum, pageSize int64) int64 {
	return (pageNum - 1) * pageSize
}
