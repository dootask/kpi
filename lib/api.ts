import axios from "axios"

// 获取API基础URL
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api"

// 创建axios实例
const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    "Content-Type": "application/json",
  },
})

// 添加请求拦截器，自动附加token
api.interceptors.request.use(
  config => {
    const token = localStorage.getItem("auth_token")
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => Promise.reject(error)
)

// 响应拦截器
api.interceptors.response.use(
  response => response.data,
  error => {
    console.error("API Error:", error)

    // 处理401错误，触发认证状态更新
    if (error.response?.status === 401) {
      if (typeof window !== "undefined") {
        window.dispatchEvent(new CustomEvent("auth_unauthorized"))
      }
    }
    return Promise.reject(error)
  }
)

// 接口类型定义
export interface Department {
  id: number
  name: string
  description: string
  created_at: string
  employees?: { length: number }
}

export interface Employee {
  id: number
  name: string
  email: string
  position: string
  department_id: number
  manager_id?: number
  role: string
  is_active: boolean
  created_at: string
  department?: { name: string }
  manager?: { name: string }
}

export interface KPITemplate {
  id: number
  name: string
  description: string
  period: string
  is_active: boolean
  created_at: string
  items?: KPIItem[]
}

export interface KPIItem {
  id: number
  template_id: number
  name: string
  description: string
  max_score: number
  order: number
  created_at: string
}

export interface KPIEvaluation {
  id: number
  employee_id: number
  template_id: number
  period: string
  year: number
  month?: number
  quarter?: number
  status: string
  total_score: number
  final_comment: string
  has_shares: boolean
  share_count: number
  created_at: string
  employee?: Employee
  template?: KPITemplate
  scores?: KPIScore[]
  shares?: EvaluationShare[]
}

export interface KPIScore {
  id: number
  evaluation_id: number
  item_id: number
  self_score?: number
  self_comment: string
  manager_score?: number
  manager_comment: string
  hr_score?: number
  hr_comment: string
  final_score?: number
  final_comment: string
  created_at: string
  item?: KPIItem
}

export interface EvaluationShare {
  id: number
  evaluation_id: number
  shared_to_id: number
  shared_by_id: number
  status: string
  message: string
  deadline?: string
  created_at: string
  updated_at: string
  evaluation?: KPIEvaluation
  shared_to?: Employee
  shared_by?: Employee
  scores?: ShareScore[]
}

export interface ShareScore {
  id: number
  share_id: number
  item_id: number
  score?: number
  comment: string
  created_at: string
  updated_at: string
  share?: EvaluationShare
  item?: KPIItem
}

export interface ShareSummary {
  item_id: number
  item_name: string
  average_score: number
  score_count: number
  scores: Array<{
    shared_to: string
    score?: number
    comment: string
  }>
}

export interface DashboardStats {
  total_employees: number
  total_departments: number
  total_evaluations: number
  pending_evaluations: number
  completed_evaluations: number
  average_score: number
  recent_evaluations: KPIEvaluation[]
}

// 统计相关接口
export interface DepartmentStats {
  name: string
  total: number
  completed: number
  pending: number
  avg_score: number
}

export interface MonthlyTrend {
  month: string
  evaluations: number
  avg_score: number
  completion_rate: number
}

export interface ScoreDistribution {
  range: string
  count: number
  color: string
}

export interface TopPerformer {
  name: string
  department: string
  score: number
  evaluations: number
}

export interface RecentEvaluation {
  id: number
  employee: string
  department: string
  template: string
  score: number
  status: string
  period: string
  year: number
  month?: number
  quarter?: number
}

// 统计数据响应类型
export interface StatisticsResponse {
  departmentStats: DepartmentStats[]
  monthlyTrends: MonthlyTrend[]
  scoreDistribution: ScoreDistribution[]
  topPerformers: TopPerformer[]
  recentEvaluations: RecentEvaluation[]
}

// 导出响应类型
export interface ExportResponse {
  file_url: string
  file_name: string
  file_size: number
  message: string
}

// 系统设置请求类型
export interface SystemSettingsRequest {
  allow_registration: boolean
}

// 系统设置响应类型
export interface SystemSettingsResponse {
  allow_registration: boolean
  system_mode: "standalone" | "integrated" // 系统模式，独立模式: standalone，集成模式: integrated
}

// 消息类型
export interface Message {
  type: "success" | "error" | "info" | "warning" | ""
  content: string
}

// 认证相关接口类型
export interface LoginRequest {
  email: string
  password: string
}

// 用户登录（DooTaskToken）请求类型
export interface LoginByDooTaskTokenRequest {
  email: string
  token: string
}

export interface RegisterRequest {
  name: string
  email: string
  password: string
  position: string
  department_id: number
}

export interface LoginResponse {
  token: string
  user: Employee
}

export interface AuthUser {
  id: number
  name: string
  email: string
  position: string
  department_id: number
  manager_id?: number
  role: string
  is_active: boolean
  created_at: string
  department?: { name: string }
  manager?: { name: string }
}

// 部门API
export const departmentApi = {
  getAll: (params?: PaginationParams): Promise<PaginatedResponse<Department>> => api.get("/departments", { params }),
  getById: (id: number): Promise<{ data: Department }> => api.get(`/departments/${id}`),
  create: (data: Omit<Department, "id" | "created_at">): Promise<{ data: Department }> =>
    api.post("/departments", data),
  update: (id: number, data: Partial<Department>): Promise<{ data: Department }> => api.put(`/departments/${id}`, data),
  delete: (id: number): Promise<void> => api.delete(`/departments/${id}`),
}

// 分页响应接口
export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  pageSize: number
  totalPages: number
  hasNext: boolean
  hasPrev: boolean
}

