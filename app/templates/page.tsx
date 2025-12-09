"use client"

import { useState, useEffect } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Dialog, DialogBody, DialogContent, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Plus, Edit, Trash2, ClipboardList, Settings, Eye } from "lucide-react"
import { templateApi, itemApi, type KPITemplate, type KPIItem } from "@/lib/api"
import { useAppContext } from "@/lib/app-context"
import { getPeriodLabel, formatScore } from "@/lib/utils"
import { LoadingInline } from "@/components/loading"

const defaultTemplateFormData = {
  name: "",
  description: "",
  period: "monthly",
  is_active: true,
}

const defaultItemFormData = {
  name: "",
  description: "",
  max_score: 0,
  order: 1,
}

export default function TemplatesPage() {
  const { Confirm } = useAppContext()
  const [templates, setTemplates] = useState<KPITemplate[]>([])
  const [selectedTemplate, setSelectedTemplate] = useState<KPITemplate | null>(null)
  const [templateTabValue, setTemplateTabValue] = useState("templates")
  const [items, setItems] = useState<KPIItem[]>([])
  const [loading, setLoading] = useState(true)
  const [templateDialogOpen, setTemplateDialogOpen] = useState(false)
  const [itemDialogOpen, setItemDialogOpen] = useState(false)
  const [editingTemplate, setEditingTemplate] = useState<KPITemplate | null>(null)
  const [editingItem, setEditingItem] = useState<KPIItem | null>(null)
  const [templateFormData, setTemplateFormData] = useState(defaultTemplateFormData)
  const [itemFormData, setItemFormData] = useState(defaultItemFormData)

  // 获取模板列表
  const fetchTemplates = async () => {
    try {
      const response = await templateApi.getAll()
      setTemplates(response.data || [])
    } catch (error) {
      console.error("获取模板列表失败:", error)
    }
    setLoading(false)
  }

  // 获取模板的KPI项目
  const fetchTemplateItems = async (templateId: number) => {
    try {
      const response = await templateApi.getItems(templateId)
      setItems(response.data || [])
    } catch (error) {
      console.error("获取KPI项目失败:", error)
      setItems([])
    }
  }

  useEffect(() => {
    fetchTemplates()
  }, [])

  // 创建或更新模板
  const handleTemplateSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      if (editingTemplate) {
        await templateApi.update(editingTemplate.id, templateFormData)
      } else {
        await templateApi.create(templateFormData)
      }

      fetchTemplates()
      setTemplateDialogOpen(false)
      setEditingTemplate(null)
      setTemplateFormData(defaultTemplateFormData)
    } catch (error) {
      console.error("保存模板失败:", error)
    }
  }

  // 创建或更新KPI项目
  const handleItemSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedTemplate) return

    try {
      const submitData = {
        ...itemFormData,
        template_id: selectedTemplate.id,
      }

      if (editingItem) {
        await itemApi.update(editingItem.id, submitData)
      } else {
        await itemApi.create(submitData)
      }

      fetchTemplateItems(selectedTemplate.id)
      setItemDialogOpen(false)
      setEditingItem(null)
      setItemFormData(defaultItemFormData)
    } catch (error) {
      console.error("保存KPI项目失败:", error)
    }
  }

  // 删除模板
  const handleDeleteTemplate = async (id: number) => {
    const result = await Confirm("删除模板", "确定要删除这个模板吗？")
    if (result) {
      try {
        await templateApi.delete(id)
        fetchTemplates()
        if (selectedTemplate?.id === id) {
          setSelectedTemplate(null)
          setItems([])
        }
      } catch (error) {
        console.error("删除模板失败:", error)
      }
    }
  }

  // 删除KPI项目
  const handleDeleteItem = async (id: number) => {
    const result = await Confirm("删除KPI项目", "确定要删除这个KPI项目吗？")
    if (result) {
      try {
        await itemApi.delete(id)
        if (selectedTemplate) {
          fetchTemplateItems(selectedTemplate.id)
        }
      } catch (error) {
        console.error("删除KPI项目失败:", error)
      }
    }
  }

  // 选择模板并加载其KPI项目
  const handleSelectTemplate = (template: KPITemplate) => {
    setSelectedTemplate(template)
    fetchTemplateItems(template.id)
    setTemplateTabValue("items")
  }

  // 打开编辑模板对话框
  const handleEditTemplate = (template: KPITemplate) => {
    setEditingTemplate(template)
    setTemplateFormData({
      name: template.name,
      description: template.description,
      period: template.period,
      is_active: template.is_active,
    })
    setTemplateDialogOpen(true)
  }

  // 打开编辑KPI项目对话框
  const handleEditItem = (item: KPIItem) => {
    setEditingItem(item)
    setItemFormData({
      name: item.name,
      description: item.description,
      max_score: item.max_score,
      order: item.order,
    })
    setItemDialogOpen(true)
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center flex-wrap gap-4">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold text-foreground">KPI模板管理</h1>
          <p className="text-muted-foreground mt-1 sm:mt-2">创建和管理KPI考核模板</p>
        </div>
        <Dialog open={templateDialogOpen} onOpenChange={setTemplateDialogOpen}>
          <DialogTrigger asChild>
            <Button className="w-full sm:w-auto lg:mt-8">
              <Plus className="w-4 h-4 mr-2" />
              创建模板
            </Button>
          </DialogTrigger>
          <DialogContent className="w-[95vw] sm:max-w-md mx-auto">
            <DialogHeader>
              <DialogTitle>{editingTemplate ? "编辑模板" : "创建模板"}</DialogTitle>
            </DialogHeader>
            <DialogBody>
              <form id="template-form" onSubmit={handleTemplateSubmit} className="space-y-4">
                <div className="flex flex-col gap-2">
                  <Label htmlFor="name">模板名称</Label>
                  <Input
                    id="name"
                    value={templateFormData.name}
                    onChange={e => setTemplateFormData({ ...templateFormData, name: e.target.value })}
                    required
                  />
                </div>
                <div className="flex flex-col gap-2">
                  <Label htmlFor="description">模板描述</Label>
                  <Input
                    id="description"
                    value={templateFormData.description}
                    onChange={e => setTemplateFormData({ ...templateFormData, description: e.target.value })}
                  />
                </div>
                <div className="flex flex-col gap-2">
                  <Label htmlFor="period">考核周期</Label>
                  <Select
                    value={templateFormData.period}
                    onValueChange={value => setTemplateFormData({ ...templateFormData, period: value })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="选择考核周期" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="monthly">月度</SelectItem>
                      <SelectItem value="quarterly">季度</SelectItem>
                      <SelectItem value="yearly">年度</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </form>
            </DialogBody>
            <DialogFooter className="flex-col-reverse sm:flex-row sm:justify-end gap-2 sm:space-x-2 sm:gap-0">
              <Button
                type="button"
                variant="outline"
                onClick={() => setTemplateDialogOpen(false)}
                className="w-full sm:w-auto"
              >
                取消
              </Button>
              <Button type="submit" form="template-form" className="w-full sm:w-auto">
                {editingTemplate ? "更新" : "创建"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      <Tabs className="w-full" value={templateTabValue} onValueChange={setTemplateTabValue}>
        <TabsList className="gap-1 mb-2 max-w-full overflow-x-auto justify-start">
          <TabsTrigger value="templates">模板列表</TabsTrigger>
          {selectedTemplate && (
            <TabsTrigger value="items">KPI项目 {selectedTemplate && `(${selectedTemplate.name})`}</TabsTrigger>
          )}
        </TabsList>

        <TabsContent value="templates">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center">
                <ClipboardList className="w-5 h-5 mr-2" />
                模板列表
              </CardTitle>
            </CardHeader>
            <CardContent>
              {loading ? (
                <LoadingInline className="py-8" message="加载中..." />
              ) : templates.length === 0 ? (
                <div className="text-center py-8 text-gray-500">暂无模板数据</div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>模板名称</TableHead>
                      <TableHead>描述</TableHead>
                      <TableHead>考核周期</TableHead>
                      {/* <TableHead>状态</TableHead> */}
                      <TableHead>创建时间</TableHead>
                      <TableHead className="text-right">操作</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {templates.map(template => (
                      <TableRow key={template.id}>
                        <TableCell className="font-medium">{template.name}</TableCell>
                        <TableCell>{template.description}</TableCell>
                        <TableCell>
                          <Badge variant="outline">{getPeriodLabel(template.period)}</Badge>
                        </TableCell>
                        {/* <TableCell>
                          <Badge variant={template.is_active ? "default" : "secondary"}>
                            {template.is_active ? "活跃" : "停用"}
                          </Badge>
                        </TableCell> */}
                        <TableCell>{new Date(template.created_at).toLocaleDateString()}</TableCell>
                        <TableCell className="text-right space-x-2">
                          <Button variant="outline" size="sm" onClick={() => handleSelectTemplate(template)}>
                            <Eye className="w-4 h-4" />
                          </Button>
                          <Button variant="outline" size="sm" onClick={() => handleEditTemplate(template)}>
                            <Edit className="w-4 h-4" />
                          </Button>
                          <Button variant="outline" size="sm" onClick={() => handleDeleteTemplate(template.id)}>
                            <Trash2 className="w-4 h-4" />
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="items">
          {selectedTemplate && (
            <Card>
              <CardHeader>
                <CardTitle className="flex flex-wrap items-center justify-between gap-4">
                  <div className="flex items-center">
                    <Settings className="w-5 h-5 mr-2" />
                    KPI项目配置
                  </div>
                  <Button
                    onClick={() => {
                      setEditingItem(null)
                      setItemFormData(defaultItemFormData)
                      setItemDialogOpen(true)
                    }}
                  >
                    <Plus className="w-4 h-4 mr-2" />
                    添加项目
                  </Button>
                  <Dialog open={itemDialogOpen} onOpenChange={setItemDialogOpen}>
                  <DialogContent className="w-[95vw] sm:max-w-md mx-auto">
                    <DialogHeader>
                      <DialogTitle>{editingItem ? "编辑KPI项目" : "添加KPI项目"}</DialogTitle>
                    </DialogHeader>
                    <DialogBody>
                      <form id="template-item-form" onSubmit={handleItemSubmit} className="space-y-4">
                        <div className="flex flex-col gap-2">
                          <Label htmlFor="item-name">项目名称</Label>
                          <Input
                            id="item-name"
                            value={itemFormData.name}
                            onChange={e => setItemFormData({ ...itemFormData, name: e.target.value })}
                            required
                          />
                        </div>
                        <div className="flex flex-col gap-2">
                          <Label htmlFor="item-description">项目描述</Label>
                          <Textarea
                            id="item-description"
                            value={itemFormData.description}
                            className="max-h-80"
                            maxLength={800}
                            onChange={e => setItemFormData({ ...itemFormData, description: e.target.value })}
                          />
                        </div>
                        <div className="flex flex-col gap-2">
                          <Label htmlFor="max-score">满分</Label>
                          <Input
                            id="max-score"
                            type="number"
                            value={itemFormData.max_score}
                            onChange={e => e.target.value !== "" && setItemFormData({ ...itemFormData, max_score: parseInt(e.target.value) })}
                            required
                          />
                          <div className="text-sm text-muted-foreground">
                            <p>1、如果满分小于0，则输入应在「满分-0」之间；</p>
                            <p>2、如果满分大于0，则输入应在「0-满分」之间；</p>
                            <p>3、如果满分等于0，则可输入任意分数。</p>
                          </div>
                        </div>
                        <div className="flex flex-col gap-2">
                          <Label htmlFor="order">排序</Label>
                          <Input
                            id="order"
                            type="number"
                            value={itemFormData.order}
                            onChange={e => e.target.value !== "" && setItemFormData({ ...itemFormData, order: parseInt(e.target.value) })}
                            required
                          />
                        </div>
                      </form>
                    </DialogBody>
                    <DialogFooter className="flex-col-reverse sm:flex-row sm:justify-end gap-2 sm:space-x-2 sm:gap-0">
                      <Button
                        type="button"
                        variant="outline"
                        onClick={() => setItemDialogOpen(false)}
                        className="w-full sm:w-auto"
                      >
                        取消
                      </Button>
                      <Button type="submit" form="template-item-form" className="w-full sm:w-auto">
                        {editingItem ? "更新" : "创建"}
                      </Button>
                    </DialogFooter>
                  </DialogContent>
                </Dialog>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="mb-4 p-4 bg-muted/50 rounded-lg">
                  <h3 className="text-sm font-medium mb-2 text-foreground">模板说明</h3>
                  <ul className="text-sm space-y-1 text-muted-foreground">
                    <li>• 当前模板: <span className="font-semibold">{selectedTemplate.name}</span></li>
                    <li>• 总项目数: <span className="font-semibold">{items.length}</span></li>
                    <li>• 总分: <span className="font-semibold">{formatScore(items.reduce((sum, item) => sum + item.max_score, 0))}分</span></li>
                  </ul>
                </div>
                {items.length === 0 ? (
                  <div className="text-center py-8 text-gray-500">暂无KPI项目，请点击&quot;添加项目&quot;创建</div>
                ) : (
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>排序</TableHead>
                        <TableHead>项目名称</TableHead>
                        <TableHead>描述</TableHead>
                        <TableHead>满分</TableHead>
                        <TableHead className="text-right">操作</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {items.map(item => (
                        <TableRow key={item.id}>
                          <TableCell>
                            <Badge variant="outline">{item.order}</Badge>
                          </TableCell>
                          <TableCell className="font-medium">{item.name}</TableCell>
                          <TableCell className="min-w-40">
                            <pre className="whitespace-pre-wrap break-words text-sm text-muted-foreground line-clamp-5">
                              {item.description}
                            </pre>
                          </TableCell>
                          <TableCell>
                            <Badge variant="secondary">{formatScore(item.max_score)}分</Badge>
                          </TableCell>
                          <TableCell className="text-right space-x-2">
                            <Button variant="outline" size="sm" onClick={() => handleEditItem(item)}>
                              <Edit className="w-4 h-4" />
                            </Button>
                            <Button variant="outline" size="sm" onClick={() => handleDeleteItem(item.id)}>
                              <Trash2 className="w-4 h-4" />
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                )}
              </CardContent>
            </Card>
          )}
        </TabsContent>
      </Tabs>
    </div>
  )
}
