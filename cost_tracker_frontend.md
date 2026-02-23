# Frontend PRD: Claude Code Cost Tracker — Vue SPA

**Version:** 0.1 — MVP  
**Status:** Draft  
**Companion to:** `claude-code-cost-tracker-prd.md`  
**Last Updated:** 2026-02-22

---

## 1. Purpose & Scope

This document specifies the frontend of `cctrack` — a Vue 3 SPA embedded in the Go binary via `go:embed`. It covers design language, component architecture, routing, state management, real-time behaviour, and animation system.

The frontend is a local-only app served from `http://localhost:7432`. There is no CDN, no external font loading at runtime, no analytics. All assets ship inside the binary.

---

## 2. Design Language

### 2.1 Concept

**"The Cost of Thinking"** — a premium editorial dashboard that treats token spend data the way a financial publication treats market data: with authority, precision, and typographic confidence. Not a SaaS product. Not a dev tool that apologises for its existence. Something you'd screenshot and share.

The aesthetic sits at the intersection of:
- A high-end financial terminal (Bloomberg, Koyfin)
- An editorial design publication (Are.na, Fonts In Use, Stripe Press)
- A premium developer tool (Linear, Raycast)

Dark canvas. Sharp geometric type for headlines. Monospace for numbers. Amber as the single accent — warm, economic, expensive-feeling.

### 2.2 Colour Palette

All colours defined as CSS custom properties on `:root`.

```css
:root {
  /* Base */
  --bg-base:        #0a0a09;   /* Near-black with warm undertone */
  --bg-surface:     #111110;   /* Cards, panels */
  --bg-elevated:    #1a1a18;   /* Hover states, tooltips */
  --bg-subtle:      #222220;   /* Input fills, table stripes */

  /* Borders */
  --border-default: #2a2a27;
  --border-subtle:  #1e1e1b;
  --border-strong:  #3a3a36;

  /* Text */
  --text-primary:   #f0ede8;   /* Warm white — not harsh */
  --text-secondary: #8c8a84;
  --text-tertiary:  #5a5855;
  --text-disabled:  #3a3835;

  /* Amber accent — the money colour */
  --amber-500:      #f59e0b;
  --amber-400:      #fbbf24;
  --amber-300:      #fcd34d;
  --amber-600:      #d97706;
  --amber-glow:     rgba(245, 158, 11, 0.15);
  --amber-glow-sm:  rgba(245, 158, 11, 0.08);

  /* Semantic */
  --cost-low:       #4ade80;   /* Green — under budget */
  --cost-mid:       #f59e0b;   /* Amber — approaching */
  --cost-high:      #f87171;   /* Red — over threshold */

  /* WebSocket status */
  --status-live:    #4ade80;
  --status-reconnecting: #f59e0b;
  --status-offline: #f87171;
}
```

**Rule:** amber is used sparingly — primary CTAs, active states, the live spend number, chart highlights. Everything else is greyscale. One colour pops because everything else steps back.

### 2.3 Typography

All fonts self-hosted in `web/public/fonts/` — no Google Fonts CDN call at runtime.

| Role | Font | Weight | Usage |
|---|---|---|---|
| Display | **Bebas Neue** | 400 | Large metric values, section headers, nav logo |
| UI / Body | **DM Sans** | 300, 400, 500 | Labels, body copy, navigation, table text |
| Data / Mono | **JetBrains Mono** | 400, 500 | Token counts, session IDs, costs, code references |

```css
/* Type scale */
--text-xs:   0.6875rem;   /* 11px — table metadata */
--text-sm:   0.8125rem;   /* 13px — labels, secondary */
--text-base: 0.9375rem;   /* 15px — body */
--text-lg:   1.0625rem;   /* 17px — UI emphasis */
--text-xl:   1.25rem;     /* 20px — card titles */
--text-2xl:  1.625rem;    /* 26px — section headers */
--text-3xl:  2.25rem;     /* 36px — summary values (DM Sans) */
--text-hero: 4.5rem;      /* 72px — primary spend metric (Bebas) */
--text-display: 7rem;     /* 112px — page-level hero number */
```

