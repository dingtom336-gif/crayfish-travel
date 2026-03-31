"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { Loader2, Eye, EyeOff } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
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
    <div className="flex flex-1 items-center justify-center bg-zinc-50 px-4 py-6 sm:py-12">
      <Card className="w-full max-w-lg animate-in fade-in slide-in-from-bottom-4 duration-500">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl font-bold" style={{ color: "#1a73e8" }}>
            小龙虾旅行
          </CardTitle>
          <CardDescription>
            描述你的理想旅程，让供应商为你竞价
          </CardDescription>
        </CardHeader>
        <CardContent>
          <fieldset disabled={loading} className="flex flex-col gap-4">
            <form onSubmit={handleSubmit} className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <Label htmlFor="name">姓名</Label>
                <Input
                  id="name"
                  required
                  value={form.name}
                  onChange={(e) => updateField("name", e.target.value)}
                  placeholder="请输入真实姓名"
                />
              </div>

              <div className="flex flex-col gap-1.5">
                <Label htmlFor="id_number">身份证号</Label>
                <div className="relative">
                  <Input
                    id="id_number"
                    required
                    type={showIdNumber ? "text" : "password"}
                    value={form.id_number}
                    onChange={(e) => updateField("id_number", e.target.value)}
                    placeholder="用于出行验证"
                    className="pr-10"
                  />
                  <button
                    type="button"
                    aria-label={showIdNumber ? "隐藏身份证号" : "显示身份证号"}
                    className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground disabled:pointer-events-none disabled:opacity-50"
                    onClick={() => setShowIdNumber((prev) => !prev)}
                  >
                    {showIdNumber ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
                  </button>
                </div>
              </div>

              <div className="flex flex-col gap-1.5">
                <Label htmlFor="phone">手机号</Label>
                <Input
                  id="phone"
                  required
                  type="tel"
                  value={form.phone}
                  onChange={(e) => updateField("phone", e.target.value)}
                  placeholder="接收订单通知"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="adults">成人</Label>
                  <Input
                    id="adults"
                    type="number"
                    min={1}
                    required
                    value={form.adults}
                    onChange={(e) => updateField("adults", Math.max(1, parseInt(e.target.value) || 1))}
                  />
                </div>
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="children">儿童</Label>
                  <Input
                    id="children"
                    type="number"
                    min={0}
                    required
                    value={form.children}
                    onChange={(e) => updateField("children", Math.max(0, parseInt(e.target.value) || 0))}
                  />
                </div>
              </div>

              <div className="flex flex-col gap-1.5">
                <Label htmlFor="description">出行需求</Label>
                <Textarea
                  id="description"
                  required
                  rows={4}
                  value={form.description}
                  onChange={(e) => updateField("description", e.target.value)}
                  placeholder="例如：暑假想带2个大人1个小孩去三亚玩5天，预算8000元左右，希望住海边带泳池的酒店..."
                />
              </div>

              {error && (
                <div className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
                  {error}
                </div>
              )}

              <Button
                type="submit"
                size="lg"
                disabled={loading}
                className="w-full text-white"
                style={{ backgroundColor: "#1a73e8" }}
              >
                {loading ? (
                  <>
                    <Loader2 className="size-4 animate-spin" />
                    正在解析中...
                  </>
                ) : (
                  "开始寻找优惠"
                )}
              </Button>
            </form>
          </fieldset>
        </CardContent>
      </Card>
    </div>
  )
}
