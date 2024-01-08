package modules

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
	"verivista/pt/database"
)

// CheckCrawler 检查反爬记录并更新
func CheckCrawler(ip string) (int, time.Time) {
	DB := database.DBClient
	var (
		lastTime  time.Time
		frequency int
	)
	now := time.Now()
	fiveMinutes := 5 * time.Minute

	// 第一次请求，初始化反爬记录
	if err := DB.QueryRow("SELECT last_time, frequency FROM t_ip WHERE ip = ?", ip).Scan(&lastTime, &frequency); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, err = DB.Exec("INSERT INTO t_ip (ip, last_time) VALUES (?,?)", ip, time.Now())
			if err != nil {
				logrus.Errorln("[警告数据新增失败]：", err)
			}
			return 0, lastTime
		} else {
			logrus.Errorln("[查询反爬记录失败]: ", err)
			return 1, lastTime
		}
	}

	// 完全触发，拦截IP请求
	if frequency >= 5 {
		return 2, lastTime
	}

	// 初步触发，开始警告记录
	if now.Sub(lastTime) < fiveMinutes {
		_, err := DB.Exec("UPDATE t_ip SET frequency = frequency + 1 WHERE ip = ?", ip)
		if err != nil {
			logrus.Errorln("[警告次数更新失败]:", err)
		}
		return 1, lastTime
	}

	return 0, lastTime
}

// ResetCrawler 重置反爬警告次数
func ResetCrawler(ip string) error {
	DB := database.DBClient
	if _, err := DB.Exec("UPDATE t_ip SET frequency = 0, last_time = ? WHERE ip = ?", time.Now(), ip); err != nil {
		return fmt.Errorf("[反爬记录重置失败]：%v, %v", ip, err)
	}
	return nil
}

// GetLastTime 获取上次敏感操作时间
func GetLastTime(c *gin.Context) {
	DB := database.DBClient
	var lastTime time.Time
	if err := DB.QueryRow("SELECT last_time FROM t_ip WHERE ip = ?", c.Request.Header.Get("X-Real-IP")).Scan(&lastTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusOK, nil)
			return
		} else {
			logrus.Errorln("[查询上次敏感操作时间失败]:", err)
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"lastTime": lastTime,
	})
}
