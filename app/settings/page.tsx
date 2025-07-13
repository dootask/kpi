"use client"

import { useState, useEffect } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import { Circle, CircleCheck } from "lucide-react"
import { RefreshCw, CheckCircle, LogOut, Monitor, Sun, Moon, Palette, Shield } from "lucide-react"
import { useAuth } from "@/lib/auth-context"
import { useAppContext } from "@/lib/app-context"
import { useTheme } from "@/lib/theme-context"
import { settingsApi, type DeadlineRules } from "@/lib/api"
import { toast } from "sonner"
import { cn } from "@/lib/utils"

type SettingTab = "appearance" | "system" | "logout"

export default function SettingsPage() {
  const { logout, isHR } = useAuth()
  const { Confirm } = useAppContext()
  const { theme, setTheme } = useTheme()
  const [allowRegistration, setAllowRegistration] = useState(true)
  const [loading, setLoading] = useState(false)
  const [activeTab, setActiveTab] = useState<SettingTab>("appearance")
  const [deadlineRules, setDeadlineRules] = useState<DeadlineRules>({
    standard_days: { self_eval: 7, manager_eval: 4, hr_review: 2, final_confirm: 1 },
    compressed_days: { self_eval: 5, manager_eval: 3, hr_review: 1, final_confirm: 1 },
    minimum_days: { self_eval: 2, manager_eval: 2, hr_review: 1, final_confirm: 1 },
    time_threshold: { standard: 14, compressed: 7, emergency: 6 },
    auto_process_overdue: true,
  })

  // 初始化设置状态
  useEffect(() => {
    const fetchSettings = async () => {
      if (!isHR) return // 非HR用户不需要加载系统设置
      
      try {
        setLoading(true)
        const [settingsResponse, deadlineRulesResponse] = await Promise.all([
          settingsApi.get(),
          settingsApi.getDeadlineRules()
        ])
        setAllowRegistration(settingsResponse.data.allow_registration)
        setDeadlineRules(deadlineRulesResponse.data)
      } catch (error) {
        console.error("获取设置失败:", error)
        toast.error("获取设置失败")
      } finally {
        setLoading(false)
      }
    }
    fetchSettings()
  }, [isHR])

  // 保存设置
  const handleSaveSettings = async () => {
    if (!isHR) {
      toast.error("只有HR用户可以保存系统设置")
      return
    }

    setLoading(true)
    try {
      const response = await settingsApi.update({
        allow_registration: allowRegistration,
      })
      toast.success(response.message || "设置保存成功！")
    } catch (error) {
      console.error("保存设置失败:", error)
      toast.error("保存设置失败，请重试。")
    } finally {
      setLoading(false)
    }
  }

  // 保存截止时间规则
  const handleSaveDeadlineRules = async () => {
    if (!isHR) {
      toast.error("只有HR用户可以保存系统设置")
      return
    }

    setLoading(true)
    try {
      const response = await settingsApi.updateDeadlineRules(deadlineRules)
      toast.success(response.message || "截止时间规则保存成功！")
    } catch (error) {
      console.error("保存截止时间规则失败:", error)
      toast.error("保存截止时间规则失败，请重试。")
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

  // 菜单项配置
  const menuItems = [
    {
      id: "appearance" as SettingTab,
      label: "外观设置",
      icon: <Palette className="w-4 h-4" />,
      available: true,
    },
    {
      id: "system" as SettingTab,
      label: "系统设置",
      icon: <Shield className="w-4 h-4" />,
      available: isHR,
    },
    {
      id: "logout" as SettingTab,
      label: "退出登录",
      icon: <LogOut className="w-4 h-4" />,
      available: true,
      action: handleLogout,
    },
  ]

  // 渲染外观设置内容
  const renderAppearanceContent = () => (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Palette className="w-5 h-5 mr-2" />
          外观设置
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        <div>
          <Label htmlFor="theme" className="text-sm font-medium mb-2 block">
            主题模式
          </Label>
          <div className="flex gap-4">
            {/* 浅色卡片 */}
            <button
              type="button"
              className={cn(
                "group flex-1 rounded-xl flex flex-col items-center p-0 overflow-hidden transition-all cursor-pointer shadow-sm hover:shadow-md",
                theme === "light"
                  ? "ring-1 ring-primary bg-background"
                  : "bg-background hover:bg-muted/50"
              )}
              onClick={() => setTheme("light")}
              aria-label="浅色模式"
            >
              {/* 预览区 */}
              <div className="w-full h-16 flex items-center justify-center bg-white border-b border-muted">
                <div className="w-8 h-8 rounded bg-background flex items-center justify-center">
                  <Sun className="w-5 h-5 text-yellow-400" />
                </div>
              </div>
              {/* 名称区 */}
              <div className="flex flex-col items-center py-3">
                <div className="flex items-center gap-1.5">
                  {theme === "light"
                    ? <CircleCheck className="w-4 h-4 text-primary" />
                    : <Circle className="w-4 h-4 text-muted-foreground" />}
                  <span className="font-medium text-sm text-foreground">浅色</span>
                </div>
              </div>
            </button>
            {/* 深色卡片 */}
            <button
              type="button"
              className={cn(
                "group flex-1 rounded-xl flex flex-col items-center p-0 overflow-hidden transition-all cursor-pointer shadow-sm hover:shadow-md",
                theme === "dark"
                  ? "ring-1 ring-primary bg-background"
                  : "bg-background hover:bg-muted/50"
              )}
              onClick={() => setTheme("dark")}
              aria-label="深色模式"
            >
              {/* 预览区 */}
              <div className="w-full h-16 flex items-center justify-center bg-zinc-900 border-b border-muted">
                <div className="w-8 h-8 rounded bg-zinc-800 flex items-center justify-center">
                  <Moon className="w-5 h-5 text-blue-400" />
                </div>
              </div>
              {/* 名称区 */}
              <div className="flex flex-col items-center py-3">
                <div className="flex items-center gap-1.5">
                  {theme === "dark"
                    ? <CircleCheck className="w-4 h-4 text-primary" />
                    : <Circle className="w-4 h-4 text-muted-foreground" />}
                  <span className="font-medium text-sm text-foreground">深色</span>
                </div>
              </div>
            </button>
            {/* 跟随系统卡片 */}
            <button
              type="button"
              className={cn(
                "group flex-1 rounded-xl flex flex-col items-center p-0 overflow-hidden transition-all cursor-pointer shadow-sm hover:shadow-md",
                theme === "system"
                  ? "ring-1 ring-primary bg-background"
                  : "bg-background hover:bg-muted/50"
              )}
              onClick={() => setTheme("system")}
              aria-label="跟随系统"
            >
              {/* 预览区 */}
              <div className="w-full h-16 flex items-center justify-center bg-gradient-to-r from-white via-zinc-900 to-white border-b border-muted">
                <div className="w-8 h-8 rounded bg-gradient-to-br from-white via-zinc-900 to-zinc-800 flex items-center justify-center">
                  <Monitor className="w-5 h-5 text-green-500" />
                </div>
              </div>
              {/* 名称区 */}
              <div className="flex flex-col items-center py-3">
                <div className="flex items-center gap-1.5">
                  {theme === "system"
                    ? <CircleCheck className="w-4 h-4 text-primary" />
                    : <Circle className="w-4 h-4 text-muted-foreground" />}
                  <span className="font-medium text-sm text-foreground">跟随系统</span>
                </div>
              </div>
            </button>
          </div>
        </div>
        <div className="p-4 bg-muted/50 rounded-lg">
          <h3 className="text-sm font-medium mb-2 text-foreground">主题说明</h3>
          <ul className="text-sm space-y-1 text-muted-foreground">
            <li>• 浅色模式：使用亮色背景和深色文字</li>
            <li>• 深色模式：使用深色背景和浅色文字，有助于减少眼部疲劳</li>
            <li>• 跟随系统：自动根据您的设备系统设置切换主题</li>
          </ul>
        </div>
      </CardContent>
    </Card>
  )

  // 渲染系统设置内容
  const renderSystemContent = () => (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Shield className="w-5 h-5 mr-2" />
          系统设置
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="flex items-center justify-between p-4 bg-muted/50 rounded-lg">
          <div className="flex-1">
            <Label htmlFor="allow_registration" className="text-sm font-medium">
              允许用户注册
            </Label>
            <p className="text-sm text-muted-foreground mt-1">
              开启后，新用户可以自行注册账户；关闭后，只能由管理员创建账户。
            </p>
          </div>
          <Switch id="allow_registration" checked={allowRegistration} onCheckedChange={setAllowRegistration} />
        </div>
        <div className="p-4 bg-muted/50 rounded-lg">
          <h3 className="text-sm font-medium mb-2 text-foreground">功能说明</h3>
          <ul className="text-sm space-y-1 text-muted-foreground">
            <li>• 开启注册：用户可以通过注册页面创建新账户</li>
            <li>• 关闭注册：注册页面将显示&quot;暂不开放注册&quot;的提示</li>
            <li>• 只有HR用户可以修改此设置</li>
          </ul>
        </div>
        <div className="flex justify-end">
          <Button onClick={handleSaveSettings} disabled={loading}>
            {loading ? <RefreshCw className="w-4 h-4 mr-2 animate-spin" /> : <CheckCircle className="w-4 h-4 mr-2" />}
            保存设置
          </Button>
        </div>
        
        {/* 截止时间规则配置 */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">截止时间规则配置</CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* 标准模式时间 */}
            <div className="p-4 bg-muted/50 rounded-lg">
              <h4 className="font-medium mb-3">标准模式时间（天）</h4>
              <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
                <div>
                  <Label htmlFor="standard_self_eval">员工自评</Label>
                  <Input
                    id="standard_self_eval"
                    type="number"
                    min="1"
                    value={deadlineRules.standard_days.self_eval}
                    onChange={e => setDeadlineRules(prev => ({
                      ...prev,
                      standard_days: { ...prev.standard_days, self_eval: parseInt(e.target.value) || 0 }
                    }))}
                  />
                </div>
                <div>
                  <Label htmlFor="standard_manager_eval">主管评分</Label>
                  <Input
                    id="standard_manager_eval"
                    type="number"
                    min="1"
                    value={deadlineRules.standard_days.manager_eval}
                    onChange={e => setDeadlineRules(prev => ({
                      ...prev,
                      standard_days: { ...prev.standard_days, manager_eval: parseInt(e.target.value) || 0 }
                    }))}
                  />
                </div>
                <div>
                  <Label htmlFor="standard_hr_review">HR审核</Label>
                  <Input
                    id="standard_hr_review"
                    type="number"
                    min="1"
                    value={deadlineRules.standard_days.hr_review}
                    onChange={e => setDeadlineRules(prev => ({
                      ...prev,
                      standard_days: { ...prev.standard_days, hr_review: parseInt(e.target.value) || 0 }
                    }))}
                  />
                </div>
                <div>
                  <Label htmlFor="standard_final_confirm">最终确认</Label>
                  <Input
                    id="standard_final_confirm"
                    type="number"
                    min="1"
                    value={deadlineRules.standard_days.final_confirm}
                    onChange={e => setDeadlineRules(prev => ({
                      ...prev,
                      standard_days: { ...prev.standard_days, final_confirm: parseInt(e.target.value) || 0 }
                    }))}
                  />
                </div>
              </div>
            </div>

            {/* 压缩模式时间 */}
            <div className="p-4 bg-muted/50 rounded-lg">
              <h4 className="font-medium mb-3">压缩模式时间（天）</h4>
              <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
                <div>
                  <Label htmlFor="compressed_self_eval">员工自评</Label>
                  <Input
                    id="compressed_self_eval"
                    type="number"
                    min="1"
                    value={deadlineRules.compressed_days.self_eval}
                    onChange={e => setDeadlineRules(prev => ({
                      ...prev,
                      compressed_days: { ...prev.compressed_days, self_eval: parseInt(e.target.value) || 0 }
                    }))}
                  />
                </div>
                <div>
                  <Label htmlFor="compressed_manager_eval">主管评分</Label>
                  <Input
                    id="compressed_manager_eval"
                    type="number"
                    min="1"
                    value={deadlineRules.compressed_days.manager_eval}
                    onChange={e => setDeadlineRules(prev => ({
                      ...prev,
                      compressed_days: { ...prev.compressed_days, manager_eval: parseInt(e.target.value) || 0 }
                    }))}
                  />
                </div>
                <div>
                  <Label htmlFor="compressed_hr_review">HR审核</Label>
                  <Input
                    id="compressed_hr_review"
                    type="number"
                    min="1"
                    value={deadlineRules.compressed_days.hr_review}
                    onChange={e => setDeadlineRules(prev => ({
                      ...prev,
                      compressed_days: { ...prev.compressed_days, hr_review: parseInt(e.target.value) || 0 }
                    }))}
                  />
                </div>
                <div>
                  <Label htmlFor="compressed_final_confirm">最终确认</Label>
                  <Input
                    id="compressed_final_confirm"
                    type="number"
                    min="1"
                    value={deadlineRules.compressed_days.final_confirm}
                    onChange={e => setDeadlineRules(prev => ({
                      ...prev,
                      compressed_days: { ...prev.compressed_days, final_confirm: parseInt(e.target.value) || 0 }
                    }))}
                  />
                </div>
              </div>
            </div>

            {/* 时间阈值设置 */}
            <div className="p-4 bg-muted/50 rounded-lg">
              <h4 className="font-medium mb-3">时间阈值设置（天）</h4>
              <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
                <div>
                  <Label htmlFor="threshold_standard">标准模式阈值</Label>
                  <Input
                    id="threshold_standard"
                    type="number"
                    min="1"
                    value={deadlineRules.time_threshold.standard}
                    onChange={e => setDeadlineRules(prev => ({
                      ...prev,
                      time_threshold: { ...prev.time_threshold, standard: parseInt(e.target.value) || 0 }
                    }))}
                  />
                  <p className="text-xs text-muted-foreground mt-1">≥该天数时使用标准模式</p>
                </div>
                <div>
                  <Label htmlFor="threshold_compressed">压缩模式阈值</Label>
                  <Input
                    id="threshold_compressed"
                    type="number"
                    min="1"
                    value={deadlineRules.time_threshold.compressed}
                    onChange={e => setDeadlineRules(prev => ({
                      ...prev,
                      time_threshold: { ...prev.time_threshold, compressed: parseInt(e.target.value) || 0 }
                    }))}
                  />
                  <p className="text-xs text-muted-foreground mt-1">≥该天数时使用压缩模式</p>
                </div>
                <div>
                  <Label htmlFor="threshold_emergency">紧急模式阈值</Label>
                  <Input
                    id="threshold_emergency"
                    type="number"
                    min="1"
                    value={deadlineRules.time_threshold.emergency}
                    onChange={e => setDeadlineRules(prev => ({
                      ...prev,
                      time_threshold: { ...prev.time_threshold, emergency: parseInt(e.target.value) || 0 }
                    }))}
                  />
                  <p className="text-xs text-muted-foreground mt-1">≥该天数时使用紧急模式</p>
                </div>
              </div>
            </div>

            {/* 自动处理设置 */}
            <div className="flex items-center justify-between p-4 bg-muted/50 rounded-lg">
              <div className="flex-1">
                <Label htmlFor="auto_process_overdue" className="text-sm font-medium">
                  自动处理超时
                </Label>
                <p className="text-sm text-muted-foreground mt-1">
                  开启后，系统会自动处理超时的考核评估，推进到下一个阶段。
                </p>
              </div>
              <Switch 
                id="auto_process_overdue" 
                checked={deadlineRules.auto_process_overdue} 
                onCheckedChange={checked => setDeadlineRules(prev => ({
                  ...prev,
                  auto_process_overdue: checked
                }))}
              />
            </div>

            <div className="flex justify-end">
              <Button onClick={handleSaveDeadlineRules} disabled={loading}>
                {loading ? <RefreshCw className="w-4 h-4 mr-2 animate-spin" /> : <CheckCircle className="w-4 h-4 mr-2" />}
                保存截止时间规则
              </Button>
            </div>
          </CardContent>
        </Card>
      </CardContent>
    </Card>
  )

  // 处理菜单项点击
  const handleMenuClick = (item: typeof menuItems[0]) => {
    if (item.action) {
      item.action()
    } else {
      setActiveTab(item.id)
    }
  }

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <div>
        <h1 className="text-3xl font-bold text-foreground">系统设置</h1>
        <p className="text-muted-foreground mt-2">管理您的个人偏好和系统配置</p>
      </div>

      {/* 主要内容区域 */}
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* 左侧导航菜单 - 小屏幕时显示在上方 */}
        <div className="lg:col-span-1">
          <Card className="sticky top-6">
            <CardHeader className="pb-3">
              <CardTitle className="text-lg">设置菜单</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {menuItems
                .filter(item => item.available)
                .map((item) => (
                  <Button
                    key={item.id}
                    variant={activeTab === item.id ? "default" : "ghost"}
                    className={cn(
                      "w-full justify-start text-left",
                      item.id === "logout" && "text-destructive hover:text-destructive/90 hover:bg-destructive/10"
                    )}
                    onClick={() => handleMenuClick(item)}
                  >
                    {item.icon}
                    <span className="ml-2">{item.label}</span>
                  </Button>
                ))}
            </CardContent>
          </Card>
        </div>

        {/* 右侧内容区域 */}
        <div className="lg:col-span-3">
          {activeTab === "appearance" && renderAppearanceContent()}
          {activeTab === "system" && isHR && renderSystemContent()}
          {activeTab === "system" && !isHR && (
            <div className="flex items-center justify-center h-64">
              <div className="text-center">
                <Shield className="w-12 h-12 mx-auto text-muted-foreground mb-4" />
                <h3 className="text-lg font-medium mb-2">权限不足</h3>
                <p className="text-muted-foreground">只有HR用户可以访问系统设置</p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
