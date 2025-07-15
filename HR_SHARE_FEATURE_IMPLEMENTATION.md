# HR共享评分功能实现文档

## 功能概述

本次开发实现了HR共享评分功能，允许HR在绩效流转到 `manager_evaluated` 状态时，将绩效共享给指定人员进行评分和评论。被共享人员的评分仅作为HR最终评分的参考，最终评分权仍在HR手中。

## 核心流程

### 原有流程
```
pending → self_evaluated → manager_evaluated → pending_confirm → completed
```

### 新增共享流程
```
pending → self_evaluated → manager_evaluated → [可选共享评分] → pending_confirm → completed
```

在 `manager_evaluated` 状态时：
- HR可以创建共享，将绩效分发给指定人员评分
- 被共享人员完成评分后，HR可以查看共享评分汇总
- HR基于所有信息（自评+主管评分+共享评分）进行最终确认
- 共享评分不影响正常的状态流转

## 数据模型设计

### 1. EvaluationShare（绩效共享表）
```go
type EvaluationShare struct {
    ID           uint       `json:"id"`
    EvaluationID uint       `json:"evaluation_id"`     // 绩效ID
    SharedToID   uint       `json:"shared_to_id"`      // 被共享人员ID
    SharedByID   uint       `json:"shared_by_id"`      // 共享人（HR）ID
    Status       string     `json:"status"`            // pending, completed, expired
    Message      string     `json:"message"`           // 共享说明
    Deadline     *time.Time `json:"deadline"`          // 评分截止时间
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
}
```

