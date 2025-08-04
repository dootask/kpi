"use client"

import { useEffect, useMemo, useState } from "react"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Label } from "@/components/ui/label"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { cn } from "@/lib/utils"
import { Check, ChevronsUpDown, CircleMinus, Loader2, Search, Users } from "lucide-react"
import { Checkbox } from "./ui/checkbox"
import { Popover, PopoverContent, PopoverTrigger } from "./ui/popover"
import { Command, CommandEmpty, CommandGroup, CommandItem, CommandList } from "./ui/command"
import { CommandInputExpand } from "./ui/command.expand"
import { employeeApi } from "@/lib/api"

export interface Employee {
  id: number
  name: string
  position: string
  department?: {
    name: string
  }
}

interface EmployeeSelectorProps {
  selectedEmployeeIds: string[]
  onSelectionChange: (employeeIds: string[]) => void
  placeholder?: string
  label?: string
  className?: string
  disabled?: boolean
  maxDisplayTags?: number
  disabledEmployeeIds?: number[]
}

export function EmployeeSelector({
  selectedEmployeeIds,
  onSelectionChange,
  placeholder = "选择员工...",
  label,
  className,
  disabled = false,
  maxDisplayTags = 2,
  disabledEmployeeIds = [],
}: EmployeeSelectorProps) {
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [searchTerm, setSearchTerm] = useState("")
  const [employees, setEmployees] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)

  // 获取员工列表
  const fetchEmployees = async (search: string) => {
    setLoading(true)
    try {
      const response = await employeeApi.getAll({ search, pageSize: search ? 10 : 100 })
      setEmployees(prev => {
        const newEmployees = response.data || []
        const newEmployeeIds = newEmployees.map(employee => employee.id.toString())
        return [...newEmployees, ...prev.filter(employee => !newEmployeeIds.includes(employee.id.toString()))]
      })
    } finally {
      setLoading(false)
    }
  }

  // 搜索员工
  useEffect(() => {
    const timer = setTimeout(
      () => {
        fetchEmployees(searchTerm)
      },
      searchTerm ? 500 : 0
    )
    return () => clearTimeout(timer)
  }, [searchTerm])

  // 获取选中的员工信息
  const getSelectedEmployees = () => {
    return employees.filter(emp => selectedEmployeeIds.includes(emp.id.toString()))
  }

  // 过滤员工列表
  const getFilteredEmployees = () => {
    if (!searchTerm) return employees
    return employees.filter(
      emp =>
        emp.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        emp.position.toLowerCase().includes(searchTerm.toLowerCase()) ||
        emp.department?.name.toLowerCase().includes(searchTerm.toLowerCase())
    )
  }

  // 按部门分组
  const getGroupedEmployees = () => {
    return getFilteredEmployees().reduce((acc: Record<string, Employee[]>, employee: Employee) => {
      const department = employee.department?.name || "未分配"
      if (!acc[department]) {
        acc[department] = []
      }
      acc[department].push(employee)
      return acc
    }, {})
  }

  // 处理员工选择
  const handleEmployeeSelect = (employeeId: string) => {
    const newSelection = selectedEmployeeIds.includes(employeeId)
      ? selectedEmployeeIds.filter(id => id !== employeeId)
      : [...selectedEmployeeIds, employeeId]

    onSelectionChange(newSelection)
  }

  // 全选功能
  const handleSelectAll = () => {
    const targetEmployees = searchTerm ? getFilteredEmployees() : employees
    const targetIds = targetEmployees.map(emp => emp.id.toString())
    const newSelection = [...new Set([...selectedEmployeeIds, ...targetIds])]
    onSelectionChange(newSelection)
  }

  // 清空选择
  const handleClearSelection = () => {
    onSelectionChange([])
  }

  // 部门全选/取消全选
  const handleDepartmentSelect = (departmentEmployees: Employee[]) => {
    const departmentIds = departmentEmployees.map(emp => emp.id.toString())
    const allSelected = departmentIds.every(id => selectedEmployeeIds.includes(id))

    if (allSelected) {
      // 如果全部选中，则取消选择该部门的所有员工
      const newSelection = selectedEmployeeIds.filter(id => !departmentIds.includes(id))
      onSelectionChange(newSelection)
    } else {
      // 如果未全选，则选择该部门的所有员工
      const newSelection = [...new Set([...selectedEmployeeIds, ...departmentIds])]
      onSelectionChange(newSelection.filter(id => !disabledEmployeeIds.includes(parseInt(id))))
    }
  }

  // 获取部门选择状态
  const getDepartmentSelectionState = (departmentEmployees: Employee[]) => {
    const departmentIds = departmentEmployees.map(emp => emp.id.toString())
    const selectedCount = departmentIds.filter(id => selectedEmployeeIds.includes(id)).length

    if (selectedCount === 0) return "none"
    if (selectedCount === departmentIds.length) return "all"
    return "partial"
  }

  // 关闭弹窗时清空搜索
  const handleDialogClose = (open: boolean) => {
    setIsDialogOpen(open)
    if (!open) {
      setSearchTerm("")
    }
  }

  const selectedEmployees = getSelectedEmployees()
  const filteredEmployees = getFilteredEmployees()
  const groupedEmployees = getGroupedEmployees()

  return (
    <div className={cn("space-y-2", className)}>
      {label && (
        <Label className="text-sm font-medium text-foreground">
          {label} {selectedEmployeeIds.length > 0 && `(${selectedEmployeeIds.length} 已选择)`}
        </Label>
      )}

      <Dialog open={isDialogOpen} onOpenChange={handleDialogClose}>
        <DialogTrigger asChild>
          <Button
            variant="outline"
            className={cn(
              "w-full justify-start min-h-9 h-auto px-3 py-2 font-normal hover:bg-transparent",
              !selectedEmployeeIds.length && "text-muted-foreground hover:text-muted-foreground",
              disabled && "cursor-not-allowed opacity-50"
            )}
            disabled={disabled}
          >
            <div className="flex flex-1 items-start gap-2">
              <div className="flex items-center justify-center h-5">
                <Users className="h-4 w-4 shrink-0" />
              </div>
              <div className="flex flex-1 items-center">
                {selectedEmployeeIds.length === 0 ? (
                  <div className="h-5 leading-5">{placeholder}</div>
                ) : (
                  <div className="flex items-center gap-1.5 flex-wrap overflow-hidden">
                    {selectedEmployees.slice(0, maxDisplayTags).map(employee => (
                      <Badge key={employee.id} variant="secondary" className="text-xs h-5 leading-5">
                        {employee.name}
                      </Badge>
                    ))}
                    {selectedEmployeeIds.length > maxDisplayTags && (
                      <Badge variant="outline" className="text-xs h-5 leading-5">
                        +{selectedEmployeeIds.length - maxDisplayTags}
                      </Badge>
                    )}
                  </div>
                )}
              </div>
            </div>
          </Button>
        </DialogTrigger>

        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Users className="h-5 w-5" />
              选择员工 ({selectedEmployeeIds.length} 已选择)
            </DialogTitle>
          </DialogHeader>

          <div className="space-y-4">
            {/* 搜索框 */}
            <div className="relative">
              {loading ? (
                <Loader2 className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground animate-spin" />
              ) : (
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              )}
              <Input
                type="text"
                placeholder="搜索员工姓名、职位或部门..."
                value={searchTerm}
                onChange={e => setSearchTerm(e.target.value)}
                className="pl-9"
                autoFocus
              />
            </div>

            {/* 快捷操作 */}
            <div className="flex justify-between items-center">
              <div className="flex gap-2">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={handleSelectAll}
                  disabled={
                    searchTerm
                      ? filteredEmployees.every(emp => selectedEmployeeIds.includes(emp.id.toString()))
                      : selectedEmployeeIds.length === employees.length
                  }
                >
                  {searchTerm ? "全选搜索结果" : "全选"}
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={handleClearSelection}
                  disabled={selectedEmployeeIds.length === 0}
                >
                  清空
                </Button>
              </div>
              <span className="text-sm text-muted-foreground">
                {searchTerm
                  ? `${filteredEmployees.filter(emp => selectedEmployeeIds.includes(emp.id.toString())).length} / ${
                      filteredEmployees.length
                    } 搜索结果`
                  : `${selectedEmployeeIds.length} / ${employees.length} 总计`}
              </span>
            </div>

            {/* 员工列表 */}
            <div className="border rounded-md max-h-96 overflow-y-auto">
              {filteredEmployees.length === 0 ? (
                <div className="p-8 text-center text-muted-foreground">
                  {searchTerm ? "未找到匹配的员工" : "暂无员工数据"}
                </div>
              ) : (
                Object.entries(groupedEmployees).map(([department, employees]) => (
                  <div key={department} className="border-b last:border-b-0">
                    <div className="sticky top-0 bg-muted/80 border-b text-sm font-medium text-foreground">
                      <Label className="pl-2 pr-4 py-3">
                        <Checkbox
                          checked={getDepartmentSelectionState(employees) === "all"}
                          onCheckedChange={() => handleDepartmentSelect(employees)}
                          className={
                            getDepartmentSelectionState(employees) === "partial"
                              ? "data-[state=unchecked]:bg-primary/20"
                              : ""
                          }
                        />
                        <div className="flex-1 flex items-center justify-between">
                          <span className={getDepartmentSelectionState(employees) === "partial" ? "text-primary" : ""}>
                            {department}
                          </span>
                          <span className="text-xs text-muted-foreground">
                            ({employees.filter(emp => selectedEmployeeIds.includes(emp.id.toString())).length}/
                            {employees.length})
                          </span>
                        </div>
                      </Label>
                    </div>
                    <div className="divide-y">
                      {employees.map(employee => (
                        <Label key={employee.id} className={cn(
                          "pl-4 pr-4 py-3 hover:bg-muted/50",
                          disabledEmployeeIds.includes(employee.id) && "opacity-50"
                        )}>
                          {disabledEmployeeIds.includes(employee.id) ? (
                            <CircleMinus className="w-4 h-4 text-muted-foreground" />
                          ) : (
                            <Checkbox
                              checked={selectedEmployeeIds.includes(employee.id.toString())}
                              onCheckedChange={() => handleEmployeeSelect(employee.id.toString())}
                            />
                          )}
                          <div className="flex-1 min-w-0 flex items-center justify-between">
                            <div className="text-foreground truncate">{employee.name}</div>
                            <div className="text-sm text-muted-foreground/80 font-normal truncate">
                              {employee.position}
                            </div>
                          </div>
                        </Label>
                      ))}
                    </div>
                  </div>
                ))
              )}
            </div>

            {/* 底部操作 */}
            <div className="flex justify-end gap-2 pt-2">
              <Button type="button" variant="outline" onClick={() => handleDialogClose(false)}>
                取消
              </Button>
              <Button type="button" onClick={() => handleDialogClose(false)}>
                确定 ({selectedEmployeeIds.length})
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}

