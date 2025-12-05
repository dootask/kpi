"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import { useRouter } from "next/navigation"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { LoadingInline } from "@/components/loading"
import { Switch } from "@/components/ui/switch"
import {
  performanceRuleApi,
  type PerformanceRule,
  type PerformanceRuleRequest,
  type PerformanceRuleNoInvitation,
  type PerformanceRuleEmployeeInvite,
} from "@/lib/api"
import { useAuth } from "@/lib/auth-context"
import { toast } from "sonner"
import { Info, RefreshCw } from "lucide-react"

type PerformanceRuleForm = {
  enabled: boolean
  no_invitation: Record<keyof PerformanceRuleNoInvitation, string>
  with_invitation: {
    employee: Record<keyof PerformanceRuleEmployeeInvite, string>
  }
}

const tolerance = 0.001

const createDefaultForm = (): PerformanceRuleForm => ({
  enabled: false,
  no_invitation: {
    self_weight: "10",
    superior_weight: "90",
  },
  with_invitation: {
    employee: {
      self_weight: "10",
      invite_superior_weight: "30",
      superior_weight: "60",
    },
  },
})

const mapRuleToForm = (rule: PerformanceRule): PerformanceRuleForm => ({
  enabled: rule.enabled ?? false,
  no_invitation: {
    self_weight: rule.no_invitation.self_weight.toString(),
    superior_weight: rule.no_invitation.superior_weight.toString(),
  },
  with_invitation: {
    employee: {
      self_weight: rule.with_invitation.employee.self_weight.toString(),
      invite_superior_weight: rule.with_invitation.employee.invite_superior_weight.toString(),
      superior_weight: rule.with_invitation.employee.superior_weight.toString(),
    },
  },
})

interface ValidationResult {
  payload: PerformanceRuleRequest | null
  errors: string[]
}

const buildPayloadAndValidate = (form: PerformanceRuleForm): ValidationResult => {
  const errors: string[] = []

  const parseValue = (label: string, value: string): number => {
    if (value === "") {
      errors.push(`${label} 不能为空`)
      return NaN
    }

    const num = Number(value)
    if (!Number.isFinite(num)) {
      errors.push(`${label} 必须是数字`)
      return NaN
    }

    if (num < 0 || num > 100) {
      errors.push(`${label} 必须在0到100之间`)
    }

    return num
  }

  const noInvitation = {
    self_weight: parseValue("无邀请评分 - 自评", form.no_invitation.self_weight),
    superior_weight: parseValue("无邀请评分 - 上级评分", form.no_invitation.superior_weight),
  }

  const employee = {
    self_weight: parseValue("有邀请评分（员工）- 自评", form.with_invitation.employee.self_weight),
    invite_superior_weight: parseValue("有邀请评分（员工）- 邀请评分", form.with_invitation.employee.invite_superior_weight),
    superior_weight: parseValue("有邀请评分（员工）- 上级评分", form.with_invitation.employee.superior_weight),
  }

  const sums = [
    {
      label: "无邀请评分",
      value: noInvitation.self_weight + noInvitation.superior_weight,
    },
    {
      label: "有邀请评分（员工）",
      value: employee.self_weight + employee.invite_superior_weight + employee.superior_weight,
    },
  ]

  for (const { label, value } of sums) {
    if (!Number.isFinite(value)) {
      continue
    }
    if (Math.abs(value - 100) > tolerance) {
      errors.push(`${label} 权重之和必须等于100，当前为 ${value.toFixed(2)}%`)
    }
  }

  if (errors.length > 0) {
    return { payload: null, errors }
  }

  return {
    payload: {
      enabled: form.enabled,
      no_invitation: noInvitation,
      with_invitation: {
        employee,
      },
    },
    errors,
  }
}

const parseForTotal = (value: string) => {
  const num = Number(value)
  return Number.isFinite(num) ? num : 0
}

