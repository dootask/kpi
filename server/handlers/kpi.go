package handlers

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"dootask-kpi-server/models"
	"dootask-kpi-server/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// KPI项目管理

// 创建KPI项目
func CreateItem(c *gin.Context) {
	var item models.KPIItem

	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	result := models.DB.Create(&item)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建KPI项目失败",
			"message": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "KPI项目创建成功",
		"data":    item,
	})
}

// 获取KPI项目
func GetItem(c *gin.Context) {
	id := c.Param("id")
	itemId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的项目ID",
		})
		return
	}

	var item models.KPIItem
	result := models.DB.Preload("Template").First(&item, itemId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "KPI项目不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": item,
	})
}

// 更新KPI项目
func UpdateItem(c *gin.Context) {
	id := c.Param("id")
	itemId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的项目ID",
		})
		return
	}

	var item models.KPIItem
	result := models.DB.First(&item, itemId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "KPI项目不存在",
		})
		return
	}

	var updateData models.KPIItem
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	result = models.DB.Model(&item).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新KPI项目失败",
			"message": result.Error.Error(),
		})
		return
	}

	if updateData.MaxScore == 0 {
		result = models.DB.Model(&item).Update("max_score", 0)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "更新MaxScore失败",
				"message": result.Error.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "KPI项目更新成功",
		"data":    item,
	})
}

// 删除KPI项目
func DeleteItem(c *gin.Context) {
	id := c.Param("id")
	itemId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的项目ID",
		})
		return
	}

	result := models.DB.Delete(&models.KPIItem{}, itemId)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除KPI项目失败",
			"message": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "KPI项目删除成功",
	})
}

// KPI评估管理

// 获取所有评估
func GetEvaluations(c *gin.Context) {
	var evaluations []models.KPIEvaluation

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := c.Query("status")
	employeeID := c.Query("employee_id")
	departmentID := c.Query("department_id")
	managerID := c.Query("manager_id")
	period := c.Query("period")
	year := c.Query("year")
	month := c.Query("month")
	quarter := c.Query("quarter")

	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询
	query := models.DB.Preload("Employee.Department").Preload("Template").Preload("Scores")

	// 添加筛选条件
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if employeeID != "" {
		query = query.Where("employee_id = ?", employeeID)
	}
	if departmentID != "" {
		query = query.Where("employee_id IN (SELECT id FROM employees WHERE department_id = ?)", departmentID)
	}
	if managerID != "" {
		query = query.Where("employee_id IN (SELECT id FROM employees WHERE manager_id = ?)", managerID)
	}
	if period != "" {
		query = query.Where("period = ?", period)
	}
	if year != "" {
		query = query.Where("year = ?", year)
	}
	if month != "" {
		query = query.Where("month = ?", month)
	}
	if quarter != "" {
		query = query.Where("quarter = ?", quarter)
	}

	// 构建基础统计查询（不含 status 筛选，用于统计卡片）
	// 只统计在职员工的评估
	buildStatsQuery := func() *gorm.DB {
		statsQuery := models.DB.Model(&models.KPIEvaluation{}).
			Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
			Where("employees.is_active = ?", true)
		if employeeID != "" {
			statsQuery = statsQuery.Where("kpi_evaluations.employee_id = ?", employeeID)
		}
		if departmentID != "" {
			statsQuery = statsQuery.Where("employees.department_id = ?", departmentID)
		}
		if managerID != "" {
			statsQuery = statsQuery.Where("employees.manager_id = ?", managerID)
		}
		if period != "" {
			statsQuery = statsQuery.Where("kpi_evaluations.period = ?", period)
		}
		if year != "" {
			statsQuery = statsQuery.Where("kpi_evaluations.year = ?", year)
		}
		if month != "" {
			statsQuery = statsQuery.Where("kpi_evaluations.month = ?", month)
		}
		if quarter != "" {
			statsQuery = statsQuery.Where("kpi_evaluations.quarter = ?", quarter)
		}
		return statsQuery
	}

	// 获取统计数据
	var statsTotal int64
	var statsPending int64
	var statsCompleted int64
	var statsAvgScore float64

	// 总评估数
	if err := buildStatsQuery().Count(&statsTotal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取统计数据失败",
			"message": err.Error(),
		})
		return
	}

	// 待处理数（pending, self_evaluated, manager_evaluated, pending_confirm）
	if err := buildStatsQuery().Where("kpi_evaluations.status IN ?", []string{"pending", "self_evaluated", "manager_evaluated", "pending_confirm"}).Count(&statsPending).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取统计数据失败",
			"message": err.Error(),
		})
		return
	}

	// 已完成数
	if err := buildStatsQuery().Where("kpi_evaluations.status = ?", "completed").Count(&statsCompleted).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取统计数据失败",
			"message": err.Error(),
		})
		return
	}

	// 平均分（只计算有分数的评估）
	var avgResult struct {
		AvgScore float64
	}
	if err := buildStatsQuery().Where("kpi_evaluations.total_score > 0").Select("COALESCE(AVG(kpi_evaluations.total_score), 0) as avg_score").Scan(&avgResult).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取统计数据失败",
			"message": err.Error(),
		})
		return
	}
	statsAvgScore = avgResult.AvgScore

	// 获取总数（用于分页，受 status 筛选影响）
	var total int64
	countQuery := models.DB.Model(&models.KPIEvaluation{})
	if status != "" {
		countQuery = countQuery.Where("status = ?", status)
	}
	if employeeID != "" {
		countQuery = countQuery.Where("employee_id = ?", employeeID)
	}
	if departmentID != "" {
		countQuery = countQuery.Where("employee_id IN (SELECT id FROM employees WHERE department_id = ?)", departmentID)
	}
	if managerID != "" {
		countQuery = countQuery.Where("employee_id IN (SELECT id FROM employees WHERE manager_id = ?)", managerID)
	}
	if period != "" {
		countQuery = countQuery.Where("period = ?", period)
	}
	if year != "" {
		countQuery = countQuery.Where("year = ?", year)
	}
	if month != "" {
		countQuery = countQuery.Where("month = ?", month)
	}
	if quarter != "" {
		countQuery = countQuery.Where("quarter = ?", quarter)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取评估总数失败",
			"message": err.Error(),
		})
		return
	}

	// 分页查询，按创建时间倒序
	offset := (page - 1) * pageSize
	result := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&evaluations)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取评估列表失败",
			"message": result.Error.Error(),
		})
		return
	}

	// 计算分页信息
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, gin.H{
		"data":       evaluations,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
		"hasNext":    page < totalPages,
		"hasPrev":    page > 1,
		"stats": gin.H{
			"total":     statsTotal,
			"pending":   statsPending,
			"completed": statsCompleted,
			"avgScore":  statsAvgScore,
		},
	})
}

