package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"websocketIMProject/model"
	"websocketIMProject/service"
	"websocketIMProject/tool"
)

type ContactController struct {
}

func (m *ContactController) Router(Engine *gin.Engine) {
	Engine.POST("/contact/addfriend", m.Addfriend)
	Engine.POST("/contact/loadfriend", m.LoadFriend)
	Engine.POST("/contact/joincommunity", m.JoinCommunity)
	Engine.POST("/contact/loadcommunity", m.LoadCommunity)
	Engine.POST("/contact/createcommunity", m.CreateCommunity)

}

func (m *ContactController) Addfriend(ctx *gin.Context) {
	ContactArgs := model.ContactArg{}
	err := ctx.ShouldBind(&ContactArgs)
	if err != nil {
		tool.Fail(ctx, fmt.Sprintln(err))
		return
	}
	contactSrv := service.ContactService{}
	err = contactSrv.AddFriend(ContactArgs.Userid, ContactArgs.Dstid)
	if err != nil {
		tool.Fail(ctx, fmt.Sprintln(err))
		return
	}

	tool.Success(ctx, "成功添加好友")
	return
}

func (m *ContactController) LoadFriend(ctx *gin.Context) {

	//参数绑定
	ContactArgs := model.ContactArg{}
	err := ctx.ShouldBind(&ContactArgs)
	if err != nil {
		tool.Fail(ctx, err)
		return
	}
	contactSrv := service.ContactService{}
	users, err := contactSrv.SearchFriend(ContactArgs.Userid)
	if err != nil {
		tool.Fail(ctx, err)
		return
	}
	tool.RespList(ctx, 0, users, len(users))
	return
}

func (m *ContactController) LoadCommunity(ctx *gin.Context) {
	//参数绑定
	ContactArgs := model.ContactArg{}
	err := ctx.ShouldBind(&ContactArgs)
	if err != nil {
		tool.Fail(ctx, err)
		return
	}
	contactSrv := service.ContactService{}
	communities, err := contactSrv.SearchComunity(ContactArgs.Userid)
	if err != nil {
		tool.Fail(ctx, err)
		return
	}
	tool.RespList(ctx, 0, communities, len(communities))
	return
}

func (m *ContactController) JoinCommunity(ctx *gin.Context) {
	//参数绑定
	ContactArgs := model.ContactArg{}
	err := ctx.ShouldBind(&ContactArgs)
	if err != nil {
		tool.Fail(ctx, fmt.Sprintln(err))
		return
	}
	contactSrv := service.ContactService{}
	err = contactSrv.JoinCommunity(ContactArgs.Userid, ContactArgs.Dstid)
	if err != nil {
		tool.Fail(ctx, fmt.Sprintln(err))
		return
	}
	//加入新群聊时，将后端维护的clientMap表中的用户node拿出来，再次刷新此用户的groupset
	AddGroupId(ContactArgs.Userid, ContactArgs.Dstid)
	tool.Success(ctx, nil)
	//加入成功后前端要刷新
	return
}

func (m *ContactController) CreateCommunity(ctx *gin.Context) {
	//参数绑定：前端传来的参数为群组创建所需参数
	Community := model.Community{}
	err := ctx.ShouldBind(&Community)
	if err != nil {
		tool.Fail(ctx, fmt.Sprintln(err))
		return
	}

	contactSrv := service.ContactService{}
	comm, err := contactSrv.CreateCommunity(Community)
	if err != nil {
		tool.Fail(ctx, fmt.Sprintln(err))
		return
	}

	tool.Success(ctx, comm)
	return
}
