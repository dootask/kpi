package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"dootask-kpi-server/models"
)

// 系统设置响应结构
type SystemSettingsResponse struct {
	AllowRegistration bool `json:"allow_registration"`
}

// 设置更新请求结构
type UpdateSettingsRequest struct {
	AllowRegistration bool `json:"allow_registration"`
}

// 获取系统设置
func GetSystemSettings(c *gin.Context) {
	var settings SystemSettingsResponse
	
	// 获取注册设置
	var allowRegistrationSetting models.SystemSetting
	if err := models.DB.Where("key = ?", "allow_registration").First(&allowRegistrationSetting).Error; err == nil {
		settings.AllowRegistration = allowRegistrationSetting.Value == "true"
	} else {
		// 如果设置不存在，默认为true
		settings.AllowRegistration = true
	}

	c.JSON(http.StatusOK, gin.H{
		"data": settings,
	})
}

// 更新系统设置
func UpdateSystemSettings(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新注册设置
	allowRegistrationValue := strconv.FormatBool(req.AllowRegistration)
	var allowRegistrationSetting models.SystemSetting
	
	// 先查找是否存在
	if err := models.DB.Where("key = ?", "allow_registration").First(&allowRegistrationSetting).Error; err != nil {
		// 如果不存在，创建新的设置
		allowRegistrationSetting = models.SystemSetting{
			Key:   "allow_registration",
			Value: allowRegistrationValue,
			Type:  "boolean",
		}
		if err := models.DB.Create(&allowRegistrationSetting).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建设置失败"})
			return
		}
	} else {
		// 如果存在，更新值
		allowRegistrationSetting.Value = allowRegistrationValue
		if err := models.DB.Save(&allowRegistrationSetting).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新设置失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "设置更新成功",
		"data": SystemSettingsResponse{
			AllowRegistration: req.AllowRegistration,
		},
	})
}

// 获取单个设置项（供其他组件使用）
func GetSetting(key string) (string, error) {
	var setting models.SystemSetting
	if err := models.DB.Where("key = ?", key).First(&setting).Error; err != nil {
		return "", err
	}
	return setting.Value, nil
}

// 设置单个设置项（供其他组件使用）
func SetSetting(key, value, settingType string) error {
	var setting models.SystemSetting
	
	// 先查找是否存在
	if err := models.DB.Where("key = ?", key).First(&setting).Error; err != nil {
		// 如果不存在，创建新的设置
		setting = models.SystemSetting{
			Key:   key,
			Value: value,
			Type:  settingType,
		}
		return models.DB.Create(&setting).Error
	} else {
		// 如果存在，更新值
		setting.Value = value
		if settingType != "" {
			setting.Type = settingType
		}
		return models.DB.Save(&setting).Error
	}
}