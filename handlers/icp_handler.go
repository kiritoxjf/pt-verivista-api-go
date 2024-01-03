package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"verivista/pt/database"
	"verivista/pt/interfaces"
)

func ICPHandler(c *gin.Context) {
	DB := database.DBClient

	type ICP struct {
		License string `json:"license"`
	}
	var icp ICP
	err := DB.QueryRow("SELECT license FROM t_icp").Scan(&icp.License)
	if err != nil {
		logrus.Errorln("[ERROR get icp]: ", err)
		c.JSON(http.StatusInternalServerError, interfaces.ErrorResponse{
			Message: "ERROR GET ICP",
		})
		return
	}
	c.JSON(http.StatusOK, icp)
}
