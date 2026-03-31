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

    try {
      await api.confirm({ session_id: sessionId, ...requirement })

      const bidding = await api.startBidding(sessionId)
      setSessionData("packages", bidding.packages)

      router.push(`/packages?session_id=${sessionId}`)
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "请求失败，请重试"
      setError(message)
    } finally {
      setLoading(false)
    }
  }

  if (!requirement) {
    return (
      <div className="flex flex-1 items-center justify-center bg-gray-50 px-4 py-12">
        <FormSkeleton />
      </div>
    )
  }

  return (
    <main className="bg-gray-50 py-8 md:py-12 flex-1">
      <div className="max-w-7xl mx-auto px-4 md:px-8">
        {/* Step Progress */}
        <div className="mb-12 flex justify-center">
          <StepProgress
            steps={[
              { label: "需求描述", status: "completed" },
              { label: "行程确认", status: "active" },
              { label: "定制方案", status: "pending" },
            ]}
          />
        </div>

        {/* Page heading */}
        <div className="mb-8">
          <h1 className="text-3xl font-extrabold font-display tracking-tight text-gray-900 mb-2">
            告诉我们您的行程需求
          </h1>
          <p className="text-gray-500">
            我们的 AI 助手将为您实时解析需求并匹配最优资源。
          </p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-10 gap-12 items-start">
          {/* Left: Chat-like UI */}
          <div className="lg:col-span-6 space-y-8">
            {/* Chat area */}
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
              <div className="bg-gray-50/50 px-6 py-4 border-b border-gray-100">
                <h2 className="text-base font-bold text-gray-800 flex items-center gap-2">
                  <span
                    className="inline-flex items-center justify-center w-6 h-6 rounded-full"
                    style={{ backgroundColor: "var(--color-trust-blue)", color: "white", fontSize: 12 }}
                  >
                    AI
                  </span>
                  告诉我您的理想行程
                </h2>
              </div>

              <div className="p-6 space-y-4 min-h-[200px]">
                {/* User message */}
                <div className="flex justify-end">
                  <div
                    className="px-4 py-2.5 rounded-2xl rounded-tr-none max-w-[85%] text-sm text-white shadow-sm"
                    style={{ backgroundColor: "var(--color-trust-blue)" }}
                  >
                    {requirement.destination}，{requirement.adults}位成人
                    {requirement.children > 0 && `${requirement.children}位儿童`}
                    ，预算{formatYuan(requirement.budget_cents)}
                  </div>
                </div>

                {/* AI response */}
                <div className="flex justify-start">
                  <div className="bg-gray-100 text-gray-700 px-4 py-2.5 rounded-2xl rounded-tl-none max-w-[85%] text-sm border border-gray-200">
                    好的，我已为您解析完行程需求，请查看右侧确认信息。
                  </div>
                </div>
              </div>

              {/* AI parsed result */}
              <div className="mx-6 mb-6 bg-blue-50/50 rounded-xl p-6 border border-blue-100 border-dashed">
                <div className="flex items-center space-x-3 mb-4">
                  <div className="flex space-x-1">
                    <div className="w-2 h-2 rounded-full" style={{ backgroundColor: "var(--color-success-green)" }} />
                    <div className="w-2 h-2 rounded-full" style={{ backgroundColor: "var(--color-success-green)" }} />
                    <div className="w-2 h-2 rounded-full" style={{ backgroundColor: "var(--color-success-green)" }} />
                  </div>
                  <span className="font-medium" style={{ color: "var(--color-success-green)" }}>
                    AI 已完成需求解析
                  </span>
                </div>
                <div className="space-y-2 text-sm text-gray-600">
                  <p><span className="font-semibold text-gray-800">目的地：</span>{requirement.destination}</p>
                  <p><span className="font-semibold text-gray-800">出行日期：</span>{formatDate(requirement.start_date)} - {formatDate(requirement.end_date)}</p>
                  <p><span className="font-semibold text-gray-800">出行人数：</span>{requirement.adults}位成人{requirement.children > 0 && ` + ${requirement.children}位儿童`}</p>
                  <p><span className="font-semibold text-gray-800">预算：</span>{formatYuan(requirement.budget_cents)}</p>
                  {requirement.preferences.length > 0 && (
                    <p><span className="font-semibold text-gray-800">偏好：</span>{requirement.preferences.join("、")}</p>
                  )}
                </div>
              </div>
            </div>
          </div>

          {/* Right: Confirmation card */}
          <div className="lg:col-span-4 sticky top-24">
            <div className="bg-white rounded-xl shadow-xl border border-gray-100 overflow-hidden">
              <div className="bg-gray-50 px-6 py-4 border-b border-gray-100">
                <h2 className="font-bold text-lg text-gray-800 flex items-center gap-2">
                  <ShieldCheck className="size-5" style={{ color: "var(--color-trust-blue)" }} />
                  已确定的行程详情
                </h2>
              </div>

              <div className="p-6 space-y-6">
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

                {/* Details */}
                <div className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-1">
                      <p className="text-xs text-gray-400 font-medium uppercase tracking-wider">目的地</p>
                      <Badge
                        className="text-white text-sm font-bold"
                        style={{ backgroundColor: "var(--color-trust-blue)" }}
                      >
                        <MapPin className="size-3 mr-1" />
                        {requirement.destination}
                      </Badge>
                    </div>
                    <div className="space-y-1">
                      <p className="text-xs text-gray-400 font-medium uppercase tracking-wider">出行人数</p>
                      <div className="flex items-center text-sm font-bold text-gray-700">
                        <Users className="size-4 mr-1" />
                        {requirement.adults}位成人
                        {requirement.children > 0 && ` + ${requirement.children}位儿童`}
                      </div>
                    </div>
                  </div>
                  <div className="grid grid-cols-3 gap-4">
                    <div className="space-y-1">
                      <p className="text-xs text-gray-400 font-medium uppercase tracking-wider">出发</p>
                      <div className="flex items-center text-sm font-bold text-gray-700">
                        <Calendar className="size-4 mr-1 shrink-0" style={{ color: "var(--color-trust-blue)" }} />
                        {formatDate(requirement.start_date)}
                      </div>
                    </div>
                    <div className="space-y-1">
                      <p className="text-xs text-gray-400 font-medium uppercase tracking-wider">返回</p>
                      <div className="flex items-center text-sm font-bold text-gray-700">
                        <Calendar className="size-4 mr-1 shrink-0" style={{ color: "var(--color-trust-blue)" }} />
                        {formatDate(requirement.end_date)}
                      </div>
                    </div>
                    <div className="space-y-1">
                      <p className="text-xs text-gray-400 font-medium uppercase tracking-wider">时长</p>
                      <p className="text-sm font-bold text-gray-700">
                        {Math.round((new Date(requirement.end_date).getTime() - new Date(requirement.start_date).getTime()) / 86400000)}天
                      </p>
                    </div>
                  </div>
                </div>

                {/* Preferences */}
                {requirement.preferences.length > 0 && (
                  <div className="space-y-2">
                    <p className="text-xs text-gray-400 font-medium uppercase tracking-wider">偏好设置</p>
                    <div className="flex flex-wrap gap-2">
                      {requirement.preferences.map((pref) => (
                        <span
                          key={pref}
                          className="px-3 py-1 rounded-lg text-xs font-bold border"
                          style={{
                            backgroundColor: "rgba(52, 168, 83, 0.1)",
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

                {/* Budget */}
                <div className="pt-4 border-t border-gray-50">
                  <p className="text-xs text-gray-400 font-medium uppercase tracking-wider mb-1">预估预算</p>
                  <div className="flex items-baseline gap-2">
                    <span
                      className="text-2xl font-extrabold tracking-tight"
                      style={{ color: "var(--color-vibrant-orange)" }}
                    >
                      {formatYuan(requirement.budget_cents)}
                    </span>
                  </div>
                </div>

                {/* Refund guarantee */}
                <div
                  className="flex items-center gap-2 p-3 rounded-lg border"
                  style={{
                    backgroundColor: "rgba(52, 168, 83, 0.05)",
                    borderColor: "rgba(52, 168, 83, 0.2)",
                  }}
                >
                  <ShieldCheck className="size-5 shrink-0" style={{ color: "var(--color-success-green)" }} />
                  <p className="text-[11px] font-medium leading-tight" style={{ color: "var(--color-success-green)" }}>
                    已激活：退改权益服务。具体退款规则请参见<a href="/refund-policy" className="underline">退款权益说明</a>。
                  </p>
                </div>

                {/* Compliance note */}
                <div className="flex items-start gap-2 pt-2">
                  <Info className="size-4 text-gray-300 mt-0.5 shrink-0" />
                  <p className="text-[10px] text-gray-400 leading-relaxed">
                    该预算已包含基础行程费用及退改权益服务费。高峰期价格可能略有上浮，具体以最终供应商竞价为准。
                  </p>
                </div>

                {error && (
                  <div className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
                    {error}
                  </div>
                )}

                {/* Action buttons */}
                <div className="grid grid-cols-1 gap-3 pt-2">
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
                        正在竞价...
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
                    className="w-full"
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
