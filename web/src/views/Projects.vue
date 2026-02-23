<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">Projects</h1>
      <div class="page-meta">{{ projects.length }} projects</div>
    </div>

    <!-- Cost by project bar chart -->
    <div class="charts-row" v-if="projects.length">
      <div class="chart-card">
        <div class="chart-header">
          <div class="chart-title">Cost by Project</div>
          <div class="chart-meta">{{ formatCostDisplay(totalCost) }} total</div>
        </div>
        <div class="chart-canvas-wrap tall">
          <Bar v-if="projectBarData" :data="projectBarData" :options="projectBarOptions" />
        </div>
      </div>

      <div class="chart-card">
        <div class="chart-header">
          <div class="chart-title">Share of Spend</div>
        </div>
        <div class="chart-canvas-wrap tall">
          <Doughnut v-if="projectDonutData" :data="projectDonutData" :options="donutOptions" />
        </div>
        <div class="donut-legend">
          <div v-for="(p, i) in topProjectsForLegend" :key="p.project" class="legend-row">
            <div class="legend-left">
              <div class="legend-dot" :style="{ background: projectColors[i] }"></div>
              <span>{{ p.project }}</span>
            </div>
            <div class="legend-val">{{ formatCostDisplay(p.total_cost) }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Monthly cost per project stacked bar -->
    <div class="chart-card full-width" v-if="monthlyData.length">
      <div class="chart-header">
        <div class="chart-title">Monthly Spend by Project</div>
      </div>
      <div class="chart-canvas-wrap tall">
        <Bar v-if="monthlyChartData" :data="monthlyChartData" :options="monthlyBarOptions" />
      </div>
    </div>

    <!-- Project table -->
    <div class="section-header">
      <div class="section-title">All Projects</div>
    </div>

    <div class="sessions-table-wrap" v-if="projects.length">
      <table>
        <thead>
          <tr>
            <th style="width:40px">#</th>
            <th>Project</th>
            <th class="right">Sessions</th>
            <th class="right">Tokens</th>
            <th>Last Active</th>
            <th class="right">Cost</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(p, i) in projects" :key="p.project">
            <td class="rank">{{ i + 1 }}</td>
            <td class="project-name">{{ p.project }}</td>
            <td class="mono right">{{ p.session_count }}</td>
            <td class="mono right dim">{{ formatTokens(p.total_tokens) }}</td>
            <td class="mono dim">{{ formatDate(p.last_activity) }}</td>
            <td class="cost-cell" :class="{ top: i === 0 }">{{ formatCostDisplay(p.total_cost) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Bar, Doughnut } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  ArcElement,
  Tooltip,
  Legend,
} from 'chart.js'
import type { ProjectSummary, ProjectMonthly } from '../types'
import { fetchProjects, fetchProjectMonthly } from '../api'
import { formatCostDisplay, formatTokens, formatDate } from '../composables/useFormatCost'

ChartJS.register(CategoryScale, LinearScale, BarElement, ArcElement, Tooltip, Legend)

const projects = ref<ProjectSummary[]>([])
const monthlyData = ref<ProjectMonthly[]>([])

const projectColors = [
  '#f59e0b', '#fbbf24', '#fcd34d', '#d97706',
  '#92400e', '#78716c', '#57534e', '#44403c',
  '#a8a29e', '#6b7280', '#4b5563', '#374151',
]

const totalCost = computed(() =>
  projects.value.reduce((sum, p) => sum + p.total_cost, 0)
)

const topProjectsForLegend = computed(() =>
  projects.value.slice(0, 8)
)

// Horizontal bar chart: cost by project
const projectBarData = computed(() => {
  if (!projects.value.length) return null
  const top = projects.value.slice(0, 12)
  return {
    labels: top.map(p => p.project),
    datasets: [{
      data: top.map(p => p.total_cost),
      backgroundColor: top.map((_, i) => projectColors[i % projectColors.length]),
      borderColor: 'transparent',
      borderWidth: 0,
      borderRadius: 0,
    }],
  }
})

const projectBarOptions = {
  responsive: true,
  maintainAspectRatio: false,
  indexAxis: 'y' as const,
  animation: { duration: 700, easing: 'easeOutQuart' as const },
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: '#1a1a18',
      borderColor: '#2a2a27',
      borderWidth: 1,
      titleColor: '#8c8a84',
      bodyColor: '#f59e0b',
      bodyFont: { family: 'JetBrains Mono', size: 13 },
      titleFont: { family: 'DM Sans', size: 11 },
      padding: 12,
      callbacks: {
        label: (ctx: any) => ' $' + ctx.parsed.x.toFixed(2),
      },
    },
  },
  scales: {
    x: {
      grid: { color: '#1e1e1b' },
      border: { color: 'transparent' },
      ticks: {
        color: '#5a5855',
        font: { family: 'JetBrains Mono', size: 10 },
        callback: (v: any) => '$' + Number(v).toFixed(0),
      },
    },
    y: {
      grid: { color: 'transparent' },
      border: { color: '#1e1e1b' },
      ticks: {
        color: '#8c8a84',
        font: { family: 'DM Sans', size: 12 },
      },
    },
  },
}

