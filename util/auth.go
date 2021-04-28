package util

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"regexp"
	"strings"
	"time"
)

func Sha256(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func Md5(data string) string {
	m := md5.New()
	m.Write([]byte(data))
	return hex.EncodeToString(m.Sum(nil))
}

// 生成一个唯一id
func GenerateDataId() string {
	id := uuid.New().String()
	return id
}

// 密码生成
func GenerateHashPasswd(loginId, rawPasswd string) string {
	d := strings.ToLower(loginId) + rawPasswd
	h := Sha256(d)
	return h
}

// token生成
func GenerateToken(userId string) string {
	d := fmt.Sprintf("%s%d", userId, time.Now().Nanosecond())
	h := Md5(d)
	h = strings.ToUpper(h)
	return h
}

// FuncVerifyMobile 手机号验证
func FuncVerifyMobile(mobile string) bool {
	regular := "^(1[3-9])\\d{9}$"
	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobile)
}

// FuncVerifyMobile 身份证号码验证
func FuncVerifyIdCard(idCard string) bool {
	idCardLen := strings.Count(idCard, "")
	if idCardLen-1 == 18 {
		regular18 := "^[1-9]\\d{5}(18|19|([23]\\d))\\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\\d{3}[0-9Xx]$"
		reg := regexp.MustCompile(regular18)
		return reg.MatchString(idCard)
	} else {
		regular15 := "^[1-9]\\d{5}\\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\\d{2}[0-9Xx]$"
		reg := regexp.MustCompile(regular15)
		return reg.MatchString(idCard)
	}
}
