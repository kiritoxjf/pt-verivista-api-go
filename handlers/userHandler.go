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
		return
	}

	// 防御
	realIp := c.Request.Header.Get("X-Real-IP")

	if crawler := modules.CheckCrawler(c, realIp, "search"); !crawler {
		return
	}

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

	if err := modules.ResetDefense(realIp, "sign"); err != nil {
		logrus.Errorln(err)
	}

	// 生成cookie
	cookie, err := modules.CookieCreate(jsonData.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "帐号注册失败, 请联系管理大大",
			})
		}
		logrus.Errorln("[生成Cookie失败]:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "生成Cookie失败, 请联系管理大大",
		})
		return
	}

	c.SetCookie("verivista_token", cookie, 432000, "/", "", false, false)

	c.JSON(http.StatusOK, gin.H{
		"message": "注册成功",
	})
}

// SignInHandler 登陆
func SignInHandler(c *gin.Context) {
	type JsonData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	realIp := c.Request.Header.Get("X-Real-IP")

	DB := database.DBClient
	var jsonData JsonData
	d, _ := c.GetRawData()
	if err := json.Unmarshal(d, &jsonData); err != nil {
		logrus.Errorln("[参数传递失败]:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "获取必要参数失败",
		})
		return
	}

	// 验证账号密码
	var password string
	if err := DB.QueryRow("SELECT password FROM t_user WHERE email = ?", jsonData.Email).Scan(&password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "这个邮箱没有注册呢~",
			})
			return
		}
		logrus.Errorln("[获取密码失败]：", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "服务器逻辑出错，请联系管理员大大",
		})
		return
	}

	if password != modules.Sha256Hash(jsonData.Password) {
		c.JSON(http.StatusConflict, gin.H{
			"message": "密码错误",
		})
		return
	}

	// 更新登录记录
	if _, err := DB.Exec("UPDATE t_user SET ip = ?, time = ? WHERE email = ?", realIp, time.Now(), jsonData.Email); err != nil {
		logrus.Errorln("[更新登录记录失败]：", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "服务器逻辑出错，请联系管理员大大",
		})
		return
	}

	// 生成cookie
	cookie, err := modules.CookieCreate(jsonData.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "帐号未注册",
			})
		}
		logrus.Errorln("[生成Cookie失败]:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "生成Cookie失败, 请联系管理大大",
		})
		return
	}

	c.SetCookie("verivista_token", cookie, 432000, "/", "", false, false)
	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
	})
}

func SignOutHandler(c *gin.Context) {
	c.SetCookie("verivista_token", "", -1, "/", "", false, false)
	c.JSON(http.StatusOK, gin.H{
		"message": "帐号已退出",
	})
}

// AuthInfoHandler 验证Cookie获取用户信息
func AuthInfoHandler(c *gin.Context) {
	DB := database.DBClient
	userId := c.MustGet("userId").(int)

	type UserInfo struct {
		ID       int    `json:"id"`
		Username string `json:"name"`
		Email    string `json:"email"`
	}
	var userInfo UserInfo
	_ = DB.QueryRow("SELECT id, username, email FROM t_user WHERE id = ?", userId).Scan(&userInfo.ID, &userInfo.Username, &userInfo.Email)

	c.JSON(http.StatusOK, userInfo)
}
