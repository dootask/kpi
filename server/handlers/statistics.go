package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"dootask-kpi-server/models"

	"github.com/gin-gonic/gin"
)

// 统计分析

// 最近评估结构
type RecentEvaluation struct {
	ID         uint           `json:"id"`
	Employee   RecentEmployee `json:"employee"`
	TotalScore float64        `json:"total_score"`
	Status     string         `json:"status"`
	Period     string         `json:"period"`
	Year       int            `json:"year"`
	Month      *int           `json:"month"`
	Quarter    *int           `json:"quarter"`
}

// 最近员工结构
type RecentEmployee struct {
	Name       string           `json:"name"`
	Position   string           `json:"position"`
	Department RecentDepartment `json:"department"`
}

// 最近部门结构
type RecentDepartment struct {
	Name string `json:"name"`
}

// 获取仪表板统计数据
func GetDashboardStats(c *gin.Context) {
	// 获取查询参数
	year := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	period := c.DefaultQuery("period", "")
	month := c.DefaultQuery("month", "")
	quarter := c.DefaultQuery("quarter", "")

	var stats struct {
		TotalEmployees       int64              `json:"total_employees"`
		TotalDepartments     int64              `json:"total_departments"`
		TotalEvaluations     int64              `json:"total_evaluations"`
		PendingEvaluations   int64              `json:"pending_evaluations"`
		CompletedEvaluations int64              `json:"completed_evaluations"`
		AverageScore         float64            `json:"average_score"`
		RecentEvaluations    []RecentEvaluation `json:"recent_evaluations"`
	}

	// 获取基本统计数据（员工数和部门数不受时间筛选影响）
	// 只统计在职员工
	models.DB.Model(&models.Employee{}).Where("is_active = ?", true).Count(&stats.TotalEmployees)
	models.DB.Model(&models.Department{}).Count(&stats.TotalDepartments)

	// 构建评估数据的时间筛选查询（只统计在职员工的评估）
	baseQuery := models.DB.Model(&models.KPIEvaluation{}).
		Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
		Where("employees.is_active = ?", true)

	// 根据时间筛选条件构建查询
	if period != "" {
		if period == "monthly" && month != "" {
			baseQuery = baseQuery.Where("period = ? AND year = ? AND month = ?", "monthly", year, month)
		} else if period == "quarterly" && quarter != "" {
			baseQuery = baseQuery.Where("period = ? AND year = ? AND quarter = ?", "quarterly", year, quarter)
		} else if period == "yearly" {
			// 兼容历史数据格式，支持 period="yearly" 和 period="年份"
			baseQuery = baseQuery.Where("(period = ? OR period = ?) AND year = ?", "yearly", year, year)
		}
	}

	// 获取筛选后的评估统计数据
	baseQuery.Count(&stats.TotalEvaluations)

	// 创建baseQuery的副本来分别查询不同状态的数据（只统计在职员工的评估）
	pendingQuery := models.DB.Model(&models.KPIEvaluation{}).
		Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
		Where("employees.is_active = ?", true)
	completedQuery := models.DB.Model(&models.KPIEvaluation{}).
		Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
		Where("employees.is_active = ?", true)
	avgQuery := models.DB.Model(&models.KPIEvaluation{}).
		Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
		Where("employees.is_active = ?", true)

	// 应用相同的时间筛选条件
	if period != "" {
		if period == "monthly" && month != "" {
			pendingQuery = pendingQuery.Where("period = ? AND year = ? AND month = ?", "monthly", year, month)
			completedQuery = completedQuery.Where("period = ? AND year = ? AND month = ?", "monthly", year, month)
			avgQuery = avgQuery.Where("period = ? AND year = ? AND month = ?", "monthly", year, month)
		} else if period == "quarterly" && quarter != "" {
			pendingQuery = pendingQuery.Where("period = ? AND year = ? AND quarter = ?", "quarterly", year, quarter)
			completedQuery = completedQuery.Where("period = ? AND year = ? AND quarter = ?", "quarterly", year, quarter)
			avgQuery = avgQuery.Where("period = ? AND year = ? AND quarter = ?", "quarterly", year, quarter)
		} else if period == "yearly" {
			pendingQuery = pendingQuery.Where("(period = ? OR period = ?) AND year = ?", "yearly", year, year)
			completedQuery = completedQuery.Where("(period = ? OR period = ?) AND year = ?", "yearly", year, year)
			avgQuery = avgQuery.Where("(period = ? OR period = ?) AND year = ?", "yearly", year, year)
		}
	}

	// 分别查询不同状态的数据
	pendingQuery.Where("status = ?", "pending").Count(&stats.PendingEvaluations)
	completedQuery.Where("status = ?", "completed").Count(&stats.CompletedEvaluations)

	// 计算筛选后的平均得分（只计算已完成的评估）
	var avgResult struct {
		AvgScore float64
	}
	avgQuery.Select("AVG(total_score) as avg_score").Where("status = ?", "completed").Scan(&avgResult)
	stats.AverageScore = avgResult.AvgScore

	// 获取最近的评估记录（应用相同的时间筛选，只显示在职员工的评估）
	var recentEvals []models.KPIEvaluation
	recentQuery := models.DB.Preload("Employee.Department").Preload("Template").
		Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
		Where("employees.is_active = ?", true)
	if period != "" {
		if period == "monthly" && month != "" {
			recentQuery = recentQuery.Where("period = ? AND year = ? AND month = ?", "monthly", year, month)
		} else if period == "quarterly" && quarter != "" {
			recentQuery = recentQuery.Where("period = ? AND year = ? AND quarter = ?", "quarterly", year, quarter)
		} else if period == "yearly" {
			// 兼容历史数据格式，支持 period="yearly" 和 period="年份"
			recentQuery = recentQuery.Where("(period = ? OR period = ?) AND year = ?", "yearly", year, year)
		}
	}
	recentQuery.Order("kpi_evaluations.created_at DESC").Limit(10).Find(&recentEvals)

	// 构建RecentEvaluation结构体
	for _, eval := range recentEvals {
		stats.RecentEvaluations = append(stats.RecentEvaluations, RecentEvaluation{
			ID: eval.ID,
			Employee: RecentEmployee{
				Name:     eval.Employee.Name,
				Position: eval.Employee.Position,
				Department: RecentDepartment{
					Name: eval.Employee.Department.Name,
				},
			},
			TotalScore: eval.TotalScore,
			Status:     eval.Status,
			Period:     eval.Period,
			Year:       eval.Year,
			Month:      eval.Month,
			Quarter:    eval.Quarter,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}

// 获取部门统计数据
func GetDepartmentStats(c *gin.Context) {
	id := c.Param("id")
	departmentId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的部门ID",
		})
		return
	}

	var stats struct {
		DepartmentInfo  models.Department `json:"department_info"`
		EmployeeCount   int64             `json:"employee_count"`
		EvaluationCount int64             `json:"evaluation_count"`
		AverageScore    float64           `json:"average_score"`
		MonthlyStats    []struct {
			Month           string  `json:"month"`
			EvaluationCount int64   `json:"evaluation_count"`
			AverageScore    float64 `json:"average_score"`
		} `json:"monthly_stats"`
		TopPerformers []struct {
			Employee     models.Employee `json:"employee"`
			AverageScore float64         `json:"average_score"`
		} `json:"top_performers"`
	}

	// 获取部门信息
	if err := models.DB.Preload("Employees").First(&stats.DepartmentInfo, departmentId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "部门不存在",
		})
		return
	}

	// 获取员工数量（只统计在职员工）
	models.DB.Model(&models.Employee{}).Where("department_id = ? AND is_active = ?", departmentId, true).Count(&stats.EmployeeCount)

	// 获取评估数量和平均分（只统计在职员工）
	models.DB.Model(&models.KPIEvaluation{}).
		Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
		Where("employees.department_id = ? AND employees.is_active = ?", departmentId, true).
		Count(&stats.EvaluationCount)

	var avgResult struct {
		AvgScore float64
	}
	models.DB.Model(&models.KPIEvaluation{}).
		Select("AVG(total_score) as avg_score").
		Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
		Where("employees.department_id = ? AND employees.is_active = ? AND kpi_evaluations.status = ?", departmentId, true, "completed").
		Scan(&avgResult)
	stats.AverageScore = avgResult.AvgScore

	// 获取月度统计（最近6个月）
	for i := 5; i >= 0; i-- {
		date := time.Now().AddDate(0, -i, 0)
		month := date.Format("2006-01")

		var monthStats struct {
			EvaluationCount int64
			AverageScore    float64
		}

		models.DB.Model(&models.KPIEvaluation{}).
			Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
			Where("employees.department_id = ? AND employees.is_active = ? AND kpi_evaluations.period LIKE ?", departmentId, true, month+"%").
			Count(&monthStats.EvaluationCount)

		var monthAvg struct {
			AvgScore float64
		}
		models.DB.Model(&models.KPIEvaluation{}).
			Select("AVG(total_score) as avg_score").
			Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
			Where("employees.department_id = ? AND employees.is_active = ? AND kpi_evaluations.period LIKE ? AND kpi_evaluations.status = ?", departmentId, true, month+"%", "completed").
			Scan(&monthAvg)

		stats.MonthlyStats = append(stats.MonthlyStats, struct {
			Month           string  `json:"month"`
			EvaluationCount int64   `json:"evaluation_count"`
			AverageScore    float64 `json:"average_score"`
		}{
			Month:           month,
			EvaluationCount: monthStats.EvaluationCount,
			AverageScore:    monthAvg.AvgScore,
		})
	}

	// 获取部门top performers
	var topPerformers []struct {
		EmployeeID   uint
		AverageScore float64
	}

	models.DB.Model(&models.KPIEvaluation{}).
		Select("employee_id, AVG(total_score) as average_score").
		Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
		Where("employees.department_id = ? AND employees.is_active = ? AND kpi_evaluations.status = ?", departmentId, true, "completed").
		Group("employee_id").
		Order("average_score DESC").
		Limit(5).
		Scan(&topPerformers)

	for _, performer := range topPerformers {
		var employee models.Employee
		models.DB.First(&employee, performer.EmployeeID)
		stats.TopPerformers = append(stats.TopPerformers, struct {
			Employee     models.Employee `json:"employee"`
			AverageScore float64         `json:"average_score"`
		}{
			Employee:     employee,
			AverageScore: performer.AverageScore,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}

// 获取员工统计数据
func GetEmployeeStats(c *gin.Context) {
	id := c.Param("id")
	employeeId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的员工ID",
		})
		return
	}

	var stats struct {
		EmployeeInfo    models.Employee `json:"employee_info"`
		EvaluationCount int64           `json:"evaluation_count"`
		AverageScore    float64         `json:"average_score"`
		LatestScore     float64         `json:"latest_score"`
		ScoreTrend      []struct {
			Period string  `json:"period"`
			Score  float64 `json:"score"`
		} `json:"score_trend"`
		KPIBreakdown []struct {
			ItemName     string  `json:"item_name"`
			AverageScore float64 `json:"average_score"`
			MaxScore     float64 `json:"max_score"`
		} `json:"kpi_breakdown"`
	}

	// 获取员工信息
	if err := models.DB.Preload("Department").Preload("Manager").First(&stats.EmployeeInfo, employeeId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "员工不存在",
		})
		return
	}

	// 获取评估数量和平均分
	models.DB.Model(&models.KPIEvaluation{}).Where("employee_id = ?", employeeId).Count(&stats.EvaluationCount)

	var avgResult struct {
		AvgScore float64
	}
	models.DB.Model(&models.KPIEvaluation{}).
		Select("AVG(total_score) as avg_score").
		Where("employee_id = ? AND status = ?", employeeId, "completed").
		Scan(&avgResult)
	stats.AverageScore = avgResult.AvgScore

	// 获取最新评估分数
	var latestEval models.KPIEvaluation
	if err := models.DB.Where("employee_id = ? AND status = ?", employeeId, "completed").
		Order("created_at DESC").First(&latestEval).Error; err == nil {
		stats.LatestScore = latestEval.TotalScore
	}

	// 获取分数趋势（最近6次评估）
	var evaluations []models.KPIEvaluation
	models.DB.Where("employee_id = ? AND status = ?", employeeId, "completed").
		Order("created_at DESC").Limit(6).Find(&evaluations)

	for i := len(evaluations) - 1; i >= 0; i-- {
		stats.ScoreTrend = append(stats.ScoreTrend, struct {
			Period string  `json:"period"`
			Score  float64 `json:"score"`
		}{
			Period: evaluations[i].Period,
			Score:  evaluations[i].TotalScore,
		})
	}

	// 获取KPI项目分析
	var kpiBreakdown []struct {
		ItemName     string
		AverageScore float64
		MaxScore     float64
	}

	models.DB.Model(&models.KPIScore{}).
		Select("kpi_items.name as item_name, AVG(kpi_scores.final_score) as average_score, kpi_items.max_score").
		Joins("JOIN kpi_items ON kpi_scores.item_id = kpi_items.id").
		Joins("JOIN kpi_evaluations ON kpi_scores.evaluation_id = kpi_evaluations.id").
		Where("kpi_evaluations.employee_id = ? AND kpi_scores.final_score IS NOT NULL", employeeId).
		Group("kpi_items.id, kpi_items.name, kpi_items.max_score").
		Scan(&kpiBreakdown)

	for _, item := range kpiBreakdown {
		stats.KPIBreakdown = append(stats.KPIBreakdown, struct {
			ItemName     string  `json:"item_name"`
			AverageScore float64 `json:"average_score"`
			MaxScore     float64 `json:"max_score"`
		}{
			ItemName:     item.ItemName,
			AverageScore: item.AverageScore,
			MaxScore:     item.MaxScore,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}

// 获取趋势分析
func GetTrends(c *gin.Context) {
	period := c.DefaultQuery("period", "monthly") // monthly, quarterly, yearly

	var trends struct {
		PeriodType  string `json:"period_type"`
		ScoreTrends []struct {
			Period          string  `json:"period"`
			AverageScore    float64 `json:"average_score"`
			EvaluationCount int64   `json:"evaluation_count"`
		} `json:"score_trends"`
		DepartmentTrends []struct {
			DepartmentName  string  `json:"department_name"`
			AverageScore    float64 `json:"average_score"`
			EvaluationCount int64   `json:"evaluation_count"`
		} `json:"department_trends"`
	}

	trends.PeriodType = period

	// 获取时间段列表
	var periods []string
	switch period {
	case "monthly":
		for i := 11; i >= 0; i-- {
			date := time.Now().AddDate(0, -i, 0)
			periods = append(periods, date.Format("2006-01"))
		}
	case "quarterly":
		for i := 7; i >= 0; i-- {
			date := time.Now().AddDate(0, -i*3, 0)
			quarter := (date.Month()-1)/3 + 1
			periods = append(periods, fmt.Sprintf("%d-Q%d", date.Year(), quarter))
		}
	case "yearly":
		for i := 4; i >= 0; i-- {
			date := time.Now().AddDate(-i, 0, 0)
			periods = append(periods, fmt.Sprintf("%d", date.Year()))
		}
	}

	// 获取每个时间段的统计数据
	for _, p := range periods {
		var periodStats struct {
			AverageScore    float64
			EvaluationCount int64
		}

		var pattern string
		switch period {
		case "monthly":
			pattern = p + "%"
		case "quarterly":
			pattern = p + "%"
		case "yearly":
			pattern = p + "%"
		}

		models.DB.Model(&models.KPIEvaluation{}).
			Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
			Where("kpi_evaluations.period LIKE ? AND kpi_evaluations.status = ? AND employees.is_active = ?", pattern, "completed", true).
			Count(&periodStats.EvaluationCount)

		var avgResult struct {
			AvgScore float64
		}
		models.DB.Model(&models.KPIEvaluation{}).
			Select("AVG(kpi_evaluations.total_score) as avg_score").
			Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
			Where("kpi_evaluations.period LIKE ? AND kpi_evaluations.status = ? AND employees.is_active = ?", pattern, "completed", true).
			Scan(&avgResult)

		trends.ScoreTrends = append(trends.ScoreTrends, struct {
			Period          string  `json:"period"`
			AverageScore    float64 `json:"average_score"`
			EvaluationCount int64   `json:"evaluation_count"`
		}{
			Period:          p,
			AverageScore:    avgResult.AvgScore,
			EvaluationCount: periodStats.EvaluationCount,
		})
	}

	// 获取部门趋势
	var deptTrends []struct {
		DepartmentName  string
		AverageScore    float64
		EvaluationCount int64
	}

	models.DB.Model(&models.KPIEvaluation{}).
		Select("departments.name as department_name, AVG(kpi_evaluations.total_score) as average_score, COUNT(*) as evaluation_count").
		Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
		Joins("JOIN departments ON employees.department_id = departments.id").
		Where("kpi_evaluations.status = ? AND employees.is_active = ?", "completed", true).
		Group("departments.id, departments.name").
		Order("average_score DESC").
		Scan(&deptTrends)

	for _, dept := range deptTrends {
		trends.DepartmentTrends = append(trends.DepartmentTrends, struct {
			DepartmentName  string  `json:"department_name"`
			AverageScore    float64 `json:"average_score"`
			EvaluationCount int64   `json:"evaluation_count"`
		}{
			DepartmentName:  dept.DepartmentName,
			AverageScore:    dept.AverageScore,
			EvaluationCount: dept.EvaluationCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": trends,
	})
}

// 导出功能（简化版，实际应用中需要使用Excel库）

// 导出单个评估
func ExportEvaluation(c *gin.Context) {
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

	// 这里应该生成Excel文件，暂时返回JSON数据
	c.JSON(http.StatusOK, gin.H{
		"message":     "导出成功",
		"data":        evaluation,
		"export_type": "evaluation",
	})
}

// 导出部门评估
func ExportDepartmentEvaluations(c *gin.Context) {
	id := c.Param("id")
	departmentId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的部门ID",
		})
		return
	}

	var evaluations []models.KPIEvaluation
	result := models.DB.Preload("Employee.Department").Preload("Template").
		Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
		Where("employees.department_id = ?", departmentId).
		Find(&evaluations)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取部门评估数据失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "导出成功",
		"data":        evaluations,
		"export_type": "department",
	})
}