**Pairing rule:** Bebas Neue only for display numerics and structural headings. DM Sans everywhere else. JetBrains Mono strictly for data — costs, tokens, IDs, timestamps.

### 2.4 Motion System

Motion is purposeful, not decorative. Two moments of delight; everything else is instant or subtle.

**Entrance animations (page load / route change)**

```css
/* Staggered reveal for cards */
@keyframes fadeSlideUp {
  from { opacity: 0; transform: translateY(12px); }
  to   { opacity: 1; transform: translateY(0); }
}

.card { animation: fadeSlideUp 0.4s ease both; }
.card:nth-child(1) { animation-delay: 0ms; }
.card:nth-child(2) { animation-delay: 60ms; }
.card:nth-child(3) { animation-delay: 120ms; }
.card:nth-child(4) { animation-delay: 180ms; }
```

**Counter tick-up (on load and on WebSocket update)**
- Spend values animate from 0 (or previous value) to new value over 800ms
- Easing: `cubic-bezier(0.16, 1, 0.3, 1)` — fast start, eases out
- Implementation: Vue composable `useCountUp(target, duration)`
- Triggers on: initial data load, `summary.updated` WebSocket event

**Chart bar animation (on load)**
- Bars grow from baseline upward over 600ms
- Staggered: each bar delayed by `index * 20ms`
- Chart.js animation config: `{ duration: 600, easing: 'easeOutQuart', delay: (ctx) => ctx.index * 20 }`

**Live pulse (active session)**
- Active session row: amber left border + subtle background pulse
- `@keyframes pulse` — opacity cycles 1 → 0.5 → 1 over 2s, `animation-iteration-count: infinite`
- Token counter in active session ticks up in real time as WebSocket events arrive

**Hover states**
- Cards: `background` transitions to `--bg-elevated` over 150ms
- Table rows: same, 100ms
- No transform lifts — feels more editorial, less app-like

**Disabled / no motion**
```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
```

### 2.5 Spatial System

8px base grid. All spacing in multiples of 4px minimum.

```css
--space-1:  4px;
--space-2:  8px;
--space-3:  12px;
--space-4:  16px;
--space-5:  20px;
--space-6:  24px;
--space-8:  32px;
--space-10: 40px;
--space-12: 48px;
--space-16: 64px;
--space-20: 80px;
```

Layout uses CSS Grid throughout. No Bootstrap, no utility-class framework — hand-written component styles using scoped `<style>` blocks.

### 2.6 Surfaces & Depth

Three levels of elevation. No drop shadows — separation is achieved through background colour steps and borders.

| Level | Background | Border | Usage |
|---|---|---|---|
| Base | `--bg-base` | — | Page canvas |
| Surface | `--bg-surface` | `--border-subtle` | Cards, panels, sidebar |
| Elevated | `--bg-elevated` | `--border-default` | Dropdowns, tooltips, hover states |

Subtle grain texture on the base layer — a CSS `background-image` SVG noise filter at 3% opacity adds depth without looking noisy. Makes the dark background feel material rather than flat.

---

## 3. Layout

### 3.1 Shell

Fixed sidebar navigation (240px wide) + scrollable main content area. No top nav bar.

```
┌─────────────────────────────────────────────────────┐
│  Sidebar (240px fixed)  │  Main content (flex: 1)   │
│                         │                           │
│  [logo / wordmark]      │  [page content]           │
│                         │                           │
│  nav links              │                           │
│                         │                           │
│  [connection status]    │                           │
│  [binary version]       │                           │
└─────────────────────────────────────────────────────┘
```

**Sidebar:**
- Background: `--bg-surface`
- Right border: `--border-subtle`
- Logo: "CCTRACK" in Bebas Neue, `--text-primary`, 22px. Small amber square glyph to the left.
- Nav items: DM Sans 400, 14px, `--text-secondary` default, `--text-primary` + amber left bar on active
- Bottom section: connection status dot + label, version string in `--text-tertiary`

**Main content:**
- Max-width: `1200px`, centred within the available space
- Padding: `--space-10` horizontal, `--space-8` top
- Page titles: Bebas Neue, 36px, `--text-primary`, letter-spacing: 0.02em

### 3.2 Responsive Behaviour

