"use client"

import { useState, useEffect, useMemo, useRef } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Textarea } from "@/components/ui/textarea"
import {
  Plus,
  Eye,
  Award,
  CheckCircle,
  Clock,
  Star,
  Edit2,
  Save,
  X,
  MessageCircle,
  Lock,
  Globe,
  Trash2,
} from "lucide-react"
import {
  evaluationApi,
  scoreApi,
  employeeApi,
  templateApi,
  commentApi,
  type KPIEvaluation,
  type KPIScore,
  type Employee,
  type KPITemplate,
  type EvaluationComment,
} from "@/lib/api"
import { useAuth } from "@/lib/auth-context"
import { useAppContext } from "@/lib/app-context"
import { getPeriodValue } from "@/lib/utils"
import { EmployeeSelector } from "@/components/employee-selector"

export default function EvaluationsPage() {
  const { Alert, Confirm, getStatusBadge } = useAppContext()
  const { user: currentUser, isManager, isHR } = useAuth()
  const detailsRef = useRef<HTMLDivElement>(null)
  const [evaluations, setEvaluations] = useState<KPIEvaluation[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [templates, setTemplates] = useState<KPITemplate[]>([])
  const [dialogOpen, setDialogOpen] = useState(false)
  const [scoreDialogOpen, setScoreDialogOpen] = useState(false)
  const [selectedEvaluation, setSelectedEvaluation] = useState<KPIEvaluation | null>(null)
  const [scores, setScores] = useState<KPIScore[]>([])
  const [activeTab, setActiveTab] = useState("details") // 新增：标签页状态
  const [editingScore, setEditingScore] = useState<number | null>(null) // 新增：正在编辑的评分ID
  const [tempScore, setTempScore] = useState<string>("") // 新增：临时评分
  const [tempComment, setTempComment] = useState<string>("") // 新增：临时评论
  const [isSubmittingSelfEvaluation, setIsSubmittingSelfEvaluation] = useState(false) // 新增：自评提交状态
  const [loading, setLoading] = useState(false) // 新增：通用加载状态
  const [error, setError] = useState<string | null>(null) // 新增：错误状态

  // 绩效评论相关状态
  const [comments, setComments] = useState<EvaluationComment[]>([]) // 评论列表
  const [isLoadingComments, setIsLoadingComments] = useState<boolean>(false) // 是否正在加载评论
  const [newComment, setNewComment] = useState<string>("") // 新评论内容
  const [newCommentPrivate, setNewCommentPrivate] = useState<boolean>(false) // 新评论是否私有
  const [isAddingComment, setIsAddingComment] = useState<boolean>(false) // 是否正在添加评论
  const [isSavingComment, setIsSavingComment] = useState<boolean>(false) // 是否正在保存评论
  const [editingCommentId, setEditingCommentId] = useState<number | null>(null) // 正在编辑的评论ID
  const [editingCommentContent, setEditingCommentContent] = useState<string>("") // 编辑中的评论内容
  const [editingCommentPrivate, setEditingCommentPrivate] = useState<boolean>(false) // 编辑中的评论是否私有
  const [formData, setFormData] = useState({
    employee_ids: [] as string[],
    template_id: "",
    period: "monthly",
    year: new Date().getFullYear(),
    month: new Date().getMonth() + 1,
    quarter: Math.floor(new Date().getMonth() / 3) + 1,
  })

  // 获取评估列表
  const fetchEvaluations = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await evaluationApi.getAll()
      setEvaluations(response.data || [])
    } catch (error) {
      console.error("获取评估列表失败:", error)
      setError("获取评估列表失败，请刷新重试")
    } finally {
      setLoading(false)
    }
  }

  // 获取员工列表
  const fetchEmployees = async () => {
    try {
      const response = await employeeApi.getAll()
      setEmployees(response.data || [])
    } catch (error) {
      console.error("获取员工列表失败:", error)
      setEmployees([])
    }
  }

  // 获取模板列表
  const fetchTemplates = async () => {
    try {
      const response = await templateApi.getAll()
      setTemplates(response.data || [])
    } catch (error) {
      console.error("获取模板列表失败:", error)
      setTemplates([])
    }
  }

  // 获取评估详情和分数
  const fetchEvaluationScores = async (evaluationId: number) => {
    try {
      const response = await scoreApi.getByEvaluation(evaluationId)
      setScores(response.data || [])
    } catch (error) {
      console.error("获取评估详情失败:", error)
    }
  }

  useEffect(() => {
    fetchEvaluations()
    fetchEmployees()
    fetchTemplates()
  }, [])

  // 创建新评估
  const handleCreateEvaluation = async (e: React.FormEvent) => {
    e.preventDefault()

    // 验证是否至少选择了一个员工
    if (formData.employee_ids.length === 0) {
      Alert("验证失败", "请至少选择一个员工进行考核")
      return
    }

    // 验证是否选择了模板
    if (!formData.template_id) {
      Alert("验证失败", "请选择考核模板")
      return
    }

    try {
      // 为每个选中的员工创建评估
      const promises = formData.employee_ids.map(employeeId =>
        evaluationApi.create({
          employee_id: parseInt(employeeId),
          template_id: parseInt(formData.template_id),
          period: formData.period,
          year: formData.year,
          month: formData.period === "monthly" ? formData.month : undefined,
          quarter: formData.period === "quarterly" ? formData.quarter : undefined,
          status: "pending",
          total_score: 0,
          final_comment: "",
        })
      )

      await Promise.all(promises)

      fetchEvaluations()
      setDialogOpen(false)
      setFormData({
        employee_ids: [],
        template_id: "",
        period: "monthly",
        year: new Date().getFullYear(),
        month: new Date().getMonth() + 1,
        quarter: Math.floor(new Date().getMonth() / 3) + 1,
      })

      // 成功提示
      Alert("创建成功", `已为 ${formData.employee_ids.length} 个员工创建考核`)
    } catch (error) {
      console.error("创建评估失败:", error)
      Alert("创建失败", "创建考核失败，请重试")
    }
  }

  // 开始编辑评分
  const handleStartEdit = (scoreId: number, currentScore?: number, currentComment?: string) => {
    setEditingScore(scoreId)
    setTempScore(currentScore ? currentScore.toString() : "")
    setTempComment(currentComment || "")
  }

  // 取消编辑评分
  const handleCancelEdit = () => {
    setEditingScore(null)
    setTempScore("")
    setTempComment("")
  }

  // 验证评分范围
  const validateScore = (score: string, maxScore: number): { isValid: boolean; message?: string } => {
    if (score === "") {
      return { isValid: false, message: "请输入评分" }
    }

    const numScore = parseFloat(score)
    if (isNaN(numScore)) {
      return { isValid: false, message: "请输入有效的数字" }
    }

    if (numScore < 0) {
      return { isValid: false, message: "评分不能小于0" }
    }

    if (numScore > maxScore) {
      return { isValid: false, message: `评分不能超过${maxScore}分` }
    }

    return { isValid: true }
  }

  // 处理输入值变化
  const handleScoreChange = (value: string, maxScore: number) => {
    setTempScore(value)

    // 实时验证并限制输入
    if (value !== "") {
      const numValue = parseFloat(value)
      if (!isNaN(numValue)) {
        if (numValue < 0) {
          setTempScore("0")
        } else if (numValue > maxScore) {
          setTempScore(maxScore.toString())
        }
      }
    }
  }

  // 找到下一个未评分的项目
  const findNextUnscored = (currentScoreId: number, type: "self" | "manager"): number | null => {
    const currentIndex = scores.findIndex(s => s.id === currentScoreId)
    if (currentIndex === -1) return null

    // 从当前项目的下一个开始查找
    for (let i = currentIndex + 1; i < scores.length; i++) {
      const score = scores[i]
      if (type === "self" && (!score.self_score || score.self_score === 0)) {
        return score.id
      }
      if (type === "manager" && (!score.manager_score || score.manager_score === 0)) {
        return score.id
      }
    }

    // 如果没找到，从头开始查找
    for (let i = 0; i < currentIndex; i++) {
      const score = scores[i]
      if (type === "self" && (!score.self_score || score.self_score === 0)) {
        return score.id
      }
      if (type === "manager" && (!score.manager_score || score.manager_score === 0)) {
        return score.id
      }
    }

    return null
  }

  // 滚动到指定的评分项目
  const scrollToNextUnscored = (currentScoreId: number, type?: "self" | "manager") => {
    let nextUnscored: number | null = currentScoreId
    if (type) {
      nextUnscored = findNextUnscored(currentScoreId, type)
      if (!nextUnscored) {
        return
      }
    }

    // 使用 setTimeout 确保DOM已更新
    requestAnimationFrame(() => {
      const element = detailsRef.current?.querySelector(`[data-score-id="${nextUnscored}"]`) as HTMLElement
      if (!element) {
        return
      }
      element.scrollIntoView({
        behavior: "smooth",
        block: "center",
        inline: "nearest",
      })
      // 添加一个视觉提示
      element.style.backgroundColor = "#f0f9ff"
      element.style.transition = "background-color 0.3s ease"
      setTimeout(() => {
        element.style.backgroundColor = ""
      }, 2000)
    })
  }

  // 保存评分
  const handleSaveScore = async (scoreId: number, type: "self" | "manager") => {
    try {
      // 获取当前编辑的评分项目
      const currentScore = scores.find(s => s.id === scoreId)
      if (!currentScore) {
        Alert("保存失败", "评分项目不存在")
        return
      }

      const maxScore = currentScore.item?.max_score || 100
      const validation = validateScore(tempScore, maxScore)

      if (!validation.isValid) {
        Alert("输入错误", validation.message || "评分输入无效")
        return
      }

      const scoreValue = parseFloat(tempScore)
      if (type === "self") {
        await scoreApi.updateSelf(scoreId, { self_score: scoreValue, self_comment: tempComment })
      } else if (type === "manager") {
        await scoreApi.updateManager(scoreId, { manager_score: scoreValue, manager_comment: tempComment })
      }

      if (selectedEvaluation) {
        await fetchEvaluationScores(selectedEvaluation.id)
        fetchEvaluations()
      }

      setEditingScore(null)
      setTempScore("")
      setTempComment("")
      scrollToNextUnscored(scoreId, type)
    } catch (error) {
      console.error("更新评分失败:", error)
      Alert("保存失败", "更新评分失败，请重试")
    }
  }

  // 完成阶段
  const handleCompleteStage = async (evaluationId: number, stage: string) => {
    // 验证状态流转
    if (selectedEvaluation) {
      const validationError = validateStageTransition(selectedEvaluation, stage)
      if (validationError) {
        Alert("验证失败", validationError)
        return
      }
    }

    // 自评阶段的特殊处理
    if (stage === "self") {
      // 检查是否所有项目都已自评
      const uncompletedItems = scores.filter(score => !score.self_score || score.self_score === 0)
      if (uncompletedItems.length > 0) {
        await Alert("自评", `请先完成所有项目的自评。还有 ${uncompletedItems.length} 个项目未评分。`)
        scrollToNextUnscored(uncompletedItems[0].id)
        return
      }

      // 确认提交自评
      const result = await Confirm("自评", "确定要提交自评吗？提交后将无法修改。")
      if (!result) {
        return
      }

      setIsSubmittingSelfEvaluation(true)
    }

    // 上级评分阶段的特殊处理
    if (stage === "manager") {
      // 检查是否所有项目都已进行主管评分
      const uncompletedItems = scores.filter(score => !score.manager_score || score.manager_score === 0)
      if (uncompletedItems.length > 0) {
        await Alert("主管评分", `请先完成所有项目的主管评分。还有 ${uncompletedItems.length} 个项目未评分。`)
        scrollToNextUnscored(uncompletedItems[0].id)
        return
      }

      // 确认提交主管评分
      const result = await Confirm("主管评分", "确定要提交主管评分吗？提交后将无法修改，评估将进入HR审核阶段。")
      if (!result) {
        return
      }
    }

    // HR审核阶段的特殊处理
    if (stage === "hr") {
      // 检查是否所有项目都已确定最终得分
      const unconfirmedItems = scores.filter(score => !score.final_score && !score.manager_score)
      if (unconfirmedItems.length > 0) {
        await Alert("HR审核", `请先确认所有项目的最终得分。还有 ${unconfirmedItems.length} 个项目待确认。`)
        scrollToNextUnscored(unconfirmedItems[0].id)
        return
      }

      // 确认完成HR审核
      const result = await Confirm("HR审核", "确定要完成HR审核吗？提交后将无法再修改，评估将进入员工确认阶段。")
      if (!result) {
        return
      }
    }

    // 员工最后确认最终得分
    if (stage === "confirm") {
      // 检查是否所有项目都已确认最终得分
      const alreadyConfirmed = scores.find(score => score.final_score)
      if (alreadyConfirmed) {
        Alert("确认最终得分", "已确认最终得分，无法再修改。")
        return
      }

      // 确认最终得分
      const result = await Confirm("确认最终得分", "确定要确认最终得分吗？确认后将无法再修改。")
      if (!result) {
        return
      }
    }

    try {
      let newStatus = ""
      switch (stage) {
        case "self":
          newStatus = isManager ? "manager_evaluated" : "self_evaluated"
          break
        case "manager":
          newStatus = "manager_evaluated"
          break
        case "hr":
          newStatus = "pending_confirm"
          break
        case "confirm":
          newStatus = "completed"
          break
      }

      // 计算并更新总分
      let totalScore = 0
      switch (stage) {
        case "self":
          // 自评完成后，总分为自评分数总和
          totalScore = scores.reduce((acc, score) => acc + (score.self_score || 0), 0)
          break
        case "manager":
          // 主管评分完成后，总分为主管评分总和
          totalScore = scores.reduce((acc, score) => acc + (score.manager_score || 0), 0)
          break
        case "hr":
        case "confirm":
          // HR审核或员工确认最终得分后，总分为最终得分总和
          totalScore = scores.reduce((acc, score) => acc + (score.final_score || score.manager_score || 0), 0)
          break
      }

      await evaluationApi.update(evaluationId, {
        status: newStatus,
        total_score: totalScore,
      })

      if (stage === "confirm") {
        // 员工确认最终得分后，更新最终得分
        setScores(scores =>
          scores.map(s => ({
            ...s,
            final_score: s.manager_score ?? s.self_score,
          }))
        )
      }

      fetchEvaluations()
      if (selectedEvaluation) {
        setSelectedEvaluation({
          ...selectedEvaluation,
          status: newStatus,
          total_score: totalScore,
        })
      }

      // 成功提示
      if (stage === "self") {
        await Alert("自评", "自评提交成功！请等待上级主管评分。")
      } else if (stage === "manager") {
        await Alert("主管评分", "主管评分提交成功！评估已转入HR审核阶段。")
      } else if (stage === "hr") {
        await Alert("HR审核", "HR审核完成！请等待员工确认最终得分。")
      } else if (stage === "confirm") {
        await Alert("确认最终得分", "最终得分确认成功！绩效评估已正式结束。")
      }
    } catch (error) {
      console.error("更新状态失败:", error)
      Alert("提交失败", "提交失败，请重试。")
    } finally {
      if (stage === "self") {
        setIsSubmittingSelfEvaluation(false)
      }
    }
  }

  // 获取评论列表
  const fetchComments = async (evaluationId: number) => {
    try {
      setIsLoadingComments(true)
      const response = await commentApi.getByEvaluation(evaluationId)
      setComments(response.data || [])
    } catch (error) {
      console.error("获取评论失败:", error)
      setComments([])
    } finally {
      setIsLoadingComments(false)
    }
  }

  // 添加评论
  const handleAddComment = async () => {
    if (!selectedEvaluation || !newComment.trim()) return

    try {
      setIsSavingComment(true)
      const response = await commentApi.create(selectedEvaluation.id, {
        content: newComment,
        is_private: newCommentPrivate,
      })

      setComments([response.data, ...comments])
      setNewComment("")
      setNewCommentPrivate(false)
      setIsAddingComment(false)
      Alert("添加成功", "评论已添加")
    } catch (error) {
      console.error("添加评论失败:", error)
      Alert("添加失败", "添加评论失败，请重试")
    } finally {
      setIsSavingComment(false)
    }
  }

  // 开始编辑评论
  const handleStartEditComment = (comment: EvaluationComment) => {
    setEditingCommentId(comment.id)
    setEditingCommentContent(comment.content)
    setEditingCommentPrivate(comment.is_private)
  }

  // 保存编辑的评论
  const handleSaveEditComment = async (commentId: number) => {
    if (!selectedEvaluation || !editingCommentContent.trim()) return

    try {
      setIsSavingComment(true)
      const response = await commentApi.update(selectedEvaluation.id, commentId, {
        content: editingCommentContent,
        is_private: editingCommentPrivate,
      })

      setComments(comments.map(c => (c.id === commentId ? response.data : c)))
      setEditingCommentId(null)
      setEditingCommentContent("")
      setEditingCommentPrivate(false)
      Alert("保存成功", "评论已更新")
    } catch (error) {
      console.error("更新评论失败:", error)
      Alert("保存失败", "更新评论失败，请重试")
    } finally {
      setIsSavingComment(false)
    }
  }

  // 取消编辑评论
  const handleCancelEditComment = () => {
    setEditingCommentId(null)
    setEditingCommentContent("")
    setEditingCommentPrivate(false)
  }

  // 删除评论
  const handleDeleteComment = async (commentId: number) => {
    if (!selectedEvaluation) return

    const confirmed = await Confirm("确认删除", "确定要删除这条评论吗？此操作无法撤销。")
    if (!confirmed) return

    try {
      await commentApi.delete(selectedEvaluation.id, commentId)
      setComments(comments.filter(c => c.id !== commentId))
      Alert("删除成功", "评论已删除")
    } catch (error) {
      console.error("删除评论失败:", error)
      Alert("删除失败", "删除评论失败，请重试")
    }
  }

  // 查看详情
  const handleViewDetails = (evaluation: KPIEvaluation) => {
    setSelectedEvaluation(evaluation)
    fetchEvaluationScores(evaluation.id)
    fetchComments(evaluation.id)
    setScoreDialogOpen(true)
    setActiveTab("details")

    // 重置评论状态
    setComments([])
    setNewComment("")
    setNewCommentPrivate(false)
    setIsAddingComment(false)
    setEditingCommentId(null)
    setEditingCommentContent("")
    setEditingCommentPrivate(false)
  }

  // 排序评估列表
  const sortEvaluations = (evaluations: KPIEvaluation[], status?: string) => {
    const sorts = ["pending", "pending_confirm"]
    if (status) {
      sorts.unshift(status)
    }
    return evaluations.sort((a, b) => {
      const aIndex = sorts.indexOf(a.status)
      const bIndex = sorts.indexOf(b.status)

      // 如果都在sorts数组中，按照sorts数组的顺序排序
      if (aIndex !== -1 && bIndex !== -1) {
        return aIndex - bIndex
      }

      // 如果只有一个在sorts数组中，在sorts中的排前面
      if (aIndex !== -1 && bIndex === -1) return -1
      if (aIndex === -1 && bIndex !== -1) return 1

      // 如果都不在sorts数组中，按照id排序
      return a.id - b.id
    })
  }

  // 根据用户角色过滤评估
  const getFilteredEvaluations = useMemo(() => {
    if (!currentUser) return []

    if (isHR) {
      return sortEvaluations(evaluations, "manager_evaluated") // HR可以看到所有评估
    }
    return sortEvaluations(
      evaluations.filter(evaluation => {
        if (isManager) {
          // 主管可以看到自己的考核 + 下属的考核
          return evaluation.employee_id === currentUser.id || evaluation.employee?.manager_id === currentUser.id
        } else {
          // 员工只能看到自己的评估
          return evaluation.employee_id === currentUser.id
        }
      }),
      isManager ? "self_evaluated" : undefined
    )
  }, [currentUser, evaluations, isHR, isManager])

  // 检查是否可以进行某个操作
  const canPerformAction = (evaluation: KPIEvaluation, action: "self" | "manager" | "hr" | "confirm") => {
    if (!currentUser) return false

    switch (action) {
      case "self":
        // 任何人都可以对自己的考核进行自评（包括主管）
        return evaluation.status === "pending" && evaluation.employee_id === currentUser.id
      case "manager":
        // 主管只能评估自己直接下属的员工，但不能评估自己
        return (
          evaluation.status === "self_evaluated" &&
          isManager &&
          evaluation.employee?.manager_id === currentUser.id &&
          evaluation.employee_id !== currentUser.id
        )
      case "hr":
        return evaluation.status === "manager_evaluated" && isHR
      case "confirm":
        return evaluation.status === "pending_confirm" && evaluation.employee_id === currentUser.id
      default:
        return false
    }
  }

  // 获取状态流转进度
  const getStatusProgress = (status: string) => {
    const statusMap = {
      pending: { step: 1, total: 4, label: "等待自评" },
      self_evaluated: { step: 2, total: 4, label: "等待主管评估" },
      manager_evaluated: { step: 3, total: 4, label: "等待HR审核" },
      pending_confirm: { step: 4, total: 4, label: "等待确认" },
      completed: { step: 4, total: 4, label: "已完成" },
    }
    return statusMap[status as keyof typeof statusMap] || { step: 0, total: 4, label: "未知状态" }
  }

  // 验证评估是否可以进入下一阶段
  const validateStageTransition = (evaluation: KPIEvaluation, stage: string): string | null => {
    const currentDate = new Date()
    const evaluationDate = new Date(evaluation.created_at)
    const daysDiff = Math.floor((currentDate.getTime() - evaluationDate.getTime()) / (1000 * 3600 * 24))

    // 检查评估是否已过期（示例：30天后过期）
    if (daysDiff > 30) {
      return "评估已过期，无法继续流转。请联系HR处理。"
    }

    // 检查用户权限和状态匹配
    if (!canPerformAction(evaluation, stage as "self" | "manager" | "hr")) {
      return "您没有权限进行此操作，或评估状态不匹配。"
    }

    return null // 验证通过
  }

  return (
    <div className="space-y-6">
      {/* 响应式头部 */}
      <div className="flex flex-col space-y-4 sm:flex-row sm:items-center sm:justify-between sm:space-y-0">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold text-foreground">考核管理</h1>
          <p className="text-muted-foreground mt-1 sm:mt-2">管理员工绩效考核流程</p>
        </div>
        {isHR && (
          <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
            <DialogTrigger asChild>
              <Button className="w-full sm:w-auto">
                <Plus className="w-4 h-4 mr-2" />
                创建考核
              </Button>
            </DialogTrigger>
            <DialogContent className="w-[95vw] sm:max-w-md mx-auto">
              <DialogHeader>
                <DialogTitle>创建新考核</DialogTitle>
              </DialogHeader>
              <form onSubmit={handleCreateEvaluation} className="space-y-4">
                <EmployeeSelector
                  employees={employees}
                  selectedEmployeeIds={formData.employee_ids}
                  onSelectionChange={employeeIds => setFormData(prev => ({ ...prev, employee_ids: employeeIds }))}
                  label="员工"
                  placeholder="选择员工..."
                  maxDisplayTags={5}
                />
                <div className="flex flex-col gap-2">
                  <Label htmlFor="template">考核模板</Label>
                  <Select
                    value={formData.template_id}
                    onValueChange={value => setFormData({ ...formData, template_id: value })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="选择模板" />
                    </SelectTrigger>
                    <SelectContent>
                      {templates.map(template => (
                        <SelectItem key={template.id} value={template.id.toString()}>
                          {template.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="flex flex-col gap-2">
                  <Label htmlFor="period">考核周期</Label>
                  <Select value={formData.period} onValueChange={value => setFormData({ ...formData, period: value })}>
                    <SelectTrigger>
                      <SelectValue placeholder="选择周期" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="monthly">月度</SelectItem>
                      <SelectItem value="quarterly">季度</SelectItem>
                      <SelectItem value="yearly">年度</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="flex flex-col gap-2">
                  <Label htmlFor="year">年份</Label>
                  <Select
                    value={formData.year.toString()}
                    onValueChange={value => setFormData({ ...formData, year: parseInt(value) })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="选择年份" />
                    </SelectTrigger>
                    <SelectContent>
                      {Array.from({ length: 10 }, (_, i) => {
                        const year = new Date().getFullYear() - i
                        return (
                          <SelectItem key={year} value={year.toString()}>
                            {year}年
                          </SelectItem>
                        )
                      })}
                    </SelectContent>
                  </Select>
                </div>
                {formData.period === "monthly" && (
                  <div className="flex flex-col gap-2">
                    <Label htmlFor="month">月份</Label>
                    <Select
                      value={formData.month.toString()}
                      onValueChange={value => setFormData({ ...formData, month: parseInt(value) })}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="选择月份" />
                      </SelectTrigger>
                      <SelectContent>
                        {[...Array(12)].map((_, i) => (
                          <SelectItem key={i + 1} value={(i + 1).toString()}>
                            {i + 1}月
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                )}
                {formData.period === "quarterly" && (
                  <div className="flex flex-col gap-2">
                    <Label htmlFor="quarter">季度</Label>
                    <Select
                      value={formData.quarter.toString()}
                      onValueChange={value => setFormData({ ...formData, quarter: parseInt(value) })}
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="选择季度" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="1">第一季度</SelectItem>
                        <SelectItem value="2">第二季度</SelectItem>
                        <SelectItem value="3">第三季度</SelectItem>
                        <SelectItem value="4">第四季度</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                )}
                <div className="flex flex-col-reverse sm:flex-row sm:justify-end gap-2 sm:space-x-2 sm:gap-0">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setDialogOpen(false)}
                    className="w-full sm:w-auto"
                  >
                    取消
                  </Button>
                  <Button type="submit" className="w-full sm:w-auto">
                    创建
                  </Button>
                </div>
              </form>
            </DialogContent>
          </Dialog>
        )}
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs sm:text-sm font-medium">总评估数</CardTitle>
            <Award className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-xl sm:text-2xl font-bold">{getFilteredEvaluations.length}</div>
            <p className="text-xs text-muted-foreground">全部考核项目</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs sm:text-sm font-medium">待处理</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-xl sm:text-2xl font-bold">
              {
                getFilteredEvaluations.filter(e =>
                  ["pending", "self_evaluated", "manager_evaluated", "pending_confirm"].includes(e.status)
                ).length
              }
            </div>
            <p className="text-xs text-muted-foreground">需要处理的考核</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs sm:text-sm font-medium">已完成</CardTitle>
            <CheckCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-xl sm:text-2xl font-bold">
              {getFilteredEvaluations.filter(e => e.status === "completed").length}
            </div>
            <p className="text-xs text-muted-foreground">已完成的考核</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs sm:text-sm font-medium">平均分</CardTitle>
            <Star className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-xl sm:text-2xl font-bold">
              {getFilteredEvaluations.length > 0
                ? Math.round(
                    getFilteredEvaluations.reduce((acc, e) => acc + e.total_score, 0) / getFilteredEvaluations.length
                  )
                : 0}
            </div>
            <p className="text-xs text-muted-foreground">总体考核平均分</p>
          </CardContent>
        </Card>
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="text-red-800 text-sm">⚠️ {error}</div>
        </div>
      )}

      {/* 评估列表 */}
      <Card>
        <CardHeader>
          <CardTitle>考核列表</CardTitle>
        </CardHeader>
        <CardContent>
          {loading && (
            <div className="text-center py-8">
              <div className="text-gray-500">加载中...</div>
            </div>
          )}
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>员工</TableHead>
                <TableHead>部门</TableHead>
                <TableHead>考核模板</TableHead>
                <TableHead>周期</TableHead>
                <TableHead>总分</TableHead>
                <TableHead>状态</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {getFilteredEvaluations.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center py-8 text-gray-500">
                    暂无考核数据
                  </TableCell>
                </TableRow>
              ) : (
                getFilteredEvaluations.map(evaluation => (
                  <TableRow key={evaluation.id}>
                    <TableCell className="font-medium">
                      {evaluation.employee?.name}
                      <div className="text-sm text-muted-foreground">{evaluation.employee?.position}</div>
                    </TableCell>
                    <TableCell>{evaluation.employee?.department?.name}</TableCell>
                    <TableCell>{evaluation.template?.name}</TableCell>
                    <TableCell>{getPeriodValue(evaluation)}</TableCell>
                    <TableCell>
                      <div className="text-lg font-semibold">{evaluation.total_score}</div>
                    </TableCell>
                    <TableCell>{getStatusBadge(evaluation.status)}</TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end space-x-2">
                        <Button variant="outline" size="sm" onClick={() => handleViewDetails(evaluation)}>
                          <Eye className="w-4 h-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* 评分详情对话框 */}
      <Dialog open={scoreDialogOpen} onOpenChange={setScoreDialogOpen}>
        <DialogContent className="w-[95vw] sm:max-w-6xl max-h-[90vh] flex flex-col">
          <DialogHeader>
            <DialogTitle>考核详情 - {selectedEvaluation?.employee?.name}</DialogTitle>
          </DialogHeader>
          {selectedEvaluation && (
            <>
              {/* 可滚动的内容区域 */}
              <div className="flex-1 overflow-y-auto space-y-4 -mx-6 px-6">
                {/* 基本信息卡片 */}
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                  <div className="bg-gray-50 p-3 rounded">
                    <Label className="text-sm text-gray-500">员工姓名</Label>
                    <p className="text-sm font-medium">{selectedEvaluation.employee?.name}</p>
                  </div>
                  <div className="bg-gray-50 p-3 rounded">
                    <Label className="text-sm text-gray-500">考核模板</Label>
                    <p className="text-sm font-medium">{selectedEvaluation.template?.name}</p>
                  </div>
                  <div className="bg-gray-50 p-3 rounded">
                    <Label className="text-sm text-gray-500">考核周期</Label>
                    <p className="text-sm font-medium">{getPeriodValue(selectedEvaluation)}</p>
                  </div>
                  <div className="bg-gray-50 p-3 rounded">
                    <Label className="text-sm text-gray-500">当前状态</Label>
                    <div className="mt-1 space-y-2">
                      {getStatusBadge(selectedEvaluation.status)}
                      {/* 状态进度条 */}
                      <div className="text-xs text-gray-500">
                        <div className="flex justify-between items-center mb-1">
                          <span>流程进度</span>
                          <span>
                            {getStatusProgress(selectedEvaluation.status).step}/
                            {getStatusProgress(selectedEvaluation.status).total}
                          </span>
                        </div>
                        <div className="w-full bg-gray-200 rounded-full h-1.5">
                          <div
                            className="bg-blue-600 h-1.5 rounded-full transition-all duration-300"
                            style={{
                              width: `${
                                (getStatusProgress(selectedEvaluation.status).step /
                                  getStatusProgress(selectedEvaluation.status).total) *
                                100
                              }%`,
                            }}
                          />
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                {/* 标签页 */}
                <Tabs value={activeTab} onValueChange={setActiveTab}>
                  <TabsList className="grid w-full grid-cols-2">
                    <TabsTrigger value="details">评分详情</TabsTrigger>
                    <TabsTrigger value="summary">总结汇总</TabsTrigger>
                  </TabsList>

                  <TabsContent value="details" className="space-y-4" ref={detailsRef}>
                    {/* 自评指导和进度信息 */}
                    {canPerformAction(selectedEvaluation, "self") && (
                      <div className="space-y-4">
                        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                          <h4 className="font-medium text-blue-900 mb-2">📝 自评指导</h4>
                          <ul className="text-sm text-blue-800 space-y-1">
                            <li>• 请根据本期间的实际工作表现进行客观评分</li>
                            <li>• 评分需要在0到满分之间，建议结合具体工作成果</li>
                            <li>• 请在评价说明中详细描述您的工作亮点和改进计划</li>
                            <li>• 完成所有项目评分后，点击&quot;完成自评&quot;提交</li>
                          </ul>
                        </div>

                        {/* 评分进度 */}
                        <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                          <h4 className="font-medium text-green-900 mb-2">📊 评分进度</h4>
                          <div className="flex items-center justify-between">
                            <span className="text-sm text-green-800">
                              已完成 {scores.filter(s => s.self_score && s.self_score > 0).length} / {scores.length} 项
                            </span>
                            <div className="flex-1 mx-4 bg-green-200 rounded-full h-2">
                              <div
                                className="bg-green-600 h-2 rounded-full transition-all duration-300"
                                style={{
                                  width: `${
                                    scores.length > 0
                                      ? (scores.filter(s => s.self_score && s.self_score > 0).length / scores.length) *
                                        100
                                      : 0
                                  }%`,
                                }}
                              />
                            </div>
                            <span className="text-sm font-medium text-green-900">
                              {scores.length > 0
                                ? Math.round(
                                    (scores.filter(s => s.self_score && s.self_score > 0).length / scores.length) * 100
                                  )
                                : 0}
                              %
                            </span>
                          </div>
                        </div>
                      </div>
                    )}

                    {/* 上级评分指导信息 */}
                    {canPerformAction(selectedEvaluation, "manager") && (
                      <div className="space-y-4">
                        <div className="bg-purple-50 border border-purple-200 rounded-lg p-4">
                          <h4 className="font-medium text-purple-900 mb-2">👔 上级评分指导</h4>
                          <ul className="text-sm text-purple-800 space-y-1">
                            <li>• 请结合员工的自评内容和实际工作表现进行评分</li>
                            <li>• 评分应客观公正，既要认可成绩，也要指出不足</li>
                            <li>• 在评价说明中提供具体的改进建议和发展方向</li>
                            <li>• 完成所有项目评分后，点击&quot;完成主管评估&quot;提交</li>
                          </ul>
                        </div>

                        {/* 评分对比和进度 */}
                        <div className="bg-orange-50 border border-orange-200 rounded-lg p-4">
                          <h4 className="font-medium text-orange-900 mb-2">📈 评分对比</h4>
                          <div className="grid grid-cols-2 gap-4 text-sm">
                            <div>
                              <span className="text-orange-800">员工自评总分：</span>
                              <span className="font-semibold text-orange-900">
                                {scores.reduce((acc, score) => acc + (score.self_score || 0), 0)} 分
                              </span>
                            </div>
                            <div>
                              <span className="text-orange-800">主管评分进度：</span>
                              <span className="font-semibold text-orange-900">
                                {scores.filter(s => s.manager_score && s.manager_score > 0).length} / {scores.length} 项
                              </span>
                            </div>
                          </div>
                          <div className="mt-3">
                            <div className="flex items-center justify-between mb-1">
                              <span className="text-xs text-orange-700">主管评分完成度</span>
                              <span className="text-xs font-medium text-orange-900">
                                {scores.length > 0
                                  ? Math.round(
                                      (scores.filter(s => s.manager_score && s.manager_score > 0).length /
                                        scores.length) *
                                        100
                                    )
                                  : 0}
                                %
                              </span>
                            </div>
                            <div className="w-full bg-orange-200 rounded-full h-2">
                              <div
                                className="bg-orange-600 h-2 rounded-full transition-all duration-300"
                                style={{
                                  width: `${
                                    scores.length > 0
                                      ? (scores.filter(s => s.manager_score && s.manager_score > 0).length /
                                          scores.length) *
                                        100
                                      : 0
                                  }%`,
                                }}
                              />
                            </div>
                          </div>
                        </div>
                      </div>
                    )}

                    {/* HR审核指导信息 */}
                    {canPerformAction(selectedEvaluation, "hr") && (
                      <div className="space-y-4">
                        <div className="bg-indigo-50 border border-indigo-200 rounded-lg p-4">
                          <h4 className="font-medium text-indigo-900 mb-2">🔍 HR审核指导</h4>
                          <ul className="text-sm text-indigo-800 space-y-1">
                            <li>• 审核员工自评与上级评分的合理性和一致性</li>
                            <li>• 检查评分是否符合公司绩效标准和政策</li>
                            <li>• 确认最终评分并可进行必要的调整</li>
                            <li>• 完成审核后，评估将进入员工确认阶段</li>
                          </ul>
                        </div>

                        {/* HR审核总结 */}
                        <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
                          <h4 className="font-medium text-gray-900 mb-3">📊 评分汇总分析</h4>
                          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
                            <div className="bg-white p-3 rounded border">
                              <div className="text-gray-600">员工自评总分</div>
                              <div className="text-2xl font-bold text-blue-600">
                                {scores.reduce((acc, score) => acc + (score.self_score || 0), 0)}
                              </div>
                              <div className="text-xs text-gray-500">
                                平均分：
                                {scores.length > 0
                                  ? (
                                      scores.reduce((acc, score) => acc + (score.self_score || 0), 0) / scores.length
                                    ).toFixed(1)
                                  : 0}
                              </div>
                            </div>
                            <div className="bg-white p-3 rounded border">
                              <div className="text-gray-600">主管评分总分</div>
                              <div className="text-2xl font-bold text-purple-600">
                                {scores.reduce((acc, score) => acc + (score.manager_score || 0), 0)}
                              </div>
                              <div className="text-xs text-gray-500">
                                平均分：
                                {scores.length > 0
                                  ? (
                                      scores.reduce((acc, score) => acc + (score.manager_score || 0), 0) / scores.length
                                    ).toFixed(1)
                                  : 0}
                              </div>
                            </div>
                            <div className="bg-white p-3 rounded border">
                              <div className="text-gray-600">评分差异分析</div>
                              <div className="text-2xl font-bold text-orange-600">
                                {Math.abs(
                                  scores.reduce((acc, score) => acc + (score.self_score || 0), 0) -
                                    scores.reduce((acc, score) => acc + (score.manager_score || 0), 0)
                                )}
                              </div>
                              <div className="text-xs text-gray-500">自评与主管评分差值</div>
                            </div>
                          </div>

                          {/* 差异分析提示 */}
                          {Math.abs(
                            scores.reduce((acc, score) => acc + (score.self_score || 0), 0) -
                              scores.reduce((acc, score) => acc + (score.manager_score || 0), 0)
                          ) > 10 && (
                            <div className="mt-3 p-3 bg-yellow-50 border border-yellow-200 rounded">
                              <div className="text-sm text-yellow-800">
                                ⚠️ <strong>注意：</strong>
                                员工自评与主管评分存在较大差异，建议重点关注并在最终评分中做出合理调整。
                              </div>
                            </div>
                          )}
                        </div>
                      </div>
                    )}

                    {scores.map(score => (
                      <Card key={score.id} role="score-item" data-score-id={score.id} className="border">
                        <CardContent className="px-4 py-2">
                          <div className="space-y-4">
                            {/* 项目信息 */}
                            <div className="flex flex-col sm:flex-row sm:justify-between sm:items-start gap-2">
                              <div className="flex-1">
                                <h4 className="font-medium text-lg">{score.item?.name}</h4>
                                <p className="text-sm text-muted-foreground">{score.item?.description}</p>
                                <p className="text-sm text-muted-foreground">满分：{score.item?.max_score}</p>
                              </div>
                              <div className="text-center sm:text-right">
                                <div className="text-2xl font-bold text-blue-600">
                                  {score.final_score || score.manager_score || score.self_score || 0}
                                </div>
                                <div className="text-sm text-muted-foreground">当前得分</div>
                              </div>
                            </div>

                            {/* 评分区域 */}
                            <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
                              {/* 自评区域 */}
                              <div className="space-y-2">
                                <Label className="text-sm font-medium flex items-center h-6">
                                  自评分数
                                  {canPerformAction(selectedEvaluation, "self") && (
                                    <Button
                                      variant="ghost"
                                      size="sm"
                                      className="ml-2 h-6 w-6 p-0"
                                      onClick={() => handleStartEdit(score.id, score.self_score, score.self_comment)}
                                    >
                                      <Edit2 className="w-3 h-3" />
                                    </Button>
                                  )}
                                </Label>

                                {editingScore === score.id && canPerformAction(selectedEvaluation, "self") ? (
                                  <div className="space-y-2">
                                    <div className="space-y-1">
                                      <Input
                                        type="number"
                                        value={tempScore}
                                        onChange={e => handleScoreChange(e.target.value, score.item?.max_score || 100)}
                                        min={0}
                                        max={score.item?.max_score}
                                        step="0.1"
                                        placeholder="评分"
                                      />
                                      <div className="text-xs text-gray-500">
                                        评分范围：0 - {score.item?.max_score || 100}分
                                      </div>
                                    </div>
                                    <Textarea
                                      value={tempComment}
                                      onChange={e => setTempComment(e.target.value)}
                                      placeholder="评价说明"
                                      rows={3}
                                    />
                                    <div className="flex space-x-2">
                                      <Button size="sm" onClick={() => handleSaveScore(score.id, "self")}>
                                        <Save className="w-3 h-3 mr-1" />
                                        保存
                                      </Button>
                                      <Button variant="outline" size="sm" onClick={handleCancelEdit}>
                                        <X className="w-3 h-3 mr-1" />
                                        取消
                                      </Button>
                                    </div>
                                  </div>
                                ) : (
                                  <div>
                                    <div className="text-sm font-medium">{score.self_score || "未评分"}</div>
                                    <div className="text-sm text-muted-foreground bg-gray-50 p-2 rounded mt-1">
                                      {score.self_comment || "暂无说明"}
                                    </div>
                                  </div>
                                )}
                              </div>

                              {/* 主管评分区域 */}
                              <div className="space-y-2">
                                <Label className="text-sm font-medium flex items-center h-6">
                                  主管评分
                                  {canPerformAction(selectedEvaluation, "manager") && (
                                    <Button
                                      variant="ghost"
                                      size="sm"
                                      className="ml-2 h-6 w-6 p-0"
                                      onClick={() =>
                                        handleStartEdit(score.id, score.manager_score, score.manager_comment)
                                      }
                                    >
                                      <Edit2 className="w-3 h-3" />
                                    </Button>
                                  )}
                                </Label>

                                {editingScore === score.id && canPerformAction(selectedEvaluation, "manager") ? (
                                  <div className="space-y-2">
                                    <div className="space-y-2">
                                      <div className="space-y-1">
                                        <Input
                                          type="number"
                                          value={tempScore}
                                          onChange={e =>
                                            handleScoreChange(e.target.value, score.item?.max_score || 100)
                                          }
                                          min={0}
                                          max={score.item?.max_score}
                                          step="0.1"
                                          placeholder="评分"
                                        />
                                        <div className="text-xs text-gray-500">
                                          评分范围：0 - {score.item?.max_score || 100}分
                                        </div>
                                      </div>
                                      {/* 评分参考标准 */}
                                      <div className="text-xs text-gray-500 bg-gray-50 p-2 rounded">
                                        <div className="font-medium mb-1">评分参考：</div>
                                        <div className="space-y-1">
                                          <div>
                                            优秀 ({Math.round((score.item?.max_score || 0) * 0.9)}-
                                            {score.item?.max_score}
                                            分)：超额完成目标，表现突出
                                          </div>
                                          <div>
                                            良好 ({Math.round((score.item?.max_score || 0) * 0.7)}-
                                            {Math.round((score.item?.max_score || 0) * 0.89)}
                                            分)：较好完成目标，有一定亮点
                                          </div>
                                          <div>
                                            合格 ({Math.round((score.item?.max_score || 0) * 0.6)}-
                                            {Math.round((score.item?.max_score || 0) * 0.69)}分)：基本完成目标，符合要求
                                          </div>
                                          <div>
                                            需改进 (0-{Math.round((score.item?.max_score || 0) * 0.59)}
                                            分)：未达成目标，需要改进
                                          </div>
                                        </div>
                                        <div className="mt-2 text-blue-600">员工自评：{score.self_score || 0}分</div>
                                      </div>
                                    </div>
                                    <Textarea
                                      value={tempComment}
                                      onChange={e => setTempComment(e.target.value)}
                                      placeholder="评价说明（请结合员工自评内容，提供具体的改进建议和发展方向）"
                                      rows={4}
                                    />
                                    <div className="flex space-x-2">
                                      <Button size="sm" onClick={() => handleSaveScore(score.id, "manager")}>
                                        <Save className="w-3 h-3 mr-1" />
                                        保存
                                      </Button>
                                      <Button variant="outline" size="sm" onClick={handleCancelEdit}>
                                        <X className="w-3 h-3 mr-1" />
                                        取消
                                      </Button>
                                    </div>
                                  </div>
                                ) : (
                                  <div>
                                    <div className="text-sm font-medium">{score.manager_score || "未评分"}</div>
                                    <div className="text-sm text-muted-foreground bg-gray-50 p-2 rounded mt-1">
                                      {score.manager_comment || "暂无说明"}
                                    </div>
                                  </div>
                                )}
                              </div>

                              {/* 最终得分区域 */}
                              <div className="space-y-2">
                                <Label className="text-sm font-medium flex items-center h-6">
                                  最终得分
                                  {canPerformAction(selectedEvaluation, "hr") && !score.final_score && (
                                    <Button
                                      variant="ghost"
                                      size="sm"
                                      className="ml-2 h-6 w-6 p-0"
                                      onClick={() =>
                                        handleStartEdit(score.id, score.manager_score, score.manager_comment)
                                      }
                                    >
                                      <Edit2 className="w-3 h-3" />
                                    </Button>
                                  )}
                                </Label>

                                {editingScore === score.id && canPerformAction(selectedEvaluation, "hr") ? (
                                  <div className="space-y-2">
                                    <div className="space-y-2">
                                      <div className="space-y-1">
                                        <Input
                                          type="number"
                                          value={tempScore}
                                          onChange={e =>
                                            handleScoreChange(e.target.value, score.item?.max_score || 100)
                                          }
                                          min={0}
                                          max={score.item?.max_score}
                                          step="0.1"
                                          placeholder="最终评分"
                                        />
                                        <div className="text-xs text-gray-500">
                                          评分范围：0 - {score.item?.max_score || 100}分
                                        </div>
                                      </div>
                                      {/* HR最终评分参考 */}
                                      <div className="text-xs text-gray-500 bg-indigo-50 p-2 rounded border border-indigo-200">
                                        <div className="font-medium mb-1">最终评分参考：</div>
                                        <div className="grid grid-cols-2 gap-2 text-xs">
                                          <div>员工自评：{score.self_score || 0}分</div>
                                          <div>主管评分：{score.manager_score || 0}分</div>
                                        </div>
                                        <div className="mt-2 text-indigo-700">
                                          💡 建议：通常采用主管评分作为最终得分，如有争议可适当调整
                                        </div>
                                      </div>
                                    </div>
                                    <Textarea
                                      value={tempComment}
                                      onChange={e => setTempComment(e.target.value)}
                                      placeholder="HR审核备注（可选）"
                                      rows={2}
                                    />
                                    <div className="flex space-x-2">
                                      <Button
                                        size="sm"
                                        onClick={() => {
                                          // HR确认最终得分
                                          handleSaveScore(score.id, "manager") // 临时使用manager类型，实际应该是final
                                        }}
                                      >
                                        <Save className="w-3 h-3 mr-1" />
                                        确认最终得分
                                      </Button>
                                      <Button variant="outline" size="sm" onClick={handleCancelEdit}>
                                        <X className="w-3 h-3 mr-1" />
                                        取消
                                      </Button>
                                    </div>
                                  </div>
                                ) : (
                                  <div>
                                    {score.final_score || score.manager_score ? (
                                      <div className="text-2xl font-bold text-green-600">
                                        {score.final_score || score.manager_score}
                                      </div>
                                    ) : (
                                      <div className="text-lg font-bold text-green-600">未评分</div>
                                    )}
                                    <div className="text-sm text-muted-foreground mt-1">
                                      {score.final_score
                                        ? "已确认"
                                        : canPerformAction(selectedEvaluation, "hr")
                                        ? "待HR确认"
                                        : "等待确认"}
                                    </div>
                                  </div>
                                )}
                              </div>
                            </div>
                          </div>
                        </CardContent>
                      </Card>
                    ))}
                  </TabsContent>

                  <TabsContent value="summary" className="space-y-4">
                    <Card>
                      <CardContent className="p-4">
                        <div className="text-center space-y-4">
                          <div>
                            <h3 className="text-2xl font-bold">总分统计</h3>
                            <div className="text-4xl font-bold text-blue-600 mt-2">
                              {scores.reduce(
                                (acc, score) =>
                                  acc + (score.final_score || score.manager_score || score.self_score || 0),
                                0
                              )}
                            </div>
                            <p className="text-muted-foreground">
                              满分 {scores.reduce((acc, score) => acc + (score.item?.max_score || 0), 0)} 分
                            </p>
                          </div>

                          <div className="grid grid-cols-3 gap-4 text-center">
                            <div>
                              <div className="text-lg font-semibold">
                                {scores.reduce((acc, score) => acc + (score.self_score || 0), 0)}
                              </div>
                              <div className="text-sm text-muted-foreground">自评总分</div>
                            </div>
                            <div>
                              <div className="text-lg font-semibold">
                                {scores.reduce((acc, score) => acc + (score.manager_score || 0), 0)}
                              </div>
                              <div className="text-sm text-muted-foreground">主管评分</div>
                            </div>
                            <div>
                              <div className="text-lg font-semibold">
                                {scores.reduce((acc, score) => acc + (score.final_score || 0), 0)}
                              </div>
                              <div className="text-sm text-muted-foreground">最终得分</div>
                            </div>
                          </div>
                        </div>
                      </CardContent>
                    </Card>

                    {/* 绩效评论卡片 */}
                    <Card className="mt-4">
                      <CardHeader>
                        <CardTitle className="flex items-center justify-between">
                          <div className="flex items-center">
                            <MessageCircle className="w-5 h-5 mr-2" />
                            绩效评论 ({comments.length})
                          </div>
                          {!isAddingComment && (
                            <Button variant="outline" size="sm" onClick={() => setIsAddingComment(true)}>
                              <Plus className="w-4 h-4 mr-1" />
                              添加评论
                            </Button>
                          )}
                        </CardTitle>
                      </CardHeader>
                      <CardContent>
                        {/* 添加评论表单 */}
                        {isAddingComment && (
                          <div className="space-y-4 mb-4 p-4 bg-gray-50 rounded-lg">
                            <div className="flex flex-col gap-2">
                              <Label htmlFor="newComment">评论内容</Label>
                              <Textarea
                                id="newComment"
                                placeholder="请输入您的评论..."
                                value={newComment}
                                onChange={e => setNewComment(e.target.value)}
                                className="mt-1 min-h-[100px] bg-white"
                              />
                            </div>
                            <div className="flex items-center space-x-2">
                              <input
                                type="checkbox"
                                id="newPrivate"
                                checked={newCommentPrivate}
                                onChange={e => setNewCommentPrivate(e.target.checked)}
                                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                              />
                              <Label htmlFor="newPrivate" className="text-sm">
                                仅自己可见
                              </Label>
                            </div>
                            <div className="flex justify-end space-x-2">
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => {
                                  setIsAddingComment(false)
                                  setNewComment("")
                                  setNewCommentPrivate(false)
                                }}
                              >
                                取消
                              </Button>
                              <Button
                                size="sm"
                                onClick={handleAddComment}
                                disabled={isSavingComment || !newComment.trim()}
                              >
                                {isSavingComment ? "保存中..." : "保存评论"}
                              </Button>
                            </div>
                          </div>
                        )}

                        {/* 评论列表 */}
                        {isLoadingComments ? (
                          <div className="text-center py-8 text-gray-500">
                            <p className="text-sm">加载评论中...</p>
                          </div>
                        ) : comments.length === 0 ? (
                          <div className="text-center py-8 text-gray-500">
                            <MessageCircle className="w-12 h-12 mx-auto mb-3 text-gray-300" />
                            <p className="text-sm">暂无评论</p>
                            <p className="text-xs mt-1">点击&quot;添加评论&quot;按钮来记录您的想法</p>
                          </div>
                        ) : (
                          <div className="space-y-4">
                            {comments.map(comment => (
                              <div key={comment.id} className="border rounded-lg p-4">
                                <div className="flex items-start justify-between">
                                  <div className="flex-1">
                                    <div className="flex items-center mb-2">
                                      <span className="font-medium text-sm">{comment.user?.name || "未知用户"}</span>
                                      <span className="text-xs text-gray-500 ml-2">{comment.user?.position}</span>
                                      <span className="text-xs text-gray-400 ml-2">
                                        {new Date(comment.created_at).toLocaleString()}
                                      </span>
                                      <div className="flex items-center ml-2 text-xs text-gray-500">
                                        {comment.is_private ? (
                                          <>
                                            <Lock className="w-3 h-3 mr-1" />
                                            仅自己可见
                                          </>
                                        ) : (
                                          <>
                                            <Globe className="w-3 h-3 mr-1" />
                                            公开可见
                                          </>
                                        )}
                                      </div>
                                    </div>

                                    {editingCommentId === comment.id ? (
                                      <div className="space-y-3">
                                        <Textarea
                                          value={editingCommentContent}
                                          onChange={e => setEditingCommentContent(e.target.value)}
                                          className="min-h-[80px]"
                                        />
                                        <div className="flex items-center space-x-2">
                                          <input
                                            type="checkbox"
                                            id={`edit-private-${comment.id}`}
                                            checked={editingCommentPrivate}
                                            onChange={e => setEditingCommentPrivate(e.target.checked)}
                                            className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                                          />
                                          <Label htmlFor={`edit-private-${comment.id}`} className="text-sm">
                                            仅自己可见
                                          </Label>
                                        </div>
                                        <div className="flex justify-end space-x-2">
                                          <Button variant="outline" size="sm" onClick={handleCancelEditComment}>
                                            取消
                                          </Button>
                                          <Button
                                            size="sm"
                                            onClick={() => handleSaveEditComment(comment.id)}
                                            disabled={isSavingComment}
                                          >
                                            {isSavingComment ? "保存中..." : "保存"}
                                          </Button>
                                        </div>
                                      </div>
                                    ) : (
                                      <p className="text-sm text-gray-800 whitespace-pre-wrap">{comment.content}</p>
                                    )}
                                  </div>

                                  {editingCommentId !== comment.id && comment.user_id === currentUser?.id && (
                                    <div className="flex items-center space-x-1 ml-2">
                                      <Button variant="ghost" size="sm" onClick={() => handleStartEditComment(comment)}>
                                        <Edit2 className="w-3 h-3" />
                                      </Button>
                                      <Button variant="ghost" size="sm" onClick={() => handleDeleteComment(comment.id)}>
                                        <Trash2 className="w-3 h-3" />
                                      </Button>
                                    </div>
                                  )}
                                </div>
                              </div>
                            ))}
                          </div>
                        )}
                      </CardContent>
                    </Card>
                  </TabsContent>
                </Tabs>
              </div>

              {/* 固定的流程控制按钮区域 */}
              <div className="flex-shrink-0 border-t pt-4">
                <div className="flex flex-col sm:flex-row justify-end gap-2 sm:space-x-2 sm:gap-0">
                  {canPerformAction(selectedEvaluation, "self") && (
                    <Button
                      onClick={() => handleCompleteStage(selectedEvaluation.id, "self")}
                      className="w-full sm:w-auto"
                      disabled={isSubmittingSelfEvaluation}
                    >
                      {isSubmittingSelfEvaluation ? "提交中..." : "完成自评"}
                    </Button>
                  )}
                  {canPerformAction(selectedEvaluation, "manager") && (
                    <Button
                      onClick={() => handleCompleteStage(selectedEvaluation.id, "manager")}
                      className="w-full sm:w-auto"
                    >
                      完成主管评估
                    </Button>
                  )}
                  {canPerformAction(selectedEvaluation, "hr") && (
                    <Button
                      onClick={() => handleCompleteStage(selectedEvaluation.id, "hr")}
                      className="w-full sm:w-auto"
                    >
                      完成HR审核
                    </Button>
                  )}
                  {canPerformAction(selectedEvaluation, "confirm") && (
                    <Button
                      onClick={() => handleCompleteStage(selectedEvaluation.id, "confirm")}
                      className="w-full sm:w-auto"
                    >
                      确认最终得分
                    </Button>
                  )}
                  <Button variant="outline" onClick={() => setScoreDialogOpen(false)} className="w-full sm:w-auto">
                    关闭
                  </Button>
                </div>
              </div>
            </>
          )}
        </DialogContent>
      </Dialog>
    </div>
  )
}
