"use client"

import { useState } from "react"
import Link from "next/link"
import { Menu, X } from "lucide-react"
import { Button } from "@/components/ui/button"
import { getPersistedSessionId } from "@/lib/session-store"

function getOrdersHref(): string {
  if (typeof window === "undefined") return "/orders"
  const sessionId = getPersistedSessionId()
  if (sessionId) return `/orders?session_id=${encodeURIComponent(sessionId)}`
  return "/orders"
}

const STATIC_NAV_LINKS = [
  { label: "首页", href: "/" },
  { label: "无忧服务", href: "#" },
  { label: "帮助", href: "#" },
] as const

export function Navbar() {
  const [menuOpen, setMenuOpen] = useState(false)

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
            {STATIC_NAV_LINKS.map((link) => (
              <Link
                key={link.label}
                href={link.href}
                className="text-sm text-gray-600 transition-colors hover:text-[var(--color-trust-blue)]"
              >
                {link.label}
              </Link>
            ))}
            <Link
              href="/orders"
              onClick={(e) => {
                e.preventDefault()
                window.location.href = getOrdersHref()
              }}
              className="text-sm text-gray-600 transition-colors hover:text-[var(--color-trust-blue)]"
            >
              我的行程
            </Link>
          </nav>
        </div>
        <div className="hidden items-center gap-3 md:flex">
          <Button
            variant="outline"
            size="sm"
            className="border-[var(--color-trust-blue)] text-[var(--color-trust-blue)]"
          >
            登录
          </Button>
        </div>

        {/* Mobile hamburger button */}
        <button
          type="button"
          className="flex items-center justify-center md:hidden"
          aria-label={menuOpen ? "关闭菜单" : "打开菜单"}
          onClick={() => setMenuOpen((prev) => !prev)}
        >
          {menuOpen ? (
            <X className="size-6 text-gray-700" />
          ) : (
            <Menu className="size-6 text-gray-700" />
          )}
        </button>
      </div>

      {/* Mobile dropdown menu */}
      {menuOpen && (
        <div className="border-t border-gray-100 bg-white md:hidden">
          <nav className="flex flex-col px-4 py-4 space-y-3">
            {STATIC_NAV_LINKS.map((link) => (
              <Link
                key={link.label}
                href={link.href}
                className="text-sm font-medium text-gray-700 py-2 transition-colors hover:text-[var(--color-trust-blue)]"
                onClick={() => setMenuOpen(false)}
              >
                {link.label}
              </Link>
            ))}
            <Link
              href="/orders"
              className="text-sm font-medium text-gray-700 py-2 transition-colors hover:text-[var(--color-trust-blue)]"
              onClick={(e) => {
                e.preventDefault()
                setMenuOpen(false)
                window.location.href = getOrdersHref()
              }}
            >
              我的行程
            </Link>
            <div className="pt-2 border-t border-gray-100">
              <Button
                variant="outline"
                size="sm"
                className="w-full border-[var(--color-trust-blue)] text-[var(--color-trust-blue)]"
              >
                登录
              </Button>
            </div>
          </nav>
        </div>
      )}
    </header>
  )
}