The app is desktop-first. On viewports under 900px:
- Sidebar collapses to a 56px icon-only rail
- Hamburger opens a slide-over drawer for nav
- Main content fills full width

This is a local developer tool — mobile support is a nice-to-have, not a core requirement.

---

## 4. Views & Components

### 4.1 Overview (Home)

The primary view. URL: `/`

**Hero metric strip**
Four stat cards in a row: Today / Week / Month / Projected.

Each card:
- Background: `--bg-surface`
- Label: DM Sans 300, 11px uppercase, letter-spacing 0.12em, `--text-tertiary`
- Value: Bebas Neue, 72px, `--text-primary` — animates in with counter tick-up on load
- Subtext: JetBrains Mono 12px, `--text-secondary` — e.g. "1.2M tokens"
- Today card only: amber left border (3px) to signal "this is now"

```
┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│ TODAY       │ │ THIS WEEK   │ │ THIS MONTH  │ │ PROJECTED   │
│             │ │             │ │             │ │             │
│ $0.84       │ │ $6.12       │ │ $24.37      │ │ $31.20      │
│ 1.2M tokens │ │ 8.9M tokens │ │ 34.1M tokens│ │ est.        │
└─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘
  ↑ amber border
```

**Daily spend chart**
- Chart.js bar chart, last 30 days
- Full width below the hero strip
- Bars: `--amber-500` at 70% opacity; today's bar: `--amber-400` at 100%
- Grid lines: `--border-subtle`
- No chart border, no legend — axis labels only
- Y-axis: dollar values, JetBrains Mono 11px
- X-axis: day labels, DM Sans 11px, `--text-tertiary`
- Tooltip: custom — `--bg-elevated` surface, amber value, mono font
- Bars animate upward on load (600ms staggered)
- On `summary.updated` WebSocket event: today's bar height updates smoothly (300ms transition)

**Token breakdown**
- Sits to the right of or below the chart (responsive)
- Donut chart (Chart.js): four segments — Input / Output / Cache Read / Cache Write
- Palette: amber shades + muted greys so the warm tokens stand out
- Legend: inline list, DM Sans 13px, percentage + token count in mono

**Top 5 sessions**
- Compact table below the charts
- Columns: Session name / Model / Cost / Tokens / Date
- Cost column: JetBrains Mono, amber for the #1 most expensive row
- Active session (if any): pulse animation on the row, live token counter ticking
- "View all sessions →" link to `/sessions`

### 4.2 Sessions

URL: `/sessions`

Full session history table.

**Table design:**
- Full-width, no card wrapper — the table *is* the content
- Alternating row backgrounds: `--bg-base` / `--bg-subtle` at 50% opacity
- Row hover: `--bg-elevated`, 100ms transition
- Columns: # / Session / Model / Started / Duration / Input / Output / Cached / **Total Cost**
- Total Cost column: right-aligned, JetBrains Mono, amber weight for top 5 rows
- Active session row: amber left border + live pulse

**Sorting:** click column headers. Active sort column: amber underline on header label.

**Pagination:** 25 rows per page. Simple prev/next with page count. No infinite scroll — predictable, fast.

**Session detail panel:**
- Clicking a row opens a right-side slide-over panel (not a new page)
- Panel width: 480px, `--bg-surface` background, `--border-default` left border
- Contents: session ID (mono, small), model badge, full token breakdown table, cost breakdown by token type, timeline (start / end / duration)
- Close: Escape or click outside

### 4.3 Settings

URL: `/settings`

Clean form layout. Not a modal — a full page. Feels like a preferences screen, not a config file.

**Sections:**

*Data Sources*
- Log directory path — text input with a "Verify" button that checks the path exists (calls `/api/v1/settings/verify`)
- DB path — read-only display, shown in mono

*Budget*
- Monthly budget (USD) — number input with dollar prefix
- When set: a thin amber progress bar appears under the "this month" hero card on Overview
- At 80%: bar colour shifts to `--cost-mid`; at 100%: `--cost-high`

*Dashboard*
- Port (read-only in UI, shown for reference)
- Open browser on serve — toggle switch
- Debounce window — read-only display (250ms), shown as an informational field

