package handlers

import (
	"net/http"
	"strconv"

	"dootask-kpi-server/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserEventRequest 用于接收来自 DooTask 的用户事件
type UserEventRequest struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"user_name"`
	Role   string `json:"user_role"`
}

// SystemUserOnboard 处理系统用户入职事件
// 规则：
// - 如果用户存在且已离职(IsActive=false)，则改为在职(IsActive=true)
// - 如果用户存在且已在职，则不变
// - 如果用户不存在，则忽略（不创建）
func SystemUserOnboard(c *gin.Context) {
	var req UserEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
		return
	}

	var user models.Employee
	var err error

	// 优先使用 DooTaskUserID
	if req.UserID != 0 {
		err = models.DB.Where("doo_task_user_id = ?", req.UserID).First(&user).Error
	}

	// 如果按 DooTaskUserID 未找到，再尝试按邮箱查找
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed", "message": err.Error()})
			return
		}
		if req.Email != "" {
			err = models.DB.Where("email = ?", req.Email).First(&user).Error
			if err != nil && err != gorm.ErrRecordNotFound {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed", "message": err.Error()})
				return
			}
		}
	}

	// 用户不存在，直接忽略
	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusOK, gin.H{"message": "user not found, skip"})
		return
	}

	// 如果用户存在且 DooTaskUserID 为空，而请求中带了，则补充映射
	if user.DooTaskUserID == nil && req.UserID != 0 {
		id := req.UserID
		user.DooTaskUserID = &id
	}

	// 如果用户已在职，保持不变
	if user.IsActive {
		c.JSON(http.StatusOK, gin.H{
			"message": "user already active",
			"user_id": user.ID,
		})
		return
	}

	// 将离职用户改为在职
	update := map[string]interface{}{
		"is_active": true,
	}
	if user.DooTaskUserID != nil {
		update["doo_task_user_id"] = user.DooTaskUserID
	}

	if err := models.DB.Model(&user).Updates(update).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user reactivated",
		"user_id": user.ID,
	})
}

// SystemUserOffboard 处理系统用户离职事件
// 规则：
// - 如果用户存在且在职(IsActive=true)，则改为离职(IsActive=false)
// - 如果用户存在且已离职，则不变
// - 如果用户不存在，则忽略
func SystemUserOffboard(c *gin.Context) {
	var req UserEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
		return
	}

	var user models.Employee
	var err error

	// 优先使用 DooTaskUserID
	if req.UserID != 0 {
		err = models.DB.Where("doo_task_user_id = ?", req.UserID).First(&user).Error
	}

	// 如果按 DooTaskUserID 未找到，再尝试按邮箱查找
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed", "message": err.Error()})
			return
		}
		if req.Email != "" {
			err = models.DB.Where("email = ?", req.Email).First(&user).Error
			if err != nil && err != gorm.ErrRecordNotFound {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed", "message": err.Error()})
				return
			}
		}
	}

	// 用户不存在，直接忽略
	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusOK, gin.H{"message": "user not found, skip"})
		return
	}

	// 如果用户存在且 DooTaskUserID 为空，而请求中带了，则补充映射
	if user.DooTaskUserID == nil && req.UserID != 0 {
		id := req.UserID
		user.DooTaskUserID = &id
	}

	// 如果用户已离职，保持不变
	if !user.IsActive {
		c.JSON(http.StatusOK, gin.H{
			"message": "user already offboarded",
			"user_id": user.ID,
		})
		return
	}

	// 将在职用户改为离职
	update := map[string]interface{}{
		"is_active": false,
	}
	if user.DooTaskUserID != nil {
		update["doo_task_user_id"] = user.DooTaskUserID
	}

	if err := models.DB.Model(&user).Updates(update).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user offboarded",
		"user_id": user.ID,
	})
}

// parseUint 尝试从字符串解析 uint，空字符串返回 0
func parseUint(s string) uint {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return uint(v)
}

