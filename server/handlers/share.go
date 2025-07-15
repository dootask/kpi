package handlers

import (
	"net/http"
	"strconv"
	"time"

	"dootask-kpi-server/models"

	"github.com/gin-gonic/gin"
)

// 创建共享请求结构
type CreateShareRequest struct {
	SharedToIDs []uint     `json:"shared_to_ids" binding:"required"`
	Message     string     `json:"message"`
	Deadline    *time.Time `json:"deadline"`
}

// 创建共享
func CreateShare(c *gin.Context) {
	id := c.Param("id")
	evaluationId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的评估ID",
		})
		return
	}

	// 获取当前用户
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未登录",
		})
		return
	}

	// 验证评估是否存在且状态为manager_evaluated
	var evaluation models.KPIEvaluation
	if err := models.DB.First(&evaluation, evaluationId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "评估不存在",
		})
		return
	}

	if evaluation.Status != "manager_evaluated" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "只能在主管评估完成后创建共享",
		})
		return
	}

	var req CreateShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	// 验证被共享人员是否存在
	var employees []models.Employee
	if err := models.DB.Where("id IN ?", req.SharedToIDs).Find(&employees).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "被共享人员不存在",
		})
		return
	}

	if len(employees) != len(req.SharedToIDs) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "部分被共享人员不存在",
		})
		return
	}

	// 创建共享记录
	var shares []models.EvaluationShare
	for _, employeeID := range req.SharedToIDs {
		share := models.EvaluationShare{
			EvaluationID: uint(evaluationId),
			SharedToID:   employeeID,
			SharedByID:   userID.(uint),
			Status:       "pending",
			Message:      req.Message,
			Deadline:     req.Deadline,
		}
		shares = append(shares, share)
	}

	// 开始事务
	tx := models.DB.Begin()

	// 批量创建共享记录
	if err := tx.Create(&shares).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建共享失败",
			"message": err.Error(),
		})
		return
	}

	// 为每个共享创建空的评分记录
	var template models.KPITemplate
	if err := tx.Preload("Items").First(&template, evaluation.TemplateID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取模板失败",
		})
		return
	}

	for _, share := range shares {
		for _, item := range template.Items {
			score := models.ShareScore{
				ShareID: share.ID,
				ItemID:  item.ID,
			}
			if err := tx.Create(&score).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "创建评分记录失败",
				})
				return
			}
		}
	}

	// 更新评估记录
	if err := tx.Model(&evaluation).Updates(map[string]interface{}{
		"has_shares":  true,
		"share_count": len(req.SharedToIDs),
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "更新评估记录失败",
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{
		"message": "共享创建成功",
		"data":    shares,
	})
}

// 获取评估的共享列表
func GetShares(c *gin.Context) {
	id := c.Param("id")
	evaluationId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的评估ID",
		})
		return
	}

	var shares []models.EvaluationShare
	if err := models.DB.Preload("SharedTo").Preload("SharedBy").
		Where("evaluation_id = ?", evaluationId).
		Find(&shares).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取共享列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": shares,
	})
}

// 获取我的共享任务
func GetMyShares(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未登录",
		})
		return
	}

	var shares []models.EvaluationShare
	if err := models.DB.Preload("Evaluation.Employee").
		Preload("Evaluation.Template").
		Preload("SharedBy").
		Where("shared_to_id = ?", userID).
		Order("created_at DESC").
		Find(&shares).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取我的共享任务失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": shares,
	})
}

// 获取共享详情
func GetShareDetail(c *gin.Context) {
	id := c.Param("shareId")
	shareId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的共享ID",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未登录",
		})
		return
	}

	var share models.EvaluationShare
	if err := models.DB.Preload("Evaluation.Employee").
		Preload("Evaluation.Template").
		Preload("Evaluation.Scores.Item").
		Preload("SharedBy").
		Where("id = ? AND shared_to_id = ?", shareId, userID).
		First(&share).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "共享不存在或无权限访问",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": share,
	})
}

// 获取共享评分
func GetShareScores(c *gin.Context) {
	id := c.Param("shareId")
	shareId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的共享ID",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未登录",
		})
		return
	}

	// 验证权限
	var share models.EvaluationShare
	if err := models.DB.Where("id = ? AND shared_to_id = ?", shareId, userID).First(&share).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "共享不存在或无权限访问",
		})
		return
	}

	var scores []models.ShareScore
	if err := models.DB.Preload("Item").
		Where("share_id = ?", shareId).
		Order("item_id").
		Find(&scores).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取评分失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": scores,
	})
}

