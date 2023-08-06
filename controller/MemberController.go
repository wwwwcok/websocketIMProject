package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"websocketIMProject/model"
	"websocketIMProject/service"
	"websocketIMProject/tool"
)

type MemberController struct {
}

func (m *MemberController) Router(Engine *gin.Engine) {
	Engine.GET("/", m.index)
	Engine.GET("/user/login.shtml", m.loginPage)
	Engine.GET("/chat/index.shtml", m.indexPage)
	Engine.POST("/user/login", m.login)
	Engine.POST("/user/register", m.register)
}
func (m *MemberController) loginPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "/user/login.shtml", nil)
}
func (m *MemberController) indexPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "/chat/index.shtml", nil)
}
func (m *MemberController) index(ctx *gin.Context) {
	ctx.Redirect(301, "http://127.0.0.1:18080/user/login.shtml")
}

func (m *MemberController) login(ctx *gin.Context) {
	mobile := ctx.PostForm("mobile")
	plainwd := ctx.PostForm("passwd")
	fmt.Println("这是请求时传输的——————————————————密码", plainwd)
	//登录验证密码
	memSrv := service.UserService{}
	tmp, err := memSrv.Login(mobile, plainwd)
	if err != nil {
		tool.Fail(ctx, fmt.Sprintf("%s", err))
		return
	}
	fmt.Println("%n这是-------------------------login---------------%n", mobile+plainwd)
	tool.Success(ctx, tmp)
	return
}
func (m *MemberController) register(ctx *gin.Context) {
	mobile := ctx.PostForm("mobile")
	plainpwd := ctx.PostForm("passwd")
	avatar := ""
	sex := model.SEX_UNKNOW
	nickname := fmt.Sprintf("user%06d", rand.Int31())
	memSrv := service.UserService{}
	tmp, err := memSrv.Register(mobile, plainpwd, avatar, sex, nickname)

	if err != nil {
		fmt.Println(err)
		tool.Fail(ctx, fmt.Sprintf("%s", err))
		return
	}
	fmt.Println("%n这是-------------------------register---------------%n", tmp)

	tool.Success(ctx, tmp)
	return
}
