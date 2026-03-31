export function TrustGuarantee() {
  return (
    <div className="mt-8 rounded-lg border border-[var(--color-success-green)]/10 bg-[var(--color-success-green)]/5 p-4">
      <div className="flex items-start gap-2">
        <svg
          className="h-5 w-5 shrink-0 text-[var(--color-success-green)]"
          viewBox="0 0 24 24"
          fill="currentColor"
        >
          <path
            fillRule="evenodd"
            d="M12.516 2.17a.75.75 0 00-1.032 0 11.209 11.209 0 01-7.877 3.08.75.75 0 00-.722.515A12.74 12.74 0 002.25 9.75c0 5.942 4.064 10.932 9.563 12.348a.749.749 0 00.374 0c5.499-1.416 9.563-6.406 9.563-12.348 0-1.39-.223-2.73-.635-3.985a.75.75 0 00-.722-.516l-.143.001c-2.996 0-5.717-1.17-7.734-3.08zm3.094 8.016a.75.75 0 10-1.22-.872l-3.236 4.53L9.53 12.22a.75.75 0 00-1.06 1.06l2.25 2.25a.75.75 0 001.14-.094l3.75-5.25z"
            clipRule="evenodd"
          />
        </svg>
        <div>
          <p className="text-xs font-bold text-[var(--color-success-green)]">100% 服务保障</p>
          <p className="mt-1 text-[10px] text-gray-500">
            所有方案均包含无忧退改服务，保障您的出行权益。
          </p>
        </div>
      </div>
    </div>
  )
}