// Donut chart: share of total spend
const projectDonutData = computed(() => {
  if (!projects.value.length) return null
  const top = projects.value.slice(0, 8)
  const otherCost = projects.value.slice(8).reduce((s, p) => s + p.total_cost, 0)
  const labels = top.map(p => p.project)
  const data = top.map(p => p.total_cost)
  if (otherCost > 0) {
    labels.push('Other')
    data.push(otherCost)
  }
  return {
    labels,
    datasets: [{
      data,
      backgroundColor: [...projectColors.slice(0, top.length), '#292524'],
      borderColor: '#0a0a09',
      borderWidth: 3,
    }],
  }
})

const donutOptions = {
  responsive: true,
  maintainAspectRatio: false,
  cutout: '68%',
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
      padding: 10,
      callbacks: {
        label: (ctx: any) => {
          const total = ctx.dataset.data.reduce((a: number, b: number) => a + b, 0)
          const pct = total > 0 ? Math.round((ctx.parsed / total) * 100) : 0
          return ` $${ctx.parsed.toFixed(2)} (${pct}%)`
        },
      },
    },
  },
}

// Monthly stacked bar chart
const monthlyChartData = computed(() => {
  if (!monthlyData.value.length) return null

  // Get unique months and top projects
  const months = [...new Set(monthlyData.value.map(d => d.month))].sort()
  const topProjects = projects.value.slice(0, 6).map(p => p.project)

  const datasets = topProjects.map((project, i) => ({
    label: project,
    data: months.map(month => {
      const entry = monthlyData.value.find(d => d.project === project && d.month === month)
      return entry ? entry.cost : 0
    }),
    backgroundColor: projectColors[i % projectColors.length],
    borderColor: 'transparent',
    borderWidth: 0,
  }))

  // Add "Other" dataset
  const otherProjects = new Set(projects.value.slice(6).map(p => p.project))
  if (otherProjects.size > 0) {
    datasets.push({
      label: 'Other',
      data: months.map(month =>
        monthlyData.value
          .filter(d => d.month === month && otherProjects.has(d.project))
          .reduce((s, d) => s + d.cost, 0)
      ),
      backgroundColor: '#292524',
      borderColor: 'transparent',
      borderWidth: 0,
    })
  }

  return {
    labels: months.map(m => {
      const [y, mo] = m.split('-')
      const d = new Date(Number(y), Number(mo) - 1)
      return d.toLocaleDateString('en-GB', { month: 'short', year: '2-digit' })
    }),
    datasets,
  }
})

