"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { Loader2, Eye, EyeOff, ArrowRight, Lock, Info } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { TrustHero } from "@/components/TrustHero"
import { api } from "@/lib/api"
import { setSessionData } from "@/lib/session-store"

export default function HomePage() {
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState("")
  const [showIdNumber, setShowIdNumber] = useState(false)

  const [form, setForm] = useState({
    name: "",
    id_number: "",
    phone: "",
    adults: 1,
    children: 0,
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
      const identity = await api.createIdentity({
        name: form.name,
        id_number: form.id_number,
        phone: form.phone,
        adults: form.adults,
        children: form.children,
      })

      setSessionData("session_id", identity.session_id)

      const parsed = await api.parse(identity.session_id, form.description)

      setSessionData("requirement", parsed.requirement)
      setSessionData("validation", parsed.validation)

      router.push(`/confirm?session_id=${identity.session_id}`)
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
                  <h2 className="text-2xl font-bold font-display">出行人信息</h2>
                </div>

                <fieldset disabled={loading} className="contents">
                  <form onSubmit={handleSubmit} className="space-y-6">
                    {/* Name */}
                    <div className="space-y-2">
                      <Label htmlFor="name" className="text-sm font-semibold text-gray-700">
                        姓名
                      </Label>
                      <Input
                        id="name"
                        required
                        value={form.name}
                        onChange={(e) => updateField("name", e.target.value)}
                        placeholder="与证件姓名一致"
                        className="h-12 bg-gray-50 border-transparent rounded-lg focus:bg-white focus:ring-2 focus:ring-blue-600 focus:border-transparent"
                      />
                    </div>

                    {/* ID Number */}
                    <div className="space-y-2">
                      <Label htmlFor="id_number" className="text-sm font-semibold text-gray-700">
                        身份证号
                      </Label>
                      <div className="relative">
                        <Input
                          id="id_number"
                          required
                          type={showIdNumber ? "text" : "password"}
                          value={form.id_number}
                          onChange={(e) => updateField("id_number", e.target.value)}
                          placeholder="请输入18位身份证号"
                          className="h-12 bg-gray-50 border-transparent rounded-lg focus:bg-white focus:ring-2 focus:ring-blue-600 focus:border-transparent pr-10"
                        />
                        <button
                          type="button"
                          aria-label={showIdNumber ? "隐藏身份证号" : "显示身份证号"}
                          className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 transition-colors"
                          onClick={() => setShowIdNumber((prev) => !prev)}
                        >
                          {showIdNumber ? <EyeOff className="size-5" /> : <Eye className="size-5" />}
                        </button>
                      </div>
                    </div>

                    {/* Phone */}
                    <div className="space-y-2">
                      <Label htmlFor="phone" className="text-sm font-semibold text-gray-700">
                        手机号
                      </Label>
                      <Input
                        id="phone"
                        required
                        type="tel"
                        value={form.phone}
                        onChange={(e) => updateField("phone", e.target.value)}
                        placeholder="接收预订短信"
                        className="h-12 bg-gray-50 border-transparent rounded-lg focus:bg-white focus:ring-2 focus:ring-blue-600 focus:border-transparent"
                      />
                    </div>

                    {/* Travelers */}
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label className="text-sm font-semibold text-gray-700">成人</Label>
                        <div className="flex items-center justify-between bg-gray-50 rounded-lg px-4 py-2 border border-gray-100">
                          <button
                            type="button"
                            className="text-xl font-bold"
                            style={{ color: "var(--color-trust-blue)" }}
                            onClick={() => updateField("adults", Math.max(1, form.adults - 1))}
                          >
                            -
                          </button>
                          <span className="font-bold">{form.adults}</span>
                          <button
                            type="button"
                            className="text-xl font-bold"
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
                            className="text-xl font-bold"
                            style={{ color: "var(--color-trust-blue)" }}
                            onClick={() => updateField("children", Math.max(0, form.children - 1))}
                          >
                            -
                          </button>
                          <span className="font-bold">{form.children}</span>
                          <button
                            type="button"
                            className="text-xl font-bold"
                            style={{ color: "var(--color-trust-blue)" }}
                            onClick={() => updateField("children", form.children + 1)}
                          >
                            +
                          </button>
                        </div>
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
                        信息将在72小时后自动删除，采用AES-256加密技术保障。点击下方按钮即表示您同意我们的
                        <a href="#" className="underline" style={{ color: "var(--color-trust-blue)" }}>退款权益说明</a>
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
