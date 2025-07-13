package handlers

import (
	"net/http"
	"strconv"
	"time"

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

	// 获取总数
	var total int64
	countQuery := models.DB.Model(&models.KPIEvaluation{})
	if status != "" {
		countQuery = countQuery.Where("status = ?", status)
	}
	if employeeID != "" {
		countQuery = countQuery.Where("employee_id = ?", employeeID)
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
// 时间检查请求结构
type TimeCheckRequest struct {
	Period   string `json:"period"`   // monthly, quarterly, yearly
	Year     int    `json:"year"`
	Month    *int   `json:"month,omitempty"`
	Quarter  *int   `json:"quarter,omitempty"`
}

// 时间检查响应结构
type TimeCheckResponse struct {
	IsValid         bool                `json:"is_valid"`
	TimeMode        string              `json:"time_mode"`
	Message         string              `json:"message,omitempty"`
	AvailableDays   int                 `json:"available_days"`
	RecommendedDeadlines utils.DeadlineSet `json:"recommended_deadlines"`
}

// 检查时间可用性
func CheckTimeAvailability(c *gin.Context) {
	var req TimeCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	// 加载系统配置
	calculator, err := loadDeadlineCalculator(req.Period, req.Year, req.Month, req.Quarter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "加载系统配置失败",
			"message": err.Error(),
		})
		return
	}

	// 计算推荐的截止时间
	deadlines := calculator.CalculateDeadlines()
	availableDays := calculator.CalculateAvailableDays()

	response := TimeCheckResponse{
		IsValid:              deadlines.IsValid,
		TimeMode:             deadlines.TimeMode,
		Message:              deadlines.Message,
		AvailableDays:        availableDays,
		RecommendedDeadlines: deadlines,
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// 创建评估（修改版，支持截止时间）
func CreateEvaluation(c *gin.Context) {
	var evaluation models.KPIEvaluation

	if err := c.ShouldBindJSON(&evaluation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	// 如果没有提供截止时间，则自动计算
	if evaluation.SelfEvalDeadline == nil {
		calculator, err := loadDeadlineCalculator(evaluation.Period, evaluation.Year, evaluation.Month, evaluation.Quarter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "加载系统配置失败",
				"message": err.Error(),
			})
			return
		}

		deadlines := calculator.CalculateDeadlines()
		if !deadlines.IsValid {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "时间不足",
				"message": deadlines.Message,
			})
			return
		}

		// 设置计算得出的截止时间
		evaluation.PeriodEndDate = deadlines.PeriodEnd
		evaluation.SelfEvalDeadline = deadlines.SelfEvalDeadline
		evaluation.ManagerEvalDeadline = deadlines.ManagerEvalDeadline
		evaluation.HRReviewDeadline = deadlines.HRReviewDeadline
		evaluation.FinalConfirmDeadline = deadlines.FinalConfirmDeadline
		evaluation.TimeMode = deadlines.TimeMode
	} else {
		// 如果提供了自定义截止时间，则验证
		calculator, err := loadDeadlineCalculator(evaluation.Period, evaluation.Year, evaluation.Month, evaluation.Quarter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "加载系统配置失败",
				"message": err.Error(),
			})
			return
		}

		customDeadlines := calculator.CreateCustomDeadlines(
			*evaluation.SelfEvalDeadline,
			*evaluation.ManagerEvalDeadline,
			*evaluation.HRReviewDeadline,
			*evaluation.FinalConfirmDeadline,
		)

		if !customDeadlines.IsValid {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "自定义截止时间无效",
				"message": customDeadlines.Message,
			})
			return
		}

		evaluation.PeriodEndDate = customDeadlines.PeriodEnd
		evaluation.TimeMode = "custom"
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

	c.JSON(http.StatusCreated, gin.H{
		"message": "评估创建成功",
		"data":    evaluation,
	})
}

// 加载截止时间计算器
func loadDeadlineCalculator(period string, year int, month *int, quarter *int) (*utils.DeadlineCalculator, error) {
	// 获取系统配置
	var standardSetting models.SystemSetting
	if err := models.DB.Where("key = ?", "deadline_standard_days").First(&standardSetting).Error; err != nil {
		return nil, err
	}
	standardDays, err := utils.ParseDeadlineDaysFromJSON(standardSetting.Value)
	if err != nil {
		return nil, err
	}

	var compressedSetting models.SystemSetting
	if err := models.DB.Where("key = ?", "deadline_compressed_days").First(&compressedSetting).Error; err != nil {
		return nil, err
	}
	compressedDays, err := utils.ParseDeadlineDaysFromJSON(compressedSetting.Value)
	if err != nil {
		return nil, err
	}

	var minimumSetting models.SystemSetting
	if err := models.DB.Where("key = ?", "deadline_minimum_days").First(&minimumSetting).Error; err != nil {
		return nil, err
	}
	minimumDays, err := utils.ParseDeadlineDaysFromJSON(minimumSetting.Value)
	if err != nil {
		return nil, err
	}

	var thresholdSetting models.SystemSetting
	if err := models.DB.Where("key = ?", "deadline_time_threshold").First(&thresholdSetting).Error; err != nil {
		return nil, err
	}
	threshold, err := utils.ParseTimeThresholdFromJSON(thresholdSetting.Value)
	if err != nil {
		return nil, err
	}

	calculator := &utils.DeadlineCalculator{
		PeriodType:     period,
		CreatedAt:      time.Now(),
		Year:           year,
		Month:          month,
		Quarter:        quarter,
		StandardDays:   standardDays,
		CompressedDays: compressedDays,
		MinimumDays:    minimumDays,
		TimeThreshold:  threshold,
	}

	return calculator, nil
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
				if s.ManagerScore != nil {
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
	models.DB.Preload("Employee").First(&evaluation, evaluationId)

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

	// 删除相关的评分记录
	models.DB.Where("evaluation_id = ?", evaluationId).Delete(&models.KPIScore{})

	result := models.DB.Delete(&models.KPIEvaluation{}, evaluationId)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除评估失败",
			"message": result.Error.Error(),
		})
		return
	}

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

	c.JSON(http.StatusOK, gin.H{
		"message": "上级评分更新成功",
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
		FinalScore *float64 `json:"final_score"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	result = models.DB.Model(&score).Updates(map[string]interface{}{
		"final_score": updateData.FinalScore,
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
