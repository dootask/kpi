"use client"

import { useState, useEffect, useRef } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"

import { Textarea } from "@/components/ui/textarea"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import {
  MessageSquare,
  Clock,
  CheckCircle,
  XCircle,
  Eye,
  Edit2,
  Save,
  Award,
  Star,
  Calendar,
  User,
  Building,
} from "lucide-react"
import {
  invitationApi,
  invitedScoreApi,
  type EvaluationInvitation,
  type InvitedScore,
  type PaginatedResponse,
} from "@/lib/api"

import { useAppContext } from "@/lib/app-context"
import { useUnreadContext } from "@/lib/unread-context"
import { getPeriodValue, scoreInputValidation } from "@/lib/utils"
import { Pagination, usePagination } from "@/components/pagination"
import { LoadingInline } from "@/components/loading"
import { toast } from "sonner"

export default function InvitationsPage() {
  const { Alert, Confirm } = useAppContext()
  const { refreshUnreadInvitations } = useUnreadContext()
  const detailsRef = useRef<HTMLDivElement>(null)
  const [invitations, setInvitations] = useState<EvaluationInvitation[]>([])
  const [selectedInvitation, setSelectedInvitation] = useState<EvaluationInvitation | null>(null)
  const [invitationScores, setInvitationScores] = useState<InvitedScore[]>([])
  const [dialogOpen, setDialogOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // 分页相关状态
  const [paginationData, setPaginationData] = useState<PaginatedResponse<EvaluationInvitation> | null>(null)
  const { currentPage, pageSize, setCurrentPage, handlePageSizeChange } = usePagination(10)

  // Popover 状态控制
  const [openPopovers, setOpenPopovers] = useState<{[key: string]: boolean}>({})

  // 获取邀请列表
  const fetchInvitations = async () => {
    try {
      setLoading(true)
      setError(null)
      const response = await invitationApi.getMy({
        page: currentPage,
        pageSize: pageSize,
      })
      setInvitations(response.data || [])
      setPaginationData(response)
      refreshUnreadInvitations()
    } catch (error) {
      console.error("获取邀请列表失败:", error)
      setError("获取邀请列表失败，请刷新重试")
      setInvitations([])
      setPaginationData(null)
    } finally {
      setLoading(false)
    }
  }

  // 获取邀请评分详情
  const fetchInvitationScores = async (invitationId: number) => {
    try {
      const response = await invitationApi.getScores(invitationId)
      setInvitationScores(response.data || [])
    } catch (error) {
      console.error("获取邀请评分失败:", error)
      setInvitationScores([])
    }
  } 
  
  // 找到下一个未评分的项目
  const findNextUnscored = (currentScoreId: number): number | null => {
    const currentIndex = invitationScores.findIndex(s => s.id === currentScoreId)
    if (currentIndex === -1) return null

    // 从当前项目的下一个开始查找
    for (let i = currentIndex + 1; i < invitationScores.length; i++) {
      const score = invitationScores[i]
      if (!score.score || score.score === 0) {
        return score.id
      }
    }

    // 如果没找到，从头开始查找
    for (let i = 0; i < currentIndex; i++) {
      const score = invitationScores[i]
      if (!score.score || score.score === 0) {
        return score.id
      }
    }

    return null
  }

  // 滚动到指定的评分项目
  const scrollToNextUnscored = (currentScoreId: number, isNext: boolean = false) => {
    const nextUnscored = isNext ? currentScoreId : findNextUnscored(currentScoreId)
    if (!nextUnscored) {
      return
    }

    // 使用 setTimeout 确保DOM已更新
    setTimeout(() => {
      const element = detailsRef.current?.querySelector(`[data-score-id="${nextUnscored}"]`) as HTMLElement
      if (!element) {
        return
      }
      element.scrollIntoView({
        behavior: "smooth",
        block: "center",
        inline: "nearest",
      })
    }, 100)
  }

  useEffect(() => {
    fetchInvitations()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentPage, pageSize])

  // 接受邀请
  const handleAcceptInvitation = async (invitationId: number) => {
    try {
      await invitationApi.accept(invitationId)
      fetchInvitations()
      
      // 如果是在详情弹窗中操作，更新selectedInvitation
      if (selectedInvitation && selectedInvitation.id === invitationId) {
        setSelectedInvitation({
          ...selectedInvitation,
          status: "accepted"
        })
        // 获取评分详情
        fetchInvitationScores(invitationId)
      }
      
      toast.success("邀请接受成功")
    } catch (error) {
      console.error("接受邀请失败:", error)
      Alert("操作失败", "接受邀请失败，请重试")
    }
  }

  // 拒绝邀请
  const handleDeclineInvitation = async (invitationId: number) => {
    const confirmed = await Confirm("确认拒绝", "确定要拒绝这个邀请吗？")
    if (!confirmed) return

    try {
      await invitationApi.decline(invitationId)
      fetchInvitations()
      
      // 如果是在详情弹窗中操作，更新selectedInvitation或关闭弹窗
      if (selectedInvitation && selectedInvitation.id === invitationId) {
        setSelectedInvitation({
          ...selectedInvitation,
          status: "declined"
        })
        // 可以选择关闭弹窗
        setDialogOpen(false)
      }
      
      toast.success("邀请已拒绝")
    } catch (error) {
      console.error("拒绝邀请失败:", error)
      Alert("操作失败", "拒绝邀请失败，请重试")
    }
  }

  // 查看邀请详情
  const handleViewInvitation = (invitation: EvaluationInvitation) => {
    setSelectedInvitation(invitation)
    fetchInvitationScores(invitation.id)
    setDialogOpen(true)
  }

  // 保存评分
  const handleSaveScore = async (scoreId: number, scoreValue: string, commentValue: string) => {
    try {
      const currentScore = invitationScores.find(s => s.id === scoreId)
      if (!currentScore) {
        Alert("保存失败", "评分项目不存在")
        return
      }

      const maxScore = currentScore.item?.max_score || 100
      const numericScore = parseFloat(scoreValue)
      
      if (isNaN(numericScore) || numericScore < 0 || numericScore > maxScore) {
        Alert("输入错误", `请输入0-${maxScore}之间的有效分数`)
        return
      }

      await invitedScoreApi.update(scoreId, {
        score: numericScore,
        comment: commentValue,
      })

      // 刷新评分数据
      if (selectedInvitation) {
        await fetchInvitationScores(selectedInvitation.id)
      }

      // 关闭 Popover
      setOpenPopovers(prev => ({
        ...prev,
        [scoreId]: false
      }))

      // 延迟执行 scrollToNextUnscored，确保 Popover 关闭动画完成
      setTimeout(() => {
        scrollToNextUnscored(scoreId)
      }, 100)

      toast.success("评分保存成功")
    } catch (error) {
      console.error("保存评分失败:", error)
      Alert("保存失败", "保存评分失败，请重试")
    }
  }

  // 完成邀请评分
  const handleCompleteInvitation = async (invitationId: number) => {
    // 检查是否所有项目都已评分
    const uncompletedItems = invitationScores.filter(score => !score.score || score.score === 0)
    if (uncompletedItems.length > 0) {
      await Alert("评分未完成", `请先完成所有项目的评分。还有 ${uncompletedItems.length} 个项目未评分。`)
      scrollToNextUnscored(uncompletedItems[0].id, true)
      return
    }

    const confirmed = await Confirm("确认提交", "确定要提交邀请评分吗？提交后将无法修改。")
    if (!confirmed) return

    try {
      await invitationApi.complete(invitationId)
      fetchInvitations()
      setDialogOpen(false)
      toast.success("邀请评分已完成")
    } catch (error) {
      console.error("完成邀请评分失败:", error)
      Alert("提交失败", "完成邀请评分失败，请重试")
    }
  }



  // 获取状态样式
  const getStatusBadge = (status: string) => {
    switch (status) {
      case "pending":
        return <Badge variant="outline" className="text-yellow-600 border-yellow-600 dark:text-yellow-400 dark:border-yellow-400">待接受</Badge>
      case "accepted":
        return <Badge variant="outline" className="text-blue-600 border-blue-600 dark:text-blue-400 dark:border-blue-400">进行中</Badge>
      case "declined":
        return <Badge variant="outline" className="text-red-600 border-red-600 dark:text-red-400 dark:border-red-400">已拒绝</Badge>
      case "cancelled":
        return <Badge variant="outline" className="text-gray-600 border-gray-600 dark:text-gray-400 dark:border-gray-400">已撤销</Badge>
      case "completed":
        return <Badge variant="outline" className="text-green-600 border-green-600 dark:text-green-400 dark:border-green-400">已完成</Badge>
      default:
        return <Badge variant="outline">未知状态</Badge>
    }
  }

  return (
    <div className="space-y-6">
      {/* 头部 */}
      <div className="flex flex-col space-y-4 sm:flex-row sm:items-center sm:justify-between sm:space-y-0">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold text-foreground">邀请评分</h1>
          <p className="text-muted-foreground mt-1 sm:mt-2">查看和处理收到的评分邀请</p>
        </div>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs sm:text-sm font-medium">总邀请数</CardTitle>
            <MessageSquare className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-xl sm:text-2xl font-bold">{invitations.length}</div>
            <p className="text-xs text-muted-foreground">全部邀请</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs sm:text-sm font-medium">待处理</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-xl sm:text-2xl font-bold">
              {invitations.filter(inv => inv.status === "pending").length}
            </div>
            <p className="text-xs text-muted-foreground">等待接受</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs sm:text-sm font-medium">进行中</CardTitle>
            <Star className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-xl sm:text-2xl font-bold">
              {invitations.filter(inv => inv.status === "accepted").length}
            </div>
            <p className="text-xs text-muted-foreground">正在评分</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-xs sm:text-sm font-medium">已完成</CardTitle>
            <CheckCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-xl sm:text-2xl font-bold">
              {invitations.filter(inv => inv.status === "completed").length}
            </div>
            <p className="text-xs text-muted-foreground">已完成评分</p>
          </CardContent>
        </Card>
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="bg-red-50 dark:bg-red-950/50 border border-red-200 dark:border-red-800 rounded-lg p-4">
          <div className="text-red-800 dark:text-red-200 text-sm">⚠️ {error}</div>
        </div>
      )}

      {/* 邀请列表 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            邀请列表
            {loading && <LoadingInline />}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>评估对象</TableHead>
                <TableHead>邀请人</TableHead>
                <TableHead>评估信息</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>创建时间</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {invitations.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center py-8 text-muted-foreground">
                    暂无邀请记录
                  </TableCell>
                </TableRow>
              ) : (
                invitations.map(invitation => (
                  <TableRow key={invitation.id}>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Building className="w-4 h-4 text-muted-foreground" />
                        <div>
                          <div className="font-medium">{invitation.evaluation?.employee?.name}</div>
                          <div className="text-sm text-muted-foreground">
                            {invitation.evaluation?.employee?.department?.name}
                          </div>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <User className="w-4 h-4 text-muted-foreground" />
                        <div>
                          <div className="font-medium">{invitation.inviter?.name}</div>
                          <div className="text-sm text-muted-foreground">{invitation.inviter?.position}</div>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <div>
                          <div className="font-medium">{invitation.evaluation?.template?.name}</div>
                          <div className="text-sm text-muted-foreground">
                            {invitation.evaluation && getPeriodValue(invitation.evaluation)}
                          </div>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>{getStatusBadge(invitation.status)}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Calendar className="w-4 h-4 text-muted-foreground" />
                        <span className="text-sm">
                          {new Date(invitation.created_at).toLocaleDateString()}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end space-x-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleViewInvitation(invitation)}
                        >
                          <Eye className="w-4 h-4" />
                        </Button>
                        {invitation.status === "pending" && (
                          <>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleAcceptInvitation(invitation.id)}
                            >
                              <CheckCircle className="w-4 h-4" />
                            </Button>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => handleDeclineInvitation(invitation.id)}
                            >
                              <XCircle className="w-4 h-4" />
                            </Button>
                          </>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>

          {/* 分页组件 */}
          {paginationData && (
            <div className="mt-6">
              <Pagination
                currentPage={currentPage}
                totalPages={paginationData.totalPages}
                pageSize={pageSize}
                totalItems={paginationData.total}
                onPageChange={setCurrentPage}
                onPageSizeChange={handlePageSizeChange}
                className="justify-center"
              />
            </div>
          )}
        </CardContent>
      </Card>

      {/* 邀请详情对话框 */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="w-[95vw] sm:max-w-6xl max-h-[90vh] flex flex-col">
          <DialogHeader>
            <DialogTitle>
              邀请评分详情 - {selectedInvitation?.evaluation?.employee?.name}
            </DialogTitle>
          </DialogHeader>
          {selectedInvitation && (
            <>
              <div className="flex-1 overflow-y-auto space-y-4 -mx-6 px-6 pb-2" ref={detailsRef}>
                {/* 基本信息 */}
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                  <div className="bg-muted/50 p-3 rounded">
                    <Label className="text-sm text-muted-foreground">评估对象</Label>
                    <p className="text-sm font-medium">{selectedInvitation.evaluation?.employee?.name}</p>
                  </div>
                  <div className="bg-muted/50 p-3 rounded">
                    <Label className="text-sm text-muted-foreground">考核模板</Label>
                    <p className="text-sm font-medium">{selectedInvitation.evaluation?.template?.name}</p>
                  </div>
                  <div className="bg-muted/50 p-3 rounded">
                    <Label className="text-sm text-muted-foreground">邀请状态</Label>
                    <div className="mt-1">{getStatusBadge(selectedInvitation.status)}</div>
                  </div>
                </div>

                {/* 邀请消息 */}
                {selectedInvitation.message && (
                  <div className="bg-blue-50/50 dark:bg-blue-950/50 p-4 rounded-lg border border-blue-200 dark:border-blue-800">
                    <Label className="text-sm text-blue-800 dark:text-blue-100 font-medium">邀请消息</Label>
                    <p className="text-sm text-blue-700 dark:text-blue-200 mt-1">{selectedInvitation.message}</p>
                  </div>
                )}

                {/* 评分区域 */}
                {selectedInvitation.status === "accepted" && (
                  <div className="space-y-4">
                    <div className="bg-green-50/50 dark:bg-green-950/50 p-4 rounded-lg border border-green-200 dark:border-green-800">
                      <h4 className="font-medium text-green-800 dark:text-green-100 mb-2">📝 评分指导</h4>
                      <ul className="text-sm text-green-700 dark:text-green-200 space-y-1">
                        <li>• 请根据您对该员工的了解进行客观评分</li>
                        <li>• 评分范围为0到各项目满分，请结合实际情况评分</li>
                        <li>• 在评价说明中详细描述您的评分依据</li>
                        <li>• 完成所有项目评分后，点击&quot;完成评分&quot;提交</li>
                      </ul>
                    </div>

                    {/* 评分进度 */}
                    <div className="bg-gray-50/50 dark:bg-gray-950/50 p-4 rounded-lg border border-gray-200 dark:border-gray-800">
                      <h4 className="font-medium text-gray-800 dark:text-gray-100 mb-2">📊 评分进度</h4>
                      <div className="flex items-center justify-between">
                        <span className="text-sm text-gray-700 dark:text-gray-200">
                          已完成 {invitationScores.filter(s => s.score && s.score > 0).length} / {invitationScores.length} 项
                        </span>
                        <div className="flex-1 mx-4 bg-gray-200 dark:bg-gray-800 rounded-full h-2">
                          <div
                            className="bg-blue-600 dark:bg-blue-400 h-2 rounded-full transition-all duration-300"
                            style={{
                              width: `${invitationScores.length > 0
                                  ? (invitationScores.filter(s => s.score && s.score > 0).length / invitationScores.length) * 100
                                  : 0
                                }%`,
                            }}
                          />
                        </div>
                        <span className="text-sm font-medium text-gray-800 dark:text-gray-100">
                          {invitationScores.length > 0
                            ? Math.round(
                              (invitationScores.filter(s => s.score && s.score > 0).length / invitationScores.length) * 100
                            )
                            : 0}
                          %
                        </span>
                      </div>
                    </div>

                    {/* 评分项目 */}
                    {invitationScores.map(score => (
                      <Card key={score.id} className="border" data-score-id={score.id}>
                        <CardContent className="px-4 py-3">
                          <div className="space-y-3">
                            <div className="flex flex-col sm:flex-row sm:justify-between sm:items-start gap-2">
                              <div className="flex-1">
                                <h4 className="font-medium text-lg">{score.item?.name}</h4>
                                <pre className="whitespace-pre-wrap break-words text-sm text-muted-foreground">
                                  {score.item?.description}
                                </pre>
                                <p className="text-sm text-muted-foreground">满分：{score.item?.max_score}</p>
                              </div>
                              <div className="text-center">
                                <div className="text-2xl font-bold text-blue-600">
                                  {score.score || 0}
                                </div>
                                <div className="text-sm text-muted-foreground">当前评分</div>
                              </div>
                            </div>

                            {/* 评分区域 */}
                            <div className="flex items-center justify-between">
                              <div className="flex-1">
                                <div className="text-sm text-muted-foreground bg-muted/50 p-2 rounded">
                                  {score.comment || "暂无评价说明"}
                                </div>
                              </div>
                              <Popover 
                                open={openPopovers[score.id] || false}
                                onOpenChange={(open) => {
                                  setOpenPopovers(prev => ({
                                    ...prev,
                                    [score.id]: open
                                  }))
                                }}
                              >
                                <PopoverTrigger asChild>
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    className="ml-2"
                                  >
                                    <Edit2 className="w-4 h-4" />
                                  </Button>
                                </PopoverTrigger>
                                <PopoverContent className="w-80" align="end">
                                  <div className="grid gap-4">
                                    <div className="space-y-2">
                                      <h4 className="leading-none font-medium">评分编辑</h4>
                                      <p className="text-muted-foreground text-sm">
                                        编辑您的评分和评价说明
                                      </p>
                                    </div>
                                    <div className="grid gap-2">
                                      <div className="grid grid-cols-3 items-center gap-4">
                                        <Label htmlFor="score">评分</Label>
                                        <Input
                                          id="score"
                                          type="number"
                                          min={0}
                                          max={score.item?.max_score}
                                          step="0.1"
                                          defaultValue={score.score?.toString() || ""}
                                          className="col-span-2 h-8"
                                          placeholder={`0-${score.item?.max_score || 100}`}
                                          onInput={(e) => scoreInputValidation(e, score.item?.max_score || 100)}
                                        />
                                      </div>
                                      <div className="grid grid-cols-3 items-start gap-4">
                                        <Label htmlFor="comment">评价说明</Label>
                                        <Textarea
                                          id="comment"
                                          defaultValue={score.comment || ""}
                                          className="col-span-2 min-h-[60px] resize-none"
                                          placeholder="请输入评价说明..."
                                        />
                                      </div>
                                      <div className="flex justify-end space-x-2 pt-2">
                                        <Button
                                          size="sm"
                                          onClick={() => {
                                            const scoreInput = document.getElementById('score') as HTMLInputElement
                                            const commentInput = document.getElementById('comment') as HTMLTextAreaElement
                                            handleSaveScore(score.id, scoreInput.value, commentInput.value)
                                          }}
                                        >
                                          <Save className="w-3 h-3 mr-1" />
                                          保存
                                        </Button>
                                      </div>
                                    </div>
                                  </div>
                                </PopoverContent>
                              </Popover>
                            </div>
                          </div>
                        </CardContent>
                      </Card>
                    ))}
                  </div>
                )}

                {/* 已完成评分的结果显示 */}
                {selectedInvitation.status === "completed" && (
                  <div className="space-y-4">
                    <div className="bg-green-50/50 dark:bg-green-950/50 p-4 rounded-lg border border-green-200 dark:border-green-800">
                      <h4 className="font-medium text-green-800 dark:text-green-100 mb-2">✅ 评分已完成</h4>
                      <p className="text-sm text-green-700 dark:text-green-200">
                        您已完成对该员工的评分，感谢您的参与！
                      </p>
                    </div>

                    {/* 评分结果展示 */}
                    <div className="bg-gray-50/50 dark:bg-gray-950/50 p-4 rounded-lg border border-gray-200 dark:border-gray-800">
                      <h4 className="font-medium text-gray-800 dark:text-gray-100 mb-3">📊 评分结果</h4>
                      <div className="space-y-3">
                        {invitationScores.map(score => (
                          <div key={score.id} className="flex items-center justify-between p-3 bg-white dark:bg-gray-800 rounded border">
                            <div className="flex-1">
                              <div className="font-medium text-sm">{score.item?.name}</div>
                              <div className="text-xs text-muted-foreground mt-1">
                                {score.comment || "暂无评价说明"}
                              </div>
                            </div>
                            <div className="text-lg font-semibold text-blue-600">
                              {score.score || 0} / {score.item?.max_score || 0}
                            </div>
                          </div>
                        ))}
                      </div>
                      <div className="mt-4 pt-3 border-t">
                        <div className="flex items-center justify-between">
                          <span className="font-medium text-gray-800 dark:text-gray-100">总分：</span>
                          <span className="text-2xl font-bold text-blue-600">
                            {invitationScores.reduce((acc, score) => acc + (score.score || 0), 0)}
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </div>

              {/* 操作按钮 */}
              <div className="flex-shrink-0">
                <div className="flex flex-col sm:flex-row justify-end gap-2 sm:space-x-2 sm:gap-0">
                  {selectedInvitation.status === "pending" && (
                    <>
                      <Button
                        onClick={() => handleAcceptInvitation(selectedInvitation.id)}
                        className="w-full sm:w-auto"
                      >
                        <CheckCircle className="w-4 h-4 mr-2" />
                        接受邀请
                      </Button>
                      <Button
                        variant="outline"
                        onClick={() => handleDeclineInvitation(selectedInvitation.id)}
                        className="w-full sm:w-auto"
                      >
                        <XCircle className="w-4 h-4 mr-2" />
                        拒绝邀请
                      </Button>
                    </>
                  )}
                  {selectedInvitation.status === "accepted" && (
                    <Button
                      onClick={() => handleCompleteInvitation(selectedInvitation.id)}
                      className="w-full sm:w-auto"
                    >
                      <Award className="w-4 h-4 mr-2" />
                      完成评分
                    </Button>
                  )}
                  <Button
                    variant="outline"
                    onClick={() => setDialogOpen(false)}
                    className="w-full sm:w-auto"
                  >
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