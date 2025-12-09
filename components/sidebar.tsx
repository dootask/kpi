"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import { cn } from "@/lib/utils"
import {
  Users,
  Building,
  ClipboardList,
  FileText,
  BarChart3,
  Settings,
  Home,
  Menu,
  X,
  HelpCircle,
  MessageSquare,
  Database,
  Scale,
} from "lucide-react"
import { useEffect, useMemo } from "react"
import { useAuth } from "@/lib/auth-context"
import { Button } from "./ui/button"
import { useRouter } from "next/navigation"
import { Badge } from "./ui/badge"
import { useDootaskContext } from "@/lib/dootask-context"
import { useUnreadContext } from "@/lib/unread-context"

interface SidebarProps {
  isMobileMenuOpen: boolean
  setIsMobileMenuOpen: (open: boolean) => void
}

export function Sidebar({ isMobileMenuOpen, setIsMobileMenuOpen }: SidebarProps) {
  const pathname = usePathname()
  const router = useRouter()
  const { user: currentUser, isHR } = useAuth()
  const { isDootask } = useDootaskContext()
  const { unreadInvitations, unreadEvaluations } = useUnreadContext()

  // 根据用户角色生成角色徽章
  const getRoleBadge = (role: string) => {
    switch (role) {
      case "hr":
        return <Badge variant="destructive">HR</Badge>
      case "manager":
        return <Badge variant="default">主管</Badge>
    }
  }

  // 根据用户角色动态生成导航菜单
  const navigation = useMemo(() => {
    let menus = []
    if (isHR) {
      menus = [
        {
          category: "核心功能",
          items: [
            { name: "仪表板", href: "/", icon: Home },
            {
              name: "考核管理",
              href: "/evaluations",
              icon: FileText,
              badge: unreadEvaluations > 0 ? unreadEvaluations : undefined,
            },
            {
              name: "邀请评分",
              href: "/invitations",
              icon: MessageSquare,
              badge: unreadInvitations > 0 ? unreadInvitations : undefined,
            },
            { name: "统计分析", href: "/statistics", icon: BarChart3 },
          ],
        },
        {
          category: "管理功能",
          items: [
            { name: "部门管理", href: "/departments", icon: Building },
            { name: "员工管理", href: "/employees", icon: Users },
            { name: "KPI模板", href: "/templates", icon: ClipboardList },
            { name: "绩效规则", href: "/performance-rules", icon: Scale },
          ],
        },
        {
          category: "其他功能",
          items: [
            { name: "备份还原", href: "/backup", icon: Database },
            { name: "系统设置", href: "/settings", icon: Settings, hidden: isDootask },
            { name: "帮助中心", href: "/help", icon: HelpCircle },
          ],
        },
      ]
    } else {
      menus = [
        {
          category: "我的功能",
          items: [
            {
              name: "考核管理",
              href: "/evaluations",
              icon: FileText,
              badge: unreadEvaluations > 0 ? unreadEvaluations : undefined,
            },
            {
              name: "邀请评分",
              href: "/invitations",
              icon: MessageSquare,
              badge: unreadInvitations > 0 ? unreadInvitations : undefined,
            },
          ],
        },
        {
          category: "系统功能",
          items: [
            { name: "系统设置", href: "/settings", icon: Settings },
            { name: "帮助中心", href: "/help", icon: HelpCircle },
          ],
        },
      ]
    }

    return menus
      .map(menu => ({
        ...menu,
        items: menu.items.filter(item => !("hidden" in item && item.hidden)),
      }))
      .filter(menu => menu.items.length > 0)
  }, [isHR, isDootask, unreadInvitations, unreadEvaluations])

  // 点击导航项时关闭移动端菜单
  const handleNavClick = () => {
    setIsMobileMenuOpen(false)
  }

  // 监听用户角色变化
  useEffect(() => {
    if (isHR) return

    if (!["/evaluations", "/invitations", "/settings", "/help", "/status"].includes(pathname)) {
      router.push("/evaluations")
    }
  }, [isHR, router, pathname])

  // 监听屏幕尺寸变化，在桌面端自动关闭移动菜单
  useEffect(() => {
    const handleResize = () => {
      if (window.innerWidth >= 1024) {
        setIsMobileMenuOpen(false)
      }
    }

    window.addEventListener("resize", handleResize)
    return () => window.removeEventListener("resize", handleResize)
  }, [setIsMobileMenuOpen])

  return (
    <>
      {/* 移动端遮罩层 */}
      {isMobileMenuOpen && (
        <div className="fixed inset-0 bg-black/50 z-10 lg:hidden" onClick={() => setIsMobileMenuOpen(false)} />
      )}

      {/* 侧边栏 */}
      <div
        className={cn(
          "fixed lg:static inset-y-0 left-0 z-20 w-64 h-screen lg:h-full bg-sidebar shadow-lg transform transition-transform duration-300 ease-in-out lg:transform-none flex flex-col pt-[var(--safe-area-top,0px)] pb-[var(--safe-area-bottom,0px)] lg:pt-0 lg:pb-0",
          isMobileMenuOpen ? "translate-x-0" : "-translate-x-full lg:translate-x-0"
        )}
      >
        {/* 头部 - 移动端显示关闭按钮 */}
        <div className="flex items-center justify-between pl-6 pr-2 lg:pr-6 py-4 lg:justify-start flex-shrink-0">
          <div className="flex flex-col gap-0.5">
            <h1 className="text-xl font-bold text-sidebar-foreground">KPI考核系统</h1>
            {currentUser && (
              <div className="text-sm text-sidebar-foreground/70 flex items-center gap-2">
                <div>{[currentUser.name, currentUser.department?.name].filter(Boolean).join(" - ")}</div>
                <div>{getRoleBadge(currentUser.role)}</div>
              </div>
            )}
          </div>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setIsMobileMenuOpen(false)}
            className="lg:hidden p-1 rounded-md hover:bg-sidebar-accent"
          >
            <X className="text-sidebar-foreground/70" />
          </Button>
        </div>

        {/* 导航菜单 - 可滚动区域 */}
        <nav className="flex-1 overflow-y-auto py-1">
          {navigation.map(section => (
            <div key={section.category} className="mb-2">
              {/* 分类标题 */}
              <div className="px-6 py-2 text-xs font-semibold text-sidebar-foreground/50 uppercase tracking-wider">
                {section.category}
              </div>

              {/* 分类下的菜单项 */}
              {section.items.map(item => {
                const isActive = pathname === item.href
                return (
                  <Link
                    key={item.name}
                    href={item.href}
                    onClick={handleNavClick}
                    className={cn(
                      "flex items-center justify-between px-6 py-3 text-sm font-medium transition-colors border-r-2 border-transparent",
                      isActive
                        ? "bg-sidebar-accent border-sidebar-primary"
                        : "text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-foreground"
                    )}
                  >
                    <div className="flex items-center">
                      <item.icon className="w-5 h-5 mr-3" />
                      {item.name}
                    </div>
                    {"badge" in item && item.badge && (
                      <Badge
                        variant="destructive"
                        className="ml-2 min-w-[20px] h-5 flex items-center justify-center text-xs"
                      >
                        {item.badge}
                      </Badge>
                    )}
                  </Link>
                )
              })}
            </div>
          ))}
        </nav>
      </div>
    </>
  )
}

// 移动端头部组件
export function MobileHeader({ onMenuClick }: { onMenuClick: () => void }) {
  const { unreadEvaluations, unreadInvitations } = useUnreadContext()
  return (
    <div className="lg:hidden bg-background shadow-sm border-b border-border flex-shrink-0">
      <div className="flex items-center justify-between px-4 py-1.5">
        <button onClick={onMenuClick} className="p-2 rounded-md hover:bg-accent relative">
          <Menu className="w-6 h-6 text-muted-foreground" />
          {unreadEvaluations + unreadInvitations > 0 && (
            <div className="absolute top-0 left-5.5 min-w-6 h-5 px-1.5 flex items-center justify-center text-xs bg-destructive/90 text-white rounded-md scale-95">
              {unreadEvaluations + unreadInvitations}
            </div>
          )}
        </button>
        <h1 className="text-lg font-semibold text-foreground">KPI考核系统</h1>
        <div className="w-10 h-10"></div>
      </div>
    </div>
  )
}
