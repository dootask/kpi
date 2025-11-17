"use client"

import { useState, useEffect } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
  CheckCircle,
  XCircle,
  Clock,
  AlertCircle,
  Wifi,
  WifiOff,
  RefreshCcw,
  Users,
  MessageSquare,
  Award,
  Star,
  TrendingUp,
} from "lucide-react"
import { useAuth } from "@/lib/auth-context"
import { useNotification, type SSEMessage } from "@/lib/notification-context"
import { useUnreadContext } from "@/lib/unread-context"
import { evaluationApi, invitationApi, sseApi, type SSEStatus } from "@/lib/api"
import { toast } from "sonner"

interface MessageTypeInfo {
  label: string
  icon: React.ReactElement
  color: string
}

export default function TestNotificationsPage() {
  const { user } = useAuth()
  const { isConnected, isConnecting, lastMessage, connectionRetries, connect, disconnect } = useNotification()
  const { unreadEvaluations, unreadInvitations, refreshUnreadEvaluations, refreshUnreadInvitations } =
    useUnreadContext()

  const [messages, setMessages] = useState<SSEMessage[]>([])
  const [sseStatus, setSseStatus] = useState<SSEStatus | null>(null)
  const [loading, setLoading] = useState(false)

  // 监听所有消息并记录
  useEffect(() => {
    if (lastMessage) {
      setMessages(prev => [lastMessage, ...prev.slice(0, 49)]) // 保留最近50条消息
    }
  }, [lastMessage])

  // 获取SSE连接状态
  const fetchSSEStatus = async () => {
    try {
      const data = await sseApi.getStatus()
      setSseStatus(data)
    } catch (error) {
      console.error("获取SSE状态失败:", error)
    }
  }

  // 测试评估通知
  const testEvaluationNotification = async () => {
    if (!user) return

    try {
      setLoading(true)
      // 创建一个测试评估
      await evaluationApi.create({
        employee_id: user.id,
        template_id: 1,
        period: "monthly",
        year: new Date().getFullYear(),
        month: new Date().getMonth() + 1,
        status: "pending",
        total_score: 0,
        final_comment: "",
        has_objection: false,
        objection_reason: "",
      })
      toast.success("测试评估通知已发送")
    } catch (error) {
      console.error("测试评估通知失败:", error)
      toast.error("测试评估通知失败")
    } finally {
      setLoading(false)
    }
  }

  // 测试邀请通知
  const testInvitationNotification = async () => {
    try {
      setLoading(true)
      // 获取待确认邀请数量来测试
      await invitationApi.getPendingCount()
      toast.success("测试邀请通知已发送")
    } catch (error) {
      console.error("测试邀请通知失败:", error)
      toast.error("测试邀请通知失败")
    } finally {
      setLoading(false)
    }
  }

  // 清空消息记录
  const clearMessages = () => {
    setMessages([])
    toast.info("消息记录已清空")
  }

  // 格式化消息类型
  const formatMessageType = (type: string): MessageTypeInfo => {
    const typeMap: { [key: string]: MessageTypeInfo } = {
      connected: { label: "连接确认", icon: <CheckCircle className="w-4 h-4" />, color: "bg-green-500" },
      heartbeat: { label: "心跳检测", icon: <Clock className="w-4 h-4" />, color: "bg-blue-500" },
      evaluation_created: { label: "评估创建", icon: <Award className="w-4 h-4" />, color: "bg-purple-500" },
      evaluation_updated: { label: "评估更新", icon: <Award className="w-4 h-4" />, color: "bg-purple-500" },
      evaluation_deleted: { label: "评估删除", icon: <Award className="w-4 h-4" />, color: "bg-red-500" },
      evaluation_status_changed: {
        label: "评估状态变更",
        icon: <TrendingUp className="w-4 h-4" />,
        color: "bg-orange-500",
      },
      invitation_created: { label: "邀请创建", icon: <Users className="w-4 h-4" />, color: "bg-indigo-500" },
      invitation_updated: { label: "邀请更新", icon: <Users className="w-4 h-4" />, color: "bg-indigo-500" },
      invitation_deleted: { label: "邀请删除", icon: <Users className="w-4 h-4" />, color: "bg-red-500" },
      invitation_status_changed: {
        label: "邀请状态变更",
        icon: <MessageSquare className="w-4 h-4" />,
        color: "bg-yellow-500",
      },
      invited_score_updated: { label: "邀请评分更新", icon: <Star className="w-4 h-4" />, color: "bg-cyan-500" },
      self_score_updated: { label: "自评分更新", icon: <Star className="w-4 h-4" />, color: "bg-green-500" },
      manager_score_updated: { label: "主管评分更新", icon: <Star className="w-4 h-4" />, color: "bg-blue-500" },
      hr_score_updated: { label: "HR评分更新", icon: <Star className="w-4 h-4" />, color: "bg-red-500" },
    }
    return typeMap[type] || { label: type, icon: <AlertCircle className="w-4 h-4" />, color: "bg-gray-500" }
  }

  useEffect(() => {
    fetchSSEStatus()
    const interval = setInterval(fetchSSEStatus, 5000)
    return () => clearInterval(interval)
  }, [])

  return (
    <div className="max-w-6xl mx-auto p-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">实时通知测试</h1>
        <Badge variant={isConnected ? "default" : "destructive"}>
          {isConnected ? <Wifi className="w-4 h-4 mr-1" /> : <WifiOff className="w-4 h-4 mr-1" />}
          {isConnected ? "已连接" : "未连接"}
        </Badge>
      </div>

      {/* 连接状态卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">连接状态</CardTitle>
            {isConnected ? (
              <CheckCircle className="h-4 w-4 text-green-500" />
            ) : (
              <XCircle className="h-4 w-4 text-red-500" />
            )}
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{isConnected ? "已连接" : isConnecting ? "连接中..." : "未连接"}</div>
            <p className="text-xs text-muted-foreground">重试次数: {connectionRetries}</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">待处理评估</CardTitle>
            <Award className="h-4 w-4 text-blue-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{unreadEvaluations}</div>
            <p className="text-xs text-muted-foreground">需要处理的评估</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">待处理邀请</CardTitle>
            <Users className="h-4 w-4 text-purple-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{unreadInvitations}</div>
            <p className="text-xs text-muted-foreground">需要处理的邀请</p>
          </CardContent>
        </Card>
      </div>

      {/* 服务器状态 */}
      {sseStatus && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">服务器状态</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-4">
              <div>
                <div className="text-sm text-muted-foreground truncate">用户ID</div>
                <div className="font-medium">{sseStatus.user_id}</div>
              </div>
              <div>
                <div className="text-sm text-muted-foreground truncate">在线状态</div>
                <Badge variant={sseStatus.is_online ? "default" : "secondary"}>
                  {sseStatus.is_online ? "在线" : "离线"}
                </Badge>
              </div>
              <div>
                <div className="text-sm text-muted-foreground truncate">在线用户数</div>
                <div className="font-medium">{sseStatus.online_users}</div>
              </div>
              <div>
                <div className="text-sm text-muted-foreground truncate">用户连接数</div>
                <div className="font-medium">{sseStatus.user_connection_count}</div>
              </div>
              <div>
                <div className="text-sm text-muted-foreground truncate">总连接数</div>
                <div className="font-medium">{sseStatus.total_connections}</div>
              </div>
              <div>
                <div className="text-sm text-muted-foreground truncate">最后更新</div>
                <div className="font-medium">{new Date().toLocaleTimeString()}</div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* 测试操作 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">测试操作</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-3">
            <Button onClick={connect} disabled={isConnected || isConnecting}>
              <Wifi className="w-4 h-4 mr-2" />
              连接SSE
            </Button>
            <Button onClick={disconnect} disabled={!isConnected} variant="outline">
              <WifiOff className="w-4 h-4 mr-2" />
              断开连接
            </Button>
            <Button onClick={testEvaluationNotification} disabled={loading}>
              <Award className="w-4 h-4 mr-2" />
              测试评估通知
            </Button>
            <Button onClick={testInvitationNotification} disabled={loading}>
              <Users className="w-4 h-4 mr-2" />
              测试邀请通知
            </Button>
            <Button onClick={refreshUnreadEvaluations} variant="outline">
              <RefreshCcw className="w-4 h-4 mr-2" />
              刷新评估数
            </Button>
            <Button onClick={refreshUnreadInvitations} variant="outline">
              <RefreshCcw className="w-4 h-4 mr-2" />
              刷新邀请数
            </Button>
            <Button onClick={clearMessages} variant="outline">
              清空消息
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* 消息记录 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center justify-between">
            消息记录
            <Badge variant="secondary">{messages.length}</Badge>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3 max-h-96 overflow-y-auto">
            {messages.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">暂无消息记录</div>
            ) : (
              messages.map((message, index) => {
                const typeInfo = formatMessageType(message.type)
                return (
                  <div key={index} className="flex items-start gap-3 p-3 border rounded-lg">
                    <div className={`p-2 rounded-full ${typeInfo.color} text-white`}>{typeInfo.icon}</div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="font-medium">{typeInfo.label}</span>
                        <span className="text-xs text-muted-foreground">
                          {new Date(message.timestamp).toLocaleString()}
                        </span>
                      </div>
                      <div className="text-sm text-muted-foreground mt-1">
                        {message.type === "connected" || message.type === "heartbeat" ? (
                          <span>系统消息</span>
                        ) : (
                          <span>{"message" in message.data ? message.data.message : JSON.stringify(message.data)}</span>
                        )}
                      </div>
                      <details className="mt-2">
                        <summary className="text-xs cursor-pointer text-muted-foreground hover:text-foreground">
                          查看详细数据
                        </summary>
                        <pre className="text-xs mt-2 p-2 bg-muted rounded overflow-x-auto">
                          {JSON.stringify(message.data, null, 2)}
                        </pre>
                      </details>
                    </div>
                  </div>
                )
              })
            )}
          </div>
        </CardContent>
      </Card>

      {/* 使用说明 */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">使用说明</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3 text-sm">
            <div>
              <strong>连接状态：</strong>显示当前SSE连接状态，包括连接状态、重试次数等信息。
            </div>
            <div>
              <strong>实时数据：</strong>显示待处理评估和邀请数量，会随着通知实时更新。
            </div>
            <div>
              <strong>测试操作：</strong>可以手动触发各种测试操作来验证通知功能。
            </div>
            <div>
              <strong>消息记录：</strong>记录所有接收到的SSE消息，包括系统消息和业务通知。
            </div>
            <div>
              <strong>验证步骤：</strong>
              <ol className="list-decimal list-inside ml-4 mt-2 space-y-1">
                <li>确保SSE连接状态为&ldquo;已连接&rdquo;</li>
                <li>在其他窗口进行评估或邀请相关操作</li>
                <li>观察此页面的消息记录和数据更新</li>
                <li>验证通知内容是否正确</li>
              </ol>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
