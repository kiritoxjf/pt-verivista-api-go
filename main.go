package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"verivista/pt/config"
	"verivista/pt/database"
	"verivista/pt/handlers"
	"verivista/pt/logger"
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

	// 创建一个Gin路由器
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/api/icp", handlers.ICPHandler)
	r.GET("/api/black", handlers.CheckBlackHandler)
	r.GET("/api/lastTime", modules.GetLastTime)

	go func() {
		if err := r.Run(":8081"); err != nil {
			logrus.Errorln("[Gin运行失败]: ", err)
		}
	}()

	logrus.Infoln("Gin程序运行成功，端口8081")
	select {}
}