// 创建评估
func CreateEvaluation(c *gin.Context) {
	var evaluation models.KPIEvaluation

	if err := c.ShouldBindJSON(&evaluation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	// 开始数据库事务
	tx := models.DB.Begin()

	// 创建评估记录
	result := tx.Create(&evaluation)
	if result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建评估失败",
			"message": result.Error.Error(),
		})
		return
	}

	// 获取模板的所有KPI项目
	var items []models.KPIItem
	tx.Where("template_id = ?", evaluation.TemplateID).Find(&items)

	// 为每个KPI项目创建评分记录
	for _, item := range items {
		score := models.KPIScore{
			EvaluationID: evaluation.ID,
			ItemID:       item.ID,
		}
		if err := tx.Create(&score).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "创建评分记录失败",
				"message": err.Error(),
			})
			return
		}
	}

	tx.Commit()

	// 获取完整的评估信息
	models.DB.Preload("Employee.Department").Preload("Template").Preload("Scores").First(&evaluation, evaluation.ID)

	// 发送 DooTask 机器人通知：创建考核 -> 员工待自评
	if evaluation.Employee.DooTaskUserID != nil {
		dooTaskClient := utils.NewDooTaskClient(c.GetHeader("DooTaskAuth"))
		periodValue := utils.GetPeriodValue(evaluation.Period, evaluation.Year, evaluation.Month, evaluation.Quarter)

		appConfigJSON := utils.BuildKPIAppConfig(evaluation.ID)

		message := fmt.Sprintf(
			"**你有一条新的绩效考核**\n- 考核模板：%s\n- 考核周期：%s\n- 被评估员工：%s\n\n> <div class=\"open-micro-app\" data-app-config='%s'>查看详情：点击查看详情</div>",
			evaluation.Template.Name,
			periodValue,
			evaluation.Employee.Name,
			appConfigJSON,
		)

		_ = dooTaskClient.SendBotMessage(evaluation.Employee.DooTaskUserID, message)
	}

	// 发送实时通知
	operatorID := c.GetUint("user_id")
	GetNotificationService().SendNotification(operatorID, EventEvaluationCreated, &evaluation)

	c.JSON(http.StatusCreated, gin.H{
		"message": "评估创建成功",
		"data":    evaluation,
	})
}

// 获取单个评估
func GetEvaluation(c *gin.Context) {
	id := c.Param("id")
	evaluationId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的评估ID",
		})
		return
	}

	var evaluation models.KPIEvaluation
	result := models.DB.Preload("Employee.Department").Preload("Template").Preload("Scores.Item").First(&evaluation, evaluationId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "评估不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": evaluation,
	})
}

