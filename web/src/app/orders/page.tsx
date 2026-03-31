"use client"

import { useSearchParams } from "next/navigation"
import { useEffect, useState, useCallback, Suspense } from "react"
import { Card, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { api, type Order, type OrderStatus } from "@/lib/api"
import { formatYuan } from "@/lib/format"
import { OrdersSkeleton, EmptyState } from "@/components/Skeleton"

const STATUS_CONFIG: Record<string, { label: string; className: string }> = {
  created: { label: "已创建", className: "bg-[var(--color-trust-blue)]/10 text-[var(--color-trust-blue)] border-0" },
  confirmed: { label: "已确认", className: "bg-[var(--color-success-green)]/10 text-[var(--color-success-green)] border-0" },
  refund_requested: { label: "退款申请中", className: "bg-[var(--color-vibrant-orange)]/10 text-[var(--color-vibrant-orange)] border-0" },
  refunded: { label: "已退款", className: "bg-gray-100 text-gray-500 border-0" },
}

function StatusBadge({ status }: { status: string }) {
  const config = STATUS_CONFIG[status] ?? { label: status, className: "bg-gray-100 text-gray-500 border-0" }
  return (
    <Badge className={`rounded-full px-3 py-1 text-xs font-bold ${config.className}`}>
      {config.label}
    </Badge>
  )
}

function canRequestRefund(status: OrderStatus): boolean {
  return status === "created" || status === "confirmed"
}

interface OrderCardProps {
  order: Order
  onRequestRefund: (orderId: string) => void
  isRefunding: boolean
}

function OrderCard({ order, onRequestRefund, isRefunding }: OrderCardProps) {
  const isRefunded = order.status === "refunded"

  return (
    <Card className={`overflow-hidden shadow-sm transition-shadow hover:shadow-md ${isRefunded ? "opacity-80" : ""}`}>
      {/* Header: Order number + Status */}
      <div className="flex items-center justify-between border-b border-border/10 p-4 md:p-6">
        <div className="flex items-center gap-3 md:gap-4">
          <span className="text-xs font-bold tracking-wider text-muted-foreground md:text-sm">
            订单号: <span className="text-foreground">{order.order_no}</span>
          </span>
          <StatusBadge status={order.status} />
        </div>
        {!isRefunded && (
          <div className="hidden items-center gap-1 text-sm font-medium text-[var(--color-success-green)] md:flex">
            <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z" />
            </svg>
            安全保障中
          </div>
        )}
      </div>

      {/* Body */}
      <CardContent className="p-4 md:p-6">
        {/* Mobile: compact layout */}
        <div className="flex gap-4 md:hidden">
          <div className="flex flex-col justify-center gap-1">
            <h2 className="text-sm font-semibold leading-tight">{order.package_title}</h2>
            <div className="flex items-center gap-1 text-[11px] text-muted-foreground">
              <span>{order.destination}</span>
              <span className="mx-1 opacity-30">|</span>
              <span>{order.start_date} - {order.end_date}</span>
            </div>
          </div>
        </div>

        {/* Desktop: expanded layout */}
        <div className="hidden md:block">
          <h3 className="mb-1 text-xl font-bold">{order.package_title}</h3>
          <p className="flex items-center gap-2 text-muted-foreground">
            {order.destination} | {order.start_date} - {order.end_date}
          </p>
        </div>

        {/* Price + Refund section */}
        <div className="mt-4 flex items-end justify-between border-t border-border/10 pt-4">
          <div className="flex flex-col">
            <span className="text-[10px] leading-none text-muted-foreground">
              基础价格 + 退改保障费
            </span>
            <span className={`mt-1 text-xl font-bold md:text-2xl ${
              isRefunded
                ? "text-muted-foreground/40"
                : "text-[var(--color-trust-blue)]"
            }`}>
              {formatYuan(order.total_amount_cents)}
            </span>
            {!isRefunded && order.base_price_cents > 0 && (
              <span className="mt-0.5 hidden text-xs text-muted-foreground md:block">
                基础价格 {formatYuan(order.base_price_cents)} + 退改保障费 {formatYuan(order.refund_guarantee_fee_cents)}
              </span>
            )}
          </div>

          {canRequestRefund(order.status) && (
            <Button
              variant="outline"
              size="sm"
              className="rounded-full border-red-200 px-5 text-xs font-bold text-red-500 hover:border-red-500 hover:text-red-600 md:rounded-xl md:px-6 md:py-2.5"
              onClick={() => onRequestRefund(order.id)}
              disabled={isRefunding}
            >
              {isRefunding ? "申请中..." : "申请退款"}
            </Button>
          )}

          {order.status === "refunded" && (
            <Button
              variant="outline"
              size="sm"
              className="cursor-not-allowed rounded-full bg-muted px-5 text-xs font-bold text-muted-foreground md:rounded-xl md:px-6 md:py-2.5"
              disabled
            >
              已退款
            </Button>
          )}

          {order.status === "refund_requested" && (
            <Badge className="rounded-full bg-[var(--color-vibrant-orange)]/10 px-4 py-1.5 text-xs font-bold text-[var(--color-vibrant-orange)] border-0">
              处理中
            </Badge>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

function OrdersContent() {
  const searchParams = useSearchParams()
  const sessionId = searchParams.get("session_id") ?? ""

  const [orders, setOrders] = useState<Order[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [refundingId, setRefundingId] = useState<string | null>(null)
  const [justCreated, setJustCreated] = useState(false)

  useEffect(() => {
    if (!sessionId) {
      setLoading(false)
      return
    }

    let timer: ReturnType<typeof setTimeout>

    async function loadOrders() {
      try {
        const result = await api.listOrders(sessionId)
        setOrders(result ?? [])
        if (result && result.length > 0) {
          const newest = result[0]
          const createdAt = new Date(newest.created_at).getTime()
          if (Date.now() - createdAt < 30000) {
            setJustCreated(true)
            timer = setTimeout(() => setJustCreated(false), 3000)
          }
        }
      } catch (err: unknown) {
        const message = err instanceof Error ? err.message : "加载订单失败"
        setError(message)
      } finally {
        setLoading(false)
      }
    }

    loadOrders()
    return () => clearTimeout(timer)
  }, [sessionId])

  const handleRequestRefund = useCallback(async (orderId: string) => {
    setRefundingId(orderId)
    setError(null)
    try {
      await api.requestRefund(orderId)
      setOrders((prev) =>
        prev.map((o) =>
          o.id === orderId ? { ...o, status: "refund_requested" as const } : o
        )
      )
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "退款申请失败"
      setError(message)
    } finally {
      setRefundingId(null)
    }
  }, [])

  if (!sessionId) {
    return (
      <main className="min-h-screen bg-[#f9f9ff]">
        <div className="mx-auto max-w-2xl px-4 py-16 text-center">
          <h2 className="text-xl font-semibold text-gray-600">未找到会话</h2>
          <p className="mt-2 text-muted-foreground">请从首页重新开始</p>
        </div>
      </main>
    )
  }

  return (
    <main className="min-h-screen bg-[#f9f9ff]">
      <div className="mx-auto max-w-3xl px-4 py-8 md:px-8 md:pt-12 md:pb-20">
        {/* Page Title */}
        <header className="mb-8 md:mb-12">
          <h1 className="text-2xl font-extrabold tracking-tight text-[var(--color-trust-blue)] md:text-4xl">
            我的订单
          </h1>
          <p className="mt-1 text-muted-foreground md:mt-2 md:text-lg">
            管理您的旅行计划与行程订单
          </p>
        </header>

        {justCreated && (
          <div className="mb-4 rounded-lg border border-[var(--color-success-green)]/30 bg-[var(--color-success-green)]/10 p-4 text-center">
            <p className="text-sm font-semibold text-[var(--color-success-green)]">订单创建成功!</p>
          </div>
        )}

        {error && (
          <div className="mb-4 rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700">
            {error}
          </div>
        )}

        {loading && <OrdersSkeleton />}

        {!loading && orders.length === 0 && (
          <EmptyState icon="📦" title="暂无订单" description="支付完成后订单将在此显示" />
        )}

        {!loading && orders.length > 0 && (
          <div className="space-y-4 md:space-y-6">
            {orders.map((order) => (
              <OrderCard
                key={order.id}
                order={order}
                onRequestRefund={handleRequestRefund}
                isRefunding={refundingId === order.id}
              />
            ))}
          </div>
        )}
      </div>
    </main>
  )
}

export default function OrdersPage() {
  return (
    <Suspense fallback={<main className="min-h-screen bg-[#f9f9ff]"><OrdersSkeleton /></main>}>
      <OrdersContent />
    </Suspense>
  )
}
