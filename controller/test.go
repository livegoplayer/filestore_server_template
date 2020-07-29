package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"
	ginHelper "github.com/livegoplayer/go_gin_helper"

	"github.com/livegoplayer/go_user_rpc/user"
	userpb "github.com/livegoplayer/go_user_rpc/user/grpc"
)

func TestHandler(c *gin.Context) {
	//test
	fmt.Printf("test")
	userClient := user.GetUserClient()
	res, err := userClient.AddUser(c, &userpb.AddUserRequest{
		UserName: "123",
		Password: "123456",
	})

	ginHelper.CheckError(err, "新建用户失败")

	data := res.GetData()
	fmt.Printf(string(data.GetUid()))

	ginHelper.SuccessResp("ok", data)
}
