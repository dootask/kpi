package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"dootask-kpi-server/models"
	"dootask-kpi-server/utils"

	"github.com/gin-gonic/gin"
)

// 邀请评分相关API

// 创建邀请请求结构
type CreateInvitationRequest struct {
	InviteeIDs []uint `json:"invitee_ids" binding:"required"`
	Message    string `json:"message"`
}

// 创建邀请
func CreateInvitation(c *gin.Context) {
	evaluationID := c.Param("id")
	evalID, err := strconv.ParseUint(evaluationID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的评估ID"})
		return
	}

	// 获取当前用户ID（邀请者）
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}
	inviterID := currentUserID.(uint)

	// 验证当前用户是否是HR
	var currentUser models.Employee
	if err := models.DB.First(&currentUser, inviterID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	if currentUser.Role != "hr" {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有HR可以发起邀请"})
		return
	}

	// 验证评估是否存在且状态为self_evaluated或manager_evaluated
	var evaluation models.KPIEvaluation
	if err := models.DB.Preload("Template").Preload("Employee").First(&evaluation, evalID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "评估不存在"})
		return
	}

	// 允许在自评完成后或主管评估完成后发起邀请
	if evaluation.Status != "self_evaluated" && evaluation.Status != "manager_evaluated" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只能在员工自评完成后或主管评估完成后发起邀请"})
		return
	}

	var req CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 禁止邀请被评估员工本人
	for _, inviteeID := range req.InviteeIDs {
		if inviteeID == evaluation.EmployeeID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "不能邀请被评估员工本人进行评分"})
			return
		}
	}

	// 获取评估的KPI项目
	var items []models.KPIItem
	if err := models.DB.Where("template_id = ?", evaluation.TemplateID).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取评估项目失败"})
		return
	}

	// 开始数据库事务
	tx := models.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var createdInvitations []models.EvaluationInvitation

	// 为每个被邀请人创建邀请
	for _, inviteeID := range req.InviteeIDs {
		// 检查是否已经邀请过这个人
		var existingInvitation models.EvaluationInvitation
		if err := tx.Where("evaluation_id = ? AND invitee_id = ?", evalID, inviteeID).First(&existingInvitation).Error; err == nil {
			// 如果已经邀请过，跳过
			continue
		}

		// 创建邀请记录
		invitation := models.EvaluationInvitation{
			EvaluationID: uint(evalID),
			InviterID:    inviterID,
			InviteeID:    inviteeID,
			Status:       "pending",
			Message:      req.Message,
		}

		if err := tx.Create(&invitation).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建邀请失败"})
			return
		}

		// 为每个KPI项目创建评分记录
		for _, item := range items {
			invitedScore := models.InvitedScore{
				InvitationID: invitation.ID,
				ItemID:       item.ID,
			}
			if err := tx.Create(&invitedScore).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "创建评分记录失败"})
				return
			}
		}

		createdInvitations = append(createdInvitations, invitation)
	}

	tx.Commit()

	// 为每个被邀请人发送 DooTask 机器人通知
	for _, invitation := range createdInvitations {
		// 获取被邀请人信息
		message := req.Message
		if message == "" {
			message = "-"
		}
		var invitee models.Employee
		if err := models.DB.First(&invitee, invitation.InviteeID).Error; err == nil {
			// 发送邀请通知
			dooTaskClient := utils.NewDooTaskClient(c.GetHeader("DooTaskAuth"))
			appConfigJSON := utils.BuildKPIInvitationAppConfig(invitation.ID, evaluation.ID)
			periodValue := utils.GetPeriodValue(evaluation.Period, evaluation.Year, evaluation.Month, evaluation.Quarter)

			if invitee.DooTaskUserID != nil {
				inviteMessage := fmt.Sprintf(
					"**你收到绩效评分邀请，需要完成评分**\n- 被评估员工：%s\n- 考核模板：%s\n- 考核周期：%s\n- 邀请人：%s\n- 邀请消息：%s\n\n> <div class=\"open-micro-app\" data-app-config='%s'>查看详情：点击查看详情</div>",
					evaluation.Employee.Name,
					evaluation.Template.Name,
					periodValue,
					c.GetString("user_name"),
					message,
					appConfigJSON,
				)
				_ = dooTaskClient.SendBotMessage(invitee.DooTaskUserID, inviteMessage)
			}
		}
	}

	// 发送实时通知
	operatorID := c.GetUint("user_id")
	for _, invitation := range createdInvitations {
		GetNotificationService().SendNotification(operatorID, EventInvitationCreated, &invitation)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "邀请创建成功",
		"data":    createdInvitations,
	})
}

