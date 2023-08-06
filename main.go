package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"websocketIMProject/controller"
	"websocketIMProject/tool"
)

func main() {
	go func() {
		err := http.ListenAndServe("localhost:6060", nil)
		if err != nil {
			panic(err)
		}
	}()
	ginServer := gin.Default()
	ginServer.Use(RemoveMultipleSlashes())
	ginServer.LoadHTMLGlob("view/**/*")
	//静态文件目录
	ginServer.Static("/asset/", "./asset/")
	ginServer.Static("/mnt/", "./mnt/")
	//这两个文件必须单独解析，因为前端传过来的是POST
	ginServer.POST("/asset/plugins/doutu/mkgif/info.json", func(c *gin.Context) {
		c.File("./asset/plugins/doutu/mkgif/info.json")
		//c.JSON(200, gin.H{"mkgif": "ok"})
	})
	ginServer.POST("/asset/plugins/doutu/emoj/info.json", func(c *gin.Context) {
		c.File("./asset/plugins/doutu/emoj/info.json")
		//c.JSON(200, gin.H{"emoj": "ok"})
	})
	go func() {
		fmt.Println("pprof start...")
		fmt.Println(http.ListenAndServe(":9876", nil))
	}()

	//注册路由
	RegisterRoute(ginServer)
	//初始化Xorm，使tool包生成mysql操作实例XormMysqlEngine
	err := tool.InitXormMysqlEngine()
	if err != nil {
		fmt.Println("初始化引擎失败:", err)
	}

	err = tool.InitRedisEngine()
	if err != nil {
		fmt.Println("获取redis客户端:", err)
	}

	ginServer.Run("127.0.0.1:18080")

}

func RegisterRoute(ginServer *gin.Engine) {
	memlogin := controller.MemberController{}
	fmt.Println("name + passwd")
	memlogin.Router(ginServer)
	contactController := controller.ContactController{}
	contactController.Router(ginServer)
	chat := controller.ChatController{}
	chat.Router(ginServer)
	attch := controller.AttchController{}
	attch.Router(ginServer)
}

func RemoveMultipleSlashes() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.Request.URL.Path, "//") {
			url := strings.Replace(c.Request.URL.Path, "//", "/", -1)
			c.Redirect(http.StatusMovedPermanently, url)
			return
		}
		c.Next()
	}
}
