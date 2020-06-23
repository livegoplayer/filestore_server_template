package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"filestore-server/helper"
	user "filestore-server/rpc"
	userpb "filestore-server/rpc/grpc"
)

func TestHandler(c *gin.Context) {
	//test
	fmt.Printf("test")
	userClient := user.GetUserClient()
	res, err := userClient.AddUser(c, &userpb.AddUserRequest{
		UserName: "123",
		Password: "123456",
	})

	helper.CheckError(err, "新建用户失败")

	data := res.GetData()
	fmt.Printf(string(data.GetUid()))

	helper.SuccessResp(c, "ok", data)
}
