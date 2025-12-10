package routes

import (
	"dootask-kpi-server/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 设置所有路由
func SetupRoutes(r *gin.RouterGroup) {

	// 健康检查（公开）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "OK",
			"message": "KPI系统后端服务正常运行",
		})
	})

	// 内部系统 Hook（公开，仅用于应用内部调用）
	hookRoutes := r.Group("/hooks")
	{
		hookRoutes.POST("/user/onboard", handlers.SystemUserOnboard)
		hookRoutes.POST("/user/offboard", handlers.SystemUserOffboard)
	}

	// 认证路由（公开）
	publicRoutes := r.Group("/auth")
	{
		publicRoutes.POST("/register", handlers.Register)
		publicRoutes.POST("/login", handlers.Login)
		publicRoutes.POST("/login-by-dootask-token", handlers.LoginByDooTaskToken)
		publicRoutes.POST("/refresh", handlers.RefreshToken)
		publicRoutes.GET("/departments", handlers.GetDepartments) // 注册时需要获取部门列表
	}

	// 系统设置（公开，部分需要HR权限）
	settingsRoutes := r.Group("/settings")
	{
		settingsRoutes.GET("", handlers.GetSystemSettings)                                   // 所有用户可以读取设置
		settingsRoutes.PUT("", handlers.RoleMiddleware("hr"), handlers.UpdateSystemSettings) // 只有HR可以修改设置
	}

	// 文件下载（公开）
	downloadRoutes := r.Group("/download")
	{
		downloadRoutes.GET("/exports/:randomKey", handlers.DownloadFile)
		downloadRoutes.GET("/backups/:randomKey", handlers.DownloadBackup)
	}

	// SSE事件流（所有认证用户）
	sseRoutes := r.Group("/events")
	{
		sseRoutes.GET("/stream", handlers.SSEHandler)                              // SSE事件流，通过URL参数token认证
		sseRoutes.GET("/status", handlers.AuthMiddleware(), handlers.GetSSEStatus) // 获取连接状态，所有认证用户
	}

	// 需要认证的路由
	protected := r.Group("/")
	protected.Use(handlers.AuthMiddleware())
	{
		// 当前用户信息
		protected.GET("/me", handlers.GetCurrentUser)

		// 部门管理（HR和管理员）
		departmentRoutes := protected.Group("/departments")
		{
			departmentRoutes.GET("", handlers.GetDepartments)
			departmentRoutes.POST("", handlers.RoleMiddleware("hr", "manager"), handlers.CreateDepartment)
			departmentRoutes.GET("/:id", handlers.GetDepartment)
			departmentRoutes.PUT("/:id", handlers.RoleMiddleware("hr", "manager"), handlers.UpdateDepartment)
			departmentRoutes.DELETE("/:id", handlers.RoleMiddleware("hr"), handlers.DeleteDepartment)
		}

		// 员工管理
		employeeRoutes := protected.Group("/employees")
		{
			employeeRoutes.GET("", handlers.GetEmployees)
			employeeRoutes.POST("", handlers.RoleMiddleware("hr", "manager"), handlers.CreateEmployee)
			employeeRoutes.GET("/:id", handlers.GetEmployee)
			employeeRoutes.PUT("/:id", handlers.RoleMiddleware("hr", "manager"), handlers.UpdateEmployee)
			employeeRoutes.DELETE("/:id", handlers.RoleMiddleware("hr"), handlers.DeleteEmployee)
			employeeRoutes.GET("/:id/subordinates", handlers.GetEmployeeSubordinates)
		}

		// KPI模板管理（HR和管理员）
		templateRoutes := protected.Group("/templates")
		{
			templateRoutes.GET("", handlers.GetTemplates)
			templateRoutes.POST("", handlers.RoleMiddleware("hr", "manager"), handlers.CreateTemplate)
			templateRoutes.GET("/:id", handlers.GetTemplate)
			templateRoutes.PUT("/:id", handlers.RoleMiddleware("hr", "manager"), handlers.UpdateTemplate)
			templateRoutes.DELETE("/:id", handlers.RoleMiddleware("hr"), handlers.DeleteTemplate)
			templateRoutes.GET("/:id/items", handlers.GetTemplateItems)
		}

		// 绩效规则管理（仅HR）
		performanceRuleRoutes := protected.Group("/performance-rules")
		{
			performanceRuleRoutes.GET("", handlers.RoleMiddleware("hr"), handlers.GetPerformanceRule)
			performanceRuleRoutes.PUT("", handlers.RoleMiddleware("hr"), handlers.UpdatePerformanceRule)
		}

		// KPI考核项目管理（HR和管理员）
		itemRoutes := protected.Group("/items")
		{
			itemRoutes.POST("", handlers.RoleMiddleware("hr", "manager"), handlers.CreateItem)
			itemRoutes.GET("/:id", handlers.GetItem)
			itemRoutes.PUT("/:id", handlers.RoleMiddleware("hr", "manager"), handlers.UpdateItem)
			itemRoutes.DELETE("/:id", handlers.RoleMiddleware("hr"), handlers.DeleteItem)
		}

		// KPI评估管理
		evaluationRoutes := protected.Group("/evaluations")
		{
			evaluationRoutes.GET("", handlers.GetEvaluations)
			evaluationRoutes.POST("", handlers.RoleMiddleware("hr", "manager"), handlers.CreateEvaluation)
			evaluationRoutes.GET("/:id", handlers.GetEvaluation)
			evaluationRoutes.PUT("/:id", handlers.UpdateEvaluation)
			evaluationRoutes.DELETE("/:id", handlers.RoleMiddleware("hr"), handlers.DeleteEvaluation)
			evaluationRoutes.GET("/employee/:employeeId", handlers.GetEmployeeEvaluations)
			evaluationRoutes.GET("/pending/:employeeId", handlers.GetPendingEvaluations)
			evaluationRoutes.GET("/pending/count", handlers.GetPendingCountEvaluations)

			// 评论管理（所有认证用户）
			evaluationRoutes.GET("/:id/comments", handlers.GetEvaluationComments)
			evaluationRoutes.POST("/:id/comments", handlers.CreateEvaluationComment)
			evaluationRoutes.PUT("/:id/comments/:comment_id", handlers.UpdateEvaluationComment)
			evaluationRoutes.DELETE("/:id/comments/:comment_id", handlers.DeleteEvaluationComment)

			// 邀请评分管理（HR发起邀请）
			evaluationRoutes.POST("/:id/invitations", handlers.RoleMiddleware("hr"), handlers.CreateInvitation)
			evaluationRoutes.GET("/:id/invitations", handlers.RoleMiddleware("hr"), handlers.GetEvaluationInvitations)

			// 异议处理
			evaluationRoutes.POST("/:id/objection", handlers.SubmitObjection)                                      // 员工提交异议
			evaluationRoutes.PUT("/:id/objection/handle", handlers.RoleMiddleware("hr"), handlers.HandleObjection) // HR处理异议
		}

		// 邀请评分管理
		invitationRoutes := protected.Group("/invitations")
		{
			invitationRoutes.GET("/my", handlers.GetMyInvitations)                                            // 获取我的邀请列表
			invitationRoutes.GET("/sent", handlers.GetMySentInvitations)                                      // 获取我发出的邀请列表
			invitationRoutes.GET("/:id", handlers.GetInvitationDetails)                                       // 获取邀请详情
			invitationRoutes.PUT("/:id/accept", handlers.AcceptInvitation)                                    // 接受邀请
			invitationRoutes.PUT("/:id/decline", handlers.DeclineInvitation)                                  // 拒绝邀请
			invitationRoutes.PUT("/:id/complete", handlers.CompleteInvitation)                                // 完成邀请评分
			invitationRoutes.GET("/:id/scores", handlers.GetInvitationScores)                                 // 获取邀请评分
			invitationRoutes.PUT("/:id/cancel", handlers.RoleMiddleware("hr"), handlers.CancelInvitation)     // 撤销邀请
			invitationRoutes.PUT("/:id/reinvite", handlers.RoleMiddleware("hr"), handlers.ReinviteInvitation) // 重新邀请
			invitationRoutes.DELETE("/:id", handlers.RoleMiddleware("hr"), handlers.DeleteInvitation)         // 删除邀请
			invitationRoutes.GET("/pending/count", handlers.GetPendingCountInvitations)                       // 获取待确认邀请数量
		}

		// 邀请评分记录管理
		invitedScoreRoutes := protected.Group("/invited-scores")
		{
			invitedScoreRoutes.PUT("/:id", handlers.UpdateInvitedScore) // 更新邀请评分
		}

		// KPI评分管理（所有认证用户）
		scoreRoutes := protected.Group("/scores")
		{
			scoreRoutes.GET("/evaluation/:evaluationId", handlers.GetEvaluationScores)
			scoreRoutes.PUT("/:id/self", handlers.UpdateSelfScore)
			scoreRoutes.PUT("/:id/manager", handlers.RoleMiddleware("manager", "hr"), handlers.UpdateManagerScore)
			scoreRoutes.PUT("/:id/hr", handlers.RoleMiddleware("hr"), handlers.UpdateHRScore)
			scoreRoutes.PUT("/:id/final", handlers.RoleMiddleware("hr"), handlers.UpdateFinalScore)
		}

		// 统计分析（所有认证用户）
		statsRoutes := protected.Group("/statistics")
		{
			statsRoutes.GET("/dashboard", handlers.GetDashboardStats)
			statsRoutes.GET("/department/:id", handlers.GetDepartmentStats)
			statsRoutes.GET("/employee/:id", handlers.GetEmployeeStats)
			statsRoutes.GET("/trends", handlers.GetTrends)
			statsRoutes.GET("/data", handlers.GetStatisticsData)
		}

		// 导出功能（管理员和HR）
		exportRoutes := protected.Group("/export")
		exportRoutes.Use(handlers.RoleMiddleware("hr", "manager"))
		{
			exportRoutes.GET("/evaluation/:id", handlers.ExportEvaluationToExcel)
			exportRoutes.GET("/department/:id", handlers.ExportDepartmentToExcel)
			exportRoutes.GET("/period/:period", handlers.ExportPeriodToExcel)
		}

		// 备份管理（仅HR）
		backupRoutes := protected.Group("/backup")
		backupRoutes.Use(handlers.RoleMiddleware("hr"))
		{
			backupRoutes.POST("", handlers.CreateBackup)
			backupRoutes.GET("", handlers.GetBackupHistory)
			backupRoutes.GET("/download/:filename", handlers.GenerateBackupDownloadURL)
			backupRoutes.POST("/restore/:filename", handlers.RestoreBackup)
			backupRoutes.DELETE("/:filename", handlers.DeleteBackup)
		}
	}
}
