package tools

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"strconv"
	"strings"

	"base-server/pkg/authx"
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
