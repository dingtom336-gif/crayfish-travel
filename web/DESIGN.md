# Design System: Crayfish Travel - 小龙虾旅行反向撮合
**Project ID:** 15920578822705138428
**Design System Assets:** 7930407182640895117 (Trust Blue), 15139940954258061869 (Crayfish Travel Design System)

---

## 1. Visual Theme & Atmosphere

A **trust-forward, clarity-driven** travel bidding platform designed for Chinese family travelers. The aesthetic is warm yet professional: a clean white canvas punctuated by confident Trust Blue headers, energetic Vibrant Orange calls-to-action, and reassuring Success Green trust signals. The overall density is **medium** -- airy enough to breathe, dense enough to convey value.

Key atmospheric qualities:
- **Reassuring**: Every screen reinforces credibility through verification badges, real case studies, and transparent pricing
- **Urgent but not anxious**: Countdown timers and lock indicators use warm orange rather than aggressive red
- **Mobile-first Chinese**: Optimized for touch targets, Chinese typographic density, and Simplified Chinese copy throughout
- **Glass-morphism accents**: Subtle frosted glass effects (rgba white + backdrop blur) on overlays and navigation elements

---

## 2. Color Palette & Roles

### Primary Colors

| Name | Hex | Role |
|------|-----|------|
| **Trust Blue** | `#1a73e8` | Primary actions, headers, navigation, trust signals, destination badges, price display |
| **Vibrant Orange** | `#ff6d3f` | CTAs (Confirm & Bidding, Select Package), countdown timers, peak season warnings, urgency indicators |
| **Orange Gradient End** | `#ff8c66` | Gradient partner for orange (135deg gradient: `#ff6d3f` -> `#ff8c66`) |
| **Success Green** | `#34a853` | Trust badges (verified), best value tags, confirmed states, preference tags |

### Neutral & Surface Colors

| Name | Hex | Role |
|------|-----|------|
| **Pure White** | `#ffffff` | Card backgrounds, page background, primary surface |
| **On-Primary White** | `#ffffff` | Text on colored backgrounds (blue/orange/green) |
| **Glass White** | `rgba(255,255,255,0.8)` | Frosted glass overlays with `backdrop-filter: blur(12px)` |
| **Light Gray Background** | `#f8f9fa` | Page background (zinc-50 equivalent) |
| **Muted Text** | `#6b7280` | Secondary labels, descriptions, itemized prices |
| **Dark Text** | `#1f2937` | Primary body text, card titles |

### Functional Colors

| Name | Hex | Role |
|------|-----|------|
| **Error Red** | `#ef4444` | Destructive actions, refunded status |
| **Warning Orange** | `#ff6d3f` | Peak season alerts, refund-requested status |

---

## 3. Typography Rules

### Font Stack
- **Headlines**: Plus Jakarta Sans, PingFang SC, Microsoft YaHei, sans-serif
- **Body & Labels**: Inter, PingFang SC, Microsoft YaHei, sans-serif
- **Chinese fallback priority**: PingFang SC (macOS/iOS) -> Microsoft YaHei (Windows) -> sans-serif

### Scale & Weight
| Usage | Size | Weight | Notes |
|-------|------|--------|-------|
| Page title / Brand | 2xl (1.5rem) | Bold (700) | Trust Blue color |
| Card title | lg (1.125rem) | Bold (700) | Dark text |
| Section label | xs (0.75rem) | Normal (400) | Muted text, uppercase-like |
| Price total | 2xl (1.5rem) | Bold (700) | Trust Blue color |
| Price breakdown | xs (0.75rem) | Normal (400) | Muted text |
| Button text | sm (0.875rem) | Medium (500) | White on colored bg |
| Body text | sm (0.875rem) | Normal (400) | Dark text |
| Badge text | xs (0.75rem) | Medium (500) | Varies by badge type |

### Icon System
- **Material Symbols Outlined**: `font-variation-settings: 'FILL' 0, 'wght' 400, 'GRAD' 0, 'opsz' 24`
- Icons are inline-block, vertically centered (`vertical-align: middle`)
- Icon size: 24px default, 20px in compact contexts, 16px in badges

---

## 4. Component Stylings

