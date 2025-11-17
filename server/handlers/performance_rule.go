package handlers

import (
	"errors"
	"fmt"
	"math"
	"net/http"

	"dootask-kpi-server/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const weightTolerance = 0.0001

// PerformanceRulePayload 绩效规则请求结构
type PerformanceRulePayload struct {
	NoInvitation   models.PerformanceRuleNoInvitation `json:"no_invitation" binding:"required"`
	WithInvitation models.PerformanceRuleWithInvite   `json:"with_invitation" binding:"required"`
	Enabled        bool                               `json:"enabled"`
}

// GetPerformanceRule 获取绩效评分规则
func GetPerformanceRule(c *gin.Context) {
	var rule models.PerformanceRule
	result := models.DB.First(&rule)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		rule = models.DefaultPerformanceRule()
		if err := models.DB.Create(&rule).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "初始化绩效规则失败",
				"message": err.Error(),
			})
			return
		}
	} else if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取绩效规则失败",
			"message": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": rule,
	})
}

// UpdatePerformanceRule 更新绩效评分规则
func UpdatePerformanceRule(c *gin.Context) {
	var payload PerformanceRulePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	if err := validatePerformanceRulePayload(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var rule models.PerformanceRule
	result := models.DB.First(&rule)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		rule = models.DefaultPerformanceRule()
	} else if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取绩效规则失败",
			"message": result.Error.Error(),
		})
		return
	}

	rule.NoInvitation = payload.NoInvitation
	rule.WithInvitation = payload.WithInvitation
	rule.Enabled = payload.Enabled

	if err := models.DB.Save(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新绩效规则失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "绩效规则更新成功",
		"data":    rule,
	})
}

func validatePerformanceRulePayload(payload PerformanceRulePayload) error {
	checks := []struct {
		label  string
		values []namedValue
	}{
		{
			label: "无邀请评分",
			values: []namedValue{
				{name: "自评", value: payload.NoInvitation.SelfWeight},
				{name: "上级评分", value: payload.NoInvitation.SuperiorWeight},
			},
		},
		{
			label: "有邀请评分-员工",
			values: []namedValue{
				{name: "自评", value: payload.WithInvitation.Employee.SelfWeight},
				{name: "邀请评分（上级）", value: payload.WithInvitation.Employee.InviteSuperiorWeight},
				{name: "上级评分", value: payload.WithInvitation.Employee.SuperiorWeight},
			},
		},
	}

	for _, check := range checks {
		sum := 0.0
		for _, item := range check.values {
			if err := validatePercentageValue(fmt.Sprintf("%s - %s", check.label, item.name), item.value); err != nil {
				return err
			}
			sum += item.value
		}

		if math.Abs(sum-100) > weightTolerance {
			return fmt.Errorf("%s 权重之和必须等于100，当前为 %.2f", check.label, sum)
		}
	}

	return nil
}

type namedValue struct {
	name  string
	value float64
}

func validatePercentageValue(field string, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("%s 的值无效", field)
	}
	if value < 0 || value > 100 {
		return fmt.Errorf("%s 必须在0到100之间", field)
	}
	return nil
}