*About*
- Binary version, rate card version, build date — all mono, `--text-tertiary`
- Link to GitHub releases

**Form controls:**
- Inputs: `--bg-subtle` fill, `--border-default` border, focus ring in `--amber-500` at 40% opacity
- Toggle switches: custom CSS — track is `--bg-elevated`, amber when on
- Save button: amber background (`--amber-500`), dark text, no border radius (sharp corners fit the geometric aesthetic)

### 4.4 Rate Card

URL: `/rates`

Simple informational view. Shows the current model rates used for cost calculation.

- Table: Model / Input / Output / Cache Read / Cache Write / Last Updated
- All prices in JetBrains Mono
- Version badge: "Rate card v1.2 — Updated 2026-02-01"
- Note: "Rates are bundled with the binary. Update cctrack to get the latest rates."

---

## 5. Component Library

All components are project-local — no UI library dependency (no Vuetify, no PrimeVue, no shadcn-vue).

### 5.1 Core Components

```
web/src/components/
├── layout/
│   ├── AppShell.vue          — sidebar + main slot
│   ├── SidebarNav.vue        — nav links, logo, status
│   └── ConnectionStatus.vue  — WS status dot + label
├── primitives/
│   ├── StatCard.vue          — hero metric card with counter animation
│   ├── DataTable.vue         — sortable, paginated table
│   ├── SlideOver.vue         — right-side detail panel
│   ├── ProgressBar.vue       — budget progress, semantic colours
│   ├── Badge.vue             — model name badge (e.g. "Sonnet 4.6")
│   ├── Toggle.vue            — custom switch
│   └── Button.vue            — primary / ghost / danger variants
├── charts/
│   ├── DailySpendChart.vue   — Chart.js bar, 30-day view
│   └── TokenDonut.vue        — Chart.js donut breakdown
└── domain/
    ├── SessionRow.vue        — table row with live pulse
    ├── SessionDetail.vue     — slide-over content for a session
    └── BudgetWarning.vue     — banner shown when threshold exceeded
```

### 5.2 StatCard

The most important component. Used four times on the Overview hero strip.

Props:
```ts
interface StatCardProps {
  label: string           // "TODAY", "THIS WEEK", etc.
  value: number           // raw dollar value
  tokens: number          // for subtext
  highlight?: boolean     // true = amber left border (Today card)
  live?: boolean          // true = value updates via WS, no re-animation
}
```

Counter animation via `useCountUp` composable — watches `value` prop, animates when it changes. On WebSocket `summary.updated` events, the value prop is patched in the Pinia store and the animation fires automatically.

### 5.3 useCountUp Composable

```ts
// web/src/composables/useCountUp.ts
export function useCountUp(target: Ref<number>, duration = 800) {
  const display = ref(0)
  let raf: number
  let start: number
  let from: number

  watch(target, (to) => {
    from = display.value
    start = performance.now()
    cancelAnimationFrame(raf)

    function tick(now: number) {
      const elapsed = now - start
      const progress = Math.min(elapsed / duration, 1)
      // cubic-bezier(0.16, 1, 0.3, 1) approximation
      const ease = 1 - Math.pow(1 - progress, 4)
      display.value = from + (to - from) * ease
      if (progress < 1) raf = requestAnimationFrame(tick)
    }

    raf = requestAnimationFrame(tick)
  }, { immediate: true })

  return display
}
```

### 5.4 useRealtimeUpdates Composable

