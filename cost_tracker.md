# PRD: Claude Code Cost Tracker (`cctrack`)

**Version:** 0.2 — MVP  
**Status:** Draft  
**Owner:** TBD  
**Last Updated:** 2026-02-22

---

## 1. Overview

`cctrack` is a self-contained CLI tool and local web dashboard that gives Claude Code users real-time visibility into their token usage and associated costs. It ships as a single Go binary with an embedded Vue SPA, distributed via Homebrew and a curl install script.

The core insight driving this product: people searching "Claude Code pricing" don't want the rate card — they want to know what they're actually spending. No existing tool answers this. `cctrack` does.

---

## 2. Problem Statement

Claude Code users have no native way to understand where their spend is coming from. The Anthropic console shows aggregate billing but gives no session-level granularity. Power users, teams, and anyone running long agentic loops have no answer to:

- "Which sessions are costing me the most?"
- "Am I on track to hit my budget this month?"
- "How much of my spend is cached vs. fresh tokens?"

This gap creates search demand ("claude code pricing", "claude code cost", "claude code usage" — ~6.6K combined monthly volume) and a real pain point that a lightweight local tool can own.

---

## 3. Goals

### MVP (v1.0)
- Parse local Claude Code session logs and extract token usage
- Calculate real costs using Anthropic's published rates
- Serve a local web dashboard (Vue SPA, embedded in binary) showing spend summary, daily chart, and top sessions
- Expose a `cctrack status` CLI command for quick terminal output
- Distribute as a single binary via Homebrew tap and curl install script

### V2
- Anthropic API billing integration (when the API exposes it)
- Cost alerts via threshold configuration
- Cloud sync for persistent history
- Email/webhook alert delivery

### Out of Scope (MVP)
- Multi-user / team dashboards
- Session comparison views
- Any remote data storage or account system

---

## 4. Target Users

**Primary:** Individual Claude Code power users — developers, AI engineers, consultants — who use Claude Code heavily and want spend awareness without leaving their terminal workflow.

**Secondary:** Technical leads or solo founders evaluating Claude Code for team adoption who need cost predictability before committing.

---

## 5. Architecture

### 5.1 Technology Stack

| Layer | Choice | Rationale |
|---|---|---|
| Language | Go 1.22+ | Single binary, fast startup, `embed.FS` support, strong CLI tooling |
| Frontend | Vue 3 + Vite | Preferred stack; Vite produces a clean `dist/` that embeds trivially |
| Embedding | `go:embed` | Vue `dist/` embedded into binary at compile time; zero runtime deps |
| HTTP server | Go `net/http` (stdlib) | Serves embedded SPA + JSON API endpoints + WebSocket upgrades |
| WebSocket | `github.com/coder/websocket` | Lightweight, stdlib-compatible; used for real-time push to dashboard |
| File watching | `github.com/fsnotify/fsnotify` | Cross-platform inotify/kqueue wrapper; watches log directory for new writes |
| Storage | SQLite via `modernc.org/sqlite` (pure Go) | No CGo dependency; stores parsed session data locally |
| CLI | `cobra` | Standard Go CLI framework |
| Charts | Chart.js (bundled in Vue build) | Lightweight, no CDN dependency at runtime |

### 5.2 Binary Structure

```
cctrack (single binary)
├── cmd/
│   ├── root.go           — root cobra command
│   ├── status.go         — `cctrack status` CLI output
│   ├── serve.go          — `cctrack serve` starts dashboard server
│   └── parse.go          — `cctrack parse` manual log ingestion
├── internal/
│   ├── parser/           — Claude Code log parser
│   ├── calculator/       — cost calculation engine
│   ├── store/            — SQLite persistence layer
│   ├── watcher/          — fsnotify wrapper + debounce buffer
│   ├── hub/              — WebSocket hub (fan-out broadcaster)
│   └── api/              — JSON API handlers + WebSocket upgrade handler
├── web/                  — Vue SPA source (compiled to dist/ at build time)
│   └── dist/             — embedded via go:embed
└── main.go
```

### 5.3 Embedding Pattern

```go
//go:embed web/dist
var webFS embed.FS

// Serve SPA for all non-API routes
http.Handle("/", http.FileServer(http.FS(webFS)))

// API routes prefixed /api/v1/
http.HandleFunc("/api/v1/summary", handleSummary)
```

