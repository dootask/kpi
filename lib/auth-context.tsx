"use client"

import { createContext, useContext, useEffect, useState } from "react"
import { authApi, LoginRequest, RegisterRequest, type AuthUser } from "@/lib/api"

interface AuthContextType {
  user: AuthUser | null
  loading: boolean
  login: (data: LoginRequest) => Promise<void>
  register: (data: RegisterRequest) => Promise<void>
  logout: () => void
  isAuthenticated: boolean
  refreshUser: () => Promise<void>
  // 角色判断
  isManager: boolean
  isHR: boolean
  isEmployee: boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null)
  const [loading, setLoading] = useState(true)

  // 初始化时检查用户状态
  useEffect(() => {
    const initializeAuth = async () => {
      try {
        const token = authApi.getToken()
        const savedUser = authApi.getUser()
        
        if (token && savedUser) {
          setUser(savedUser)
          // 验证token是否仍然有效
          try {
            const response = await authApi.getCurrentUser()
            setUser(response.data)
            authApi.setAuth(token, response.data)
          } catch {
            // Token无效，清除本地存储
            authApi.logout()
            setUser(null)
          }
        }
      } catch (error) {
        console.error("初始化认证失败:", error)
        authApi.logout()
        setUser(null)
      } finally {
        setLoading(false)
      }
    }

    initializeAuth()
  }, [])

  const login = async (data: LoginRequest) => {
    try {
      const response = await authApi.login(data)
      authApi.setAuth(response.token, response.user)
      setUser(response.user)
    } catch (error) {
      throw error
    }
  }

  const register = async (data: {
    name: string
    email: string
    password: string
    position: string
    department_id: number
  }) => {
    try {
      const response = await authApi.register(data)
      authApi.setAuth(response.token, response.user)
      setUser(response.user)
    } catch (error) {
      throw error
    }
  }

  const logout = () => {
    authApi.logout()
    setUser(null)
  }

  const refreshUser = async () => {
    try {
      const response = await authApi.getCurrentUser()
      setUser(response.data)
      const token = authApi.getToken()
      if (token) {
        authApi.setAuth(token, response.data)
      }
    } catch (error) {
      console.error("刷新用户信息失败:", error)
      logout()
    }
  }

  const value: AuthContextType = {
    user,
    loading,
    login,
    register,
    logout,
    isAuthenticated: !!user,
    refreshUser,
    // 角色判断
    isManager: user?.role === "manager",
    isHR: user?.role === "hr",
    isEmployee: user?.role === "employee",
  }

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider")
  }
  return context
} 