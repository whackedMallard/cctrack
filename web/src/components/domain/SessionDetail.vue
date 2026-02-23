<template>
  <div class="session-detail" v-if="session">
    <div class="detail-header">
      <h2 class="detail-title">{{ session.project || session.id.slice(0, 12) }}</h2>
      <Badge :label="formatModel(session.model)" />
    </div>

    <div class="detail-meta">
      <div class="meta-row">
        <span class="meta-label">Session ID</span>
        <span class="meta-value mono">{{ session.id }}</span>
      </div>
      <div class="meta-row">
        <span class="meta-label">Project</span>
        <span class="meta-value">{{ session.project }}</span>
      </div>
      <div class="meta-row" v-if="session.slug">
        <span class="meta-label">Slug</span>
        <span class="meta-value mono">{{ session.slug }}</span>
      </div>
      <div class="meta-row">
        <span class="meta-label">Started</span>
        <span class="meta-value mono">{{ session.started_at }}</span>
      </div>
      <div class="meta-row">
        <span class="meta-label">Last Activity</span>
        <span class="meta-value mono">{{ session.last_activity }}</span>
      </div>
    </div>

    <div class="detail-section">
      <div class="section-label">Token Breakdown</div>
      <table class="breakdown-table">
        <tr>
          <td>Input</td>
          <td class="mono right">{{ formatTokensRaw(session.total_input) }}</td>
        </tr>
        <tr>
          <td>Output</td>
          <td class="mono right">{{ formatTokensRaw(session.total_output) }}</td>
        </tr>
        <tr>
          <td>Cache Read</td>
          <td class="mono right">{{ formatTokensRaw(session.total_cache_read) }}</td>
        </tr>
        <tr>
          <td>Cache Write</td>
          <td class="mono right">{{ formatTokensRaw(session.total_cache_write) }}</td>
        </tr>
        <tr class="total-row">
          <td>Total</td>
          <td class="mono right">{{ formatTokensRaw(totalTokens) }}</td>
        </tr>
      </table>
    </div>

    <div class="detail-section">
      <div class="section-label">Cost</div>
      <div class="cost-display">{{ formatCostDisplay(session.total_cost) }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Session } from '../../types'
import Badge from '../primitives/Badge.vue'
import { formatCostDisplay, formatTokensRaw, formatModel } from '../../composables/useFormatCost'

const props = defineProps<{ session: Session | null }>()

const totalTokens = computed(() => {
  if (!props.session) return 0
  return props.session.total_input + props.session.total_output +
    props.session.total_cache_read + props.session.total_cache_write
})
</script>

<style scoped>
.session-detail {
  display: flex;
  flex-direction: column;
  gap: var(--space-8);
}
.detail-header {
  display: flex;
  align-items: center;
  gap: var(--space-4);
}
.detail-title {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 28px;
  letter-spacing: 0.02em;
  color: var(--text-primary);
}
.detail-meta {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}
.meta-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
}
.meta-label {
  color: var(--text-tertiary);
}
.meta-value {
  color: var(--text-secondary);
}
.meta-value.mono {
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
}
.detail-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}
.section-label {
  font-size: 10.5px;
  font-weight: 500;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--text-tertiary);
}
.breakdown-table {
  width: 100%;
}
.breakdown-table td {
  padding: var(--space-2) 0;
  font-size: 13px;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--border-subtle);
}
.breakdown-table .right {
  text-align: right;
}
.breakdown-table .mono {
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
}
.total-row td {
  color: var(--text-primary);
  font-weight: 500;
  border-bottom: none;
  border-top: 1px solid var(--border-default);
}
.cost-display {
  font-family: 'Bebas Neue', sans-serif;
  font-size: 48px;
  color: var(--amber-400);
  line-height: 1;
}
</style>
