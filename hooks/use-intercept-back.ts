import { useEffect } from "react"
import { interceptBack } from "@dootask/tools"

let interceptRegistered = false
const handlers: (() => void)[] = []

function ensureInterceptRegistration() {
  if (typeof window === "undefined" || interceptRegistered) {
    return
  }

  try {
    interceptBack(() => {
      const handler = handlers.pop()
      if (handler) {
        handler()
        return true
      }
      return false
    })
    interceptRegistered = true
  } catch (error) {
    console.error("Failed to register back interception", error)
  }
}

export function useInterceptBack(handler?: () => void) {
  useEffect(() => {
    if (!handler) {
      return
    }

    ensureInterceptRegistration()
    handlers.push(handler)

    return () => {
      setTimeout(() => {
        handlers.splice(handlers.indexOf(handler), 1)
      }, 100);
    }
  }, [handler])
}