// 分页查询参数接口
export interface PaginationParams {
  page?: number
  pageSize?: number
  search?: string
}

// 评估分页查询参数接口
export interface EvaluationPaginationParams extends PaginationParams {
  status?: string
  employee_id?: string
}

// 员工API
export const employeeApi = {
  getAll: (params?: PaginationParams): Promise<PaginatedResponse<Employee>> => api.get("/employees", { params }),
  getById: (id: number): Promise<{ data: Employee }> => api.get(`/employees/${id}`),
  create: (data: Omit<Employee, "id" | "created_at">): Promise<{ data: Employee }> => api.post("/employees", data),
  update: (id: number, data: Partial<Employee>): Promise<{ data: Employee }> => api.put(`/employees/${id}`, data),
  delete: (id: number): Promise<void> => api.delete(`/employees/${id}`),
  getSubordinates: (id: number): Promise<{ data: Employee[]; total: number }> =>
    api.get(`/employees/${id}/subordinates`),
}

// KPI模板API
export const templateApi = {
  getAll: (): Promise<{ data: KPITemplate[]; total: number }> => api.get("/templates"),
  getById: (id: number): Promise<{ data: KPITemplate }> => api.get(`/templates/${id}`),
  create: (data: Omit<KPITemplate, "id" | "created_at">): Promise<{ data: KPITemplate }> =>
    api.post("/templates", data),
  update: (id: number, data: Partial<KPITemplate>): Promise<{ data: KPITemplate }> => api.put(`/templates/${id}`, data),
  delete: (id: number): Promise<void> => api.delete(`/templates/${id}`),
  getItems: (id: number): Promise<{ data: KPIItem[]; total: number }> => api.get(`/templates/${id}/items`),
}

// KPI项目API
export const itemApi = {
  getById: (id: number): Promise<{ data: KPIItem }> => api.get(`/items/${id}`),
  create: (data: Omit<KPIItem, "id" | "created_at">): Promise<{ data: KPIItem }> => api.post("/items", data),
  update: (id: number, data: Partial<KPIItem>): Promise<{ data: KPIItem }> => api.put(`/items/${id}`, data),
  delete: (id: number): Promise<void> => api.delete(`/items/${id}`),
}

