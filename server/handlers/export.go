package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"dootask-kpi-server/global"
	"dootask-kpi-server/models"
	"dootask-kpi-server/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

const (
	// 导出文件目录
	ExportDir = "./public/exports"
)

// 格式化周期显示
func formatPeriodDisplay(period string, year int, month *int, quarter *int) string {
	switch period {
	case "yearly":
		return fmt.Sprintf("年度 %d", year)
	case "quarterly":
		if quarter != nil {
			return fmt.Sprintf("季度 %d年第%d季度", year, *quarter)
		}
		return fmt.Sprintf("年度 %d", year)
	case "monthly":
		if month != nil {
			return fmt.Sprintf("月度 %d年%d月", year, *month)
		}
		return fmt.Sprintf("年度 %d", year)
	default:
		// 兼容历史数据格式，period可能是年份字符串
		if yearInt, err := strconv.Atoi(period); err == nil {
			return fmt.Sprintf("年度 %d", yearInt)
		}
		return fmt.Sprintf("年度 %d", year)
	}
}

// 导出响应结构
type ExportResponse struct {
	FileURL  string `json:"file_url"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	Message  string `json:"message"`
}

// 导出评估报告为Excel
func ExportEvaluationToExcel(c *gin.Context) {
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

	// 创建Excel文件
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheetName := "评估报告"
	f.SetSheetName("Sheet1", sheetName)

	// 设置标题
	f.SetCellValue(sheetName, "A1", "绩效考核评估报告")
	f.MergeCell(sheetName, "A1", "G1")

	// 设置标题样式
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 16,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	f.SetCellStyle(sheetName, "A1", "G1", titleStyle)

	// 设置基本信息
	row := 3
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "员工姓名:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), evaluation.Employee.Name)
	f.SetCellValue(sheetName, "D"+strconv.Itoa(row), "部门:")
	f.SetCellValue(sheetName, "E"+strconv.Itoa(row), evaluation.Employee.Department.Name)

	row++
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "考核模板:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), evaluation.Template.Name)
	f.SetCellValue(sheetName, "D"+strconv.Itoa(row), "考核周期:")
	// 使用格式化的周期显示
	periodDisplay := formatPeriodDisplay(evaluation.Period, evaluation.Year, evaluation.Month, evaluation.Quarter)
	f.SetCellValue(sheetName, "E"+strconv.Itoa(row), periodDisplay)

	row++
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "总分:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), evaluation.TotalScore)
	f.SetCellValue(sheetName, "D"+strconv.Itoa(row), "状态:")
	f.SetCellValue(sheetName, "E"+strconv.Itoa(row), getStatusText(evaluation.Status))

	// 设置表头
	row += 2
	headers := []string{"考核项目", "满分", "自评分", "自评说明", "主管评分", "主管说明", "最终得分"}
	for i, header := range headers {
		cell := string(rune('A'+i)) + strconv.Itoa(row)
		f.SetCellValue(sheetName, cell, header)
	}

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E6E6FA"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	for i := 0; i < len(headers); i++ {
		cell := string(rune('A'+i)) + strconv.Itoa(row)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 设置数据
	for _, score := range evaluation.Scores {
		row++
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), score.Item.Name)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(row), score.Item.MaxScore)

		if score.SelfScore != nil {
			f.SetCellValue(sheetName, "C"+strconv.Itoa(row), *score.SelfScore)
		}
		f.SetCellValue(sheetName, "D"+strconv.Itoa(row), score.SelfComment)

		if score.ManagerScore != nil {
			f.SetCellValue(sheetName, "E"+strconv.Itoa(row), *score.ManagerScore)
		}
		f.SetCellValue(sheetName, "F"+strconv.Itoa(row), score.ManagerComment)

		if score.FinalScore != nil {
			f.SetCellValue(sheetName, "G"+strconv.Itoa(row), *score.FinalScore)
		}

		// 设置数据行样式
		dataStyle, _ := f.NewStyle(&excelize.Style{
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
			},
		})
		for i := 0; i < len(headers); i++ {
			cell := string(rune('A'+i)) + strconv.Itoa(row)
			f.SetCellStyle(sheetName, cell, cell, dataStyle)
		}
	}

	// 添加总结
	if evaluation.FinalComment != "" {
		row += 2
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "总结评价:")
		f.MergeCell(sheetName, "A"+strconv.Itoa(row), "G"+strconv.Itoa(row))
		row++
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), evaluation.FinalComment)
		f.MergeCell(sheetName, "A"+strconv.Itoa(row), "G"+strconv.Itoa(row+2))
	}

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 20)
	f.SetColWidth(sheetName, "B", "B", 10)
	f.SetColWidth(sheetName, "C", "C", 10)
	f.SetColWidth(sheetName, "D", "D", 25)
	f.SetColWidth(sheetName, "E", "E", 10)
	f.SetColWidth(sheetName, "F", "F", 25)
	f.SetColWidth(sheetName, "G", "G", 10)

	// 创建公共导出目录
	if err := os.MkdirAll(ExportDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建导出目录失败",
		})
		return
	}

	// 生成文件名
	periodForFileName := formatPeriodDisplay(evaluation.Period, evaluation.Year, evaluation.Month, evaluation.Quarter)
	fileName := fmt.Sprintf("评估报告-%s-%s-%d.xlsx",
		evaluation.Employee.Name,
		periodForFileName,
		time.Now().Unix())
	filePath := filepath.Join(ExportDir, fileName)

	// 保存文件
	if err := f.SaveAs(filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "保存文件失败: " + err.Error(),
		})
		return
	}

	// 获取文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取文件信息失败",
		})
		return
	}

	// 缓存到内存
	randomKey := "export_" + uuid.New().String()
	global.Cache.Set(randomKey, fileName, time.Minute*5)

	// 返回下载URL
	downloadURL := utils.GetFileURL(c.GetString("base_url"), fmt.Sprintf("/api/download/exports/%s", randomKey))

	c.JSON(http.StatusOK, ExportResponse{
		FileURL:  downloadURL,
		FileName: fileName,
		FileSize: fileInfo.Size(),
		Message:  "导出成功",
	})

	// 30分钟后删除文件
	go func() {
		time.Sleep(time.Minute * 30)
		os.Remove(filePath)
	}()
}

// 导出部门评估汇总
func ExportDepartmentToExcel(c *gin.Context) {
	id := c.Param("id")
	departmentId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的部门ID",
		})
		return
	}

	var department models.Department
	if err := models.DB.First(&department, departmentId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "部门不存在",
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

	// 创建Excel文件
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheetName := "部门评估汇总"
	f.SetSheetName("Sheet1", sheetName)

	// 设置标题
	f.SetCellValue(sheetName, "A1", fmt.Sprintf("%s 评估汇总报告", department.Name))
	f.MergeCell(sheetName, "A1", "H1")

	// 设置标题样式
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 16,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	f.SetCellStyle(sheetName, "A1", "H1", titleStyle)

	// 设置表头
	row := 3
	headers := []string{"序号", "员工姓名", "考核模板", "考核周期", "总分", "状态", "创建时间", "最后更新"}
	for i, header := range headers {
		cell := string(rune('A'+i)) + strconv.Itoa(row)
		f.SetCellValue(sheetName, cell, header)
	}

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E6E6FA"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	for i := 0; i < len(headers); i++ {
		cell := string(rune('A'+i)) + strconv.Itoa(row)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 设置数据
	for idx, evaluation := range evaluations {
		row++
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), idx+1)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(row), evaluation.Employee.Name)
		f.SetCellValue(sheetName, "C"+strconv.Itoa(row), evaluation.Template.Name)
		f.SetCellValue(sheetName, "D"+strconv.Itoa(row), evaluation.Period)
		f.SetCellValue(sheetName, "E"+strconv.Itoa(row), evaluation.TotalScore)
		f.SetCellValue(sheetName, "F"+strconv.Itoa(row), getStatusText(evaluation.Status))
		f.SetCellValue(sheetName, "G"+strconv.Itoa(row), evaluation.CreatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheetName, "H"+strconv.Itoa(row), evaluation.UpdatedAt.Format("2006-01-02 15:04:05"))

		// 设置数据行样式
		dataStyle, _ := f.NewStyle(&excelize.Style{
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
			},
		})
		for i := 0; i < len(headers); i++ {
			cell := string(rune('A'+i)) + strconv.Itoa(row)
			f.SetCellStyle(sheetName, cell, cell, dataStyle)
		}
	}

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 8)
	f.SetColWidth(sheetName, "B", "B", 15)
	f.SetColWidth(sheetName, "C", "C", 25)
	f.SetColWidth(sheetName, "D", "D", 15)
	f.SetColWidth(sheetName, "E", "E", 10)
	f.SetColWidth(sheetName, "F", "F", 15)
	f.SetColWidth(sheetName, "G", "G", 20)
	f.SetColWidth(sheetName, "H", "H", 20)

	// 创建公共导出目录
	if err := os.MkdirAll(ExportDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建导出目录失败",
		})
		return
	}

	// 生成文件名
	fileName := fmt.Sprintf("部门评估汇总-%s-%d.xlsx",
		department.Name,
		time.Now().Unix())
	filePath := filepath.Join(ExportDir, fileName)

	// 保存文件
	if err := f.SaveAs(filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "保存文件失败: " + err.Error(),
		})
		return
	}

	// 获取文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取文件信息失败",
		})
		return
	}

	// 缓存到内存
	randomKey := "export_" + uuid.New().String()
	global.Cache.Set(randomKey, fileName, time.Minute*5)

	// 返回下载URL
	downloadURL := utils.GetFileURL(c.GetString("base_url"), fmt.Sprintf("/api/download/exports/%s", randomKey))

	c.JSON(http.StatusOK, ExportResponse{
		FileURL:  downloadURL,
		FileName: fileName,
		FileSize: fileInfo.Size(),
		Message:  "导出成功",
	})

	// 30分钟后删除文件
	go func() {
		time.Sleep(time.Minute * 30)
		os.Remove(filePath)
	}()
}

// 导出周期评估统计
func ExportPeriodToExcel(c *gin.Context) {
	period := c.Param("period")
	year := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	month := c.DefaultQuery("month", "")
	quarter := c.DefaultQuery("quarter", "")

	var evaluations []models.KPIEvaluation
	query := models.DB.Preload("Employee.Department").Preload("Template").Preload("Scores.Item")

	// 根据周期类型筛选（与统计页面查询逻辑保持一致）
	if period == "monthly" && month != "" {
		query = query.Where("period = ? AND year = ? AND month = ?", "monthly", year, month)
	} else if period == "quarterly" && quarter != "" {
		query = query.Where("period = ? AND year = ? AND quarter = ?", "quarterly", year, quarter)
	} else if period == "yearly" {
		// 兼容历史数据格式，支持 period="yearly" 和 period="年份"
		query = query.Where("(period = ? OR period = ?) AND year = ?", "yearly", year, year)
	}

	result := query.Find(&evaluations)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取周期评估数据失败",
		})
		return
	}

	// 获取所有评估的邀请评分数据
	var evaluationIDs []uint
	for _, evaluation := range evaluations {
		evaluationIDs = append(evaluationIDs, evaluation.ID)
	}

	var invitations []models.EvaluationInvitation
	if len(evaluationIDs) > 0 {
		models.DB.Preload("Inviter").Preload("Invitee").Preload("Scores.Item").
			Where("evaluation_id IN ?", evaluationIDs).
			Find(&invitations)
	}

	// 创建Excel文件
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// 构建标题和文件名标识
	var title, fileNamePeriod string
	switch period {
	case "monthly":
		title = fmt.Sprintf("%s年%s月 综合评估报告", year, month)
		fileNamePeriod = fmt.Sprintf("%s年%s月", year, month)
	case "quarterly":
		title = fmt.Sprintf("%s年第%s季度 综合评估报告", year, quarter)
		fileNamePeriod = fmt.Sprintf("%s年Q%s", year, quarter)
	case "yearly":
		title = fmt.Sprintf("%s年度 综合评估报告", year)
		fileNamePeriod = fmt.Sprintf("%s年度", year)
	default:
		title = "综合评估报告"
		fileNamePeriod = year
	}

	// 第一个工作表：总览表
	overviewSheetName := "总览表"
	f.SetSheetName("Sheet1", overviewSheetName)
	createOverviewSheet(f, overviewSheetName, title, evaluations)

	// 为每个员工创建详细工作表
	for idx, evaluation := range evaluations {
		// 获取该员工的邀请评分
		var evalInvitations []models.EvaluationInvitation
		for _, invitation := range invitations {
			if invitation.EvaluationID == evaluation.ID {
				evalInvitations = append(evalInvitations, invitation)
			}
		}

		sheetName := evaluation.Employee.Name
		// Excel工作表名称有字符限制，如果名称过长或有特殊字符，使用简化名称
		if len(sheetName) > 31 {
			sheetName = fmt.Sprintf("员工%d", idx+1)
		}

		// 创建新的工作表
		f.NewSheet(sheetName)
		createDetailSheet(f, sheetName, evaluation, evalInvitations)
	}

	// 设置活动工作表为总览表
	f.SetActiveSheet(0)

	// 创建公共导出目录
	if err := os.MkdirAll(ExportDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建导出目录失败",
		})
		return
	}

	// 生成文件名
	fileName := fmt.Sprintf("综合评估报告-%s-%d.xlsx",
		fileNamePeriod,
		time.Now().Unix())
	filePath := filepath.Join(ExportDir, fileName)

	// 保存文件
	if err := f.SaveAs(filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "保存文件失败: " + err.Error(),
		})
		return
	}

	// 获取文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取文件信息失败",
		})
		return
	}

	// 缓存到内存
	randomKey := "export_" + uuid.New().String()
	global.Cache.Set(randomKey, fileName, time.Minute*5)

	// 返回下载URL
	downloadURL := utils.GetFileURL(c.GetString("base_url"), fmt.Sprintf("/api/download/exports/%s", randomKey))

	c.JSON(http.StatusOK, ExportResponse{
		FileURL:  downloadURL,
		FileName: fileName,
		FileSize: fileInfo.Size(),
		Message:  "导出成功",
	})

	// 30分钟后删除文件
	go func() {
		time.Sleep(time.Minute * 30)
		os.Remove(filePath)
	}()
}

// 文件下载处理
func DownloadFile(c *gin.Context) {
	randomKey := c.Param("randomKey")

	fileName, ok := global.Cache.Get(randomKey)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "文件不存在",
		})
		return
	}
	filePath := filepath.Join(ExportDir, fileName.(string))

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "文件不存在",
		})
		return
	}

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+fileName.(string))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	// 返回文件
	c.File(filePath)
}

// 创建总览表工作表
func createOverviewSheet(f *excelize.File, sheetName, title string, evaluations []models.KPIEvaluation) {
	// 设置标题
	f.SetCellValue(sheetName, "A1", title)
	f.MergeCell(sheetName, "A1", "J1")

	// 设置标题样式
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 16,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	f.SetCellStyle(sheetName, "A1", "J1", titleStyle)

	// 设置表头
	row := 3
	headers := []string{"序号", "员工姓名", "部门", "考核模板", "考核周期", "员工自评", "主管评分", "邀请评分", "最终得分", "状态"}
	for i, header := range headers {
		cell := string(rune('A'+i)) + strconv.Itoa(row)
		f.SetCellValue(sheetName, cell, header)
	}

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E6E6FA"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	for i := 0; i < len(headers); i++ {
		cell := string(rune('A'+i)) + strconv.Itoa(row)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 设置数据
	for idx, evaluation := range evaluations {
		row++
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), idx+1)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(row), evaluation.Employee.Name)
		f.SetCellValue(sheetName, "C"+strconv.Itoa(row), evaluation.Employee.Department.Name)
		f.SetCellValue(sheetName, "D"+strconv.Itoa(row), evaluation.Template.Name)

		// 使用格式化的周期显示
		periodDisplay := formatPeriodDisplay(evaluation.Period, evaluation.Year, evaluation.Month, evaluation.Quarter)
		f.SetCellValue(sheetName, "E"+strconv.Itoa(row), periodDisplay)

		// 计算员工自评和主管评分
		var selfTotalScore, managerTotalScore float64
		for _, score := range evaluation.Scores {
			if score.SelfScore != nil {
				selfTotalScore += *score.SelfScore
			}
			if score.ManagerScore != nil {
				managerTotalScore += *score.ManagerScore
			}
		}

		f.SetCellValue(sheetName, "F"+strconv.Itoa(row), fmt.Sprintf("%.2f", selfTotalScore))
		f.SetCellValue(sheetName, "G"+strconv.Itoa(row), fmt.Sprintf("%.2f", managerTotalScore))

		// 计算邀请评分列表
		var invitations []models.EvaluationInvitation
		models.DB.Preload("Scores").Where("evaluation_id = ? AND status = ?", evaluation.ID, "completed").Find(&invitations)

		var invitationScores []string
		for _, invitation := range invitations {
			var invitationTotal float64
			for _, invScore := range invitation.Scores {
				if invScore.Score != nil {
					invitationTotal += *invScore.Score
				}
			}
			if invitationTotal > 0 {
				invitationScores = append(invitationScores, fmt.Sprintf("%.0f", invitationTotal))
			}
		}

		var invitationScoreText string
		if len(invitationScores) > 0 {
			invitationScoreText = invitationScores[0]
			for i := 1; i < len(invitationScores); i++ {
				invitationScoreText += "、" + invitationScores[i]
			}
		}

		f.SetCellValue(sheetName, "H"+strconv.Itoa(row), invitationScoreText)
		f.SetCellValue(sheetName, "I"+strconv.Itoa(row), evaluation.TotalScore)
		f.SetCellValue(sheetName, "J"+strconv.Itoa(row), getStatusText(evaluation.Status))

		// 设置数据行样式
		dataStyle, _ := f.NewStyle(&excelize.Style{
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
			},
		})
		for i := 0; i < len(headers); i++ {
			cell := string(rune('A'+i)) + strconv.Itoa(row)
			f.SetCellStyle(sheetName, cell, cell, dataStyle)
		}
	}

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 8)
	f.SetColWidth(sheetName, "B", "B", 15)
	f.SetColWidth(sheetName, "C", "C", 15)
	f.SetColWidth(sheetName, "D", "D", 25)
	f.SetColWidth(sheetName, "E", "E", 20)
	f.SetColWidth(sheetName, "F", "F", 12)
	f.SetColWidth(sheetName, "G", "G", 12)
	f.SetColWidth(sheetName, "H", "H", 18)
	f.SetColWidth(sheetName, "I", "I", 12)
	f.SetColWidth(sheetName, "J", "J", 15)
}

// 创建详细工作表
func createDetailSheet(f *excelize.File, sheetName string, evaluation models.KPIEvaluation, invitations []models.EvaluationInvitation) {
	currentRow := 1

	// 设置标题
	title := fmt.Sprintf("%s 详细评估报告", evaluation.Employee.Name)
	f.SetCellValue(sheetName, "A"+strconv.Itoa(currentRow), title)
	f.MergeCell(sheetName, "A"+strconv.Itoa(currentRow), "I"+strconv.Itoa(currentRow))

	// 设置标题样式
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 16,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	f.SetCellStyle(sheetName, "A"+strconv.Itoa(currentRow), "I"+strconv.Itoa(currentRow), titleStyle)
	currentRow += 2

	// 基本信息
	f.SetCellValue(sheetName, "A"+strconv.Itoa(currentRow), "员工姓名:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(currentRow), evaluation.Employee.Name)
	f.SetCellValue(sheetName, "D"+strconv.Itoa(currentRow), "部门:")
	f.SetCellValue(sheetName, "E"+strconv.Itoa(currentRow), evaluation.Employee.Department.Name)
	currentRow++

	f.SetCellValue(sheetName, "A"+strconv.Itoa(currentRow), "考核模板:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(currentRow), evaluation.Template.Name)
	f.SetCellValue(sheetName, "D"+strconv.Itoa(currentRow), "考核周期:")
	periodDisplay := formatPeriodDisplay(evaluation.Period, evaluation.Year, evaluation.Month, evaluation.Quarter)
	f.SetCellValue(sheetName, "E"+strconv.Itoa(currentRow), periodDisplay)
	currentRow++

	f.SetCellValue(sheetName, "A"+strconv.Itoa(currentRow), "考核状态:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(currentRow), getStatusText(evaluation.Status))
	f.SetCellValue(sheetName, "D"+strconv.Itoa(currentRow), "最终得分:")
	f.SetCellValue(sheetName, "E"+strconv.Itoa(currentRow), evaluation.TotalScore)
	currentRow += 2

	// 考核指标详细表
	f.SetCellValue(sheetName, "A"+strconv.Itoa(currentRow), "考核指标详细")
	f.MergeCell(sheetName, "A"+strconv.Itoa(currentRow), "I"+strconv.Itoa(currentRow))
	sectionStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 14,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#D3D3D3"},
			Pattern: 1,
		},
	})
	f.SetCellStyle(sheetName, "A"+strconv.Itoa(currentRow), "I"+strconv.Itoa(currentRow), sectionStyle)
	currentRow++

	// 考核指标表头
	headers := []string{"考核项目", "满分", "自评分", "自评说明", "主管评分", "主管说明", "HR评分", "HR说明", "最终得分"}
	for i, header := range headers {
		cell := string(rune('A'+i)) + strconv.Itoa(currentRow)
		f.SetCellValue(sheetName, cell, header)
	}

	// 设置表头样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E6E6FA"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	for i := 0; i < len(headers); i++ {
		cell := string(rune('A'+i)) + strconv.Itoa(currentRow)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}
	currentRow++

	// 考核指标数据
	var totalMaxScore, totalSelfScore, totalManagerScore, totalHRScore, totalFinalScore float64
	for _, score := range evaluation.Scores {
		f.SetCellValue(sheetName, "A"+strconv.Itoa(currentRow), score.Item.Name)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(currentRow), score.Item.MaxScore)
		totalMaxScore += score.Item.MaxScore

		if score.SelfScore != nil {
			f.SetCellValue(sheetName, "C"+strconv.Itoa(currentRow), *score.SelfScore)
			totalSelfScore += *score.SelfScore
		}
		f.SetCellValue(sheetName, "D"+strconv.Itoa(currentRow), score.SelfComment)

		if score.ManagerScore != nil {
			f.SetCellValue(sheetName, "E"+strconv.Itoa(currentRow), *score.ManagerScore)
			totalManagerScore += *score.ManagerScore
		}
		f.SetCellValue(sheetName, "F"+strconv.Itoa(currentRow), score.ManagerComment)

		if score.HRScore != nil {
			f.SetCellValue(sheetName, "G"+strconv.Itoa(currentRow), *score.HRScore)
			totalHRScore += *score.HRScore
		}
		f.SetCellValue(sheetName, "H"+strconv.Itoa(currentRow), score.HRComment)

		if score.FinalScore != nil {
			f.SetCellValue(sheetName, "I"+strconv.Itoa(currentRow), *score.FinalScore)
			totalFinalScore += *score.FinalScore
		}

		// 设置数据行样式
		dataStyle, _ := f.NewStyle(&excelize.Style{
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
			},
		})
		for i := 0; i < len(headers); i++ {
			cell := string(rune('A'+i)) + strconv.Itoa(currentRow)
			f.SetCellStyle(sheetName, cell, cell, dataStyle)
		}
		currentRow++
	}

	// 添加总分行
	f.SetCellValue(sheetName, "A"+strconv.Itoa(currentRow), "总分")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(currentRow), totalMaxScore)
	if totalSelfScore > 0 {
		f.SetCellValue(sheetName, "C"+strconv.Itoa(currentRow), totalSelfScore)
	}
	f.SetCellValue(sheetName, "D"+strconv.Itoa(currentRow), "")
	if totalManagerScore > 0 {
		f.SetCellValue(sheetName, "E"+strconv.Itoa(currentRow), totalManagerScore)
	}
	f.SetCellValue(sheetName, "F"+strconv.Itoa(currentRow), "")
	if totalHRScore > 0 {
		f.SetCellValue(sheetName, "G"+strconv.Itoa(currentRow), totalHRScore)
	}
	f.SetCellValue(sheetName, "H"+strconv.Itoa(currentRow), "")
	if totalFinalScore > 0 {
		f.SetCellValue(sheetName, "I"+strconv.Itoa(currentRow), totalFinalScore)
	}

	// 设置总分行样式（加粗）
	totalStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#F0F0F0"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	for i := 0; i < len(headers); i++ {
		cell := string(rune('A'+i)) + strconv.Itoa(currentRow)
		f.SetCellStyle(sheetName, cell, cell, totalStyle)
	}
	currentRow++

	// 邀请评分部分
	if len(invitations) > 0 {
		currentRow += 2
		f.SetCellValue(sheetName, "A"+strconv.Itoa(currentRow), "邀请评分详细")
		f.MergeCell(sheetName, "A"+strconv.Itoa(currentRow), "I"+strconv.Itoa(currentRow))
		f.SetCellStyle(sheetName, "A"+strconv.Itoa(currentRow), "I"+strconv.Itoa(currentRow), sectionStyle)
		currentRow++

		for invIdx, invitation := range invitations {
			// 邀请人信息
			f.SetCellValue(sheetName, "A"+strconv.Itoa(currentRow), fmt.Sprintf("邀请人 %d: %s", invIdx+1, invitation.Invitee.Name))
			f.SetCellValue(sheetName, "C"+strconv.Itoa(currentRow), "状态:")
			f.SetCellValue(sheetName, "D"+strconv.Itoa(currentRow), getInvitationStatusText(invitation.Status))
			currentRow++

			// 邀请评分表头
			invHeaders := []string{"考核项目", "满分", "邀请评分", "评价说明"}
			for i, header := range invHeaders {
				cell := string(rune('A'+i)) + strconv.Itoa(currentRow)
				f.SetCellValue(sheetName, cell, header)
			}

			// 设置表头样式
			for i := 0; i < len(invHeaders); i++ {
				cell := string(rune('A'+i)) + strconv.Itoa(currentRow)
				f.SetCellStyle(sheetName, cell, cell, headerStyle)
			}
			currentRow++

			// 邀请评分数据
			var invTotalMaxScore, invTotalScore float64
			for _, invScore := range invitation.Scores {
				f.SetCellValue(sheetName, "A"+strconv.Itoa(currentRow), invScore.Item.Name)
				f.SetCellValue(sheetName, "B"+strconv.Itoa(currentRow), invScore.Item.MaxScore)
				invTotalMaxScore += invScore.Item.MaxScore

				if invScore.Score != nil {
					f.SetCellValue(sheetName, "C"+strconv.Itoa(currentRow), *invScore.Score)
					invTotalScore += *invScore.Score
				}
				f.SetCellValue(sheetName, "D"+strconv.Itoa(currentRow), invScore.Comment)

				// 设置数据行样式
				dataStyle, _ := f.NewStyle(&excelize.Style{
					Border: []excelize.Border{
						{Type: "left", Color: "000000", Style: 1},
						{Type: "top", Color: "000000", Style: 1},
						{Type: "bottom", Color: "000000", Style: 1},
						{Type: "right", Color: "000000", Style: 1},
					},
				})
				for i := 0; i < len(invHeaders); i++ {
					cell := string(rune('A'+i)) + strconv.Itoa(currentRow)
					f.SetCellStyle(sheetName, cell, cell, dataStyle)
				}
				currentRow++
			}

			// 添加邀请评分总分行
			f.SetCellValue(sheetName, "A"+strconv.Itoa(currentRow), "总分")
			f.SetCellValue(sheetName, "B"+strconv.Itoa(currentRow), invTotalMaxScore)
			if invTotalScore > 0 {
				f.SetCellValue(sheetName, "C"+strconv.Itoa(currentRow), invTotalScore)
			}
			f.SetCellValue(sheetName, "D"+strconv.Itoa(currentRow), "")

			// 设置邀请评分总分行样式
			invTotalStyle, _ := f.NewStyle(&excelize.Style{
				Font: &excelize.Font{
					Bold: true,
				},
				Fill: excelize.Fill{
					Type:    "pattern",
					Color:   []string{"#F0F0F0"},
					Pattern: 1,
				},
				Border: []excelize.Border{
					{Type: "left", Color: "000000", Style: 1},
					{Type: "top", Color: "000000", Style: 2},
					{Type: "bottom", Color: "000000", Style: 1},
					{Type: "right", Color: "000000", Style: 1},
				},
			})
			for i := 0; i < len(invHeaders); i++ {
				cell := string(rune('A'+i)) + strconv.Itoa(currentRow)
				f.SetCellStyle(sheetName, cell, cell, invTotalStyle)
			}
			currentRow += 2
		}
	}

	// 总结评价
	if evaluation.FinalComment != "" {
		currentRow += 1
		f.SetCellValue(sheetName, "A"+strconv.Itoa(currentRow), "总结评价:")
		f.MergeCell(sheetName, "A"+strconv.Itoa(currentRow), "I"+strconv.Itoa(currentRow))
		f.SetCellStyle(sheetName, "A"+strconv.Itoa(currentRow), "I"+strconv.Itoa(currentRow), sectionStyle)
		currentRow++
		f.SetCellValue(sheetName, "A"+strconv.Itoa(currentRow), evaluation.FinalComment)
		f.MergeCell(sheetName, "A"+strconv.Itoa(currentRow), "I"+strconv.Itoa(currentRow+2))
	}

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 20)
	f.SetColWidth(sheetName, "B", "B", 10)
	f.SetColWidth(sheetName, "C", "C", 10)
	f.SetColWidth(sheetName, "D", "D", 25)
	f.SetColWidth(sheetName, "E", "E", 10)
	f.SetColWidth(sheetName, "F", "F", 25)
	f.SetColWidth(sheetName, "G", "G", 10)
	f.SetColWidth(sheetName, "H", "H", 25)
	f.SetColWidth(sheetName, "I", "I", 10)
}

// 获取邀请状态文本
func getInvitationStatusText(status string) string {
	switch status {
	case "pending":
		return "待评分"
	case "accepted":
		return "已接受"
	case "declined":
		return "已拒绝"
	case "completed":
		return "已完成"
	default:
		return "未知状态"
	}
}

// 获取状态文本
func getStatusText(status string) string {
	switch status {
	case "pending":
		return "待自评"
	case "self_evaluated":
		return "待主管评估"
	case "manager_evaluated":
		return "待HR审核"
	case "pending_confirm":
		return "待确认"
	case "completed":
		return "已完成"
	default:
		return "未知状态"
	}
}

// 定期清理导出文件
func CleanupExportFiles() {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			// 清理导出文件
			files, err := os.ReadDir(ExportDir)
			if err != nil {
				continue
			}

			cutoffTime := time.Now().Add(-1 * time.Hour)

			for _, file := range files {
				filePath := filepath.Join(ExportDir, file.Name())
				if info, err := file.Info(); err == nil {
					if info.ModTime().Before(cutoffTime) {
						os.Remove(filePath)
					}
				}
			}
		}
	}()
}
