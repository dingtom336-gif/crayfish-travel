"use client"

import { Suspense, useEffect, useState } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { Loader2, MapPin, Calendar, Wallet, Users, Sparkles, AlertTriangle } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
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

    if (storedReq) setRequirement(storedReq)
    if (storedVal) setValidation(storedVal)

    if (!storedReq && !sessionId) {
      router.replace("/")
    }
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
      <div className="flex flex-1 items-center justify-center bg-zinc-50 px-4 py-12">
        <FormSkeleton />
      </div>
    )
  }

  const totalTravelers = requirement.adults + requirement.children

  return (
    <div className="flex flex-1 items-center justify-center bg-zinc-50 px-4 py-6 sm:py-12">
      <Card className="w-full max-w-lg">
        <CardHeader>
          <CardTitle className="text-xl font-bold">确认行程信息</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-5">
          {validation?.is_peak_season && (
            <div
              className="flex items-center gap-2 rounded-lg px-4 py-3 text-sm font-medium text-white"
              style={{ backgroundColor: "#ff6d3f" }}
            >
              <AlertTriangle className="size-4 shrink-0" />
              <span>
                旺季提醒：{validation.peak_type === "summer" ? "暑期高峰" : "春节高峰"}，价格可能略有上浮
              </span>
            </div>
          )}

          <div className="flex flex-col gap-4">
            <div className="flex items-center gap-3">
              <MapPin className="size-5 shrink-0" style={{ color: "#1a73e8" }} />
              <div>
                <div className="text-xs text-muted-foreground">目的地</div>
                <Badge
                  className="mt-0.5 text-white"
                  style={{ backgroundColor: "#1a73e8" }}
                >
                  {requirement.destination}
                </Badge>
              </div>
            </div>

            <Separator />

            <div className="flex items-center gap-3">
              <Calendar className="size-5 shrink-0" style={{ color: "#1a73e8" }} />
              <div>
                <div className="text-xs text-muted-foreground">出行日期</div>
                <div className="text-sm font-medium">
                  {formatDate(requirement.start_date)} — {formatDate(requirement.end_date)}
                </div>
              </div>
            </div>

            <Separator />

            <div className="flex items-center gap-3">
              <Wallet className="size-5 shrink-0" style={{ color: "#ff6d3f" }} />
              <div>
                <div className="text-xs text-muted-foreground">预算</div>
                <div className="text-lg font-bold" style={{ color: "#ff6d3f" }}>
                  {formatYuan(requirement.budget_cents)}
                </div>
              </div>
            </div>

            <Separator />

            <div className="flex items-center gap-3">
              <Users className="size-5 shrink-0" style={{ color: "#1a73e8" }} />
              <div>
                <div className="text-xs text-muted-foreground">出行人数</div>
                <div className="text-sm font-medium">
                  {requirement.adults}位成人
                  {requirement.children > 0 && (
                    <> + {requirement.children}位儿童</>
                  )}
                  （共{totalTravelers}人）
                </div>
              </div>
            </div>

            {requirement.preferences.length > 0 && (
              <>
                <Separator />
                <div className="flex items-start gap-3">
                  <Sparkles className="mt-0.5 size-5 shrink-0" style={{ color: "#34a853" }} />
                  <div>
                    <div className="text-xs text-muted-foreground">偏好</div>
                    <div className="mt-1 flex flex-wrap gap-1.5">
                      {requirement.preferences.map((pref) => (
                        <Badge
                          key={pref}
                          variant="secondary"
                          className="text-white"
                          style={{ backgroundColor: "#34a853" }}
                        >
                          {pref}
                        </Badge>
                      ))}
                    </div>
                  </div>
                </div>
              </>
            )}
          </div>

          {error && (
            <div className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {error}
            </div>
          )}

          <div className="flex gap-3 pt-2">
            <Button
              variant="outline"
              size="lg"
              className="flex-1"
              onClick={() => router.push("/")}
            >
              修改
            </Button>
            <Button
              size="lg"
              className="flex-1 text-white"
              style={{ backgroundColor: "#ff6d3f" }}
              disabled={loading}
              onClick={handleConfirm}
            >
              {loading ? (
                <>
                  <Loader2 className="size-4 animate-spin" />
                  正在竞价...
                </>
              ) : (
                "确认并开始竞价"
              )}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export default function ConfirmPage() {
  return (
    <Suspense
      fallback={
        <div className="flex flex-1 items-center justify-center bg-zinc-50 px-4 py-12">
          <FormSkeleton />
        </div>
      }
    >
      <ConfirmContent />
    </Suspense>
  )
}
