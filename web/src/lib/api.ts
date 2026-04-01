const API_BASE = process.env.NEXT_PUBLIC_API_BASE ?? "http://localhost:8080/api/v1"

function generateTraceId(): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID()
  }
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    return (c === "x" ? r : (r & 0x3) | 0x8).toString(16)
  })
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const maxRetries = 3
  const baseTimeout = 30000 // 30s

  for (let attempt = 0; attempt < maxRetries; attempt++) {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), baseTimeout)

    try {
      const res = await fetch(`${API_BASE}${path}`, {
        ...options,
        signal: controller.signal,
        headers: {
          "Content-Type": "application/json",
          "X-Trace-ID": generateTraceId(),
          ...options?.headers,
        },
      })
      clearTimeout(timeoutId)

      if (!res.ok) {
        const body = await res.json().catch(() => ({}))
        throw new Error(body.error || `请求失败: ${res.status}`)
      }

      return res.json()
    } catch (err) {
      clearTimeout(timeoutId)
      if (err instanceof DOMException && err.name === "AbortError") {
        if (attempt < maxRetries - 1) {
          await new Promise(r => setTimeout(r, 1000 * Math.pow(2, attempt)))
          continue
        }
        throw new Error("请求超时，请检查网络后重试")
      }
      if (attempt < maxRetries - 1 && err instanceof TypeError) {
        await new Promise(r => setTimeout(r, 1000 * Math.pow(2, attempt)))
        continue
      }
      throw err
    }
  }
  throw new Error("请求失败，请稍后重试")
}

export interface IdentityResponse {
  session_id: string
  expires_at: string
  trace_id: string
}

export interface TravelRequirement {
  destination: string
  start_date: string
  end_date: string
  budget_cents: number
  adults: number
  children: number
  preferences: string[]
}

export interface DateValidation {
  is_valid: boolean
  is_peak_season: boolean
  peak_type: string
  warning: string
}

export interface ParseResponse {
  session_id: string
  requirement: TravelRequirement
  validation: DateValidation
  trace_id: string
}

export interface ConfirmResponse {
  session_id: string
  status: string
  is_peak_season: boolean
  peak_type: string
  trace_id: string
}

export interface RankedQuote {
  id?: string
  supplier: string
  package_title: string
  destination: string
  duration_days: number
  duration_nights: number
  base_price_cents: number
  refund_guarantee_fee_cents: number
  total_price_cents: number
  star_rating: number
  review_count: number
  hotel_name: string
  highlights: string[]
  inclusions: string[]
  image_url: string
  rank: number
  is_best_value: boolean
}

export interface BiddingResponse {
  session_id: string
  packages: RankedQuote[]
  count: number
  trace_id: string
}

export interface LockResponse {
  lock_session_id: string
  state: string
  expires_at: string
  ttl_seconds: number
  trace_id: string
}

export interface PaymentResponse {
  payment_id: string
  out_trade_no: string
  qr_code_url: string
  voice_token: string
  method: string
  trace_id: string
}

export type OrderStatus = "created" | "confirmed" | "fulfilling" | "completed" | "refund_requested" | "refunded"

export interface Order {
  id: string
  order_no: string
  status: OrderStatus
  package_title: string
  destination: string
  start_date: string
  end_date: string
  total_amount_cents: number
  base_price_cents: number
  refund_guarantee_fee_cents: number
  supplier: string
  adults: number
  children: number
  created_at: string
}

export interface SessionResponse {
  session_id: string
  destination: string
  start_date: string
  end_date: string
  budget_cents: number
  adults: number
  children: number
  preferences: string[]
  status: string
}

