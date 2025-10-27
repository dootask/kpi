"use client"

import { useState, useEffect, useCallback } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Plus, Edit, Trash2, Users, Search } from "lucide-react"
import { departmentApi, type Department, type PaginatedResponse, type PaginationParams } from "@/lib/api"
import { useAppContext } from "@/lib/app-context"
import { Pagination, usePagination } from "@/components/pagination"
import { LoadingInline } from "@/components/loading"
import { AxiosError } from "axios"

export default function DepartmentsPage() {
  const { Confirm, Alert } = useAppContext()
  const [departments, setDepartments] = useState<Department[]>([])
  const [loading, setLoading] = useState(true)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingDepartment, setEditingDepartment] = useState<Department | null>(null)
  const [formData, setFormData] = useState({
    name: "",
    description: "",
  })

  // 分页相关状态
  const [paginationData, setPaginationData] = useState<PaginatedResponse<Department> | null>(null)
  const [search, setSearch] = useState<string>("")

  // 使用分页Hook
  const { currentPage, pageSize, setCurrentPage, handlePageSizeChange, resetPagination } = usePagination(10)

  // 获取部门列表
  const fetchDepartments = useCallback(async () => {
    try {
      setLoading(true)

      const params: PaginationParams = {
        page: currentPage,
        pageSize: pageSize,
      }

      if (search.trim()) {
        params.search = search.trim()
      }

      const response = await departmentApi.getAll(params)
      setDepartments(response.data || [])
      setPaginationData(response)
    } catch (error) {
      console.error("获取部门列表失败:", error)
      setDepartments([])
      setPaginationData(null)
    } finally {
      setLoading(false)
    }
  }, [currentPage, pageSize, search])

  useEffect(() => {
    fetchDepartments()
  }, [fetchDepartments])

  // 搜索处理
  const handleSearchChange = (value: string) => {
    setSearch(value)
    resetPagination()
  }

  // 创建或更新部门
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      if (editingDepartment) {
        await departmentApi.update(editingDepartment.id, formData)
      } else {
        await departmentApi.create(formData)
      }

      fetchDepartments()
      setDialogOpen(false)
      setEditingDepartment(null)
      setFormData({ name: "", description: "" })
    } catch (error) {
      console.error("保存部门失败:", error)
    }
  }

  // 删除部门
  const handleDelete = async (id: number) => {
    const result = await Confirm("删除部门", "确定要删除这个部门吗？")
    if (result) {
      try {
        await departmentApi.delete(id)
        fetchDepartments()
      } catch (error: unknown) {
        console.error("删除部门失败:", error)
        if (error instanceof AxiosError) {
          Alert("删除失败", error.response?.data?.error || "删除部门失败，请重试")
        } else {
          Alert("删除失败", "删除部门失败，请重试")
        }
      }
    }
  }

  // 打开编辑对话框
  const handleEdit = (department: Department) => {
    setEditingDepartment(department)
    setFormData({
      name: department.name,
      description: department.description,
    })
    setDialogOpen(true)
  }

  // 打开新增对话框
  const handleAdd = () => {
    setEditingDepartment(null)
    setFormData({ name: "", description: "" })
    setDialogOpen(true)
  }

  return (
    <div className="space-y-6">
      {/* 响应式头部 */}
      <div className="flex justify-between items-center flex-wrap gap-4">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold text-foreground">部门管理</h1>
          <p className="text-muted-foreground mt-1 sm:mt-2">管理组织架构和部门信息</p>
        </div>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button onClick={handleAdd} className="w-full sm:w-auto lg:mt-8">
              <Plus className="w-4 h-4 mr-2" />
              添加部门
            </Button>
          </DialogTrigger>
          <DialogContent className="w-[95vw] sm:max-w-md mx-auto">
            <DialogHeader>
              <DialogTitle>{editingDepartment ? "编辑部门" : "添加部门"}</DialogTitle>
            </DialogHeader>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="flex flex-col gap-2">
                <Label htmlFor="name">部门名称</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={e => setFormData({ ...formData, name: e.target.value })}
                  required
                />
              </div>
              <div className="flex flex-col gap-2">
                <Label htmlFor="description">部门描述</Label>
                <Input
                  id="description"
                  value={formData.description}
                  onChange={e => setFormData({ ...formData, description: e.target.value })}
                />
              </div>
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
                  {editingDepartment ? "更新" : "创建"}
                </Button>
              </div>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex flex-wrap items-center justify-between gap-4">
            <div className="flex items-center">
              <Users className="w-5 h-5 mr-2" />
              部门列表
            </div>
            <div className="flex items-center gap-4">
              <div className="relative">
                <Search className="w-4 h-4 absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder="搜索部门..."
                  value={search}
                  onChange={e => handleSearchChange(e.target.value)}
                  className="pl-10 w-48 sm:w-64"
                />
              </div>
            </div>
          </CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <LoadingInline className="py-8" message="加载中..." />
          ) : departments.length === 0 ? (
            <div className="text-center py-8 text-gray-500">暂无部门数据</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>部门名称</TableHead>
                  <TableHead>描述</TableHead>
                  <TableHead>员工数量</TableHead>
                  <TableHead>创建时间</TableHead>
                  <TableHead className="text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {departments.map(department => (
                  <TableRow key={department.id}>
                    <TableCell className="font-medium">{department.name}</TableCell>
                    <TableCell>{department.description}</TableCell>
                    <TableCell>
                      <Badge variant="secondary">{department.employees?.length || 0} 人</Badge>
                    </TableCell>
                    <TableCell>{new Date(department.created_at).toLocaleDateString()}</TableCell>
                    <TableCell className="text-right space-x-2">
                      <Button variant="outline" size="sm" onClick={() => handleEdit(department)}>
                        <Edit className="w-4 h-4" />
                      </Button>
                      <Button variant="outline" size="sm" onClick={() => handleDelete(department.id)}>
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}

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
    </div>
  )
}
