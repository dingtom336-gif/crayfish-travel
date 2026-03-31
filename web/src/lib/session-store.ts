type SessionKey = "session_id" | "requirement" | "validation" | "packages" | "selected_package"

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
