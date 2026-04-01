import Link from "next/link"
import { CheckCircle, Brain, Search, BarChart3, User, ArrowRight, Shield, Lock, Trash2, Calendar, MessageSquare, ShoppingBag, QrCode, ChevronRight } from "lucide-react"

const steps = [
  {
    num: 1,
    title: "描述需求",
    tag: "自然语言",
    tagColor: "bg-blue-50 text-blue-600",
    icon: <Calendar className="size-6 text-gray-400" />,
    wireframe: (
      <div className="space-y-3">
        <div className="h-3 w-3/4 bg-gray-200 rounded-full" />
        <div className="h-8 w-full border border-gray-200 rounded-lg flex items-center px-3">
          <Calendar className="size-3 text-gray-300" />
          <div className="ml-2 h-2 w-16 bg-gray-200 rounded" />
        </div>
        <div className="h-12 w-full border border-gray-200 rounded-lg" />
      </div>
    ),
  },
  {
    num: 2,
    title: "AI 解析",
    tag: "结构化需求",
    tagColor: "bg-blue-50 text-blue-600",
    icon: <MessageSquare className="size-6 text-gray-400" />,
    wireframe: (
      <div className="space-y-3">
        <div className="flex justify-end">
          <div className="h-6 w-24 bg-blue-100 rounded-xl rounded-tr-none" />
        </div>
        <div className="flex items-start gap-2">
          <div className="w-5 h-5 rounded-full bg-blue-100 shrink-0" />
          <div className="h-10 flex-1 bg-gray-100 rounded-xl rounded-tl-none" />
        </div>
        <div className="h-6 w-20 bg-green-50 rounded-lg border border-green-200 mx-auto" />
      </div>
    ),
  },
  {
    num: 3,
    title: "供应商竞价",
    tag: "最优方案",
    tagColor: "bg-orange-50 text-orange-600",
    icon: <ShoppingBag className="size-6 text-gray-400" />,
    wireframe: (
      <div className="space-y-2">
        {[1, 2, 3].map((i) => (
          <div key={i} className={`h-7 w-full rounded-lg flex items-center justify-between px-3 ${i === 3 ? "border-2 border-blue-200 bg-blue-50/50" : "bg-gray-100"}`}>
            <div className="h-2 w-12 bg-gray-200 rounded" />
            <div className="h-3 w-8 bg-orange-200 rounded text-[6px] text-center font-bold text-orange-600">{(3200 + i * 400)}</div>
          </div>
        ))}
      </div>
    ),
  },
  {
    num: 4,
    title: "锁定支付",
    tag: "订单确认",
    tagColor: "bg-blue-50 text-blue-600",
    icon: <QrCode className="size-6 text-gray-400" />,
    wireframe: (
      <div className="flex flex-col items-center justify-center gap-2">
        <div className="w-14 h-14 border-2 border-gray-200 rounded-lg grid grid-cols-3 gap-0.5 p-1.5">
          {Array.from({ length: 9 }).map((_, i) => (
            <div key={i} className={`rounded-sm ${i % 3 === 0 ? "bg-gray-800" : "bg-gray-300"}`} />
          ))}
        </div>
        <div className="h-2 w-16 bg-gray-200 rounded" />
      </div>
    ),
  },
  {
    num: 5,
    title: "出行无忧",
    tag: "服务启动",
    tagColor: "bg-green-50 text-green-600",
    icon: <CheckCircle className="size-6 text-gray-400" />,
    wireframe: (
      <div className="flex flex-col items-center justify-center">
        <div className="w-14 h-14 rounded-full bg-green-50 flex items-center justify-center">
          <CheckCircle className="size-8 text-green-500" />
        </div>
      </div>
    ),
  },
]

const archNodes = [
  { icon: <User className="size-7 text-blue-600" />, label: "用户", sub: "自然语言输入", bg: "bg-white", border: "border-gray-200" },
  null,
  { icon: <Brain className="size-7 text-white" />, label: "MiniMax M2.5", sub: "大模型解析引擎", bg: "bg-blue-600", border: "border-blue-700", text: "text-white" },
  null,
  { icon: <Search className="size-7 text-white" />, label: "FlyAI 匹配", sub: "飞猪真实数据", bg: "bg-orange-500", border: "border-orange-600", text: "text-white" },
  null,
  { icon: <BarChart3 className="size-7 text-white" />, label: "智能排序", sub: "多因子加权", bg: "bg-green-600", border: "border-green-700", text: "text-white" },
  null,
  { icon: <CheckCircle className="size-7 text-green-600" />, label: "用户选择", sub: "锁定最优方案", bg: "bg-white", border: "border-gray-200" },
]

