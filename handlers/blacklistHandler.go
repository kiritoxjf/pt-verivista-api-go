package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
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
		Email       string    `json:"email"`
		Reporter    string    `json:"reporter"`
		Description string    `json:"description"`
		Date        time.Time `json:"date"`
	}
	var black Black
	email := c.Query("email")

	if err := DB.QueryRow("SELECT tb.email, tu.email, description, DATE(date) FROM t_blacklist tb LEFT JOIN pt.t_user tu on tu.id = tb.reporter WHERE tb.email = ?", email).Scan(&black.Email, &black.Reporter, &black.Description, &black.Date); err != nil {
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
	if _, err := c.Request.Cookie("verivista_token"); err != nil {
		black.Reporter = ""
	} else {
		black.Reporter = strings.Split(black.Reporter, "@")[0]
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
	if err := DB.QueryRow("SELECT id FROM t_blacklist WHERE email = ?", jsonData.Email).Scan(&id); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logrus.Errorln("[查询是否在榜失败]:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "查询是否在榜失败，请联系管理员",
			})
			return
		}
	} else {
		c.JSON(http.StatusConflict, gin.H{
			"message": "此邮箱已在榜上",
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
