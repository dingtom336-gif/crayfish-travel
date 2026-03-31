import Link from "next/link"

const FOOTER_LINKS = [
  { label: "服务条款", href: "#" },
  { label: "退款权益说明", href: "#" },
  { label: "隐私政策", href: "#" },
] as const

export function Footer() {
  return (
    <footer className="w-full border-t border-gray-200 bg-gray-50 py-8">
      <div className="mx-auto flex max-w-7xl flex-col items-center gap-3 px-4 text-center">
        <div className="flex gap-6">
          {FOOTER_LINKS.map((link) => (
            <Link
              key={link.label}
              href={link.href}
              className="text-xs text-gray-400 transition-colors hover:text-[var(--color-trust-blue)]"
            >
              {link.label}
            </Link>
          ))}
        </div>
        <div className="flex flex-col gap-1 text-xs text-gray-400">
          <p>2024 小龙虾旅行 | 京ICP备12345678号-1</p>
          <p>增值电信业务经营许可证：京B2-20240001</p>
        </div>
      </div>
    </footer>
  )
}
