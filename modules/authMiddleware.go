package modules

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
	"verivista/pt/database"
)

// AuthMiddleware 身份验证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		DB := database.DBClient
		token, err := c.Request.Cookie("verivista_token")
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		var cookieTime time.Time
		var userId int
		err = DB.QueryRow("SELECT time, user FROM t_cookie WHERE token = ? ORDER BY time DESC", token.Value).Scan(&cookieTime, &userId)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if time.Now().Sub(cookieTime) > 120*time.Hour {
			c.SetCookie("verivista_token", "", -1, "/", "", false, false)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("userId", userId)
		c.Next()
	}
}

// CookieCreate Cookie创建
func CookieCreate(email string) (string, error) {
	DB := database.DBClient
	// 查询用户ID
	var userId int
	if err := DB.QueryRow("SELECT id FROM t_user WHERE email = ?", email).Scan(&userId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", sql.ErrNoRows
		}
		return "", err
	}
	// 生成Cookie
	now := time.Now()
	cookie := Sha256Hash(strconv.Itoa(userId) + "_" + email + "_" + now.String())
	if _, err := DB.Exec("INSERT INTO t_cookie (token, time, user) VALUES (?,?,?)", cookie, now, userId); err != nil {
		return "", err
	}
	return cookie, nil
}
