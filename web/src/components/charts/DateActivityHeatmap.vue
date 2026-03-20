<template>
  <div class="heatmap-card">
    <div class="chart-header">
      <div class="chart-title">{{ title }}</div>
    </div>

    <!-- 30-day variant: single horizontal row with date labels above -->
    <div v-if="variant === '30day'" class="heatmap-30">
      <div class="heatmap-30-labels">
        <div v-for="day in days30" :key="'l'+day.date" class="heatmap-30-label">
          {{ day.monthLabel }}<br>{{ day.dayLabel }}
        </div>
      </div>
      <div class="heatmap-30-row">
        <div
          v-for="day in days30"
          :key="'c'+day.date"
          class="heatmap-30-cell"
          :class="colorClass(day.cost)"
          :title="`${day.tooltipDate}: ${formatCost(day.cost)}`"
        ></div>
      </div>
    </div>

    <!-- 365-day variant: GitHub contribution graph layout -->
    <div v-if="variant === '365day'" class="heatmap-365-wrap">
      <!-- Left row labels -->
      <div class="heatmap-365-labels-left">
        <div v-for="label in dayLabels" :key="'ll'+label" class="heatmap-365-row-label">{{ label }}</div>
      </div>
      <!-- Grid area: month labels on top, cells below -->
      <div class="heatmap-365-grid-area">
        <div class="heatmap-365-months" :style="{ gridTemplateColumns: `repeat(${numCols}, 1fr)`, gap: '2px' }">
          <span
            v-for="ml in monthLabels365"
            :key="ml.col"
            class="heatmap-365-month-label"
            :style="{ gridColumn: ml.col + 1 }"
          >{{ ml.label }}</span>
        </div>
        <div class="heatmap-365-grid">
          <div
            v-for="cell in grid365"
            :key="cell.date"
            class="heatmap-365-cell"
            :class="cell.cls"
            :title="cell.tooltip"
          ></div>
        </div>
      </div>
      <!-- Right row labels -->
      <div class="heatmap-365-labels-right">
        <div v-for="label in dayLabels" :key="'rl'+label" class="heatmap-365-row-label">{{ label }}</div>
      </div>
    </div>

    <!-- Legend (30-day only) -->
    <div v-if="variant === '30day'" class="legend">
      <span class="legend-label">Less</span>
      <div class="legend-swatch c0"></div>
      <div class="legend-swatch c1"></div>
      <div class="legend-swatch c2"></div>
      <div class="legend-swatch c4"></div>
      <div class="legend-swatch c5"></div>
      <span class="legend-label">More</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { DailyHeatmapCell } from '../../types'
import { formatCostDisplay } from '../../composables/useFormatCost'

const props = defineProps<{
  cells: DailyHeatmapCell[]
  title: string
  variant: '30day' | '365day'
}>()

const monthNames = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec']
const dayLabels = ['M','T','W','Th','F','Sa','S']

// Build a cost lookup map from cells
const costMap = computed(() => {
  const m = new Map<string, number>()
  for (const c of props.cells) {
    m.set(c.date, c.cost)
  }
  return m
})

// Max cost for intensity scaling
const maxCost = computed(() => {
  let max = 0
  for (const c of props.cells) {
    if (c.cost > max) max = c.cost
  }
  return max || 1
})

// Map cost to a color class (c0-c5)
function colorClass(cost: number): string {
  if (cost === 0) return 'c0'
  const intensity = cost / maxCost.value
  if (intensity < 0.15) return 'c1'
  if (intensity < 0.35) return 'c2'
  if (intensity < 0.55) return 'c3'
  if (intensity < 0.75) return 'c4'
  return 'c5'
}

function formatCost(cost: number): string {
  return formatCostDisplay(cost)
}

