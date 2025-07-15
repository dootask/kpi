"use client"

import { useState, useEffect } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { ArrowLeft, Save, Send, User, Calendar, Clock, Star } from "lucide-react"
import { shareApi, type EvaluationShare, type ShareScore } from "@/lib/api"
import { useAppContext } from "@/lib/app-context"
import { useRouter } from "next/navigation"
import { format } from "date-fns"
import { zhCN } from "date-fns/locale"
import { toast } from "sonner"

export default function ShareDetailPage({ params }: { params: { id: string } }) {
  const { Alert } = useAppContext()
  const router = useRouter()
  const [share, setShare] = useState<EvaluationShare | null>(null)
  const [scores, setScores] = useState<ShareScore[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    fetchShareDetail()
  }, [params.id])

  const fetchShareDetail = async () => {
    try {
      setLoading(true)
      const [shareResponse, scoresResponse] = await Promise.all([
        shareApi.getDetail(parseInt(params.id)),
        shareApi.getScores(parseInt(params.id))
      ])
      setShare(shareResponse.data)
      setScores(scoresResponse.data)
    } catch (error) {
      Alert({
        title: "错误",
        description: "获取共享详情失败",
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }

  const handleScoreChange = (itemId: number, field: 'score' | 'comment', value: string | number) => {
    setScores(prev => prev.map(score => 
      score.item_id === itemId 
        ? { ...score, [field]: field === 'score' ? (value === '' ? undefined : Number(value)) : value }
        : score
    ))
  }

  const handleSaveScore = async (itemId: number) => {
    const score = scores.find(s => s.item_id === itemId)
    if (!score) return

    try {
      setSaving(true)
      await shareApi.updateScore(parseInt(params.id), itemId, {
        score: score.score,
        comment: score.comment
      })
      toast.success("评分保存成功")
    } catch (error) {
      toast.error("评分保存失败")
    } finally {
      setSaving(false)
    }
  }

  const handleSubmitShare = async () => {
    try {
      setSubmitting(true)
      await shareApi.submit(parseInt(params.id))
      toast.success("评分提交成功")
      router.push('/shares')
    } catch (error) {
      toast.error("评分提交失败")
    } finally {
      setSubmitting(false)
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

  const formatDeadline = (deadline?: string) => {
    if (!deadline) return "无限期"
    return format(new Date(deadline), "yyyy-MM-dd HH:mm", { locale: zhCN })
  }

  const isReadonly = share?.status === "completed"

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  if (!share) {
    return (
      <div className="text-center py-8">
        <p className="text-gray-500">共享不存在</p>
      </div>
    )
  }

  return (
    <div className="container mx-auto py-6 space-y-6">
      <div className="flex items-center gap-4 mb-6">
        <Button variant="outline" onClick={() => router.back()}>
          <ArrowLeft className="w-4 h-4 mr-2" />
          返回
        </Button>
        <h1 className="text-2xl font-bold">协助评分详情</h1>
        {getStatusBadge(share.status)}
      </div>

      <Tabs defaultValue="info" className="w-full">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="info">基本信息</TabsTrigger>
          <TabsTrigger value="scores">评分详情</TabsTrigger>
        </TabsList>

        <TabsContent value="info" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <User className="w-5 h-5" />
                员工信息
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label className="text-sm font-medium">姓名</Label>
                  <p className="text-sm">{share.evaluation?.employee?.name || "未知"}</p>
                </div>
                <div>
                  <Label className="text-sm font-medium">职位</Label>
                  <p className="text-sm">{share.evaluation?.employee?.position || "未知"}</p>
                </div>
                <div>
                  <Label className="text-sm font-medium">部门</Label>
                  <p className="text-sm">{share.evaluation?.employee?.department?.name || "未知"}</p>
                </div>
                <div>
                  <Label className="text-sm font-medium">邮箱</Label>
                  <p className="text-sm">{share.evaluation?.employee?.email || "未知"}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Calendar className="w-5 h-5" />
                评估信息
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label className="text-sm font-medium">评估模板</Label>
                  <p className="text-sm">{share.evaluation?.template?.name || "未知"}</p>
                </div>
                <div>
                  <Label className="text-sm font-medium">评估期间</Label>
                  <p className="text-sm">{share.evaluation?.period || "未知"}</p>
                </div>
                <div>
                  <Label className="text-sm font-medium">共享人</Label>
                  <p className="text-sm">{share.shared_by?.name || "未知"}</p>
                </div>
                <div>
                  <Label className="text-sm font-medium">截止时间</Label>
                  <p className="text-sm">{formatDeadline(share.deadline)}</p>
                </div>
              </div>
              {share.message && (
                <div>
                  <Label className="text-sm font-medium">共享说明</Label>
                  <p className="text-sm bg-gray-50 p-3 rounded-md">{share.message}</p>
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>现有评分（参考）</CardTitle>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>考核项目</TableHead>
                    <TableHead>自评分数</TableHead>
                    <TableHead>自评评价</TableHead>
                    <TableHead>主管评分</TableHead>
                    <TableHead>主管评价</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {share.evaluation?.scores?.map((score) => (
                    <TableRow key={score.id}>
                      <TableCell className="font-medium">
                        {score.item?.name}
                      </TableCell>
                      <TableCell>
                        {score.self_score !== undefined ? (
                          <div className="flex items-center gap-1">
                            <Star className="w-4 h-4 text-yellow-500" />
                            {score.self_score}
                          </div>
                        ) : (
                          <span className="text-gray-400">未评分</span>
                        )}
                      </TableCell>
                      <TableCell>
                        <div className="max-w-xs truncate" title={score.self_comment}>
                          {score.self_comment || "无"}
                        </div>
                      </TableCell>
                      <TableCell>
                        {score.manager_score !== undefined ? (
                          <div className="flex items-center gap-1">
                            <Star className="w-4 h-4 text-blue-500" />
                            {score.manager_score}
                          </div>
                        ) : (
                          <span className="text-gray-400">未评分</span>
                        )}
                      </TableCell>
                      <TableCell>
                        <div className="max-w-xs truncate" title={score.manager_comment}>
                          {score.manager_comment || "无"}
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="scores" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Star className="w-5 h-5" />
                我的评分
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-6">
                {scores.map((score) => (
                  <div key={score.id} className="border rounded-lg p-4">
                    <div className="flex items-center justify-between mb-4">
                      <h3 className="font-medium">{score.item?.name}</h3>
                      <div className="text-sm text-gray-500">
                        最高分: {score.item?.max_score || 100}
                      </div>
                    </div>
                    
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <Label htmlFor={`score-${score.id}`}>评分</Label>
                        <Input
                          id={`score-${score.id}`}
                          type="number"
                          min="0"
                          max={score.item?.max_score || 100}
                          value={score.score || ''}
                          onChange={(e) => handleScoreChange(score.item_id, 'score', e.target.value)}
                          disabled={isReadonly}
                        />
                      </div>
                      <div>
                        <Label htmlFor={`comment-${score.id}`}>评价</Label>
                        <Textarea
                          id={`comment-${score.id}`}
                          value={score.comment || ''}
                          onChange={(e) => handleScoreChange(score.item_id, 'comment', e.target.value)}
                          disabled={isReadonly}
                          rows={3}
                        />
                      </div>
                    </div>
                    
                    {!isReadonly && (
                      <div className="mt-4 flex justify-end">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleSaveScore(score.item_id)}
                          disabled={saving}
                        >
                          <Save className="w-4 h-4 mr-2" />
                          保存
                        </Button>
                      </div>
                    )}
                  </div>
                ))}
              </div>
              
              {!isReadonly && (
                <div className="mt-6 flex justify-center">
                  <Button
                    onClick={handleSubmitShare}
                    disabled={submitting}
                    className="px-8"
                  >
                    <Send className="w-4 h-4 mr-2" />
                    提交评分
                  </Button>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}