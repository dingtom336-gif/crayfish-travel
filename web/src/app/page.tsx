"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { Loader2, ArrowRight, Info, Calendar } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { TrustHero } from "@/components/TrustHero"
import { api } from "@/lib/api"
import { formatDate } from "@/lib/format"
import { setSessionData, setPersistedSessionId } from "@/lib/session-store"

export default function HomePage() {
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState("")
  const [progress, setProgress] = useState("")

  const [form, setForm] = useState({
    adults: 1,
    children: 0,
    start_date: "",
    end_date: "",
    description: "",
  })

  function updateField<K extends keyof typeof form>(key: K, value: (typeof form)[K]) {
    setForm((prev) => ({ ...prev, [key]: value }))
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError("")

    if (!form.start_date || !form.end_date) {
      setError("请选择出发日期和返回日期")
      return
    }
    if (!form.description.trim()) {
      setError("请描述您的出行需求")
      return
    }

    setLoading(true)
    setProgress("")

    try {
      setProgress("正在创建会话...")
      const session = await api.createSession({
        adults: form.adults,
        children: form.children,
      })

      setPersistedSessionId(session.session_id)

      // Prefetch confirm page as soon as session is created
      router.prefetch(`/confirm?session_id=${session.session_id}`)

      const parsed = await api.parseStream(
        session.session_id,
        form.description,
        (_step, message) => setProgress(message),
      )

      setProgress("解析完成，正在跳转...")

      const requirement = {
        ...parsed.requirement,
        start_date: form.start_date,
        end_date: form.end_date,
        adults: form.adults,
        children: form.children,
      }
      setSessionData("requirement", requirement)
      setSessionData("validation", parsed.validation)

      router.push(`/confirm?session_id=${session.session_id}`)
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "请求失败，请重试"
      setError(message)
    } finally {
      setLoading(false)
      setProgress("")
    }
  }

  return (
    <main className="bg-gray-50 py-12 md:py-20">
      <div className="max-w-7xl mx-auto px-4 md:px-8">
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-12 items-start">
          {/* Left: Trust Building */}
          <div className="lg:col-span-7">
            <TrustHero />
          </div>

          {/* Right: Form Card */}
          <div className="lg:col-span-5">
            <div className="bg-white rounded-2xl shadow-xl border border-gray-100 overflow-hidden sticky top-24">
              <div className="p-8">
                <div className="flex items-center justify-between mb-8">
                  <h2 className="text-2xl font-bold font-display">描述您的旅行需求</h2>
                </div>

                <fieldset disabled={loading} className="contents">
                  <form onSubmit={handleSubmit} className="space-y-6">
                    {/* Travelers */}
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label className="text-sm font-semibold text-gray-700">成人</Label>
                        <div className="flex items-center justify-between bg-gray-50 rounded-lg px-4 py-2 border border-gray-100">
                          <button
                            type="button"
                            className="w-10 h-10 flex items-center justify-center text-xl font-bold rounded-lg hover:bg-blue-50 active:bg-blue-100 transition-colors cursor-pointer select-none"
                            style={{ color: "var(--color-trust-blue)" }}
                            onClick={() => updateField("adults", Math.max(1, form.adults - 1))}
                          >
                            -
                          </button>
                          <span className="font-bold">{form.adults}</span>
                          <button
                            type="button"
                            className="w-10 h-10 flex items-center justify-center text-xl font-bold rounded-lg hover:bg-blue-50 active:bg-blue-100 transition-colors cursor-pointer select-none"
                            style={{ color: "var(--color-trust-blue)" }}
                            onClick={() => updateField("adults", form.adults + 1)}
                          >
                            +
                          </button>
                        </div>
                      </div>
                      <div className="space-y-2">
                        <Label className="text-sm font-semibold text-gray-700">儿童</Label>
                        <div className="flex items-center justify-between bg-gray-50 rounded-lg px-4 py-2 border border-gray-100">
                          <button
                            type="button"
                            className="w-10 h-10 flex items-center justify-center text-xl font-bold rounded-lg hover:bg-blue-50 active:bg-blue-100 transition-colors cursor-pointer select-none"
                            style={{ color: "var(--color-trust-blue)" }}
                            onClick={() => updateField("children", Math.max(0, form.children - 1))}
                          >
                            -
                          </button>
                          <span className="font-bold">{form.children}</span>
                          <button
                            type="button"
                            className="w-10 h-10 flex items-center justify-center text-xl font-bold rounded-lg hover:bg-blue-50 active:bg-blue-100 transition-colors cursor-pointer select-none"
                            style={{ color: "var(--color-trust-blue)" }}
                            onClick={() => updateField("children", form.children + 1)}
                          >
                            +
                          </button>
                        </div>
                      </div>
                    </div>

                    {/* Dates - Stitch pill-shaped design */}
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <Label className="block text-xs font-bold text-gray-400 uppercase tracking-wider mb-2 ml-1">出发日期</Label>
                        <button
                          type="button"
                          onClick={() => (document.getElementById("start_date_input") as HTMLInputElement)?.showPicker()}
                          className="w-full flex items-center gap-3 bg-gray-50 hover:bg-gray-100 px-4 py-3.5 rounded-xl border border-gray-200 shadow-sm transition-all cursor-pointer text-left"
                        >
                          <div className="flex items-center justify-center w-9 h-9 rounded-lg" style={{ backgroundColor: "rgba(26, 115, 232, 0.1)" }}>
                            <Calendar className="size-4" style={{ color: "var(--color-trust-blue)" }} />
                          </div>
                          <div className="flex-1 min-w-0">
                            <div className="text-[10px] font-semibold text-gray-400 uppercase tracking-wider">出发</div>
                            <div className={`text-sm font-bold truncate ${form.start_date ? "text-gray-900" : "text-gray-300"}`}>
                              {form.start_date ? formatDate(form.start_date) : "选择日期"}
                            </div>
                          </div>
                        </button>
                        <input
                          id="start_date_input"
                          type="date"
                          tabIndex={-1}
                          value={form.start_date}
                          onChange={(e) => updateField("start_date", e.target.value)}
                          min={new Date().toISOString().split("T")[0]}
                          className="sr-only"
                          aria-hidden="true"
                        />
                      </div>
                      <div>
                        <Label className="block text-xs font-bold text-gray-400 uppercase tracking-wider mb-2 ml-1">返回日期</Label>
                        <button
                          type="button"
                          onClick={() => (document.getElementById("end_date_input") as HTMLInputElement)?.showPicker()}
                          className="w-full flex items-center gap-3 bg-gray-50 hover:bg-gray-100 px-4 py-3.5 rounded-xl border border-gray-200 shadow-sm transition-all cursor-pointer text-left"
                        >
                          <div className="flex items-center justify-center w-9 h-9 rounded-lg" style={{ backgroundColor: "rgba(26, 115, 232, 0.1)" }}>
                            <Calendar className="size-4" style={{ color: "var(--color-trust-blue)" }} />
                          </div>
                          <div className="flex-1 min-w-0">
                            <div className="text-[10px] font-semibold text-gray-400 uppercase tracking-wider">返回</div>
                            <div className={`text-sm font-bold truncate ${form.end_date ? "text-gray-900" : "text-gray-300"}`}>
                              {form.end_date ? formatDate(form.end_date) : "选择日期"}
                            </div>
                          </div>
                        </button>
                        <input
                          id="end_date_input"
                          type="date"
                          tabIndex={-1}
                          value={form.end_date}
                          onChange={(e) => updateField("end_date", e.target.value)}
                          min={form.start_date || new Date().toISOString().split("T")[0]}
                          className="sr-only"
                          aria-hidden="true"
                        />
                      </div>
                    </div>

                    {/* Description */}
                    <div className="space-y-2">
                      <Label htmlFor="description" className="text-sm font-semibold text-gray-700">
                        出行需求
                      </Label>
                      <Textarea
                        id="description"
                        required
                        rows={3}
                        value={form.description}
                        onChange={(e) => updateField("description", e.target.value)}
                        placeholder="例如：暑假想去三亚玩5天，预算8000元左右，希望住海边带泳池的酒店..."
                        className="bg-gray-50 border-transparent rounded-lg focus:bg-white focus:ring-2 focus:ring-blue-600 focus:border-transparent"
                      />
                    </div>

                    {/* Privacy notice */}
                    <div className="flex items-start gap-3 p-4 bg-gray-50 rounded-xl">
                      <Info className="size-4 text-gray-400 mt-0.5 shrink-0" />
                      <p className="text-xs text-gray-500 leading-relaxed">
                        本页面不收集任何个人身份信息。点击下方按钮即表示您同意我们的
                        <a href="#" className="underline" style={{ color: "var(--color-trust-blue)" }}>服务条款</a>
                        及隐私政策。
                      </p>
                    </div>

                    {error && (
                      <div className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
                        {error}
                      </div>
                    )}

                    {/* CTA */}
                    <Button
                      type="submit"
                      size="lg"
                      disabled={loading}
                      className="w-full h-14 text-white font-bold text-base rounded-xl shadow-lg"
                      style={{ backgroundColor: "var(--color-trust-blue)" }}
                    >
                      {loading ? (
                        <>
                          <Loader2 className="size-4 animate-spin" />
                          {progress || "正在解析中..."}
                        </>
                      ) : (
                        <>
                          开始寻找优惠
                          <ArrowRight className="size-5 ml-2" />
                        </>
                      )}
                    </Button>
                  </form>
                </fieldset>
              </div>

              {/* Bottom price preview */}
              <div className="bg-blue-50 p-6 flex justify-between items-center border-t border-blue-100">
                <div>
                  <p className="text-xs font-medium" style={{ color: "var(--color-trust-blue)" }}>
                    预计总额
                  </p>
                  <p className="text-2xl font-black" style={{ color: "var(--color-trust-blue)" }}>
                    &yen; --.--
                  </p>
                </div>
                <div className="text-right">
                  <p className="text-[10px] uppercase font-bold tracking-widest" style={{ color: "var(--color-trust-blue)", opacity: 0.6 }}>
                    包含
                  </p>
                  <p className="text-xs" style={{ color: "var(--color-trust-blue)" }}>
                    基础行程费 + 退改权益费
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  )
}
