import { Card, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"

const STORIES = [
  {
    name: "张先生一家",
    trip: "三亚 5天4晚",
    quote: "通过退改无忧服务，临时改签节省了大量费用，客服响应非常及时。",
    savings: "3200",
  },
  {
    name: "李小姐",
    trip: "大理 自由行",
    quote: "全款预付也不担心，资金托管让人很放心，退改特权非常实在。",
    savings: "1580",
  },
] as const

const TRUST_BADGES = [
  { label: "官方授权服务平台", bg: "bg-green-50/50", border: "border-green-100", color: "text-[var(--color-success-green)]" },
  { label: "资金由支付宝托管", bg: "bg-blue-50/50", border: "border-blue-100", color: "text-[var(--color-trust-blue)]" },
  { label: "100% 退改权益保障", bg: "bg-orange-50/50", border: "border-orange-100", color: "text-[var(--color-vibrant-orange)]" },
  { label: "2000+ 家庭信赖之选", bg: "bg-purple-50/50", border: "border-purple-100", color: "text-purple-600" },
] as const

const BADGE_ICONS: Record<string, React.ReactNode> = {
  "官方授权服务平台": (
    <svg className="mx-auto size-7 sm:size-8" viewBox="0 0 24 24" fill="currentColor">
      <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm-2 16l-4-4 1.41-1.41L10 14.17l6.59-6.59L18 9l-8 8z" />
    </svg>
  ),
  "资金由支付宝托管": (
    <svg className="mx-auto size-7 sm:size-8" viewBox="0 0 24 24" fill="currentColor">
      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" />
    </svg>
  ),
  "100% 退改权益保障": (
    <svg className="mx-auto size-7 sm:size-8" viewBox="0 0 24 24" fill="currentColor">
      <path d="M12.5 6.9c1.78 0 2.44.85 2.5 2.1h2.21c-.07-1.72-1.12-3.3-3.21-3.81V3h-3v2.16c-.53.12-1.03.3-1.48.54l1.47 1.47c.41-.17.91-.27 1.51-.27zM5.33 4.06L4.06 5.33 7.5 8.77c0 2.08 1.56 3.22 3.91 3.91l3.51 3.51c-.34.49-1.05.91-2.42.91-2.06 0-2.87-.92-2.98-2.1h-2.2c.12 2.19 1.76 3.42 3.68 3.83V21h3v-2.15c.96-.18 1.83-.55 2.46-1.12l2.22 2.22 1.27-1.27L5.33 4.06z" />
    </svg>
  ),
  "2000+ 家庭信赖之选": (
    <svg className="mx-auto size-7 sm:size-8" viewBox="0 0 24 24" fill="currentColor">
      <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z" />
    </svg>
  ),
}

export function TrustHero() {
  return (
    <div className="space-y-8 lg:space-y-10">
      {/* Headline */}
      <div className="space-y-4">
        <h1 className="font-display text-3xl font-extrabold leading-tight tracking-tight text-gray-900 md:text-4xl lg:text-5xl">
          告诉我们您的行程需求，
          <br className="hidden sm:block" />
          我们将为您寻找最佳优惠
        </h1>
        <p className="flex items-center gap-2 text-base text-gray-500 lg:text-lg">
          <svg className="size-5 text-[var(--color-success-green)]" viewBox="0 0 24 24" fill="currentColor">
            <path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zM9 8V6c0-1.66 1.34-3 3-3s3 1.34 3 3v2H9z" />
          </svg>
          您的信息已加密，仅用于本次预订
        </p>
      </div>

      {/* Success Stories */}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        {STORIES.map((story) => (
          <Card key={story.name} className="border-gray-100 transition-shadow hover:shadow-md">
            <CardContent className="flex flex-col justify-between p-5 sm:p-6">
              <div className="mb-4 flex items-center gap-3">
                <div className="flex size-10 shrink-0 items-center justify-center rounded-full bg-blue-50 font-display text-sm font-bold text-[var(--color-trust-blue)] sm:size-12">
                  {story.name.charAt(0)}
                </div>
                <div>
                  <p className="font-bold text-gray-900">{story.name}</p>
                  <p className="text-xs text-gray-500">{story.trip}</p>
                </div>
              </div>
              <p className="mb-4 text-sm italic text-gray-600">
                &ldquo;{story.quote}&rdquo;
              </p>
              <Badge variant="secondary" className="w-fit bg-blue-50 text-[var(--color-trust-blue)] hover:bg-blue-50">
                节省了 {story.savings} 元
              </Badge>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Trust Badges */}
      <div className="grid grid-cols-2 gap-3 md:grid-cols-4 md:gap-4">
        {TRUST_BADGES.map((badge) => (
          <div
            key={badge.label}
            className={`flex flex-col items-center gap-2 rounded-xl border p-3 text-center sm:p-4 ${badge.bg} ${badge.border}`}
          >
            <span className={`text-2xl font-bold sm:text-3xl ${badge.color}`}>
              {BADGE_ICONS[badge.label]}
            </span>
            <span className="text-xs font-bold text-gray-800">{badge.label}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
