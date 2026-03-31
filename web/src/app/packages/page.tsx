"use client"

import { useSearchParams, useRouter } from "next/navigation"
import { useState, useCallback, Suspense } from "react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { api, type RankedQuote, type TravelRequirement } from "@/lib/api"
import { getSessionData, setSessionData } from "@/lib/session-store"
import { formatYuan, LOCK_DURATION_SECONDS } from "@/lib/format"
import { CountdownTimer } from "@/components/CountdownTimer"
import { PackagesSkeleton, EmptyState } from "@/components/Skeleton"
import { TripSidebar } from "@/components/TripSidebar"
import { TrustGuarantee } from "@/components/TrustGuarantee"

const HIGHLIGHT_COLORS = [
  "bg-blue-50 text-[var(--color-trust-blue)]",
  "bg-orange-50 text-[var(--color-vibrant-orange)]",
  "bg-green-50 text-[var(--color-success-green)]",
  "bg-purple-50 text-purple-600",
  "bg-pink-50 text-pink-600",
]

interface PackageCardProps {
  pkg: RankedQuote
  onSelect: (pkg: RankedQuote) => void
  isLocking: boolean
}

function PackageCard({ pkg, onSelect, isLocking }: PackageCardProps) {
  return (
    <div className="group relative overflow-hidden rounded-xl border border-gray-100 border-l-4 border-l-[var(--color-success-green)] bg-white shadow-sm transition-all hover:shadow-xl">
      {pkg.is_best_value && (
        <div className="absolute left-0 top-4 z-20 rounded-r-full bg-[var(--color-success-green)] px-4 py-1 text-sm font-bold text-white shadow-md">
          超值首选
        </div>
      )}

      {pkg.image_url && (
        <div className="relative h-56 overflow-hidden">
          <img
            src={pkg.image_url}
            alt={pkg.package_title}
            className="h-full w-full object-cover transition-transform duration-700 group-hover:scale-110"
            onError={(e) => { (e.target as HTMLImageElement).style.display = "none" }}
          />
          <div className="absolute right-4 top-4 flex items-center gap-1 rounded bg-white/90 px-2 py-1 text-xs font-bold text-[var(--color-success-green)] shadow-sm backdrop-blur">
            <svg className="h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
              <path fillRule="evenodd" d="M8.603 3.799A4.49 4.49 0 0112 2.25c1.357 0 2.573.6 3.397 1.549a4.49 4.49 0 013.498 1.307 4.491 4.491 0 011.307 3.497A4.49 4.49 0 0121.75 12a4.49 4.49 0 01-1.549 3.397 4.491 4.491 0 01-1.307 3.497 4.491 4.491 0 01-3.497 1.307A4.49 4.49 0 0112 21.75a4.49 4.49 0 01-3.397-1.549 4.49 4.49 0 01-3.498-1.306 4.491 4.491 0 01-1.307-3.498A4.49 4.49 0 012.25 12c0-1.357.6-2.573 1.549-3.397a4.49 4.49 0 011.307-3.497 4.49 4.49 0 013.497-1.307zm7.007 6.387a.75.75 0 10-1.22-.872l-3.236 4.53L9.53 12.22a.75.75 0 00-1.06 1.06l2.25 2.25a.75.75 0 001.14-.094l3.75-5.25z" clipRule="evenodd" />
            </svg>
            <span>飞猪验证</span>
          </div>
        </div>
      )}

      <div className="p-6">
        <div className="mb-3 flex items-start justify-between">
          <h3 className="text-xl font-bold font-display">{pkg.package_title}</h3>
          <div className="flex items-center gap-1 rounded bg-gray-50 px-2 py-1">
            <svg className="h-4 w-4 text-[var(--color-vibrant-orange)]" viewBox="0 0 24 24" fill="currentColor">
              <path fillRule="evenodd" d="M10.788 3.21c.448-1.077 1.976-1.077 2.424 0l2.082 5.007 5.404.433c1.164.093 1.636 1.545.749 2.305l-4.117 3.527 1.257 5.273c.271 1.136-.964 2.033-1.96 1.425L12 18.354 7.373 21.18c-.996.608-2.231-.29-1.96-1.425l1.257-5.273-4.117-3.527c-.887-.76-.415-2.212.749-2.305l5.404-.433 2.082-5.006z" clipRule="evenodd" />
            </svg>
            <span className="text-xs font-bold">{pkg.star_rating}</span>
            {pkg.review_count > 0 && (
              <span className="text-[10px] text-gray-400">({pkg.review_count}+ 评价)</span>
            )}
          </div>
        </div>

        <div className="mb-6 flex flex-wrap gap-2">
          {pkg.highlights.slice(0, 3).map((h, i) => (
            <span
              key={h || i}
              className={`rounded px-2 py-1 text-[11px] font-bold uppercase tracking-wider ${HIGHLIGHT_COLORS[i % HIGHLIGHT_COLORS.length]}`}
            >
              {h}
            </span>
          ))}
        </div>

        <div className="border-t border-gray-100 pt-4">
          <div className="mb-4 flex items-end justify-between">
            <div>
              <p className="text-2xl font-black font-display text-[var(--color-trust-blue)]">
                {formatYuan(pkg.total_price_cents)}
              </p>
              <p className="mt-1 text-[10px] text-gray-400">
                基础行程 {formatYuan(pkg.base_price_cents)} + 无忧退改服务费 {formatYuan(pkg.refund_guarantee_fee_cents)}
              </p>
            </div>
            <Button
              className="rounded-full bg-gradient-to-r from-[var(--color-vibrant-orange)] to-[#ff8c66] px-6 py-3 font-bold text-white shadow-lg shadow-orange-200 hover:brightness-110"
              onClick={() => onSelect(pkg)}
              disabled={isLocking}
            >
              {isLocking ? "锁定中..." : "选择此方案"}
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}

function PackagesContent() {
  const searchParams = useSearchParams()
  const router = useRouter()
  const sessionId = searchParams.get("session_id") ?? ""

  const [packages] = useState<RankedQuote[]>(() => getSessionData<RankedQuote[]>("packages") ?? [])
  const [requirement] = useState<TravelRequirement | null>(() => getSessionData<TravelRequirement>("requirement"))
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
    <main className="min-h-screen bg-gray-50 pb-28">
      {/* Lock Banner - orange gradient */}
      <section className="relative w-full overflow-hidden bg-gradient-to-r from-[var(--color-vibrant-orange)] to-[#ff8c66] py-4 text-white shadow-lg">
        <div className="relative z-10 mx-auto flex max-w-7xl flex-col items-center justify-between gap-2 px-6 md:flex-row">
          <div className="flex items-center gap-3">
            <svg className="h-8 w-8" viewBox="0 0 24 24" fill="currentColor">
              <path fillRule="evenodd" d="M12 1.5a5.25 5.25 0 00-5.25 5.25v3a3 3 0 00-3 3v6.75a3 3 0 003 3h10.5a3 3 0 003-3v-6.75a3 3 0 00-3-3v-3c0-2.9-2.35-5.25-5.25-5.25zm3.75 8.25v-3a3.75 3.75 0 10-7.5 0v3h7.5z" clipRule="evenodd" />
            </svg>
            <div>
              <h2 className="text-xl font-bold tracking-tight font-display">
                {packages.length}个方案已为您锁定
              </h2>
              <p className="text-sm opacity-90">倒计时结束前方案将为您保留</p>
            </div>
          </div>
          <div className="flex items-center gap-4 rounded-xl border border-white/30 bg-white/20 px-6 py-2 backdrop-blur-md">
            <span className="text-sm font-medium">剩余时间</span>
            {expiresAt > 0 && <CountdownTimer expiresAt={expiresAt} />}
          </div>
        </div>
        {/* Decorative wave */}
        <div className="pointer-events-none absolute right-0 top-0 opacity-10">
          <svg width="400" height="100" viewBox="0 0 400 100" fill="none">
            <path d="M0 50C50 20 150 80 200 50C250 20 350 80 400 50" stroke="white" strokeWidth="2" />
            <path d="M0 70C50 40 150 100 200 70C250 40 350 100 400 70" stroke="white" strokeWidth="2" />
          </svg>
        </div>
      </section>

      <div className="mx-auto max-w-7xl px-6 py-8">
        {error && (
          <div className="mb-4 rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700">
            {error}
          </div>
        )}

        {/* Mobile: trip summary pill */}
        {requirement && (
          <div className="mb-6 lg:hidden">
            <div className="flex items-center justify-between gap-3 overflow-x-auto whitespace-nowrap rounded-full border border-gray-100 bg-white px-5 py-2.5 shadow-sm scrollbar-hide">
              <div className="flex items-center gap-3 text-xs font-medium text-gray-600">
                <span className="flex items-center gap-1">
                  <svg className="h-4 w-4 text-[var(--color-trust-blue)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M15 10.5a3 3 0 11-6 0 3 3 0 016 0z" />
                    <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1115 0z" />
                  </svg>
                  {requirement.destination}
                </span>
                <span className="h-3 w-px bg-gray-200" />
                <span>{formatYuan(requirement.budget_cents)}预算</span>
                <span className="h-3 w-px bg-gray-200" />
                <span>{requirement.adults + requirement.children}位</span>
              </div>
            </div>
          </div>
        )}

        <div className="flex flex-col gap-8 lg:flex-row">
          {/* Package Cards - left 2/3 */}
          <div className="flex-1">
            <div className="grid grid-cols-1 items-start gap-6 md:grid-cols-2">
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

          {/* Sidebar - right 1/3 (desktop only) */}
          {requirement && (
            <aside className="hidden w-80 shrink-0 lg:block">
              <TripSidebar
                destination={requirement.destination}
                startDate={requirement.start_date}
                endDate={requirement.end_date}
                adults={requirement.adults}
                children={requirement.children}
                budgetCents={requirement.budget_cents}
              />
              <TrustGuarantee />
            </aside>
          )}
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
