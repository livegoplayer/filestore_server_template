package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"

	myHelper "github.com/livegoplayer/go_helper"
	user "github.com/livegoplayer/go_user_rpc/user"
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

	myHelper.CheckError(err, "新建用户失败")

	data := res.GetData()
	fmt.Printf(string(data.GetUid()))

	myHelper.SuccessResp(c, "ok", data)
}
