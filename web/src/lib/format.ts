export function formatYuan(cents: number): string {
  const yuan = cents / 100
  const formatted = yuan % 1 === 0
    ? yuan.toLocaleString("zh-CN")
    : yuan.toLocaleString("zh-CN", { minimumFractionDigits: 0, maximumFractionDigits: 0 })
  return `${formatted}元`
}

export function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleDateString("zh-CN", {
    year: "numeric",
    month: "long",
    day: "numeric",
    weekday: "short",
  })
}

export const LOCK_DURATION_SECONDS = 15 * 60
