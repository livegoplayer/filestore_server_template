package controller

import (
	"github.com/gin-gonic/gin"
	ginHelper "github.com/livegoplayer/go_gin_helper"
	"github.com/livegoplayer/go_user_rpc/user"
	userpb "github.com/livegoplayer/go_user_rpc/user/grpc"
)

type checkUserStatusRes struct {
	IsLogin     bool                 `json:"isLogin"`
	UserSession *userpb.UserSessions `json:"userSession"`
	Token       string               `json:"token"`
}

//子服务器请求检查是否登录
func CheckTokenHandler(c *gin.Context) {
	//设置本域名下的cookie
	token, err := c.Cookie("us_user_cookie")
	if token == "" {
		token = c.Request.FormValue("token")
	}
	//检查session是否存在
	//如果没有token，证明没有登录
	data := &checkUserStatusRes{}
	if token == "" {
		ginHelper.SuccessResp("ok", data)
	}

	checkUserStatusRequest := &userpb.CheckUserStatusRequest{}
	checkUserStatusRequest.Token = token

	userClient := user.GetUserClient()
	res, err := userClient.CheckUserStatus(c, checkUserStatusRequest)
	ginHelper.CheckError(err, "检查用户登录状态失败")

	data.UserSession = res.GetData().UserSession
	data.IsLogin = res.GetData().IsLogin
	data.Token = res.GetData().Token

	ginHelper.SuccessResp("ok", data)
}
