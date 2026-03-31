"use client"

import Link from "next/link"
import { Button } from "@/components/ui/button"

const NAV_LINKS = [
  { label: "首页", href: "/" },
  { label: "我的行程", href: "/orders" },
  { label: "无忧服务", href: "#" },
  { label: "帮助", href: "#" },
] as const

export function Navbar() {
  return (
    <header className="fixed top-0 z-50 w-full border-b border-gray-100 bg-white shadow-sm">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 md:px-8">
        <div className="flex items-center gap-8">
          <Link
            href="/"
            className="font-display text-xl font-bold text-[var(--color-trust-blue)]"
          >
            小龙虾旅行
          </Link>
          <nav className="hidden items-center gap-6 md:flex">
            {NAV_LINKS.map((link) => (
              <Link
                key={link.label}
                href={link.href}
                className="text-sm text-gray-600 transition-colors hover:text-[var(--color-trust-blue)]"
              >
                {link.label}
              </Link>
            ))}
          </nav>
        </div>
        <div className="hidden items-center gap-3 md:flex">
          <Button
            variant="outline"
            size="sm"
            className="border-[var(--color-trust-blue)] text-[var(--color-trust-blue)]"
          >
            Sign In
          </Button>
        </div>
        {/* Mobile: brand only */}
        <div className="flex items-center gap-2 md:hidden">
          <span className="text-xs font-medium text-gray-500">安全加密中</span>
        </div>
      </div>
    </header>
  )
}
