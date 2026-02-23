<template>
  <Teleport to="body">
    <Transition name="slide">
      <div v-if="open" class="slide-over-backdrop" @click.self="$emit('close')">
        <div class="slide-over-panel" @keydown.escape="$emit('close')">
          <button class="close-btn" @click="$emit('close')" aria-label="Close panel">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M4 4l8 8M12 4l-8 8"/>
            </svg>
          </button>
          <slot />
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
defineProps<{ open: boolean }>()
defineEmits<{ close: [] }>()
</script>

<style scoped>
.slide-over-backdrop {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  justify-content: flex-end;
}
.slide-over-panel {
  width: 480px;
  max-width: 100%;
  height: 100vh;
  background: var(--bg-surface);
  border-left: 1px solid var(--border-default);
  overflow-y: auto;
  padding: var(--space-8);
  position: relative;
}
.close-btn {
  position: absolute;
  top: var(--space-6);
  right: var(--space-6);
  background: none;
  border: none;
  color: var(--text-tertiary);
  cursor: pointer;
  padding: var(--space-2);
  transition: color 150ms;
}
.close-btn:hover {
  color: var(--text-primary);
}

.slide-enter-active,
.slide-leave-active {
  transition: transform 200ms ease;
}
.slide-enter-active .slide-over-panel,
.slide-leave-active .slide-over-panel {
  transition: transform 200ms ease;
}
.slide-enter-from .slide-over-panel,
.slide-leave-to .slide-over-panel {
  transform: translateX(100%);
}
</style>
