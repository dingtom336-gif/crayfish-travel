import type { Metadata } from "next"
import { Geist, Geist_Mono } from "next/font/google"
import "./globals.css"

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
})

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
})

export const metadata: Metadata = {
  title: "Crayfish Travel",
  description: "反向撮合旅行平台 - 让供应商为你竞价",
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html
      lang="zh-CN"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col">
        <nav className="border-b bg-white px-6 py-3">
          <span className="text-lg font-bold" style={{ color: "#1a73e8" }}>
            小龙虾旅行
          </span>
        </nav>
        <main className="flex flex-1 flex-col">{children}</main>
      </body>
    </html>
  )
}
