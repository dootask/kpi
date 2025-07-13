package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

// 时间配置结构
type DeadlineDays struct {
	SelfEval      int `json:"self_eval"`
	ManagerEval   int `json:"manager_eval"`
	HRReview      int `json:"hr_review"`
	FinalConfirm  int `json:"final_confirm"`
}

type TimeThreshold struct {
	Standard   int `json:"standard"`
	Compressed int `json:"compressed"`
	Emergency  int `json:"emergency"`
}

// 截止时间集合
type DeadlineSet struct {
	PeriodEnd         *time.Time `json:"period_end"`
	SelfEvalDeadline  *time.Time `json:"self_eval_deadline"`
	ManagerEvalDeadline *time.Time `json:"manager_eval_deadline"`
	HRReviewDeadline   *time.Time `json:"hr_review_deadline"`
	FinalConfirmDeadline *time.Time `json:"final_confirm_deadline"`
	TimeMode          string     `json:"time_mode"`
	IsValid           bool       `json:"is_valid"`
	Message           string     `json:"message,omitempty"`
}

// 截止时间计算器
type DeadlineCalculator struct {
	PeriodType        string
	CreatedAt         time.Time
	Year              int
	Month             *int
	Quarter           *int
	StandardDays      DeadlineDays
	CompressedDays    DeadlineDays
	MinimumDays       DeadlineDays
	TimeThreshold     TimeThreshold
}

// 计算周期结束时间
func (d *DeadlineCalculator) CalculatePeriodEnd() time.Time {
	switch d.PeriodType {
	case "monthly":
		if d.Month != nil {
			// 计算指定月份的最后一天
			return time.Date(d.Year, time.Month(*d.Month+1), 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1)
		}
		// 如果没有指定月份，使用当前月份
		now := time.Now()
		return time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1)
	case "quarterly":
		if d.Quarter != nil {
			// 计算指定季度的最后一天
			endMonth := *d.Quarter * 3
			return time.Date(d.Year, time.Month(endMonth+1), 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1)
		}
		// 如果没有指定季度，使用当前季度
		now := time.Now()
		currentQuarter := int((now.Month()-1)/3) + 1
		endMonth := currentQuarter * 3
		return time.Date(now.Year(), time.Month(endMonth+1), 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1)
	case "yearly":
		return time.Date(d.Year, 12, 31, 0, 0, 0, 0, time.Local)
	default:
		// 默认为年度
		return time.Date(d.Year, 12, 31, 0, 0, 0, 0, time.Local)
	}
}

// 计算可用天数
func (d *DeadlineCalculator) CalculateAvailableDays() int {
	periodEnd := d.CalculatePeriodEnd()
	return int(periodEnd.Sub(d.CreatedAt).Hours() / 24)
}

// 计算截止时间
func (d *DeadlineCalculator) CalculateDeadlines() DeadlineSet {
	periodEnd := d.CalculatePeriodEnd()
	availableDays := d.CalculateAvailableDays()
	
	// 计算最小所需天数
	minRequired := d.MinimumDays.SelfEval + d.MinimumDays.ManagerEval + d.MinimumDays.HRReview + d.MinimumDays.FinalConfirm
	
	result := DeadlineSet{
		PeriodEnd: &periodEnd,
		IsValid:   true,
	}
	
	// 判断时间模式
	var days DeadlineDays
	switch {
	case availableDays < minRequired:
		// 时间不足，无法完成考核
		result.IsValid = false
		result.Message = "时间不足，建议延期到下个周期"
		result.TimeMode = "insufficient"
		return result
	case availableDays < d.TimeThreshold.Emergency:
		result.TimeMode = "emergency"
		days = d.MinimumDays
	case availableDays < d.TimeThreshold.Compressed:
		result.TimeMode = "compressed"
		days = d.CompressedDays
	default:
		result.TimeMode = "standard"
		days = d.StandardDays
	}
	
	// 计算各阶段截止时间
	selfEvalDeadline := d.CreatedAt.AddDate(0, 0, days.SelfEval)
	managerEvalDeadline := selfEvalDeadline.AddDate(0, 0, days.ManagerEval)
	hrReviewDeadline := managerEvalDeadline.AddDate(0, 0, days.HRReview)
	finalConfirmDeadline := hrReviewDeadline.AddDate(0, 0, days.FinalConfirm)
	
	result.SelfEvalDeadline = &selfEvalDeadline
	result.ManagerEvalDeadline = &managerEvalDeadline
	result.HRReviewDeadline = &hrReviewDeadline
	result.FinalConfirmDeadline = &finalConfirmDeadline
	
	return result
}