// 更新评估
func UpdateEvaluation(c *gin.Context) {
	id := c.Param("id")
	evaluationId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的评估ID",
		})
		return
	}

	var evaluation models.KPIEvaluation
	result := models.DB.Preload("Employee").First(&evaluation, evaluationId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "评估不存在",
		})
		return
	}

	var updateData models.KPIEvaluation
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	// 特殊处理：如果状态是self_evaluated，检查员工是否有主管
	// 注意：由于在UpdateSelfScore中已经限制了所有员工（包括HR）必须有直属上级才能开始自评
	// 所以这里理论上不应该出现没有主管的情况，但为了数据一致性，仍然进行检查
	if updateData.Status == "self_evaluated" {
		// 检查员工是否有直属主管
		if evaluation.Employee.ManagerID == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "暂无直属上级，请联系HR",
			})
			return
		}
	}

	result = models.DB.Model(&evaluation).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新评估失败",
			"message": result.Error.Error(),
		})
		return
	}

	// 当流程进入等待HR审核阶段时，根据配置自动计算HR评分
	if updateData.Status == "manager_evaluated" {
		// 检查绩效规则是否启用
		var rule models.PerformanceRule
		if err := models.DB.First(&rule).Error; err == nil && rule.Enabled {
			// 如果规则已启用，自动计算HR评分
			if err := applyPerformanceRuleForEvaluation(evaluation.ID); err == nil {
				// 如果绩效规则应用成功，自动将状态推进到pending_confirm
				models.DB.Model(&evaluation).Update("status", "pending_confirm")
				updateData.Status = "pending_confirm"
			}
		}
		// 如果规则未启用或应用失败，保持manager_evaluated状态，等待HR手动审核
	}

	// 如果状态变为completed，自动计算最终得分
	if updateData.Status == "completed" {
		var scores []models.KPIScore
		if err := models.DB.Where("evaluation_id = ?", evaluation.ID).Find(&scores).Error; err == nil {
			// 更新各项目的final_score
			for _, s := range scores {
				var final float64
				if s.HRScore != nil {
					final = *s.HRScore
				} else if s.ManagerScore != nil {
					final = *s.ManagerScore
				} else if s.SelfScore != nil {
					final = *s.SelfScore
				}
				models.DB.Model(&s).Update("final_score", final)
			}

			// 如果前端发送了total_score（说明可能已经过异议处理调整），使用前端发送的值
			// 否则，如果存在异议处理（有final_comment），保持现有的total_score
			// 否则，重新计算total_score
			if updateData.TotalScore > 0 {
				// 使用前端发送的total_score（已通过异议处理调整）
				models.DB.Model(&evaluation).Update("total_score", updateData.TotalScore)
			} else if evaluation.FinalComment != "" {
				// 存在异议处理，保持现有的total_score
				// 不更新total_score
			} else {
				// 没有异议处理，正常计算最终得分
				total := 0.0
				for _, s := range scores {
					var final float64
					if s.HRScore != nil {
						final = *s.HRScore
					} else if s.ManagerScore != nil {
						final = *s.ManagerScore
					} else if s.SelfScore != nil {
						final = *s.SelfScore
					}
					total += final
				}
				models.DB.Model(&evaluation).Update("total_score", total)
			}
		}
	}

	// 重新加载更新后的数据
	models.DB.Preload("Employee.Manager").Preload("Template").First(&evaluation, evaluationId)

	// 发送DooTask机器人通知（根据状态变更）
	if updateData.Status != "" {
		dooTaskClient := utils.NewDooTaskClient(c.GetHeader("DooTaskAuth"))
		periodValue := utils.GetPeriodValue(evaluation.Period, evaluation.Year, evaluation.Month, evaluation.Quarter)

		switch updateData.Status {
		case "self_evaluated":
			// 完成自评：通知主管
			if evaluation.Employee.Manager != nil && evaluation.Employee.Manager.DooTaskUserID != nil {
				appConfigJSON := utils.BuildKPIAppConfig(evaluation.ID)

				message := fmt.Sprintf(
					"**你有一条绩效考核待评估**\n- 考核模板：%s\n- 考核周期：%s\n- 部门员工：%s\n\n> <div class=\"open-micro-app\" data-app-config='%s'>查看详情：点击查看详情</div>",
					evaluation.Template.Name,
					periodValue,
					evaluation.Employee.Name,
					appConfigJSON,
				)
				_ = dooTaskClient.SendBotMessage(evaluation.Employee.Manager.DooTaskUserID, message)
			}

		case "manager_evaluated":
			// 完成主管评分：如仍处于待HR审核阶段，则通知HR
			// （当启用了绩效规则且自动推进到 pending_confirm 时，这里不会进入）
			appConfigJSON := utils.BuildKPIAppConfig(evaluation.ID)

			hrUserIDs := GetNotificationService().GetAllHRUsers()
			for _, hrID := range hrUserIDs {
				var hr models.Employee
				if err := models.DB.First(&hr, hrID).Error; err != nil || hr.DooTaskUserID == nil {
					continue
				}

				message := fmt.Sprintf(
					"**你有一条绩效考核待审核**\n- 考核模板：%s\n- 考核周期：%s\n- 部门员工：%s\n\n> <div class=\"open-micro-app\" data-app-config='%s'>查看详情：点击查看详情</div>",
					evaluation.Template.Name,
					periodValue,
					evaluation.Employee.Name,
					appConfigJSON,
				)

				_ = dooTaskClient.SendBotMessage(hr.DooTaskUserID, message)
			}

		case "pending_confirm":
			// 完成审核（HR人工或规则自动）：通知员工确认
			if evaluation.Employee.DooTaskUserID != nil {
				appConfigJSON := utils.BuildKPIAppConfig(evaluation.ID)

				message := fmt.Sprintf(
					"**你的绩效已审核完成，请确认结果**\n- 考核模板：%s\n- 考核周期：%s\n- 总分：%.1f\n\n> <div class=\"open-micro-app\" data-app-config='%s'>查看详情：点击查看详情</div>",
					evaluation.Template.Name,
					periodValue,
					evaluation.TotalScore,
					appConfigJSON,
				)
				_ = dooTaskClient.SendBotMessage(evaluation.Employee.DooTaskUserID, message)
			}

		case "completed":
			// 员工确认完成：通知 HR
			appConfigJSON := utils.BuildKPIAppConfig(evaluation.ID)

			hrUserIDs := GetNotificationService().GetAllHRUsers()
			for _, hrID := range hrUserIDs {
				var hr models.Employee
				if err := models.DB.First(&hr, hrID).Error; err != nil || hr.DooTaskUserID == nil {
					continue
				}

				message := fmt.Sprintf(
					"**绩效已完成**\n- 考核模板：%s\n- 考核周期：%s\n- 部门员工：%s\n- 总分：%.1f\n\n> <div class=\"open-micro-app\" data-app-config='%s'>查看详情：点击查看详情</div>",
					evaluation.Template.Name,
					periodValue,
					evaluation.Employee.Name,
					evaluation.TotalScore,
					appConfigJSON,
				)

				_ = dooTaskClient.SendBotMessage(hr.DooTaskUserID, message)
			}
		}
	}

	// 发送实时通知
	operatorID := c.GetUint("user_id")
	if updateData.Status != "" {
		// 状态变更通知
		GetNotificationService().SendNotification(operatorID, EventEvaluationStatusChange, &evaluation)
	} else {
		// 一般更新通知
		GetNotificationService().SendNotification(operatorID, EventEvaluationUpdated, &evaluation)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "评估更新成功",
		"data":    evaluation,
	})
}

