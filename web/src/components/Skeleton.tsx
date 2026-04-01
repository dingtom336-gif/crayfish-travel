"use client"

import { cn } from "@/lib/utils"

function Skeleton({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn("animate-pulse rounded-md bg-muted", className)}
      {...props}
    />
  )
}

export function CardSkeleton() {
  return (
    <div className="rounded-xl border bg-card p-6 shadow-sm">
      <Skeleton className="mb-3 h-5 w-3/4" />
      <Skeleton className="mb-2 h-4 w-1/2" />
      <Skeleton className="mb-4 h-4 w-2/3" />
      <div className="space-y-2">
        <Skeleton className="h-3 w-full" />
        <Skeleton className="h-3 w-5/6" />
        <Skeleton className="h-3 w-4/6" />
      </div>
      <Skeleton className="mt-4 h-8 w-1/3" />
      <Skeleton className="mt-4 h-10 w-full" />
    </div>
  )
}

export function FormSkeleton() {
  return (
    <div className="mx-auto w-full max-w-lg rounded-xl border bg-card p-6 shadow-sm">
      <Skeleton className="mx-auto mb-2 h-7 w-40" />
      <Skeleton className="mx-auto mb-6 h-4 w-64" />
      <div className="space-y-4">
        {/* Chat area */}
        <div className="space-y-3 p-4 rounded-xl bg-gray-50">
          <div className="flex justify-end">
            <Skeleton className="h-8 w-48 rounded-2xl" />
          </div>
          <div className="flex items-start gap-2">
            <Skeleton className="h-6 w-6 rounded-full shrink-0" />
            <Skeleton className="h-10 w-64 rounded-2xl" />
          </div>
        </div>
        {/* Confirm card */}
        <div className="space-y-3 p-4 rounded-xl border">
          <Skeleton className="h-5 w-32" />
          <div className="grid grid-cols-2 gap-3">
            <Skeleton className="h-8 w-full rounded-lg" />
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-20" />
            <Skeleton className="h-4 w-28" />
          </div>
          <Skeleton className="h-8 w-24" />
        </div>
        <Skeleton className="h-14 w-full rounded-xl" />
      </div>
    </div>
  )
}

export function PackagesSkeleton() {
  return (
    <div className="mx-auto max-w-6xl px-4 py-8">
      {/* Lock banner skeleton */}
      <Skeleton className="mb-6 h-16 w-full rounded-xl bg-gradient-to-r from-orange-100 to-orange-50" />

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Cards column */}
        <div className="lg:col-span-2 grid grid-cols-1 md:grid-cols-2 gap-4">
          {Array.from({ length: 4 }, (_, i) => (
            <div key={i} className="rounded-xl border bg-card overflow-hidden shadow-sm">
              {/* Image area */}
              <Skeleton className="aspect-[3/2] w-full" />
              <div className="p-4 space-y-3">
                {/* Title */}
                <Skeleton className="h-5 w-4/5" />
                {/* Rating */}
                <div className="flex items-center gap-2">
                  <Skeleton className="h-4 w-8" />
                  <Skeleton className="h-4 w-16" />
                </div>
                {/* Tags */}
                <div className="flex gap-2">
                  <Skeleton className="h-6 w-14 rounded-full" />
                  <Skeleton className="h-6 w-20 rounded-full" />
                  <Skeleton className="h-6 w-24 rounded-full" />
                </div>
                {/* Price + button */}
                <div className="flex items-end justify-between pt-2">
                  <Skeleton className="h-8 w-24" />
                  <Skeleton className="h-10 w-28 rounded-full" />
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* Sidebar skeleton */}
        <div className="space-y-4">
          <div className="rounded-xl border bg-card p-6 shadow-sm space-y-4">
            <Skeleton className="h-5 w-32" />
            <Skeleton className="h-4 w-24" />
            <Skeleton className="h-4 w-40" />
            <Skeleton className="h-4 w-20" />
            <Skeleton className="h-4 w-28" />
          </div>
          <div className="rounded-xl border bg-card p-6 shadow-sm">
            <Skeleton className="h-4 w-full" />
            <Skeleton className="mt-2 h-4 w-3/4" />
          </div>
        </div>
      </div>
    </div>
  )
}

export function OrdersSkeleton() {
  return (
    <div className="mx-auto max-w-2xl px-4 py-8">
      <Skeleton className="mb-6 h-7 w-32" />
      <div className="space-y-4">
        {Array.from({ length: 2 }, (_, i) => (
          <div key={i} className="rounded-xl border bg-card p-6 shadow-sm space-y-4">
            <div className="flex items-center justify-between">
              <Skeleton className="h-5 w-48" />
              <Skeleton className="h-6 w-16 rounded-full" />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-4 w-32" />
              <Skeleton className="h-4 w-20" />
              <Skeleton className="h-4 w-28" />
            </div>
            <div className="flex items-center justify-between pt-2 border-t">
              <Skeleton className="h-7 w-24" />
              <Skeleton className="h-9 w-24 rounded-lg" />
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export function ErrorState({ message, onRetry }: { message: string; onRetry?: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <div className="mb-4 rounded-full bg-red-50 p-4">
        <svg className="h-8 w-8 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
        </svg>
      </div>
      <h3 className="text-lg font-semibold text-gray-700">出了点问题</h3>
      <p className="mt-1 max-w-sm text-sm text-muted-foreground">{message}</p>
      {onRetry && (
        <button
          onClick={onRetry}
          className="mt-4 rounded-lg bg-[var(--color-trust-blue)] px-4 py-2 text-sm font-medium text-white hover:bg-[var(--color-trust-blue-dark)]"
        >
          重试
        </button>
      )}
    </div>
  )
}

export function EmptyState({ icon, title, description }: { icon: string; title: string; description: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <div className="mb-4 text-4xl" aria-hidden="true">{icon}</div>
      <h3 className="text-lg font-semibold text-gray-600">{title}</h3>
      <p className="mt-1 max-w-sm text-sm text-muted-foreground">{description}</p>
    </div>
  )
}

export { Skeleton }
