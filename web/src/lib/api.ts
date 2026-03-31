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
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      "X-Trace-ID": generateTraceId(),
      ...options?.headers,
    },
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `Request failed: ${res.status}`)
  }

  return res.json()
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
