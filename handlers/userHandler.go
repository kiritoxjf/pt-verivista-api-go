package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"time"
	"verivista/pt/database"
	"verivista/pt/modules"
)

// SignOnHandler 注册
func SignOnHandler(c *gin.Context) {
	type JsonData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Code     string `json:"code"`
	}

	DB := database.DBClient

	var jsonData JsonData
	d, _ := c.GetRawData()
	err := json.Unmarshal(d, &jsonData)
	if err != nil {
		logrus.Errorln("[参数传递失败]:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "获取必要参数失败",
		})
	}

	// 反爬
	realIp := c.Request.Header.Get("X-Real-IP")

	var codeTime time.Time
	err = DB.QueryRow("SELECT time FROM t_auth WHERE ip = ? AND  email = ? AND code = ?", realIp, jsonData.Email, jsonData.Code).Scan(&codeTime)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logrus.Errorln("[获取验证码记录失败]:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "获取验证码记录失败，请联系管理员排查",
			})
			return
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "没有查询到验证码记录呐",
			})
			return
		}
	}

	if time.Now().Sub(codeTime) > 5*time.Minute {
		c.JSON(http.StatusExpectationFailed, gin.H{
			"message": "验证码已过期",
		})
		return
	}
	// 注册录入
	_, err = DB.Exec("INSERT INTO t_user (username, email, password, ip) VALUES (?,?,?,?)", strings.Split(jsonData.Email, "@")[0], jsonData.Email, modules.Sha256Hash(jsonData.Password), realIp)
	if err != nil {
		logrus.Errorln("[注册信息录入失败]:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "注册信息录入失败，请联系管理员",
		})
		return
	}

	var userId int
	err = DB.QueryRow("SELECT id FROM t_user WHERE email = ?", jsonData.Email).Scan(&userId)
	if err != nil {
		logrus.Errorln("[注册的用户未查询到]:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "注册的用户未查询到，请联系管理员排查",
		})
		return
	}
	now := time.Now()

	// 生成Cookie Token
	cookie := modules.Sha256Hash(strconv.Itoa(userId) + "_" + jsonData.Email + "_" + now.String())
	_, err = DB.Exec("INSERT INTO t_cookie (token, time, user) VALUES (?,?,?)", cookie, time.Now(), userId)
	if err != nil {
		logrus.Errorln("[生成Cookie失败]:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "生成Cookie失败，请联系管理员排查",
		})
		return
	}

	c.SetCookie("verivista_token", cookie, 432000, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "注册成功",
	})
}

// SignInHandler 登陆
func SignInHandler(c *gin.Context) {

}

func UserHandler(c *gin.Context) {
	//DB := database.DBClient
	//type User struct {
	//	Name      string `json:"name"`
	//	Nick      string `json:"nick"`
	//	Avatar    string `json:"avatar"`
	//	Signature string `json:"signature"`
	//}
	//var user User
	//name := c.Query("name")
	//// 获取用户信息
	//err := DB.QueryRow("SELECT name, nick, avatar,signature from t_user where name = ?", name).Scan(&user.Name, &user.Nick, &user.Avatar, &user.Signature)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, interfaces.ErrorResponse{
	//		Message: "ERROR GET USER",
	//	})
	//	return
	//}
	//c.JSON(http.StatusOK, user)
}
