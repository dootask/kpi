package handlers

import (
	"net/http"
	"strconv"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"dootask-kpi-server/models"
	"dootask-kpi-server/utils"
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

// 截止时间规则响应结构
type DeadlineRulesResponse struct {
	StandardDays   utils.DeadlineDays   `json:"standard_days"`
	CompressedDays utils.DeadlineDays   `json:"compressed_days"`
	MinimumDays    utils.DeadlineDays   `json:"minimum_days"`
	TimeThreshold  utils.TimeThreshold  `json:"time_threshold"`
	AutoProcessOverdue bool              `json:"auto_process_overdue"`
}

// 获取截止时间规则
func GetDeadlineRules(c *gin.Context) {
	var rules DeadlineRulesResponse
	
	// 获取标准模式时间
	if standardStr, err := GetSetting("deadline_standard_days"); err == nil {
		if standardDays, err := utils.ParseDeadlineDaysFromJSON(standardStr); err == nil {
			rules.StandardDays = standardDays
		}
	}
	
	// 获取压缩模式时间
	if compressedStr, err := GetSetting("deadline_compressed_days"); err == nil {
		if compressedDays, err := utils.ParseDeadlineDaysFromJSON(compressedStr); err == nil {
			rules.CompressedDays = compressedDays
		}
	}
	
	// 获取最小时间要求
	if minimumStr, err := GetSetting("deadline_minimum_days"); err == nil {
		if minimumDays, err := utils.ParseDeadlineDaysFromJSON(minimumStr); err == nil {
			rules.MinimumDays = minimumDays
		}
	}
	
	// 获取时间阈值
	if thresholdStr, err := GetSetting("deadline_time_threshold"); err == nil {
		if threshold, err := utils.ParseTimeThresholdFromJSON(thresholdStr); err == nil {
			rules.TimeThreshold = threshold
		}
	}
	
	// 获取自动处理超时设置
	if autoProcessStr, err := GetSetting("auto_process_overdue"); err == nil {
		rules.AutoProcessOverdue = autoProcessStr == "true"
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data": rules,
	})
}

// 更新截止时间规则
func UpdateDeadlineRules(c *gin.Context) {
	var req DeadlineRulesResponse
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// 更新标准模式时间
	if standardJson, err := json.Marshal(req.StandardDays); err == nil {
		if err := SetSetting("deadline_standard_days", string(standardJson), "json"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新标准模式时间失败"})
			return
		}
	}
	
	// 更新压缩模式时间
	if compressedJson, err := json.Marshal(req.CompressedDays); err == nil {
		if err := SetSetting("deadline_compressed_days", string(compressedJson), "json"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新压缩模式时间失败"})
			return
		}
	}
	
	// 更新最小时间要求
	if minimumJson, err := json.Marshal(req.MinimumDays); err == nil {
		if err := SetSetting("deadline_minimum_days", string(minimumJson), "json"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新最小时间要求失败"})
			return
		}
	}
	
	// 更新时间阈值
	if thresholdJson, err := json.Marshal(req.TimeThreshold); err == nil {
		if err := SetSetting("deadline_time_threshold", string(thresholdJson), "json"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新时间阈值失败"})
			return
		}
	}
	
	// 更新自动处理超时设置
	autoProcessValue := strconv.FormatBool(req.AutoProcessOverdue)
	if err := SetSetting("auto_process_overdue", autoProcessValue, "boolean"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新自动处理设置失败"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "截止时间规则更新成功",
		"data": req,
	})
}