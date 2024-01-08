package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
	"verivista/pt/database"
	"verivista/pt/interfaces"
)

func ICPHandler(c *gin.Context) {
	//header := c.Request.Header.Get("X-Real-Ip")
	DB := database.DBClient

	type ICP struct {
		License string `json:"license"`
	}
	var icp ICP
	err := DB.QueryRow("SELECT license FROM t_icp").Scan(&icp.License)
	if err != nil {
		logrus.Errorln("[ERROR get icp]: ", err)
		c.JSON(http.StatusInternalServerError, interfaces.ErrorResponse{
			Message: "ERROR GET ICP",
		})
		return
	}
	// 设置浏览器缓存策略 1小时
	c.Header("Cache-Control", "public, max-age=3600")
	c.Header("Last-Modified", time.Now().Format(http.TimeFormat))
	c.JSON(http.StatusOK, icp)
}
