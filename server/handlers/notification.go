package handlers

import (
	"fmt"
	"time"

	"dootask-kpi-server/models"
)

// 通知事件类型
const (
	// 绩效评估相关事件
	EventEvaluationCreated      = "evaluation_created"
	EventEvaluationUpdated      = "evaluation_updated"
	EventEvaluationDeleted      = "evaluation_deleted"
	EventEvaluationStatusChange = "evaluation_status_changed"

	// 邀请评分相关事件
	EventInvitationCreated      = "invitation_created"
	EventInvitationUpdated      = "invitation_updated"
	EventInvitationDeleted      = "invitation_deleted"
	EventInvitationStatusChange = "invitation_status_changed"

	// 评分更新事件
	EventInvitedScoreUpdated = "invited_score_updated"
	EventSelfScoreUpdated    = "self_score_updated"
	EventManagerScoreUpdated = "manager_score_updated"
	EventHRScoreUpdated      = "hr_score_updated"

	// 异议相关事件
	EventObjectionSubmitted = "objection_submitted" // 员工提交异议
	EventObjectionHandled   = "objection_handled"   // HR处理异议
)

// 通知服务
type NotificationService struct{}

// 全局通知服务实例
var notificationService = &NotificationService{}

// 获取所有HR用户
func (n *NotificationService) GetAllHRUsers() []uint {
	var hrUsers []models.Employee
	models.DB.Where("role = ?", "hr").Find(&hrUsers)

	var hrUserIDs []uint
	for _, user := range hrUsers {
		hrUserIDs = append(hrUserIDs, user.ID)
	}
	return hrUserIDs
}

