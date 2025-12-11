"use client"

import { createContext, useContext, useState } from "react"
import { usePathname } from "next/navigation"
import ProtectedRoute from "@/components/protected-route"
import { MobileHeader } from "@/components/sidebar"
import { Sidebar } from "@/components/sidebar"

interface LayoutContextType {
  isMobileMenuOpen: boolean
  setIsMobileMenuOpen: (isOpen: boolean) => void
}

const LayoutContext = createContext<LayoutContextType | undefined>(undefined)

export function LayoutProvider({ children }: { children: React.ReactNode }) {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const pathname = usePathname()
  const isAuthPage = pathname.startsWith("/auth")

  if (isAuthPage) {
    // 认证页面布局（不需要侧边栏）
    return <ProtectedRoute requireAuth={false}>{children}</ProtectedRoute>
  }

  // 主要应用程序布局（需要认证）
  return (
    <ProtectedRoute requireAuth={true}>
      <div className="h-screen flex flex-col lg:flex-row overflow-hidden">
        {/* 移动端头部 */}
        <MobileHeader onMenuClick={() => setIsMobileMenuOpen(true)} />

        {/* 侧边栏 */}
        <Sidebar isMobileMenuOpen={isMobileMenuOpen} setIsMobileMenuOpen={setIsMobileMenuOpen} />

        {/* 主内容区域 */}
        <main className="flex-1 bg-muted/30 min-w-0 overflow-y-auto lg:pt-[var(--safe-area-top,0px)] pb-[var(--safe-area-bottom,0px)]">
          <div className="p-4 lg:p-6">
            <div className="max-w-7xl mx-auto">{children}</div>
          </div>
        </main>
      </div>
    </ProtectedRoute>
  )
}

export function useLayoutContext() {
  const context = useContext(LayoutContext)
  if (context === undefined) {
    throw new Error("useLayoutContext must be used within an LayoutProvider")
  }
  return context
}
