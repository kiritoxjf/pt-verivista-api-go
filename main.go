package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"verivista/pt/config"
	"verivista/pt/database"
	"verivista/pt/handlers"
	"verivista/pt/logger"
	"verivista/pt/mail"
	"verivista/pt/modules"
)

func main() {
	err := logger.SetLogOutPut()
	if err != nil {
		logrus.Errorln(err)
		return
	}
	logrus.Infoln("设置日志程序成功")

	err = config.GetConfig()
	if err != nil {
		logrus.Errorln(err)
		return
	}
	logrus.Infoln("获取配置文件成功")

	err = database.InitDB(config.Config.DB)
	if err != nil {
		logrus.Errorln(err)
		return
	}
	logrus.Infoln("数据库初始化成功")

	mail.ConnMailClient()

	// 创建一个Gin路由器
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 公用接口
	commonGroup := r.Group("/api/com/")
	commonGroup.POST("/record", handlers.IpRecordHandler)
	commonGroup.GET("/icp", handlers.ICPHandler)
	commonGroup.GET("/search", handlers.SearchHandler)
	commonGroup.GET("/defense", modules.GetDefense)
	commonGroup.POST("/signCode", handlers.SignCodeHandler)
	commonGroup.POST("/signOn", handlers.SignOnHandler)
	commonGroup.POST("/signIn", handlers.SignInHandler)

	// 私密接口
	authGroup := r.Group("/api/auth/")
	authGroup.Use(modules.AuthMiddleware())
	authGroup.GET("/signOut", handlers.SignOutHandler)
	authGroup.GET("/authInfo", handlers.AuthInfoHandler)
	authGroup.POST("/report", handlers.ReportHandler)

	go func() {
		if err := r.Run(":8081"); err != nil {
			logrus.Errorln("[Gin运行失败]: ", err)
		}
	}()

	logrus.Infoln("Gin程序运行成功，端口8081")
	select {}
}
