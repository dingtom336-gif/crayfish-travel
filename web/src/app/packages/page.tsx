"use client"

import { useSearchParams, useRouter } from "next/navigation"
import { useState, useCallback, Suspense } from "react"
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { api, type RankedQuote } from "@/lib/api"
import { getSessionData, setSessionData } from "@/lib/session-store"
import { formatYuan, LOCK_DURATION_SECONDS } from "@/lib/format"
import { CountdownTimer } from "@/components/CountdownTimer"
import { PackagesSkeleton, EmptyState } from "@/components/Skeleton"

function StarRating({ rating, reviewCount }: { rating: number; reviewCount: number }) {
  const fullStars = Math.floor(rating)
  const hasHalf = rating - fullStars >= 0.5

  return (
    <div className="flex items-center gap-1">
      <div className="flex">
        {Array.from({ length: 5 }, (_, i) => (
          <span key={i} className={i < fullStars ? "text-amber-400" : i === fullStars && hasHalf ? "text-amber-400" : "text-gray-300"}>
            {i < fullStars ? "\u2605" : i === fullStars && hasHalf ? "\u2605" : "\u2606"}
          </span>
        ))}
      </div>
      <span className="text-xs text-muted-foreground">({reviewCount})</span>
    </div>
  )
}

interface PackageCardProps {
  pkg: RankedQuote
  onSelect: (pkg: RankedQuote) => void
  isLocking: boolean
}

function PackageCard({ pkg, onSelect, isLocking }: PackageCardProps) {
  return (
    <Card className="relative flex flex-col">
      {pkg.is_best_value && (
        <div className="absolute top-3 right-3">
          <Badge className="bg-[#34a853] text-white">超值首选</Badge>
        </div>
      )}
      <CardHeader>
        <CardTitle className="text-lg font-bold">{pkg.package_title}</CardTitle>
        <div className="flex items-center gap-2 pt-1">
          <Badge variant="outline" className="border-[#1a73e8] text-[#1a73e8]">
            {pkg.supplier} 已认证
          </Badge>
        </div>
        <StarRating rating={pkg.star_rating} reviewCount={pkg.review_count} />
      </CardHeader>
      <CardContent className="flex flex-1 flex-col gap-3">
        <ul className="space-y-1 text-sm text-muted-foreground">
          {pkg.highlights.slice(0, 3).map((h, i) => (
            <li key={i} className="flex items-start gap-2">
              <span className="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-[#1a73e8]" />
              {h}
            </li>
          ))}
        </ul>
        <Separator />
        <div className="space-y-1">
          <div className="text-2xl font-bold text-[#1a73e8]">
            {formatYuan(pkg.total_price_cents)}
          </div>
          <div className="text-xs text-muted-foreground">
            基础价格 {formatYuan(pkg.base_price_cents)} + 退改保障费 {formatYuan(pkg.refund_guarantee_fee_cents)}
          </div>
        </div>
      </CardContent>
      <CardFooter>
        <Button
          className="w-full bg-[#ff6d3f] text-white hover:bg-[#e55a30]"
          size="lg"
          onClick={() => onSelect(pkg)}
          disabled={isLocking}
        >
          {isLocking ? "锁定中..." : "选择此方案"}
        </Button>
      </CardFooter>
    </Card>
  )
}

function PackagesContent() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const sessionId = searchParams.get("session_id") ?? ""

  const [packages] = useState<RankedQuote[]>(() => getSessionData<RankedQuote[]>("packages") ?? [])
  const [expiresAt] = useState(() => Date.now() + LOCK_DURATION_SECONDS * 1000)
  const [lockingQuoteId, setLockingQuoteId] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const handleSelect = useCallback(async (pkg: RankedQuote) => {
    if (!sessionId || !pkg.id) return
    setLockingQuoteId(pkg.id)
    setError(null)
    try {
      await api.acquireLock(sessionId, pkg.id)
      setSessionData("selected_package", pkg)
      router.push(`/payment?session_id=${sessionId}&quote_id=${pkg.id}`)
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "锁定失败，请重试"
      setError(message)
      setLockingQuoteId(null)
    }
  }, [sessionId, router])

  if (!sessionId) {
    return (
      <main className="min-h-screen bg-gray-50">
        <div className="mx-auto max-w-6xl px-4 py-16 text-center">
          <h2 className="text-xl font-semibold text-gray-600">未找到会话</h2>
          <p className="mt-2 text-muted-foreground">请从首页重新开始</p>
        </div>
      </main>
    )
  }

  if (packages.length === 0) {
    return (
      <main className="min-h-screen bg-gray-50">
        <EmptyState
          icon="📭"
          title="暂无可用方案"
          description="未找到竞价结果，请返回首页重新提交需求"
        />
      </main>
    )
  }

  return (
    <main className="min-h-screen bg-gray-50">
      <div className="mx-auto max-w-6xl px-4 py-8">
        <div className="mb-6 flex items-center justify-between rounded-lg bg-white p-4 shadow-sm ring-1 ring-foreground/10">
          <div className="flex items-center gap-3">
            <h1 className="text-xl font-bold text-[#1a73e8]">
              已为您锁定 {packages.length} 个方案
            </h1>
          </div>
          {expiresAt > 0 && <CountdownTimer expiresAt={expiresAt} />}
        </div>

        {error && (
          <div className="mb-4 rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700">
            {error}
          </div>
        )}

        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          {packages.map((pkg) => (
              <PackageCard
                key={pkg.id || pkg.rank}
                pkg={pkg}
                onSelect={handleSelect}
                isLocking={lockingQuoteId === pkg.id}
              />
          ))}
        </div>
      </div>
    </main>
  )
}

export default function PackagesPage() {
  return (
    <Suspense
      fallback={
        <main className="min-h-screen bg-gray-50">
          <PackagesSkeleton />
        </main>
      }
    >
      <PackagesContent />
    </Suspense>
  )
}
