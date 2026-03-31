"use client"

import { useSearchParams } from "next/navigation"
import { useEffect, useState, useCallback, Suspense } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { api, type Order, type OrderStatus } from "@/lib/api"
import { formatYuan } from "@/lib/format"
import { OrdersSkeleton, EmptyState } from "@/components/Skeleton"

const STATUS_CONFIG: Record<string, { label: string; className: string }> = {
  created: { label: "已创建", className: "bg-[#1a73e8] text-white" },
  confirmed: { label: "已确认", className: "bg-[#34a853] text-white" },
  refund_requested: { label: "退款申请中", className: "bg-[#ff6d3f] text-white" },
  refunded: { label: "已退款", className: "bg-red-500 text-white" },
}

function StatusBadge({ status }: { status: string }) {
  const config = STATUS_CONFIG[status] ?? { label: status, className: "bg-gray-500 text-white" }
  return <Badge className={config.className}>{config.label}</Badge>
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
  return (
    <Card>
      <CardHeader>
        <div className="flex flex-wrap items-start justify-between gap-2">
          <div className="min-w-0 flex-1">
            <CardTitle className="truncate text-base font-bold">{order.order_no}</CardTitle>
            <p className="mt-1 text-sm text-muted-foreground">{order.package_title}</p>
          </div>
          <StatusBadge status={order.status} />
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="grid grid-cols-2 gap-2 text-sm">
          <div>
            <span className="text-muted-foreground">目的地</span>
            <p className="font-medium">{order.destination}</p>
          </div>
          <div>
            <span className="text-muted-foreground">日期</span>
            <p className="font-medium">{order.start_date} - {order.end_date}</p>
          </div>
        </div>
        <Separator />
        <div className="flex items-baseline justify-between">
          <span className="text-sm text-muted-foreground">总金额</span>
          <span className="text-lg font-bold text-[#1a73e8]">{formatYuan(order.total_amount_cents)}</span>
        </div>
        {canRequestRefund(order.status) && (
          <Button
            variant="destructive"
            size="sm"
            className="w-full"
            onClick={() => onRequestRefund(order.id)}
            disabled={isRefunding}
          >
            {isRefunding ? "申请中..." : "申请退款"}
          </Button>
        )}
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
          o.id === orderId ? { ...o, status: "refund_requested" } : o
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
      <main className="min-h-screen bg-gray-50">
        <div className="mx-auto max-w-2xl px-4 py-16 text-center">
          <h2 className="text-xl font-semibold text-gray-600">未找到会话</h2>
          <p className="mt-2 text-muted-foreground">请从首页重新开始</p>
        </div>
      </main>
    )
  }

  return (
    <main className="min-h-screen bg-gray-50">
      <div className="mx-auto max-w-2xl px-4 py-8">
        <h1 className="mb-6 text-xl font-bold text-[#1a73e8]">我的订单</h1>

        {justCreated && (
          <div className="mb-4 rounded-lg border border-[#34a853]/30 bg-[#34a853]/10 p-4 text-center">
            <p className="text-sm font-semibold text-[#34a853]">订单创建成功!</p>
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
          <div className="space-y-4">
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
    <Suspense fallback={<main className="min-h-screen bg-gray-50"><OrdersSkeleton /></main>}>
      <OrdersContent />
    </Suspense>
  )
}
