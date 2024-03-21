package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
	"verivista/pt/database"
	"verivista/pt/mail"
	"verivista/pt/modules"
)

// SearchHandler 查询是否被挂
func SearchHandler(c *gin.Context) {
	// 防御
	realIp := c.Request.Header.Get("X-Real-IP")
	if crawler := modules.CheckCrawler(c, realIp, "search"); !crawler {
		return
	}

	DB := database.DBClient
	type Black struct {
		Total       int       `json:"total"`
		Email       string    `json:"email"`
		Description string    `json:"description"`
		Date        time.Time `json:"date"`
	}
	var black Black
	email := c.Query("email")

	// 查询被挂数量
	if err := DB.QueryRow("SELECT COUNT(*) as total FROM t_blacklist WHERE email = ?", email).Scan(&black.Total); err != nil {
		logrus.Errorln("[查询被挂数量失败]：", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查人失败，可联系管理员解决",
		})
		return
	}

	if black.Total == 0 {
		c.JSON(http.StatusOK, gin.H{
			"black":    false,
			"lastTime": time.Now(),
		})
		if err := modules.ResetDefense(realIp, "search"); err != nil {
			logrus.Errorln(err)
		}
		return
	}

	if err := DB.QueryRow(
		"SELECT email, description, DATE(date) from t_blacklist WHERE email = ?",
		email).Scan(&black.Email, &black.Description, &black.Date); err != nil {
		logrus.Errorln("[获取查人结果失败]：", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查人失败，可联系管理员解决",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"black":       true,
		"total":       black.Total,
		"email":       black.Email,
		"description": black.Description,
		"date":        black.Date.Format("2006-01-02"),
		"lastTime":    time.Now(),
	})
	if err := modules.ResetDefense(realIp, "search"); err != nil {
		logrus.Errorln(err)
	}
}

// ReportHandler 挂人
func ReportHandler(c *gin.Context) {
	type JsonData struct {
		Email       string `json:"email"`
		Description string `json:"description"`
	}

	DB := database.DBClient
	userId := c.MustGet("userId").(int)

	var jsonData JsonData
	d, _ := c.GetRawData()
	if err := json.Unmarshal(d, &jsonData); err != nil {
		logrus.Errorln("[参数传递失败]:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "获取必要参数失败",
		})
		return
	}

	// 防御
	realIp := c.Request.Header.Get("X-Real-IP")

	if crawler := modules.CheckCrawler(c, realIp, "report"); !crawler {
		return
	}

	var id int
	if err := DB.QueryRow("SELECT id FROM t_blacklist WHERE email = ? AND id = ?", jsonData.Email, userId).Scan(&id); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logrus.Errorln("[查询是否在榜失败]:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "查询是否在榜失败，请联系管理员",
			})
			return
		}
	} else {
		c.JSON(http.StatusConflict, gin.H{
			"message": "您已经挂过这个人啦~",
		})
		return
	}

	if _, err := DB.Exec("INSERT INTO t_blacklist (email, reporter, description) VALUES (?,?,?)", jsonData.Email, userId, jsonData.Description); err != nil {
		logrus.Errorln("[挂人信息录入失败]：", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "挂人信息录入失败，请联系管理员",
		})
		return
	}

	if err := mail.SendWarnMail(jsonData.Email); err != nil {
		logrus.Errorln("[发送被挂通知邮件失败]:", err)
	}

	if err := modules.ResetDefense(realIp, "report"); err != nil {
		logrus.Errorln(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "挂人成功",
	})
}

// ReportListHandler 查询被挂列表
func ReportListHandler(c *gin.Context) {
	searchEmail := c.Query("email")
	userId := c.MustGet("userId")
	realIp := c.Request.Header.Get("X-Real-IP")

	DB := database.DBClient

	var userEmail string
	_ = DB.QueryRow("SELECT * FROM t_user WHERE id = ?", userId).Scan(&userEmail)

	if searchEmail != userEmail {
		// 防御
		if crawler := modules.CheckCrawler(c, realIp, "report_list"); !crawler {
			return
		}
	}

	type ReportCard struct {
		Id          int       `json:"id"`
		Description string    `json:"description"`
		Date        time.Time `json:"date"`
	}

	query, err := DB.Query("SELECT id, description, date FROM t_blacklist WHERE email = ?", searchEmail)
	if err != nil {
		logrus.Errorln("[查询举报单列表失败]:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询举报单列表失败",
		})
		return
	}
	defer func(query *sql.Rows) {
		_ = query.Close()
	}(query)

	var list []ReportCard

	for query.Next() {
		var row ReportCard
		if err := query.Scan(&row.Id, &row.Description, &row.Date); err != nil {
			logrus.Errorln("[扫描举报单列表失败]:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "扫描举报单列表失败",
			})
			return
		}
		list = append(list, row)
	}

	if searchEmail != userEmail {
		if err := modules.ResetDefense(realIp, "report"); err != nil {
			logrus.Errorln(err)
		}
	}

	if len(list) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"reportList": []interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reportList": list,
	})
}
