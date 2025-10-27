package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"dootask-kpi-server/models"
	"dootask-kpi-server/utils"

	"github.com/gin-gonic/gin"
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
	// 周期筛选参数
	period := c.Query("period")
	yearStr := c.Query("year")
	monthStr := c.Query("month")
	quarterStr := c.Query("quarter")

	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询
	query := models.DB.Preload("Employee.Department").Preload("Template").Preload("Scores").Preload("Scores.Item")

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

	// 添加周期筛选条件
	if period != "" {
		query = query.Where("period = ?", period)
	}
	if yearStr != "" {
		year, _ := strconv.Atoi(yearStr)
		query = query.Where("year = ?", year)
	}
	if monthStr != "" {
		month, _ := strconv.Atoi(monthStr)
		query = query.Where("month = ?", month)
	}
	if quarterStr != "" {
		quarter, _ := strconv.Atoi(quarterStr)
		query = query.Where("quarter = ?", quarter)
	}

	// 获取总数
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

	// 在计数查询中也添加周期筛选条件
	if period != "" {
		countQuery = countQuery.Where("period = ?", period)
	}
	if yearStr != "" {
		year, _ := strconv.Atoi(yearStr)
		countQuery = countQuery.Where("year = ?", year)
	}
	if monthStr != "" {
		month, _ := strconv.Atoi(monthStr)
		countQuery = countQuery.Where("month = ?", month)
	}
	if quarterStr != "" {
		quarter, _ := strconv.Atoi(quarterStr)
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

	row := models.KPIEvaluation{}
	result := tx.Preload("Employee").
		Where("employee_id = ? AND template_id = ? AND period = ? AND year = ? AND month = ?",
			evaluation.EmployeeID, evaluation.TemplateID, evaluation.Period, evaluation.Year, evaluation.Month).
		Find(&row).Limit(1)
	if result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建评估失败",
		})
		return
	}
	if row.ID > 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("员工【%s】评估记录已存在", row.Employee.Name),
		})
		return
	}

	// 创建评估记录
	result = tx.Create(&evaluation)
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

	// 发送 DooTask 机器人通知
	dooTaskClient := utils.NewDooTaskClient(c.GetHeader("DooTaskAuth"))
	dooTaskClient.SendBotMessage(evaluation.Employee.DooTaskUserID, fmt.Sprintf(
		"### 📋 您有新的考核任务，请及时处理。\n\n- **考核模板：** %s\n- **考核周期：** %s\n- **考核时间：** %s\n- **发起人：** %s\n\n> 请前往「应用 - 绩效考核」中查看详情。",
		evaluation.Template.Name,
		utils.GetPeriodValue(evaluation.Period, evaluation.Year, evaluation.Month, evaluation.Quarter),
		evaluation.CreatedAt.Format("2006-01-02"),
		c.GetString("user_name"),
	))

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
	if updateData.Status == "self_evaluated" {
		// 检查员工是否有直属主管
		if evaluation.Employee.ManagerID == nil {
			// 如果没有主管，直接将状态改为manager_evaluated
			updateData.Status = "manager_evaluated"

			// 自动填入主管评分：将自评分数复制到主管评分
			var scores []models.KPIScore
			if err := models.DB.Where("evaluation_id = ?", evaluation.ID).Find(&scores).Error; err == nil {
				for _, score := range scores {
					if score.SelfScore != nil {
						// 将自评分数复制到主管评分
						comment := "（自评分数）"
						models.DB.Model(&score).Updates(map[string]interface{}{
							"manager_score":   *score.SelfScore,
							"manager_comment": comment,
							"manager_auto":    true,
						})
					}
				}
			}
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

	// 如果状态变为completed，自动计算最终得分
	if updateData.Status == "completed" {
		var scores []models.KPIScore
		if err := models.DB.Where("evaluation_id = ?", evaluation.ID).Find(&scores).Error; err == nil {
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
				models.DB.Model(&s).Update("final_score", final)
				total += final
			}
			models.DB.Model(&evaluation).Update("total_score", total)
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
				message := fmt.Sprintf(
					"### 📋 「%s」已完成自评，请您进行主管评估。\n\n- **考核模板：** %s\n- **考核周期：** %s\n- **员工姓名：** %s\n\n> 请前往「应用 - 绩效考核」中查看详情。",
					evaluation.Employee.Name,
					evaluation.Template.Name,
					periodValue,
					evaluation.Employee.Name,
				)
				dooTaskClient.SendBotMessage(evaluation.Employee.Manager.DooTaskUserID, message)
			}

		case "pending_confirm":
			// 完成HR审核：通知员工确认
			if evaluation.Employee.DooTaskUserID != nil {
				message := fmt.Sprintf(
					"### 📋 您的考核已完成HR审核，请确认最终得分。\n\n- **考核模板：** %s\n- **考核周期：** %s\n- **总分：** %.1f\n\n> 请前往「应用 - 绩效考核」中查看详情并确认。",
					evaluation.Template.Name,
					periodValue,
					evaluation.TotalScore,
				)
				dooTaskClient.SendBotMessage(evaluation.Employee.DooTaskUserID, message)
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

	var count int64
	if err := models.DB.Model(&models.KPIEvaluation{}).
		Where("employee_id = ? AND status = ?", userID, "pending").
		Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取待确认评估数量失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": count,
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
	result := models.DB.First(&score, scoreId)
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
