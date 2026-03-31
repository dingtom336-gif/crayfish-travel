"use client"

import { Button } from "@/components/ui/button"

interface CompareBarProps {
  selectedCount: number
  onCompare: () => void
}

export function CompareBar({ selectedCount, onCompare }: CompareBarProps) {
  return (
    <div className="fixed bottom-0 left-0 right-0 z-50 pointer-events-none">
      <div className="mx-auto max-w-7xl px-6 pb-6">
        <div className="pointer-events-auto rounded-2xl border border-white bg-white/80 p-4 shadow-2xl backdrop-blur-xl">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <span className="rounded-lg bg-[var(--color-trust-blue)]/10 p-2">
                <svg className="h-5 w-5 text-[var(--color-trust-blue)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M7.5 21L3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" />
                </svg>
              </span>
              <div>
                <p className="text-sm font-bold">对比选中方案</p>
                <p className="text-[10px] text-gray-500">已选择 {selectedCount}/3 个方案</p>
              </div>
            </div>

            <div className="flex items-center gap-3">
              <div className="flex -space-x-3">
                {Array.from({ length: Math.min(selectedCount, 3) }, (_, i) => (
                  <div
                    key={i}
                    className="h-10 w-10 overflow-hidden rounded-full border-2 border-white bg-gray-100 shadow-sm"
                  />
                ))}
                {selectedCount < 3 && (
                  <div className="flex h-10 w-10 items-center justify-center rounded-full border-2 border-white bg-gray-50 text-gray-300">
                    <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
                    </svg>
                  </div>
                )}
              </div>
              <Button
                onClick={onCompare}
                className="rounded-full bg-[var(--color-trust-blue)] px-6 text-sm font-bold text-white hover:opacity-90"
              >
                开始对比
              </Button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
