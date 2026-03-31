"use client"

import { useSearchParams, useRouter } from "next/navigation"
import { useEffect, useState, useCallback, useRef, Suspense } from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { api, type RankedQuote, type PaymentResponse } from "@/lib/api"
import { getSessionData } from "@/lib/session-store"
import { formatYuan, LOCK_DURATION_SECONDS } from "@/lib/format"
import { CountdownTimer } from "@/components/CountdownTimer"
import { CardSkeleton, EmptyState } from "@/components/Skeleton"

function PaymentContent() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const sessionId = searchParams.get("session_id") ?? ""
  const quoteId = searchParams.get("quote_id") ?? ""

  const [method, setMethod] = useState<"qr" | "voice_token">("qr")
  const [payment, setPayment] = useState<PaymentResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [checking, setChecking] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const [expiresAt] = useState(() => Date.now() + LOCK_DURATION_SECONDS * 1000)
  const [selectedPackage] = useState<RankedQuote | null>(() => getSessionData<RankedQuote>("selected_package"))

  const createPayment = useCallback(async (payMethod: "qr" | "voice_token") => {
    if (!sessionId || !quoteId) return
    setLoading(true)
    setError(null)
    try {
      const result = await api.createPayment(sessionId, quoteId, payMethod)
      setPayment(result)
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "创建支付失败"
      setError(message)
    } finally {
      setLoading(false)
    }
  }, [sessionId, quoteId])

  const hasMounted = useRef(false)
  useEffect(() => {
    if (hasMounted.current) return
    hasMounted.current = true
    createPayment(method)
  }, [createPayment, method])

  const handleMethodChange = useCallback((newMethod: "qr" | "voice_token") => {
    setMethod(newMethod)
    setPayment(null)
    createPayment(newMethod)
  }, [createPayment])

  const handleCheckStatus = useCallback(async () => {
    if (!sessionId) return
    setChecking(true)
    try {
      const orders = await api.listOrders(sessionId)
      if (orders && orders.length > 0) {
        router.push(`/orders?session_id=${sessionId}`)
      } else {
        setError("支付尚未确认，请稍后再试")
      }
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "查询状态失败"
      setError(message)
    } finally {
      setChecking(false)
    }
  }, [sessionId, router])

  const handleCopyToken = useCallback(async () => {
    if (!payment?.voice_token) return
    try {
      await navigator.clipboard.writeText(payment.voice_token)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      setError("复制失败，请手动复制口令")
    }
  }, [payment])

  if (!sessionId || !quoteId) {
    return (
      <main className="min-h-screen bg-gray-50">
        <div className="mx-auto max-w-2xl px-4 py-8">
          <EmptyState icon="🔗" title="缺少会话信息" description="请返回选择方案" />
        </div>
      </main>
    )
  }

  return (
    <main className="min-h-screen bg-gray-50">
      <div className="mx-auto max-w-2xl px-4 py-8">
        <div className="mb-6 flex items-center justify-between">
          <h1 className="text-xl font-bold text-[#1a73e8]">支付</h1>
          <CountdownTimer expiresAt={expiresAt} />
        </div>

        {selectedPackage && (
          <Card className="mb-6">
            <CardHeader>
              <CardTitle className="text-base">已选方案</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <div className="font-semibold">{selectedPackage.package_title}</div>
              <div className="text-sm text-muted-foreground">
                {selectedPackage.destination} - {selectedPackage.duration_days}天{selectedPackage.duration_nights}晚
              </div>
              <Separator />
              <div className="flex items-baseline justify-between">
                <span className="text-sm text-muted-foreground">总价</span>
                <span className="text-xl font-bold text-[#1a73e8]">{formatYuan(selectedPackage.total_price_cents)}</span>
              </div>
              <div className="text-xs text-muted-foreground">
                基础价格 {formatYuan(selectedPackage.base_price_cents)} + 退改保障费 {formatYuan(selectedPackage.refund_guarantee_fee_cents)}
              </div>
            </CardContent>
          </Card>
        )}

        <Card className="mb-6">
          <CardHeader>
            <CardTitle className="text-base">支付方式</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex gap-2">
              <Button
                variant={method === "qr" ? "default" : "outline"}
                className={method === "qr" ? "bg-[#1a73e8] text-white" : ""}
                onClick={() => handleMethodChange("qr")}
                disabled={loading}
              >
                扫码支付
              </Button>
              <Button
                variant={method === "voice_token" ? "default" : "outline"}
                className={method === "voice_token" ? "bg-[#1a73e8] text-white" : ""}
                onClick={() => handleMethodChange("voice_token")}
                disabled={loading}
              >
                吱口令
              </Button>
            </div>

            {loading && (
              <div className="py-4">
                <CardSkeleton />
              </div>
            )}

            {!loading && payment && method === "qr" && (
              <div className="space-y-3">
                <div className="flex aspect-square w-full items-center justify-center rounded-lg border-2 border-dashed border-gray-300 bg-gray-50 p-4 sm:max-w-[240px]">
                  <div className="text-center">
                    <p className="text-xs text-muted-foreground">支付二维码：</p>
                    <p className="mt-1 break-all text-xs font-mono text-[#1a73e8]">{payment.qr_code_url}</p>
                  </div>
                </div>
                {payment.out_trade_no && (
                  <p className="text-xs text-muted-foreground">交易号：{payment.out_trade_no}</p>
                )}
              </div>
            )}

            {!loading && payment && method === "voice_token" && (
              <div className="space-y-3">
                <div className="relative">
                  <div className="rounded-lg border bg-gray-50 p-6 text-center">
                    <p className="text-xs text-muted-foreground mb-2">吱口令</p>
                    <p className="text-3xl font-bold tracking-widest text-[#1a73e8]">{payment.voice_token}</p>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    className="absolute top-2 right-2"
                    onClick={handleCopyToken}
                  >
                    {copied ? "已复制" : "复制"}
                  </Button>
                </div>
                {payment.out_trade_no && (
                  <p className="text-xs text-muted-foreground">交易号：{payment.out_trade_no}</p>
                )}
              </div>
            )}
          </CardContent>
        </Card>

        {error && (
          <div className="mb-4 rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700">
            {error}
          </div>
        )}

        <Button
          className="w-full bg-[#1a73e8] text-white hover:bg-[#1565c0]"
          size="lg"
          onClick={handleCheckStatus}
          disabled={checking || loading}
        >
          {checking ? "查询中..." : "查看支付状态"}
        </Button>
      </div>
    </main>
  )
}

function PaymentSkeleton() {
  return (
    <main className="min-h-screen bg-gray-50">
      <div className="mx-auto max-w-2xl px-4 py-8 space-y-6">
        <CardSkeleton />
        <CardSkeleton />
      </div>
    </main>
  )
}

export default function PaymentPage() {
  return (
    <Suspense fallback={<PaymentSkeleton />}>
      <PaymentContent />
    </Suspense>
  )
}
