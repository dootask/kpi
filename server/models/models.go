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
	ID            uint      `json:"id" gorm:"primaryKey"`
	Name          string    `json:"name" gorm:"not null"`
	Email         string    `json:"email" gorm:"unique;not null"`
	DooTaskUserID *uint     `json:"-"`                 // DooTask用户ID，JSON中不返回
	Password      string    `json:"-" gorm:"not null"` // 密码，JSON中不返回
	Position      string    `json:"position"`
	DepartmentID  uint      `json:"department_id"`
	ManagerID     *uint     `json:"manager_id"`                   // 直属上级ID，可以为空
	Role          string    `json:"role" gorm:"default:employee"` // employee, manager, hr
	IsActive      bool      `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

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
	ID              uint      `json:"id" gorm:"primaryKey"`
	EmployeeID      uint      `json:"employee_id"`
	TemplateID      uint      `json:"template_id"`
	Period          string    `json:"period"` // 2024-01, 2024-Q1, 2024
	Year            int       `json:"year"`
	Month           *int      `json:"month,omitempty"`
	Quarter         *int      `json:"quarter,omitempty"`
	Status          string    `json:"status" gorm:"default:pending"` // pending, self_evaluated, manager_evaluated, pending_confirm, completed
	TotalScore      float64   `json:"total_score"`
	FinalComment    string    `json:"final_comment"`
	HasObjection    bool      `json:"has_objection" gorm:"default:false"` // 是否有异议
	ObjectionReason string    `json:"objection_reason"`                   // 异议原因（员工填写）
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

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
	SelfScore      *float64  `json:"self_score,omitempty"`              // 自评分数
	SelfComment    string    `json:"self_comment"`                      // 自评评价
	ManagerScore   *float64  `json:"manager_score,omitempty"`           // 上级评分
	ManagerComment string    `json:"manager_comment"`                   // 上级评价
	ManagerAuto    bool      `json:"manager_auto" gorm:"default:false"` // 是否自动填入上级评分
	HRScore        *float64  `json:"hr_score,omitempty"`                // HR评分
	HRComment      string    `json:"hr_comment"`                        // HR评价
	FinalScore     *float64  `json:"final_score,omitempty"`             // 最终得分
	FinalComment   string    `json:"final_comment"`                     // 最终得分说明
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

// 评估邀请模型
type EvaluationInvitation struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	EvaluationID uint      `json:"evaluation_id"`
	InviterID    uint      `json:"inviter_id"`                    // 邀请者ID（HR）
	InviteeID    uint      `json:"invitee_id"`                    // 被邀请者ID
	Status       string    `json:"status" gorm:"default:pending"` // pending, accepted, declined, completed
	Message      string    `json:"message"`                       // 邀请消息
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联关系
	Evaluation KPIEvaluation  `json:"evaluation,omitempty" gorm:"foreignKey:EvaluationID"`
	Inviter    Employee       `json:"inviter,omitempty" gorm:"foreignKey:InviterID"`
	Invitee    Employee       `json:"invitee,omitempty" gorm:"foreignKey:InviteeID"`
	Scores     []InvitedScore `json:"scores,omitempty" gorm:"foreignKey:InvitationID"`
}

// 邀请评分模型
type InvitedScore struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	InvitationID uint      `json:"invitation_id"`
	ItemID       uint      `json:"item_id"`
	Score        *float64  `json:"score,omitempty"` // 邀请评分
	Comment      string    `json:"comment"`         // 邀请评价说明
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联关系
	Invitation EvaluationInvitation `json:"invitation,omitempty" gorm:"foreignKey:InvitationID"`
	Item       KPIItem              `json:"item,omitempty" gorm:"foreignKey:ItemID"`
}

// 绩效评分规则模型
type PerformanceRule struct {
	ID             uint                        `json:"id" gorm:"primaryKey"`
	NoInvitation   PerformanceRuleNoInvitation `json:"no_invitation" gorm:"embedded;embeddedPrefix:no_invitation_"`
	WithInvitation PerformanceRuleWithInvite   `json:"with_invitation" gorm:"embedded;embeddedPrefix:with_invitation_"`
	Enabled        bool                        `json:"enabled" gorm:"not null;default:false"`
	CreatedAt      time.Time                   `json:"created_at"`
	UpdatedAt      time.Time                   `json:"updated_at"`
}

// 无邀请评分规则
type PerformanceRuleNoInvitation struct {
	SelfWeight     float64 `json:"self_weight" gorm:"not null;default:10"`
	SuperiorWeight float64 `json:"superior_weight" gorm:"not null;default:90"`
}

// 有邀请评分规则
type PerformanceRuleWithInvite struct {
	Employee PerformanceRuleEmployee `json:"employee" gorm:"embedded;embeddedPrefix:employee_"`
}

// 员工邀请评分规则
type PerformanceRuleEmployee struct {
	SelfWeight           float64 `json:"self_weight" gorm:"not null;default:10"`
	InviteSuperiorWeight float64 `json:"invite_superior_weight" gorm:"not null;default:30"`
	SuperiorWeight       float64 `json:"superior_weight" gorm:"not null;default:60"`
}

// DefaultPerformanceRule 返回默认绩效规则配置
func DefaultPerformanceRule() PerformanceRule {
	return PerformanceRule{
		Enabled: false,
		NoInvitation: PerformanceRuleNoInvitation{
			SelfWeight:     10,
			SuperiorWeight: 90,
		},
		WithInvitation: PerformanceRuleWithInvite{
			Employee: PerformanceRuleEmployee{
				SelfWeight:           10,
				InviteSuperiorWeight: 30,
				SuperiorWeight:       60,
			},
		},
	}
}
