"use client"

import { useState, useEffect } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import { Label } from "@/components/ui/label"
import { Alert, AlertDescription } from "@/components/ui/alert"
import {
  Settings,
  RefreshCw,
  CheckCircle,
  AlertCircle,
  LogOut,
} from "lucide-react"
import { useAuth } from "@/lib/auth-context"
import { useAppContext } from "@/lib/app-context"
import { settingsApi } from "@/lib/api"

export default function SettingsPage() {
  const { logout } = useAuth()
  const { Confirm } = useAppContext()
  const [allowRegistration, setAllowRegistration] = useState(true)
  const [loading, setLoading] = useState(false)
  const [message, setMessage] = useState<{ type: string; content: string }>({ type: "", content: "" })

  // 初始化设置状态
  useEffect(() => {
    const fetchSettings = async () => {
      try {
        setLoading(true)
        const response = await settingsApi.get()
        setAllowRegistration(response.data.allow_registration)
      } catch (error) {
        console.error("获取设置失败:", error)
        setMessage({ type: "error", content: "获取设置失败" })
      } finally {
        setLoading(false)
      }
    }
    fetchSettings()
  }, [])

  // 保存设置
  const handleSaveSettings = async () => {
    setLoading(true)
    try {
      const response = await settingsApi.update({
        allow_registration: allowRegistration,
      })
      setMessage({ type: "success", content: response.message || "设置保存成功！" })
      setTimeout(() => setMessage({ type: "", content: "" }), 3000)
    } catch (error) {
      console.error("保存设置失败:", error)
      setMessage({ type: "error", content: "保存设置失败，请重试。" })
    } finally {
      setLoading(false)
    }
  }

  // 退出登录
  const handleLogout = async () => {
    const result = await Confirm("退出登录", "确定要退出当前账户吗？")
    if (result) {
      logout()
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center flex-wrap gap-4">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">系统设置</h1>
          <p className="text-gray-600 mt-2">管理系统基本配置</p>
        </div>
        <div className="flex items-center space-x-2">
          <Button variant="outline" onClick={handleLogout} className="text-red-600 hover:text-red-700">
            <LogOut className="w-4 h-4 mr-2" />
            退出登录
          </Button>
          <Button onClick={handleSaveSettings} disabled={loading}>
            {loading ? <RefreshCw className="w-4 h-4 mr-2 animate-spin" /> : <CheckCircle className="w-4 h-4 mr-2" />}
            保存设置
          </Button>
        </div>
      </div>

      {/* 消息提示 */}
      {message.content && (
        <Alert className={message.type === "success" ? "border-green-500" : "border-red-500"}>
          {message.type === "success" ? <CheckCircle className="h-4 w-4" /> : <AlertCircle className="h-4 w-4" />}
          <AlertDescription>{message.content}</AlertDescription>
        </Alert>
      )}

      {/* 系统设置 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center">
            <Settings className="w-5 h-5 mr-2" />
            系统设置
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
            <div className="flex-1">
              <Label htmlFor="allow_registration" className="text-sm font-medium">
                允许用户注册
              </Label>
              <p className="text-sm text-gray-600 mt-1">
                开启后，新用户可以自行注册账户；关闭后，只能由管理员创建账户。
              </p>
            </div>
            <Switch
              id="allow_registration"
              checked={allowRegistration}
              onCheckedChange={setAllowRegistration}
            />
          </div>

          <div className="p-4 bg-blue-50 rounded-lg">
            <h3 className="text-sm font-medium text-blue-900 mb-2">功能说明</h3>
            <ul className="text-sm text-blue-800 space-y-1">
              <li>• 开启注册：用户可以通过注册页面创建新账户</li>
              <li>• 关闭注册：注册页面将显示"暂不开放注册"的提示</li>
              <li>• 退出登录：点击右上角的"退出登录"按钮安全退出系统</li>
            </ul>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
