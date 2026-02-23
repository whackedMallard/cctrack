<template>
  <span class="model-badge">{{ label }}</span>
</template>

<script setup lang="ts">
defineProps<{ label: string }>()
</script>

<style scoped>
.model-badge {
  font-family: 'JetBrains Mono', monospace;
  font-size: 10.5px;
  color: var(--text-tertiary);
  background: var(--bg-subtle);
  border: 1px solid var(--border-subtle);
  padding: 2px 7px;
  display: inline-block;
}
</style>