// 删除评估
func DeleteEvaluation(c *gin.Context) {
	id := c.Param("id")
	evaluationId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的评估ID",
		})
		return
	}

	// 在删除前获取评估信息用于通知
	var evaluation models.KPIEvaluation
	if err := models.DB.Preload("Employee").First(&evaluation, evaluationId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "评估不存在",
		})
		return
	}

	// 删除相关的评分记录
	models.DB.Where("evaluation_id = ?", evaluationId).Delete(&models.KPIScore{})

	// 删除相关的邀请记录
	models.DB.Where("evaluation_id = ?", evaluationId).Delete(&models.EvaluationInvitation{})

	result := models.DB.Delete(&models.KPIEvaluation{}, evaluationId)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除评估失败",
			"message": result.Error.Error(),
		})
		return
	}

	// 发送实时通知
	operatorID := c.GetUint("user_id")
	GetNotificationService().SendNotification(operatorID, EventEvaluationDeleted, &evaluation)

	c.JSON(http.StatusOK, gin.H{
		"message": "评估删除成功",
	})
}

// 获取员工的评估记录
func GetEmployeeEvaluations(c *gin.Context) {
	employeeId := c.Param("employeeId")
	empId, err := strconv.ParseUint(employeeId, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的员工ID",
		})
		return
	}

	var evaluations []models.KPIEvaluation
	result := models.DB.Preload("Template").Preload("Scores").Where("employee_id = ?", empId).Find(&evaluations)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取员工评估记录失败",
			"message": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  evaluations,
		"total": len(evaluations),
	})
}

// 获取待处理的评估
func GetPendingEvaluations(c *gin.Context) {
	employeeId := c.Param("employeeId")
	empId, err := strconv.ParseUint(employeeId, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的员工ID",
		})
		return
	}

	var evaluations []models.KPIEvaluation

	// 获取需要当前员工处理的评估
	result := models.DB.Preload("Employee.Department").Preload("Template").Where("employee_id = ? AND status IN ?", empId, []string{"pending", "self_evaluated"}).Find(&evaluations)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取待处理评估失败",
			"message": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  evaluations,
		"total": len(evaluations),
	})
}

