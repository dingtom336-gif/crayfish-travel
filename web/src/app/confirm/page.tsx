"use client"

import { Suspense, useEffect, useState } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import {
  Loader2,
  MapPin,
  Calendar,
  Users,
  Sparkles,
  AlertTriangle,
  ShieldCheck,
  Plane,
  Info,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { StepProgress } from "@/components/StepProgress"
import { api, type TravelRequirement, type DateValidation } from "@/lib/api"
import { getSessionData, setSessionData } from "@/lib/session-store"
import { formatYuan, formatDate } from "@/lib/format"
import { FormSkeleton } from "@/components/Skeleton"

function ConfirmContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const sessionId = searchParams.get("session_id")

  const [requirement, setRequirement] = useState<TravelRequirement | null>(null)
  const [validation, setValidation] = useState<DateValidation | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState("")
  const [progress, setProgress] = useState("")

  // Prefetch the packages page as soon as we have a session
  useEffect(() => {
    if (sessionId) {
      router.prefetch(`/packages?session_id=${sessionId}`)
    }
  }, [sessionId, router])

  useEffect(() => {
    const storedReq = getSessionData<TravelRequirement>("requirement")
    const storedVal = getSessionData<DateValidation>("validation")

    if (storedReq) {
      setRequirement(storedReq)
      if (storedVal) setValidation(storedVal)
      return
    }

    // Fallback: fetch from backend if sessionStorage is empty (e.g. after refresh)
    if (sessionId) {
      api.getSession(sessionId).then((session) => {
        const req: TravelRequirement = {
          destination: session.destination,
          start_date: session.start_date,
          end_date: session.end_date,
          budget_cents: session.budget_cents,
          adults: session.adults,
          children: session.children,
          preferences: session.preferences || [],
        }
        setRequirement(req)
      }).catch(() => {
        router.replace("/")
      })
      return
    }

    router.replace("/")
  }, [sessionId, router])

  async function handleConfirm() {
    if (!sessionId || !requirement) return
    setError("")
    setLoading(true)
    setProgress("")

    try {
      setProgress("正在确认行程信息...")
      await api.confirm({ session_id: sessionId, ...requirement })

      const bidding = await api.startBiddingStream(
        sessionId,
        (_step, message) => setProgress(message),
      )
      setSessionData("packages", bidding.packages)

      setProgress("已获取方案，正在跳转...")
      router.push(`/packages?session_id=${sessionId}`)
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "请求失败，请重试"
      setError(message)
    } finally {
      setLoading(false)
      setProgress("")
    }
  }

  if (!requirement) {
    return (
      <div className="flex flex-1 items-center justify-center bg-gray-50 px-4 py-12">
        <FormSkeleton />
      </div>
    )
  }

  const travelDays = (() => {
    if (!requirement.start_date || !requirement.end_date) return 0
    const start = new Date(requirement.start_date).getTime()
    const end = new Date(requirement.end_date).getTime()
    if (isNaN(start) || isNaN(end)) return 0
    return Math.round((end - start) / 86400000)
  })()

  return (
    <main className="bg-gray-50 py-8 md:py-12 flex-1">
      <div className="max-w-7xl mx-auto px-4 md:px-8">
        {/* Step Progress */}
        <div className="mb-10 flex justify-center">
          <StepProgress
            steps={[
              { label: "需求描述", status: "completed" },
              { label: "行程确认", status: "active" },
              { label: "定制方案", status: "pending" },
            ]}
          />
        </div>

        <div className="max-w-2xl mx-auto">
          <div className="bg-white rounded-2xl shadow-xl border border-gray-100 overflow-hidden">
            {/* Destination hero image */}
            <div className="h-40 bg-gradient-to-br from-blue-50 to-blue-100 flex items-center justify-center">
              <MapPin className="size-10 text-blue-200" />
            </div>

            <div className="p-6 md:p-8 space-y-6">
              {/* Title inside card */}
              <div>
                <h1 className="text-2xl font-extrabold font-display tracking-tight text-gray-900 flex items-center gap-2">
                  <ShieldCheck className="size-6 shrink-0" style={{ color: "var(--color-trust-blue)" }} />
                  确认您的行程信息
                </h1>
                <p className="text-sm text-gray-500 mt-1">
                  AI 已为您智能解析需求，请确认以下信息后开始竞价
                </p>
              </div>

              {/* Peak season alert */}
              {validation?.is_peak_season && (
                <div
                  className="flex items-center gap-2 rounded-lg px-4 py-2 text-xs font-bold text-white"
                  style={{ backgroundColor: "var(--color-vibrant-orange)" }}
                >
                  <AlertTriangle className="size-4 shrink-0" />
                  暑假高峰期 - 价格可能略有上浮
                </div>
              )}

              {/* Destination + Travelers */}
              <div className="grid grid-cols-2 gap-4">
                <div className="bg-gray-50 rounded-xl p-4 space-y-2">
                  <p className="text-xs text-gray-400 font-medium">目的地</p>
                  <Badge
                    className="text-white text-sm font-bold"
                    style={{ backgroundColor: "var(--color-trust-blue)" }}
                  >
                    <MapPin className="size-3 mr-1" />
                    {requirement.destination}
                  </Badge>
                </div>
                <div className="bg-gray-50 rounded-xl p-4 space-y-2">
                  <p className="text-xs text-gray-400 font-medium">出行人数</p>
                  <div className="flex items-center text-sm font-bold text-gray-700">
                    <Users className="size-4 mr-1.5" style={{ color: "var(--color-trust-blue)" }} />
                    {requirement.adults}位成人
                    {requirement.children > 0 && ` + ${requirement.children}位儿童`}
                  </div>
                </div>
              </div>

              {/* Dates + Duration */}
              <div className="grid grid-cols-3 gap-3">
                <div className="bg-gray-50 rounded-xl p-4 space-y-1">
                  <p className="text-xs text-gray-400 font-medium">出发日期</p>
                  <p className="text-base font-bold text-gray-800">{formatDate(requirement.start_date)}</p>
                </div>
                <div className="bg-gray-50 rounded-xl p-4 space-y-1">
                  <p className="text-xs text-gray-400 font-medium">返回日期</p>
                  <p className="text-base font-bold text-gray-800">{formatDate(requirement.end_date)}</p>
                </div>
                <div
                  className="rounded-xl p-4 flex flex-col items-center justify-center"
                  style={{ backgroundColor: "rgba(26, 115, 232, 0.06)" }}
                >
                  <Calendar className="size-5 mb-1" style={{ color: "var(--color-trust-blue)" }} />
                  <span className="text-lg font-extrabold" style={{ color: "var(--color-trust-blue)" }}>
                    {travelDays > 0 ? `${travelDays}天` : "-"}
                  </span>
                </div>
              </div>

              {/* Budget - orange left border */}
              <div
                className="rounded-xl p-4 border-l-4"
                style={{
                  backgroundColor: "rgba(234, 67, 53, 0.04)",
                  borderLeftColor: "var(--color-vibrant-orange)",
                }}
              >
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-xs text-gray-400 font-medium mb-1">预算要求</p>
                    <span
                      className="text-xl font-extrabold tracking-tight"
                      style={{ color: "var(--color-vibrant-orange)" }}
                    >
                      {requirement.budget_cents > 0 ? formatYuan(requirement.budget_cents) : "由供应商竞价"}
                    </span>
                    {requirement.budget_cents === 0 && (
                      <p className="text-xs text-gray-400 mt-1 flex items-center gap-1">
                        <Info className="size-3 shrink-0" />
                        未指定预算，系统将推荐最优方案
                      </p>
                    )}
                  </div>
                  <div
                    className="flex items-center justify-center size-10 rounded-full"
                    style={{ backgroundColor: "rgba(234, 67, 53, 0.1)" }}
                  >
                    <span className="text-lg" style={{ color: "var(--color-vibrant-orange)" }}>$</span>
                  </div>
                </div>
              </div>

              {/* Preferences */}
              {(requirement.preferences ?? []).length > 0 && (
                <div className="space-y-2">
                  <p className="text-xs text-gray-400 font-medium">旅行偏好</p>
                  <div className="flex flex-wrap gap-2">
                    {(requirement.preferences ?? []).map((pref) => (
                      <span
                        key={pref}
                        className="px-3 py-1.5 rounded-full text-xs font-bold border"
                        style={{
                          backgroundColor: "rgba(52, 168, 83, 0.08)",
                          color: "var(--color-success-green)",
                          borderColor: "rgba(52, 168, 83, 0.2)",
                        }}
                      >
                        {pref}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {/* Refund guarantee banner */}
              <div
                className="flex items-center gap-2.5 px-4 py-3 rounded-xl"
                style={{ backgroundColor: "rgba(52, 168, 83, 0.08)" }}
              >
                <ShieldCheck className="size-5 shrink-0" style={{ color: "var(--color-success-green)" }} />
                <span className="text-sm font-bold" style={{ color: "var(--color-success-green)" }}>
                  100% 无忧退款保障
                </span>
              </div>

              {/* Sync message */}
              <p className="text-center text-xs text-gray-400">
                确认后需求将实时同步给 5000+ 认证供应商进行竞价
              </p>

              {error && (
                <div className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
                  {error}
                </div>
              )}

              {/* Action buttons */}
              <div className="grid grid-cols-1 gap-3">
                <Button
                  size="lg"
                  className="w-full h-14 text-white font-extrabold text-lg rounded-xl shadow-lg"
                  style={{ backgroundColor: "var(--color-vibrant-orange)" }}
                  disabled={loading}
                  onClick={handleConfirm}
                >
                  {loading ? (
                    <>
                      <Loader2 className="size-4 animate-spin" />
                      {progress || "正在竞价..."}
                    </>
                  ) : (
                    <>
                      确认并开始竞价
                      <Plane className="size-5 ml-2" />
                    </>
                  )}
                </Button>
                <Button
                  variant="outline"
                  size="lg"
                  className="w-full rounded-xl"
                  onClick={() => router.push("/")}
                >
                  修改行程需求
                </Button>
              </div>
            </div>
          </div>

          {/* Trust footer */}
          <div className="mt-6 flex items-center justify-center gap-6 text-gray-400">
            <div className="flex items-center gap-1">
              <ShieldCheck className="size-4" />
              <span className="text-[10px] uppercase font-bold tracking-widest">安全加密</span>
            </div>
            <div className="flex items-center gap-1">
              <Sparkles className="size-4" />
              <span className="text-[10px] uppercase font-bold tracking-widest">官方认证</span>
            </div>
          </div>
        </div>
      </div>
    </main>
  )
}

export default function ConfirmPage() {
  return (
    <Suspense
      fallback={
        <div className="flex flex-1 items-center justify-center bg-gray-50 px-4 py-12">
          <FormSkeleton />
        </div>
      }
    >
      <ConfirmContent />
    </Suspense>
  )
}
