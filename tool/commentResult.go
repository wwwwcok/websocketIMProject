package tool

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type H struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data,omitempty"`
	Rows  interface{} `json:"rows,omitempty"`
	Total interface{} `json:"total,omitempty"`
}

func Fail(ctx *gin.Context, r any) {
	ctx.JSON(http.StatusOK, H{
		Code: 1,
		Data: r,
		Msg:  "失败",
	})
}

func Success(ctx *gin.Context, r any) {
	ctx.JSON(http.StatusOK, H{
		Code: 0,
		Data: r,
		Msg:  "成功",
	})
}

func RespList(ctx *gin.Context, c int, r any, t any) {
	ctx.JSON(http.StatusOK, H{
		Code:  c,
		Rows:  r,
		Total: t,
	})
}

func md5Encode(data string) string {
	md5 := md5.New()
	md5.Write([]byte(data))
	digest := md5.Sum(nil)
	return hex.EncodeToString(digest)
}

func MD5Encode(data string) string {
	return strings.ToUpper(md5Encode(data))
}

func ValidatePasswd(plainwd, salt, pwd string) bool {
	fmt.Println(MD5Encode(plainwd+salt), plainwd, salt, "-----这是验证密码--------------", pwd)
	return MD5Encode(plainwd+salt) == pwd
}