```ts
// web/src/composables/useRealtimeUpdates.ts
export function useRealtimeUpdates() {
  const store = useDashboardStore()
  const status = ref<'connected' | 'reconnecting' | 'offline'>('reconnecting')
  let ws: WebSocket
  let retryDelay = 1000

  function connect() {
    ws = new WebSocket(`ws://localhost:${window.CCTRACK_PORT ?? 7432}/api/v1/ws`)

    ws.onopen = () => {
      status.value = 'connected'
      retryDelay = 1000
    }

    ws.onmessage = (msg) => {
      const event = JSON.parse(msg.data)
      store.applyEvent(event)  // Pinia action handles all event types
    }

    ws.onclose = () => {
      status.value = retryDelay >= 16000 ? 'offline' : 'reconnecting'
      setTimeout(() => { retryDelay = Math.min(retryDelay * 2, 30000); connect() }, retryDelay)
    }
  }

  onMounted(connect)
  onUnmounted(() => ws?.close())

  return { status }
}
```

---

## 6. State Management (Pinia)

One store per domain concern.

```
web/src/stores/
├── dashboard.ts    — summary totals, daily series, top sessions
├── sessions.ts     — full session list, pagination, sort state
└── settings.ts     — user config, dirty state, save status
```

### 6.1 Dashboard Store

```ts
interface DashboardState {
  summary: {
    today:     { cost: number; tokens: number }
    week:      { cost: number; tokens: number }
    month:     { cost: number; tokens: number }
    projected: number
  }
  daily: Array<{ date: string; cost: number }>
  topSessions: Session[]
  activeSessionId: string | null
  loaded: boolean
}
```

`applyEvent(event: WsEvent)` — the single entry point for all WebSocket messages:
- `summary.updated` → patch `state.summary` (triggers counter animation via watched refs)
- `session.updated` → patch session in `topSessions` if present; patch active session token count
- `session.created` → prepend to `topSessions`, set `activeSessionId`
- `ping` → no-op

### 6.2 Data Fetching

Vue Query (`@tanstack/vue-query`) for REST calls — handles loading states, caching, and background refetch on reconnect. Avoids manual `isLoading` / `error` state management in components.

```ts
// Overview page
const { data: summary } = useQuery({
  queryKey: ['summary'],
  queryFn: () => fetch('/api/v1/summary').then(r => r.json()),
  staleTime: 30_000,
})
```

On WebSocket reconnect, `queryClient.invalidateQueries()` is called to force a fresh REST sync before resuming live updates.

---

## 7. Routing

Vue Router 4. Hash mode to avoid needing server-side catch-all config (the Go server handles this, but hash mode is simpler for an embedded app).

```ts
const routes = [
  { path: '/',         component: Overview,  name: 'overview' },
  { path: '/sessions', component: Sessions,  name: 'sessions' },
  { path: '/settings', component: Settings,  name: 'settings' },
  { path: '/rates',    component: RateCard,  name: 'rates' },
]
```

Route transitions: `fade` — opacity 0→1 over 150ms. Keeps navigation feeling instant without jarring cuts.

---

## 8. Build Configuration

### 8.1 Vite Config

```ts
// web/vite.config.ts
export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    rollupOptions: {
      output: {
        // Deterministic filenames — Go embed doesn't need content hashing
        entryFileNames: 'assets/[name].js',
        chunkFileNames: 'assets/[name].js',
        assetFileNames: 'assets/[name].[ext]',
      }
    }
  },
  base: '/',
})
```

Deterministic filenames matter here — the Go binary embeds the `dist/` directory at compile time, so hashed filenames would change on every build and make debugging harder.

### 8.2 Dependencies

```json
{
  "dependencies": {
    "vue": "^3.4",
    "vue-router": "^4.3",
    "pinia": "^2.1",
    "@tanstack/vue-query": "^5.0",
    "chart.js": "^4.4",
    "vue-chartjs": "^5.3"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.0",
    "vite": "^5.0",
    "typescript": "^5.4",
    "vue-tsc": "^2.0"
  }
}
```

No UI component library. No icon library with 10,000 unused icons — use inline SVGs for the handful of icons needed (nav icons, sort arrows, status dots). Keeps the bundle lean.

### 8.3 Self-Hosted Fonts

Fonts live in `web/public/fonts/`. The CSS `@font-face` declarations reference them with relative paths. Vite copies `public/` contents to `dist/` verbatim, so Go's embed picks them up.

```
web/public/fonts/
├── BebasNeue-Regular.woff2
├── DMSans-Light.woff2
├── DMSans-Regular.woff2
├── DMSans-Medium.woff2
├── JetBrainsMono-Regular.woff2
└── JetBrainsMono-Medium.woff2
```

Total font payload target: under 300KB (woff2 compressed). Load order: CSS declares fonts, browser loads them in parallel with JS. No FOUT — dark background + dark text means font swap is invisible at load.

### 8.4 Bundle Size Target

| Asset | Target |
|---|---|
| JS (all chunks) | < 350KB gzipped |
| CSS | < 30KB gzipped |
| Fonts | < 300KB total |
| **Total** | **< 680KB** |

This keeps the Go binary under the 20MB target from the main PRD with plenty of headroom.

---

## 9. Accessibility

- All interactive elements keyboard-navigable (tab order follows visual order)
- Focus rings: amber at 40% opacity, 2px offset — visible but not garish
- ARIA labels on icon-only buttons (sort arrows, close panel)
- Table: `<thead>`, `scope="col"`, `aria-sort` on sorted column
- Connection status: `role="status"` with `aria-live="polite"`
- Colour is never the sole means of conveying information (status dots have text labels; budget states have text alongside colour)
- `prefers-reduced-motion` respected throughout (see §2.4)

---

## 10. Acceptance Criteria

| # | Criterion |
|---|---|
| 1 | All four hero stat cards display and animate counter tick-up on first load |
| 2 | Daily chart bars animate upward on load; today's bar is full amber |
| 3 | WebSocket connection status indicator updates within 1s of connect/disconnect |
| 4 | `summary.updated` WS event patches stat card values and re-triggers counter animation |
| 5 | Active session row pulses in the sessions table |
| 6 | Session detail slide-over opens/closes without page navigation |
| 7 | All fonts load from local binary — no external network requests |
| 8 | Budget progress bar appears under Today card when monthly budget is set |
| 9 | Table sorts correctly on all columns; sort state persists during the session |
| 10 | `prefers-reduced-motion` disables all animations without breaking layout |
| 11 | Full JS + CSS bundle under 350KB gzipped |
| 12 | App is functional and usable at 1280px, 1440px, and 1920px widths |
| 13 | No external network requests from the browser at any point |
| 14 | Keyboard navigation works for all interactive elements |

---

## 11. File Structure

```
web/
├── public/
│   ├── fonts/              — self-hosted woff2 files
│   └── favicon.ico
├── src/
│   ├── main.ts             — app entry, router + pinia install
│   ├── App.vue             — root component, AppShell slot
│   ├── assets/
│   │   ├── styles/
│   │   │   ├── tokens.css      — CSS custom properties
│   │   │   ├── typography.css  — @font-face + type classes
│   │   │   ├── animations.css  — @keyframes definitions
│   │   │   └── reset.css       — minimal box-model reset
│   ├── components/         — see §5.1
│   ├── composables/
│   │   ├── useCountUp.ts
│   │   ├── useRealtimeUpdates.ts
│   │   └── useFormatCost.ts    — $0.0000 / $0.00 / $1,234 formatting
│   ├── stores/             — see §6
│   ├── views/
│   │   ├── Overview.vue
│   │   ├── Sessions.vue
│   │   ├── Settings.vue
│   │   └── RateCard.vue
│   ├── router/
│   │   └── index.ts
│   └── types/
│       └── index.ts        — shared TS interfaces (Session, Summary, WsEvent, etc.)
├── index.html
├── vite.config.ts
├── tsconfig.json
└── package.json
```

---

## 12. Open Questions

| # | Question | Notes |
|---|---|---|
| 1 | Session naming — Claude Code logs use file paths or UUIDs as session identifiers. How do we derive a human-readable name (e.g. "refactor-auth") for display? | Parse working directory from log metadata if available; fall back to truncated session ID |
| 2 | Token count formatting — at what threshold do we switch from raw numbers to "1.2M"? | Suggest: < 10,000 show raw; ≥ 10,000 show abbreviated |
| 3 | Cost precision — show 2 decimal places ($0.84) or 4 ($0.8423)? | 2dp for totals, 4dp for per-session breakdown where sub-cent precision matters |
| 4 | Chart.js vs a lighter alternative (uPlot) — Chart.js is ~60KB gzipped. Worth evaluating if bundle size becomes a concern. | Stick with Chart.js for MVP; revisit if bundle target is threatened |
| 5 | Dark mode only or light mode option in Settings? | Dark only for MVP — the design is built for dark; light would require a full re-skin |