interface EmployeeComboboxProps {
  value?: string
  onValueChange?: (value: string) => void
  placeholder?: string
  emptyPlaceholder?: string
  className?: string
  contentClassName?: string
  align?: "start" | "center" | "end"
}

export function EmployeeCombobox({
  value,
  onValueChange,
  placeholder = "选择员工...",
  emptyPlaceholder = "未找到匹配的员工",
  className,
  contentClassName,
  align = "center",
}: EmployeeComboboxProps) {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState("")
  const [employees, setEmployees] = useState<Employee[]>([])

  const fetchEmployees = async (search: string) => {
    setLoading(true)
    try {
      const response = await employeeApi.getAll({ search, pageSize: 100 })
      setEmployees(prev => {
        const newEmployees = response.data || []
        const newEmployeeIds = newEmployees.map(employee => employee.id.toString())
        return [...newEmployees, ...prev.filter(employee => !newEmployeeIds.includes(employee.id.toString()))]
      })
    } catch (error) {
      console.error("获取员工列表失败:", error)
      setEmployees([])
    } finally {
      setLoading(false)
    }
  }

  const filteredEmployees = useMemo(() => {
    return [
      {
        id: "all",
        name: "全部员工",
        position: "",
        department: undefined,
      },
      ...employees,
    ]
  }, [employees])

  useEffect(() => {
    const timer = setTimeout(
      () => {
        fetchEmployees(searchTerm)
      },
      searchTerm ? 500 : 0
    )
    return () => clearTimeout(timer)
  }, [searchTerm])

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button variant="outline" role="combobox" aria-expanded={open} className={className}>
          {value ? filteredEmployees.find(employee => employee.id.toString() === value)?.name : placeholder}
          <ChevronsUpDown className="opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className={cn("p-0", contentClassName)} align={align}>
        <Command>
          <CommandInputExpand
            placeholder={placeholder}
            className="h-9"
            value={searchTerm}
            onValueChange={setSearchTerm}
            loading={loading}
          />
          <CommandList>
            <CommandEmpty>{emptyPlaceholder}</CommandEmpty>
            <CommandGroup>
              {filteredEmployees.map(employee => (
                <CommandItem
                  key={employee.id}
                  value={`${employee.id}-${employee.name}-${employee.position}-${employee.department?.name}`}
                  onSelect={currentValue => {
                    onValueChange?.(currentValue === value ? "" : currentValue.split("-")[0])
                    setOpen(false)
                  }}
                >
                  <div className="flex w-full items-center justify-between gap-2">
                    <div className="flex-1 min-w-0 flex items-center justify-between gap-2">
                      <div className="shrink-0 text-foreground truncate">{employee.name}</div>
                      {employee.position && (
                        <div className="text-sm text-muted-foreground/80 font-normal truncate">
                          {[employee.department?.name, employee.position].filter(Boolean).join(" - ")}
                        </div>
                      )}
                    </div>
                    <Check className={value === employee.id.toString() ? "opacity-100" : "opacity-0"} />
                  </div>
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