// 获取评估的邀请列表
func GetEvaluationInvitations(c *gin.Context) {
	evaluationID := c.Param("id")
	evalID, err := strconv.ParseUint(evaluationID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的评估ID"})
		return
	}

	// 获取当前用户ID
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}
	userID := currentUserID.(uint)

	// 验证当前用户
	var currentUser models.Employee
	if err := models.DB.First(&currentUser, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 验证评估是否存在且评估对象存在
	var evaluation models.KPIEvaluation
	if err := models.DB.Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
		First(&evaluation, evalID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "评估不存在或评估对象已被删除"})
		return
	}

	// 权限检查：HR可以查看所有邀请，被邀请人可以查看自己相关的邀请，被评估员工可以查看自己的评估的邀请
	var invitations []models.EvaluationInvitation
	query := models.DB.Preload("Invitee").Preload("Inviter")

	if currentUser.Role == "hr" {
		// HR可以查看所有邀请
		query = query.Where("evaluation_id = ?", evalID)
	} else if evaluation.EmployeeID == userID {
		// 被评估员工可以查看自己评估的所有邀请
		query = query.Where("evaluation_id = ?", evalID)
	} else {
		// 被邀请人只能查看自己相关的邀请
		query = query.Where("evaluation_id = ? AND invitee_id = ?", evalID, userID)
	}

	if err := query.Find(&invitations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取邀请列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": invitations,
	})
}

// 获取我的邀请列表
func GetMyInvitations(c *gin.Context) {
	// 获取当前用户ID
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}
	userID := currentUserID.(uint)

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := c.Query("status")

	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询条件 - 只查询评估对象存在的邀请，排除已离职员工
	query := models.DB.Model(&models.EvaluationInvitation{}).
		Joins("JOIN kpi_evaluations ON evaluation_invitations.evaluation_id = kpi_evaluations.id").
		Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
		Where("evaluation_invitations.invitee_id = ? AND employees.is_active = ?", userID, true)
	if status != "" && status != "all" {
		query = query.Where("evaluation_invitations.status = ?", status)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取邀请总数失败"})
		return
	}

	// 分页查询
	var invitations []models.EvaluationInvitation
	offset := (page - 1) * pageSize
	queryBuilder := models.DB.Preload("Evaluation.Employee.Department").Preload("Evaluation.Employee").Preload("Evaluation.Template").
		Preload("Inviter").
		Joins("JOIN kpi_evaluations ON evaluation_invitations.evaluation_id = kpi_evaluations.id").
		Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
		Where("evaluation_invitations.invitee_id = ? AND employees.is_active = ?", userID, true)

	if status != "" && status != "all" {
		queryBuilder = queryBuilder.Where("evaluation_invitations.status = ?", status)
	}

	if err := queryBuilder.Order("evaluation_invitations.created_at DESC").Offset(offset).Limit(pageSize).Find(&invitations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取邀请列表失败"})
		return
	}

	// 计算分页信息
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, gin.H{
		"data":       invitations,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
		"hasNext":    page < totalPages,
		"hasPrev":    page > 1,
	})
}