// 获取待确认评估数量
func GetPendingCountEvaluations(c *gin.Context) {
	userID := c.GetUint("user_id")

	// 获取当前用户信息，用于根据角色计算待处理数量
	var user models.Employee
	if err := models.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}

	var totalCount int64

	// 所有角色：自己的待自评 + 待确认
	if err := models.DB.Model(&models.KPIEvaluation{}).
		Where("employee_id = ? AND status IN ?", userID, []string{"pending", "pending_confirm"}).
		Count(&totalCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取待确认评估数量失败"})
		return
	}

	// 主管：增加部门内员工的 self_evaluated（待主管评估）
	if user.Role == "manager" {
		var deptSelfEvaluatedCount int64
		if err := models.DB.Model(&models.KPIEvaluation{}).
			Joins("JOIN employees ON employees.id = kpi_evaluations.employee_id").
			Where("employees.department_id = ? AND kpi_evaluations.status = ?", user.DepartmentID, "self_evaluated").
			Count(&deptSelfEvaluatedCount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取待确认评估数量失败"})
			return
		}
		totalCount += deptSelfEvaluatedCount
	}

	// HR：增加所有 manager_evaluated（待HR审核）+ 部门内员工的 self_evaluated
	if user.Role == "hr" {
		var managerEvaluatedCount int64
		if err := models.DB.Model(&models.KPIEvaluation{}).
			Where("status = ?", "manager_evaluated").
			Count(&managerEvaluatedCount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取待确认评估数量失败"})
			return
		}
		totalCount += managerEvaluatedCount

		var deptSelfEvaluatedCount int64
		if err := models.DB.Model(&models.KPIEvaluation{}).
			Joins("JOIN employees ON employees.id = kpi_evaluations.employee_id").
			Where("employees.department_id = ? AND kpi_evaluations.status = ?", user.DepartmentID, "self_evaluated").
			Count(&deptSelfEvaluatedCount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取待确认评估数量失败"})
			return
		}
		totalCount += deptSelfEvaluatedCount
	}

	c.JSON(http.StatusOK, gin.H{
		"count": totalCount,
	})
}

// 获取评估的评分记录
func GetEvaluationScores(c *gin.Context) {
	evaluationId := c.Param("evaluationId")
	evalId, err := strconv.ParseUint(evaluationId, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的评估ID",
		})
		return
	}

	var scores []models.KPIScore
	result := models.DB.Preload("Item").Where("evaluation_id = ?", evalId).Find(&scores)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取评分记录失败",
			"message": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  scores,
		"total": len(scores),
	})
}

// 更新自评分数
func UpdateSelfScore(c *gin.Context) {
	id := c.Param("id")
	scoreId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的评分ID",
		})
		return
	}

	var score models.KPIScore
	result := models.DB.Preload("Evaluation.Employee").First(&score, scoreId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "评分记录不存在",
		})
		return
	}

	var updateData struct {
		SelfScore   *float64 `json:"self_score"`
		SelfComment string   `json:"self_comment"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	// 检查是否是第一次保存自评分数（开始自评时）
	// 如果之前没有自评分数，且现在要保存自评分数，则检查是否有直属上级
	if score.SelfScore == nil && updateData.SelfScore != nil {
		// 检查被评估员工是否有直属上级（所有角色都需要检查）
		if score.Evaluation.Employee.ManagerID == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "暂无直属上级，请联系HR",
			})
			return
		}
	}

	result = models.DB.Model(&score).Updates(map[string]interface{}{
		"self_score":   updateData.SelfScore,
		"self_comment": updateData.SelfComment,
	})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新自评分数失败",
			"message": result.Error.Error(),
		})
		return
	}

	// 发送实时通知
	operatorID := c.GetUint("user_id")
	GetNotificationService().SendNotification(operatorID, EventSelfScoreUpdated, &score)

	c.JSON(http.StatusOK, gin.H{
		"message": "自评分数更新成功",
		"data":    score,
	})
}

// 更新上级评分
func UpdateManagerScore(c *gin.Context) {
	id := c.Param("id")
	scoreId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的评分ID",
		})
		return
	}

	var score models.KPIScore
	result := models.DB.First(&score, scoreId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "评分记录不存在",
		})
		return
	}

	var updateData struct {
		ManagerScore   *float64 `json:"manager_score"`
		ManagerComment string   `json:"manager_comment"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	result = models.DB.Model(&score).Updates(map[string]interface{}{
		"manager_score":   updateData.ManagerScore,
		"manager_comment": updateData.ManagerComment,
	})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新上级评分失败",
			"message": result.Error.Error(),
		})
		return
	}

	// 发送实时通知
	operatorID := c.GetUint("user_id")
	GetNotificationService().SendNotification(operatorID, EventManagerScoreUpdated, &score)

	c.JSON(http.StatusOK, gin.H{
		"message": "上级评分更新成功",
		"data":    score,
	})
}

