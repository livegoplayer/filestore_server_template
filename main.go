package main

import (
	"fmt"

	"github.com/gin-gonic/gin"

	. "github.com/livegoplayer/go_helper"
	. "github.com/livegoplsyer/filestore-server/controller"
)

func main() {
	// 初始化一个http服务对象
	//默认有写入控制台的日志
	r := gin.Default()
	// 把这两个处理器替换
	r.NoMethod(HandleNotFound)
	r.NoRoute(HandleNotFound)
	//增加一个recover在 中间件的执行链的最内层，不破坏原来Recover handler的结构，在最内层渲染并且返回api请求结果
	r.Use(ErrHandler())

	// 设置一个get请求的路由，url为/ping, 处理函数（或者叫控制器函数）是一个闭包函数。
	r.POST("/api/file/upload", UpLoadHandler)
	r.GET("/api/file/test", TestHandler)

	err := r.Run(":9090") // 监听并在 9090 上启动服务
	if err != nil {
		fmt.Printf("server start error : " + err.Error())
		return
	}

	fmt.Printf("server is running !")
}
