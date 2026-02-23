<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">Sessions</h1>
      <div class="page-meta">{{ store.total }} total</div>
    </div>

    <div class="sessions-table-wrap">
      <table>
        <thead>
          <tr>
            <th style="width:40px">#</th>
            <th class="sortable" @click="store.setSort('date')">
              Session
              <span v-if="store.sortBy === 'date'" class="sort-arrow">{{ store.sortDir === 'desc' ? '↓' : '↑' }}</span>
            </th>
            <th class="sortable" @click="store.setSort('model')">
              Model
              <span v-if="store.sortBy === 'model'" class="sort-arrow">{{ store.sortDir === 'desc' ? '↓' : '↑' }}</span>
            </th>
            <th class="sortable" @click="store.setSort('started')">
              Started
              <span v-if="store.sortBy === 'started'" class="sort-arrow">{{ store.sortDir === 'desc' ? '↓' : '↑' }}</span>
            </th>
            <th class="right sortable" @click="store.setSort('tokens')">
              Tokens
              <span v-if="store.sortBy === 'tokens'" class="sort-arrow">{{ store.sortDir === 'desc' ? '↓' : '↑' }}</span>
            </th>
            <th class="right sortable" @click="store.setSort('cost')">
              Cost
              <span v-if="store.sortBy === 'cost'" class="sort-arrow">{{ store.sortDir === 'desc' ? '↓' : '↑' }}</span>
            </th>
          </tr>
        </thead>
        <tbody>
          <SessionRow
            v-for="(session, i) in store.sessions"
            :key="session.id"
            :session="session"
            :rank="store.offset + i + 1"
            @select="store.selectSession"
          />
        </tbody>
      </table>
    </div>

    <div class="pagination" v-if="store.total > store.limit">
      <button @click="store.prevPage()" :disabled="store.offset === 0">← Prev</button>
      <span class="page-info">
        {{ store.offset + 1 }}–{{ Math.min(store.offset + store.limit, store.total) }} of {{ store.total }}
      </span>
      <button @click="store.nextPage()" :disabled="store.offset + store.limit >= store.total">Next →</button>
    </div>

    <SlideOver :open="!!store.selectedSession" @close="store.clearSelection()">
      <SessionDetail :session="store.selectedSession" />
    </SlideOver>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useSessionsStore } from '../stores/sessions'
import SessionRow from '../components/domain/SessionRow.vue'
import SessionDetail from '../components/domain/SessionDetail.vue'
import SlideOver from '../components/primitives/SlideOver.vue'

const store = useSessionsStore()

onMounted(() => {
  store.load()
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

.sessions-table-wrap {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  overflow: hidden;
  animation: fadeSlideUp 0.45s ease both;
  animation-delay: 100ms;
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
thead th.sortable {
  cursor: pointer;
  user-select: none;
  transition: color 150ms;
}
thead th.sortable:hover { color: var(--text-secondary); }
.sort-arrow {
  color: var(--amber-500);
  margin-left: 4px;
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-6);
  margin-top: var(--space-6);
}
.pagination button {
  background: var(--bg-surface);
  border: 1px solid var(--border-default);
  color: var(--text-secondary);
  padding: var(--space-2) var(--space-4);
  font-size: 13px;
  cursor: pointer;
  transition: background 150ms, color 150ms;
}
.pagination button:hover:not(:disabled) {
  background: var(--bg-elevated);
  color: var(--text-primary);
}
.pagination button:disabled {
  opacity: 0.3;
  cursor: default;
}
.page-info {
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
  color: var(--text-tertiary);
}
</style>
