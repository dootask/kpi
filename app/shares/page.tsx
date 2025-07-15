"use client"

import { useState, useEffect } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Eye, Clock, CheckCircle, Calendar } from "lucide-react"
import { shareApi, type EvaluationShare } from "@/lib/api"
import { useAppContext } from "@/lib/app-context"
import { useRouter } from "next/navigation"
import { format } from "date-fns"
import { zhCN } from "date-fns/locale"

export default function SharesPage() {
  const { Alert } = useAppContext()
  const router = useRouter()
  const [shares, setShares] = useState<EvaluationShare[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchShares()
  }, [])

  const fetchShares = async () => {
    try {
      setLoading(true)
      const response = await shareApi.getMy()
      setShares(response.data)
    } catch (error) {
      Alert({
        title: "错误",
        description: "获取共享任务失败",
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "pending":
        return <Badge variant="outline" className="text-yellow-600 border-yellow-600">待评分</Badge>
      case "completed":
        return <Badge variant="outline" className="text-green-600 border-green-600">已完成</Badge>
      case "expired":
        return <Badge variant="outline" className="text-red-600 border-red-600">已过期</Badge>
      default:
        return <Badge variant="outline">{status}</Badge>
    }
  }

  const getPriorityColor = (deadline?: string) => {
    if (!deadline) return "text-gray-500"
    
    const deadlineDate = new Date(deadline)
    const now = new Date()
    const diffDays = Math.ceil((deadlineDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
    
    if (diffDays < 0) return "text-red-600" // 已过期
    if (diffDays <= 1) return "text-red-500" // 1天内
    if (diffDays <= 3) return "text-orange-500" // 3天内
    return "text-gray-500"
  }

  const formatDeadline = (deadline?: string) => {
    if (!deadline) return "无限期"
    return format(new Date(deadline), "yyyy-MM-dd HH:mm", { locale: zhCN })
  }

  const handleViewShare = (share: EvaluationShare) => {
    router.push(`/shares/${share.id}`)
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  return (
    <div className="container mx-auto py-6 space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Eye className="w-5 h-5" />
            协助评分任务
          </CardTitle>
        </CardHeader>
        <CardContent>
          {shares.length === 0 ? (
            <div className="text-center py-8">
              <div className="text-gray-500 mb-4">
                <Clock className="w-12 h-12 mx-auto mb-2" />
                <p>暂无共享评分任务</p>
              </div>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>员工姓名</TableHead>
                  <TableHead>职位</TableHead>
                  <TableHead>评估模板</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>截止时间</TableHead>
                  <TableHead>共享人</TableHead>
                  <TableHead>共享说明</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {shares.map((share) => (
                  <TableRow key={share.id}>
                    <TableCell className="font-medium">
                      {share.evaluation?.employee?.name || "未知"}
                    </TableCell>
                    <TableCell>
                      {share.evaluation?.employee?.position || "未知"}
                    </TableCell>
                    <TableCell>
                      {share.evaluation?.template?.name || "未知"}
                    </TableCell>
                    <TableCell>
                      {getStatusBadge(share.status)}
                    </TableCell>
                    <TableCell>
                      <div className={`flex items-center gap-1 ${getPriorityColor(share.deadline)}`}>
                        <Calendar className="w-4 h-4" />
                        {formatDeadline(share.deadline)}
                      </div>
                    </TableCell>
                    <TableCell>
                      {share.shared_by?.name || "未知"}
                    </TableCell>
                    <TableCell>
                      <div className="max-w-xs truncate" title={share.message}>
                        {share.message || "无"}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleViewShare(share)}
                        >
                          <Eye className="w-4 h-4 mr-1" />
                          {share.status === "completed" ? "查看" : "评分"}
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  )
}