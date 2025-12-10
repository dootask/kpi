"use client"

import { createContext, useContext, useEffect, useState, useCallback, useRef } from "react"
import { useAuth } from "./auth-context"
import { sseApi } from "./api"
import { storage } from "./storage"

// 通知事件类型
export type NotificationEventType =
  | "evaluation_created"
  | "evaluation_updated"
  | "evaluation_deleted"
  | "evaluation_status_changed"
  | "invitation_created"
  | "invitation_updated"
  | "invitation_deleted"
  | "invitation_status_changed"
  | "invited_score_updated"
  | "self_score_updated"
  | "manager_score_updated"
  | "hr_score_updated"

// 通知事件数据结构
export interface NotificationEventData {
  id: number
  employee_id: number
  operator_id: number
  operator_name: string
  message: string
  timestamp: string
  payload?: Record<string, unknown>
}

// 连接确认数据
export interface ConnectionData {
  user_id: number
  timestamp: string
}

// 心跳数据
export interface HeartbeatData {
  timestamp: string
}

// SSE消息结构
export interface SSEMessage {
  type: NotificationEventType | "connected" | "heartbeat"
  data: NotificationEventData | ConnectionData | HeartbeatData
  timestamp: string
  id: string
}

// 通知上下文类型
interface NotificationContextType {
  isConnected: boolean
  isConnecting: boolean
  lastMessage: SSEMessage | null
  connectionRetries: number
  connect: () => void
  disconnect: () => void
  onMessage: (callback: (message: SSEMessage) => void) => () => void
}

// 最大重试次数，-1表示无限重试
const MAX_RETRIES = -1

// 重试间隔（秒）
const RETRY_INTERVALS = [1, 2, 5, 10, 30]

// 是否输出日志
const LOG_ENABLED = process.env.NODE_ENV === "development" ? true : false

// 日志输出
const logger = {
  info: (message: string, ...args: unknown[]) => {
    if (LOG_ENABLED) {
      console.log(message, ...args)
    }
  },
  error: (message: string, ...args: unknown[]) => {
    if (LOG_ENABLED) {
      console.error(message, ...args)
    }
  },
}

const NotificationContext = createContext<NotificationContextType | undefined>(undefined)

export function NotificationProvider({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, user } = useAuth()
  const [isConnected, setIsConnected] = useState(false)
  const [isConnecting, setIsConnecting] = useState(false)
  const [lastMessage, setLastMessage] = useState<SSEMessage | null>(null)
  const [connectionRetries, setConnectionRetries] = useState(0)

  const eventSourceRef = useRef<EventSource | null>(null)
  const messageCallbacksRef = useRef<Set<(message: SSEMessage) => void>>(new Set())
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)

  // 连接SSE
  const connect = useCallback(() => {
    if (!isAuthenticated || !user || eventSourceRef.current) return

    setIsConnecting(true)

    try {
      const token = storage.getItem("auth_token")
      if (!token) {
        console.error("No auth token found")
        setIsConnecting(false)
        return
      }

      const eventSource = sseApi.getStream()

      eventSourceRef.current = eventSource

      eventSource.onopen = () => {
        logger.info("SSE连接已建立")
        setIsConnected(true)
        setIsConnecting(false)
        setConnectionRetries(0)
      }

      eventSource.onmessage = event => {
        try {
          const message: SSEMessage = JSON.parse(event.data)
          logger.info("收到SSE消息:", message)

          setLastMessage(message)

          // 处理不同类型的消息
          if (message.type === "connected") {
            logger.info("SSE连接确认:", message.data)
          } else if (message.type === "heartbeat") {
            // 心跳消息，不需要特殊处理
            logger.info("收到心跳消息")
          } else {
            // 通知所有监听器
            messageCallbacksRef.current.forEach(callback => {
              try {
                callback(message)
              } catch (error) {
                logger.error("消息回调处理错误:", error)
              }
            })
          }
        } catch (error) {
          logger.error("解析SSE消息失败:", error)
        }
      }

      eventSource.onerror = error => {
        logger.error("SSE连接错误:", error)
        setIsConnected(false)
        setIsConnecting(false)

        if (eventSource.readyState === EventSource.CLOSED) {
          logger.info("SSE连接已关闭，尝试重连...")

          // 自动重连
          if (MAX_RETRIES === -1 || connectionRetries < MAX_RETRIES) {
            const retryInterval = RETRY_INTERVALS[Math.min(connectionRetries, RETRY_INTERVALS.length - 1)]
            logger.info(`${retryInterval}秒后重试连接...`)

            reconnectTimeoutRef.current = setTimeout(() => {
              setConnectionRetries(prev => prev + 1)
              // 先断开现有连接
              if (eventSourceRef.current) {
                eventSourceRef.current.close()
                eventSourceRef.current = null
              }
              connect()
            }, retryInterval * 1000)
          } else {
            logger.error("SSE连接重试次数已达上限")
          }
        }
      }
    } catch (error) {
      logger.error("创建SSE连接失败:", error)
      setIsConnecting(false)
    }
  }, [isAuthenticated, user, connectionRetries])

  // 断开连接
  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }

    if (eventSourceRef.current) {
      eventSourceRef.current.close()
      eventSourceRef.current = null
    }

    setIsConnected(false)
    setIsConnecting(false)
    setConnectionRetries(0)
    setLastMessage(null)

    logger.info("SSE连接已断开")
  }, [])

  // 添加消息监听器
  const onMessage = useCallback((callback: (message: SSEMessage) => void) => {
    messageCallbacksRef.current.add(callback)

    // 返回移除监听器的函数
    return () => {
      messageCallbacksRef.current.delete(callback)
    }
  }, [])

  // 自动连接管理
  useEffect(() => {
    if (isAuthenticated && user) {
      connect()
    } else {
      disconnect()
    }

    return () => {
      disconnect()
    }
  }, [isAuthenticated, user, connect, disconnect])

  // 页面可见性变化时重连
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState === "visible" && isAuthenticated && user && !isConnected && !isConnecting) {
        console.log("页面可见，重新连接SSE...")
        connect()
      }
    }

    document.addEventListener("visibilitychange", handleVisibilityChange)
    return () => document.removeEventListener("visibilitychange", handleVisibilityChange)
  }, [isAuthenticated, user, isConnected, isConnecting, connect])

  // 在线状态变化时重连
  useEffect(() => {
    const handleOnline = () => {
      if (isAuthenticated && user && !isConnected && !isConnecting) {
        console.log("网络恢复，重新连接SSE...")
        connect()
      }
    }

    const handleOffline = () => {
      console.log("网络离线，断开SSE连接...")
      disconnect()
    }

    window.addEventListener("online", handleOnline)
    window.addEventListener("offline", handleOffline)

    return () => {
      window.removeEventListener("online", handleOnline)
      window.removeEventListener("offline", handleOffline)
    }
  }, [isAuthenticated, user, isConnected, isConnecting, connect, disconnect])

  // 清理资源
  useEffect(() => {
    return () => {
      disconnect()
    }
  }, [disconnect])

  const value: NotificationContextType = {
    isConnected,
    isConnecting,
    lastMessage,
    connectionRetries,
    connect,
    disconnect,
    onMessage,
  }

  return <NotificationContext.Provider value={value}>{children}</NotificationContext.Provider>
}

export function useNotification() {
  const context = useContext(NotificationContext)
  if (context === undefined) {
    throw new Error("useNotification must be used within a NotificationProvider")
  }
  return context
}