// 获取我发出的邀请列表
func GetMySentInvitations(c *gin.Context) {
	// 获取当前用户ID
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}
	userID := currentUserID.(uint)

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := c.Query("status")

	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询条件 - 只查询评估对象存在的邀请，排除已离职员工
	query := models.DB.Model(&models.EvaluationInvitation{}).
		Joins("JOIN kpi_evaluations ON evaluation_invitations.evaluation_id = kpi_evaluations.id").
		Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
		Where("evaluation_invitations.inviter_id = ? AND employees.is_active = ?", userID, true)
	if status != "" && status != "all" {
		query = query.Where("evaluation_invitations.status = ?", status)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取邀请总数失败"})
		return
	}

	// 分页查询
	var invitations []models.EvaluationInvitation
	offset := (page - 1) * pageSize
	queryBuilder := models.DB.Preload("Evaluation.Employee.Department").Preload("Evaluation.Employee").Preload("Evaluation.Template").
		Preload("Invitee").
		Joins("JOIN kpi_evaluations ON evaluation_invitations.evaluation_id = kpi_evaluations.id").
		Joins("JOIN employees ON kpi_evaluations.employee_id = employees.id").
		Where("evaluation_invitations.inviter_id = ? AND employees.is_active = ?", userID, true)

	if status != "" && status != "all" {
		queryBuilder = queryBuilder.Where("evaluation_invitations.status = ?", status)
	}

	if err := queryBuilder.Order("evaluation_invitations.created_at DESC").Offset(offset).Limit(pageSize).Find(&invitations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取邀请列表失败"})
		return
	}

	// 计算分页信息
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, gin.H{
		"data":       invitations,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
		"hasNext":    page < totalPages,
		"hasPrev":    page > 1,
	})
}

// 接受邀请
func AcceptInvitation(c *gin.Context) {
	invitationID := c.Param("id")
	inviteID, err := strconv.ParseUint(invitationID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的邀请ID"})
		return
	}

	// 获取当前用户ID
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}
	userID := currentUserID.(uint)

	// 验证邀请是否存在且属于当前用户
	var invitation models.EvaluationInvitation
	if err := models.DB.First(&invitation, inviteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "邀请不存在"})
		return
	}

	if invitation.InviteeID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作此邀请"})
		return
	}

	if invitation.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "邀请状态不正确"})
		return
	}

	// 更新邀请状态
	invitation.Status = "accepted"
	if err := models.DB.Save(&invitation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新邀请状态失败"})
		return
	}

	// 发送实时通知
	operatorID := c.GetUint("user_id")
	GetNotificationService().SendNotification(operatorID, EventInvitationStatusChange, &invitation)

	c.JSON(http.StatusOK, gin.H{
		"message": "邀请接受成功",
		"data":    invitation,
	})
}

