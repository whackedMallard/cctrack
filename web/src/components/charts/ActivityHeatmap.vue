<template>
  <div class="heatmap-card">
    <div class="chart-header">
      <div class="chart-title">Activity Heatmap — Last 7 Days</div>
    </div>
    <div class="heatmap-grid">
      <!-- Hour labels across top -->
      <div class="heatmap-corner"></div>
      <div v-for="h in hourLabels" :key="'h'+h.hour" class="hour-label">{{ h.label }}</div>

      <!-- Rows: one per day, oldest at top, today at bottom -->
      <template v-for="row in dateRows" :key="row.date">
        <div class="day-label">{{ row.label }}</div>
        <div
          v-for="h in 24"
          :key="row.date + '-' + (h-1)"
          class="heatmap-cell"
          :class="{ future: row.date === todayDate && (h-1) > currentHour }"
          :style="{ background: cellColor(row.date, h - 1) }"
          :title="cellTooltip(row.date, row.tooltipLabel, h - 1)"
        ></div>
      </template>
    </div>
    <div class="heatmap-legend">
      <span class="legend-label">Less</span>
      <div class="legend-swatch" style="background: var(--bg-subtle)"></div>
      <div class="legend-swatch" style="background: rgba(245,158,11,0.15)"></div>
      <div class="legend-swatch" style="background: rgba(245,158,11,0.35)"></div>
      <div class="legend-swatch" style="background: rgba(245,158,11,0.6)"></div>
      <div class="legend-swatch" style="background: #f59e0b"></div>
      <span class="legend-label">More</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { DateHeatmapCell } from '../../types'
import { formatCostDisplay } from '../../composables/useFormatCost'

const props = defineProps<{ cells: DateHeatmapCell[] }>()

const dayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

// Track today's date and current hour for future-cell styling
const todayDate = formatDateISO(new Date())
const currentHour = new Date().getHours()

const hourLabels = computed(() => {
  const labels = []
  for (let h = 0; h < 24; h++) {
    labels.push({
      hour: h,
      label: h % 3 === 0 ? (h === 0 ? '12a' : h < 12 ? `${h}a` : h === 12 ? '12p' : `${h-12}p`) : '',
    })
  }
  return labels
})

// Format date to "YYYY-MM-DD"
function formatDateISO(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

// Generate 7 rows: 6 days ago at top, today at bottom
const dateRows = computed(() => {
  const today = new Date()
  const rows: { date: string; label: string; tooltipLabel: string }[] = []
  for (let i = 6; i >= 0; i--) {
    const d = new Date(today)
    d.setDate(today.getDate() - i)
    const dateStr = formatDateISO(d)
    const dayName = dayNames[d.getDay()]
    let label: string
    let tooltipLabel: string
    if (i === 0) {
      label = 'Today'
      tooltipLabel = 'Today'
    } else if (i === 1) {
      label = dayName
      tooltipLabel = 'Yesterday'
    } else {
      label = dayName
      tooltipLabel = dayName
    }
    rows.push({ date: dateStr, label, tooltipLabel })
  }
  return rows
})

// Build a map of {date-hour -> cost} for quick lookup
const cellMap = computed(() => {
  const m = new Map<string, number>()
  for (const c of props.cells) {
    m.set(`${c.date}-${c.hour}`, c.cost)
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

function cellColor(date: string, hour: number): string {
  const cost = cellMap.value.get(`${date}-${hour}`) || 0
  if (cost === 0) return 'var(--bg-subtle)'
  const intensity = cost / maxCost.value
  if (intensity < 0.15) return 'rgba(245,158,11,0.10)'
  if (intensity < 0.35) return 'rgba(245,158,11,0.22)'
  if (intensity < 0.55) return 'rgba(245,158,11,0.38)'
  if (intensity < 0.75) return 'rgba(245,158,11,0.58)'
  return '#f59e0b'
}

function cellTooltip(date: string, dayLabel: string, hour: number): string {
  const cost = cellMap.value.get(`${date}-${hour}`) || 0
  const hourStr = hour === 0 ? '12am' : hour < 12 ? `${hour}am` : hour === 12 ? '12pm' : `${hour-12}pm`
  return `${dayLabel} ${hourStr}: ${formatCostDisplay(cost)}`
}
</script>

<style scoped>
.heatmap-card {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
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
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--text-tertiary);
}
.heatmap-grid {
  display: grid;
  grid-template-columns: 36px repeat(24, 1fr);
  gap: 2px;
}
.heatmap-corner {
  /* empty top-left corner */
}
.hour-label {
  font-family: 'JetBrains Mono', monospace;
  font-size: 11px;
  color: var(--text-tertiary);
  text-align: center;
  padding-bottom: 2px;
}
.day-label {
  font-family: 'JetBrains Mono', monospace;
  font-size: 11px;
  color: var(--text-tertiary);
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding-right: 4px;
}
.heatmap-cell {
  height: 24px;
  transition: background 300ms;
  cursor: default;
}
.heatmap-cell:hover {
  outline: 1px solid var(--amber-500);
  outline-offset: -1px;
  z-index: 1;
}
/* Future hours (after current hour today) */
.heatmap-cell.future {
  background: transparent !important;
  border: 1px dashed var(--border-subtle);
}
.heatmap-legend {
  display: flex;
  align-items: center;
  gap: 3px;
  justify-content: flex-end;
  margin-top: var(--space-3);
}
.legend-label {
  font-family: 'JetBrains Mono', monospace;
  font-size: 11px;
  color: var(--text-tertiary);
  padding: 0 3px;
}
.legend-swatch {
  width: 10px;
  height: 10px;
}
</style>