// 导出周期评估
func ExportPeriodEvaluations(c *gin.Context) {
	period := c.Param("period")

	var evaluations []models.KPIEvaluation
	result := models.DB.Preload("Employee.Department").Preload("Template").
		Where("period LIKE ?", period+"%").
		Find(&evaluations)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取周期评估数据失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "导出成功",
		"data":        evaluations,
		"export_type": "period",
		"period":      period,
	})
}

// 获取完整的统计分析数据
func GetStatisticsData(c *gin.Context) {
	// 获取查询参数
	year := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	period := c.DefaultQuery("period", "monthly")
	month := c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month())))
	quarter := c.DefaultQuery("quarter", strconv.Itoa((int(time.Now().Month())-1)/3+1))

	// 部门统计结构
	type DepartmentStat struct {
		Name      string  `json:"name"`
		Total     int64   `json:"total"`
		Completed int64   `json:"completed"`
		Pending   int64   `json:"pending"`
		AvgScore  float64 `json:"avg_score"`
	}

	// 月度趋势结构
	type MonthlyTrend struct {
		Month          string  `json:"month"`
		Evaluations    int64   `json:"evaluations"`
		AvgScore       float64 `json:"avg_score"`
		CompletionRate float64 `json:"completion_rate"`
	}

	// 分数分布结构
	type ScoreDistrib struct {
		Range string `json:"range"`
		Count int64  `json:"count"`
		Color string `json:"color"`
	}

	// 顶级表现者结构
	type TopPerformer struct {
		Name        string  `json:"name"`
		Department  string  `json:"department"`
		Score       float64 `json:"score"`
		Evaluations int64   `json:"evaluations"`
	}

	// 最近评估结构
	type RecentEvaluation struct {
		ID         uint    `json:"id"`
		Employee   string  `json:"employee"`
		Department string  `json:"department"`
		Template   string  `json:"template"`
		Score      float64 `json:"score"`
		Status     string  `json:"status"`
		Period     string  `json:"period"`
		Year       int     `json:"year"`
		Month      *int    `json:"month"`
		Quarter    *int    `json:"quarter"`
	}

	var response struct {
		DepartmentStats   []DepartmentStat   `json:"departmentStats"`
		MonthlyTrends     []MonthlyTrend     `json:"monthlyTrends"`
		ScoreDistribution []ScoreDistrib     `json:"scoreDistribution"`
		TopPerformers     []TopPerformer     `json:"topPerformers"`
		RecentEvaluations []RecentEvaluation `json:"recentEvaluations"`
	}

	// 1. 获取部门统计数据
	var departments []models.Department
	models.DB.Find(&departments)

	for _, dept := range departments {
		var stat DepartmentStat
		stat.Name = dept.Name

		// 总评估数（只统计在职员工）
		query := models.DB.Model(&models.KPIEvaluation{}).
			Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
			Where("employees.department_id = ? AND employees.is_active = ?", dept.ID, true)

		switch period {
		case "monthly":
			query = query.Where("kpi_evaluations.period = ? AND kpi_evaluations.year = ? AND kpi_evaluations.month = ?", "monthly", year, month)
		case "quarterly":
			query = query.Where("kpi_evaluations.period = ? AND kpi_evaluations.year = ? AND kpi_evaluations.quarter = ?", "quarterly", year, quarter)
		default:
			// 兼容历史数据格式，支持 period="yearly" 和 period="年份"
			query = query.Where("(kpi_evaluations.period = ? OR kpi_evaluations.period = ?) AND kpi_evaluations.year = ?", "yearly", year, year)
		}

		query.Count(&stat.Total)

		// 已完成评估数
		query.Where("kpi_evaluations.status = ?", "completed").Count(&stat.Completed)

		// 待处理评估数
		stat.Pending = stat.Total - stat.Completed

		// 平均分
		var avgResult struct {
			AvgScore float64
		}
		query.Select("AVG(kpi_evaluations.total_score) as avg_score").Scan(&avgResult)
		stat.AvgScore = avgResult.AvgScore

		response.DepartmentStats = append(response.DepartmentStats, stat)
	}

	// 2. 获取月度趋势数据
	for i := 5; i >= 0; i-- {
		date := time.Now().AddDate(0, -i, 0)
		monthStr := date.Format("1月")

		var trend MonthlyTrend
		trend.Month = monthStr

		// 该月评估数量（只统计在职员工）
		models.DB.Model(&models.KPIEvaluation{}).
			Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
			Where("kpi_evaluations.year = ? AND kpi_evaluations.month = ? AND employees.is_active = ?", date.Year(), int(date.Month()), true).
			Count(&trend.Evaluations)

		// 该月平均分（只统计在职员工）
		var avgResult struct {
			AvgScore float64
		}
		models.DB.Model(&models.KPIEvaluation{}).
			Select("AVG(kpi_evaluations.total_score) as avg_score").
			Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
			Where("kpi_evaluations.year = ? AND kpi_evaluations.month = ? AND kpi_evaluations.status = ? AND employees.is_active = ?", date.Year(), int(date.Month()), "completed", true).
			Scan(&avgResult)
		trend.AvgScore = avgResult.AvgScore

		// 完成率（只统计在职员工）
		var completed int64
		models.DB.Model(&models.KPIEvaluation{}).
			Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
			Where("kpi_evaluations.year = ? AND kpi_evaluations.month = ? AND kpi_evaluations.status = ? AND employees.is_active = ?", date.Year(), int(date.Month()), "completed", true).
			Count(&completed)

		if trend.Evaluations > 0 {
			trend.CompletionRate = float64(completed) / float64(trend.Evaluations) * 100
		}

		response.MonthlyTrends = append(response.MonthlyTrends, trend)
	}

	// 3. 获取分数分布
	scoreRanges := []struct {
		min   float64
		max   float64
		label string
		color string
	}{
		{90, 100, "90-100", "#22c55e"},
		{80, 89, "80-89", "#3b82f6"},
		{70, 79, "70-79", "#f59e0b"},
		{60, 69, "60-69", "#ef4444"},
		{0, 59, "60以下", "#6b7280"},
	}

	for _, scoreRange := range scoreRanges {
		var count int64
		query := models.DB.Model(&models.KPIEvaluation{}).
			Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
			Where("kpi_evaluations.status = ? AND kpi_evaluations.total_score >= ? AND kpi_evaluations.total_score <= ? AND employees.is_active = ?", "completed", scoreRange.min, scoreRange.max, true)

		switch period {
		case "monthly":
			query = query.Where("kpi_evaluations.period = ? AND kpi_evaluations.year = ? AND kpi_evaluations.month = ?", "monthly", year, month)
		case "quarterly":
			query = query.Where("kpi_evaluations.period = ? AND kpi_evaluations.year = ? AND kpi_evaluations.quarter = ?", "quarterly", year, quarter)
		default:
			// 兼容历史数据格式，支持 period="yearly" 和 period="年份"
			query = query.Where("(kpi_evaluations.period = ? OR kpi_evaluations.period = ?) AND kpi_evaluations.year = ?", "yearly", year, year)
		}

		query.Count(&count)

		response.ScoreDistribution = append(response.ScoreDistribution, ScoreDistrib{
			Range: scoreRange.label,
			Count: count,
			Color: scoreRange.color,
		})
	}

	// 4. 获取顶级表现者
	var topPerformers []struct {
		EmployeeID   uint
		EmployeeName string
		DeptName     string
		AvgScore     float64
		EvalCount    int64
	}

	query := models.DB.Model(&models.KPIEvaluation{}).
		Select("employees.id as employee_id, employees.name as employee_name, departments.name as dept_name, AVG(kpi_evaluations.total_score) as avg_score, COUNT(*) as eval_count").
		Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
		Joins("JOIN departments ON employees.department_id = departments.id").
		Where("kpi_evaluations.status = ? AND employees.is_active = ?", "completed", true)

	switch period {
	case "monthly":
		query = query.Where("kpi_evaluations.period = ? AND kpi_evaluations.year = ? AND kpi_evaluations.month = ?", "monthly", year, month)
	case "quarterly":
		query = query.Where("kpi_evaluations.period = ? AND kpi_evaluations.year = ? AND kpi_evaluations.quarter = ?", "quarterly", year, quarter)
	default:
		// 兼容历史数据格式，支持 period="yearly" 和 period="年份"
		query = query.Where("(kpi_evaluations.period = ? OR kpi_evaluations.period = ?) AND kpi_evaluations.year = ?", "yearly", year, year)
	}

	query.Group("employees.id, employees.name, departments.name").
		Order("avg_score DESC").
		Limit(5).
		Scan(&topPerformers)

	for _, performer := range topPerformers {
		response.TopPerformers = append(response.TopPerformers, TopPerformer{
			Name:        performer.EmployeeName,
			Department:  performer.DeptName,
			Score:       performer.AvgScore,
			Evaluations: performer.EvalCount,
		})
	}

	// 5. 获取最近评估记录（只显示在职员工的评估）
	var recentEvals []models.KPIEvaluation
	models.DB.Preload("Employee.Department").Preload("Template").
		Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
		Where("employees.is_active = ?", true).
		Order("kpi_evaluations.created_at DESC").
		Limit(10).
		Find(&recentEvals)

	for _, eval := range recentEvals {
		response.RecentEvaluations = append(response.RecentEvaluations, RecentEvaluation{
			ID:         eval.ID,
			Employee:   eval.Employee.Name,
			Department: eval.Employee.Department.Name,
			Template:   eval.Template.Name,
			Score:      eval.TotalScore,
			Status:     eval.Status,
			Period:     eval.Period,
			Year:       eval.Year,
			Month:      eval.Month,
			Quarter:    eval.Quarter,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}
