export function formatYuan(cents: number): string {
  if (!cents && cents !== 0) return "0元"
  const yuan = cents / 100
  const formatted = yuan % 1 === 0
    ? yuan.toLocaleString("zh-CN")
    : yuan.toLocaleString("zh-CN", { minimumFractionDigits: 0, maximumFractionDigits: 0 })
  return `${formatted}元`
}

export function formatDate(dateStr: string, showYear = false): string {
  if (!dateStr) return "待定"
  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return "待定"
  const now = new Date()
  const sameYear = date.getFullYear() === now.getFullYear()
  return date.toLocaleDateString("zh-CN", {
    year: (!sameYear || showYear) ? "numeric" : undefined,
    month: "short",
    day: "numeric",
    weekday: "short",
  })
}

export const LOCK_DURATION_SECONDS = 15 * 60