// readSSE reads an SSE response stream and dispatches events to the handler.
function readSSE<T>(
  path: string,
  body: unknown,
  onProgress: (step: string, message: string) => void,
): Promise<T> {
  return new Promise((resolve, reject) => {
    const controller = new AbortController()
    const timeout = setTimeout(() => {
      controller.abort()
      reject(new Error("请求超时"))
    }, 60000)

    fetch(`${API_BASE}${path}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Trace-ID": generateTraceId(),
      },
      body: JSON.stringify(body),
      signal: controller.signal,
    })
      .then((res) => {
        if (!res.ok) {
          return res
            .json()
            .catch(() => ({}))
            .then((b: Record<string, string>) => {
              throw new Error(b.error || `请求失败: ${res.status}`)
            })
        }

        const reader = res.body!.getReader()
        const decoder = new TextDecoder()
        let buffer = ""

        function read(): Promise<void> {
          return reader.read().then(({ done, value }) => {
            if (done) {
              clearTimeout(timeout)
              return
            }
            buffer += decoder.decode(value, { stream: true })
            const lines = buffer.split("\n")
            buffer = lines.pop() || ""
            let currentEvent = ""

            for (const line of lines) {
              if (line.startsWith("event: ")) {
                currentEvent = line.slice(7)
              } else if (line.startsWith("data: ")) {
                try {
                  const data = JSON.parse(line.slice(6))
                  if (currentEvent === "progress") {
                    onProgress(data.step, data.message)
                  } else if (currentEvent === "error") {
                    clearTimeout(timeout)
                    reject(new Error(data.message))
                    return
                  } else if (currentEvent === "result") {
                    clearTimeout(timeout)
                    resolve(data as T)
                  }
                } catch {
                  // skip malformed JSON lines
                }
              }
            }
            return read()
          })
        }
        read().catch((err) => {
          clearTimeout(timeout)
          reject(err)
        })
      })
      .catch((err) => {
        clearTimeout(timeout)
        reject(err)
      })
  })
}

export const api = {
  createIdentity(data: {
    name: string
    id_number: string
    phone: string
    adults: number
    children: number
  }) {
    return request<IdentityResponse>("/identity", {
      method: "POST",
      body: JSON.stringify(data),
    })
  },

  createSession(data: { adults: number; children: number }) {
    return request<{ session_id: string }>("/sessions", {
      method: "POST",
      body: JSON.stringify(data),
    })
  },

  getSession(session_id: string) {
    return request<SessionResponse>(`/sessions/${encodeURIComponent(session_id)}`)
  },

  parse(session_id: string, raw_input: string) {
    return request<ParseResponse>("/nlp/parse", {
      method: "POST",
      body: JSON.stringify({ session_id, raw_input }),
    })
  },

  parseStream(
    session_id: string,
    raw_input: string,
    onProgress: (step: string, message: string) => void,
  ): Promise<ParseResponse> {
    return readSSE<ParseResponse>(
      "/nlp/parse/stream",
      { session_id, raw_input },
      onProgress,
    )
  },

  confirm(data: {
    session_id: string
    destination: string
    start_date: string
    end_date: string
    budget_cents: number
    adults: number
    children: number
    preferences: string[]
  }) {
    return request<ConfirmResponse>("/nlp/confirm", {
      method: "POST",
      body: JSON.stringify(data),
    })
  },

  startBidding(session_id: string) {
    return request<BiddingResponse>("/bidding/start", {
      method: "POST",
      body: JSON.stringify({ session_id }),
    })
  },

  startBiddingStream(
    session_id: string,
    onProgress: (step: string, message: string) => void,
  ): Promise<BiddingResponse> {
    return readSSE<BiddingResponse>(
      "/bidding/start/stream",
      { session_id },
      onProgress,
    )
  },

  acquireLock(session_id: string, quote_id: string) {
    return request<LockResponse>("/lock/acquire", {
      method: "POST",
      body: JSON.stringify({ session_id, quote_id }),
    })
  },

  createPayment(session_id: string, quote_id: string, method: "qr" | "voice_token") {
    return request<PaymentResponse>("/payment/create", {
      method: "POST",
      body: JSON.stringify({ session_id, quote_id, method }),
    })
  },

  async listOrders(session_id: string): Promise<Order[]> {
    // Backend wraps orders in { orders: [...] } envelope
    const result = await request<{ orders: Order[] }>(`/orders?session_id=${encodeURIComponent(session_id)}`)
    return result.orders ?? []
  },

  requestRefund(order_id: string) {
    return request<{ status: string }>(`/orders/${encodeURIComponent(order_id)}/refund`, {
      method: "POST",
    })
  },
}