// 拒绝邀请
func DeclineInvitation(c *gin.Context) {
	invitationID := c.Param("id")
	inviteID, err := strconv.ParseUint(invitationID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的邀请ID"})
		return
	}

	// 获取当前用户ID
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}
	userID := currentUserID.(uint)

	// 验证邀请是否存在且属于当前用户
	var invitation models.EvaluationInvitation
	if err := models.DB.First(&invitation, inviteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "邀请不存在"})
		return
	}

	if invitation.InviteeID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作此邀请"})
		return
	}

	if invitation.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "邀请状态不正确"})
		return
	}

	// 更新邀请状态
	invitation.Status = "declined"
	if err := models.DB.Save(&invitation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新邀请状态失败"})
		return
	}

	// 发送实时通知
	operatorID := c.GetUint("user_id")
	GetNotificationService().SendNotification(operatorID, EventInvitationStatusChange, &invitation)

	// 检查评估状态，如果处于manager_evaluated且启用了绩效规则，检查是否所有邀请都已完成
	var evaluation models.KPIEvaluation
	if err := models.DB.First(&evaluation, invitation.EvaluationID).Error; err == nil {
		if evaluation.Status == "manager_evaluated" {
			// 检查绩效规则是否启用
			var rule models.PerformanceRule
			if err := models.DB.First(&rule).Error; err == nil && rule.Enabled {
				// 检查是否所有邀请都已完成
				allCompleted, err := areAllInvitationsCompleted(invitation.EvaluationID)
				if err == nil && allCompleted {
					// 所有邀请都已完成，自动计算HR评分
					if err := applyPerformanceRuleForEvaluation(invitation.EvaluationID); err == nil {
						// 如果绩效规则应用成功，自动将状态推进到pending_confirm
						models.DB.Model(&evaluation).Update("status", "pending_confirm")
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "邀请拒绝成功",
		"data":    invitation,
	})
}

// 获取邀请的评分记录
func GetInvitationScores(c *gin.Context) {
	invitationID := c.Param("id")
	inviteID, err := strconv.ParseUint(invitationID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的邀请ID"})
		return
	}

	// 获取当前用户ID
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}
	userID := currentUserID.(uint)

	// 验证邀请是否存在
	var invitation models.EvaluationInvitation
	if err := models.DB.Preload("Evaluation").First(&invitation, inviteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "邀请不存在"})
		return
	}

	// 验证权限：被邀请人、被评估员工或HR可以查看
	var currentUser models.Employee
	if err := models.DB.First(&currentUser, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 检查权限：被邀请人、被评估员工或HR可以查看
	canView := invitation.InviteeID == userID || // 被邀请人
		currentUser.Role == "hr" || // HR
		invitation.Evaluation.EmployeeID == userID // 被评估员工

	if !canView {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限查看此邀请的评分"})
		return
	}

	// 获取评分记录
	var scores []models.InvitedScore
	if err := models.DB.Preload("Item").Where("invitation_id = ?", inviteID).Find(&scores).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取评分记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": scores,
	})
}

