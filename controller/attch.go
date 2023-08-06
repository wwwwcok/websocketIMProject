package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
	"websocketIMProject/tool"
)

func init() {
	os.MkdirAll("./mnt", os.ModePerm)
}

type AttchController struct {
}

func (m *AttchController) Router(Engine *gin.Engine) {
	Engine.POST("/attach/upload", m.Upload)

}

func (m *AttchController) Upload(ctx *gin.Context) {
	src, err := ctx.FormFile("file")
	if err != nil {
		tool.Fail(ctx, fmt.Sprintln(err))
		return
	}
	//获得上传来的文件
	srcFile, err := src.Open()
	if err != nil {
		tool.Fail(ctx, fmt.Sprintln(err))
		return
	}
	//开始拼接新文件的名称，先写文件后缀
	suffix := ".png"
	//可以从完整路径名里提取单个文件名
	srcFn := filepath.Base(src.Filename)
	//或者表单里的文件名包含了后缀
	tmp := strings.Split(srcFn, ".")
	if len(tmp) > 1 {
		suffix = "." + tmp[len(tmp)-1]
	}
	//或者前端指定了文件类型(应该不会用到，好像http包里没有这个字段)
	fileType := ctx.PostForm("filetype")
	//假如有fileType，那么就后缀suffix就设置为fileType
	if len(fileType) > 0 {
		suffix = fileType
	}
	fileName := fmt.Sprintf("%d%04d%s",
		time.Now().Unix(), rand.Int31(), suffix)

	//io操作创建新文件
	dstFile, err := os.Create("./mnt/" + fileName)
	defer dstFile.Close()
	if err != nil {
		tool.Fail(ctx, fmt.Sprintln(err))
		return
	}

	//将文件复制到新文件夹中
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		tool.Fail(ctx, fmt.Sprintln(err))
		return
	}
	//组建新文件的url地址
	url := "/mnt/" + fileName
	//响应前端
	tool.Success(ctx, url)

}
