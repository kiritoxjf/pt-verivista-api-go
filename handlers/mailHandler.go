package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"math/big"
	"net/http"
	"verivista/pt/database"
	"verivista/pt/mail"
	"verivista/pt/modules"
)

// SignCodeHandler 注册验证码
func SignCodeHandler(c *gin.Context) {
	type JsonData struct {
		Email string `json:"email"`
	}

	var jsonData JsonData
	d, _ := c.GetRawData()
	err := json.Unmarshal(d, &jsonData)
	if err != nil {
		logrus.Errorln("[参数传递失败]:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "获取必要参数失败",
		})
		return
	}

	// 校验是否已注册
	DB := database.DBClient
	var id int
	if err := DB.QueryRow("SELECT id FROM t_user WHERE email = ?", jsonData.Email).Scan(&id); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logrus.Errorln("[查询邮箱是否已注册失败]:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "查询邮箱是否已注册失败，请联系管理员排查",
			})
			return
		}
	}
	if id > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"message": "邮箱已注册",
		})
		return
	}

	// 防御
	realIp := c.Request.Header.Get("X-Real-IP")

	if crawler := modules.CheckCrawler(c, realIp, "email"); !crawler {
		return
	}

	// 生成验证码(6位随机数)
	random, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		logrus.Errorln("[生成验证码失败]:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
		})
		return
	}
	code := int(random.Int64()) + 100000

	codeType := "sign"

	// 写入数据库
	_, err = DB.Exec("INSERT INTO t_auth (ip, code, type, email) VALUES (?,?,?,?)", realIp, code, codeType, jsonData.Email)
	if err != nil {
		logrus.Errorln("[写入验证码失败]:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
		})
		return
	}

	// 发送邮件
	if err := mail.SendAuthMail(jsonData.Email, realIp, code); err != nil {
		logrus.Errorln("[发送验证码邮件失败]:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})

	if err := modules.ResetDefense(realIp, "email"); err != nil {
		logrus.Errorln(err)
	}
}
