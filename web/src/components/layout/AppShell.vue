<template>
  <div class="app-shell">
    <SidebarNav :connectionStatus="connectionStatus" />
    <main class="main">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
import type { ConnectionStatus } from '../../types'
import SidebarNav from './SidebarNav.vue'

defineProps<{ connectionStatus: ConnectionStatus }>()
</script>

<style scoped>
.app-shell {
  display: flex;
  min-height: 100vh;
}
.main {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-10) var(--space-10) var(--space-16);
  max-width: calc(1200px + var(--space-10) * 2);
}
.main::-webkit-scrollbar { width: 6px; }
.main::-webkit-scrollbar-track { background: transparent; }
.main::-webkit-scrollbar-thumb { background: var(--border-default); }
.main::-webkit-scrollbar-thumb:hover { background: var(--border-strong); }
</style>
