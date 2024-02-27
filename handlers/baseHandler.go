package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
	"verivista/pt/database"
)

// GetStatistic 获取统计数据
func GetStatistic(c *gin.Context) {
	DB := database.DBClient
	realIp := c.Request.Header.Get("X-Real-IP")

	// 更新当前IP在线时间
	if _, err := DB.Exec("UPDATE t_ip SET time = ? WHERE ip = ?", time.Now(), realIp); err != nil {
		logrus.Errorln("[更新在线时间失败]:", err)
	}

	// 查询5分钟内在线人数
	var online int
	if err := DB.QueryRow("SELECT COUNT(*) FROM t_ip WHERE time >= ?", time.Now().Add(-5*time.Minute)).Scan(&online); err != nil {
		logrus.Errorln("[获取在线人数失败]:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "获取在线人数失败",
		})
		return
	}

	// 查询被挂人数
	var report int
	if err := DB.QueryRow("SELECT COUNT(DISTINCT email) FROM t_blacklist ").Scan(&report); err != nil {
		logrus.Errorln("[查询被挂人数失败]:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询被挂人数失败",
		})
		return
	}

	// 查询注册人数
	var sign int
	if err := DB.QueryRow("SELECT COUNT(*) FROM t_user").Scan(&sign); err != nil {
		logrus.Errorln("[查询注册人数失败]:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询注册人数失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"online": online,
		"report": report,
		"sign":   sign,
	})
}
