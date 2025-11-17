package models

import (
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// 初始化数据库连接
func InitDB() {
	var err error

	// 创建db目录
	os.MkdirAll("db", 0755)

	// 连接SQLite数据库
	DB, err = gorm.Open(sqlite.Open("db/kpi.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	log.Println("数据库连接成功")

	// 自动迁移数据库表
	err = DB.AutoMigrate(
		&Department{},
		&Employee{},
		&KPITemplate{},
		&KPIItem{},
		&KPIEvaluation{},
		&KPIScore{},
		&EvaluationComment{},
		&EvaluationInvitation{},
		&InvitedScore{},
		&SystemSetting{},
		&PerformanceRule{},
	)
	if err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	log.Println("数据库表迁移完成")
}

// 创建测试数据
func CreateTestData() {
	// 检查是否已有数据
	var count int64
	var count2 int64
	var ruleCount int64
	DB.Model(&Department{}).Count(&count)
	DB.Model(&KPITemplate{}).Count(&count2)
	DB.Model(&PerformanceRule{}).Count(&ruleCount)
	if count > 0 && count2 > 0 && ruleCount > 0 {
		log.Println("测试数据已存在，跳过创建")
		return
	}

	// 根据系统模式创建测试数据
	if os.Getenv("SYSTEM_MODE") == "integrated" {
		log.Println("集成模式，仅创建KPI模板")
	} else {
		log.Println("独立模式，创建测试数据")
		CreateTestDataForUser()
	}
	CreateTestDataForTemplate()
	CreateTestDataForPerformanceRule()

	log.Println("测试数据创建完成")
}

// 创建测试数据（用户）
func CreateTestDataForUser() {
	// 创建部门
	departments := []Department{
		{Name: "技术部", Description: "负责产品研发和技术支持"},
		{Name: "市场部", Description: "负责市场营销和客户关系"},
		{Name: "人事部", Description: "负责人力资源管理"},
		{Name: "财务部", Description: "负责财务管理和会计"},
	}

	for _, dept := range departments {
		DB.Create(&dept)
	}

	// 生成默认密码哈希
	defaultPassword := "123456"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("密码哈希生成失败:", err)
	}

	// 创建员工
	employees := []Employee{
		{Name: "张三", Email: "zhangsan@company.com", Password: string(hashedPassword), Position: "技术总监", DepartmentID: 1, Role: "manager"},
		{Name: "李四", Email: "lisi@company.com", Password: string(hashedPassword), Position: "高级开发工程师", DepartmentID: 1, ManagerID: getUintPtr(1), Role: "employee"},
		{Name: "王五", Email: "wangwu@company.com", Password: string(hashedPassword), Position: "前端开发工程师", DepartmentID: 1, ManagerID: getUintPtr(1), Role: "employee"},
		{Name: "赵六", Email: "zhaoliu@company.com", Password: string(hashedPassword), Position: "市场总监", DepartmentID: 2, Role: "manager"},
		{Name: "钱七", Email: "qianqi@company.com", Password: string(hashedPassword), Position: "市场专员", DepartmentID: 2, ManagerID: getUintPtr(4), Role: "employee"},
		{Name: "孙八", Email: "sunba@company.com", Password: string(hashedPassword), Position: "HR经理", DepartmentID: 3, Role: "hr"},
		{Name: "周九", Email: "zhoujiu@company.com", Password: string(hashedPassword), Position: "财务经理", DepartmentID: 4, Role: "manager"},
	}

	for _, emp := range employees {
		DB.Create(&emp)
	}
}

// 创建测试数据（KPI模板）
func CreateTestDataForTemplate() {
	// 创建KPI模板
	templates := []KPITemplate{
		{Name: "技术岗位月度考核", Description: "适用于技术人员的月度绩效考核", Period: "monthly"},
		{Name: "市场岗位季度考核", Description: "适用于市场人员的季度绩效考核", Period: "quarterly"},
		{Name: "管理岗位年度考核", Description: "适用于管理人员的年度绩效考核", Period: "yearly"},
	}

	for _, template := range templates {
		DB.Create(&template)
	}

	// 创建KPI考核项目
	items := []KPIItem{
		// 技术岗位月度考核项目
		{TemplateID: 1, Name: "代码质量", Description: "代码规范性、可维护性评估", MaxScore: 20, Order: 1},
		{TemplateID: 1, Name: "任务完成度", Description: "按时完成分配的开发任务", MaxScore: 25, Order: 2},
		{TemplateID: 1, Name: "技术创新", Description: "技术方案创新和改进", MaxScore: 15, Order: 3},
		{TemplateID: 1, Name: "团队协作", Description: "与团队成员的协作配合", MaxScore: 20, Order: 4},
		{TemplateID: 1, Name: "学习成长", Description: "技术学习和个人提升", MaxScore: 20, Order: 5},

		// 市场岗位季度考核项目
		{TemplateID: 2, Name: "销售业绩", Description: "季度销售目标达成情况", MaxScore: 40, Order: 1},
		{TemplateID: 2, Name: "客户维护", Description: "客户关系维护和满意度", MaxScore: 30, Order: 2},
		{TemplateID: 2, Name: "市场活动", Description: "市场推广活动执行效果", MaxScore: 30, Order: 3},

		// 管理岗位年度考核项目
		{TemplateID: 3, Name: "团队管理", Description: "团队建设和人员管理", MaxScore: 30, Order: 1},
		{TemplateID: 3, Name: "业务发展", Description: "部门业务发展和目标达成", MaxScore: 35, Order: 2},
		{TemplateID: 3, Name: "战略规划", Description: "部门战略规划和执行", MaxScore: 35, Order: 3},
	}

	for _, item := range items {
		DB.Create(&item)
	}

	// 创建默认系统设置
	settings := []SystemSetting{
		{Key: "allow_registration", Value: "true", Type: "boolean"},
	}

	for _, setting := range settings {
		DB.Create(&setting)
	}
}

// 创建测试数据（绩效规则）
func CreateTestDataForPerformanceRule() {
	var count int64
	DB.Model(&PerformanceRule{}).Count(&count)
	if count > 0 {
		return
	}

	defaultRule := DefaultPerformanceRule()
	DB.Create(&defaultRule)
}

// 辅助函数：获取uint指针
func getUintPtr(val uint) *uint {
	return &val
}