The Vue app is built with `vite build` and output goes into `web/dist/`. The Go build then embeds this directory. The result is a single binary that opens a browser to `http://localhost:7432` when `cctrack serve` is run.

### 5.4 Data Flow

```
Claude Code session logs (JSONL)
        ↓
    watcher/        — fsnotify watches ~/.claude/projects/ for file writes
        ↓
    debounce        — 250ms buffer groups rapid sequential writes into one event
        ↓
    parser/         — reads new lines only (tail from last byte offset); extracts session ID, timestamps, token counts
        ↓
    calculator/     — applies published Anthropic rates → USD cost
        ↓
    store/ (SQLite) — persists parsed sessions (~/.cctrack/cctrack.db)
        ↓
    hub/            — WebSocket fan-out broadcaster; notifies all connected clients
        ↓
    Vue dashboard   — receives push event, re-fetches affected data via REST or applies patch from WS message
```

**Initial load** (browser open / refresh) uses the REST API to hydrate all state. The WebSocket connection is then established and maintained for incremental live updates only. This keeps the REST API as the source of truth and the WebSocket layer purely additive.

---

## 6. Log Parser

### 6.1 Source

Claude Code writes session logs to `~/.claude/projects/` as JSONL files. Each line is a message event containing role, token counts, and model metadata.

### 6.2 Fields to Extract

| Field | Source | Notes |
|---|---|---|
| `session_id` | File path / metadata | Unique per session |
| `timestamp` | Message event | ISO8601 |
| `model` | Message metadata | e.g. `claude-sonnet-4-6` |
| `input_tokens` | Usage object | Fresh input tokens |
| `output_tokens` | Usage object | |
| `cache_read_tokens` | Usage object | Cheaper rate |
| `cache_write_tokens` | Usage object | Slightly above input rate |

### 6.3 Parsing Strategy

**Startup (cold parse)**
- Scan `~/.claude/projects/` recursively for all JSONL log files
- For each file, check the DB for the last processed byte offset
- Parse from that offset to EOF; store the new offset after each successful parse
- Idempotent — re-processing the same byte range produces no duplicate records

**Real-time watching**
- `fsnotify` watcher registered on the log directory (recursive on supported platforms; polling fallback on others)
- On `WRITE` or `CREATE` event for a `.jsonl` file, a debounce timer is reset
- **Debounce window: 250ms** — chosen to group the rapid burst of writes Claude Code emits during an active response (tool calls, streaming chunks) into a single parse cycle without introducing noticeable lag
- After the debounce window elapses with no further events for that file, the parser reads from the last stored byte offset to the current EOF
- Parsed events are written to SQLite, then the hub broadcasts a WebSocket update to all connected clients

**Byte-offset tracking**
- Per-file offset stored in SQLite (`file_offsets` table: `path`, `offset`, `updated_at`)
- Protects against re-parsing on restart; also handles log rotation gracefully (new file = zero offset)

**Edge cases**
- File truncation detected by offset > file size → reset offset to 0, re-parse full file
- Incomplete final line (mid-write) → skip and retain current offset; next debounce cycle will pick it up

---

## 7. Cost Calculator

Rates stored as a versioned config struct in the binary, updated each release. No API call required.

```go
type ModelRates struct {
    Model            string
    InputPerMToken   float64
    OutputPerMToken  float64
    CacheReadPerMToken  float64
    CacheWritePerMToken float64
}
```

Initial rates at time of MVP (to be confirmed at build time):

| Model | Input | Output | Cache Read | Cache Write |
|---|---|---|---|---|
| claude-opus-4-6 | $15.00 | $75.00 | $1.50 | $18.75 |
| claude-sonnet-4-6 | $3.00 | $15.00 | $0.30 | $3.75 |
| claude-haiku-4-5 | $0.80 | $4.00 | $0.08 | $1.00 |

*Per million tokens. Rates are per Anthropic's published pricing; will be kept current in releases.*

---

## 8. Real-Time WebSocket Layer

### 8.1 Overview

The WebSocket connection provides live push updates to the dashboard without polling. It is a fan-out broadcaster: one server-side hub accepts connections from N browser tabs and pushes the same event payload to all of them.

### 8.2 Hub Design