// 创建自定义截止时间
func (d *DeadlineCalculator) CreateCustomDeadlines(selfEval, managerEval, hrReview, finalConfirm time.Time) DeadlineSet {
	periodEnd := d.CalculatePeriodEnd()
	
	result := DeadlineSet{
		PeriodEnd:            &periodEnd,
		SelfEvalDeadline:     &selfEval,
		ManagerEvalDeadline:  &managerEval,
		HRReviewDeadline:     &hrReview,
		FinalConfirmDeadline: &finalConfirm,
		TimeMode:             "custom",
		IsValid:              true,
	}
	
	// 验证自定义时间
	if err := d.ValidateCustomDeadlines(result); err != nil {
		result.IsValid = false
		result.Message = err.Error()
	}
	
	return result
}

// 验证自定义截止时间
func (d *DeadlineCalculator) ValidateCustomDeadlines(deadlines DeadlineSet) error {
	// 验证时间顺序
	if deadlines.SelfEvalDeadline.After(*deadlines.ManagerEvalDeadline) {
		return fmt.Errorf("自评截止时间不能晚于主管评分截止时间")
	}
	if deadlines.ManagerEvalDeadline.After(*deadlines.HRReviewDeadline) {
		return fmt.Errorf("主管评分截止时间不能晚于HR审核截止时间")
	}
	if deadlines.HRReviewDeadline.After(*deadlines.FinalConfirmDeadline) {
		return fmt.Errorf("HR审核截止时间不能晚于最终确认截止时间")
	}
	
	// 验证不能早于创建时间
	if deadlines.SelfEvalDeadline.Before(d.CreatedAt) {
		return fmt.Errorf("自评截止时间不能早于创建时间")
	}
	
	// 验证不能超过周期结束时间
	if deadlines.FinalConfirmDeadline.After(*deadlines.PeriodEnd) {
		return fmt.Errorf("最终确认时间不能超过考核周期结束时间")
	}
	
	// 验证最小时间间隔
	if deadlines.SelfEvalDeadline.Sub(d.CreatedAt).Hours()/24 < float64(d.MinimumDays.SelfEval) {
		return fmt.Errorf("自评时间至少需要 %d 天", d.MinimumDays.SelfEval)
	}
	
	if deadlines.ManagerEvalDeadline.Sub(*deadlines.SelfEvalDeadline).Hours()/24 < float64(d.MinimumDays.ManagerEval) {
		return fmt.Errorf("主管评分时间至少需要 %d 天", d.MinimumDays.ManagerEval)
	}
	
	if deadlines.HRReviewDeadline.Sub(*deadlines.ManagerEvalDeadline).Hours()/24 < float64(d.MinimumDays.HRReview) {
		return fmt.Errorf("HR审核时间至少需要 %d 天", d.MinimumDays.HRReview)
	}
	
	if deadlines.FinalConfirmDeadline.Sub(*deadlines.HRReviewDeadline).Hours()/24 < float64(d.MinimumDays.FinalConfirm) {
		return fmt.Errorf("最终确认时间至少需要 %d 天", d.MinimumDays.FinalConfirm)
	}
	
	return nil
}

// 从JSON配置创建时间配置
func ParseDeadlineDaysFromJSON(jsonStr string) (DeadlineDays, error) {
	var days DeadlineDays
	err := json.Unmarshal([]byte(jsonStr), &days)
	return days, err
}

// 从JSON配置创建时间阈值
func ParseTimeThresholdFromJSON(jsonStr string) (TimeThreshold, error) {
	var threshold TimeThreshold
	err := json.Unmarshal([]byte(jsonStr), &threshold)
	return threshold, err
}

// 检查是否超时
func IsOverdue(deadline *time.Time) bool {
	if deadline == nil {
		return false
	}
	return time.Now().After(*deadline)
}

// 计算剩余天数
func GetRemainingDays(deadline *time.Time) int {
	if deadline == nil {
		return 0
	}
	remaining := deadline.Sub(time.Now()).Hours() / 24
	if remaining < 0 {
		return 0
	}
	return int(remaining)
}

// 格式化截止时间显示
func FormatDeadline(deadline *time.Time) string {
	if deadline == nil {
		return ""
	}
	return deadline.Format("2006-01-02 15:04")
}