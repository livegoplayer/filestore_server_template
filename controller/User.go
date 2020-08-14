package controller

import (
	"time"

	"github.com/gin-gonic/gin"
	ginHelper "github.com/livegoplayer/go_gin_helper"
	myLogger "github.com/livegoplayer/go_logger"
	"github.com/livegoplayer/go_user_rpc/user"
	userpb "github.com/livegoplayer/go_user_rpc/user/grpc"
	"github.com/spf13/viper"
)

type checkUserStatusRes struct {
	IsLogin     bool                 `json:"isLogin"`
	UserSession *userpb.UserSessions `json:"userSession"`
	Token       string               `json:"token"`
}

//子服务器请求检查是否登录
func CheckTokenHandler(c *gin.Context) {
	//获取token，如果没有就设置
	token, err := c.Cookie("us_user_cookie")
	if token == "" {
		token = c.Request.FormValue("token")

		if token == "" {
			ginHelper.AuthResp("没有权限，请先登录", viper.GetString("user_app_host"))
		}
		//设置一下cookie
		c.SetCookie("us_user_cookie", token, int(time.Hour.Seconds()*6), "/", "", false, false)
	}

	//如果没有token，证明没有登录
	data := &checkUserStatusRes{}
	checkUserStatusRequest := &userpb.CheckUserStatusRequest{}
	checkUserStatusRequest.Token = token

	userClient := user.GetUserClient()
	res, err := userClient.CheckUserStatus(c, checkUserStatusRequest)
	if err != nil {
		if gin.IsDebugging() {
			ginHelper.CheckError(err)
		} else {
			myLogger.Error("获取用户鉴权失败" + err.Error())
			ginHelper.AuthResp("没有权限，请先登录", viper.GetString("user_app_host"))
		}
	}
	//如果没有登录的话
	if res.GetData().IsLogin == false {
		ginHelper.AuthResp("没有权限，请先登录", viper.GetString("user_app_host"))
	}
	data.UserSession = res.GetData().UserSession
	data.IsLogin = res.GetData().IsLogin
	data.Token = res.GetData().Token

	ginHelper.SuccessResp("ok", data)
}

func LogoutHandler(c *gin.Context) {
	//设置本域名下的cookie
	c.SetCookie("us_user_cookie", "", -1, "/", "", false, false)
	logoutRequest := &userpb.LogoutRequest{}
	err := c.Bind(logoutRequest)
	ginHelper.CheckError(err)

	userClient := user.GetUserClient()
	response, err := userClient.Logout(c, logoutRequest)
	ginHelper.CheckError(err)

	data := response.GetData()

	ginHelper.SuccessResp("ok", data)
}

//主服务器请求检查是否登录
func CommonCheckTokenHandler(c *gin.Context) {
	//获取token，如果没有就设置
	token, err := c.Cookie("us_user_cookie")
	if token == "" {
		token = c.Request.FormValue("token")
		if token == "" {
			ginHelper.AuthResp("没有权限，请先登录", viper.GetString("user_app_host"))
		}
		//设置一下cookie
		c.SetCookie("us_user_cookie", token, int(time.Hour.Seconds()*6), "/", "", false, false)
	}

	//如果没有token，证明没有登录
	data := &checkUserStatusRes{}
	checkUserStatusRequest := &userpb.CheckUserStatusRequest{}
	checkUserStatusRequest.Token = token

	userClient := user.GetUserClient()
	res, err := userClient.CheckUserStatus(c, checkUserStatusRequest)
	if err != nil {
		if gin.IsDebugging() {
			ginHelper.CheckError(err)
		} else {
			myLogger.Error("获取用户鉴权失败" + err.Error())
			ginHelper.AuthResp("没有权限，请先登录", viper.GetString("user_app_host"))
		}
	}

	//如果没有登录的话
	if res.GetData().IsLogin == false {
		ginHelper.AuthResp("没有权限，请先登录", viper.GetString("user_app_host"))
	}
	data.UserSession = res.GetData().UserSession
	data.IsLogin = res.GetData().IsLogin
	data.Token = res.GetData().Token
}
