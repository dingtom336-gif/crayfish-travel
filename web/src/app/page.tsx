"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { Loader2, ArrowRight, Info } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { TrustHero } from "@/components/TrustHero"
import { api } from "@/lib/api"
import { setSessionData, setPersistedSessionId } from "@/lib/session-store"

export default function HomePage() {
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState("")

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
    setLoading(true)

    try {
      const session = await api.createSession({
        adults: form.adults,
        children: form.children,
      })

      setPersistedSessionId(session.session_id)

      const parsed = await api.parse(session.session_id, form.description)

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

                    {/* Dates */}
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label htmlFor="start_date" className="text-sm font-semibold text-gray-700">出发日期</Label>
                        <Input
                          id="start_date"
                          type="date"
                          required
                          value={form.start_date}
                          onChange={(e) => updateField("start_date", e.target.value)}
                          min={new Date().toISOString().split("T")[0]}
                          className="h-12 bg-gray-50 border-transparent rounded-lg focus:bg-white focus:ring-2 focus:ring-blue-600 focus:border-transparent"
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="end_date" className="text-sm font-semibold text-gray-700">返回日期</Label>
                        <Input
                          id="end_date"
                          type="date"
                          required
                          value={form.end_date}
                          onChange={(e) => updateField("end_date", e.target.value)}
                          min={form.start_date || new Date().toISOString().split("T")[0]}
                          className="h-12 bg-gray-50 border-transparent rounded-lg focus:bg-white focus:ring-2 focus:ring-blue-600 focus:border-transparent"
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
                          正在解析中...
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
