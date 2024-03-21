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
func CheckCrawler(c *gin.Context, ip string, operation string) bool {
	DB := database.DBClient
	var (
		lastTime  time.Time
		frequency int
	)
	now := time.Now()
	fiveMinutes := 1 * time.Minute

	// 第一次请求，初始化反爬记录
	if err := DB.QueryRow("SELECT last_time, frequency FROM t_defence td LEFT JOIN pt.t_ip ti on ti.id = td.ip_id WHERE ip = ? AND type = ?", ip, operation).Scan(&lastTime, &frequency); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if _, err := DB.Exec("INSERT INTO t_defence (ip_id, type) SELECT id, ? FROM t_ip WHERE ip = ?", operation, ip); err != nil {
				logrus.Errorln("[防御数据新增失败]：", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "防御数据新增失败，请联系管理员",
				})
				return false
			}
			return true
		} else {
			logrus.Errorln("[查询防御数据失败]: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "查询防御数据失败，请联系管理员",
			})
			return false
		}
	}

	// 完全触发，拦截IP请求
	if frequency >= 5 {
		c.JSON(http.StatusLocked, gin.H{
			"message": "已完全触发反爬虫，拦截您的请求，请联系管理员申述解锁",
		})
		return false
	}

	// 初步触发，开始警告记录
	if now.Sub(lastTime) < fiveMinutes {
		_, err := DB.Exec("UPDATE t_ip SET frequency = frequency + 1 WHERE ip = ?", ip)
		if err != nil {
			logrus.Errorln("[防御数据更新失败]:", err)
		}
		c.JSON(http.StatusLocked, gin.H{
			"lastTime": lastTime,
			"message":  "触发反爬虫机制，请确保您两次敏感操作间隔超过1分钟",
		})
		return false
	}
	return true
}

// ResetDefense 重置反爬警告次数
func ResetDefense(ip string, operation string) error {
	DB := database.DBClient
	if _, err := DB.Exec("UPDATE t_defence td LEFT JOIN pt.t_ip ti on ti.id = td.ip_id SET last_time = ? WHERE ip = ? AND type = ?", time.Now(), ip, operation); err != nil {
		return fmt.Errorf("[反爬记录重置失败]：%v, %v", ip, err)
	}
	return nil
}

// GetDefense 查询防御信息
func GetDefense(c *gin.Context) {
	DB := database.DBClient
	realIP := c.Request.Header.Get("X-Real-IP")

	type Operation struct {
		Sign       int
		Email      int
		Search     int
		Report     int
		ReportList int
	}

	var operation = Operation{
		Sign:       0,
		Email:      0,
		Search:     0,
		Report:     0,
		ReportList: 0,
	}

	type QueryResult struct {
		Type     string    `json:"type"`
		LastTime time.Time `json:"last_time"`
	}

	query, err := DB.Query("SELECT type, last_time FROM t_defence LEFT JOIN t_ip ON t_defence.ip_id = t_ip.id WHERE ip = ?", realIP)
	if err != nil {
		logrus.Errorln("[查询防御表失败]:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询防御表失败",
		})
		return
	}

	for query.Next() {
		var row QueryResult
		if err := query.Scan(&row.Type, &row.LastTime); err != nil {
			logrus.Errorln("[扫描防御表查询结果失败]:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "扫描防御表查询结果失败",
			})
			return
		}

		switch row.Type {
		case "sign":
			operation.Sign = int(row.LastTime.Unix()) * 1000
		case "email":
			operation.Email = int(row.LastTime.Unix()) * 1000
		case "search":
			operation.Search = int(row.LastTime.Unix()) * 1000
		case "report":
			operation.Report = int(row.LastTime.Unix()) * 1000
		case "report_list":
			operation.ReportList = int(row.LastTime.Unix()) * 1000
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"sign":       operation.Sign,
		"email":      operation.Email,
		"search":     operation.Search,
		"report":     operation.Report,
		"reportList": operation.ReportList,
	})
}