// 更新HR评分
func UpdateHRScore(c *gin.Context) {
	id := c.Param("id")
	scoreId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的评分ID",
		})
		return
	}

	var score models.KPIScore
	result := models.DB.First(&score, scoreId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "评分记录不存在",
		})
		return
	}

	var updateData struct {
		HRScore   *float64 `json:"hr_score"`
		HRComment string   `json:"hr_comment"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	result = models.DB.Model(&score).Updates(map[string]interface{}{
		"hr_score":   updateData.HRScore,
		"hr_comment": updateData.HRComment,
	})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新HR评分失败",
			"message": result.Error.Error(),
		})
		return
	}

	// 发送实时通知
	operatorID := c.GetUint("user_id")
	GetNotificationService().SendNotification(operatorID, EventHRScoreUpdated, &score)

	c.JSON(http.StatusOK, gin.H{
		"message": "HR评分更新成功",
		"data":    score,
	})
}

// 更新最终得分
func UpdateFinalScore(c *gin.Context) {
	id := c.Param("id")
	scoreId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的评分ID",
		})
		return
	}

	var score models.KPIScore
	result := models.DB.First(&score, scoreId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "评分记录不存在",
		})
		return
	}

	var updateData struct {
		FinalScore   *float64 `json:"final_score"`
		FinalComment string   `json:"final_comment"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	result = models.DB.Model(&score).Updates(map[string]interface{}{
		"final_score":   updateData.FinalScore,
		"final_comment": updateData.FinalComment,
	})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新最终得分失败",
			"message": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "最终得分更新成功",
		"data":    score,
	})
}

// 员工提交异议
func SubmitObjection(c *gin.Context) {
	id := c.Param("id")
	evaluationId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的评估ID",
		})
		return
	}

	var evaluation models.KPIEvaluation
	result := models.DB.Preload("Employee").Preload("Employee.Manager").First(&evaluation, evaluationId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "评估不存在",
		})
		return
	}

	// 检查权限：只有被考核员工本人可以提交异议
	userID := c.GetUint("user_id")
	if evaluation.EmployeeID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "无权操作此评估",
		})
		return
	}

	// 检查状态：只有在待确认状态才能提交异议
	if evaluation.Status != "pending_confirm" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "只有在待确认状态才能提交异议",
		})
		return
	}

	// 检查是否已有异议（包括已处理的情况）
	if evaluation.HasObjection || evaluation.ObjectionReason != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "已提交过异议，不可重复提交",
		})
		return
	}

	var objectionData struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&objectionData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	// 更新评估，添加异议
	result = models.DB.Model(&evaluation).Updates(map[string]interface{}{
		"has_objection":    true,
		"objection_reason": objectionData.Reason,
	})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "提交异议失败",
			"message": result.Error.Error(),
		})
		return
	}

	// 重新加载评估数据
	models.DB.Preload("Employee").Preload("Employee.Manager").Preload("Template").First(&evaluation, evaluationId)

	// 发送 DooTask 机器人通知给主管和所有HR
	dooTaskClient := utils.NewDooTaskClient(c.GetHeader("DooTaskAuth"))
	periodValue := utils.GetPeriodValue(evaluation.Period, evaluation.Year, evaluation.Month, evaluation.Quarter)

	appConfigJSON := utils.BuildKPIAppConfig(evaluation.ID)

	// 通知主管（如果有）
	if evaluation.Employee.Manager != nil && evaluation.Employee.Manager.DooTaskUserID != nil {
		message := fmt.Sprintf(
			"**有绩效异议待处理**\n- 部门员工：%s\n- 考核模板：%s\n- 考核周期：%s\n- 异议原因：%s\n\n> <div class=\"open-micro-app\" data-app-config='%s'>查看详情：点击查看详情</div>",
			evaluation.Employee.Name,
			evaluation.Template.Name,
			periodValue,
			objectionData.Reason,
			appConfigJSON,
		)
		_ = dooTaskClient.SendBotMessage(evaluation.Employee.Manager.DooTaskUserID, message)
	}

	// 通知所有HR
	hrUserIDsForBot := GetNotificationService().GetAllHRUsers()
	for _, hrID := range hrUserIDsForBot {
		var hr models.Employee
		if err := models.DB.First(&hr, hrID).Error; err != nil || hr.DooTaskUserID == nil {
			continue
		}

		message := fmt.Sprintf(
			"**有绩效异议待处理**\n- 部门员工：%s\n- 考核模板：%s\n- 考核周期：%s\n- 异议原因：%s\n\n> <div class=\"open-micro-app\" data-app-config='%s'>查看详情：点击查看详情</div>",
			evaluation.Employee.Name,
			evaluation.Template.Name,
			periodValue,
			objectionData.Reason,
			appConfigJSON,
		)
		_ = dooTaskClient.SendBotMessage(hr.DooTaskUserID, message)
	}

	// 发送通知给上级和HR
	notificationService := GetNotificationService()

	// 通知上级（如果有）
	if evaluation.Employee.Manager != nil {
		notificationService.SendNotification(evaluation.Employee.Manager.ID, EventObjectionSubmitted, &evaluation)
	}

	// 通知所有HR
	hrUserIDs := notificationService.GetAllHRUsers()
	for _, hrID := range hrUserIDs {
		notificationService.SendNotification(hrID, EventObjectionSubmitted, &evaluation)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "异议提交成功",
		"data":    evaluation,
	})
}