// 去重用户ID
func (n *NotificationService) DeduplicateUsers(userIDs []uint) []uint {
	seen := make(map[uint]bool)
	result := []uint{}

	for _, id := range userIDs {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	return result
}

// 获取相关用户
func (n *NotificationService) GetRelatedUsers(eventType string, data interface{}) []uint {
	var relatedUsers []uint

	switch eventType {
	case EventEvaluationCreated, EventEvaluationUpdated, EventEvaluationDeleted, EventEvaluationStatusChange:
		evaluation := data.(*models.KPIEvaluation)

		// 预加载相关数据
		models.DB.Preload("Employee").First(&evaluation, evaluation.ID)

		// 被评估员工
		relatedUsers = append(relatedUsers, evaluation.EmployeeID)

		// 被评估员工的主管
		if evaluation.Employee.ManagerID != nil {
			relatedUsers = append(relatedUsers, *evaluation.Employee.ManagerID)
		}

		// 所有HR用户
		hrUsers := n.GetAllHRUsers()
		relatedUsers = append(relatedUsers, hrUsers...)

	case EventInvitationCreated, EventInvitationUpdated, EventInvitationDeleted, EventInvitationStatusChange:
		invitation := data.(*models.EvaluationInvitation)

		// 预加载相关数据
		models.DB.Preload("Evaluation.Employee").First(&invitation, invitation.ID)

		// 被邀请人
		relatedUsers = append(relatedUsers, invitation.InviteeID)

		// 被评估员工
		relatedUsers = append(relatedUsers, invitation.Evaluation.EmployeeID)

		// 被评估员工的主管
		if invitation.Evaluation.Employee.ManagerID != nil {
			relatedUsers = append(relatedUsers, *invitation.Evaluation.Employee.ManagerID)
		}

		// 所有HR用户
		hrUsers := n.GetAllHRUsers()
		relatedUsers = append(relatedUsers, hrUsers...)

		// 邀请发起人
		relatedUsers = append(relatedUsers, invitation.InviterID)

	case EventInvitedScoreUpdated:
		score := data.(*models.InvitedScore)

		// 预加载相关数据
		models.DB.Preload("Invitation.Evaluation.Employee").First(&score, score.ID)

		// 被邀请人（评分者）
		relatedUsers = append(relatedUsers, score.Invitation.InviteeID)

		// 被评估员工
		relatedUsers = append(relatedUsers, score.Invitation.Evaluation.EmployeeID)

		// 被评估员工的主管
		if score.Invitation.Evaluation.Employee.ManagerID != nil {
			relatedUsers = append(relatedUsers, *score.Invitation.Evaluation.Employee.ManagerID)
		}

		// 所有HR用户
		hrUsers := n.GetAllHRUsers()
		relatedUsers = append(relatedUsers, hrUsers...)

		// 邀请发起人
		relatedUsers = append(relatedUsers, score.Invitation.InviterID)

	case EventSelfScoreUpdated, EventManagerScoreUpdated, EventHRScoreUpdated:
		score := data.(*models.KPIScore)

		// 预加载相关数据
		models.DB.Preload("Evaluation.Employee").First(&score, score.ID)

		// 被评估员工
		relatedUsers = append(relatedUsers, score.Evaluation.EmployeeID)

		// 被评估员工的主管
		if score.Evaluation.Employee.ManagerID != nil {
			relatedUsers = append(relatedUsers, *score.Evaluation.Employee.ManagerID)
		}

		// 所有HR用户
		hrUsers := n.GetAllHRUsers()
		relatedUsers = append(relatedUsers, hrUsers...)
	}

	return n.DeduplicateUsers(relatedUsers)
}

// 获取用户信息
func (n *NotificationService) GetUserInfo(userID uint) (*models.Employee, error) {
	var user models.Employee
	if err := models.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// 个性化消息生成
func (n *NotificationService) PersonalizeMessage(userID uint, operatorID uint, eventType string, data interface{}) string {
	// 获取操作者信息
	operator, err := n.GetUserInfo(operatorID)
	if err != nil {
		return "系统通知"
	}

	switch eventType {
	case EventEvaluationCreated:
		evaluation := data.(*models.KPIEvaluation)
		models.DB.Preload("Employee").Preload("Template").First(&evaluation, evaluation.ID)

		if userID == evaluation.EmployeeID {
			return fmt.Sprintf("您的绩效评估已创建：%s", evaluation.Template.Name)
		} else if evaluation.Employee.ManagerID != nil && userID == *evaluation.Employee.ManagerID {
			return fmt.Sprintf("您的下属 %s 的绩效评估已创建", evaluation.Employee.Name)
		} else {
			return fmt.Sprintf("员工 %s 的绩效评估已创建", evaluation.Employee.Name)
		}

	case EventEvaluationUpdated:
		evaluation := data.(*models.KPIEvaluation)
		models.DB.Preload("Employee").Preload("Template").First(&evaluation, evaluation.ID)

		if userID == evaluation.EmployeeID {
			return fmt.Sprintf("您的绩效评估已更新：%s", evaluation.Template.Name)
		} else if evaluation.Employee.ManagerID != nil && userID == *evaluation.Employee.ManagerID {
			return fmt.Sprintf("您的下属 %s 的绩效评估已更新", evaluation.Employee.Name)
		} else {
			return fmt.Sprintf("员工 %s 的绩效评估已更新", evaluation.Employee.Name)
		}

	case EventEvaluationDeleted:
		evaluation := data.(*models.KPIEvaluation)
		models.DB.Preload("Employee").Preload("Template").First(&evaluation, evaluation.ID)

		if userID == evaluation.EmployeeID {
			return fmt.Sprintf("您的绩效评估已删除：%s", evaluation.Template.Name)
		} else if evaluation.Employee.ManagerID != nil && userID == *evaluation.Employee.ManagerID {
			return fmt.Sprintf("您的下属 %s 的绩效评估已删除", evaluation.Employee.Name)
		} else {
			return fmt.Sprintf("员工 %s 的绩效评估已删除", evaluation.Employee.Name)
		}

	case EventEvaluationStatusChange:
		evaluation := data.(*models.KPIEvaluation)
		models.DB.Preload("Employee").Preload("Template").First(&evaluation, evaluation.ID)

		statusText := n.getStatusText(evaluation.Status)

		if userID == evaluation.EmployeeID {
			return fmt.Sprintf("您的绩效评估状态已更新为：%s", statusText)
		} else if evaluation.Employee.ManagerID != nil && userID == *evaluation.Employee.ManagerID {
			return fmt.Sprintf("您的下属 %s 的绩效评估状态已更新为：%s", evaluation.Employee.Name, statusText)
		} else {
			return fmt.Sprintf("员工 %s 的绩效评估状态已更新为：%s", evaluation.Employee.Name, statusText)
		}

	case EventInvitationCreated:
		invitation := data.(*models.EvaluationInvitation)
		models.DB.Preload("Evaluation.Employee").Preload("Evaluation.Template").First(&invitation, invitation.ID)

		if userID == invitation.InviteeID {
			return fmt.Sprintf("您收到了 %s 的绩效评分邀请", invitation.Evaluation.Employee.Name)
		} else if userID == invitation.Evaluation.EmployeeID {
			return fmt.Sprintf("%s 已被邀请为您评分", operator.Name)
		} else if invitation.Evaluation.Employee.ManagerID != nil && userID == *invitation.Evaluation.Employee.ManagerID {
			return fmt.Sprintf("%s 已被邀请为您的下属 %s 评分", operator.Name, invitation.Evaluation.Employee.Name)
		} else {
			return fmt.Sprintf("%s 已被邀请为员工 %s 评分", operator.Name, invitation.Evaluation.Employee.Name)
		}

	case EventInvitationStatusChange:
		invitation := data.(*models.EvaluationInvitation)
		models.DB.Preload("Evaluation.Employee").Preload("Invitee").First(&invitation, invitation.ID)

		statusText := n.getInvitationStatusText(invitation.Status)

		if userID == invitation.InviteeID {
			return fmt.Sprintf("您对 %s 的评分邀请状态已更新为：%s", invitation.Evaluation.Employee.Name, statusText)
		} else if userID == invitation.Evaluation.EmployeeID {
			return fmt.Sprintf("%s 对您的评分邀请状态已更新为：%s", invitation.Invitee.Name, statusText)
		} else if invitation.Evaluation.Employee.ManagerID != nil && userID == *invitation.Evaluation.Employee.ManagerID {
			return fmt.Sprintf("%s 对您的下属 %s 的评分邀请状态已更新为：%s", invitation.Invitee.Name, invitation.Evaluation.Employee.Name, statusText)
		} else {
			return fmt.Sprintf("%s 对员工 %s 的评分邀请状态已更新为：%s", invitation.Invitee.Name, invitation.Evaluation.Employee.Name, statusText)
		}

	case EventInvitedScoreUpdated:
		score := data.(*models.InvitedScore)
		models.DB.Preload("Invitation.Evaluation.Employee").Preload("Invitation.Invitee").First(&score, score.ID)

		if userID == score.Invitation.InviteeID {
			return fmt.Sprintf("您已更新对 %s 的评分", score.Invitation.Evaluation.Employee.Name)
		} else if userID == score.Invitation.Evaluation.EmployeeID {
			return fmt.Sprintf("%s 已更新对您的评分", score.Invitation.Invitee.Name)
		} else if score.Invitation.Evaluation.Employee.ManagerID != nil && userID == *score.Invitation.Evaluation.Employee.ManagerID {
			return fmt.Sprintf("%s 已更新对您的下属 %s 的评分", score.Invitation.Invitee.Name, score.Invitation.Evaluation.Employee.Name)
		} else {
			return fmt.Sprintf("%s 已更新对员工 %s 的评分", score.Invitation.Invitee.Name, score.Invitation.Evaluation.Employee.Name)
		}

	default:
		return "系统通知"
	}
}

// 获取状态文本
func (n *NotificationService) getStatusText(status string) string {
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

// 获取邀请状态文本
func (n *NotificationService) getInvitationStatusText(status string) string {
	switch status {
	case "pending":
		return "待接受"
	case "accepted":
		return "已接受"
	case "declined":
		return "已拒绝"
	case "completed":
		return "已完成"
	case "cancelled":
		return "已取消"
	default:
		return "未知状态"
	}
}

// 发送通知
func (n *NotificationService) SendNotification(operatorID uint, eventType string, data interface{}) {
	// 获取相关用户
	relatedUsers := n.GetRelatedUsers(eventType, data)

	// 移除操作者本人（避免自己通知自己）
	filteredUsers := []uint{}
	for _, userID := range relatedUsers {
		if userID != operatorID {
			filteredUsers = append(filteredUsers, userID)
		}
	}

	// 获取操作者信息
	operator, err := n.GetUserInfo(operatorID)
	if err != nil {
		fmt.Printf("获取操作者信息失败: %v\n", err)
		return
	}

	// 为每个用户生成个性化消息并发送
	for _, userID := range filteredUsers {
		message := n.PersonalizeMessage(userID, operatorID, eventType, data)

		// 创建SSE消息
		sseMessage := SSEMessage{
			Type: eventType,
			Data: SSEEventData{
				ID:           n.getDataID(data),
				EmployeeID:   n.getEmployeeID(data),
				OperatorID:   operatorID,
				OperatorName: operator.Name,
				Message:      message,
				Timestamp:    time.Now().Format(time.RFC3339),
				Payload:      data,
			},
			Timestamp: time.Now().Format(time.RFC3339),
			ID:        fmt.Sprintf("%s-%d", eventType, time.Now().UnixNano()),
		}

		// 发送SSE消息
		sseManager.SendToUser(userID, sseMessage)
	}
}

// 获取数据ID
func (n *NotificationService) getDataID(data interface{}) uint {
	switch v := data.(type) {
	case *models.KPIEvaluation:
		return v.ID
	case *models.EvaluationInvitation:
		return v.ID
	case *models.InvitedScore:
		return v.ID
	case *models.KPIScore:
		return v.ID
	default:
		return 0
	}
}

// 获取员工ID
func (n *NotificationService) getEmployeeID(data interface{}) uint {
	switch v := data.(type) {
	case *models.KPIEvaluation:
		return v.EmployeeID
	case *models.EvaluationInvitation:
		models.DB.Preload("Evaluation").First(&v, v.ID)
		return v.Evaluation.EmployeeID
	case *models.InvitedScore:
		models.DB.Preload("Invitation.Evaluation").First(&v, v.ID)
		return v.Invitation.Evaluation.EmployeeID
	case *models.KPIScore:
		models.DB.Preload("Evaluation").First(&v, v.ID)
		return v.Evaluation.EmployeeID
	default:
		return 0
	}
}

// 导出通知服务实例
func GetNotificationService() *NotificationService {
	return notificationService
}
