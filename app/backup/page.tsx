"use client"

import { useState, useEffect, useCallback } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from "@/components/ui/alert-dialog"
import { Download, RotateCcw, Trash2, Database, HardDrive } from "lucide-react"
import { toast } from "sonner"
import { useAppContext } from "@/lib/app-context"
import { LoadingInline } from "@/components/loading"
import { backupApi, type BackupHistory } from "@/lib/api"
import { downloadUrl } from "@dootask/tools"

export default function BackupPage() {
  const { Confirm } = useAppContext()
  const [backups, setBackups] = useState<BackupHistory[]>([])
  const [loading, setLoading] = useState(true)
  const [backupLoading, setBackupLoading] = useState(false)

  // 获取备份历史
  const fetchBackups = useCallback(async () => {
    try {
      setLoading(true)
      const response = await backupApi.getHistory()
      setBackups(response.data || [])
    } catch (error) {
      console.error('获取备份历史失败:', error)
      toast.error('获取备份历史失败')
      setBackups([]) // 出错时也设置为数组
    } finally {
      setLoading(false)
    }
  }, [])

  // 创建备份
  const createBackup = async () => {
    try {
      setBackupLoading(true)
      const result = await backupApi.create()
      toast.success(result.message || '备份创建成功')
      fetchBackups() // 刷新列表
    } catch (error) {
      console.error('创建备份失败:', error)
      toast.error('创建备份失败')
    } finally {
      setBackupLoading(false)
    }
  }

  // 下载备份
  const downloadBackup = async (download_url: string) => {
    try {
      // 直接跳转到下载URL
      try {
        await downloadUrl(download_url)
      } catch {
        window.open(download_url, "_blank")
      }
    } catch (error) {
      console.error('下载备份失败:', error)
      toast.error(error instanceof Error ? error.message : '下载备份失败')
    }
  }

  // 恢复备份
  const restoreBackup = async (fileName: string) => {
    const confirmed = await Confirm(
      '确认恢复数据库',
      `确定要从备份文件 "${fileName}" 恢复数据库吗？此操作将覆盖当前所有数据，且不可撤销。`
    )

    if (!confirmed) return

    try {
      const result = await backupApi.restore(fileName)
      toast.success(result.message || '数据库恢复成功')

      // 恢复成功后，建议用户刷新页面
      setTimeout(() => {
        window.location.reload()
      }, 2000)
    } catch (error) {
      console.error('恢复备份失败:', error)
      toast.error(error instanceof Error ? error.message : '恢复备份失败')
    }
  }

  // 删除备份
  const deleteBackup = async (fileName: string) => {
    const confirmed = await Confirm(
      '确认删除备份',
      `确定要删除备份文件 "${fileName}" 吗？此操作不可撤销。`
    )

    if (!confirmed) return

    try {
      const result = await backupApi.delete(fileName)
      toast.success(result.message || '备份删除成功')
      fetchBackups() // 刷新列表
    } catch (error) {
      console.error('删除备份失败:', error)
      toast.error(error instanceof Error ? error.message : '删除备份失败')
    }
  }

  // 格式化文件大小
  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes'
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  // 格式化时间
  const formatDateTime = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    })
  }

  useEffect(() => {
    fetchBackups()
  }, [fetchBackups])

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">备份还原</h1>
          <p className="text-muted-foreground">
            管理和维护系统数据库备份
          </p>
        </div>
      </div>

      {/* 创建备份卡片 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Database className="h-5 w-5" />
            创建备份
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-4">
            <Button
              onClick={createBackup}
              disabled={backupLoading}
              className="flex items-center gap-2"
            >
              {backupLoading ? (
                <LoadingInline />
              ) : (
                <HardDrive className="h-4 w-4" />
              )}
              {backupLoading ? '正在创建备份...' : '一键备份'}
            </Button>
            <p className="text-sm text-muted-foreground">
              点击创建当前数据库的完整备份，所有业务数据将被保存到备份文件中。
            </p>
          </div>
        </CardContent>
      </Card>

      {/* 备份历史卡片 */}
      <Card>
        <CardHeader>
          <CardTitle>备份历史</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <LoadingInline />
            </div>
          ) : !backups || backups.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Database className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>暂无备份记录</p>
              <p className="text-sm">点击&ldquo;一键备份&rdquo;创建第一个备份</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>文件名</TableHead>
                  <TableHead>文件大小</TableHead>
                  <TableHead>文件类型</TableHead>
                  <TableHead>创建时间</TableHead>
                  <TableHead className="w-[250px]">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {backups.map((backup) => (
                  <TableRow key={backup.file_name}>
                    <TableCell className="font-medium">
                      {backup.file_name}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">
                        {formatFileSize(backup.file_size)}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant={backup.file_type === 'db' ? 'default' : 'secondary'}>
                        {backup.file_type === 'db' ? '数据库文件' : 'SQL脚本'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {formatDateTime(backup.created_at)}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => downloadBackup(backup.download_url)}
                          className="flex items-center gap-1"
                        >
                          <Download className="h-3 w-3" />
                          下载
                        </Button>

                        <AlertDialog>
                          <AlertDialogTrigger asChild>
                            <Button
                              variant="outline"
                              size="sm"
                              className="flex items-center gap-1 text-orange-600 hover:text-orange-700"
                            >
                              <RotateCcw className="h-3 w-3" />
                              恢复
                            </Button>
                          </AlertDialogTrigger>
                          <AlertDialogContent>
                            <AlertDialogHeader>
                              <AlertDialogTitle>确认恢复数据库</AlertDialogTitle>
                              <AlertDialogDescription>
                                您即将从备份文件 <strong>{backup.file_name}</strong> 恢复数据库。
                                此操作将覆盖当前所有数据，且不可撤销。
                                <br /><br />
                                <span className="text-red-600 font-semibold">
                                  ⚠️ 请确保您了解此操作的后果！
                                </span>
                              </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                              <AlertDialogCancel>取消</AlertDialogCancel>
                              <AlertDialogAction
                                onClick={() => restoreBackup(backup.file_name)}
                                className="bg-red-600 hover:bg-red-700"
                              >
                                确认恢复
                              </AlertDialogAction>
                            </AlertDialogFooter>
                          </AlertDialogContent>
                        </AlertDialog>
                        <AlertDialog>
                          <AlertDialogTrigger asChild>
                            <Button
                              variant="outline"
                              size="sm"
                              className="flex items-center gap-1 text-red-600 hover:text-red-700"
                            >
                              <Trash2 className="h-3 w-3" />
                              删除
                            </Button>
                          </AlertDialogTrigger>
                          <AlertDialogContent>
                            <AlertDialogHeader>
                              <AlertDialogTitle>确认删除备份</AlertDialogTitle>
                              <AlertDialogDescription>
                                您即将删除备份文件 <strong>{backup.file_name}</strong>。
                                此操作不可撤销。
                              </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                              <AlertDialogCancel>取消</AlertDialogCancel>
                              <AlertDialogAction
                                onClick={() => deleteBackup(backup.file_name)}
                                className="bg-red-600 hover:bg-red-700"
                              >
                                确认删除
                              </AlertDialogAction>
                            </AlertDialogFooter>
                          </AlertDialogContent>
                        </AlertDialog>
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