// HR处理异议
func HandleObjection(c *gin.Context) {
	id := c.Param("id")
	evaluationId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的评估ID",
		})
		return
	}

	var evaluation models.KPIEvaluation
	result := models.DB.Preload("Employee").First(&evaluation, evaluationId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "评估不存在",
		})
		return
	}

	// 检查是否有异议
	if !evaluation.HasObjection {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "该评估没有异议需要处理",
		})
		return
	}

	var handleData struct {
		TotalScore   float64 `json:"total_score" binding:"required"`
		FinalComment string  `json:"final_comment" binding:"required"`
	}

	if err := c.ShouldBindJSON(&handleData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	// 更新评估：处理异议，清除异议状态，更新最终得分和处理原因
	result = models.DB.Model(&evaluation).Updates(map[string]interface{}{
		"has_objection": false,
		"total_score":   handleData.TotalScore,
		"final_comment": handleData.FinalComment,
	})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "处理异议失败",
			"message": result.Error.Error(),
		})
		return
	}

	// 重新加载评估数据
	models.DB.Preload("Employee").Preload("Template").First(&evaluation, evaluationId)

	// 发送 DooTask 机器人通知给员工（和可选主管）
	dooTaskClient := utils.NewDooTaskClient(c.GetHeader("DooTaskAuth"))
	periodValue := utils.GetPeriodValue(evaluation.Period, evaluation.Year, evaluation.Month, evaluation.Quarter)

	appConfigJSON := utils.BuildKPIAppConfig(evaluation.ID)

	// 通知员工
	if evaluation.Employee.DooTaskUserID != nil {
		message := fmt.Sprintf(
			"**你的绩效异议已处理，请重新确认**\n- 考核模板：%s\n- 考核周期：%s\n- 调整后总分：%.2f\n- 处理说明：%s\n\n> <div class=\"open-micro-app\" data-app-config='%s'>查看详情：点击查看详情</div>",
			evaluation.Template.Name,
			periodValue,
			handleData.TotalScore,
			handleData.FinalComment,
			appConfigJSON,
		)
		_ = dooTaskClient.SendBotMessage(evaluation.Employee.DooTaskUserID, message)
	}

	// 通知主管（如果有）
	var employee models.Employee
	if err := models.DB.Preload("Manager").First(&employee, evaluation.EmployeeID).Error; err == nil {
		if employee.Manager != nil && employee.Manager.DooTaskUserID != nil {
			message := fmt.Sprintf(
				"**部门员工的绩效异议已处理**\n- 部门员工：%s\n- 考核模板：%s\n- 考核周期：%s\n- 调整后总分：%.2f\n\n> <div class=\"open-micro-app\" data-app-config='%s'>查看详情：点击查看详情</div>",
				employee.Name,
				evaluation.Template.Name,
				periodValue,
				handleData.TotalScore,
				appConfigJSON,
			)
			_ = dooTaskClient.SendBotMessage(employee.Manager.DooTaskUserID, message)
		}
	}

	// 发送通知给员工
	GetNotificationService().SendNotification(evaluation.EmployeeID, EventObjectionHandled, &evaluation)

	c.JSON(http.StatusOK, gin.H{
		"message": "异议处理成功",
		"data":    evaluation,
	})
}

const (
	performanceScenarioNoInvitation       = "no_invitation"
	performanceScenarioEmployeeInvitation = "employee_invitation"

	autoHRScoreComment = "系统根据绩效规则自动计算"
)

type invitationAggregate struct {
	Average float64
	Count   int
}

type scoreComponent struct {
	weight  float64
	value   float64
	present bool
}

