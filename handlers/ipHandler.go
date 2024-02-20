package handlers

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"verivista/pt/database"
)

// IpRecordHandler 记录IP
func IpRecordHandler(c *gin.Context) {
	DB := database.DBClient
	realIp := c.Request.Header.Get("X-Real-IP")

	// 查询是否已记录
	var id int
	if err := DB.QueryRow("SELECT id FROM t_ip WHERE ip = ?", realIp).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if _, err = DB.Exec("INSERT INTO t_ip (ip) VALUES (?)", realIp); err != nil {
				logrus.Errorln("[记录IP失败]：", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "记录IP失败",
				})
				return
			}
		}
		logrus.Errorln("[查询IP记录失败]：", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询IP记录失败",
		})
		return
	}
	if _, err := DB.Exec("UPDATE t_ip SET time = now() WHERE id = ?", id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "更新IP访问时间失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
