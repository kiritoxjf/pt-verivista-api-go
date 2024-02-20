package handlers

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
	"verivista/pt/database"
	"verivista/pt/modules"
)

// CheckBlackHandler 查询是否被挂
func CheckBlackHandler(c *gin.Context) {
	// 防御
	realIp := c.Request.Header.Get("X-Real-IP")

	if crawler := modules.CheckCrawler(c, realIp, "search"); !crawler {
		return
	}

	DB := database.DBClient
	type Black struct {
		Email       string    `json:"email"`
		Reporter    string    `json:"reporter"`
		Description string    `json:"description"`
		Date        time.Time `json:"date"`
	}
	var black Black
	email := c.Query("email")

	if err := DB.QueryRow("SELECT email, reporter, description, DATE(date) FROM t_blacklist WHERE email = ?", email).Scan(&black.Email, &black.Reporter, &black.Description, &black.Date); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusOK, gin.H{
				"black":    false,
				"lastTime": time.Now(),
			})
			if err := modules.ResetDefense(realIp, "search"); err != nil {
				logrus.Errorln(err)
			}
			return
		} else {
			logrus.Errorln("[获取查人结果失败]：", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "查人失败，可联系管理员解决",
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"black":       true,
		"email":       black.Email,
		"reporter":    black.Reporter,
		"description": black.Description,
		"date":        black.Date.Format("2006-01-02"),
		"lastTime":    time.Now(),
	})
	if err := modules.ResetDefense(realIp, "search"); err != nil {
		logrus.Errorln(err)
	}
}