// applyPerformanceRuleForEvaluation 根据启用的绩效规则自动计算HR评分和总分
func applyPerformanceRuleForEvaluation(evaluationID uint) error {
	var rule models.PerformanceRule
	if err := models.DB.First(&rule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	if !rule.Enabled {
		return nil
	}

	var evaluation models.KPIEvaluation
	if err := models.DB.Preload("Employee").First(&evaluation, evaluationID).Error; err != nil {
		return err
	}

	var scores []models.KPIScore
	if err := models.DB.Where("evaluation_id = ?", evaluationID).Find(&scores).Error; err != nil {
		return err
	}
	if len(scores) == 0 {
		return nil
	}

	var invitations []models.EvaluationInvitation
	if err := models.DB.
		Preload("Invitee").
		Preload("Scores").
		Where("evaluation_id = ? AND status = ?", evaluationID, "completed").
		Find(&invitations).Error; err != nil {
		return err
	}

	scenario, relevantInvitations := determinePerformanceRuleScenario(invitations)
	invitationAverages := buildInvitationAverages(relevantInvitations)

	tx := models.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	totalScore := 0.0
	updatedCount := 0

	for _, score := range scores {
		aggregate := invitationAverages[score.ItemID]
		hrScore, ok := calculateHRScoreByScenario(score, aggregate, scenario, rule)
		if !ok {
			continue
		}

		comment := score.HRComment
		if strings.TrimSpace(comment) == "" {
			comment = autoHRScoreComment
		}

		if err := tx.Model(&models.KPIScore{}).Where("id = ?", score.ID).Updates(map[string]interface{}{
			"hr_score":   hrScore,
			"hr_comment": comment,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}

		totalScore += hrScore
		updatedCount++
	}

	if updatedCount > 0 {
		totalScore = math.Round(totalScore*100) / 100
		if err := tx.Model(&models.KPIEvaluation{}).Where("id = ?", evaluationID).Update("total_score", totalScore).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// determinePerformanceRuleScenario 根据邀请情况确定使用的绩效规则场景
func determinePerformanceRuleScenario(invitations []models.EvaluationInvitation) (string, []models.EvaluationInvitation) {
	validInvitations := make([]models.EvaluationInvitation, 0, len(invitations))
	for _, invitation := range invitations {
		if invitationHasCompletedScore(invitation) {
			validInvitations = append(validInvitations, invitation)
		}
	}

	if len(validInvitations) == 0 {
		return performanceScenarioNoInvitation, nil
	}

	return performanceScenarioEmployeeInvitation, validInvitations
}

func invitationHasCompletedScore(invitation models.EvaluationInvitation) bool {
	for _, score := range invitation.Scores {
		if score.Score != nil {
			return true
		}
	}
	return false
}

func buildInvitationAverages(invitations []models.EvaluationInvitation) map[uint]invitationAggregate {
	aggregates := make(map[uint]invitationAggregate)

	for _, invitation := range invitations {
		for _, score := range invitation.Scores {
			if score.Score == nil {
				continue
			}

			aggregate := aggregates[score.ItemID]
			aggregate.Average += *score.Score
			aggregate.Count++
			aggregates[score.ItemID] = aggregate
		}
	}

	for itemID, aggregate := range aggregates {
		if aggregate.Count > 0 {
			aggregate.Average = aggregate.Average / float64(aggregate.Count)
			aggregates[itemID] = aggregate
		} else {
			delete(aggregates, itemID)
		}
	}

	return aggregates
}

func calculateHRScoreByScenario(score models.KPIScore, aggregate invitationAggregate, scenario string, rule models.PerformanceRule) (float64, bool) {
	selfValue, hasSelf := valueFromPointer(score.SelfScore)
	managerValue, hasManager := valueFromPointer(score.ManagerScore)
	inviteValue := aggregate.Average
	hasInvite := aggregate.Count > 0

	var components []scoreComponent

	switch scenario {
	case performanceScenarioEmployeeInvitation:
		components = []scoreComponent{
			{weight: rule.WithInvitation.Employee.SelfWeight, value: selfValue, present: hasSelf},
			{weight: rule.WithInvitation.Employee.InviteSuperiorWeight, value: inviteValue, present: hasInvite},
			{weight: rule.WithInvitation.Employee.SuperiorWeight, value: managerValue, present: hasManager},
		}
	default:
		components = []scoreComponent{
			{weight: rule.NoInvitation.SelfWeight, value: selfValue, present: hasSelf},
			{weight: rule.NoInvitation.SuperiorWeight, value: managerValue, present: hasManager},
		}
	}

	result, ok := weightedAverage(components)
	if !ok {
		return 0, false
	}

	result = math.Round(result*100) / 100
	return result, true
}

func valueFromPointer(val *float64) (float64, bool) {
	if val == nil {
		return 0, false
	}
	return *val, true
}

func weightedAverage(components []scoreComponent) (float64, bool) {
	totalWeight := 0.0
	for _, component := range components {
		if component.present && component.weight > 0 {
			totalWeight += component.weight
		}
	}

	if totalWeight == 0 {
		return 0, false
	}

	totalValue := 0.0
	for _, component := range components {
		if component.present && component.weight > 0 {
			totalValue += (component.weight / totalWeight) * component.value
		}
	}

	return totalValue, true
}
