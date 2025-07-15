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

					// 评论管理（所有认证用户）
		evaluationRoutes.GET("/:id/comments", handlers.GetEvaluationComments)
		evaluationRoutes.POST("/:id/comments", handlers.CreateEvaluationComment)
		evaluationRoutes.PUT("/:id/comments/:comment_id", handlers.UpdateEvaluationComment)
		evaluationRoutes.DELETE("/:id/comments/:comment_id", handlers.DeleteEvaluationComment)

		// 共享管理（仅HR）
		evaluationRoutes.POST("/:id/shares", handlers.RoleMiddleware("hr"), handlers.CreateShare)
		evaluationRoutes.GET("/:id/shares", handlers.RoleMiddleware("hr"), handlers.GetShares)
		evaluationRoutes.DELETE("/:id/shares/:shareId", handlers.RoleMiddleware("hr"), handlers.DeleteShare)
		evaluationRoutes.GET("/:id/share-summary", handlers.RoleMiddleware("hr"), handlers.GetShareSummary)
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

		// 共享评分管理（被共享人员）
		shareRoutes := protected.Group("/shares")
		{
			shareRoutes.GET("/my", handlers.GetMyShares)
			shareRoutes.GET("/:shareId", handlers.GetShareDetail)
			shareRoutes.GET("/:shareId/scores", handlers.GetShareScores)
			shareRoutes.PUT("/:shareId/scores/:itemId", handlers.UpdateShareScore)
			shareRoutes.POST("/:shareId/submit", handlers.SubmitShareScore)
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
	}
}