// 更新邀请评分
func UpdateInvitedScore(c *gin.Context) {
	scoreID := c.Param("id")
	scoreIDUint, err := strconv.ParseUint(scoreID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的评分ID"})
		return
	}

	// 获取当前用户ID
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}
	userID := currentUserID.(uint)

	// 验证评分记录是否存在
	var score models.InvitedScore
	if err := models.DB.Preload("Invitation").First(&score, scoreIDUint).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "评分记录不存在"})
		return
	}

	// 验证权限：只有被邀请人可以更新评分
	if score.Invitation.InviteeID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限更新此评分"})
		return
	}

	// 验证邀请状态
	if score.Invitation.Status != "accepted" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只有接受的邀请才能进行评分"})
		return
	}

	var updateData struct {
		Score   *float64 `json:"score"`
		Comment string   `json:"comment"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新评分
	if err := models.DB.Model(&score).Updates(map[string]interface{}{
		"score":   updateData.Score,
		"comment": updateData.Comment,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新评分失败"})
		return
	}

	// 发送实时通知
	operatorID := c.GetUint("user_id")
	GetNotificationService().SendNotification(operatorID, EventInvitedScoreUpdated, &score)

	c.JSON(http.StatusOK, gin.H{
		"message": "评分更新成功",
		"data":    score,
	})
}

// 完成邀请评分
func CompleteInvitation(c *gin.Context) {
	invitationID := c.Param("id")
	inviteID, err := strconv.ParseUint(invitationID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的邀请ID"})
		return
	}

	// 获取当前用户ID
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}
	userID := currentUserID.(uint)

	// 验证邀请是否存在且属于当前用户
	var invitation models.EvaluationInvitation
	if err := models.DB.First(&invitation, inviteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "邀请不存在"})
		return
	}

	if invitation.InviteeID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作此邀请"})
		return
	}

	if invitation.Status != "accepted" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "邀请状态不正确"})
		return
	}

	// 检查是否所有项目都已评分
	var totalScores int64
	var completedScores int64
	models.DB.Model(&models.InvitedScore{}).Where("invitation_id = ?", inviteID).Count(&totalScores)
	models.DB.Model(&models.InvitedScore{}).Where("invitation_id = ? AND score IS NOT NULL", inviteID).Count(&completedScores)

	if completedScores < totalScores {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请先完成所有项目的评分",
		})
		return
	}

	// 更新邀请状态为已完成
	invitation.Status = "completed"
	if err := models.DB.Save(&invitation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新邀请状态失败"})
		return
	}

	// 计算邀请评分总分并创建自动评论
	var invitedScores []models.InvitedScore
	if err := models.DB.Where("invitation_id = ?", inviteID).Find(&invitedScores).Error; err == nil {
		totalInvitedScore := 0.0
		for _, score := range invitedScores {
			if score.Score != nil {
				totalInvitedScore += *score.Score
			}
		}

		// 获取被邀请人信息
		var invitee models.Employee
		if err := models.DB.First(&invitee, invitation.InviteeID).Error; err == nil {
			// 创建自动评论：邀请评分（XXX），总分X
			commentContent := fmt.Sprintf("邀请评分（%s），总分%s", invitee.Name, formatScore(totalInvitedScore))
			comment := models.EvaluationComment{
				EvaluationID: invitation.EvaluationID,
				UserID:       invitation.InviteeID,
				Content:      commentContent,
				IsPrivate:    false,
			}
			if err := models.DB.Create(&comment).Error; err != nil {
				// 评论创建失败不影响主流程，仅记录错误
				fmt.Printf("创建邀请评分自动评论失败: %v\n", err)
			}
		}
	}

	// 发送实时通知
	operatorID := c.GetUint("user_id")
	GetNotificationService().SendNotification(operatorID, EventInvitationStatusChange, &invitation)

	// 检查评估状态，如果处于manager_evaluated且启用了绩效规则，检查是否所有邀请都已完成
	var evaluation models.KPIEvaluation
	if err := models.DB.First(&evaluation, invitation.EvaluationID).Error; err == nil {
		if evaluation.Status == "manager_evaluated" {
			// 检查绩效规则是否启用
			var rule models.PerformanceRule
			if err := models.DB.First(&rule).Error; err == nil && rule.Enabled {
				// 检查是否所有邀请都已完成
				allCompleted, err := areAllInvitationsCompleted(invitation.EvaluationID)
				if err == nil && allCompleted {
					// 所有邀请都已完成，自动计算HR评分
					if err := applyPerformanceRuleForEvaluation(invitation.EvaluationID); err == nil {
						// 如果绩效规则应用成功，自动将状态推进到pending_confirm
						models.DB.Model(&evaluation).Update("status", "pending_confirm")
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "邀请评分完成",
		"data":    invitation,
	})
}

// 获取邀请的详细信息（包括评分）
func GetInvitationDetails(c *gin.Context) {
	invitationID := c.Param("id")
	inviteID, err := strconv.ParseUint(invitationID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的邀请ID"})
		return
	}

	// 获取当前用户ID
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}
	userID := currentUserID.(uint)

	// 验证邀请是否存在
	var invitation models.EvaluationInvitation
	if err := models.DB.Preload("Evaluation.Employee.Department").Preload("Evaluation.Employee").Preload("Evaluation.Template").
		Preload("Inviter").Preload("Invitee").First(&invitation, inviteID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "邀请不存在"})
		return
	}

	// 验证权限：只有被邀请人或HR可以查看详情
	var currentUser models.Employee
	if err := models.DB.First(&currentUser, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	if invitation.InviteeID != userID && currentUser.Role != "hr" {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限查看此邀请详情"})
		return
	}

	// 获取评分记录
	var scores []models.InvitedScore
	if err := models.DB.Preload("Item").Where("invitation_id = ?", inviteID).Find(&scores).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取评分记录失败"})
		return
	}

	invitation.Scores = scores

	c.JSON(http.StatusOK, gin.H{
		"data": invitation,
	})
}

// 撤销邀请
func CancelInvitation(c *gin.Context) {
	// 获取邀请ID
	inviteIDParam := c.Param("id")
	inviteID, err := strconv.ParseUint(inviteIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "邀请ID格式错误"})
		return
	}

	// 获取当前用户ID
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}
	userID := currentUserID.(uint)

	// 验证用户是否为HR
	var currentUser models.Employee
	if err := models.DB.First(&currentUser, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	if currentUser.Role != "hr" {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有HR可以撤销邀请"})
		return
	}

	// 验证邀请是否存在
	var invitation models.EvaluationInvitation
	if err := models.DB.Preload("Invitee").Preload("Inviter").First(&invitation, uint(inviteID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "邀请不存在"})
		return
	}

	// 只有待接受状态的邀请才可以撤销
	if invitation.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只有待接受状态的邀请才可以撤销"})
		return
	}

	// 更新邀请状态为已撤销
	if err := models.DB.Model(&invitation).Update("status", "cancelled").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "撤销邀请失败"})
		return
	}

	// 重新加载邀请数据以获取最新状态
	models.DB.First(&invitation, uint(inviteID))

	// 检查评估状态，如果处于manager_evaluated且启用了绩效规则，检查是否所有邀请都已完成
	var evaluation models.KPIEvaluation
	if err := models.DB.First(&evaluation, invitation.EvaluationID).Error; err == nil {
		if evaluation.Status == "manager_evaluated" {
			// 检查绩效规则是否启用
			var rule models.PerformanceRule
			if err := models.DB.First(&rule).Error; err == nil && rule.Enabled {
				// 检查是否所有邀请都已完成
				allCompleted, err := areAllInvitationsCompleted(invitation.EvaluationID)
				if err == nil && allCompleted {
					// 所有邀请都已完成，自动计算HR评分
					if err := applyPerformanceRuleForEvaluation(invitation.EvaluationID); err == nil {
						// 如果绩效规则应用成功，自动将状态推进到pending_confirm
						models.DB.Model(&evaluation).Update("status", "pending_confirm")
					}
				}
			}
		}
	}

	// 发送实时通知
	operatorID := c.GetUint("user_id")
	GetNotificationService().SendNotification(operatorID, EventInvitationStatusChange, &invitation)

	c.JSON(http.StatusOK, gin.H{
		"data":    invitation,
		"message": "邀请已撤销",
	})
}

// 重新邀请
func ReinviteInvitation(c *gin.Context) {
	// 获取邀请ID
	inviteIDParam := c.Param("id")
	inviteID, err := strconv.ParseUint(inviteIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "邀请ID格式错误"})
		return
	}

	// 获取当前用户ID
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}
	userID := currentUserID.(uint)

	// 验证用户是否为HR
	var currentUser models.Employee
	if err := models.DB.First(&currentUser, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	if currentUser.Role != "hr" {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有HR可以重新邀请"})
		return
	}

	// 验证邀请是否存在
	var invitation models.EvaluationInvitation
	if err := models.DB.Preload("Invitee").Preload("Inviter").Preload("Evaluation.Employee").Preload("Evaluation.Template").First(&invitation, uint(inviteID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "邀请不存在"})
		return
	}

	// 只有已拒绝状态的邀请才可以重新邀请
	if invitation.Status != "declined" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只有已拒绝状态的邀请才可以重新邀请"})
		return
	}

	// 更新邀请状态为待接受
	if err := models.DB.Model(&invitation).Update("status", "pending").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重新邀请失败"})
		return
	}

	// 发送 DooTask 机器人通知
	dooTaskClient := utils.NewDooTaskClient(c.GetHeader("DooTaskAuth"))
	appConfigJSON := utils.BuildKPIInvitationAppConfig(invitation.ID, invitation.EvaluationID)
	periodValue := utils.GetPeriodValue(invitation.Evaluation.Period, invitation.Evaluation.Year, invitation.Evaluation.Month, invitation.Evaluation.Quarter)

	if invitation.Invitee.DooTaskUserID != nil {
		reinviteMessage := fmt.Sprintf(
			"**【重新邀请】你收到绩效评分邀请，请及时处理**\n- 被评估员工：%s\n- 考核模板：%s\n- 考核周期：%s\n- 邀请人：%s\n\n> <div class=\"open-micro-app\" data-app-config='%s'>查看详情：点击查看详情</div>",
			invitation.Evaluation.Employee.Name,
			invitation.Evaluation.Template.Name,
			periodValue,
			c.GetString("user_name"),
			appConfigJSON,
		)
		_ = dooTaskClient.SendBotMessage(invitation.Invitee.DooTaskUserID, reinviteMessage)
	}

	// 发送实时通知
	operatorID := c.GetUint("user_id")
	GetNotificationService().SendNotification(operatorID, EventInvitationStatusChange, &invitation)

	c.JSON(http.StatusOK, gin.H{
		"data":    invitation,
		"message": "重新邀请成功",
	})
}

// 删除邀请
func DeleteInvitation(c *gin.Context) {
	// 获取邀请ID
	inviteIDParam := c.Param("id")
	inviteID, err := strconv.ParseUint(inviteIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "邀请ID格式错误"})
		return
	}

	// 获取当前用户ID
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}
	userID := currentUserID.(uint)

	// 验证用户是否为HR
	var currentUser models.Employee
	if err := models.DB.First(&currentUser, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	if currentUser.Role != "hr" {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有HR可以删除邀请"})
		return
	}

	// 验证邀请是否存在
	var invitation models.EvaluationInvitation
	if err := models.DB.Preload("Evaluation").First(&invitation, uint(inviteID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "邀请不存在"})
		return
	}

	// 保存评估ID和状态，用于删除后重新计算HR评分
	evaluationID := invitation.EvaluationID
	evaluationStatus := invitation.Evaluation.Status
	wasCompleted := invitation.Status == "completed"

	// 开始事务
	tx := models.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "开始事务失败"})
		return
	}

	// 删除相关的评分记录
	if err := tx.Where("invitation_id = ?", inviteID).Delete(&models.InvitedScore{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除评分记录失败"})
		return
	}

	// 删除邀请记录
	if err := tx.Delete(&invitation).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除邀请失败"})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交事务失败"})
		return
	}

	// 如果删除的是已完成的邀请，且评估处于pending_confirm状态，且启用了绩效规则，需要重新计算HR评分
	if wasCompleted && evaluationStatus == "pending_confirm" {
		// 检查绩效规则是否启用
		var rule models.PerformanceRule
		if err := models.DB.First(&rule).Error; err == nil && rule.Enabled {
			// 重新计算HR评分
			if err := applyPerformanceRuleForEvaluation(evaluationID); err != nil {
				// 重新计算失败不影响删除操作，仅记录错误
				fmt.Printf("删除邀请后重新计算HR评分失败: %v\n", err)
			}
		}
	}

	// 发送实时通知
	operatorID := c.GetUint("user_id")
	GetNotificationService().SendNotification(operatorID, EventInvitationDeleted, &invitation)

	c.JSON(http.StatusOK, gin.H{
		"message": "邀请删除成功",
	})
}

// 获取待确认邀请数量
func GetPendingCountInvitations(c *gin.Context) {
	// 获取当前用户ID
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未找到用户信息"})
		return
	}

	userID := currentUserID.(uint)

	var count int64
	if err := models.DB.Model(&models.EvaluationInvitation{}).
		Where("invitee_id = ? AND status = ?", userID, "pending").
		Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取待确认邀请数量失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}