// Format date to "YYYY-MM-DD"
function formatDateISO(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

// ── 30-day variant data ──
const days30 = computed(() => {
  const today = new Date()
  const result: { date: string; monthLabel: string; dayLabel: string; cost: number; tooltipDate: string }[] = []
  // 31 boxes: today - 30 through today
  for (let i = 30; i >= 0; i--) {
    const d = new Date(today)
    d.setDate(today.getDate() - i)
    const dateStr = formatDateISO(d)
    const cost = costMap.value.get(dateStr) || 0
    result.push({
      date: dateStr,
      monthLabel: monthNames[d.getMonth()],
      dayLabel: String(d.getDate()),
      cost,
      tooltipDate: `${monthNames[d.getMonth()]} ${d.getDate()}`,
    })
  }
  return result
})

// ── 365-day variant data ──

// Grid cells ordered column-by-column (week by week, Mon-Sun per column)
// The CSS grid uses grid-auto-flow: column with 7 rows
const grid365 = computed(() => {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const todayStr = formatDateISO(today)

  // Find the start: the Monday on or before (today - 364)
  const rawStart = new Date(today)
  rawStart.setDate(today.getDate() - 364)
  const startDow = rawStart.getDay() // 0=Sun, 1=Mon ... 6=Sat
  const mondayOffset = startDow === 0 ? -6 : 1 - startDow
  rawStart.setDate(rawStart.getDate() + mondayOffset)

  // The 365-day window boundary
  const windowStart = new Date(today)
  windowStart.setDate(today.getDate() - 364)
  windowStart.setHours(0, 0, 0, 0)

  const cells: { date: string; cls: string; tooltip: string }[] = []
  const current = new Date(rawStart)

  while (true) {
    // Process 7 days (Mon-Sun) for this column
    for (let d = 0; d < 7; d++) {
      const dateStr = formatDateISO(current)
      const cost = costMap.value.get(dateStr) || 0

      if (current > today) {
        // Future day — dashed outline
        cells.push({ date: dateStr, cls: 'future', tooltip: '' })
      } else if (current < windowStart) {
        // Before 365-day window — empty
        cells.push({ date: dateStr, cls: 'c0', tooltip: '' })
      } else {
        const tooltipStr = `${monthNames[current.getMonth()]} ${current.getDate()}: ${formatCostDisplay(cost)}`
        cells.push({ date: dateStr, cls: colorClass(cost), tooltip: tooltipStr })
      }
      current.setDate(current.getDate() + 1)
    }

    // If we've completed the week that contains today, stop
    // The current date is now the Monday after the last Sunday we processed
    if (current > today && current.getDay() === 1) break
  }

  return cells
})

// Number of week columns in the 365-day grid
const numCols = computed(() => Math.ceil(grid365.value.length / 7))

// Month labels positioned at the column where a month starts
const monthLabels365 = computed(() => {
  const today = new Date()
  today.setHours(0, 0, 0, 0)

  const rawStart = new Date(today)
  rawStart.setDate(today.getDate() - 364)
  const startDow = rawStart.getDay()
  const mondayOffset = startDow === 0 ? -6 : 1 - startDow
  rawStart.setDate(rawStart.getDate() + mondayOffset)

  const labels: { col: number; label: string }[] = []
  let prevMonth = -1
  let col = 0
  const current = new Date(rawStart)

  while (true) {
    const weekStart = new Date(current)

    // Check if any day in this week is the 1st-7th of a new month
    for (let d = 0; d < 7; d++) {
      const check = new Date(weekStart)
      check.setDate(weekStart.getDate() + d)
      if (check.getDate() <= 7 && check.getMonth() !== prevMonth) {
        labels.push({ col, label: monthNames[check.getMonth()] })
        prevMonth = check.getMonth()
        break
      }
    }

    col++
    current.setDate(current.getDate() + 7)

    if (weekStart > today) break
  }

  return labels
})
</script>

<style scoped>
.heatmap-card {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: 12px;
  padding: var(--space-6);
  animation: fadeSlideUp 0.45s ease both;
  animation-delay: 450ms;
  display: flex;
  flex-direction: column;
}
.chart-header {
  margin-bottom: var(--space-4);
}
.chart-title {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--text-tertiary);
}

/* === 30-DAY HEATMAP === */
.heatmap-30 {
  display: flex;
  flex-direction: column;
}
.heatmap-30-labels {
  display: grid;
  grid-template-columns: repeat(31, 1fr);
  gap: 2px;
}
.heatmap-30-label {
  text-align: center;
  font-family: 'JetBrains Mono', monospace;
  font-size: 8px;
  line-height: 1.3;
  color: var(--text-disabled);
  margin-bottom: 4px;
}
.heatmap-30-row {
  display: grid;
  grid-template-columns: repeat(31, 1fr);
  gap: 2px;
}
.heatmap-30-cell {
  aspect-ratio: 1;
  border-radius: 3px;
  transition: outline 150ms;
  cursor: default;
}
.heatmap-30-cell:hover {
  outline: 1px solid #f59e0b;
  outline-offset: -1px;
}

/* === 365-DAY HEATMAP === */
.heatmap-365-wrap {
  display: flex;
  align-items: stretch;
  gap: 0;
}
.heatmap-365-labels-left,
.heatmap-365-labels-right {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding-top: 18px; /* align with cells below month labels */
}
.heatmap-365-labels-left { padding-right: 6px; }
.heatmap-365-labels-right { padding-left: 6px; }
.heatmap-365-row-label {
  flex: 1; /* match grid row heights */
  font-family: 'JetBrains Mono', monospace;
  font-size: 9px;
  color: var(--text-disabled);
  display: flex;
  align-items: center;
  line-height: 1;
}
.heatmap-365-labels-left .heatmap-365-row-label { justify-content: flex-end; }
.heatmap-365-labels-right .heatmap-365-row-label { justify-content: flex-start; }
.heatmap-365-grid-area {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.heatmap-365-months {
  display: grid;
  height: 16px;
  margin-bottom: 2px;
}
.heatmap-365-month-label {
  font-family: 'JetBrains Mono', monospace;
  font-size: 9px;
  color: var(--text-disabled);
}
.heatmap-365-grid {
  display: grid;
  grid-template-rows: repeat(7, 1fr);
  grid-auto-flow: column;
  grid-auto-columns: 1fr;
  gap: 2px;
}
.heatmap-365-cell {
  aspect-ratio: 1;
  border-radius: 2px;
  cursor: default;
}
.heatmap-365-cell:hover {
  outline: 1px solid #f59e0b;
  outline-offset: -1px;
}
/* Future days (after today) */
.heatmap-365-cell.future {
  background: transparent !important;
  border: 1px dashed var(--border-subtle);
}

/* === Shared color levels === */
.c0 { background: var(--bg-subtle); }
.c1 { background: rgba(245,158,11,0.10); }
.c2 { background: rgba(245,158,11,0.22); }
.c3 { background: rgba(245,158,11,0.38); }
.c4 { background: rgba(245,158,11,0.58); }
.c5 { background: #f59e0b; }

/* === Legend === */
.legend {
  display: flex;
  align-items: center;
  gap: 3px;
  justify-content: flex-end;
  margin-top: var(--space-3);
}
.legend-label {
  font-family: 'JetBrains Mono', monospace;
  font-size: 9px;
  color: var(--text-disabled);
  padding: 0 3px;
}
.legend-swatch {
  width: 10px;
  height: 10px;
  border-radius: 2px;
}
</style>