export default function PerformanceRulesPage() {
  const router = useRouter()
  const { isHR, loading: authLoading } = useAuth()

  const [formData, setFormData] = useState<PerformanceRuleForm>(createDefaultForm)
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [errorMessages, setErrorMessages] = useState<string[]>([])
  const [fetchError, setFetchError] = useState<string | null>(null)

  const fetchRule = useCallback(async () => {
    setLoading(true)
    setFetchError(null)
    try {
      const response = await performanceRuleApi.get()
      setFormData(mapRuleToForm(response.data))
    } catch (error) {
      console.error("获取绩效规则失败:", error)
      const message =
        (error as { response?: { data?: { error?: string; message?: string } } }).response?.data?.error ||
        (error as { response?: { data?: { error?: string; message?: string } } }).response?.data?.message ||
        (error instanceof Error ? error.message : "获取绩效规则失败")
      setFetchError(message)
      toast.error(message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    if (!authLoading && !isHR) {
      router.replace("/evaluations")
    }
  }, [authLoading, isHR, router])

  useEffect(() => {
    if (!authLoading && isHR) {
      fetchRule()
    }
  }, [authLoading, isHR, fetchRule])

  const totals = useMemo(
    () => ({
      noInvitation:
        parseForTotal(formData.no_invitation.self_weight) + parseForTotal(formData.no_invitation.superior_weight),
      employee:
        parseForTotal(formData.with_invitation.employee.self_weight) +
        parseForTotal(formData.with_invitation.employee.invite_superior_weight) +
        parseForTotal(formData.with_invitation.employee.superior_weight),
    }),
    [formData]
  )

  const isBusy = loading || saving

  const handleNoInvitationChange =
    (field: keyof PerformanceRuleForm["no_invitation"]) => (value: string) =>
      setFormData(prev => ({
        ...prev,
        no_invitation: {
          ...prev.no_invitation,
          [field]: value,
        },
      }))

  const handleEmployeeChange =
    (field: keyof PerformanceRuleForm["with_invitation"]["employee"]) => (value: string) =>
      setFormData(prev => ({
        ...prev,
        with_invitation: {
          employee: {
            ...prev.with_invitation.employee,
            [field]: value,
          },
        },
      }))

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setErrorMessages([])

    const { payload, errors } = buildPayloadAndValidate(formData)
    if (errors.length > 0 || !payload) {
      setErrorMessages(errors)
      return
    }

    try {
      setSaving(true)
      const response = await performanceRuleApi.update(payload)
      setFormData(mapRuleToForm(response.data))
      toast.success("绩效规则已保存")
    } catch (error) {
      console.error("更新绩效规则失败:", error)
      const message =
        (error as { response?: { data?: { error?: string; message?: string } } }).response?.data?.error ||
        (error as { response?: { data?: { error?: string; message?: string } } }).response?.data?.message ||
        (error instanceof Error ? error.message : "更新绩效规则失败")
      toast.error(message)
      setErrorMessages(prev => [...prev, message])
    } finally {
      setSaving(false)
    }
  }

  const handleRestoreDefault = () => {
    setFormData(createDefaultForm())
    setErrorMessages([])
  }

  if (authLoading) {
    return (
      <div className="py-12">
        <LoadingInline message="加载中..." />
      </div>
    )
  }

  if (!isHR) {
    return (
      <Alert variant="destructive" className="mt-4">
        <AlertTitle>权限不足</AlertTitle>
        <AlertDescription>只有HR可以访问绩效规则配置。</AlertDescription>
      </Alert>
    )
  }

  return (
    <div className="space-y-6">
      {/* 响应式头部 */}
      <div className="flex justify-between items-center flex-wrap gap-4">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold text-foreground">绩效规则配置</h1>
          <p className="text-muted-foreground mt-1 sm:mt-2">
            设置不同场景下自评、上级及邀请评分的权重比例
          </p>
        </div>
        <Button type="button" className="w-full sm:w-auto lg:mt-8" onClick={fetchRule} disabled={isBusy}>
          <RefreshCw className="w-4 h-4 mr-2" />
          重新加载
        </Button>
      </div>

      <Alert>
        <Info className="h-4 w-4 text-sky-600" aria-hidden="true" />
        <AlertTitle>使用说明</AlertTitle>
        <AlertDescription>
          <p>请根据公司制度调整各场景下的权重百分比，并确保每个场景的合计为100%。保存后立即生效。</p>
        </AlertDescription>
      </Alert>

      <form className="space-y-6" onSubmit={handleSubmit}>
        <Card>
          <CardContent className="flex flex-wrap items-center justify-between gap-4 py-4">
            <div>
              <p className="text-base font-semibold text-foreground">启用绩效规则</p>
              <p className="text-sm text-muted-foreground">
                关闭后将暂停使用该套权重配置，启用后立即生效。
              </p>
            </div>
            <Switch
              checked={formData.enabled}
              onCheckedChange={checked => setFormData(prev => ({ ...prev, enabled: checked }))}
              disabled={isBusy}
            />
          </CardContent>
        </Card>

        {fetchError && (
          <Alert variant="destructive">
            <AlertTitle>加载失败</AlertTitle>
            <AlertDescription>
              <p className="mb-2">{fetchError}</p>
              <Button type="button" variant="outline" size="sm" onClick={fetchRule} disabled={loading}>
                重新尝试
              </Button>
            </AlertDescription>
          </Alert>
        )}

        {errorMessages.length > 0 && (
          <Alert variant="destructive">
            <AlertTitle>校验失败</AlertTitle>
            <AlertDescription>
              <ul className="list-disc pl-4 space-y-1 text-sm">
                {errorMessages.map((message, index) => (
                  <li key={`${message}-${index}`}>{message}</li>
                ))}
              </ul>
            </AlertDescription>
          </Alert>
        )}

        {loading ? (
          <Card>
            <CardContent className="py-12">
              <LoadingInline message="加载绩效规则..." />
            </CardContent>
          </Card>
        ) : (
          <>
            <Card>
              <CardHeader>
                <CardTitle>无邀请评分</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid gap-4 sm:grid-cols-2">
                  <WeightInput
                    label="自评"
                    value={formData.no_invitation.self_weight}
                    onChange={handleNoInvitationChange("self_weight")}
                    disabled={isBusy}
                  />
                  <WeightInput
                    label="上级评分"
                    value={formData.no_invitation.superior_weight}
                    onChange={handleNoInvitationChange("superior_weight")}
                    disabled={isBusy}
                  />
                </div>
                <SectionTotal total={totals.noInvitation} />
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>有邀请评分 - 员工</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="grid gap-4 lg:grid-cols-3 sm:grid-cols-2">
                  <WeightInput
                    label="自评"
                    value={formData.with_invitation.employee.self_weight}
                    onChange={handleEmployeeChange("self_weight")}
                    disabled={isBusy}
                  />
                  <WeightInput
                    label="邀请评分"
                    value={formData.with_invitation.employee.invite_superior_weight}
                    onChange={handleEmployeeChange("invite_superior_weight")}
                    disabled={isBusy}
                  />
                  <WeightInput
                    label="上级评分"
                    value={formData.with_invitation.employee.superior_weight}
                    onChange={handleEmployeeChange("superior_weight")}
                    disabled={isBusy}
                  />
                </div>
                <SectionTotal total={totals.employee} />
              </CardContent>
            </Card>

          </>
        )}

        <div className="flex flex-wrap justify-end gap-3">
          <Button type="button" variant="outline" onClick={handleRestoreDefault} disabled={isBusy}>
            恢复默认
          </Button>
          <Button type="submit" disabled={isBusy}>
            {saving ? "保存中..." : "保存设置"}
          </Button>
        </div>
      </form>
    </div>
  )
}

function SectionTotal({ total }: { total: number }) {
  const isValid = Math.abs(total - 100) <= tolerance
  return (
    <p className={`mt-4 text-sm ${isValid ? "text-muted-foreground" : "text-destructive font-medium"}`}>
      合计：{total.toFixed(1)}%
    </p>
  )
}

interface WeightInputProps {
  label: string
  value: string
  onChange: (value: string) => void
  disabled?: boolean
}

function WeightInput({ label, value, onChange, disabled }: WeightInputProps) {
  return (
    <div className="space-y-2">
      <Label className="text-sm font-medium text-foreground">{label}</Label>
      <div className="flex items-center gap-2">
        <Input
          type="number"
          inputMode="decimal"
          min="0"
          max="100"
          step="0.1"
          value={value ?? ""}
          onChange={event => onChange(event.target.value)}
          disabled={disabled}
          className="text-right"
        />
        <span className="text-muted-foreground text-sm">%</span>
      </div>
    </div>
  )
}