// 更新共享评分
func UpdateShareScore(c *gin.Context) {
	shareId := c.Param("shareId")
	itemId := c.Param("itemId")
	
	shareIdInt, err := strconv.ParseUint(shareId, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的共享ID",
		})
		return
	}

	itemIdInt, err := strconv.ParseUint(itemId, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的项目ID",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未登录",
		})
		return
	}

	// 验证权限
	var share models.EvaluationShare
	if err := models.DB.Where("id = ? AND shared_to_id = ?", shareIdInt, userID).First(&share).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "共享不存在或无权限访问",
		})
		return
	}

	// 检查共享状态
	if share.Status == "completed" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "共享已完成，无法修改评分",
		})
		return
	}

	type UpdateScoreRequest struct {
		Score   *float64 `json:"score"`
		Comment string   `json:"comment"`
	}

	var req UpdateScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	// 更新评分
	var score models.ShareScore
	if err := models.DB.Where("share_id = ? AND item_id = ?", shareIdInt, itemIdInt).First(&score).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "评分记录不存在",
		})
		return
	}

	if err := models.DB.Model(&score).Updates(map[string]interface{}{
		"score":   req.Score,
		"comment": req.Comment,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "更新评分失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "评分更新成功",
		"data":    score,
	})
}

// 提交共享评分
func SubmitShareScore(c *gin.Context) {
	id := c.Param("shareId")
	shareId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的共享ID",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未登录",
		})
		return
	}

	// 验证权限
	var share models.EvaluationShare
	if err := models.DB.Where("id = ? AND shared_to_id = ?", shareId, userID).First(&share).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "共享不存在或无权限访问",
		})
		return
	}

	// 检查是否已完成
	if share.Status == "completed" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "共享已完成",
		})
		return
	}

	// 更新共享状态
	if err := models.DB.Model(&share).Update("status", "completed").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "提交失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "评分提交成功",
	})
}

// 获取共享评分汇总（HR查看）
func GetShareSummary(c *gin.Context) {
	id := c.Param("id")
	evaluationId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的评估ID",
		})
		return
	}

	// 获取共享记录
	var shares []models.EvaluationShare
	if err := models.DB.Preload("SharedTo").Preload("Scores.Item").
		Where("evaluation_id = ?", evaluationId).
		Find(&shares).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取共享汇总失败",
		})
		return
	}

	// 计算汇总数据
	type ItemSummary struct {
		ItemID      uint    `json:"item_id"`
		ItemName    string  `json:"item_name"`
		AverageScore float64 `json:"average_score"`
		ScoreCount  int     `json:"score_count"`
		Scores      []struct {
			SharedTo string   `json:"shared_to"`
			Score    *float64 `json:"score"`
			Comment  string   `json:"comment"`
		} `json:"scores"`
	}

	itemSummaries := make(map[uint]*ItemSummary)
	
	for _, share := range shares {
		for _, score := range share.Scores {
			if itemSummaries[score.ItemID] == nil {
				itemSummaries[score.ItemID] = &ItemSummary{
					ItemID:   score.ItemID,
					ItemName: score.Item.Name,
					Scores:   []struct {
						SharedTo string   `json:"shared_to"`
						Score    *float64 `json:"score"`
						Comment  string   `json:"comment"`
					}{},
				}
			}
			
			summary := itemSummaries[score.ItemID]
			summary.Scores = append(summary.Scores, struct {
				SharedTo string   `json:"shared_to"`
				Score    *float64 `json:"score"`
				Comment  string   `json:"comment"`
			}{
				SharedTo: share.SharedTo.Name,
				Score:    score.Score,
				Comment:  score.Comment,
			})
			
			if score.Score != nil {
				summary.AverageScore = (summary.AverageScore*float64(summary.ScoreCount) + *score.Score) / float64(summary.ScoreCount+1)
				summary.ScoreCount++
			}
		}
	}

	// 转换为切片
	var summaries []ItemSummary
	for _, summary := range itemSummaries {
		summaries = append(summaries, *summary)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": summaries,
	})
}

// 删除共享
func DeleteShare(c *gin.Context) {
	evaluationId := c.Param("id")
	shareId := c.Param("shareId")
	
	shareIdInt, err := strconv.ParseUint(shareId, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的共享ID",
		})
		return
	}

	// 验证权限
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "用户未登录",
		})
		return
	}

	var share models.EvaluationShare
	if err := models.DB.Where("id = ? AND shared_by_id = ?", shareIdInt, userID).First(&share).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "共享不存在或无权限删除",
		})
		return
	}

	// 开始事务
	tx := models.DB.Begin()

	// 删除相关的评分记录
	if err := tx.Where("share_id = ?", shareIdInt).Delete(&models.ShareScore{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "删除评分记录失败",
		})
		return
	}

	// 删除共享记录
	if err := tx.Delete(&share).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "删除共享失败",
		})
		return
	}

	// 更新评估的共享计数
	var evaluation models.KPIEvaluation
	if err := tx.First(&evaluation, evaluationId).Error; err == nil {
		newCount := evaluation.ShareCount - 1
		hasShares := newCount > 0
		
		tx.Model(&evaluation).Updates(map[string]interface{}{
			"share_count": newCount,
			"has_shares":  hasShares,
		})
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "共享删除成功",
	})
}