const dataLabels = ["文本描述", "目的地/预算/偏好", "5个实时方案", "最优推荐"]

export default function AboutPage() {
  return (
    <main className="bg-white">
      {/* Hero */}
      <section className="relative overflow-hidden bg-gradient-to-br from-blue-50 via-white to-orange-50 py-24 md:py-32">
        <div className="absolute inset-0 opacity-30" style={{ backgroundImage: "radial-gradient(circle, #e0e2ec 1px, transparent 1px)", backgroundSize: "32px 32px" }} />
        <div className="relative max-w-5xl mx-auto px-4 text-center">
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-blue-50 border border-blue-100 text-blue-700 text-sm font-semibold mb-8">
            <Brain className="size-4" /> AI 驱动的反向撮合旅行平台
          </div>
          <h1 className="font-display text-4xl md:text-6xl font-extrabold tracking-tight text-gray-900 mb-6 leading-tight">
            告诉我们你想去哪，<br />
            <span style={{ color: "var(--color-vibrant-orange)" }}>供应商为你竞价</span>
          </h1>
          <p className="max-w-2xl mx-auto text-lg text-gray-500 mb-10 leading-relaxed">
            用自然语言描述旅行需求，MiniMax M2.5 大模型实时解析，FlyAI 对接真实供应商竞价，1.3 秒锁定最优方案
          </p>
          <Link
            href="/crayfish"
            className="inline-flex items-center gap-2 px-8 py-4 rounded-xl text-white font-bold text-lg shadow-lg hover:shadow-xl transition-all hover:-translate-y-0.5"
            style={{ backgroundColor: "var(--color-vibrant-orange)" }}
          >
            立即体验 <ArrowRight className="size-5" />
          </Link>
        </div>
      </section>

      {/* 5-Step Flow */}
      <section className="py-20 md:py-28 bg-gray-50/50">
        <div className="max-w-6xl mx-auto px-4">
          <div className="text-center mb-16">
            <h2 className="font-display text-3xl md:text-4xl font-bold text-gray-900 mb-3">智能解析与实时竞价流程</h2>
            <div className="h-1 w-16 rounded-full mx-auto" style={{ backgroundColor: "var(--color-trust-blue)" }} />
          </div>

          {/* Desktop flow */}
          <div className="hidden lg:block">
            {/* Connection line */}
            <div className="relative">
              <div className="absolute top-[72px] left-[10%] right-[10%] h-0.5 bg-gradient-to-r from-blue-200 via-orange-200 to-green-200" />
            </div>
            <div className="grid grid-cols-5 gap-6">
              {steps.map((step, i) => (
                <div key={step.num} className="relative flex flex-col items-center text-center">
                  {/* Number */}
                  <div className="w-12 h-12 rounded-full flex items-center justify-center text-white font-bold text-lg mb-6 z-10 shadow-md" style={{ backgroundColor: "var(--color-trust-blue)" }}>
                    {step.num}
                  </div>
                  {/* Wireframe card */}
                  <div className="bg-white p-5 rounded-xl shadow-sm border border-gray-100 w-full aspect-[4/3] flex flex-col justify-center mb-4 hover:shadow-md transition-shadow">
                    {step.wireframe}
                  </div>
                  {/* Label */}
                  <h3 className="font-semibold text-base text-gray-800 mb-2">{step.title}</h3>
                  <span className={`text-xs font-semibold px-3 py-1 rounded-full ${step.tagColor}`}>{step.tag}</span>
                  {/* Arrow */}
                  {i < 4 && (
                    <div className="absolute -right-3 top-[72px] z-20">
                      <ChevronRight className="size-5 text-gray-300" />
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>

          {/* Mobile flow */}
          <div className="lg:hidden space-y-6">
            {steps.map((step, i) => (
              <div key={step.num} className="flex items-start gap-4">
                <div className="flex flex-col items-center">
                  <div className="w-10 h-10 rounded-full flex items-center justify-center text-white font-bold shadow-md shrink-0" style={{ backgroundColor: "var(--color-trust-blue)" }}>
                    {step.num}
                  </div>
                  {i < 4 && <div className="w-0.5 h-12 bg-blue-100 mt-2" />}
                </div>
                <div className="flex-1 pb-4">
                  <h3 className="font-semibold text-base text-gray-800">{step.title}</h3>
                  <span className={`text-xs font-semibold px-2 py-0.5 rounded-full ${step.tagColor}`}>{step.tag}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Data Flow Architecture */}
      <section className="py-20 md:py-28">
        <div className="max-w-6xl mx-auto px-4">
          <div className="mb-16">
            <h2 className="font-display text-3xl md:text-4xl font-bold text-gray-900 mb-2">数据流转架构</h2>
            <p className="text-gray-500">基于大规模语言模型的智能旅行决策引擎</p>
          </div>

          <div className="overflow-x-auto">
            <div className="flex items-center justify-between gap-3 min-w-[900px] p-8 bg-gray-50 rounded-2xl">
              {archNodes.map((node, i) => {
                if (!node) {
                  const labelIdx = Math.floor(i / 2)
                  return (
                    <div key={i} className="flex flex-col items-center gap-1 shrink-0">
                      <ArrowRight className="size-5 text-gray-300" />
                      <span className="text-[10px] text-gray-400 whitespace-nowrap bg-white px-2 py-0.5 rounded-full border border-gray-100 shadow-sm">
                        {dataLabels[labelIdx]}
                      </span>
                    </div>
                  )
                }
                return (
                  <div key={i} className={`flex flex-col items-center p-5 rounded-2xl border-2 shadow-sm shrink-0 ${node.bg} ${node.border}`}>
                    <div className="mb-2">{node.icon}</div>
                    <span className={`text-sm font-bold ${node.text || "text-gray-800"}`}>{node.label}</span>
                    <span className={`text-[10px] mt-0.5 ${node.text ? "opacity-70" : "text-gray-400"}`}>{node.sub}</span>
                  </div>
                )
              })}
            </div>
          </div>
        </div>
      </section>

      {/* Key Metrics */}
      <section className="py-20 bg-gray-50/50">
        <div className="max-w-5xl mx-auto px-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {[
              { value: "1.3秒", label: "平均解析速度", desc: "Heuristic 优先，LLM 兜底" },
              { value: "5+", label: "实时竞价方案", desc: "FlyAI 真实酒店+机票数据" },
              { value: "100%", label: "退改权益保障", desc: "平台风控资金池支持" },
            ].map((m) => (
              <div key={m.label} className="bg-white p-10 rounded-2xl shadow-sm border border-gray-100 text-center hover:shadow-md hover:-translate-y-1 transition-all">
                <div className="text-5xl font-black mb-2 font-display" style={{ color: "var(--color-trust-blue)" }}>{m.value}</div>
                <div className="text-gray-800 font-semibold mb-1">{m.label}</div>
                <div className="text-xs text-gray-400">{m.desc}</div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Trust Badges */}
      <section className="py-16 border-t border-gray-100">
        <div className="max-w-5xl mx-auto px-4">
          <div className="flex flex-wrap justify-center items-center gap-10">
            {[
              { icon: <Shield className="size-5" />, label: "飞猪验证" },
              { icon: <Lock className="size-5" />, label: "AES-256 加密" },
              { icon: <Trash2 className="size-5" />, label: "72 小时自动删除" },
            ].map((b) => (
              <div key={b.label} className="flex items-center gap-2 text-gray-500 hover:text-blue-600 transition-colors">
                {b.icon}
                <span className="font-semibold tracking-tight">{b.label}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="py-20 bg-gradient-to-r from-blue-600 to-blue-700 text-center">
        <div className="max-w-3xl mx-auto px-4">
          <h2 className="text-3xl md:text-4xl font-bold text-white mb-4 font-display">准备好开始你的旅程了吗？</h2>
          <p className="text-blue-100 mb-8">只需一句话，AI 为你搞定一切</p>
          <Link
            href="/crayfish"
            className="inline-flex items-center gap-2 px-8 py-4 rounded-xl text-blue-700 bg-white font-bold text-lg shadow-lg hover:shadow-xl transition-all hover:-translate-y-0.5"
          >
            开始体验 <ArrowRight className="size-5" />
          </Link>
        </div>
      </section>
    </main>
  )
}
