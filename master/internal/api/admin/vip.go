package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jumpfrp/master/internal/service"
	"gorm.io/gorm"
)

func listOrders(db *gorm.DB) gin.HandlerFunc {
	vipSvc := service.NewVIPService(db)
	return func(c *gin.Context) {
		status := c.Query("status")
		orders, err := vipSvc.GetAllOrders(status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": orders})
	}
}

func adminGrantVIP(db *gorm.DB) gin.HandlerFunc {
	vipSvc := service.NewVIPService(db)
	return func(c *gin.Context) {
		var req struct {
			UserID       uint `json:"user_id" binding:"required"`
			VIPLevel     int  `json:"vip_level" binding:"min=0,max=3"`
			DurationDays int  `json:"duration_days" binding:"required,min=1"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		if err := vipSvc.AdminGrantVIP(req.UserID, req.VIPLevel, req.DurationDays); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "VIP 开通成功"})
	}
}
