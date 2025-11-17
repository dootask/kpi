package handlers

import (
	"net/http"
	"strconv"

	"dootask-kpi-server/models"

	"github.com/gin-gonic/gin"
)

// 获取所有部门
func GetDepartments(c *gin.Context) {
	var departments []models.Department

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	search := c.Query("search")

	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询，只预加载在职员工
	query := models.DB.Preload("Employees", "is_active = ?", true)

	// 添加搜索条件
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name LIKE ? OR description LIKE ?",
			searchPattern, searchPattern)
	}

	// 获取总数
	var total int64
	countQuery := models.DB.Model(&models.Department{})
	if search != "" {
		searchPattern := "%" + search + "%"
		countQuery = countQuery.Where("name LIKE ? OR description LIKE ?",
			searchPattern, searchPattern)
	}

	if err := countQuery.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取部门总数失败",
			"message": err.Error(),
		})
		return
	}

	// 分页查询
	offset := (page - 1) * pageSize
	result := query.Offset(offset).Limit(pageSize).Find(&departments)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取部门列表失败",
			"message": result.Error.Error(),
		})
		return
	}

	// 计算分页信息
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	c.JSON(http.StatusOK, gin.H{
		"data":       departments,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": totalPages,
		"hasNext":    page < totalPages,
		"hasPrev":    page > 1,
	})
}

// 创建部门
func CreateDepartment(c *gin.Context) {
	var department models.Department

	if err := c.ShouldBindJSON(&department); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	result := models.DB.Create(&department)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建部门失败",
			"message": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "部门创建成功",
		"data":    department,
	})
}

// 获取单个部门
func GetDepartment(c *gin.Context) {
	id := c.Param("id")
	departmentId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的部门ID",
		})
		return
	}

	var department models.Department
	result := models.DB.Preload("Employees", "is_active = ?", true).First(&department, departmentId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "部门不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": department,
	})
}

// 更新部门
func UpdateDepartment(c *gin.Context) {
	id := c.Param("id")
	departmentId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的部门ID",
		})
		return
	}

	var department models.Department
	result := models.DB.First(&department, departmentId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "部门不存在",
		})
		return
	}

	var updateData models.Department
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	result = models.DB.Model(&department).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新部门失败",
			"message": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "部门更新成功",
		"data":    department,
	})
}

// 删除部门
func DeleteDepartment(c *gin.Context) {
	id := c.Param("id")
	departmentId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的部门ID",
		})
		return
	}

	// 检查是否有在职员工在该部门
	var employeeCount int64
	models.DB.Model(&models.Employee{}).Where("department_id = ? AND is_active = ?", departmentId, true).Count(&employeeCount)
	if employeeCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "该部门下还有在职员工，无法删除",
		})
		return
	}

	result := models.DB.Delete(&models.Department{}, departmentId)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除部门失败",
			"message": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "部门删除成功",
	})
}
