type SessionKey = "session_id" | "requirement" | "validation" | "packages" | "selected_package"

const LOCAL_STORAGE_SESSION_KEY = "crayfish_session_id"

export function setSessionData<T>(key: SessionKey, data: T): void {
  if (typeof window === "undefined") return
  sessionStorage.setItem(key, JSON.stringify(data))
}

export function getSessionData<T>(key: SessionKey): T | null {
  if (typeof window === "undefined") return null
  const raw = sessionStorage.getItem(key)
  if (!raw) return null
  try {
    return JSON.parse(raw) as T
  } catch {
    return null
  }
}

export function clearSessionData(): void {
  if (typeof window === "undefined") return
  const keys: SessionKey[] = ["session_id", "requirement", "validation", "packages", "selected_package"]
  for (const key of keys) {
    sessionStorage.removeItem(key)
  }
}

export function setPersistedSessionId(id: string): void {
  if (typeof window === "undefined") return
  localStorage.setItem(LOCAL_STORAGE_SESSION_KEY, id)
  sessionStorage.setItem("session_id", JSON.stringify(id))
}

export function getPersistedSessionId(): string | null {
  if (typeof window === "undefined") return null
  // Try sessionStorage first, then localStorage
  const fromSession = sessionStorage.getItem("session_id")
  if (fromSession) {
    try { return JSON.parse(fromSession) } catch { /* fall through */ }
  }
  return localStorage.getItem(LOCAL_STORAGE_SESSION_KEY)
}