// KPI评估API
export const evaluationApi = {
  getAll: (params?: EvaluationPaginationParams): Promise<PaginatedResponse<KPIEvaluation>> =>
    api.get("/evaluations", { params }),
  getById: (id: number): Promise<{ data: KPIEvaluation }> => api.get(`/evaluations/${id}`),
  create: (data: Omit<KPIEvaluation, "id" | "created_at">): Promise<{ data: KPIEvaluation }> =>
    api.post("/evaluations", data),
  update: (id: number, data: Partial<KPIEvaluation>): Promise<{ data: KPIEvaluation }> =>
    api.put(`/evaluations/${id}`, data),
  delete: (id: number): Promise<void> => api.delete(`/evaluations/${id}`),
  getByEmployee: (employeeId: number): Promise<{ data: KPIEvaluation[]; total: number }> =>
    api.get(`/evaluations/employee/${employeeId}`),
  getPending: (employeeId: number): Promise<{ data: KPIEvaluation[]; total: number }> =>
    api.get(`/evaluations/pending/${employeeId}`),
}

// KPI评分API
export const scoreApi = {
  getByEvaluation: (evaluationId: number): Promise<{ data: KPIScore[]; total: number }> =>
    api.get(`/scores/evaluation/${evaluationId}`),
  updateSelf: (id: number, data: { self_score?: number; self_comment: string }): Promise<{ data: KPIScore }> =>
    api.put(`/scores/${id}/self`, data),
  updateManager: (id: number, data: { manager_score?: number; manager_comment: string }): Promise<{ data: KPIScore }> =>
    api.put(`/scores/${id}/manager`, data),
  updateHR: (id: number, data: { hr_score?: number; hr_comment: string }): Promise<{ data: KPIScore }> =>
    api.put(`/scores/${id}/hr`, data),
  updateFinal: (id: number, data: { final_score?: number; final_comment?: string }): Promise<{ data: KPIScore }> =>
    api.put(`/scores/${id}/final`, data),
}

// 统计API
export const statisticsApi = {
  getDashboard: (params?: {
    year?: string
    period?: string
    month?: string
    quarter?: string
  }): Promise<{ data: DashboardStats }> => api.get("/statistics/dashboard", { params }),
  getDepartment: (id: number): Promise<StatisticsResponse> => api.get("/statistics/department/" + id),
  getEmployee: (id: number): Promise<StatisticsResponse> => api.get("/statistics/employee/" + id),
  getTrends: (period?: string): Promise<StatisticsResponse> => api.get("/statistics/trends", { params: { period } }),
  getData: (params?: {
    year?: string
    period?: string
    month?: string
    quarter?: string
  }): Promise<{ data: StatisticsResponse }> => api.get("/statistics/data", { params }),
}

// 导出API
export const exportApi = {
  evaluation: (id: number): Promise<ExportResponse> => api.get(`/export/evaluation/${id}`),
  department: (id: number): Promise<ExportResponse> => api.get(`/export/department/${id}`),
  period: (period: string, params?: { year?: string; month?: string; quarter?: string }): Promise<ExportResponse> =>
    api.get(`/export/period/${period}`, { params }),
}

// 评论接口类型
export interface EvaluationComment {
  id: number
  evaluation_id: number
  user_id: number
  content: string
  is_private: boolean
  created_at: string
  updated_at: string
  user?: {
    id: number
    name: string
    email: string
    position: string
    department?: {
      name: string
    }
  }
}

// 评论API
export const commentApi = {
  getByEvaluation: (evaluationId: number, params?: PaginationParams): Promise<PaginatedResponse<EvaluationComment>> =>
    api.get(`/evaluations/${evaluationId}/comments`, { params }),
  create: (
    evaluationId: number,
    data: { content: string; is_private: boolean }
  ): Promise<{ data: EvaluationComment }> => api.post(`/evaluations/${evaluationId}/comments`, data),
  update: (
    evaluationId: number,
    commentId: number,
    data: { content: string; is_private: boolean }
  ): Promise<{ data: EvaluationComment }> => api.put(`/evaluations/${evaluationId}/comments/${commentId}`, data),
  delete: (evaluationId: number, commentId: number): Promise<void> =>
    api.delete(`/evaluations/${evaluationId}/comments/${commentId}`),
}

