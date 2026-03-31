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
// StepProgress is used elsewhere with 3 steps; payment uses a simple inline indicator
import { CardSkeleton, EmptyState } from "@/components/Skeleton"
import { QRCodeSVG } from "qrcode.react"

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

  // Auto-poll for payment completion
  useEffect(() => {
    if (!payment || !sessionId) return
    const interval = setInterval(async () => {
      try {
        const result = await api.listOrders(sessionId)
        if (result && result.length > 0) {
          clearInterval(interval)
          router.push(`/orders?session_id=${sessionId}`)
        }
      } catch { /* ignore polling errors */ }
    }, 3000)
    return () => clearInterval(interval)
  }, [payment, sessionId, router])

  if (!sessionId || !quoteId) {
    return (
      <main className="min-h-screen bg-[#f9f9ff]">
        <div className="mx-auto max-w-2xl px-4 py-8">
          <EmptyState icon="🔗" title="缺少会话信息" description="请返回选择方案" />
        </div>
      </main>
    )
  }

  return (
    <>
      {/* Desktop Layout */}
      <main className="hidden min-h-screen bg-[#f9f9ff] md:block">
        <div className="mx-auto max-w-5xl px-6 pt-8 pb-20">
          <div className="flex flex-row gap-8">
            {/* Left Column (60%) */}
            <div className="w-[60%] space-y-8">
              <div className="flex items-center gap-4">
                <span className="text-sm font-semibold text-muted-foreground">Step 4 of 5</span>
                <div className="flex gap-2">
                  {[1, 2, 3, 4, 5].map((s) => (
                    <div
                      key={s}
                      className={`h-2.5 w-2.5 rounded-full ${
                        s === 4 ? "bg-[var(--color-trust-blue)]" : "bg-muted-foreground/20"
                      }`}
                    />
                  ))}
                </div>
              </div>

              {/* Payment Method Card */}
              <Card className="border-border/10 shadow-sm">
                <CardContent className="p-8">
                  {/* Tabs */}
                  <div className="mb-10 flex border-b border-border/20">
                    <button
                      className={`px-8 py-3 text-lg font-semibold transition-colors ${
                        method === "qr"
                          ? "border-b-2 border-[var(--color-trust-blue)] text-[var(--color-trust-blue)]"
                          : "text-muted-foreground hover:text-foreground"
                      }`}
                      onClick={() => handleMethodChange("qr")}
                      disabled={loading}
                    >
                      扫码支付
                    </button>
                    <button
                      className={`px-8 py-3 text-lg font-medium transition-colors ${
                        method === "voice_token"
                          ? "border-b-2 border-[var(--color-trust-blue)] text-[var(--color-trust-blue)]"
                          : "text-muted-foreground hover:text-foreground"
                      }`}
                      onClick={() => handleMethodChange("voice_token")}
                      disabled={loading}
                    >
                      吱口令
                    </button>
                  </div>

                  {/* QR / Token Area */}
                  <div className="flex flex-col items-center justify-center space-y-6 py-4">
                    {loading && <CardSkeleton />}

                    {!loading && payment && method === "qr" && (
                      <div className="relative rounded-xl border-2 border-dashed border-muted-foreground/20 bg-white p-6">
                        <QRCodeSVG value={payment.qr_code_url} size={240} level="M" />
                      </div>
                    )}

                    {!loading && payment && method === "voice_token" && (
                      <div className="w-full max-w-sm space-y-3">
                        <div className="relative rounded-xl border bg-muted/30 p-8 text-center">
                          <p className="mb-3 text-sm text-muted-foreground">吱口令</p>
                          <p className="text-4xl font-bold tracking-widest text-[var(--color-trust-blue)]">
                            {payment.voice_token}
                          </p>
                          <Button
                            variant="outline"
                            size="sm"
                            className="absolute right-3 top-3"
                            onClick={handleCopyToken}
                          >
                            {copied ? "已复制" : "复制"}
                          </Button>
                        </div>
                      </div>
                    )}

                    {payment?.out_trade_no && (
                      <div className="text-center">
                        <p className="text-sm tracking-wider text-muted-foreground">
                          交易单号: <span className="font-mono">{payment.out_trade_no}</span>
                        </p>
                      </div>
                    )}

                    {error && (
                      <div className="w-full max-w-sm rounded-lg border border-red-200 bg-red-50 p-3 text-center text-sm text-red-700">
                        {error}
                      </div>
                    )}

                    <div className="w-full max-w-sm pt-6">
                      <Button
                        className="w-full bg-gradient-to-br from-[var(--color-trust-blue)] to-[var(--color-trust-blue-dark)] py-6 text-lg font-bold text-white shadow-lg shadow-[var(--color-trust-blue)]/20 hover:shadow-xl"
                        onClick={handleCheckStatus}
                        disabled={checking || loading}
                      >
                        {checking ? "查询中..." : "查看支付状态"}
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>

            {/* Right Column (40%) */}
            <div className="w-[40%]">
              <div className="sticky top-8 space-y-6">
                {/* Selected Package Summary */}
                {selectedPackage && (
                  <Card className="overflow-hidden border-border/10 shadow-sm">
                    <CardHeader className="border-b border-border/10 bg-muted/30">
                      <CardTitle className="text-xl font-bold">已选方案</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-6 p-6">
                      <div className="space-y-4">
                        <div className="flex flex-col gap-1">
                          <h3 className="text-lg font-bold leading-tight">
                            {selectedPackage.package_title}
                          </h3>
                          <div className="mt-1 flex items-center gap-3">
                            <span className="inline-flex items-center gap-1 text-sm text-muted-foreground">
                              目的地: {selectedPackage.destination}
                            </span>
                            <span className="inline-flex items-center gap-1 text-sm text-muted-foreground">
                              时长: {selectedPackage.duration_days}天{selectedPackage.duration_nights}晚
                            </span>
                          </div>
                        </div>
                        <Separator />
                        <div className="flex items-baseline justify-between">
                          <span className="font-medium text-muted-foreground">总计金额</span>
                          <span className="text-4xl font-extrabold tracking-tighter text-[var(--color-trust-blue)]">
                            {formatYuan(selectedPackage.total_price_cents)}
                          </span>
                        </div>
                        <p className="text-right text-xs text-muted-foreground">
                          包含: 基础价格 {formatYuan(selectedPackage.base_price_cents)} + 退改保障费{" "}
                          {formatYuan(selectedPackage.refund_guarantee_fee_cents)}
                        </p>
                      </div>

                      {/* Countdown Timer */}
                      <div className="flex items-center justify-center gap-3 rounded-xl bg-[var(--color-vibrant-orange)]/10 p-4 font-bold text-[var(--color-vibrant-orange)]">
                        <span>剩余支付时间</span>
                        <CountdownTimer expiresAt={expiresAt} />
                      </div>
                    </CardContent>
                  </Card>
                )}

                {/* Security Badge */}
                <div className="flex items-center gap-4 px-4 text-[var(--color-success-green)]">
                  <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z" />
                  </svg>
                  <span className="text-sm font-semibold">支付保障 - 官方严选行程</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>

      {/* Mobile Layout */}
      <main className="min-h-screen bg-[#f9f9ff] pb-24 md:hidden">
        {/* Top Bar */}
        <header className="fixed top-0 z-50 flex h-14 w-full items-center justify-between bg-white/70 px-4 shadow-sm backdrop-blur-xl">
          <button
            className="text-[var(--color-trust-blue)] transition-opacity hover:opacity-80"
            onClick={() => router.back()}
          >
            <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 19.5L3 12m0 0l7.5-7.5M3 12h18" />
            </svg>
          </button>
          <h1 className="text-lg font-bold">支付</h1>
          <CountdownTimer expiresAt={expiresAt} />
        </header>

        <div className="space-y-6 px-4 pt-20">
          {/* Package Summary Compact */}
          {selectedPackage && (
            <Card className="shadow-sm">
              <CardContent className="p-5">
                <div className="flex items-start gap-4">
                  <div className="flex-1 space-y-1">
                    <h2 className="text-base font-bold leading-tight">
                      {selectedPackage.package_title}
                    </h2>
                    <p className="text-sm text-muted-foreground">
                      {selectedPackage.destination} - {selectedPackage.duration_days}天{selectedPackage.duration_nights}晚
                    </p>
                    <div className="flex items-end justify-between pt-2">
                      <div className="text-xl font-bold text-[var(--color-vibrant-orange)]">
                        {formatYuan(selectedPackage.total_price_cents)}
                      </div>
                      <span className="rounded-full bg-muted px-2 py-0.5 text-[10px] uppercase tracking-widest text-muted-foreground">
                        已选套餐
                      </span>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Payment Method Toggle */}
          <div className="flex rounded-xl bg-muted p-1">
            <button
              className={`flex-1 rounded-lg py-3 text-sm font-semibold transition-all ${
                method === "qr"
                  ? "bg-white text-[var(--color-trust-blue)] shadow-sm"
                  : "text-muted-foreground hover:text-foreground"
              }`}
              onClick={() => handleMethodChange("qr")}
              disabled={loading}
            >
              扫码支付
            </button>
            <button
              className={`flex-1 rounded-lg py-3 text-sm font-medium transition-all ${
                method === "voice_token"
                  ? "bg-white text-[var(--color-trust-blue)] shadow-sm"
                  : "text-muted-foreground hover:text-foreground"
              }`}
              onClick={() => handleMethodChange("voice_token")}
              disabled={loading}
            >
              吱口令
            </button>
          </div>

          {/* QR / Token Area Full Width */}
          <div className="relative overflow-hidden rounded-2xl bg-muted/50 p-8">
            <div className="absolute -right-10 -top-10 h-32 w-32 rounded-full bg-[var(--color-trust-blue)]/5 blur-3xl" />
            <div className="absolute -bottom-10 -left-10 h-32 w-32 rounded-full bg-[var(--color-vibrant-orange)]/5 blur-3xl" />

            <div className="relative z-10 flex flex-col items-center justify-center space-y-6">
              {loading && <CardSkeleton />}

              {!loading && payment && method === "qr" && (
                <>
                  <div className="rounded-2xl bg-white p-4 shadow-xl">
                    <QRCodeSVG value={payment.qr_code_url} size={192} level="M" />
                  </div>
                  <div className="space-y-2 text-center">
                    <p className="font-medium">使用支付宝或微信扫码</p>
                    <p className="max-w-[200px] text-xs text-muted-foreground">
                      二维码每 30 秒自动刷新，请尽快完成支付
                    </p>
                  </div>
                </>
              )}

              {!loading && payment && method === "voice_token" && (
                <div className="w-full space-y-3">
                  <div className="relative rounded-xl border bg-white p-6 text-center shadow-sm">
                    <p className="mb-2 text-sm text-muted-foreground">吱口令</p>
                    <p className="text-3xl font-bold tracking-widest text-[var(--color-trust-blue)]">
                      {payment.voice_token}
                    </p>
                    <Button
                      variant="outline"
                      size="sm"
                      className="absolute right-2 top-2"
                      onClick={handleCopyToken}
                    >
                      {copied ? "已复制" : "复制"}
                    </Button>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Fee Details */}
          {selectedPackage && (
            <Card className="shadow-sm">
              <CardContent className="space-y-3 p-5">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">退改保障费</span>
                  <span className="font-semibold">{formatYuan(selectedPackage.refund_guarantee_fee_cents)}</span>
                </div>
                <Separator className="bg-border/20" />
                <div className="flex items-center justify-between pt-1">
                  <span className="font-bold">总计支付</span>
                  <span className="text-2xl font-extrabold text-[var(--color-trust-blue)]">
                    {formatYuan(selectedPackage.total_price_cents)}
                  </span>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Trust Indicator */}
          <div className="flex items-center justify-center gap-2 py-4">
            <svg className="h-4 w-4 text-[var(--color-success-green)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z" />
            </svg>
            <span className="text-[10px] uppercase tracking-widest text-muted-foreground">
              安全加密支付
            </span>
          </div>

          {error && (
            <div className="rounded-lg border border-red-200 bg-red-50 p-3 text-center text-sm text-red-700">
              {error}
            </div>
          )}
        </div>

        {/* Sticky Bottom CTA */}
        <div className="fixed inset-x-0 bottom-0 flex items-center justify-between border-t border-border/10 bg-white/80 px-4 py-4 backdrop-blur-md">
          <div className="flex flex-col">
            <span className="text-[10px] uppercase tracking-tighter text-muted-foreground">
              总金额
            </span>
            {selectedPackage && (
              <span className="text-xl font-extrabold">
                {formatYuan(selectedPackage.total_price_cents)}
              </span>
            )}
          </div>
          <Button
            className="bg-gradient-to-br from-[var(--color-trust-blue)] to-[var(--color-trust-blue-dark)] px-8 py-3.5 text-sm font-bold text-white shadow-lg shadow-[var(--color-trust-blue)]/20"
            onClick={handleCheckStatus}
            disabled={checking || loading}
          >
            {checking ? "查询中..." : "查看支付状态"}
          </Button>
        </div>
      </main>
    </>
  )
}

function PaymentSkeleton() {
  return (
    <main className="min-h-screen bg-[#f9f9ff]">
      <div className="mx-auto max-w-2xl space-y-6 px-4 py-8">
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