### Buttons
| Variant | Background | Text | Shape | Usage |
|---------|-----------|------|-------|-------|
| **Primary (Trust Blue)** | `#1a73e8` | White | Generously rounded (`rounded-xl`, ~1rem) | Submit forms, confirm actions |
| **Accent (Vibrant Orange)** | `linear-gradient(135deg, #ff6d3f, #ff8c66)` | White | Generously rounded | CTA: "Confirm & Start Bidding", "Select Package" |
| **Secondary (Outline)** | Transparent | Dark text | Rounded with 1px border | Edit, Back, Cancel |
| **Ghost** | Transparent | Trust Blue | No border | Navigation links |

### Cards & Containers
- **Corner roundness**: Generously rounded (1.5rem / `rounded-2xl`) for main cards, standard rounded (1rem / `rounded-xl`) for nested elements
- **Background**: Pure white (`#ffffff`)
- **Shadow**: Whisper-soft diffused shadow (`0 1px 3px rgba(0,0,0,0.1), 0 1px 2px rgba(0,0,0,0.06)`)
- **Ring**: Subtle 1px ring (`ring-1 ring-foreground/10`) for secondary elevation

### Package Cards (Stitch Design)
- Hero image at top (40% card height, `object-cover`)
- Verification badge floating on image (top-right): green circle + white checkmark + "飞猪验证"
- Best Value badge floating on image (top-left): green pill "超值首选"
- Title below image, bold, dark
- Highlights as colored pill tags (not bullet list)
- Price: large Trust Blue + small muted breakdown
- CTA button: full-width Vibrant Orange gradient

### Trust Badges
- Green circle icon + bold credential text
- Layout: horizontal row on desktop, scrollable on mobile
- Examples: "持牌旅行服务平台", "支付宝资金保障", "100%退改保障", "2000+家庭已服务"

### Countdown Timer
- **Container**: Vibrant Orange pill (`rounded-full`)
- **Text**: White, bold, monospace-like
- **Format**: `MM:SS` with zero-padding

### Navigation Bar
- Height: 48-56px
- Background: White with subtle bottom border
- Brand text: Trust Blue, bold
- Right items: Help icon, account icon (Material Symbols)
- Mobile: simplified, brand only

### Progress Indicator (Step bar)
- Horizontal dots/steps at bottom of confirm page
- Active: Trust Blue filled circle
- Completed: Green checkmark
- Upcoming: Gray outline circle

---

## 5. Layout Principles

### Spacing Strategy
- **Page padding**: 16px (px-4) on mobile, 32px on desktop
- **Card internal padding**: 24px (p-6)
- **Section gap**: 20px (gap-5) between info sections
- **Grid gap**: 16px (gap-4) between cards

### Grid System
- **Desktop packages**: 2-column grid + right sidebar (3-column layout)
- **Mobile packages**: single column, full-width cards
- **Forms**: single column, max-width 32rem (512px)
- **Content pages**: max-width 42rem (672px) for payment/orders

### Breakpoints
- **Mobile**: < 640px (sm) -- single column, full-width
- **Tablet**: 640px - 1024px (md) -- 2 columns where applicable
- **Desktop**: > 1024px (lg) -- full layout with sidebars

### Whitespace
- Generous vertical spacing between sections (2rem+)
- Cards have consistent internal padding (1.5rem)
- Tight spacing within card content sections (0.5-1rem)

---

## 6. Stitch Screen Inventory

| Screen ID | Title | Device | Dimensions | Page Mapping |
|-----------|-------|--------|------------|-------------|
| `af1234ec` | Identity Collection Page | Desktop | 2560x2342 | `/` (Homepage) |
| `da7a5156` | Identity Collection Mobile | Mobile | 780x2818 | `/` (Homepage) |
| `dc1d1c68` | Requirements Input Page | Desktop | 2560x2094 | `/confirm` |
| `1e22d5dc` | Requirements Input Mobile | Mobile | 780x2230 | `/confirm` |
| `d736599f` | Package Selection Page | Desktop | 2560x3780 | `/packages` |
| `3d73af1b` | Package Display Mobile | Mobile | 780x4196 | `/packages` |

### Missing Stitch Screens
- **Payment page**: No Desktop or Mobile design exists
- **Orders page**: No Desktop or Mobile design exists
- **Layout/Nav**: Embedded in each screen, no standalone design

---

## 7. Compliance Rules (CRITICAL)

- **NEVER** use terms: insurance, premium, claim, underwrite, policy (or Chinese equivalents: 保险, 保费, 理赔, 承保, 投保)
- **ALWAYS** use: refund guarantee (退改保障), worry-free refund service (无忧退改服务), refund privilege (退款特权), service fee (服务费)
- **Price display MUST** show breakdown: base price + refund guarantee fee (基础价格 + 退改保障费)