const monthlyBarOptions = {
  responsive: true,
  maintainAspectRatio: false,
  animation: { duration: 700, easing: 'easeOutQuart' as const },
  plugins: {
    legend: {
      display: true,
      position: 'bottom' as const,
      labels: {
        color: '#8c8a84',
        font: { family: 'DM Sans', size: 11 },
        boxWidth: 8,
        boxHeight: 8,
        padding: 16,
      },
    },
    tooltip: {
      backgroundColor: '#1a1a18',
      borderColor: '#2a2a27',
      borderWidth: 1,
      titleColor: '#8c8a84',
      bodyColor: '#f0ede8',
      bodyFont: { family: 'JetBrains Mono', size: 12 },
      padding: 12,
      callbacks: {
        label: (ctx: any) => ` ${ctx.dataset.label}: $${ctx.parsed.y.toFixed(2)}`,
      },
    },
  },
  scales: {
    x: {
      stacked: true,
      grid: { color: 'transparent' },
      border: { color: '#1e1e1b' },
      ticks: {
        color: '#5a5855',
        font: { family: 'DM Sans', size: 11 },
      },
    },
    y: {
      stacked: true,
      grid: { color: '#1e1e1b' },
      border: { color: 'transparent' },
      ticks: {
        color: '#5a5855',
        font: { family: 'JetBrains Mono', size: 10 },
        callback: (v: any) => '$' + Number(v).toFixed(0),
      },
    },
  },
}

onMounted(async () => {
  const [p, m] = await Promise.all([fetchProjects(), fetchProjectMonthly()])
  projects.value = p || []
  monthlyData.value = m || []
})
</script>

<style scoped>
.page-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  margin-bottom: var(--space-8);
  animation: fadeSlideUp 0.4s ease both;
}
.page-title {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 36px;
  letter-spacing: 0.04em;
  color: var(--text-primary);
  line-height: 1;
}
.page-meta {
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
  color: var(--text-tertiary);
  padding-bottom: 4px;
}

.charts-row {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: var(--space-5);
  margin-bottom: var(--space-6);
}

.chart-card {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  padding: var(--space-6);
  animation: fadeSlideUp 0.45s ease both;
  animation-delay: 100ms;
}
.chart-card.full-width {
  margin-bottom: var(--space-8);
  animation-delay: 200ms;
}
.chart-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-5);
}
.chart-title {
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--text-tertiary);
}
.chart-meta {
  font-family: 'JetBrains Mono', monospace;
  font-size: 11px;
  color: var(--text-tertiary);
}
.chart-canvas-wrap {
  height: 180px;
  position: relative;
}
.chart-canvas-wrap.tall {
  height: 260px;
}

.donut-legend {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  margin-top: var(--space-4);
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

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-4);
  animation: fadeSlideUp 0.45s ease both;
  animation-delay: 300ms;
}
.section-title {
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--text-tertiary);
}

.sessions-table-wrap {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  overflow: hidden;
  animation: fadeSlideUp 0.45s ease both;
  animation-delay: 350ms;
}
table { width: 100%; font-size: 13px; }
thead th {
  padding: var(--space-3) var(--space-5);
  text-align: left;
  font-size: 10.5px;
  font-weight: 500;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-tertiary);
  border-bottom: 1px solid var(--border-subtle);
  white-space: nowrap;
}
thead th.right { text-align: right; }

tbody tr {
  border-bottom: 1px solid var(--border-subtle);
  transition: background 100ms;
}
tbody tr:last-child { border-bottom: none; }
tbody tr:hover { background: var(--bg-elevated); }

td {
  padding: var(--space-4) var(--space-5);
  color: var(--text-secondary);
  vertical-align: middle;
}
td.right { text-align: right; }
.rank {
  font-family: 'JetBrains Mono', monospace;
  font-size: 11px;
  color: var(--text-disabled);
  width: 32px;
  text-align: right;
  padding-right: var(--space-2);
}
tbody tr:first-child .rank { color: var(--amber-500); }
.project-name {
  color: var(--text-primary);
  font-weight: 400;
}
.mono {
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
}
.dim { color: var(--text-tertiary); }
.cost-cell {
  font-family: 'JetBrains Mono', monospace;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  text-align: right;
}
.cost-cell.top { color: var(--amber-400); }
</style>
