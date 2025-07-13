package models

import (
	"time"
)

// 部门模型
type Department struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联关系
	Employees []Employee `json:"employees,omitempty" gorm:"foreignKey:DepartmentID"`
}

// 员工模型
type Employee struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null"`
	Email        string    `json:"email" gorm:"unique;not null"`
	Password     string    `json:"-" gorm:"not null"` // 密码，JSON中不返回
	Position     string    `json:"position"`
	DepartmentID uint      `json:"department_id"`
	ManagerID    *uint     `json:"manager_id"`                   // 直属上级ID，可以为空
	Role         string    `json:"role" gorm:"default:employee"` // employee, manager, hr
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联关系
	Department   Department `json:"department,omitempty" gorm:"foreignKey:DepartmentID"`
	Manager      *Employee  `json:"manager,omitempty" gorm:"foreignKey:ManagerID"`
	Subordinates []Employee `json:"subordinates,omitempty" gorm:"foreignKey:ManagerID"`
}

// KPI模板模型
type KPITemplate struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Period      string    `json:"period"` // monthly, quarterly, yearly
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联关系
	Items []KPIItem `json:"items,omitempty" gorm:"foreignKey:TemplateID"`
}

// KPI考核项目模型
type KPIItem struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	TemplateID  uint      `json:"template_id"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	MaxScore    float64   `json:"max_score"` // 满分
	Order       int       `json:"order"`     // 排序
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联关系
	Template KPITemplate `json:"template,omitempty" gorm:"foreignKey:TemplateID"`
}

// KPI评估记录模型
type KPIEvaluation struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	EmployeeID   uint      `json:"employee_id"`
	TemplateID   uint      `json:"template_id"`
	Period       string    `json:"period"` // 2024-01, 2024-Q1, 2024
	Year         int       `json:"year"`
	Month        *int      `json:"month,omitempty"`
	Quarter      *int      `json:"quarter,omitempty"`
	Status       string    `json:"status" gorm:"default:pending"` // pending, self_evaluated, manager_evaluated, pending_confirm, completed
	TotalScore   float64   `json:"total_score"`
	FinalComment string    `json:"final_comment"`
	
	// 截止时间相关字段
	PeriodEndDate         *time.Time `json:"period_end_date,omitempty"`         // 考核周期结束时间
	SelfEvalDeadline      *time.Time `json:"self_eval_deadline,omitempty"`      // 自评截止时间
	ManagerEvalDeadline   *time.Time `json:"manager_eval_deadline,omitempty"`   // 主管评分截止时间
	HRReviewDeadline      *time.Time `json:"hr_review_deadline,omitempty"`      // HR审核截止时间
	FinalConfirmDeadline  *time.Time `json:"final_confirm_deadline,omitempty"`  // 最终确认截止时间
	TimeMode              string     `json:"time_mode" gorm:"default:standard"` // standard, compressed, emergency, custom
	IsOverdue             bool       `json:"is_overdue" gorm:"default:false"`   // 是否超时
	
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联关系
	Employee Employee    `json:"employee,omitempty" gorm:"foreignKey:EmployeeID"`
	Template KPITemplate `json:"template,omitempty" gorm:"foreignKey:TemplateID"`
	Scores   []KPIScore  `json:"scores,omitempty" gorm:"foreignKey:EvaluationID"`
}

// KPI具体得分模型
type KPIScore struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	EvaluationID   uint      `json:"evaluation_id"`
	ItemID         uint      `json:"item_id"`
	SelfScore      *float64  `json:"self_score,omitempty"`    // 自评分数
	SelfComment    string    `json:"self_comment"`            // 自评评价
	ManagerScore   *float64  `json:"manager_score,omitempty"` // 上级评分
	ManagerComment string    `json:"manager_comment"`         // 上级评价
	FinalScore     *float64  `json:"final_score,omitempty"`   // 最终得分
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// 关联关系
	Evaluation KPIEvaluation `json:"evaluation,omitempty" gorm:"foreignKey:EvaluationID"`
	Item       KPIItem       `json:"item,omitempty" gorm:"foreignKey:ItemID"`
}

// 评论模型
type EvaluationComment struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	EvaluationID uint      `json:"evaluation_id"`
	UserID       uint      `json:"user_id"` // 评论者ID
	Content      string    `json:"content" gorm:"not null"`
	IsPrivate    bool      `json:"is_private" gorm:"default:false"` // 是否仅自己可见
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联关系
	Evaluation KPIEvaluation `json:"evaluation,omitempty" gorm:"foreignKey:EvaluationID"`
	User       Employee      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// 系统设置模型
type SystemSetting struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Key       string    `json:"key" gorm:"unique;not null"`
	Value     string    `json:"value" gorm:"not null"`
	Type      string    `json:"type" gorm:"default:string"` // string, boolean, number, json
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
