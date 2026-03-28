package tools

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"sort"
	"strconv"
	"strings"

	"base-server/pkg/authx"

	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

// GetUserId 从jwt中获取userID
func GetUserId(ctx context.Context) string {
	return authx.UserID(ctx)
}

// UserPasswdEncrypt 将用户密码加密
func UserPasswdEncrypt(passwd, salt string) string {
	// 数据库不方便使用明文密码，在此处自定义加密逻辑
	h := sha1.New()
	h.Write([]byte(passwd))
	ciphertext := h.Sum(nil)
	return string(ciphertext)
}

func DeptStrSplitToInt(dept string) (int64, error) {
	deptList := strings.Split(dept, "-")
	deptId := deptList[len(deptList)-1]
	return strconv.ParseInt(deptId, 10, 64)
}

func GetPageOffset(pageNum, pageSize int64) int64 {
	return (pageNum - 1) * pageSize
}

// MD5 计算字符串的 MD5
func MD5(text string) string {
	data := []byte(text)
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

type WalkRouteItem struct {
	URL    string
	Method string
}

func WalkHTTPRoutes(server *kratoshttp.Server) ([]WalkRouteItem, error) {
	if server == nil {
		return nil, errors.New("http server is nil")
	}
	items := make([]WalkRouteItem, 0)
	if err := server.WalkRoute(func(info kratoshttp.RouteInfo) error {
		items = append(items, WalkRouteItem{URL: info.Path, Method: info.Method})
		return nil
	}); err != nil {
		return nil, err
	}
	return items, nil
}

func SortAndUniqueWalkRoutes(items []WalkRouteItem) []WalkRouteItem {
	seen := make(map[string]struct{}, len(items))
	result := make([]WalkRouteItem, 0, len(items))
	for _, item := range items {
		key := item.Method + "\x00" + item.URL
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].URL == result[j].URL {
			return result[i].Method < result[j].Method
		}
		return result[i].URL < result[j].URL
	})
	return result
}