### 2. ShareScore（共享评分表）
```go
type ShareScore struct {
    ID        uint      `json:"id"`
    ShareID   uint      `json:"share_id"`      // 共享ID
    ItemID    uint      `json:"item_id"`       // 考核项ID
    Score     *float64  `json:"score"`         // 评分
    Comment   string    `json:"comment"`       // 评价
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### 3. 扩展现有KPIEvaluation模型
```go
type KPIEvaluation struct {
    // ... 现有字段
    HasShares    bool `json:"has_shares"`     // 是否有共享
    ShareCount   int  `json:"share_count"`    // 共享数量
}
```

## 后端API实现

### 1. 共享管理API（HR专用）
- `POST /api/evaluations/{id}/shares` - 创建共享
- `GET /api/evaluations/{id}/shares` - 获取共享列表
- `DELETE /api/evaluations/{id}/shares/{shareId}` - 删除共享
- `GET /api/evaluations/{id}/share-summary` - 获取共享评分汇总

### 2. 共享评分API（被共享人员）
- `GET /api/shares/my` - 获取我的共享任务
- `GET /api/shares/{shareId}` - 获取共享详情
- `GET /api/shares/{shareId}/scores` - 获取共享评分
- `PUT /api/shares/{shareId}/scores/{itemId}` - 更新单项评分
- `POST /api/shares/{shareId}/submit` - 提交共享评分

### 3. 权限控制
- 创建共享：仅HR且绩效状态为 `manager_evaluated`
- 查看共享：被共享人员只能查看自己的共享
- 评分权限：被共享人员只能评分自己的共享任务

## 前端界面实现

### 1. HR共享管理界面
**位置：** `/app/evaluations/page.tsx` - 共享评分Tab

**功能：**
- 在绩效详情中新增"共享评分"Tab（仅HR可见）
- 显示已创建的共享列表，包括被共享人员、状态、截止时间
- 提供"添加共享"按钮，打开共享创建对话框
- 显示共享评分汇总，包括各项目的平均分和详细评分

**界面元素：**
- 共享列表：显示被共享人员信息和状态
- 共享评分汇总：按考核项目展示所有共享评分
- 创建共享对话框：选择人员、设置说明和截止时间

### 2. 被共享人员界面
**位置：** `/app/shares/page.tsx` - 共享任务列表

**功能：**
- 显示收到的所有共享评分任务
- 按状态区分：待评分、已完成、已过期
- 显示任务详情：员工信息、模板、截止时间、共享说明
- 提供评分入口

**界面元素：**
- 任务列表表格：显示基本信息和操作按钮
- 状态标识：用颜色区分不同状态
- 截止时间提醒：临近截止时间的任务高亮显示

### 3. 共享评分详情界面
**位置：** `/app/shares/[id]/page.tsx` - 评分详情

**功能：**
- 显示员工基本信息和评估信息
- 显示现有评分（自评+主管评分）作为参考
- 提供独立的评分输入区域
- 支持单项保存和整体提交

**界面元素：**
- 基本信息Tab：员工信息、评估信息、现有评分
- 评分详情Tab：独立的评分输入表单
- 状态指示：显示共享状态和完成情况

### 4. 导航菜单更新
**位置：** `/components/sidebar.tsx`

**更新：**
- 在非HR用户菜单中添加"协助评分"选项
- 路由权限检查包含 `/shares` 路径

## 核心功能特点

### 1. 权限控制严格
- 仅HR可以创建和管理共享
- 被共享人员只能查看和评分自己的共享任务
- 共享评分不影响原有的自评、主管评分

### 2. 状态管理清晰
- 共享状态独立于评估状态
- 共享不影响正常的评估流程
- 支持共享的创建、完成、过期状态

### 3. 数据安全保护
- 被共享人员不能查看其他人的共享评分
- 共享评分不能修改原有评分
- 完整的操作权限验证

### 4. 用户体验优化
- 直观的状态指示和进度显示
- 友好的错误提示和加载状态
- 响应式设计适配各种设备

## 业务价值

### 1. 提高评估准确性
- 通过多人评分减少主观偏差
- 提供更全面的员工表现评估
- 增强评估结果的可信度

### 2. 增强决策支持
- 为HR提供更多参考信息
- 支持基于多维度数据的决策
- 保持HR的最终决定权

### 3. 优化协作流程
- 简化多人参与的评估过程
- 提高评估效率和质量
- 增强团队协作透明度

## 测试验证

### 1. 功能测试
- ✅ 数据模型创建和迁移
- ✅ 后端API接口实现
- ✅ 前端界面开发
- ✅ 权限控制验证
- ✅ 状态流转测试

### 2. 用户体验测试
- ✅ 界面交互流畅性
- ✅ 错误处理完整性
- ✅ 响应式设计适配
- ✅ 加载状态显示

### 3. 业务流程测试
- ✅ 完整的共享创建流程
- ✅ 被共享人员评分流程
- ✅ HR查看汇总和决策流程
- ✅ 异常情况处理

## 部署说明

### 1. 数据库迁移
系统启动时会自动创建新的数据表：
- `evaluation_shares` - 绩效共享表
- `share_scores` - 共享评分表
- 更新 `kpi_evaluations` 表结构

### 2. API路由
新增API路由已自动注册：
- 共享管理相关路由
- 共享评分相关路由
- 权限中间件保护

### 3. 前端页面
新增页面文件：
- `/app/shares/page.tsx` - 共享任务列表
- `/app/shares/[id]/page.tsx` - 评分详情
- 更新现有评估页面的共享功能

## 使用指南

### HR用户操作流程：
1. 进入考核管理页面
2. 查看状态为"主管已评估"的考核
3. 点击"查看详情"打开考核详情
4. 切换到"共享评分"Tab
5. 点击"添加共享"创建新的共享
6. 选择评分人员，设置说明和截止时间
7. 查看共享评分汇总，作为最终评分参考

### 被共享人员操作流程：
1. 在左侧导航点击"协助评分"
2. 查看收到的共享评分任务
3. 点击"评分"进入评分详情页面
4. 查看员工信息和现有评分
5. 在"评分详情"Tab中进行评分
6. 单项保存或整体提交评分

## 未来扩展

### 1. 通知系统
- 共享创建通知
- 评分截止提醒
- 完成状态通知

### 2. 评分分析
- 共享评分统计分析
- 评分一致性分析
- 评分质量监控

### 3. 高级权限
- 分级共享权限
- 部门内共享限制
- 评分结果可见性控制

---

## 总结

本次开发成功实现了HR共享评分功能，完整保证了功能的闭环性：
- ✅ 数据模型设计合理
- ✅ 后端API功能完整
- ✅ 前端界面用户友好
- ✅ 权限控制严格
- ✅ 业务流程清晰

该功能在保持原有评估流程完整性的同时，为HR提供了更多的决策支持信息，提升了绩效评估的质量和准确性。