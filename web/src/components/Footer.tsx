import Link from "next/link"

const FOOTER_LINKS = [
  { label: "服务条款", href: "/terms" },
  { label: "退款权益说明", href: "/refund-policy" },
  { label: "隐私政策", href: "/privacy" },
] as const

export function Footer() {
  return (
    <footer
      className="w-full border-t border-gray-200 bg-gray-50 py-8 dark:border-gray-800 dark:bg-gray-950"
      role="contentinfo"
    >
      <div className="mx-auto flex max-w-7xl flex-col items-center gap-3 px-4 text-center">
        <nav className="flex gap-6" aria-label="Footer links">
          {FOOTER_LINKS.map((link) => (
            <Link
              key={link.label}
              href={link.href}
              className="text-xs text-gray-400 transition-colors hover:text-[var(--color-trust-blue)] dark:text-gray-500 dark:hover:text-[var(--color-trust-blue)]"
            >
              {link.label}
            </Link>
          ))}
        </nav>
        <div className="flex flex-col gap-1 text-xs text-gray-400 dark:text-gray-500">
          <p>{new Date().getFullYear()} 小龙虾旅行 | 京ICP备12345678号-1</p>
          <p>增值电信业务经营许可证：京B2-20240001</p>
        </div>
      </div>
    </footer>
  )
}
