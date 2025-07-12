"use client"

import { useState, useEffect } from "react"
import Link from "next/link"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { authApi, RegisterRequest, type Department, settingsApi } from "@/lib/api"
import { useAuth } from "@/lib/auth-context"
import { useAppContext } from "@/lib/app-context"
import { toast } from "sonner"
import { UserX } from "lucide-react"

export default function RegisterPage() {
  const { register } = useAuth()
  const { Alert } = useAppContext()
  const [loading, setLoading] = useState(false)
  const [departments, setDepartments] = useState<Department[]>([])
  const [allowRegistration, setAllowRegistration] = useState(true)
  const [formData, setFormData] = useState<RegisterRequest>({
    name: "",
    email: "",
    password: "",
    position: "",
    department_id: 0,
  })

  // 检查是否允许注册
  useEffect(() => {
    const checkRegistrationSettings = async () => {
      try {
        const response = await settingsApi.get()
        setAllowRegistration(response.data.allow_registration)
      } catch (error) {
        console.error("获取注册设置失败:", error)
        // 如果获取失败，默认允许注册
        setAllowRegistration(true)
      }
    }
    checkRegistrationSettings()
  }, [])

  // 加载部门列表
  useEffect(() => {
    const fetchDepartments = async () => {
      try {
        const response = await authApi.getDepartments()
        setFormData(prev => ({ ...prev, department_id: response.data[0]?.id || 0 }))
        setDepartments(response.data)
      } catch (error) {
        console.error("获取部门列表失败:", error)
        await Alert("错误", "获取部门列表失败")
      }
    }
    if (allowRegistration) {
      fetchDepartments()
    }
  }, [Alert, allowRegistration])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!formData.department_id) {
      await Alert("提示", "请选择部门")
      return
    }

    setLoading(true)

    try {
      await register(formData)
      toast.success("注册成功")
    } catch (error: unknown) {
      let errorMessage = "注册失败，请重试"

      if (error && typeof error === "object" && "response" in error) {
        const response = (error as { response?: { data?: { error?: string } } }).response
        if (response?.data?.error) {
          errorMessage = response.data.error
        }
      }

      await Alert("注册失败", errorMessage)
    } finally {
      setLoading(false)
    }
  }

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setFormData(prev => ({ ...prev, [name]: value }))
  }

  const handleDepartmentChange = (value: string) => {
    if (!value) {
      return
    }
    setFormData(prev => ({ ...prev, department_id: parseInt(value) }))
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-sm py-10">
        <div className="text-center mb-6">
          <h1 className="text-2xl font-bold text-gray-900">绩效管理系统</h1>
          <p className="text-gray-600 mt-2">
            {allowRegistration ? "创建新账户" : "用户注册"}
          </p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center">
              {!allowRegistration && <UserX className="w-5 h-5 mr-2 text-red-500" />}
              {allowRegistration ? "注册账户" : "注册暂不开放"}
            </CardTitle>
            <CardDescription>
              {allowRegistration 
                ? "请填写以下信息创建您的账户"
                : "系统管理员已关闭用户注册功能"
              }
            </CardDescription>
          </CardHeader>
          <CardContent>
            {allowRegistration ? (
              <form onSubmit={handleSubmit}>
                <div className="flex flex-col gap-6">
                  <div className="grid gap-3">
                    <Label htmlFor="name">姓名</Label>
                    <Input
                      id="name"
                      name="name"
                      type="text"
                      placeholder="请输入姓名"
                      value={formData.name}
                      onChange={handleChange}
                      required
                    />
                  </div>
                  <div className="grid gap-3">
                    <Label htmlFor="email">邮箱</Label>
                    <Input
                      id="email"
                      name="email"
                      type="email"
                      placeholder="请输入邮箱"
                      value={formData.email}
                      onChange={handleChange}
                      required
                    />
                  </div>
                  <div className="grid gap-3">
                    <Label htmlFor="password">密码</Label>
                    <Input
                      id="password"
                      name="password"
                      type="password"
                      placeholder="请输入密码（至少6位）"
                      value={formData.password}
                      onChange={handleChange}
                      minLength={6}
                      required
                    />
                  </div>
                  <div className="grid gap-3">
                    <Label htmlFor="position">职位</Label>
                    <Input
                      id="position"
                      name="position"
                      type="text"
                      placeholder="请输入职位"
                      value={formData.position}
                      onChange={handleChange}
                      required
                    />
                  </div>
                  <div className="grid gap-3">
                    <Label htmlFor="department">部门</Label>
                    <Select value={formData.department_id.toString()} onValueChange={handleDepartmentChange}>
                      <SelectTrigger>
                        <SelectValue placeholder="请选择部门" />
                      </SelectTrigger>
                      <SelectContent>
                        {departments.map(dept => (
                          <SelectItem key={dept.id} value={dept.id.toString()}>
                            {dept.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="flex flex-col gap-3">
                    <Button type="submit" className="w-full" disabled={loading}>
                      {loading ? "注册中..." : "注册"}
                    </Button>
                  </div>
                </div>
                <div className="mt-4 text-center text-sm">
                  已有账户？{" "}
                  <Link href="/auth/login" className="underline underline-offset-4 hover:text-blue-600">
                    立即登录
                  </Link>
                </div>
              </form>
            ) : (
              <div className="text-center py-8">
                <div className="space-y-4">
                  <div className="text-gray-600">
                    当前系统暂不开放用户注册功能。
                  </div>
                  <div className="text-sm text-gray-500">
                    如需创建账户，请联系系统管理员。
                  </div>
                  <Button asChild className="w-full">
                    <Link href="/auth/login">
                      返回登录
                    </Link>
                  </Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
