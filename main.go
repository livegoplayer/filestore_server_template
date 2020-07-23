package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	. "github.com/livegoplayer/filestore-server/controller"

	dbHelper "github.com/livegoplayer/go_db_helper"
	ginHelper "github.com/livegoplayer/go_gin_helper"
	. "github.com/livegoplayer/go_helper"
	myLogger "github.com/livegoplayer/go_logger"
)

func main() {
	// 初始化一个http服务对象
	//默认有写入控制台的日志
	// 把这两个处理器替换
	r := gin.New()
	r.NoMethod(ginHelper.HandleNotFound)
	r.NoRoute(ginHelper.HandleNotFound)

	//加载.env文件
	LoadEnv()

	//设置gin的运行模式
	switch viper.GetString("ENV") {
	case PRODUCTION_ENV:
		gin.SetMode(gin.ReleaseMode)
	case DEVELOPMENT_ENV:
		gin.SetMode(gin.DebugMode)
		//额外放置一个可以在控制台打印access_log的中间件
		r.Use(gin.Logger())
	default:
		gin.SetMode(gin.DebugMode)
		r.Use(gin.Logger())
	}

	r.Use(ginHelper.ErrHandler())

	//gin的格式化参数
	//改造access log, 输出到文件
	r.Use(myLogger.GetGinAccessFileLogger(viper.GetString("log.access_log_file_path"), viper.GetString("log.access_log_file_name")))
	//如果是debug模式的话，使用logger另外打印一份输出到控制台的logger
	if gin.IsDebugging() {
		r.Use(gin.Logger())
		//额外输出错误异常栈
	}

	//app_log
	//如果是debug模式的话，直接打印到控制台
	var appLogger *logrus.Logger
	if gin.IsDebugging() {
		appLogger = myLogger.GetConsoleLogger()
	} else {
		appLogger = myLogger.GetMysqlLogger(viper.GetString("log.app_log_mysql_host"), viper.GetString("log.app_log_mysql_port"), viper.GetString("log.app_log_mysql_db_name"), viper.GetString("log.app_log_mysql_table_name"), viper.GetString("log.app_log_mysql_user"), viper.GetString("log.app_log_mysql_pass"))
	}
	myLogger.SetLogger(appLogger)

	//解决跨域问题的中间件
	r.Use(ginHelper.Cors(viper.GetStringSlice("client_list")))

	dbHelper.InitDbHelper(&dbHelper.MysqlConfig{Username: viper.GetString("database.username"), Password: viper.GetString("database.password"), Host: viper.GetString("database.host"), Port: int32(viper.GetInt("database.port")), Dbname: viper.GetString("database.dbname")}, viper.GetBool("database.log_mode"), viper.GetInt("database.max_open_connection"), viper.GetInt("database.max_idle_connection"))

	//更换校验器
	binding.Validator = ValidatorV10

	// 设置一个get请求的路由，url为/ping, 处理函数（或者叫控制器函数）是一个闭包函数。
	r.POST("/api/file/upload", UpLoadHandler)
	r.GET("/api/file/test", TestHandler)

	r.POST("/api/user/checkToken", CheckTokenHandler)

	//获取文件列表
	r.GET("/api/file/getFileList", GetFileListHandler)
	r.GET("/api/file/getPathList", GetUserPathListHandler)

	err := r.Run(":9090") // 监听并在 9090 上启动服务
	if err != nil {
		fmt.Printf("server start error : " + err.Error())
		return
	}

	fmt.Printf("server is running !")
}
