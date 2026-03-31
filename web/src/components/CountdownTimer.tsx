"use client"

import { useEffect, useState } from "react"

interface CountdownTimerProps {
  expiresAt: number
}

export function CountdownTimer({ expiresAt }: CountdownTimerProps) {
  const [remaining, setRemaining] = useState(() =>
    Math.max(0, Math.floor((expiresAt - Date.now()) / 1000))
  )

  useEffect(() => {
    const timer = setInterval(() => {
      const left = Math.max(0, Math.floor((expiresAt - Date.now()) / 1000))
      setRemaining((prev) => (prev === left ? prev : left))
      if (left <= 0) clearInterval(timer)
    }, 1000)

    return () => clearInterval(timer)
  }, [expiresAt])

  const minutes = Math.floor(remaining / 60)
  const seconds = remaining % 60

  return (
    <span className="inline-flex items-center rounded-full bg-[#ff6d3f] px-3 py-1 text-sm font-semibold text-white">
      {minutes.toString().padStart(2, "0")}:{seconds.toString().padStart(2, "0")}
    </span>
  )
}