```go
// internal/hub/hub.go

type Hub struct {
    clients   map[*Client]struct{}
    broadcast chan Event
    register  chan *Client
    unregister chan *Client
    mu        sync.RWMutex
}

type Event struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"`
}
```

The hub runs in a single goroutine and serialises all register/unregister/broadcast operations to avoid locking overhead on the hot broadcast path.

### 8.3 Event Types

| Event Type | Triggered By | Payload |
|---|---|---|
| `session.updated` | New tokens parsed for an existing session | `{ session_id, delta_cost, delta_tokens, running_total }` |
| `session.created` | New session file detected | `{ session_id, started_at, model }` |
| `summary.updated` | Any parse cycle that changes totals | `{ today, week, month, projected }` |
| `ping` | Every 30s (keepalive) | `{}` |

Keeping payloads small — the Vue client re-fetches full detail from REST only when it needs to render a drill-down view. The WebSocket events carry just enough data to update running counters and charts in place.

### 8.4 WebSocket Endpoint

```
GET /api/v1/ws
```

Upgraded via stdlib `net/http` + `coder/websocket`. No auth required (local-only server). Connection lifecycle:

1. Client connects → hub registers client
2. Server sends current `summary.updated` snapshot immediately on connect (so a freshly opened tab is in sync)
3. Server sends events as they arrive from the watcher/parser pipeline
4. Client disconnects → hub unregisters, goroutine exits cleanly

### 8.5 Vue Client Integration

The Vue app uses a composable `useRealtimeUpdates()` that wraps a native `WebSocket`:

```js
// web/src/composables/useRealtimeUpdates.js

