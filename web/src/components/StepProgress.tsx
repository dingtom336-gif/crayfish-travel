import { cn } from "@/lib/utils"

interface Step {
  label: string
  status: "completed" | "active" | "pending"
}

interface StepProgressProps {
  steps: Step[]
}

export function StepProgress({ steps }: StepProgressProps) {
  return (
    <div className="flex w-full max-w-3xl items-center">
      {steps.map((step, i) => (
        <div key={step.label} className="flex flex-1 items-center">
          <div className="flex flex-col items-center">
            <div
              className={cn(
                "flex size-10 items-center justify-center rounded-full text-sm font-bold",
                step.status === "completed" && "bg-[var(--color-success-green)] text-white",
                step.status === "active" && "bg-[var(--color-trust-blue)] text-white shadow-lg shadow-[var(--color-trust-blue)]/30",
                step.status === "pending" && "bg-gray-200 text-gray-500"
              )}
            >
              {step.status === "completed" ? (
                <svg className="size-5" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" />
                </svg>
              ) : (
                i + 1
              )}
            </div>
            <span
              className={cn(
                "mt-2 text-xs font-medium sm:text-sm",
                step.status === "completed" && "text-[var(--color-success-green)]",
                step.status === "active" && "font-bold text-[var(--color-trust-blue)]",
                step.status === "pending" && "text-gray-500"
              )}
            >
              {step.label}
            </span>
          </div>
          {i < steps.length - 1 && (
            <div
              className={cn(
                "-mt-5 h-1 flex-1",
                step.status === "completed" ? "bg-[var(--color-trust-blue)]" : "bg-gray-200"
              )}
            />
          )}
        </div>
      ))}
    </div>
  )
}
