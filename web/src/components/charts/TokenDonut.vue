<template>
  <div class="donut-card">
    <div class="chart-header">
      <div class="chart-title">Cost Breakdown</div>
    </div>
    <div class="donut-wrap">
      <Doughnut v-if="chartData" :data="chartData" :options="chartOptions" />
    </div>
    <div class="donut-legend">
      <div v-for="(item, i) in legendItems" :key="i" class="legend-row">
        <div class="legend-left">
          <div class="legend-dot" :style="{ background: item.color }"></div>
          <span>{{ item.label }}</span>
        </div>
        <div class="legend-val">{{ formatCostDisplay(item.value) }} <span class="legend-pct">{{ item.pct }}%</span></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Doughnut } from 'vue-chartjs'
import {
  Chart as ChartJS,
  ArcElement,
  Tooltip,
} from 'chart.js'
import { formatCostDisplay } from '../../composables/useFormatCost'

ChartJS.register(ArcElement, Tooltip)

const props = defineProps<{
  inputCost: number
  outputCost: number
  cacheReadCost: number
  cacheWriteCost: number
}>()

const segments = computed(() => [
  { label: 'Input', value: props.inputCost, color: '#f59e0b' },
  { label: 'Output', value: props.outputCost, color: '#fbbf24' },
  { label: 'Cache Read', value: props.cacheReadCost, color: '#78716c' },
  { label: 'Cache Write', value: props.cacheWriteCost, color: '#44403c' },
])

const total = computed(() =>
  props.inputCost + props.outputCost + props.cacheReadCost + props.cacheWriteCost
)

const legendItems = computed(() =>
  segments.value.map(s => ({
    ...s,
    pct: total.value > 0 ? Math.round((s.value / total.value) * 100) : 0,
  }))
)

const chartData = computed(() => ({
  labels: segments.value.map(s => s.label),
  datasets: [{
    data: segments.value.map(s => s.value),
    backgroundColor: segments.value.map(s => s.color),
    borderColor: '#0a0a09',
    borderWidth: 3,
    hoverBorderColor: '#0a0a09',
  }],
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  cutout: '72%',
  animation: { duration: 800, easing: 'easeOutQuart' as const },
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: '#1a1a18',
      borderColor: '#2a2a27',
      borderWidth: 1,
      titleColor: '#8c8a84',
      bodyColor: '#f0ede8',
      bodyFont: { family: 'JetBrains Mono', size: 12 },
      titleFont: { family: 'DM Sans', size: 11 },
      padding: 10,
      callbacks: {
        label: (ctx: any) => {
          const t = ctx.dataset.data.reduce((a: number, b: number) => a + b, 0)
          const pct = t > 0 ? Math.round((ctx.parsed / t) * 100) : 0
          return `  $${ctx.parsed.toFixed(2)} (${pct}%)`
        },
      },
    },
  },
}
</script>

<style scoped>
.donut-card {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  padding: var(--space-6);
  animation: fadeSlideUp 0.45s ease both;
  animation-delay: 320ms;
  display: flex;
  flex-direction: column;
}
.chart-header {
  margin-bottom: var(--space-5);
}
.chart-title {
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--text-tertiary);
}
.donut-wrap {
  height: 150px;
  position: relative;
  margin-bottom: var(--space-5);
}
.donut-legend {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}
.legend-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 12px;
}
.legend-left {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  color: var(--text-secondary);
}
.legend-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}
.legend-val {
  font-family: 'JetBrains Mono', monospace;
  font-size: 11px;
  color: var(--text-tertiary);
}
.legend-pct {
  color: var(--text-disabled);
  font-size: 10px;
}
</style>