export function useRealtimeUpdates(onEvent) {
  const ws = ref(null)

  function connect() {
    ws.value = new WebSocket(`ws://localhost:7432/api/v1/ws`)
    ws.value.onmessage = (msg) => {
      const event = JSON.parse(msg.data)
      onEvent(event)
    }
    ws.value.onclose = () => {
      // Reconnect with exponential backoff (max 30s)
      setTimeout(connect, Math.min(retryDelay * 2, 30000))
    }
  }

  onMounted(connect)
  onUnmounted(() => ws.value?.close())
}
```

Stores (Pinia) handle incoming events and apply patches directly to reactive state — no full re-fetch needed for `session.updated` and `summary.updated` events.

### 8.6 Reconnection Behaviour

- Exponential backoff starting at 1s, capped at 30s
- Visual indicator in dashboard header: green dot (connected) / amber dot (reconnecting) / red dot (disconnected >30s)
- On reconnect, client calls REST `/api/v1/summary` to re-sync before resuming WebSocket updates

---

## 9. Dashboard (Vue SPA)

### 9.1 Views

**Home / Summary**
- Total spend: today / this week / this month (card strip) — updated live via WebSocket
- Projected monthly cost based on current-month run-rate
- Daily spend bar chart (last 30 days) — Chart.js, appends to current day's bar in real time
- Token breakdown donut: input / output / cache read / cache write
- Live connection status indicator (green / amber / red dot, top-right)

**Sessions**
- Table of all sessions, sortable by cost / date / token count
- Top 5 most expensive sessions highlighted
- Per-session detail: token breakdown, model used, duration, cost
- Active session row pulses to indicate live token ingestion

**Settings**
- Log directory path override
- Monthly budget threshold (triggers a visual warning, not an alert in MVP)
- Rate card view (read-only, shows current rates in use)

### 9.2 API Endpoints

All prefixed `/api/v1/`:

| Endpoint | Method | Description |
|---|---|---|
| `/summary` | GET | Spend totals, projections, token breakdown |
| `/sessions` | GET | Paginated session list with cost data |
| `/sessions/:id` | GET | Single session detail |
| `/daily` | GET | Daily spend timeseries (last 30 days) |
| `/settings` | GET/POST | Read/write user config |
| `/ws` | GET (WS upgrade) | Real-time event stream |

### 9.3 Design Principles

- Dark mode default (developer audience)
- Zero external network calls from the browser — all data from local API
- Responsive but desktop-first (dashboard will mostly be viewed on dev machines)
- No telemetry, no analytics, no phone-home behaviour — must be explicit in README
- REST for initial hydration; WebSocket for incremental live updates only

---

## 10. CLI Interface

### Commands

```
cctrack serve          — parse logs, start dashboard, open browser
cctrack status         — print summary to stdout (today / week / month spend)
cctrack parse          — manually trigger log parsing (useful in scripts)
cctrack config         — open/show config file path
cctrack version        — show binary version + rate card version
```

### `cctrack status` Output Example

```
cctrack v1.0.0
─────────────────────────────────────
Today          $0.84   (1.2M tokens)
This week      $6.12   (8.9M tokens)
This month    $24.37  (34.1M tokens)
Projected     $31.20
─────────────────────────────────────
Top session today: "refactor-auth" — $0.43
```

---

## 11. Configuration

Stored at `~/.cctrack/config.json`:

```json
{
  "log_dir": "~/.claude/projects",
  "db_path": "~/.cctrack/cctrack.db",
  "port": 7432,
  "monthly_budget_usd": 0,
  "open_browser_on_serve": true
}
```

---

## 12. Distribution

### Homebrew Tap

```
brew tap <org>/cctrack
brew install cctrack
```

Tap repo maintained separately. Formula points to GitHub release artifacts (darwin-arm64, darwin-amd64, linux-amd64, linux-arm64).

### Curl Install

```bash
curl -fsSL https://cctrack.dev/install.sh | bash
```

Script detects OS/arch, downloads the correct binary from GitHub releases, places in `/usr/local/bin`.

### Build Pipeline

- GitHub Actions on tag push
- `go build` with `GOOS`/`GOARCH` matrix
- Vue built first (`vite build`), output committed to `web/dist/` or built in CI before Go build
- Binary signed (macOS notarisation for Gatekeeper compliance)
- Checksums published alongside releases

---

## 13. Subscriber Gate

The tool is distributed as a lead magnet / subscriber-only download. The gate is **distribution-level, not feature-level** — the binary itself has no licence check or phone-home. Subscribers get a download link; the binary works fully offline.

This keeps the architecture clean and trust high (no telemetry, no activation), while the subscriber gate is handled entirely by the content/email platform.

---

## 14. MVP Acceptance Criteria

| # | Criterion |
|---|---|
| 1 | Binary parses Claude Code JSONL logs from `~/.claude/projects/` without configuration |
| 2 | Cost calculation matches manual calculation using published Anthropic rates |
| 3 | `cctrack serve` opens browser dashboard within 2 seconds on a cold start |
| 4 | `cctrack status` prints spend summary in under 500ms |
| 5 | Dashboard displays: total spend (today/week/month), daily chart, top 5 sessions |
| 6 | Re-parsing the same logs produces identical results (idempotent) |
| 7 | Binary size under 20MB (Vue build + Go binary) |
| 8 | Works on macOS (arm64 + amd64) and Linux (amd64 + arm64) with no runtime dependencies |
| 9 | No network calls at runtime — fully air-gapped capable |
| 10 | README clearly states no telemetry / data collection |
| 11 | New token events written to a log file appear in the dashboard within 500ms of the debounce window closing |
| 12 | WebSocket reconnects automatically after server restart or network blip without user action |
| 13 | Multiple browser tabs connected simultaneously all receive the same events |
| 14 | Dashboard summary counters update in place without a full page reload when a WebSocket event arrives |

---

## 15. Open Questions

| # | Question | Owner | Notes |
|---|---|---|---|
| 1 | What is the exact JSONL schema for Claude Code session logs? Needs verification against actual log output before parser is written. | Dev | Inspect `~/.claude/projects/` on a machine with Claude Code installed |
| 2 | Does Anthropic expose any usage API today that could supplement log parsing? | Dev | Check docs.anthropic.com — V2 feature if available |
| 3 | Port 7432 — any conflicts? Should this be configurable only or also auto-detect a free port? | Dev | Make configurable, default 7432 |
| 4 | Subscriber gate mechanism — email platform, Gumroad, or custom? | Product | Affects download link flow but not binary design |
| 5 | Rate card update cadence — ship a new binary release every time Anthropic changes pricing, or pull rates from a hosted JSON? | Dev | Hosted JSON adds a network call but keeps rates fresh without a release |

---

## 16. Milestones

| Milestone | Scope | Target |
|---|---|---|
| M0 — Spike | Validate log schema, confirm token field names, prototype parser | 1 week |
| M1 — Core | Parser + debounce watcher + calculator + SQLite store + `cctrack status` CLI | 2 weeks |
| M2 — Dashboard | Vue SPA embedded, summary + sessions views, serve command | 2 weeks |
| M3 — Real-time | WebSocket hub, Vue composable, live counter updates, connection indicator | 1 week |
| M4 — Distribution | Homebrew tap, curl installer, CI/CD pipeline, macOS signing | 1 week |
| M5 — Polish | README, no-telemetry statement, settings view, subscriber delivery | 1 week |
| **MVP Total** | | **~8 weeks** |