// 认证API
export const authApi = {
  // 用户登录
  login: (data: LoginRequest): Promise<LoginResponse> => api.post("/auth/login", data),

  // 用户登录（DooTaskToken）
  loginByDooTaskToken: (data: LoginByDooTaskTokenRequest): Promise<LoginResponse> => api.post("/auth/login-by-dootask-token", data),

  // 用户注册
  register: (data: RegisterRequest): Promise<LoginResponse> => api.post("/auth/register", data),

  // 获取当前用户信息
  getCurrentUser: (): Promise<{ data: AuthUser }> => api.get("/me"),

  // 刷新token
  refreshToken: (): Promise<{ token: string }> => api.post("/auth/refresh"),

  // 获取部门列表（公开接口，用于注册）
  getDepartments: (): Promise<{ data: Department[] }> => api.get("/auth/departments"),

  // 登出（清除本地token）
  logout: () => {
    localStorage.removeItem("auth_token")
    localStorage.removeItem("user_info")
  },

  // 检查是否已认证
  isAuthenticated: (): boolean => {
    const token = localStorage.getItem("auth_token")
    return !!token
  },

  // 获取当前用户token
  getToken: (): string | null => {
    return localStorage.getItem("auth_token")
  },

  // 设置用户token和信息
  setAuth: (token: string, user: AuthUser) => {
    localStorage.setItem("auth_token", token)
    localStorage.setItem("user_info", JSON.stringify(user))
  },

  // 获取用户信息
  getUser: (): AuthUser | null => {
    const userInfo = localStorage.getItem("user_info")
    return userInfo ? JSON.parse(userInfo) : null
  },
}

// 共享API
export const shareApi = {
  // 创建共享
  create: (evaluationId: number, data: { shared_to_ids: number[]; message: string; deadline?: string }): Promise<{ data: EvaluationShare[]; message: string }> =>
    api.post(`/evaluations/${evaluationId}/shares`, data),

  // 获取评估的共享列表
  getByEvaluation: (evaluationId: number): Promise<{ data: EvaluationShare[] }> =>
    api.get(`/evaluations/${evaluationId}/shares`),

  // 删除共享
  delete: (evaluationId: number, shareId: number): Promise<{ message: string }> =>
    api.delete(`/evaluations/${evaluationId}/shares/${shareId}`),

  // 获取共享评分汇总
  getSummary: (evaluationId: number): Promise<{ data: ShareSummary[] }> =>
    api.get(`/evaluations/${evaluationId}/share-summary`),

  // 获取我的共享任务
  getMy: (): Promise<{ data: EvaluationShare[] }> =>
    api.get("/shares/my"),

  // 获取共享详情
  getDetail: (shareId: number): Promise<{ data: EvaluationShare }> =>
    api.get(`/shares/${shareId}`),

  // 获取共享评分
  getScores: (shareId: number): Promise<{ data: ShareScore[] }> =>
    api.get(`/shares/${shareId}/scores`),

  // 更新共享评分
  updateScore: (shareId: number, itemId: number, data: { score?: number; comment: string }): Promise<{ data: ShareScore; message: string }> =>
    api.put(`/shares/${shareId}/scores/${itemId}`, data),

  // 提交共享评分
  submit: (shareId: number): Promise<{ message: string }> =>
    api.post(`/shares/${shareId}/submit`),
}

// 系统设置API
export const settingsApi = {
  // 获取系统设置
  get: (): Promise<{ data: SystemSettingsResponse }> => api.get("/settings"),

  // 更新系统设置
  update: (data: SystemSettingsRequest): Promise<{ data: SystemSettingsResponse; message: string }> =>
    api.put("/settings", data),
}

export default api
